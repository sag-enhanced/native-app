package bindings

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"sync"
	"time"

	browserAPI "github.com/sag-enhanced/native-app/src/browser"
)

var browserHandles = map[string]*browserAPI.BrowserChannels{}
var browserHandleLock = sync.Mutex{}

func (b *Bindings) BrowserNew(pageUrl string, code string, browser string, proxy *string, profileId int32) (string, error) {
	rawHandle := make([]byte, 16)
	var err error
	if _, err := rand.Read(rawHandle); err != nil {
		return "", err
	}
	handle := fmt.Sprintf("%x", rawHandle)
	if b.options.Verbose {
		fmt.Println("Created new browser instance with handle", handle)
	}
	var parsedProxy *url.URL
	if proxy != nil {
		parsedProxy, err = url.Parse(*proxy)
		if err != nil {
			return "", err
		}
	}

	channels := &browserAPI.BrowserChannels{
		Result: make(chan string, 5),
		Stop:   make(chan string, 5),
	}
	go func() {
		err := browserAPI.RunBrowser(channels, b.options, pageUrl, code, browser, parsedProxy, profileId)
		if err != nil {
			fmt.Println("Error running browser:", err)
		}
	}()

	browserHandleLock.Lock()
	browserHandles[handle] = channels
	browserHandleLock.Unlock()
	return handle, nil
}
func (b *Bindings) BrowserGet(handle string, timeout int64) (string, error) {
	browserHandleLock.Lock()
	browser, ok := browserHandles[handle]
	browserHandleLock.Unlock()
	if !ok {
		return "", errors.New("invalid handle")
	}

	select {
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return "", errors.New("timeout")
	case msg := <-browser.Result:
		if msg == "closed" {
			browserHandleLock.Lock()
			delete(browserHandles, handle)
			browserHandleLock.Unlock()
			return "", errors.New("closed")
		}
		return msg, nil
	}
}
func (b *Bindings) BrowserDestroy(handle string) {
	browserHandleLock.Lock()
	defer browserHandleLock.Unlock()
	browser, ok := browserHandles[handle]
	if !ok {
		return
	}
	delete(browserHandles, handle)
	if b.options.Verbose {
		fmt.Println("Destroying browser instance with handle", handle)
	}
	browser.Stop <- "quit"
}

func (b *Bindings) BrowserDestroyProfile(browser string) error {
	profilePath := path.Join(b.options.DataDirectory, "profiles", browser)
	return os.RemoveAll(profilePath)
}
