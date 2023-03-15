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

type TestSuitRouter interface {
	InitTestSuitRouter(gocdRouter *mux.Router)
}
type TestSuitRouterImpl struct {
	testSuitRouter restHandler.TestSuitRestHandler
}

func NewTestSuitRouterImpl(testSuitRouter restHandler.TestSuitRestHandler) *TestSuitRouterImpl {
	return &TestSuitRouterImpl{testSuitRouter: testSuitRouter}
}

func (impl TestSuitRouterImpl) InitTestSuitRouter(configRouter *mux.Router) {
	configRouter.Path("/suites/proxy").HandlerFunc(impl.testSuitRouter.SuitesProxy).Methods("POST")
	configRouter.Path("/suites/list").HandlerFunc(impl.testSuitRouter.GetTestSuites).Methods("GET")
	configRouter.Path("/suites/list/detail").HandlerFunc(impl.testSuitRouter.DetailedTestSuites).Methods("GET")
	configRouter.Path("/suites/{pipelineId}").HandlerFunc(impl.testSuitRouter.GetAllSuitByID).Methods("GET")
	configRouter.Path("/cases").HandlerFunc(impl.testSuitRouter.GetAllTestCases).Methods("GET")
	configRouter.Path("/cases/{pipelineId}").HandlerFunc(impl.testSuitRouter.GetTestCaseByID).Methods("GET")
	configRouter.Path("/trigger/{pipelineId}").HandlerFunc(impl.testSuitRouter.RedirectTriggerForApp).Methods("GET")
	configRouter.Path("/trigger/{pipelineId}/{triggerId}").HandlerFunc(impl.testSuitRouter.RedirectTriggerForEnv).Methods("GET")
}
