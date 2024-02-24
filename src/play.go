package app

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/playwright-community/playwright-go"
)

func installPlaywright() error {
	fmt.Println("Note the following browser related messages are from playwright. This may take a while if it's the first time.")
	// technically we also support firefox, but as its not the default, it will likely not be used
	// as much and thus we dont install it by default
	// this is no issue, as playwright will install it on demand anyway
	return playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
		Verbose:  true,
	})
}

func runPlaywright(chint chan string, chout chan string, url string, code string, browser_name string, proxy *url.URL) error {
	args := []string{
		// this unsets navigator.webdriver, which is used to detect automation
		"--disable-blink-features=AutomationControlled",
		// minimum window size that works with the recaptcha popup shown
		"--window-size=440,720",
		// force the language to english for nopecha
		"--lang=en",
	}
	extensions, err := getExtensionList(browser_name)
	if err != nil {
		return err
	}
	for _, extension := range extensions {
		// dont ask me why we have to do this twice
		args = append(args, "--disable-extensions-except="+extension)
		args = append(args, "--load-extension="+extension)
	}

	fmt.Println("Running playwright with args", args)

	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	defer pw.Stop()

	var playwright_proxy *playwright.Proxy
	if proxy != nil {
		var username, password *string
		if proxy.User != nil {
			username = playwright.String(proxy.User.Username())
			temp_password, ok := proxy.User.Password()
			if ok {
				password = &temp_password
			}
		}
		playwright_proxy = &playwright.Proxy{
			Server:   proxy.Scheme + "://" + proxy.Host,
			Username: username,
			Password: password,
		}
	}

	profile_name := "pw-profile"
	if browser_name != "chromium" {
		profile_name += "-" + browser_name
	}

	profile_path := path.Join(getStoragePath(), profile_name)
	var browser playwright.BrowserContext
	options := playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: playwright.Bool(false),
		Args:     args,
		IgnoreDefaultArgs: []string{
			// disables "Chrome is being controlled by automated test software"
			"--enable-automation",
		},
		NoViewport: playwright.Bool(true),
		Viewport: &playwright.Size{
			Width:  440,
			Height: 620,
		},
		Proxy: playwright_proxy,
	}
	if browser_name == "chromium" {
		browser, err = pw.Chromium.LaunchPersistentContext(profile_path, options)
	} else {
		browser, err = pw.Firefox.LaunchPersistentContext(profile_path, options)
	}

	if err != nil {
		return err
	}
	defer browser.Close()

	page, err := browser.NewPage()
	if err != nil {
		return err
	}
	defer page.Close()

	page.AddInitScript(playwright.Script{
		Content: playwright.String(code),
	})

	if _, err := page.Goto(url); err != nil {
		return err
	}
	if err := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	}); err != nil {
		return err
	}

	page.OnDialog(func(d playwright.Dialog) {
		message := d.Message()
		fmt.Println("Dialog:", message)
		if strings.HasPrefix(message, "SAGE#") {
			chout <- message[5:]
		}
		d.Dismiss()
	})

	page.OnClose(func(_ playwright.Page) {
		fmt.Println("Page closed")
		chint <- "quit"
	})

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
