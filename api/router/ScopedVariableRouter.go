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
	"github.com/devtron-labs/devtron/api/restHandler/scopedVariable"
	"github.com/gorilla/mux"
)

type ScopedVariableRouter interface {
	InitScopedVariableRouter(router *mux.Router)
}

type ScopedVariableRouterImpl struct {
	scopedVariableRestHandler scopedVariable.ScopedVariableRestHandler
}

func NewScopedVariableRouterImpl(scopedVariableRestHandler scopedVariable.ScopedVariableRestHandler) *ScopedVariableRouterImpl {
	router := &ScopedVariableRouterImpl{
		scopedVariableRestHandler: scopedVariableRestHandler,
	}
	return router
}

func (impl ScopedVariableRouterImpl) InitScopedVariableRouter(router *mux.Router) {
	router.Path("/variables").
		HandlerFunc(impl.scopedVariableRestHandler.CreateVariables).
		Methods("POST")
	router.Path("/variables").
		HandlerFunc(impl.scopedVariableRestHandler.GetScopedVariables).
		Methods("GET")
	router.Path("/variables/detail").
		HandlerFunc(impl.scopedVariableRestHandler.GetJsonForVariables).
		Methods("GET")

}
