package app

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/shirou/gopsutil/v3/process"
)

// looking where the steam executable is from currently running processes
// seemed like the most reliable way to find it on all platforms
func (app App) findSteamExecutable() (string, error) {
	storagePath := getStoragePath()
	cache := path.Join(storagePath, "steam_executable.txt")
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
