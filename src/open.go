package app

import (
	"os/exec"
	"runtime"
)

func GetDefaultOpenCommand() []string {
	if runtime.GOOS == "windows" {
		return []string{"rundll32", "url.dll,FileProtocolHandler"}
	} else if runtime.GOOS == "darwin" {
		return []string{"open"}
	}
	return []string{"xdg-open"}
}

func (app App) open(url string) {
	args := append(app.options.OpenCommand, url)
	exec.Command(args[0], args[1:]...).Run()
}
