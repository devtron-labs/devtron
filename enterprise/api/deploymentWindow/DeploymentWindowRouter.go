package deploymentWindow

import (
	"github.com/gorilla/mux"
)

type DeploymentWindowRouter interface {
	InitDeploymentWindowRouter(configRouter *mux.Router)
}

type DeploymentWindowRouterImpl struct {
	DeploymentWindowRestHandler DeploymentWindowRestHandler
}

func NewDeploymentWindowRouterImpl(DeploymentWindowRestHandler DeploymentWindowRestHandler) *DeploymentWindowRouterImpl {
	return &DeploymentWindowRouterImpl{DeploymentWindowRestHandler: DeploymentWindowRestHandler}
}

func (router *DeploymentWindowRouterImpl) InitDeploymentWindowRouter(configRouter *mux.Router) {
	configRouter.
		Path("/profile").
		HandlerFunc(router.DeploymentWindowRestHandler.CreateDeploymentWindowProfile).
		Methods("POST")
	configRouter.
		Path("/profile").
		HandlerFunc(router.DeploymentWindowRestHandler.UpdateDeploymentWindowProfile).
		Methods("PUT")
	configRouter.
		Path("/profile").
		HandlerFunc(router.DeploymentWindowRestHandler.DeleteDeploymentWindowProfile).
		Queries("profileId", "{profileId}").
		Methods("DELETE")
	configRouter.
		Path("/profile").
		HandlerFunc(router.DeploymentWindowRestHandler.GetDeploymentWindowProfile).
		Queries("profileId", "{profileId}").
		Methods("GET")
	configRouter.
		Path("/profile/list").
		HandlerFunc(router.DeploymentWindowRestHandler.ListAppDeploymentWindowProfiles).
		Methods("GET")
	configRouter.
		Path("/overview").
		HandlerFunc(router.DeploymentWindowRestHandler.GetDeploymentWindowProfileAppOverview).
		Queries("appId", "{appId}").
		//Queries("envIds", "{envIds}").
		Methods("GET")
	configRouter.
		Path("/state").
		HandlerFunc(router.DeploymentWindowRestHandler.GetDeploymentWindowProfileStateForApp).
		Queries("appId", "{appId}").
		//Queries("envIds", "{envIds}").
		Methods("GET")
	configRouter.
		Path("/state/appgroup").
		HandlerFunc(router.DeploymentWindowRestHandler.GetDeploymentWindowProfileStateForAppGroup).
		Methods("POST")
}
