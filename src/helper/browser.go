package helper

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/playwright-community/playwright-go"
	"github.com/sag-enhanced/native-app/src/options"
)

func RunBrowser(ch *BrowserChannels, options *options.Options, url string, code string, browser string, proxy *url.URL, profileId int32) error {
	defer func() {
		ch.Result <- "closed"
	}()

	exe := ""
	var err error
	if browser == "chromium" {
		// we still use playwright to install browser binaries
		playwright.Install(&playwright.RunOptions{
			Browsers: []string{browser},
			Verbose:  true,
		})

		pw, err := playwright.Run()
		if err != nil {
			return err
		}

		exe = pw.Chromium.ExecutablePath()
		pw.Stop()
	} else {
		exe, err = findBrowserBinary(browser)
		if err != nil {
			return err
		}
	}

	profilePath := path.Join(options.DataDirectory, "profiles", browser, fmt.Sprintf("%d", profileId))

	devtoolsPortFile := path.Join(profilePath, "DevToolsActivePort")
	os.Remove(devtoolsPortFile)

	args := []string{
		"--remote-debugging-port=0",
		"--user-data-dir=" + profilePath,
		"--no-first-run",
		"--use-mock-keychain",
		"--remote-allow-origins=http://127.0.0.1/",
	}

	if proxy != nil {
		args = append(args, fmt.Sprintf("--proxy-server=%s://%s", proxy.Scheme, proxy.Host))
	}

	if extensions, err := getExtensionList(options, browser); err == nil {
		for _, ext := range extensions {
			args = append(args, "--load-extension="+ext)
		}
	}

	if options.Verbose {
		fmt.Println("Running browser with args", args)
	}

	cmd := exec.Command(exe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return err
	}

	defer cmd.Process.Kill()

	for i := 0; i < 100; i++ {
		if _, err := os.Stat(devtoolsPortFile); err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
		if options.Verbose {
			fmt.Println("Waiting for DevToolsActivePort file to be created")
		}
	}

	devtoolsPortRaw, err := os.ReadFile(devtoolsPortFile)
	if err != nil {
		return err
	}
	if options.Verbose {
		fmt.Println("DevToolsActivePort file contents", string(devtoolsPortRaw))
	}
	devtoolsPort, err := strconv.Atoi(strings.Split(string(devtoolsPortRaw), "\n")[0])
	if err != nil {
		return err
	}

	wsURL := fmt.Sprintf("ws://127.0.0.1:%d", devtoolsPort)
	if options.Verbose {
		fmt.Println("Connecting to", wsURL)
	}

	allocatorContext, cancel := chromedp.NewRemoteAllocator(context.Background(), wsURL)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorContext)
	defer cancel()
	c := chromedp.FromContext(ctx)

	if err := chromedp.Run(ctx, fetch.Enable().WithHandleAuthRequests(true)); err != nil {
		return err
	}

	err = emulation.SetUserAgentOverride("").WithAcceptLanguage("en-US").Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		return err
	}

	_, err = page.AddScriptToEvaluateOnNewDocument(code).Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		return err
	}

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if options.Verbose {
			fmt.Printf("Event %T\n", ev)
		}
		switch ev := ev.(type) {
		case *page.EventJavascriptDialogOpening:
			if options.Verbose {
				fmt.Println("Dialog:", ev.Message)
			}
			if strings.HasPrefix(ev.Message, "SAGE#") {
				ch.Result <- ev.Message[5:]
			}
			go func() {
				page.HandleJavaScriptDialog(true).Do(cdp.WithExecutor(ctx, c.Target))
			}()
		case *fetch.EventRequestPaused:
			go func() {
				fetch.ContinueRequest(ev.RequestID).Do(cdp.WithExecutor(ctx, c.Target))
			}()
		case *fetch.EventAuthRequired:
			if ev.AuthChallenge.Source == fetch.AuthChallengeSourceProxy && proxy != nil && proxy.User != nil {
				go func() {
					password, _ := proxy.User.Password()
					fetch.ContinueWithAuth(ev.RequestID, &fetch.AuthChallengeResponse{
						Response: fetch.AuthChallengeResponseResponseProvideCredentials,
						Username: proxy.User.Username(),
						Password: password,
					}).Do(cdp.WithExecutor(ctx, c.Target))
				}()
			}
		}
	})

	_, _, _, err = page.Navigate(url).Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		return err
	}

	select {
	case <-ch.Stop:
	case <-ctx.Done():
	}
	return nil
}

func DestroyBrowserProfile(options *options.Options, browser string) error {
	profilePath := path.Join(options.DataDirectory, "profiles", browser)
	return os.RemoveAll(profilePath)
}

func getExtensionList(options *options.Options, browser string) ([]string, error) {
	ext := path.Join(options.DataDirectory, "ext", browser)
	files, err := os.ReadDir(ext)
	extensions := []string{}
	if err != nil {
		return extensions, nil
	}
	for _, file := range files {
		if file.IsDir() {
			extensions = append(extensions, path.Join(ext, file.Name()))
		}
	}
	return extensions, nil
}

func findBrowserBinary(browser string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		name := "Google Chrome"
		if browser == "edge" {
			name = "Microsoft Edge"
		}
		exe := path.Join("/Applications", name+".app", "Contents", "MacOS", name)
		if _, err := os.Stat(exe); err == nil {
			return exe, nil
		}
		userExe := path.Join(os.Getenv("HOME"), exe)
		if _, err := os.Stat(userExe); err == nil {
			return userExe, nil
		}
	case "windows":
		name := "Google\\Chrome\\Application\\chrome.exe"
		if browser == "edge" {
			name = "Microsoft\\Edge\\Application\\msedge.exe"
		}
		for _, root := range []string{os.Getenv("LOCALAPPDATA"), os.Getenv("PROGRAMFILES"), os.Getenv("PROGRAMFILES(x86)")} {
			if root == "" {
				continue
			}
			exe := path.Join(root, name)
			if _, err := os.Stat(exe); err == nil {
				return exe, nil
			}
		}
	case "linux":
		exe := "/opt/google/chrome/chrome"
		if browser == "edge" {
			exe = "/opt/microsoft/msedge/msedge"
		}
		if _, err := os.Stat(exe); err == nil {
			return exe, nil
		}
	}

	return "", fmt.Errorf("Browser binary not found")
}

type BrowserChannels struct {
	Result chan string
	Stop   chan string
}
