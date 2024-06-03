/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package appStoreDeployment

import (
	"github.com/gorilla/mux"
)

type CommonDeploymentRouter interface {
	Init(configRouter *mux.Router)
}

type CommonDeploymentRouterImpl struct {
	commonDeploymentRestHandler CommonDeploymentRestHandler
}

func NewCommonDeploymentRouterImpl(commonDeploymentRestHandler CommonDeploymentRestHandler) *CommonDeploymentRouterImpl {
	return &CommonDeploymentRouterImpl{
		commonDeploymentRestHandler: commonDeploymentRestHandler,
	}
}

func (router CommonDeploymentRouterImpl) Init(configRouter *mux.Router) {
	configRouter.Path("/deployment-history").
		HandlerFunc(router.commonDeploymentRestHandler.GetDeploymentHistory).Methods("GET")

	configRouter.Path("/deployment-history/info").
		Queries("version", "{version}").
		HandlerFunc(router.commonDeploymentRestHandler.GetDeploymentHistoryValues).Methods("GET")

	configRouter.Path("/rollback").
		HandlerFunc(router.commonDeploymentRestHandler.RollbackApplication).Methods("PUT")
}
