package configDiff

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/configDiff"
	"github.com/gorilla/mux"
)

type DeploymentConfigurationRouter interface {
	InitDeploymentConfigurationRouter(configRouter *mux.Router)
}

type DeploymentConfigurationRouterImpl struct {
	deploymentGroupRestHandler configDiff.DeploymentConfigurationRestHandler
}

func NewDeploymentConfigurationRouter(deploymentGroupRestHandler configDiff.DeploymentConfigurationRestHandler) *DeploymentConfigurationRouterImpl {
	router := &DeploymentConfigurationRouterImpl{
		deploymentGroupRestHandler: deploymentGroupRestHandler,
	}
	return router
}

func (router DeploymentConfigurationRouterImpl) InitDeploymentConfigurationRouter(configRouter *mux.Router) {
	configRouter.Path("/autocomplete").
		HandlerFunc(router.deploymentGroupRestHandler.ConfigAutoComplete).
		Methods("GET")
	configRouter.Path("/data").
		HandlerFunc(router.deploymentGroupRestHandler.GetConfigData).
		Methods("GET")
	configRouter.Path("/compare/{resource}").
		HandlerFunc(router.deploymentGroupRestHandler.CompareCategoryWiseConfigData).
		Methods("GET")

}
