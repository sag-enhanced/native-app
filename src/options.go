package app

import "runtime"

type Realm = string

const (
	StableRealm Realm = "stable"
	BetaRealm   Realm = "beta"
	LocalRealm  Realm = "local"
)

type UI = string

const (
	PlaywrightUI UI = "playwright"
	WebviewUI    UI = "webview"
)

type Options struct {
	Verbose         bool
	Realm           Realm
	OpenCommand     []string
	RemotejsSession string
	UI              UI
}

func (options *Options) getRealmOrigin() string {
	switch options.Realm {
	case BetaRealm:
		return "https://app-beta.sage.party"
	case LocalRealm:
		return "http://localhost:5173"
	default:
		return "https://app.sage.party"
	}
}

func GetPreferredUI() UI {
	if runtime.GOOS == "linux" {
		return PlaywrightUI
	}
	return WebviewUI
}
