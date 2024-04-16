package options

type Realm = string

const (
	StableRealm Realm = "stable"
	BetaRealm   Realm = "beta"
	LocalRealm  Realm = "local"
)

func (options *Options) GetRealmOrigin() string {
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
