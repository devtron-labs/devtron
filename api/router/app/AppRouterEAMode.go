/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
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
