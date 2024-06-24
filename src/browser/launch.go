package browser

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

func prepareArguments(profile string, proxy *url.URL) []string {
	args := []string{
		"--remote-debugging-port=0",
		"--user-data-dir=" + profile,
		"--no-first-run",
		"--use-mock-keychain", // required or macOS will prompt for keychain access
		"--remote-allow-origins=http://127.0.0.1/",
	}
	if proxy != nil {
		// we cant pass the authentication information to the browser here
		args = append(args, fmt.Sprintf("--proxy-server=%s://%s", proxy.Scheme, proxy.Host))
	}
	return args
}

func prepareExtensions(args []string, extensions []string) []string {
	for _, ext := range extensions {
		args = append(args, "--load-extension="+ext)
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

func waitForDevToolsActivePort(profile string) (int, error) {
	devtoolsPortFile := path.Join(profile, "DevToolsActivePort")
	for i := 0; i < 100; i++ {
		if _, err := os.Stat(devtoolsPortFile); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	content, err := os.ReadFile(devtoolsPortFile)
	if err != nil {
		return 0, err
	}

	port, err := strconv.Atoi(strings.Split(string(content), "\n")[0])
	if err != nil {
		return 0, err
	}
	return port, nil
}
