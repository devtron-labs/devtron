package router

import (
	"github.com/devtron-labs/devtron/api/restHandler/resourceFilter"
	"github.com/gorilla/mux"
)

type ResourceFilterRouter interface {
	InitResourceFilterRouter(router *mux.Router)
}

type ResourceFilterRouterImpl struct {
	resourceFilterRestHandler resourceFilter.ResourceFilterRestHandler
}

func NewResourceFilterRouterImpl(resourceFilterRestHandler resourceFilter.ResourceFilterRestHandler) *ResourceFilterRouterImpl {
	router := &ResourceFilterRouterImpl{
		resourceFilterRestHandler: resourceFilterRestHandler,
	}
	return router
}

func (impl *ResourceFilterRouterImpl) InitResourceFilterRouter(router *mux.Router) {
	router.Path("").
		HandlerFunc(impl.resourceFilterRestHandler.ListFilters).
		Methods("GET")

	router.Path("/{id}").
		HandlerFunc(impl.resourceFilterRestHandler.GetFilterById).
		Methods("GET")

	router.Path("").
		HandlerFunc(impl.resourceFilterRestHandler.CreateFilter).
		Methods("PUT")

	router.Path("/{id}").
		HandlerFunc(impl.resourceFilterRestHandler.UpdateFilter).
		Methods("POST")

	router.Path("/{id}").
		HandlerFunc(impl.resourceFilterRestHandler.DeleteFilter).
		Methods("DELETE")

	router.Path("/expression/validate").
		HandlerFunc(impl.resourceFilterRestHandler.ValidateExpression).
		Methods("POST")

	router.Path("/pipeline/{pipelineId}").
		HandlerFunc(impl.resourceFilterRestHandler.GetFiltersByPipelineId).
		Methods("GET")
}
