package ui

import (
	"fmt"
	"net/url"

	"github.com/playwright-community/playwright-go"
	"github.com/sag-enhanced/native-app/src/options"
)

type PlaywrightUII struct {
	page        playwright.Page
	mainThread  chan func()
	options     *options.Options
	bindHandler bindHandler
}

func createPlaywrightUII(options *options.Options) *PlaywrightUII {
	return &PlaywrightUII{options: options}
}

func (pwui *PlaywrightUII) Run() {
	options := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
		},
		IgnoreDefaultArgs: []string{
			// disables "Chrome is being controlled by automated test software" banner
			"--enable-automation",
		},
	}

	playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})

	pw, err := playwright.Run()

	if err != nil {
		fmt.Println("Error while starting playwright: ", err)
		return
	}
	defer pw.Stop()

	browser, err := pw.Chromium.Launch(options)
	if err != nil {
		fmt.Println("Error while launching browser: ", err)
		return
	}
	defer browser.Close()

	pwui.page, err = browser.NewPage(playwright.BrowserNewPageOptions{
		NoViewport: playwright.Bool(true),
	})
	if err != nil {
		fmt.Println("Error while creating new page: ", err)
		return
	}
	defer pwui.page.Close()

	pwui.mainThread = make(chan func())
	pwui.initBinding()

	scripts := getScripts(pwui.options)
	for _, script := range scripts {
		pwui.page.AddInitScript(playwright.Script{
			Content: playwright.String(script),
		})
	}

	pwui.page.OnClose(func(_ playwright.Page) {
		// this will wake-up the main thread which will then realize its time to exit
		pwui.mainThread <- func() {}
	})

	pwui.page.Goto(pwui.options.GetRealmOrigin())

	for !pwui.page.IsClosed() {
		select {
		case fn := <-pwui.mainThread:
			fn()
		}
	}
}

func (pwui *PlaywrightUII) Navigate(url string) {
	if pwui.options.Verbose {
		fmt.Println("Navigate:", url)
	}
	pwui.mainThread <- func() {
		pwui.page.Goto(url)
	}
}

func (pwui *PlaywrightUII) Eval(code string) {
	if pwui.options.Verbose {
		fmt.Println("Eval:", code)
	}
	pwui.mainThread <- func() {
		pwui.page.Evaluate(code)
	}
}

func (pwui *PlaywrightUII) Quit() {
	pwui.mainThread <- func() {
		pwui.page.Close()
	}
}

func (pwui *PlaywrightUII) initBinding() {
	pwui.page.ExposeBinding("sage", func(source *playwright.BindingSource, args ...any) any {
		if len(args) != 3 {
			return fmt.Errorf("sage() expects 3 arguments")
		}
		caller, err := url.Parse(source.Frame.URL())
		if err != nil {
			return fmt.Errorf("failed to parse caller URL: %w", err)
		}

		callerOrigin := fmt.Sprintf("%s://%s", caller.Scheme, caller.Host)
		if callerOrigin != pwui.options.GetRealmOrigin() && callerOrigin != "https://id.sage.party" {
			return fmt.Errorf("sage() is not allowed to be called from %q", callerOrigin)
		}

		method := args[0].(string)
		callId := args[1].(int)
		params := args[2].(string)

		return pwui.bindHandler(method, callId, params)
	})
}

func (pwui *PlaywrightUII) SetBindHandler(handler bindHandler) {
	pwui.bindHandler = handler
}
