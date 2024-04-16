package options

import "runtime"

type UI = string

const (
	PlaywrightUI UI = "playwright"
	WebviewUI    UI = "webview"
)

func GetPreferredUI() UI {
	if runtime.GOOS == "linux" {
		return PlaywrightUI
	}
	return WebviewUI
}
