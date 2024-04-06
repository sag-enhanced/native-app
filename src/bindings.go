package app

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/denisbrodbeck/machineid"
	"github.com/gen2brain/beeep"
	"github.com/sqweek/dialog"
)

func (app *App) registerBindings() {
	app.bind("build", func() uint32 {
		return build
	})

	app.bind("start", func() int64 {
		return app.start
	})

	app.bind("open", func(target string) {
		// some sanization checks
		if strings.ContainsAny(target, "\n\r'\"` {}$|;") {
			return
		}
		url, err := url.Parse(target)
		// only allow https urls and block any path traversal attempts
		if err != nil || url.Scheme != "https" || strings.Contains(url.Path, "..") {
			return
		}
		fmt.Println("Opening URL", url.String())
		// re-assemble url to string to avoid any funny business
		app.open(url.String())
	})

	app.bind("save", func(filename string, data string) error {
		path, err := dialog.File().Title("Save file").SetStartFile(filename).Filter("All files", "*").Save()
		if err != nil {
			return err
		}

		return os.WriteFile(path, []byte(data), 0644)
	})

	app.bind("read", func(filterText string, filter string) (string, error) {
		path, err := dialog.File().Title("Open file").Filter(filterText, filter).Load()
		if err != nil {
			return "", err
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(data), nil
	})

	app.bind("alert", func(message string) {
		dialog.Message(message).Title("Alert").Info()
	})

	app.bind("notify", func(title string, message string, alert bool) error {
		err := beeep.Notify(title, message, "")
		if err != nil {
			return err
		}
		if alert {
			return beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
		}
		return nil
	})

	app.bind("steamDesktopLogin", func(username string, password string) error {
		err := app.closeSteam()
		if err != nil {
			return err
		}
		return app.runSteamWithArguments("-login", username, password)
	})

	app.bind("steamPatch", func(js string) error {
		exe, err := app.findSteamExecutable()
		if err != nil {
			return err
		}

		if app.options.Verbose {
			fmt.Println("Steam executable found at", exe)
		}

		data, err := app.findSteamDataDir()
		if err != nil {
			return err
		}
		if app.options.Verbose {
			fmt.Println("Steam data directory found at", data)
		}

		entryFile := path.Join(data, "steamui", "library.js")
		content, err := os.ReadFile(entryFile)
		if err != nil {
			return err
		}

		// inject our code into the steam client
		lines := strings.Split(string(content), "\n")[:3]
		if js != "" {
			lines = append(lines, js)
		}

		return os.WriteFile(entryFile, []byte(strings.Join(lines, "\n")), 0644)
	})

	app.bind("steamRun", func() error {
		app.closeSteam()

		if app.options.Verbose {
			fmt.Println("Starting Steam with injected code...")
		}
		// -noverifyfiles is required to prevent steam from checking the files
		// and redownloading them if they are modified
		return app.runSteamWithArguments("-noverifyfiles")
	})

	app.bind("id", func() string {
		return app.identity.Id()
	})

	app.bind("sign", func(message string) ([]byte, error) {
		return app.identity.Sign([]byte(message))
	})

	app.bind("get", func(key string) (string, error) {
		file := path.Join(getStoragePath(), key+".dat")
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		reader := flate.NewReader(bytes.NewReader(data))
		decompressed, err := io.ReadAll(reader)
		if err != nil {
			return "", err
		}
		return string(decompressed), nil
	})

	app.bind("set", func(key string, value string) error {
		file := path.Join(getStoragePath(), key+".dat")
		if key == "accounts" {
			if stat, err := os.Stat(file); err == nil && int(stat.Size()) > len(value) && stat.Size() > 100 {
				fmt.Println("New value is smaller than the old one, refusing to overwrite")
				fmt.Println("Old size:", stat.Size(), "New size:", len(value))
				fmt.Println("This is a bug (that could've nuked your accounts!), please report it")
				return errors.New("new value is smaller than the old one, refusing to overwrite")
			}
		}
		fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0644)
		defer fd.Close()
		writer, err := flate.NewWriter(fd, flate.BestCompression)
		if err != nil {
			return err
		}
		defer writer.Close()
		writer.Write([]byte(value))
		return nil
	})

	clientHandles := map[string]http.Client{}
	app.bind("httpClient", func(proxyUrl *string) (string, error) {
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
		if app.options.Verbose {
			fmt.Println("Created new HTTP client with handle", handle)
		}
		clientHandles[handle] = http.Client{Jar: jar, Transport: &http.Transport{Proxy: proxy}}
		return handle, nil
	})

	app.bind("httpRequest", func(handle string, method string, url string, headers map[string]string, body string) (*HTTPResponse, error) {
		client, ok := clientHandles[handle]
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

		if app.options.Verbose {
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
	})

	app.bind("httpCookie", func(handle string, domain string, name string, value *string) (string, error) {
		client, ok := clientHandles[handle]
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
	})

	app.bind("httpDestroy", func(handle string) {
		if app.options.Verbose {
			fmt.Println("Destroying HTTP client with handle", handle)
		}
		delete(clientHandles, handle)
	})

	app.bind("ext", func(browser string) (*map[string]string, error) {
		dir := path.Join(getStoragePath(), "ext", browser)
		extensions := map[string]string{}
		files, err := os.ReadDir(dir)
		// probably directory doesnt exist, so no extensions
		if err != nil {
			return &extensions, nil
		}

		for _, file := range files {
			if !file.IsDir() {
				continue
			}
			manifest, err := os.ReadFile(path.Join(dir, file.Name(), "manifest.json"))
			if err != nil {
				continue
			}
			var parsedManifest Manifest
			err = json.Unmarshal(manifest, &parsedManifest)
			if err != nil {
				continue
			}

			extensions[file.Name()] = parsedManifest.Version
		}

		return &extensions, nil
	})

	app.bind("extInstall", func(name string, browser string, download string) error {
		if !strings.HasPrefix(download, "https://github.com/") || strings.Contains(download, "..") {
			return errors.New("invalid download URL")
		}
		if strings.Contains(name, "..") {
			return errors.New("invalid extension name")
		}

		installExtensionFromGithub(name, browser, download)

		return nil
	})

	app.bind("extGetManifest", func(name string, browser string) (string, error) {
		if strings.Contains(name, "..") {
			return "", errors.New("invalid extension name")
		}

		manifest := path.Join(getStoragePath(), "ext", browser, name, "manifest.json")

		data, err := os.ReadFile(manifest)
		if err != nil {
			return "", err
		}
		return string(data), nil
	})

	app.bind("extSetManifest", func(name string, browser string, manifest string) error {
		if strings.Contains(name, "..") {
			return errors.New("invalid extension name")
		}

		manifestPath := path.Join(getStoragePath(), "ext", browser, name, "manifest.json")
		return os.WriteFile(manifestPath, []byte(manifest), 0644)
	})

	app.bind("extUninstall", func(name string, browser string) error {
		if strings.Contains(name, "..") {
			return errors.New("invalid extension name")
		}

		dir := path.Join(getStoragePath(), "ext", browser, name)
		return os.RemoveAll(dir)
	})

	// browser automation
	browserResultHandles := map[string]chan string{}
	browserStopHandles := map[string]chan string{}
	app.bind("browserNew", func(pageUrl string, code string, browser string, proxy *string, profileId int32) (string, error) {
		rawHandle := make([]byte, 16)
		var err error
		if _, err := rand.Read(rawHandle); err != nil {
			return "", err
		}
		handle := fmt.Sprintf("%x", rawHandle)
		if app.options.Verbose {
			fmt.Println("Created new browser instance with handle", handle)
		}
		var parsedProxy *url.URL
		if proxy != nil {
			parsedProxy, err = url.Parse(*proxy)
			if err != nil {
				return "", err
			}
		}

		chResult := make(chan string, 5)
		chStop := make(chan string, 5)
		go func() {
			err := app.runBrowser(chResult, chStop, pageUrl, code, browser, parsedProxy, profileId)
			if err != nil {
				fmt.Println("Error running browser:", err)
			}
		}()

		browserResultHandles[handle] = chResult
		browserStopHandles[handle] = chStop
		return handle, nil
	})
	app.bind("browserGet", func(handle string, timeout int64) (string, error) {
		chResult, ok := browserResultHandles[handle]
		if !ok {
			return "", errors.New("invalid handle")
		}

		select {
		case <-time.After(time.Duration(timeout) * time.Millisecond):
			return "", errors.New("timeout")
		case msg := <-chResult:
			if msg == "closed" {
				delete(browserStopHandles, handle)
				delete(browserResultHandles, handle)
				return "", errors.New("closed")
			}
			return msg, nil
		}
	})
	app.bind("browserDestroy", func(handle string) {
		delete(browserResultHandles, handle)

		chStop, ok := browserStopHandles[handle]
		if !ok {
			return
		}
		if app.options.Verbose {
			fmt.Println("Destroying browser instance with handle", handle)
		}
		chStop <- "quit"
		delete(browserResultHandles, handle)
	})

	app.bind("browserDestroyProfile", func(browser string) error {
		return app.destroyBrowserProfile(browser)
	})

	var server *http.Server
	app.bind("serverNew", func() {
		if server != nil {
			return
		}

		server = &http.Server{
			Addr: "127.0.0.1:" + strconv.Itoa(loopbackPort),
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
				app.ui.eval(fmt.Sprintf("sages(%s, %d)", encoded, sequence))
			}),
		}

		go server.ListenAndServe()
	})

	app.bind("serverDestroy", func() {
		if server != nil {
			server.Shutdown(context.Background())
			server = nil
		}
	})

	app.bind("info", func() (map[string]any, error) {
		id, err := machineid.ID()
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"build": build,
			"path":  getStoragePath(),
			"os":    runtime.GOOS,
			"arch":  runtime.GOARCH,
			"id":    id,
			"port":  loopbackPort,
		}, nil
	})

	app.bind("quit", func() {
		app.ui.quit()
	})
}

type HTTPResponse struct {
	StatusCode int               `json:"status"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

type Manifest struct {
	Version string `json:"version"`
}
