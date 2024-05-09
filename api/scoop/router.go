package scoop

import "github.com/gorilla/mux"

type Router interface {
	InitScoopRouter(router *mux.Router)
}

type RouterImpl struct {
	restHandler RestHandler
}

func NewRouterImpl(restHandler RestHandler) *RouterImpl {
	return &RouterImpl{
		restHandler: restHandler,
	}
}
func (impl RouterImpl) InitScoopRouter(router *mux.Router) {
	router.Path("/notify").
		HandlerFunc(impl.restHandler.HandleNotificationEvent).Methods("POST")
}
