package app

import (
	"os"

	"github.com/sag-enhanced/native-app/src/bindings"
	"github.com/sag-enhanced/native-app/src/file"
	"github.com/sag-enhanced/native-app/src/options"
	"github.com/sag-enhanced/native-app/src/ui"
)

func Run(options *options.Options) {
	os.MkdirAll(options.DataDirectory, 0755)

	fm, err := file.NewFileManager(options)
	if err != nil {
		panic(err)
	}
	ui := ui.NewUI(options)

	bindings := bindings.NewBindings(options, ui, fm)
	ui.SetBindHandler(bindings.BindHandler)
	ui.Run()
}
