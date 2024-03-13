package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/sag-enhanced/native-app/src"
)

func main() {
	var options app.Options
	var openCommand string
	flag.StringVar(&options.RemotejsSession, "remote", "", "Allow remote debugging with the specified session ID.")
	flag.StringVar(&options.Realm, "realm", "stable", "Run the app in the specified realm")
	flag.BoolVar(&options.Verbose, "verbose", false, "Enable VERY verbose logging")
	flag.StringVar(&openCommand, "open", "", "Command to open URLs")
	flag.Parse()

	if openCommand != "" {
		options.OpenCommand = strings.Split(openCommand, " ")
	} else {
		options.OpenCommand = app.GetDefaultOpenCommand()
	}

	if options.RemotejsSession != "" {
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

	app := app.NewApp(options)
	start := time.Now()

	app.Run()

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
