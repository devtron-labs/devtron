package k8s

import (
	"github.com/google/wire"
)

var K8sApplicationWireSet = wire.NewSet(
	NewK8sApplicationRouterImpl,
	wire.Bind(new(K8sApplicationRouter), new(*K8sApplicationRouterImpl)),
	NewK8sApplicationRestHandlerImpl,
	wire.Bind(new(K8sApplicationRestHandler), new(*K8sApplicationRestHandlerImpl)),
	NewK8sApplicationServiceImpl,
	wire.Bind(new(K8sApplicationService), new(*K8sApplicationServiceImpl)),
)
