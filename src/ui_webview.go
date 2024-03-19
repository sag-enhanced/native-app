package app

import (
	"fmt"

	webview_go "github.com/webview/webview_go"
)

type WebviewUII struct {
	app     *App
	webview webview_go.WebView
}

func createWebviewUII(app *App) *WebviewUII {
	return &WebviewUII{
		app: app,
	}
}

func (wui *WebviewUII) run() {
	wui.webview = webview_go.New(true)
	defer wui.webview.Destroy()

	wui.webview.SetTitle(fmt.Sprintf("SAG Enhanced (b%d)", build))
	wui.webview.SetSize(800, 600, webview_go.HintNone)

	origin := wui.app.options.getRealmOrigin()
	// security measure to prevent any funny business
	js := fmt.Sprintf("if(location.origin !== %q)location.href=%q", origin, origin)
	wui.webview.Init(js)

	wui.webview.Bind("sage", wui.app.bindHandler)

	if wui.app.options.RemotejsSession != "" {
		js := fmt.Sprintf("addEventListener('DOMContentLoaded', () => {const s = document.createElement('script'); s.src='https://remotejs.com/agent/agent.js'; s.setAttribute('data-consolejs-channel', %q); document.head.appendChild(s)});", wui.app.options.RemotejsSession)
		wui.webview.Init(js)
	}

	wui.webview.Navigate(origin)
	wui.webview.Run()
}

func (wui *WebviewUII) eval(code string) {
	wui.webview.Dispatch(func() {
		wui.webview.Eval(code)
	})
}

func (wui *WebviewUII) quit() {
	wui.webview.Dispatch(func() {
		wui.webview.Terminate()
	})
}

//
// func (app *App) initWebviewBindings(webview webview_go.WebView) {
// 	webview.Bind("sage", func(method string, callId int, params string) error {
// 		return app.bindHandler(method, callId, params)
// 	})
// }
