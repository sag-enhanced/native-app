package app

import (
	"fmt"

	webview_go "github.com/webview/webview_go"
)

func (app *App) runWebview() {
	webview := webview_go.New(true)
	defer webview.Destroy()

	app.initWebviewBindings(webview)

	webview.SetTitle(fmt.Sprintf("SAG Enhanced (b%d)", build))
	webview.SetSize(800, 600, webview_go.HintNone)

	origin := app.options.getRealmOrigin()
	// security measure to prevent any funny business
	js := fmt.Sprintf("if(location.origin !== %q)location.href=%q", origin, origin)
	webview.Init(js)

	if app.options.RemotejsSession != "" {
		js := fmt.Sprintf("addEventListener('DOMContentLoaded', () => {const s = document.createElement('script'); s.src='https://remotejs.com/agent/agent.js'; s.setAttribute('data-consolejs-channel', %q); document.head.appendChild(s)});", app.options.RemotejsSession)
		webview.Init(js)
	}

	webview.Navigate(origin)
	webview.Run()
}

func (app *App) initWebviewBindings(webview webview_go.WebView) {
	eval := func(code string) {
		webview.Dispatch(func() {
			webview.Eval(code)
		})
	}
	webview.Bind("sage", func(method string, callId int, params string) error {
		return app.bindHandler(method, callId, params, eval)
	})

	// special bindings
	app.bind("quit", func() {
		webview.Dispatch(webview.Terminate)
	})
}
