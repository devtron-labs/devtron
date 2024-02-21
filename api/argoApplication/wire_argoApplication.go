package argoApplication

import (
	"github.com/devtron-labs/devtron/pkg/argoApplication"
	"github.com/google/wire"
)

var ArgoApplicationWireSet = wire.NewSet(
	argoApplication.NewArgoApplicationServiceImpl,
	wire.Bind(new(argoApplication.ArgoApplicationService), new(*argoApplication.ArgoApplicationServiceImpl)),

	NewArgoApplicationRestHandlerImpl,
	wire.Bind(new(ArgoApplicationRestHandler), new(*ArgoApplicationRestHandlerImpl)),

	NewArgoApplicationRouterImpl,
	wire.Bind(new(ArgoApplicationRouter), new(*ArgoApplicationRouterImpl)),
)
