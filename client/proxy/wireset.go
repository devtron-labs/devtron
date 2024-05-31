/*
 * Copyright (c) 2024. Devtron Inc.
 */

package proxy

import "github.com/google/wire"

var ProxyWireSet = wire.NewSet(
	GetProxyConfig,
	NewProxyRouterImpl,
	wire.Bind(new(ProxyRouter), new(*ProxyRouterImpl)),
)
