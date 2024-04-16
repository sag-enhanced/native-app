package bindings

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
	"unicode/utf8"
)

var httpClients = make(map[string]http.Client)
var httpHandleLock = sync.Mutex{}

func (b *Bindings) HttpClient(proxyUrl *string) (string, error) {
	rawHandle := make([]byte, 16)
	if _, err := rand.Read(rawHandle); err != nil {
		return "", err
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", err
	}
	var proxy func(*http.Request) (*url.URL, error)
	if proxyUrl != nil {
		parsedProxyUrl, err := url.Parse(*proxyUrl)
		if err != nil {
			return "", err
		}
		proxy = http.ProxyURL(parsedProxyUrl)
	}

	handle := fmt.Sprintf("%x", rawHandle)
	if b.options.Verbose {
		fmt.Println("Created new HTTP client with handle", handle)
	}
	httpHandleLock.Lock()
	httpClients[handle] = http.Client{Jar: jar, Transport: &http.Transport{Proxy: proxy}}
	httpHandleLock.Unlock()
	return handle, nil
}

func (b *Bindings) HttpRequest(handle string, method string, url string, headers map[string]string, body string) (*HTTPResponse, error) {
	httpHandleLock.Lock()
	client, ok := httpClients[handle]
	httpHandleLock.Unlock()
	if !ok {
		return nil, errors.New("invalid handle")
	}
	var reader io.Reader
	if strings.HasPrefix(body, "data:") {
		reader = base64.NewDecoder(base64.StdEncoding, strings.NewReader(strings.Split(body, ",")[1]))
	} else {
		reader = strings.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	responseHeaders := map[string]string{}
	for key, value := range resp.Header {
		responseHeaders[key] = value[0]
	}

	if b.options.Verbose {
		fmt.Println("HTTP request", method, url, "returned", resp.StatusCode)
	}

	var stringifiedBody string
	if utf8.Valid(responseBody) {
		stringifiedBody = string(responseBody)
	} else {
		stringifiedBody = "data:;base64," + base64.StdEncoding.EncodeToString(responseBody)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    responseHeaders,
		Body:       stringifiedBody,
	}, nil
}

func (b *Bindings) HttpCookie(handle string, domain string, name string, value *string) (string, error) {
	httpHandleLock.Lock()
	client, ok := httpClients[handle]
	httpHandleLock.Unlock()
	if !ok {
		return "", errors.New("invalid handle")
	}
	if value != nil {
		client.Jar.SetCookies(&url.URL{Scheme: "https", Host: domain}, []*http.Cookie{
			{Name: name, Value: *value},
		})
	}
	cookies := client.Jar.Cookies(&url.URL{Scheme: "https", Host: domain})
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value, nil
		}
	}
	return "", errors.New("cookie not found")
}

func (b *Bindings) HttpDestroy(handle string) {
	if b.options.Verbose {
		fmt.Println("Destroying HTTP client with handle", handle)
	}
	httpHandleLock.Lock()
	delete(httpClients, handle)
	httpHandleLock.Unlock()
}

type HTTPResponse struct {
	StatusCode int               `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}
