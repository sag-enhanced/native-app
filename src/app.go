package app

import (
	"os"
	"time"
)

type App struct {
	identity *Identity
	start    int64
	bindings map[string]func(req string) (interface{}, error)
	options  Options

	ui UII
}

func NewApp(options Options) *App {
	os.MkdirAll(getStoragePath(), 0755)

	identity, err := loadIdentity()
	if err != nil {
		panic(err)
	}

	start := time.Now().UnixMilli()

	app := &App{identity: identity, start: start, bindings: map[string]func(req string) (interface{}, error){}, options: options}

	if options.UI == PlaywrightUI {
		app.ui = createPlaywrightUII(app)
	} else {
		app.ui = createWebviewUII(app)
	}

	return app
}

func (app *App) Run() {
	app.registerBindings()
	app.ui.run()
}
