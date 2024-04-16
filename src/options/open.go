package options

import "runtime"

func GetDefaultOpenCommand() []string {
	if runtime.GOOS == "windows" {
		return []string{"rundll32", "url.dll,FileProtocolHandler"}
	} else if runtime.GOOS == "darwin" {
		return []string{"open"}
	}
	return []string{"xdg-open"}
}
