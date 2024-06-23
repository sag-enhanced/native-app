package steam

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/sag-enhanced/native-app/src/options"
)

func RunSteamWithArguments(options *options.Options, args ...string) error {
	var executable string
	var err error

	if runtime.GOOS == "linux" {
		// steam is a shell script on linux, so we need to run it with bash
		executable = "/bin/bash"
		args = append([]string{"-c", "steam"}, args...)
	} else {
		executable, err = FindSteamExecutable(options)
		if err != nil {
			return err
		}
	}
	if options.SteamDev {
		args = append(args, "-dev")
	}

	if options.Verbose {
		fmt.Println("Running Steam with arguments:", executable, args)
	}
	cmd := exec.Command(executable, args...)
	// steam is dying without having stdout attached
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if runtime.GOOS == "windows" {
		// so windows apparently doesnt have the bootstrapper and it will directly start steam
		// and if we ran .Run() it would block the process until steam is closed, which is not what we want
		cmd.Start()
	} else {
		// this will run the bootstrapper and then block until its done and starts the actual steam process
		cmd.Run()
	}
	return nil
}
