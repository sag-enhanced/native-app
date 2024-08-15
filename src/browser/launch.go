package browser

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
)

func prepareArguments(profile string, proxy *url.URL) []string {
	args := []string{
		"--user-data-dir=" + profile,
		"--no-first-run",
		"--use-mock-keychain", // required or macOS will prompt for keychain access
	}
	if proxy != nil {
		// we cant pass the authentication information to the browser here
		args = append(args, fmt.Sprintf("--proxy-server=%s://%s", proxy.Scheme, proxy.Host))
	}
	return args
}

func prepareExtensions(args []string, extensions []string) []string {
	if len(extensions) > 0 {
		args = append(args, "--load-extension="+strings.Join(extensions, ","))
	}
	return args
}

func launchBrowser(exe string, args []string) (*os.Process, error) {
	cmd := exec.Command(exe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd.Process, nil
}
