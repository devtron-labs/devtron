package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type GlobalPluginRouter interface {
	initGlobalPluginRouter(globalPluginRouter *mux.Router)
}

func NewGlobalPluginRouter(logger *zap.SugaredLogger, globalPluginRestHandler restHandler.GlobalPluginRestHandler) *GlobalPluginRouterImpl {
	return &GlobalPluginRouterImpl{
		logger:                  logger,
		globalPluginRestHandler: globalPluginRestHandler,
	}
}

type GlobalPluginRouterImpl struct {
	logger                  *zap.SugaredLogger
	globalPluginRestHandler restHandler.GlobalPluginRestHandler
}

func (impl *GlobalPluginRouterImpl) initGlobalPluginRouter(globalPluginRouter *mux.Router) {
	globalPluginRouter.Path("").
		HandlerFunc(impl.globalPluginRestHandler.PatchPlugin).Methods("POST")
	globalPluginRouter.Path("/detail/all").
		HandlerFunc(impl.globalPluginRestHandler.GetAllDetailedPluginInfo).Methods("GET")
	globalPluginRouter.Path("/detail/{pluginId}").
		HandlerFunc(impl.globalPluginRestHandler.GetDetailedPluginInfoByPluginId).Methods("GET")

	globalPluginRouter.Path("/list/global-variable").
		HandlerFunc(impl.globalPluginRestHandler.GetAllGlobalVariables).Methods("GET")

	globalPluginRouter.Path("/list").
		HandlerFunc(impl.globalPluginRestHandler.ListAllPlugins).Methods("GET")

	globalPluginRouter.Path("/{pluginId}").
		HandlerFunc(impl.globalPluginRestHandler.GetPluginDetailById).Methods("GET")
}
