//go:build !windows && !darwin

package browser

import (
	"os"
	"os/exec"
)

func findBrowserBinarySystem(browser string) (string, error) {
	if browser == "firefox" {
		exe, err := exec.LookPath("firefox")
		if err == nil {
			return exe, nil
		}
		return "", BrowserNotFoundErr
	}

	exe := "/opt/google/chrome/chrome"
	if browser == "edge" {
		exe = "/opt/microsoft/msedge/msedge"
	}
	if _, err := os.Stat(exe); err == nil {
		return exe, nil
	}

	return "", BrowserNotFoundErr
}
