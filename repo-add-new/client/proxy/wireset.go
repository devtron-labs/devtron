package proxy

import "github.com/google/wire"

var ProxyWireSet = wire.NewSet(
	GetProxyConfig,
	NewProxyRouterImpl,
	wire.Bind(new(ProxyRouter), new(*ProxyRouterImpl)),
)
