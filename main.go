package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/sag-enhanced/native-app/src"
	"github.com/sag-enhanced/native-app/src/options"
)

func main() {
	opt := options.NewOptions()
	var openCommand string
	var buildOverride int
	var releaseOverride int
	var loopbackPort int
	flag.StringVar(&opt.DataDirectory, "data", opt.DataDirectory, "Data directory to use")
	flag.StringVar(&opt.Realm, "realm", options.StableRealm, "Run the app in the specified realm")
	flag.BoolVar(&opt.Verbose, "verbose", false, "Enable VERY verbose logging")
	flag.StringVar(&openCommand, "open", "", "Command to open URLs")
	flag.StringVar(&opt.UI, "ui", opt.UI, "UI to use (webview or playwright)")
	flag.BoolVar(&opt.SteamDev, "steamdev", false, "Enable Steam Dev mode")
	flag.BoolVar(&opt.NoCompress, "nocompress", false, "Disable file compression")
	flag.IntVar(&buildOverride, "build", -1, "Override/spoof build number (NOT RECOMMENDED)")
	flag.IntVar(&releaseOverride, "release", -1, "Override/spoof release number (NOT RECOMMENDED)")
	flag.IntVar(&loopbackPort, "loopback", -1, fmt.Sprintf("Port to use for loopback connections (default: %d) (NOT RECOMMENDED)", opt.LoopbackPort))
	flag.StringVar(&opt.ForceBrowser, "forcebrowser", "", "Force a specific browser to be used (specify full executable path)")
	flag.Parse()

	if openCommand != "" {
		opt.OpenCommand = strings.Split(openCommand, " ")
	}
	if buildOverride != -1 {
		fmt.Println("WARNING: Build number override is not recommended and may cause issues.")
		opt.Build = uint32(buildOverride)
	}
	if releaseOverride != -1 {
		fmt.Println("WARNING: Release number override is not recommended and may cause issues.")
		opt.Release = uint32(releaseOverride)
	}
	if loopbackPort != -1 {
		opt.LoopbackPort = uint16(loopbackPort)
	}
	if opt.Realm != options.StableRealm {
		fmt.Println("WARNING: Using experimental realm. This may cause issues.")
	}

	if err := app.Run(opt); err != nil {
		fmt.Println(err)
	}
}
