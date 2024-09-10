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

type GlobalCMCSRouter interface {
	initGlobalCMCSRouter(configRouter *mux.Router)
}
type GlobalCMCSRouterImpl struct {
	restHandler restHandler.GlobalCMCSRestHandler
}

func NewGlobalCMCSRouterImpl(restHandler restHandler.GlobalCMCSRestHandler) *GlobalCMCSRouterImpl {
	return &GlobalCMCSRouterImpl{restHandler: restHandler}

}

func (router GlobalCMCSRouterImpl) initGlobalCMCSRouter(configRouter *mux.Router) {
	configRouter.Path("/{config_name}/{config_type}").
		HandlerFunc(router.restHandler.GetGlobalCMCSDataByConfigTypeAndName).Methods("GET")
	configRouter.Path("/all").
		HandlerFunc(router.restHandler.GetAllGlobalCMCSData).Methods("GET")
	configRouter.Path("").
		HandlerFunc(router.restHandler.CreateGlobalCMCSConfig).Methods("POST")
	configRouter.Path("/update").
		HandlerFunc(router.restHandler.UpdateGlobalCMCSDataById).Methods("PUT")
	configRouter.Path("/delete/{id}").
		HandlerFunc(router.restHandler.DeleteByID).Methods("DELETE")
}
