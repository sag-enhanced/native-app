package bindings

import "net/url"

// webview has no builtin way to get the current url, so we neded a tamper proof way to access it
// we use a secret that is only known to the app and the bindings
// this way we cant be tricked into setting the url to something else
//
// we need the current URL to properly restrict certain bindings to only work on certain pages
func (b *Bindings) SetUrl(currentUrl string, secret string) error {
	// someone is doing something fishy
	if b.options.CurrentUrlSecret != secret {
		b.ui.Quit()
		return nil
	}

	url, err := url.Parse(currentUrl)
	b.currentUrl = url // we don't care if it's nil
	return err
}
