package globalConfig

import "github.com/gorilla/mux"

type AuthorisationConfigRouter interface {
	InitAuthorisationConfigRouter(router *mux.Router)
}

type AuthorisationConfigRouterImpl struct {
	handler AuthorisationConfigRestHandler
}

func NewGlobalConfigAuthorisationConfigRouterImpl(handler AuthorisationConfigRestHandler) *AuthorisationConfigRouterImpl {
	router := &AuthorisationConfigRouterImpl{
		handler: handler,
	}
	return router
}

func (router *AuthorisationConfigRouterImpl) InitAuthorisationConfigRouter(authorisationConfigRouter *mux.Router) {
	authorisationConfigRouter.Path("/global-config").
		HandlerFunc(router.handler.CreateOrUpdateAuthorisationConfig).Methods("POST")
	authorisationConfigRouter.Path("/global-config").
		HandlerFunc(router.handler.GetAllActiveAuthorisationConfig).Methods("GET")
}
