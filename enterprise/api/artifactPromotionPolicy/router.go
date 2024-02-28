package artifactPromotionPolicy

import (
	"github.com/gorilla/mux"
)

type Router interface {
	InitRouter(router *mux.Router)
}

type RouterImpl struct {
	restHandler RestHandler
}

func NewCommonPolicyRouterImpl(restHandler RestHandler) *RouterImpl {
	return &RouterImpl{
		restHandler: restHandler,
	}
}

func (r *RouterImpl) InitRouter(router *mux.Router) {
	router.Path("").HandlerFunc(r.restHandler.CreatePolicy).
		Methods("POST")
	router.Path("").HandlerFunc(r.restHandler.UpdatePolicy).
		Methods("PUT")
	router.Path("").HandlerFunc(r.restHandler.GetPoliciesMetadata).
		Methods("GET")
	router.Path("/{name}").HandlerFunc(r.restHandler.DeletePolicy).
		Methods("DELETE")
	router.Path("/{name}").HandlerFunc(r.restHandler.GetPolicyByName).
		Methods("GET")

}
