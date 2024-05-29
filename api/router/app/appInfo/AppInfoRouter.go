/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package appInfo

import (
	"github.com/devtron-labs/devtron/api/restHandler/app/appInfo"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type AppInfoRouter interface {
	InitAppInfoRouter(router *mux.Router)
}

type AppInfoRouterImpl struct {
	logger  *zap.SugaredLogger
	handler appInfo.AppInfoRestHandler
}

func NewAppInfoRouterImpl(logger *zap.SugaredLogger, handler appInfo.AppInfoRestHandler) *AppInfoRouterImpl {
	router := &AppInfoRouterImpl{
		logger:  logger,
		handler: handler,
	}
	return router
}

func (router AppInfoRouterImpl) InitAppInfoRouter(appRouter *mux.Router) {
	appRouter.Path("/labels/list").
		HandlerFunc(router.handler.GetAllLabels).Methods("GET")
	appRouter.Path("/meta/info/{appId}").
		HandlerFunc(router.handler.GetAppMetaInfo).Methods("GET")

	appRouter.Path("/helm/meta/info/{appId}").
		HandlerFunc(router.handler.GetHelmAppMetaInfo).Methods("GET")

	appRouter.Path("/edit").
		HandlerFunc(router.handler.UpdateApp).Methods("POST")
	appRouter.Path("/edit/projects").
		HandlerFunc(router.handler.UpdateProjectForApps).Methods("POST")

	appRouter.Path("/min").HandlerFunc(router.handler.GetAppListByTeamIds).Methods("GET")

	appRouter.Path("/note").
		Methods("PUT").
		HandlerFunc(router.handler.UpdateAppNote)
}
