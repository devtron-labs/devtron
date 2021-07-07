package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type BulkUpdateRouter interface {
	initBulkUpdateRouter(bulkRouter *mux.Router)
}

type BulkUpdateRouterImpl struct {
	restHandler restHandler.BulkUpdateRestHandler
}

func NewBulkUpdateRouterImpl(handler restHandler.BulkUpdateRestHandler) *BulkUpdateRouterImpl {
	router := &BulkUpdateRouterImpl{
		restHandler: handler,
	}
	return router
}
func (router BulkUpdateRouterImpl) initBulkUpdateRouter(bulkRouter *mux.Router) {
	bulkRouter.Path("/v1beta1/application/see-example").HandlerFunc(router.restHandler.GetExampleOperationBulkUpdate).Methods("GET")
	bulkRouter.Path("/v1beta1/application/dryrun").HandlerFunc(router.restHandler.GetAppNameDeploymentTemplate).Methods("POST")
	bulkRouter.Path("/v1beta1/application").HandlerFunc(router.restHandler.BulkUpdateDeploymentTemplate).Methods("POST")
}
