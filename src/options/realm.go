package options

type Realm = string

const (
	StableRealm Realm = "stable"
	BetaRealm   Realm = "beta"
	DevRealm    Realm = "dev"
	LocalRealm  Realm = "local"
)

func (options *Options) GetRealmOrigin() string {
	switch options.Realm {
	case StableRealm:
		return "https://app.sage.party"
	case BetaRealm:
		return "https://app-beta.sage.party"
	case DevRealm:
		return "https://app-dev.sage.party"
	case LocalRealm:
		return "http://localhost:5173"
	}
	return "https://" + options.Realm
}
