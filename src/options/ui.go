package options

type UI = string

const (
	PlaywrightUI UI = "playwright"
	WebviewUI    UI = "webview"
)

func GetPreferredUI() UI {
	if isWebviewAvailable() {
		return WebviewUI
	}
	return PlaywrightUI
}
