//go:build linux

package options

import "os"

// sage requires webkit2gtk-4.1-dev to be installed, otherwise it'll segfault when trying to use webview
func isWebviewAvailable() bool {
	if _, err := os.Stat("/usr/include/webkit2gtk-4.1"); err == nil {
		return true
	}
	return false
}
