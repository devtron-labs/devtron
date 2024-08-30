package fluxApplication

import (
	"github.com/devtron-labs/devtron/pkg/fluxApplication"
	"github.com/google/wire"
)

var FluxApplicationWireSet = wire.NewSet(
	fluxApplication.NewFluxApplicationServiceImpl,
	wire.Bind(new(fluxApplication.FluxApplicationService), new(*fluxApplication.FluxApplicationServiceImpl)),

	NewFluxApplicationRestHandlerImpl,
	wire.Bind(new(FluxApplicationRestHandler), new(*FluxApplicationRestHandlerImpl)),

	NewFluxApplicationRouterImpl,
	wire.Bind(new(FluxApplicationRouter), new(*FluxApplicationRouterImpl)),
)
