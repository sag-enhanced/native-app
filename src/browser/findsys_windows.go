package browser

import (
	"os"
	"path"
)

func findBrowserBinarySystem(browser string) (string, error) {
	name := "Google\\Chrome\\Application\\chrome.exe"
	switch browser {
	case "edge":
		name = "Microsoft\\Edge\\Application\\msedge.exe"
	case "firefox":
		name = "Mozilla Firefox\\firefox.exe"
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
	return "", BrowserNotFoundErr
}
