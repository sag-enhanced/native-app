package bindings

import "net/url"

var currentUrl *url.URL

// webview has no builtin way to get the current url, so we neded a tamper proof way to access it
// we use a secret that is only known to the app and the bindings
// this way we cant be tricked into setting the url to something else
//
// we need the current URL to properly restrict certain bindings to only work on certain pages
func (b *Bindings) SetUrl(currentPageUrl string, secret string) error {
	// someone is doing something fishy
	if b.options.CurrentUrlSecret != secret {
		b.ui.Quit()
		return nil
	}

	var err error
	currentUrl, err = url.Parse(currentPageUrl)
	return err
}
