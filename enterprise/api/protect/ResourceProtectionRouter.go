package protect

import "github.com/gorilla/mux"

type ResourceProtectionRouter interface {
	InitResourceProtectionRouter(protectRouter *mux.Router)
}

type ResourceProtectionRouterImpl struct {
	resourceProtectionRestHandler ResourceProtectionRestHandler
}

func NewResourceProtectionRouterImpl(resourceProtectionRestHandler ResourceProtectionRestHandler) *ResourceProtectionRouterImpl {
	return &ResourceProtectionRouterImpl{resourceProtectionRestHandler: resourceProtectionRestHandler}
}

func (router *ResourceProtectionRouterImpl) InitResourceProtectionRouter(protectRouter *mux.Router) {
	protectRouter.Path("").HandlerFunc(router.resourceProtectionRestHandler.ConfigureResourceProtect).
		Methods("POST")
}
