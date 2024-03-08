package app

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/playwright-community/playwright-go"
)

const (
	// minimum window size that works with the recaptcha popup shown
	WIDTH  = 440
	HEIGHT = 620
)

func runPlaywright(chint chan string, chout chan string, url string, code string, browserName string, proxy *url.URL, options Options) error {
	args := []string{}
	if browserName == "chromium" {
		// this unsets navigator.webdriver, which is used to detect automation
		args = append(args, "--disable-blink-features=AutomationControlled")
		args = append(args, fmt.Sprintf("--window-size=%d,%d", WIDTH, HEIGHT))
	} else {
		// unsetting navigator.webdriver is not required in firefox
		args = append(args, fmt.Sprintf("--width=%d", WIDTH))
		args = append(args, fmt.Sprintf("--height=%d", HEIGHT))
	}

	extensions, err := getExtensionList(browserName)
	if err != nil {
		return err
	}
	for _, extension := range extensions {
		// dont ask me why we have to do this twice
		args = append(args, "--disable-extensions-except="+extension)
		args = append(args, "--load-extension="+extension)
	}

	if options.Verbose {
		fmt.Println("Running playwright with args", args)
	}

	pw, err := playwright.Run(&playwright.RunOptions{
		Browsers: []string{browserName},
	})
	if err != nil {
		return err
	}
	defer pw.Stop()

	var playwrightProxy *playwright.Proxy
	if proxy != nil {
		var username, password *string
		if proxy.User != nil {
			username = playwright.String(proxy.User.Username())
			tempPassword, ok := proxy.User.Password()
			if ok {
				password = &tempPassword
			}
		}
		playwrightProxy = &playwright.Proxy{
			Server:   proxy.Scheme + "://" + proxy.Host,
			Username: username,
			Password: password,
		}
	}

	profileName := "pw-profile"
	if browserName != "chromium" {
		profileName += "-" + browserName
	}

	profilePath := path.Join(getStoragePath(), profileName)
	var browser playwright.BrowserContext
	playwrightOptions := playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: playwright.Bool(false),
		Args:     args,
		IgnoreDefaultArgs: []string{
			// disables "Chrome is being controlled by automated test software" banner
			"--enable-automation",
		},
		NoViewport: playwright.Bool(true),
		Viewport: &playwright.Size{
			Width:  WIDTH,
			Height: HEIGHT,
		},
		Proxy:  playwrightProxy,
		Locale: playwright.String("en-US"),
	}
	if browserName == "chromium" {
		browser, err = pw.Chromium.LaunchPersistentContext(profilePath, playwrightOptions)
	} else {
		browser, err = pw.Firefox.LaunchPersistentContext(profilePath, playwrightOptions)
	}

	if err != nil {
		return err
	}
	defer browser.Close()

	fullyLoaded := false
	browser.OnClose(func(_ playwright.BrowserContext) {
		if options.Verbose {
			fmt.Println("Browser closed")
		}
		if fullyLoaded {
			chint <- "quit"
		} else {
			chout <- "closed"
			fmt.Println("Browser was closed before it was fully loaded, this is likely caused by a bad proxy.")
			fmt.Println("This is a memory leak which we can't currently fix, so please don't do this too often.")
		}
	})

	if options.Verbose {
		fmt.Println("Browser running; waiting for page")
	}
	page, err := browser.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	page.AddInitScript(playwright.Script{
		Content: playwright.String(code),
	})

	if options.Verbose {
		fmt.Println("Navigating to", url)
	}
	if _, err := page.Goto(url); err != nil {
		return err
	}
	if options.Verbose {
		fmt.Println("Waiting for page to load")
	}
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	}); err != nil {
		return err
	}

	page.OnDialog(func(d playwright.Dialog) {
		message := d.Message()
		if options.Verbose {
			fmt.Println("Dialog:", message)
		}
		if strings.HasPrefix(message, "SAGE#") {
			chout <- message[5:]
		}
		d.Dismiss()
	})

	page.OnClose(func(_ playwright.Page) {
		if options.Verbose {
			fmt.Println("Page closed")
		}
		chint <- "quit"
	})

	fullyLoaded = true

	if options.Verbose {
		fmt.Println("Page loaded")
	}
	for {
		msg := <-chint
		if msg == "quit" {
			break
		}
		page.Evaluate(msg)
	}
	chout <- "closed" // signal to .Get() that we are closed
	return nil
}

func getExtensionList(browser string) ([]string, error) {
	ext := path.Join(getStoragePath(), "ext", browser)
	files, err := os.ReadDir(ext)
	extensions := []string{}
	if err != nil {
		return extensions, nil
	}
	for _, file := range files {
		if file.IsDir() {
			extensions = append(extensions, path.Join(ext, file.Name()))
		}
	}
	return extensions, nil
}
