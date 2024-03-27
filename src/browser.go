package app

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/playwright-community/playwright-go"
)

func (app *App) runBrowser(chResult chan string, chStop chan string, url string, code string, proxy *string) error {
	defer func() {
		chResult <- "closed"
	}()

	// we still use playwright to install browser binaries
	playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
		Verbose:  true,
	})

	pw, err := playwright.Run()
	if err != nil {
		return err
	}

	exe := pw.Chromium.ExecutablePath()
	pw.Stop()

	profileName := "manual-profile"
	profilePath := path.Join(getStoragePath(), profileName)

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
		args = append(args, "--proxy-server="+*proxy)
	}

	if extensions, err := getExtensionList(); err == nil {
		for _, ext := range extensions {
			args = append(args, "--load-extension="+ext)
		}
	}

	if app.options.Verbose {
		fmt.Println("Running chromedriver with args", args)
	}

	cmd := exec.Command(exe, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		return err
	}

	defer cmd.Process.Kill()

	for _, err := os.Stat(devtoolsPortFile); err != nil; _, err = os.Stat(devtoolsPortFile) {
		time.Sleep(100 * time.Millisecond)
		if app.options.Verbose {
			fmt.Println("Waiting for DevToolsActivePort file to be created")
		}
	}

	devtoolsPortRaw, err := os.ReadFile(devtoolsPortFile)
	if err != nil {
		return err
	}
	if app.options.Verbose {
		fmt.Println("DevToolsActivePort file contents", string(devtoolsPortRaw))
	}
	devtoolsPort, err := strconv.Atoi(strings.Split(string(devtoolsPortRaw), "\n")[0])
	if err != nil {
		return err
	}

	wsURL := fmt.Sprintf("ws://127.0.0.1:%d", devtoolsPort)
	if app.options.Verbose {
		fmt.Println("Connecting to", wsURL)
	}

	allocatorContext, cancel := chromedp.NewRemoteAllocator(context.Background(), wsURL)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocatorContext)
	defer cancel()
	c := chromedp.FromContext(ctx)

	if err := chromedp.Run(ctx); err != nil {
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
		if app.options.Verbose {
			fmt.Printf("Event %T\n", ev)
		}
		switch ev := ev.(type) {
		case *page.EventJavascriptDialogOpening:
			if app.options.Verbose {
				fmt.Println("Dialog:", ev.Message)
			}
			if strings.HasPrefix(ev.Message, "SAGE#") {
				chResult <- ev.Message[5:]
			}
			go func() {
				page.HandleJavaScriptDialog(true).Do(cdp.WithExecutor(ctx, c.Target))
			}()
		}
	})

	_, _, _, err = page.Navigate(url).Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		return err
	}

	select {
	case <-chStop:
	case <-ctx.Done():
	}
	return nil
}

func (app *App) destroyBrowserProfile() error {
	profileName := "manual-profile"
	profilePath := path.Join(getStoragePath(), profileName)

	return os.RemoveAll(profilePath)
}

func getExtensionList() ([]string, error) {
	ext := path.Join(getStoragePath(), "ext", "chromium")
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
