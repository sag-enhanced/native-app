package helper

import (
	"os/exec"

	"github.com/sag-enhanced/native-app/src/options"
)

func Open(url string, options *options.Options) {
	args := append(options.OpenCommand, url)
	exec.Command(args[0], args[1:]...).Run()
}
