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
	userAuth              logger.UserAuth
}

func NewDashboardTelemetryRouterImpl(deploymentRestHandler DashboardTelemetryRestHandler, userAuth logger.UserAuth) *DashboardTelemetryRouterImpl {
	return &DashboardTelemetryRouterImpl{
		deploymentRestHandler: deploymentRestHandler,
		userAuth:              userAuth,
	}
}

func (router DashboardTelemetryRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Use(router.userAuth.LoggingMiddleware)
	configRouter.Path("/dashboardAccessed").
		HandlerFunc(router.deploymentRestHandler.SendDashboardAccessedEvent).Methods("GET")
	configRouter.Path("/dashboardLoggedIn").
		HandlerFunc(router.deploymentRestHandler.SendDashboardLoggedInEvent).Methods("GET")
}
