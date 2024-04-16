package bindings

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var server *http.Server

func (b *Bindings) ServerNew() {
	if server != nil {
		return
	}

	server = &http.Server{
		Addr: fmt.Sprintf("127.0.0.1:%d", b.options.LoopbackPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			w.WriteHeader(204)
			r.URL.Host = r.Host
			r.URL.Scheme = "http"
			request := map[string]any{
				"method":  r.Method,
				"url":     r.URL.String(),
				"headers": map[string]string{},
				"body":    body,
			}
			for key, value := range r.Header {
				request["headers"].(map[string]string)[key] = value[0]
			}
			encoded, err := json.Marshal(request)
			if err != nil {
				return
			}
			sequence := time.Now().UnixNano()
			b.ui.Eval(fmt.Sprintf("sages(%s, %d)", encoded, sequence))
		}),
	}

	go server.ListenAndServe()
}

func (b *Bindings) ServerDestroy() {
	if server != nil {
		server.Shutdown(context.Background())
		server = nil
	}
}
