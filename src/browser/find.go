package browser

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/playwright-community/playwright-go"
)

func findBrowserBinary(browser string) (string, error) {
	if browser == "chromium" {
		// we use playwright to manage our chromium installation

		playwright.Install(&playwright.RunOptions{
			Browsers: []string{browser},
			Verbose:  true,
		})

		pw, err := playwright.Run()
		if err != nil {
			return "", err
		}
		defer pw.Stop()

		return pw.Chromium.ExecutablePath(), nil
	}

	switch runtime.GOOS {
	case "darwin":
		name := "Google Chrome"
		if browser == "edge" {
			name = "Microsoft Edge"
		}
		exe := path.Join("/Applications", name+".app", "Contents", "MacOS", name)
		if _, err := os.Stat(exe); err == nil {
			return exe, nil
		}
		userExe := path.Join(os.Getenv("HOME"), exe)
		if _, err := os.Stat(userExe); err == nil {
			return userExe, nil
		}
	case "windows":
		name := "Google\\Chrome\\Application\\chrome.exe"
		if browser == "edge" {
			name = "Microsoft\\Edge\\Application\\msedge.exe"
		}
		for _, root := range []string{os.Getenv("LOCALAPPDATA"), os.Getenv("PROGRAMFILES"), os.Getenv("PROGRAMFILES(x86)")} {
			if root == "" {
				continue
			}
			exe := path.Join(root, name)
			if _, err := os.Stat(exe); err == nil {
				return exe, nil
			}
		}
	case "linux":
		exe := "/opt/google/chrome/chrome"
		if browser == "edge" {
			exe = "/opt/microsoft/msedge/msedge"
		}
		if _, err := os.Stat(exe); err == nil {
			return exe, nil
		}
	}

	return "", fmt.Errorf("Browser binary not found")
}
