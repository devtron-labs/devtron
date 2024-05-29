/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package appStore

import (
	"github.com/gorilla/mux"
)

type AppStoreRouterEnterprise interface {
	Init(configRouter *mux.Router)
}

type AppStoreRouterEnterpriseImpl struct {
	deployRestHandler InstalledAppRestHandlerEnterprise
}

func NewAppStoreRouterEnterpriseImpl(restHandler InstalledAppRestHandlerEnterprise) *AppStoreRouterEnterpriseImpl {
	return &AppStoreRouterEnterpriseImpl{
		deployRestHandler: restHandler,
	}
}

func (router AppStoreRouterEnterpriseImpl) Init(configRouter *mux.Router) {

	configRouter.Path("/helm/manifest/download/{installedAppId}/{envId}").
		HandlerFunc(router.deployRestHandler.GetChartForLatestDeployment).Methods("GET")

	configRouter.Path("/helm/manifest/download/{installedAppId}/{envId}/{installedAppVersionHistoryId}").
		HandlerFunc(router.deployRestHandler.GetChartForParticularTrigger).Methods("GET")

}
