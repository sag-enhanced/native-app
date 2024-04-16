package options

type Options struct {
	Build        uint32
	LoopbackPort uint16

	Verbose         bool
	Realm           Realm
	OpenCommand     []string
	RemotejsSession string
	UI              UI
	SteamDev        bool
}

func NewOptions() *Options {
	return &Options{
		Build:        7,
		LoopbackPort: 8666,

		UI:          GetPreferredUI(),
		OpenCommand: GetDefaultOpenCommand(),
	}
}
