//go:build !windows && !linux

package options

func isWebviewAvailable() bool {
	return true
}
