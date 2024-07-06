package bindings

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/sag-enhanced/native-app/src/helper"
)

func (b *Bindings) Open(target string) error {
	// some sanization checks
	if strings.ContainsAny(target, "\n\r'\"` {}$|;") {
		return fmt.Errorf("Invalid URL")
	}
	url, err := url.Parse(target)
	// only allow https urls and block any path traversal attempts
	if err != nil || url.Scheme != "https" || strings.Contains(url.Path, "..") {
		return fmt.Errorf("Invalid URL")
	}
	fmt.Println("Opening URL", url.String())
	// re-assemble url to string to avoid any funny business
	helper.Open(url.String(), b.options)
  return nil
}
