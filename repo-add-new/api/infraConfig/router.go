package infraConfig

import "github.com/gorilla/mux"

type InfraConfigRouter interface {
	InitInfraConfigRouter(configRouter *mux.Router)
}

type InfraConfigRouterImpl struct {
	infraConfigRestHandler InfraConfigRestHandler
}

func NewInfraProfileRouterImpl(infraConfigRestHandler InfraConfigRestHandler) *InfraConfigRouterImpl {
	return &InfraConfigRouterImpl{
		infraConfigRestHandler: infraConfigRestHandler,
	}
}

func (impl *InfraConfigRouterImpl) InitInfraConfigRouter(configRouter *mux.Router) {
	configRouter.Path("/profile/{name}").
		HandlerFunc(impl.infraConfigRestHandler.GetProfile).
		Methods("GET")

	configRouter.Path("/profile/{name}").
		HandlerFunc(impl.infraConfigRestHandler.UpdateInfraProfile).
		Methods("PUT")
}
