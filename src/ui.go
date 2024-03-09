package app

import (
	"fmt"

	"github.com/webview/webview_go"
)

func (app *App) runUI() {
	defer app.webview.Destroy()

	app.webview.SetTitle(fmt.Sprintf("SAG Enhanced (b%d)", build))
	app.webview.SetSize(800, 600, webview.HintNone)

	origin := app.options.getRealmOrigin()

	// security measure to prevent any funny business
	js := fmt.Sprintf("if(location.origin !== %q)location.href=%q", origin, origin)
	app.webview.Init(js)

	app.webview.Navigate(origin)

	app.webview.Run()
}
