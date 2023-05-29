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
	sse2 "github.com/devtron-labs/devtron/api/sse"
	"github.com/gorilla/mux"
	"github.com/juju/errors"
	_ "k8s.io/client-go/rest"
	"net/http"
	"strconv"
	"time"
)

var sse *sse2.SSE

type PipelineTriggerRouter interface {
	initPipelineTriggerRouter(pipelineTriggerRouter *mux.Router)
}

func PollTopic(r *http.Request) (string, error) {
	parameters := mux.Vars(r)
	if parameters == nil {
		return "", errors.New("missing mandatory parameters")
	}
	name := parameters["name"]
	if name == "" {
		return "", errors.New("missing mandatory parameters")
	}
	return "/" + name, nil
}

func NewPipelineTriggerRouter(pipelineRestHandler restHandler.PipelineTriggerRestHandler, sseChannel *sse2.SSE) *PipelineTriggerRouterImpl {
	routerImpl := &PipelineTriggerRouterImpl{restHandler: pipelineRestHandler}
	sse = sseChannel
	return routerImpl
}

type PipelineTriggerRouterImpl struct {
	restHandler restHandler.PipelineTriggerRestHandler
}

func (router PipelineTriggerRouterImpl) initPipelineTriggerRouter(pipelineTriggerRouter *mux.Router) {
	pipelineTriggerRouter.Path("/cd-pipeline/trigger").HandlerFunc(router.restHandler.OverrideConfig).Methods("POST")
	pipelineTriggerRouter.Path("/update-release-status").HandlerFunc(router.restHandler.ReleaseStatusUpdate).Methods("POST")
	pipelineTriggerRouter.Path("/stop-start-app").HandlerFunc(router.restHandler.StartStopApp).Methods("POST")
	pipelineTriggerRouter.Path("/stop-start-dg").HandlerFunc(router.restHandler.StartStopDeploymentGroup).Methods("POST")
	pipelineTriggerRouter.Path("/release/").
		Handler(sse2.SubscribeHandler(sse.Broker, PollTopic, fetchReleaseData)).
		Methods("GET").
		Queries("name", "{name}")

	pipelineTriggerRouter.Path("/deployment-configuration/latest/saved/{appId}/{pipelineId}").HandlerFunc(router.restHandler.GetAllLatestDeploymentConfiguration).Methods("GET")
	pipelineTriggerRouter.Path("/manifest/download/{appId}/{envId}").Queries("runner", "{runner}").HandlerFunc(router.restHandler.DownloadManifest).Methods("GET")
	pipelineTriggerRouter.Path("/manifest/download/{appId}/{envId}/{cd_workflow_id}").HandlerFunc(router.restHandler.DownloadManifestForSpecificTrigger).Methods("GET")
}

func fetchReleaseData(r *http.Request, receive <-chan int, send chan<- int) {
	parameters := mux.Vars(r)
	name := parameters["name"]
	for i := 0; i <= 10; i++ {
		select {
		case <-receive:
			return
		default:
		}
		time.Sleep(1 * time.Second)
		data := []byte(time.Now().String() + "-" + strconv.Itoa(i))
		sse.OutboundChannel <- sse2.SSEMessage{"", data, "/" + name}
	}
	send <- 1
}
