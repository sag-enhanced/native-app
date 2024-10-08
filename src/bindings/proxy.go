package bindings

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"github.com/sag-enhanced/native-app/src/options"
	_ "github.com/wzshiming/anyproxy/proxies/socks5"
	"github.com/wzshiming/bridge/chain"
	"github.com/wzshiming/bridge/config"
	"github.com/wzshiming/bridge/logger"
	_ "github.com/wzshiming/bridge/protocols/connect"
	_ "github.com/wzshiming/bridge/protocols/socks4"
	_ "github.com/wzshiming/bridge/protocols/socks5"
	_ "github.com/wzshiming/bridge/protocols/ssh"
	_ "github.com/wzshiming/bridge/protocols/tls"
)

var proxyClients = make(map[string]context.CancelFunc)
var proxyHandleLock = sync.Mutex{}

func (b *Bindings) ProxyNew(proxyUrl string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())

	parsedProxyUrl, err := url.Parse(proxyUrl)
	if err != nil {
		cancel()
		return "", err
	}

	localProxy, err := createProxyProxy(parsedProxyUrl, b.options, ctx)
	if err != nil {
		cancel()
		return "", err
	}
	if b.options.Verbose {
		fmt.Println("Created new proxy with handle", proxyUrl, localProxy)
	}

	proxyHandleLock.Lock()
	proxyClients[localProxy.String()] = cancel
	proxyHandleLock.Unlock()

	return localProxy.String(), nil
}

func (b *Bindings) ProxyDestroy(handle string) error {
	proxyHandleLock.Lock()
	cancel, ok := proxyClients[handle]
	proxyHandleLock.Unlock()
	if !ok {
		return fmt.Errorf("invalid handle %s", handle)
	}
	cancel()
	return nil
}

func createProxyProxy(proxy *url.URL, options *options.Options, stop context.Context) (*url.URL, error) {
	freePort, err := getFreePort()
	if err != nil {
		return nil, err
	}

	localProxy := &url.URL{
		Scheme: "socks5",
		Host:   fmt.Sprintf("127.0.0.1:%d", freePort),
	}
	if options.Verbose {
		fmt.Println("Local proxy", localProxy)
	}

	cfg := config.Chain{
		Bind: []config.Node{
			{
				LB: []string{localProxy.String()},
			},
		},
		Proxy: []config.Node{
			{
				LB: []string{"-"},
			},
			{
				LB: []string{proxy.String()},
			},
		},
		IdleTimeout: 120 * time.Second,
	}
	b := chain.NewBridge(logger.Std, false)

	go func() {
		if err := b.BridgeWithConfig(stop, cfg); err != nil {
			fmt.Println("Error running proxy", err)
		}
	}()

	return localProxy, nil
}

func getFreePort() (int, error) {
	l, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
