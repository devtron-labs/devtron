package router

import (
	"github.com/devtron-labs/devtron/api/logger"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type GlobalPluginRouter interface {
	initGlobalPluginRouter(globalPluginRouter *mux.Router)
}

func NewGlobalPluginRouter(logger *zap.SugaredLogger, globalPluginRestHandler restHandler.GlobalPluginRestHandler, userAuth logger.UserAuth) *GlobalPluginRouterImpl {
	return &GlobalPluginRouterImpl{
		logger:                  logger,
		globalPluginRestHandler: globalPluginRestHandler,
		userAuth:                userAuth,
	}
}

type GlobalPluginRouterImpl struct {
	logger                  *zap.SugaredLogger
	globalPluginRestHandler restHandler.GlobalPluginRestHandler
	userAuth                logger.UserAuth
}

func (impl *GlobalPluginRouterImpl) initGlobalPluginRouter(globalPluginRouter *mux.Router) {
	globalPluginRouter.Use(impl.userAuth.LoggingMiddleware)
	globalPluginRouter.Path("/global/list/global-variable").
		HandlerFunc(impl.globalPluginRestHandler.GetAllGlobalVariables).Methods("GET")

	globalPluginRouter.Path("/global/list").
		HandlerFunc(impl.globalPluginRestHandler.ListAllPlugins).Methods("GET")

	globalPluginRouter.Path("/global/{pluginId}").
		HandlerFunc(impl.globalPluginRestHandler.GetPluginDetailById).Methods("GET")
}
