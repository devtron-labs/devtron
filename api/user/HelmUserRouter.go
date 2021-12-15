package user

import (
	"github.com/argoproj/argo-cd/util/settings"
	"github.com/gorilla/mux"
)

type HelmUserRouter interface {
	InitHelmUserRouter(helmRouter *mux.Router)
}

type HelmUserRouterImpl struct {
	userRestHandler HelmUserRestHandler
}

func NewHelmUserRouterImpl(userRestHandler HelmUserRestHandler, settings *settings.ArgoCDSettings) *HelmUserRouterImpl {
	tlsConfig := settings.TLSConfig()
	if tlsConfig != nil {
		tlsConfig.InsecureSkipVerify = true
	}
	router := &HelmUserRouterImpl{
		userRestHandler: userRestHandler,
	}
	return router
}

func (router HelmUserRouterImpl) InitHelmUserRouter(userAuthRouter *mux.Router) {
	userAuthRouter.Path("/{id}").
		HandlerFunc(router.userRestHandler.GetById).Methods("GET")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.GetAll).Methods("GET")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.CreateHelmUser).Methods("POST")
	userAuthRouter.Path("").
		HandlerFunc(router.userRestHandler.UpdateHelmUser).Methods("PUT")
}
