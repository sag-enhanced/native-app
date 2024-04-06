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
	SteamDev        bool
}

func (options *Options) getRealmOrigin() string {
	switch options.Realm {
	case BetaRealm:
		return "https://app-beta.sage.party"
	case LocalRealm:
		return "http://localhost:5173"
	case StableRealm:
		return "https://app.sage.party"
	}
	return "https://" + options.Realm
}

func GetPreferredUI() UI {
	if runtime.GOOS == "linux" {
		return PlaywrightUI
	}
	return WebviewUI
}
