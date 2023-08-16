package deployment

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/gorilla/mux"
)

type DeploymentConfigRouter interface {
	Init(configRouter *mux.Router)
}

type DeploymentConfigRouterImpl struct {
	deploymentRestHandler DeploymentConfigRestHandler
	userAuth              logger.UserAuth
}

func NewDeploymentRouterImpl(deploymentRestHandler DeploymentConfigRestHandler, userAuth logger.UserAuth) *DeploymentConfigRouterImpl {
	return &DeploymentConfigRouterImpl{
		deploymentRestHandler: deploymentRestHandler,
		userAuth:              userAuth,
	}
}

func (router DeploymentConfigRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Use(router.userAuth.LoggingMiddleware)
	configRouter.Path("/validate").
		HandlerFunc(router.deploymentRestHandler.CreateChartFromFile).Methods("POST")
	configRouter.Path("/upload").
		HandlerFunc(router.deploymentRestHandler.SaveChart).Methods("PUT")
	configRouter.Path("/fetch").
		HandlerFunc(router.deploymentRestHandler.GetUploadedCharts).Methods("GET")
}
