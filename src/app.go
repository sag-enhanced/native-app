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
}

func NewApp(options Options) *App {
	os.MkdirAll(getStoragePath(), 0755)

	identity, err := loadIdentity()
	if err != nil {
		panic(err)
	}

	start := time.Now().UnixMilli()

	return &App{identity: identity, start: start, bindings: map[string]func(req string) (interface{}, error){}, options: options}
}

func (app *App) Run() {
	app.registerBindings()
	app.runWebview()
}
