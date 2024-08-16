package browser

import (
	"os"
	"path"
)

func findBrowserBinarySystem(browser string) (string, error) {
	name := "Google Chrome"
	switch browser {
	case "edge":
		name = "Microsoft Edge"
	case "firefox":
		name = "Firefox"
	}

	exe := path.Join("/Applications", name+".app", "Contents", "MacOS", name)
	if _, err := os.Stat(exe); err == nil {
		return exe, nil
	}
	userExe := path.Join(os.Getenv("HOME"), exe)
	if _, err := os.Stat(userExe); err == nil {
		return userExe, nil
	}

	return "", BrowserNotFoundErr
}
