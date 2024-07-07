//go:build linux

package options

// webview *is* available, but for some reason it's using HTTP/1.1 instead of HTTP/2
// when connecting to the server. This will be rejected by the server, so we'lL
// disable it for now.
func isWebviewAvailable() bool {
	return false
}
