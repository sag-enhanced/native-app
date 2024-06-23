package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"

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
	flag.StringVar(&opt.RemotejsSession, "remote", "", "Allow remote debugging with the specified session ID.")
	flag.StringVar(&opt.Realm, "realm", options.StableRealm, "Run the app in the specified realm")
	flag.BoolVar(&opt.Verbose, "verbose", false, "Enable VERY verbose logging")
	flag.StringVar(&openCommand, "open", "", "Command to open URLs")
	flag.StringVar(&opt.UI, "ui", opt.UI, "UI to use (webview or playwright)")
	flag.BoolVar(&opt.SteamDev, "steamdev", false, "Enable Steam Dev mode")
	flag.BoolVar(&opt.NoCompress, "nocompress", false, "Disable file compression")
	flag.IntVar(&buildOverride, "build", -1, "Override/spoof build number (NOT RECOMMENDED)")
	flag.IntVar(&releaseOverride, "release", -1, "Override/spoof release number (NOT RECOMMENDED)")
	flag.IntVar(&loopbackPort, "loopback", -1, fmt.Sprintf("Port to use for loopback connections (default: %d) (NOT RECOMMENDED)", opt.LoopbackPort))
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
		fmt.Println("WARNING: Loopback port override is not recommended and may cause issues.")
		opt.LoopbackPort = uint16(loopbackPort)
	}
	if opt.Realm != options.StableRealm {
		fmt.Println("WARNING: Using experimental realm. This may cause issues.")
	}

	if opt.RemotejsSession != "" {
		var allow string
		fmt.Println("Debug session requested using -remote flag")
		fmt.Println("A debug session will allow others to connect to your app and debug it remotely. Please make sure you are communicating with official staff.")
		fmt.Print("Allow debug session? (y/N): ")
		fmt.Scanln(&allow)
		if allow != "y" && allow != "Y" {
			fmt.Println("Aborted.")
			return
		}
	}

	start := time.Now()

	app.Run(opt)

	elapsed := time.Since(start)
	if elapsed.Seconds() < 2 {
		fmt.Println("App exited too quickly. Everything ok?")
		if runtime.GOOS == "windows" {
			fmt.Println("Windows 10 users need to install the following program:")
			fmt.Println("https://developer.microsoft.com/en-us/microsoft-edge/webview2/")
		}
		fmt.Scanln()
	}
}
