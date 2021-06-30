package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type BulkUpdateRouter interface {
	initBulkUpdateRouter(bulkRouter *mux.Router)
}

type BulkUpdateRouterImpl struct {
restHandler   restHandler.BulkUpdateRestHandler
}

func NewBulkUpdateRouterImpl(handler restHandler.BulkUpdateRestHandler) *BulkUpdateRouterImpl {
	router := &BulkUpdateRouterImpl{
		restHandler: handler,
	}
	return router
}
func (router BulkUpdateRouterImpl) initBulkUpdateRouter(bulkRouter *mux.Router) {
	bulkRouter.Path("/application/see-example").HandlerFunc(router.restHandler.GetExampleInputBulkUpdate).Methods("GET")
	bulkRouter.Path("/application/{dryrun}").HandlerFunc(router.restHandler.GetAppNameDeploymentTemplate).Methods("POST")
	bulkRouter.Path("/application").HandlerFunc(router.restHandler.BulkUpdateDeploymentTemplate).Methods("POST")
}