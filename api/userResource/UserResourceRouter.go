package userResource

import "github.com/gorilla/mux"

type Router interface {
	InitUserResourceRouter(userResourceRouter *mux.Router)
}

type RouterImpl struct {
	restHandler RestHandler
}

func NewUserResourceRouterImpl(restHandler RestHandler) *RouterImpl {
	return &RouterImpl{
		restHandler: restHandler,
	}
}

func (router *RouterImpl) InitUserResourceRouter(userResourceRouter *mux.Router) {
	userResourceRouter.Path("/options/{kind:[a-zA-Z0-9/-]+}/{version:[a-zA-Z0-9]+}").
		HandlerFunc(router.restHandler.GetResourceOptions).Methods("POST")

}
