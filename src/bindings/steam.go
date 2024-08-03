package bindings

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/sag-enhanced/native-app/src/steam"
)

func (b *Bindings) SteamPatch(js string) error {
	exe, err := steam.FindSteamExecutable(b.options)
	if err != nil {
		return err
	}

	if b.options.Verbose {
		fmt.Println("Steam executable found at", exe)
	}

	data, err := steam.FindSteamDataDir(b.options)
	if err != nil {
		return err
	}
	if b.options.Verbose {
		fmt.Println("Steam data directory found at", data)
	}

	entryFile := path.Join(data, "steamui", "library.js")
	content, err := os.ReadFile(entryFile)
	if err != nil {
		return err
	}

	// inject our code into the steam client
	lines := strings.Split(string(content), "\n")
	if len(lines) > 2 {
		lines = lines[:3]
	}
	if js != "" {
		lines = append(lines, js)
	}

	return os.WriteFile(entryFile, []byte(strings.Join(lines, "\n")), 0644)
}

func (b *Bindings) SteamRun() error {
	steam.CloseSteam(b.options)

	if b.options.Verbose {
		fmt.Println("Starting Steam with injected code...")
	}
	// -noverifyfiles is required to prevent steam from checking the files
	// and redownloading them if they are modified
	return steam.RunSteamWithArguments(b.options, "-noverifyfiles")
}
