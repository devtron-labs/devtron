package fluxApplication

import (
	"github.com/gorilla/mux"
)

type FluxApplicationRouter interface {
	InitFluxApplicationRouter(fluxApplicationRouter *mux.Router)
}

type FluxApplicationRouterImpl struct {
	fluxApplicationRestHandler FluxApplicationRestHandler
}

func NewFluxApplicationRouterImpl(fluxApplicationRestHandler FluxApplicationRestHandler) *FluxApplicationRouterImpl {
	return &FluxApplicationRouterImpl{
		fluxApplicationRestHandler: fluxApplicationRestHandler,
	}
}

func (impl *FluxApplicationRouterImpl) InitFluxApplicationRouter(fluxApplicationRouter *mux.Router) {
	fluxApplicationRouter.Path("").
		Methods("GET").
		HandlerFunc(impl.fluxApplicationRestHandler.ListFluxApplications)

	fluxApplicationRouter.Path("/app").Queries("appId", "{appId}").
		HandlerFunc(impl.fluxApplicationRestHandler.GetApplicationDetail).Methods("GET")
}
