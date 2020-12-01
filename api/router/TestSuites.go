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
	configRouter.Path("/testsuite/{appId}").HandlerFunc(impl.testSuitRouter.RedirectTestSuites).Methods("GET")
	configRouter.Path("/testsuite/{appId}/detailed").HandlerFunc(impl.testSuitRouter.RedirectDetailedTestSuites).Methods("GET")
	configRouter.Path("/testsuite/{appId}/{pipelineId}").HandlerFunc(impl.testSuitRouter.RedirectSuitByID).Methods("GET")

	configRouter.Path("/testcase/{appId}").HandlerFunc(impl.testSuitRouter.RedirectTestCases).Methods("GET")
	configRouter.Path("/testcase/{appId}/{pipelineId}").HandlerFunc(impl.testSuitRouter.RedirectTestCaseByID).Methods("GET")

	configRouter.Path("/triggers/{appId}/{pipelineId}").HandlerFunc(impl.testSuitRouter.RedirectTriggerForPipeline).Methods("GET")
	configRouter.Path("/triggers/{appId}/{pipelineId}/{triggerId}").HandlerFunc(impl.testSuitRouter.RedirectTriggerForBuild).Methods("GET")

	configRouter.Path("/filters/{appId}/{pipelineId}").HandlerFunc(impl.testSuitRouter.RedirectFilterForPipeline).Methods("GET")
	configRouter.Path("/filters/{appId}/{pipelineId}/{triggerId}").HandlerFunc(impl.testSuitRouter.RedirectFilterForBuild).Methods("GET")

}
