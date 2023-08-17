package dashboardEvent

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/gorilla/mux"
)

type DashboardTelemetryRouter interface {
	Init(configRouter *mux.Router)
}

type DashboardTelemetryRouterImpl struct {
	deploymentRestHandler DashboardTelemetryRestHandler
}

func NewDashboardTelemetryRouterImpl(deploymentRestHandler DashboardTelemetryRestHandler, userAuth logger.UserAuth) *DashboardTelemetryRouterImpl {
	return &DashboardTelemetryRouterImpl{
		deploymentRestHandler: deploymentRestHandler,
	}
}

func (router DashboardTelemetryRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("/dashboardAccessed").
		HandlerFunc(router.deploymentRestHandler.SendDashboardAccessedEvent).Methods("GET")
	configRouter.Path("/dashboardLoggedIn").
		HandlerFunc(router.deploymentRestHandler.SendDashboardLoggedInEvent).Methods("GET")
}
