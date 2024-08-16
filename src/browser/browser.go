package browser

import (
	"context"
	"fmt"
	"net/url"
	"path"

	"github.com/sag-enhanced/native-app/src/options"
)

func RunBrowser(stop context.Context, options *options.Options, browserUrl string, browser string, proxy *url.URL, profileId int32) error {
	var err error

	profile := path.Join(options.DataDirectory, "profiles", browser, fmt.Sprint(profileId))

	args := prepareArguments(profile, proxy)
	if extensions, err := getExtensionList(options, browser); err == nil {
		args = prepareExtensions(args, extensions)
	}
	args = append(args, browserUrl)

	exe := options.ForceBrowser
	if exe == "" {
		var err error
		exe, err = findBrowserBinary(browser)
		if err != nil {
			return err
		}
	}

	if options.Verbose {
		fmt.Println("Running browser with args", exe, args)
	}

	proc, err := launchBrowser(exe, args)
	if err != nil {
		return err
	}
	defer proc.Kill()

	processDone := make(chan struct{})
	go func() {
		proc.Wait()
		close(processDone)
	}()

	select {
	case <-stop.Done():
	case <-processDone:
	}

	return nil
}
