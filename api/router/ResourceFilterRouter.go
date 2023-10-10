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
	//if no param passed, it will list all the active filters
	//if pipelineId is passed, will fetch all the filters that will be applied on this pipeline(app,env)
	//if other params are added in the future, please follow a weightage for the params
	router.Path("").Queries("pipelineId", "{pipelineId}").
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

}
