package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type DeploymentConfigurationRouter interface {
	initDeploymentConfigurationRouter(configRouter *mux.Router)
}

type DeploymentConfigurationRouterImpl struct {
	deploymentGroupRestHandler restHandler.DeploymentConfigurationRestHandler
}

func NewDeploymentConfigurationRouter(deploymentGroupRestHandler restHandler.DeploymentConfigurationRestHandler) *DeploymentConfigurationRouterImpl {
	router := &DeploymentConfigurationRouterImpl{
		deploymentGroupRestHandler: deploymentGroupRestHandler,
	}
	return router
}

func (router DeploymentConfigurationRouterImpl) initDeploymentConfigurationRouter(configRouter *mux.Router) {
	configRouter.Path("/autocomplete").
		HandlerFunc(router.deploymentGroupRestHandler.ConfigAutoComplete).Methods("POST")
}
