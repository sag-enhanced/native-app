package bindings

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

var server *http.Server
var serverRequests = make(map[int]chan serverResponse)
var serverRequestLock = sync.Mutex{}

func (b *Bindings) ServerNew() {
	if server != nil {
		return
	}

	requestId := 0
	server = &http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", b.options.LoopbackPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				w.WriteHeader(500)
				return
			}
			r.URL.Host = r.Host
			r.URL.Scheme = "http"

			request := serverRequest{
				Id:      requestId,
				Method:  r.Method,
				Url:     r.URL.String(),
				Headers: map[string]string{},
				Body:    string(body),
			}
			requestId++
			for key, value := range r.Header {
				request.Headers[key] = value[0]
			}

			if b.options.Verbose {
				fmt.Println("received HTTP request", request.Method, request.Url)
			}

			encoded, err := json.Marshal(request)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			sequence := time.Now().UnixNano()

			responseChannel := make(chan serverResponse)

			serverRequestLock.Lock()
			serverRequests[requestId] = responseChannel
			serverRequestLock.Unlock()

			b.ui.Eval(fmt.Sprintf("sages(%s, %d)", encoded, sequence))

			select {
			case response := <-responseChannel:
				for key, value := range response.Headers {
					w.Header().Set(key, value)
				}
				w.WriteHeader(response.StatusCode)
				w.Write([]byte(response.Body))
				if b.options.Verbose {
					fmt.Println("HTTP request", request.Method, request.Url, "returned", response.StatusCode)
				}
			case <-time.After(10 * time.Second):
				w.WriteHeader(502)
				if b.options.Verbose {
					fmt.Println("HTTP request timed out")
				}
			}

			serverRequestLock.Lock()
			delete(serverRequests, requestId)
			serverRequestLock.Unlock()
		}),
	}

	go server.ListenAndServe()
}

func (b *Bindings) ServerRespond(requestId int, statusCode int, headers map[string]string, body string) error {
	serverRequestLock.Lock()
	responseChannel, ok := serverRequests[requestId]
	if ok {
		delete(serverRequests, requestId)
	}
	serverRequestLock.Unlock()

	if !ok {
		return fmt.Errorf("invalid request id")
	}

	response := serverResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
	}
	responseChannel <- response

	return nil
}

func (b *Bindings) ServerDestroy() {
	if server != nil {
		server.Shutdown(context.Background())
		server = nil
	}
}

type serverResponse struct {
	StatusCode int               `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type serverRequest struct {
	Id      int               `json:"id"`
	Method  string            `json:"method"`
	Url     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}
