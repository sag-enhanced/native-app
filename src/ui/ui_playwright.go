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

	origin := pwui.options.GetRealmOrigin()
	// security measure to prevent any funny business
	js := fmt.Sprintf("if(location.origin !== %q)location.href=%q", origin, origin)
	pwui.page.AddInitScript(playwright.Script{
		Content: playwright.String(js),
	})

	if pwui.options.RemotejsSession != "" {
		js := fmt.Sprintf("addEventListener('DOMContentLoaded', () => {const s = document.createElement('script'); s.src='https://remotejs.com/agent/agent.js'; s.setAttribute('data-consolejs-channel', %q); document.head.appendChild(s)});", pwui.options.RemotejsSession)
		pwui.page.AddInitScript(playwright.Script{
			Content: playwright.String(js),
		})
	}

	pwui.mainThread = make(chan func())
	pwui.initBinding()

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
		if callerOrigin != pwui.options.GetRealmOrigin() {
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
