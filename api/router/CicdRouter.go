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

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type CicdRouter interface {
	initCicdRouter(helmRouter *mux.Router)
}

type CicdRouterImpl struct {
	cicdRestHandler restHandler.CicdRestHandler
}

func NewCicdRouterImpl(cicdRestHandler restHandler.CicdRestHandler) *CicdRouterImpl {
	router := &CicdRouterImpl{
		cicdRestHandler: cicdRestHandler,
	}
	return router
}

func (router CicdRouterImpl) initCicdRouter(cicdRouter *mux.Router) {

	cicdRouter.Path("/cicd/save").
		HandlerFunc(router.cicdRestHandler.SaveCICD).
		Methods("POST")

	cicdRouter.Path("/cicd/update").
		HandlerFunc(router.cicdRestHandler.UpdateCICD).
		Methods("POST")

	cicdRouter.Path("/cicd/delete").
		HandlerFunc(router.cicdRestHandler.DeleteCICD).
		Methods("POST")

	cicdRouter.Path("/cicd/find").
		HandlerFunc(router.cicdRestHandler.FindByCICD).
		Methods("POST")
}
