//go:build windows

package options

import "golang.org/x/sys/windows/registry"

func isWebviewAvailable() bool {
	// https://learn.microsoft.com/en-us/microsoft-edge/webview2/concepts/distribution?tabs=dotnetcsharp#detect-if-a-webview2-runtime-is-already-installed
	for _, edge := range []edgeLocation{
		{registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`},
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`},
		{registry.CURRENT_USER, `Software\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`},
	} {
		key, err := registry.OpenKey(edge.registry, edge.key, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		defer key.Close()

		s, _, err := key.GetStringValue(`pv`)
		if err == nil && s != "" && s != "0.0.0.0" {
			return true
		}
	}

	return false
}

type edgeLocation struct {
	registry registry.Key
	key      string
}
