package app

import (
	"fmt"
	"net/url"

	"github.com/playwright-community/playwright-go"
)

type PlaywrightUII struct {
	app        *App
	page       playwright.Page
	mainThread chan func()
}

func createPlaywrightUII(app *App) *PlaywrightUII {
	return &PlaywrightUII{
		app: app,
	}
}

func (pwui *PlaywrightUII) run() {
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

	origin := pwui.app.options.getRealmOrigin()
	// security measure to prevent any funny business
	js := fmt.Sprintf("if(location.origin !== %q)location.href=%q", origin, origin)
	pwui.page.AddInitScript(playwright.Script{
		Content: playwright.String(js),
	})

	pwui.mainThread = make(chan func())
	pwui.initBinding()

	pwui.page.OnClose(func(_ playwright.Page) {
		// this will wake-up the main thread which will then realize its time to exit
		pwui.mainThread <- func() {}
	})

	pwui.page.Goto(pwui.app.options.getRealmOrigin())

	for !pwui.page.IsClosed() {
		select {
		case fn := <-pwui.mainThread:
			fn()
		}
	}
}

func (pwui *PlaywrightUII) eval(code string) {
	if pwui.app.options.Verbose {
		fmt.Println("Eval:", code)
	}
	pwui.mainThread <- func() {
		pwui.page.Evaluate(code)
	}
}

func (pwui *PlaywrightUII) quit() {
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
		if callerOrigin != pwui.app.options.getRealmOrigin() {
			return fmt.Errorf("sage() is not allowed to be called from %q", callerOrigin)
		}

		method := args[0].(string)
		callId := args[1].(int)
		params := args[2].(string)

		return pwui.app.bindHandler(method, callId, params)
	})
}
