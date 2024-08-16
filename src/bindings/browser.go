package bindings

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"

	browserAPI "github.com/sag-enhanced/native-app/src/browser"
)

var browserHandles = map[string]context.CancelFunc{}
var browserHandleLock = sync.Mutex{}

func (b *Bindings) BrowserNew(pageUrl string, browser string, proxy *string, profileId int32) (string, error) {
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
		if parsedProxy.Hostname() != "127.0.0.1" {
			return "", errors.New("Only local proxies are allowed.")
		}
	}

	if _, err := url.Parse(pageUrl); err != nil {
		return "", err
	}

	cancelCtx, cancel := context.WithCancel(context.Background())

	go func() {
		defer cancel()
		defer func() {
			browserHandleLock.Lock()
			defer browserHandleLock.Unlock()
			delete(browserHandles, handle)

			if b.options.Verbose {
				fmt.Println("Destroying browser instance with handle", handle)
			}

			b.ui.Eval(fmt.Sprintf("sagebd(%q)", handle))
		}()
		err := browserAPI.RunBrowser(cancelCtx, b.options, pageUrl, browser, parsedProxy, profileId)
		if err != nil {
			fmt.Println("Error running browser:", err)
		}
	}()

	browserHandleLock.Lock()
	browserHandles[handle] = cancel
	browserHandleLock.Unlock()
	return handle, nil
}

func (b *Bindings) BrowserDestroy(handle string) {
	browserHandleLock.Lock()
	defer browserHandleLock.Unlock()
	cancelCtx, ok := browserHandles[handle]
	if !ok {
		return
	}
	delete(browserHandles, handle)
	if b.options.Verbose {
		fmt.Println("Destroying browser instance with handle", handle)
	}
	cancelCtx()
}

func (b *Bindings) BrowserDestroyProfile(browser string, profileId string) error {
	if strings.ContainsAny(browser+profileId, "/\\.;:") {
		return errors.New("invalid browser name")
	}
	profilePath := path.Join(b.options.DataDirectory, "profiles", browser)
	if profileId != "" {
		profilePath = path.Join(profilePath, profileId)
	}
	return os.RemoveAll(profilePath)
}
