package app

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// looking where the steam executable is from currently running processes
// seemed like the most reliable way to find it on all platforms
func (app App) findSteamExecutable() (string, error) {
	storagePath := getStoragePath()
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
		app.open("steam://open/console")
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

func (app *App) findSteamDataDir() (string, error) {
	if runtime.GOOS == "darwin" {
		// the application in /Applications is just the bootstrapper, the real executable
		// is installed per user right here:
		return filepath.Join(os.Getenv("HOME"), "Library/Application Support/Steam/Steam.AppBundle/Steam/Contents/MacOS"), nil
	}
	executable, err := app.findSteamExecutable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(executable), nil
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

func (app *App) runSteamWithArguments(args ...string) error {
	executable, err := app.findSteamExecutable()
	if err != nil {
		return err
	}
	if app.options.Verbose {
		fmt.Println("Running Steam with arguments:", executable, args)
	}
	cmd := exec.Command(executable, args...)
	// steam is dying without having stdout attached
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	return nil
}

func (app *App) closeSteam() error {
	killed := int32(0)
	for {
		var process *process.Process
		var err error
		if process, err = findSteamProcess(); err != nil {
			break
		}
		if process.Pid != killed {
			// new process found (this can happen if we close steam while its still bootstrapping)
			if app.options.Verbose {
				fmt.Println("Steam running, shutting it down...")
			}

			app.runSteamWithArguments("-shutdown")
			killed = process.Pid
		}
		if app.options.Verbose {
			fmt.Println("Waiting for Steam to shut down...", process.Pid)
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}
