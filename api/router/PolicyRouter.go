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

type PolicyRouter interface {
	InitPolicyRouter(configRouter *mux.Router)
}
type PolicyRouterImpl struct {
	policyRestHandler restHandler.PolicyRestHandler
}

func NewPolicyRouterImpl(policyRestHandler restHandler.PolicyRestHandler) *PolicyRouterImpl {
	return &PolicyRouterImpl{
		policyRestHandler: policyRestHandler,
	}
}
func (impl PolicyRouterImpl) InitPolicyRouter(configRouter *mux.Router) {
	configRouter.Path("/save").HandlerFunc(impl.policyRestHandler.SavePolicy).Methods("POST")
	configRouter.Path("/update").HandlerFunc(impl.policyRestHandler.UpdatePolicy).Methods("POST")
	configRouter.Path("/list").HandlerFunc(impl.policyRestHandler.GetPolicy).Methods("GET")
	configRouter.Path("/verify/webhook").HandlerFunc(impl.policyRestHandler.VerifyImage).Methods("POST")
}
