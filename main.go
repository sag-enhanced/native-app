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
	var remoteSession string
	var options app.Options
	var openCommand string
	flag.StringVar(&remoteSession, "remote", "", "Allow remote debugging with the specified session ID.")
	flag.BoolVar(&options.Debug, "debug", false, "Enable devtools and verbose logging")
	flag.BoolVar(&options.Local, "local", false, "Run the app in local mode")
	flag.BoolVar(&options.Verbose, "verbose", false, "Enable VERY verbose logging")
	flag.StringVar(&openCommand, "open", "", "Command to open URLs")
	flag.Parse()

	if openCommand != "" {
		options.OpenCommand = strings.Split(openCommand, " ")
	} else {
		options.OpenCommand = app.GetDefaultOpenCommand()
	}

	if remoteSession != "" {
		var allow string
		fmt.Println("Debug session requested: " + remoteSession)
		fmt.Println("A debug session will allow others to connect to your app and debug it remotely.")
		fmt.Print("Allow debug session? (y/N): ")
		fmt.Scanln(&allow)
		if allow != "y" && allow != "Y" {
			fmt.Println("Aborted.")
			return
		}
	}
	if options.Debug {
		fmt.Println("Debug mode enabled.")
	}

	app := app.NewApp(options)
	if remoteSession != "" {
		app.InstallDebugger(remoteSession)
	}

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
