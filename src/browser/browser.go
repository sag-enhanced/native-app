package browser

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/sag-enhanced/native-app/src/options"
)

func RunBrowser(ch *BrowserChannels, options *options.Options, url string, code string, browser string, proxy *url.URL, profileId int32) error {
	profile := path.Join(options.DataDirectory, "profiles", browser, fmt.Sprint(profileId))
	args := prepareArguments(profile, proxy)
	if extensions, err := getExtensionList(options, browser); err == nil {
		args = prepareExtensions(args, extensions)
	}

	exe, err := findBrowserBinary(browser)
	if err != nil {
		return err
	}

	if options.Verbose {
		fmt.Println("Running browser with args", exe, args)
	}

	proc, err := launchBrowser(exe, args)
	if err != nil {
		return err
	}
	defer proc.Kill()

	devtoolsPort, err := waitForDevToolsActivePort(profile)
	if err != nil {
		return err
	}

	wsURL := fmt.Sprintf("ws://127.0.0.1:%d", devtoolsPort)
	if options.Verbose {
		fmt.Println("Connecting to devtools on", wsURL)
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
			// we got a captcha token
			if strings.HasPrefix(ev.Message, "SAGE#") {
				ch.Result <- ev.Message[5:]
			}
			// recaptcha also does alerts if it fails to connect to its servers
			// so we just dismiss all of them
			go func() {
				page.HandleJavaScriptDialog(true).Do(cdp.WithExecutor(ctx, c.Target))
			}()
		case *fetch.EventRequestPaused:
			// we dont do anything with these, but we get them because we need to handle auth requests
			go func() {
				fetch.ContinueRequest(ev.RequestID).Do(cdp.WithExecutor(ctx, c.Target))
			}()
		case *fetch.EventAuthRequired:
			// proxy authentication!
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
	case <-ctx.Done(): // browser closed
	}
	return nil
}

type BrowserChannels struct {
	Result chan string
	Stop   chan string
}
