package app

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

func (app *App) runPlaywrightUI() {
	args := []string{}
	options := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Args:     args,
		IgnoreDefaultArgs: []string{
			// disables "Chrome is being controlled by automated test software" banner
			"--enable-automation",
		},
	}

	pw, err := playwright.Run(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})

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

	page, err := browser.NewPage(playwright.BrowserNewPageOptions{
		NoViewport: playwright.Bool(true),
	})
	if err != nil {
		fmt.Println("Error while creating new page: ", err)
		return
	}
	defer page.Close()

	origin := app.options.getRealmOrigin()
	// security measure to prevent any funny business
	js := fmt.Sprintf("if(location.origin !== %q)location.href=%q", origin, origin)
	page.AddInitScript(playwright.Script{
		Content: playwright.String(js),
	})

	mainThread := make(chan func())
	app.initPlaywrightBindings(page, mainThread)

	page.Goto(app.options.getRealmOrigin())

	for !page.IsClosed() {
		select {
		case fn := <-mainThread:
			fn()
		}
	}
}

func (app *App) initPlaywrightBindings(page playwright.Page, mainThread chan func()) {
	eval := func(js string) {
		mainThread <- func() {
			page.Evaluate(js)
		}
	}
	page.ExposeBinding("sage", func(source *playwright.BindingSource, args ...any) any {
		if len(args) != 3 {
			return fmt.Errorf("sage() expects 3 arguments")
		}
		// TODO: add an origin check to source

		method := args[0].(string)
		callId := args[1].(int)
		params := args[2].(string)

		return app.bindHandler(method, callId, params, eval)
	})

	page.OnClose(func(_ playwright.Page) {
		// this will wake-up the main thread which will then realize its time to exit
		mainThread <- func() {}
	})

	app.bind("quit", func() {
		mainThread <- func() {
			page.Close()
		}
	})
}
