package app

type Realm = string

const (
	StableRealm Realm = "stable"
	BetaRealm   Realm = "beta"
	LocalRealm  Realm = "local"
)

type Options struct {
	Verbose         bool
	Realm           Realm
	OpenCommand     []string
	RemotejsSession string
	PlaywrightUI    bool
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
