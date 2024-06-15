package helper

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/sag-enhanced/native-app/src/options"
	"github.com/shirou/gopsutil/v3/process"
)

// looking where the steam executable is from currently running processes
// seemed like the most reliable way to find it on all platforms
func FindSteamExecutable(options *options.Options) (string, error) {
	storagePath := GetStoragePath()
	cache := filepath.Join(storagePath, "steam_executable.txt")
	if _, err := os.Stat(cache); err == nil {
		data, err := os.ReadFile(cache)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	fmt.Println("Searching for Steam executable...")
	process, err := findSteamProcess()
	if err != nil {
		fmt.Println("Steam process not found; starting it...")
		// we are opening steam now
		// opening the console just for why not
		// the code that is calling this will close steam immediately afterwards anyway
		Open("steam://open/console", options)
		for {
			if process, err = findSteamProcess(); err != nil {
				break
			}
			fmt.Println("Waiting for Steam process...")
			time.Sleep(1 * time.Second)
		}
	}
	fmt.Println("Steam process found: ", process.Pid)
	exe, err := process.Exe()
	if err != nil {
		return "", err
	}

	os.WriteFile(cache, []byte(exe), 0644)
	return exe, nil
}

func FindSteamDataDir(options *options.Options) (string, error) {
	if runtime.GOOS == "darwin" {
		// the application in /Applications is just the bootstrapper, the real executable
		// is installed per user right here:
		return filepath.Join(os.Getenv("HOME"), "Library/Application Support/Steam/Steam.AppBundle/Steam/Contents/MacOS"), nil
	}
	executable, err := FindSteamExecutable(options)
	if err != nil {
		return "", err
	}
	parent := filepath.Dir(executable)
	for len(parent) > 1 {
		_, err := os.Stat(filepath.Join(parent, "steamui"))
		if err == nil || !os.IsNotExist(err) {
			break
		}
		parent = filepath.Dir(parent)
	}
	if len(parent) <= 1 {
		return "", errors.New("Steam data directory not found")
	}
	return parent, nil
}

func findSteamProcess() (*process.Process, error) {
	processList, err := process.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range processList {
		name, err := p.Name()
		if err != nil {
			continue
		}
		if name == "steam_osx" || name == "steam.exe" || name == "steam" {
			return p, nil
		}
	}
	return nil, errors.New("Steam process not found")
}

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

func CloseSteam(options *options.Options) error {
	killed := int32(0)
	for {
		var process *process.Process
		var err error
		if process, err = findSteamProcess(); err != nil {
			break
		}
		if process.Pid != killed {
			// new process found (this can happen if we close steam while its still bootstrapping)
			if options.Verbose {
				fmt.Println("Steam running, shutting it down...")
			}

			RunSteamWithArguments(options, "-shutdown")
			killed = process.Pid
		}
		if options.Verbose {
			fmt.Println("Waiting for Steam to shut down...", process.Pid)
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
