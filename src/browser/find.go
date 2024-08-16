package browser

import (
	"fmt"

	"github.com/playwright-community/playwright-go"
)

var BrowserNotFoundErr = fmt.Errorf("Browser binary not found")

func findBrowserBinary(browser string) (string, error) {
	if browser == "chromium" || browser == "firefox" {
		// we use playwright to manage our chromium installation

		playwright.Install(&playwright.RunOptions{
			Browsers: []string{browser},
			Verbose:  true,
		})

		pw, err := playwright.Run()
		if err != nil {
			return "", err
		}
		defer pw.Stop()

		if browser == "firefox" {
			return pw.Firefox.ExecutablePath(), nil
		}
		return pw.Chromium.ExecutablePath(), nil
	}

	return findBrowserBinarySystem(browser)
}
