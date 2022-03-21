package deployment

import "github.com/gorilla/mux"

type DeploymentRouter interface {
	Init(configRouter *mux.Router)
}

type DeploymentRouterImpl struct {
	deploymentRestHandler DeploymentRestHandler
}

func NewDeploymentRouterImpl(deploymentRestHandler DeploymentRestHandler) *DeploymentRouterImpl {
	return &DeploymentRouterImpl{
		deploymentRestHandler: deploymentRestHandler,
	}
}

func (router DeploymentRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("/upload").
		HandlerFunc(router.deploymentRestHandler.CreateChartFromFile).Methods("POST")
}
