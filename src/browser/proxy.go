package browser

import (
	"context"
	"fmt"
	"net"
	"net/url"

	_ "github.com/wzshiming/anyproxy/proxies/httpproxy"
	_ "github.com/wzshiming/bridge/protocols/connect"
	_ "github.com/wzshiming/bridge/protocols/socks4"
	_ "github.com/wzshiming/bridge/protocols/socks5"
	_ "github.com/wzshiming/bridge/protocols/tls"
	_ "github.com/wzshiming/bridge/protocols/ssh"
	"github.com/sag-enhanced/native-app/src/options"
	"github.com/wzshiming/bridge/chain"
	"github.com/wzshiming/bridge/config"
	"github.com/wzshiming/bridge/logger"
)


func createProxyProxy(proxy *url.URL, options *options.Options, stop context.Context) (*url.URL, error) {
  freePort, err := getFreePort()
  if err != nil {
    return nil, err
  }

  localProxy := &url.URL{
    Scheme: "http",
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
    IdleTimeout: 120,
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
