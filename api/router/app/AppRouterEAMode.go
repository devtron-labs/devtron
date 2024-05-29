/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package app

import (
	"github.com/devtron-labs/devtron/api/router/app/appInfo"
	"github.com/devtron-labs/devtron/api/router/app/appList"
	"github.com/gorilla/mux"
)

type AppRouterEAMode interface {
	InitAppRouterEAMode(helmRouter *mux.Router)
}

type AppRouterEAModeImpl struct {
	appInfoRouter      appInfo.AppInfoRouter
	appFilteringRouter appList.AppFilteringRouter
}

func NewAppRouterEAModeImpl(appInfoRouter appInfo.AppInfoRouter, appFilteringRouter appList.AppFilteringRouter) *AppRouterEAModeImpl {
	router := &AppRouterEAModeImpl{
		appInfoRouter:      appInfoRouter,
		appFilteringRouter: appFilteringRouter,
	}
	return router
}

func (router AppRouterEAModeImpl) InitAppRouterEAMode(AppRouterEAMode *mux.Router) {
	router.appInfoRouter.InitAppInfoRouter(AppRouterEAMode)
	appFilterRouter := AppRouterEAMode.PathPrefix("/filter").Subrouter()
	router.appFilteringRouter.InitAppFilteringRouter(appFilterRouter)
}
