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
