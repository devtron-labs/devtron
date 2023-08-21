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
	protectRouter.Path("").HandlerFunc(router.resourceProtectionRestHandler.GetResourceProtectMetadata).
		Queries("appId", "{appId}").
		Methods("GET")
	protectRouter.Path("/env").HandlerFunc(router.resourceProtectionRestHandler.GetResourceProtectMetadataForEnv).
		Queries("envId", "{envId}").
		Methods("GET")
}
