package app

import (
	"os/exec"
	"runtime"
)

func GetDefaultOpenCommand() string {
	if runtime.GOOS == "windows" {
		return "explorer"
	}
	return "open"
}

func (app App) open(url string) {
	exec.Command(app.options.OpenCommand, url).Run()
}
