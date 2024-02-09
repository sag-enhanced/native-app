package app

import (
	"fmt"
	"os"
	"time"

	"github.com/webview/webview_go"
)

type App struct {
	webview  webview.WebView
	identity *Identity
	start    int64
	bindings map[string]func(req string) (interface{}, error)
	options  Options
}

type Options struct {
	Debug       bool
	Verbose     bool
	Local       bool
	OpenCommand []string
}

func NewApp(options Options) *App {
	if options.Debug {
		fmt.Println("Storage location:", getStoragePath())
	}

	installPlaywright()

	webview := webview.New(options.Debug)

	os.MkdirAll(getStoragePath(), 0755)

	identity, err := loadIdentity()
	if err != nil {
		panic(err)
	}

	start := time.Now().UnixMilli()

	return &App{webview: webview, identity: identity, start: start, bindings: map[string]func(req string) (interface{}, error){}, options: options}
}

func (app *App) Run() {
	app.initBindings()
	app.registerBindings()
	app.runUI()
}
