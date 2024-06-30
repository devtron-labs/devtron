/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type BulkUpdateRouter interface {
	initBulkUpdateRouter(bulkRouter *mux.Router)
}

type BulkUpdateRouterImpl struct {
	restHandler restHandler.BulkUpdateRestHandler
}

func NewBulkUpdateRouterImpl(handler restHandler.BulkUpdateRestHandler) *BulkUpdateRouterImpl {
	router := &BulkUpdateRouterImpl{
		restHandler: handler,
	}
	return router
}
func (router BulkUpdateRouterImpl) initBulkUpdateRouter(bulkRouter *mux.Router) {
	bulkRouter.Path("/{apiVersion}/{kind}/readme").HandlerFunc(router.restHandler.FindBulkUpdateReadme).Methods("GET")
	bulkRouter.Path("/v1beta1/application/dryrun").HandlerFunc(router.restHandler.GetImpactedAppsName).Methods("POST")
	bulkRouter.Path("/v1beta1/application").HandlerFunc(router.restHandler.BulkUpdate).Methods("POST")

	bulkRouter.Path("/v1beta1/hibernate").HandlerFunc(router.restHandler.BulkHibernate).Methods("POST")
	bulkRouter.Path("/v1beta1/unhibernate").HandlerFunc(router.restHandler.BulkUnHibernate).Methods("POST")
	bulkRouter.Path("/v1beta1/deploy").HandlerFunc(router.restHandler.BulkDeploy).Methods("POST")
	bulkRouter.Path("/v1beta1/build").HandlerFunc(router.restHandler.BulkBuildTrigger).Methods("POST")
	bulkRouter.Path("/v1beta1/cd-pipeline").HandlerFunc(router.restHandler.HandleCdPipelineBulkAction).Methods("POST")

}
