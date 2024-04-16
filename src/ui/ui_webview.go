package ui

import (
	"fmt"

	"github.com/sag-enhanced/native-app/src/options"
	webview_go "github.com/webview/webview_go"
)

type WebviewUII struct {
	webview     webview_go.WebView
	options     *options.Options
	bindHandler bindHandler
}

func createWebviewUII(options *options.Options) *WebviewUII {
	return &WebviewUII{
		options: options,
	}
}

func (wui *WebviewUII) Run() {
	wui.webview = webview_go.New(true)
	defer wui.webview.Destroy()

	wui.webview.SetTitle(fmt.Sprintf("SAG Enhanced (b%d)", wui.options.Build))
	wui.webview.SetSize(800, 600, webview_go.HintNone)

	origin := wui.options.GetRealmOrigin()
	// security measure to prevent any funny business
	js := fmt.Sprintf("if(location.origin !== %q)location.href=%q", origin, origin)
	wui.webview.Init(js)

	wui.webview.Bind("sage", wui.bindHandler)

	if wui.options.RemotejsSession != "" {
		js := fmt.Sprintf("addEventListener('DOMContentLoaded', () => {const s = document.createElement('script'); s.src='https://remotejs.com/agent/agent.js'; s.setAttribute('data-consolejs-channel', %q); document.head.appendChild(s)});", wui.options.RemotejsSession)
		wui.webview.Init(js)
	}

	wui.webview.Navigate(origin)
	wui.webview.Run()
}

func (wui *WebviewUII) Eval(code string) {
	if wui.options.Verbose {
		fmt.Println("Eval:", code)
	}
	// there seems to be a rare bug in webview where sometimes the eval doesn't work
	// so we try it a few times (the code is idempotent so it's safe to retry)
	// 3 tries should be enough
	// for i := 0; i < 3; i++ {
	wui.webview.Dispatch(func() {
		wui.webview.Eval(code)
	})
	// }
}

func (wui *WebviewUII) Quit() {
	wui.webview.Dispatch(func() {
		wui.webview.Terminate()
	})
}

func (wui *WebviewUII) SetBindHandler(handler bindHandler) {
	wui.bindHandler = handler
}
