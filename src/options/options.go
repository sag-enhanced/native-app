package options

import (
	"crypto/rand"
	"encoding/base64"
)

type Options struct {
	Build        uint32
	Release      uint32
	LoopbackPort uint16

	Verbose       bool
	Realm         Realm
	OpenCommand   []string
	UI            UI
	SteamDev      bool
	DataDirectory string
	NoCompress    bool

	CurrentUrlSecret string
}

func NewOptions() *Options {
	secret := make([]byte, 32)
	rand.Read(secret) // let's pray this doesn't fail

	return &Options{
		Build:        12,
		Release:      3,
		LoopbackPort: 8666,

		UI:            GetPreferredUI(),
		OpenCommand:   GetDefaultOpenCommand(),
		DataDirectory: GetDefaultStoragePath(),

		CurrentUrlSecret: base64.RawURLEncoding.EncodeToString(secret),
	}
}
