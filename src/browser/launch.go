package browser

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
)

func prepareChromiumArguments(profile string, proxy *url.URL) []string {
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

func prepareFirefoxArguments(profile string, proxy *url.URL) ([]string, error) {
	args := []string{
		"-profile", profile,
		"-no-remote",
		"-new-instance",
	}
	prefs := []string{
		"user_pref(\"browser.shell.checkDefaultBrowser\", false);",
		"user_pref(\"browser.shell.defaultBrowserCheckCount\", 1);",
		"user_pref(\"datareporting.policy.firstRunURL\", \"\");",
		"user_pref(\"browser.warnOnQuit\", false);",
		"user_pref(\"browser.warnOnQuitShortcut\", false);",
		"user_pref(\"browser.tabs.warnOnClose\", false);",
		"user_pref(\"browser.tabs.warnOnCloseOtherTabs\", false);",
		"user_pref(\"browser.sessionstore.restore_on_demand\", false);",
		"user_pref(\"xpinstall.signatures.required\", false);",
		"user_pref(\"xpinstall.whitelist.required\", false);",
		"user_pref(\"extensions.update.enabled\", false);",
		"user_pref(\"extensions.autoDisableScopes\", 10);",
		"user_pref(\"extensions.enabledScopes\", 5);",
		"user_pref(\"extensions.installDistroAddons\", false);",
		"user_pref(\"datareporting.policy.dataSubmissionEnabled\", false);",
		"user_pref(\"app.update.enabled\", false);",
	}
	if proxy != nil {
		prefs = append(prefs, "user_pref(\"network.proxy.type\", 1);")
		prefs = append(prefs, fmt.Sprintf("user_pref(\"network.proxy.%s\", %q);", proxy.Scheme, proxy.Hostname()))
		prefs = append(prefs, fmt.Sprintf("user_pref(\"network.proxy.%s_port\", %s);", proxy.Scheme, proxy.Port()))
	}
	if _, err := os.Stat(profile); os.IsNotExist(err) {
		os.MkdirAll(profile, 0700)
	}
	prefFile := path.Join(profile, "user.js")

	if err := os.WriteFile(prefFile, []byte(strings.Join(prefs, "\n")), 0600); err != nil {
		return nil, err
	}
	return args, nil
}

func prepareChromiumExtensions(args []string, extensions []string) []string {
	if len(extensions) > 0 {
		args = append(args, "--load-extension="+strings.Join(extensions, ","))
	}
	return args
}

func prepareFirefoxExtensions(profile string, extensions []string) error {
	extensionsDir := path.Join(profile, "extensions")
	if _, err := os.Stat(extensionsDir); !os.IsNotExist(err) {
		os.RemoveAll(extensionsDir)
	}
	os.MkdirAll(extensionsDir, 0700)

	for _, ext := range extensions {
		fmt.Println("Adding extension", ext)
		manifest := path.Join(ext, "manifest.json")
		content, err := os.ReadFile(manifest)
		if err != nil {
			return err
		}
		var m Manifest
		if err := json.Unmarshal(content, &m); err != nil {
			return err
		}
		id := m.BrowserSpecificSettings.Gecko.Id
		if id == "" {
			return fmt.Errorf("Manifest %s is missing browser_specific_settings.gecko.id", manifest)
		}

		proxy := path.Join(extensionsDir, id)
		os.WriteFile(proxy, []byte(ext), 0600)
	}

	return nil
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

type Manifest struct {
	BrowserSpecificSettings struct {
		Gecko struct {
			Id string `json:"id"`
		} `json:"gecko"`
	} `json:"browser_specific_settings"`
}
