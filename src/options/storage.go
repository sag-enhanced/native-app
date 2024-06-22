package options

import (
	"os"
	"runtime"
)

func GetDefaultStoragePath() string {
	if runtime.GOOS == "windows" {
		return os.ExpandEnv("${APPDATA}/sage")
	} else if runtime.GOOS == "darwin" {
		return os.ExpandEnv("${HOME}/Library/Application Support/sage")
	}
	return os.ExpandEnv("${HOME}/.config/sage")
}
