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

type HelmRouter interface {
	initHelmRouter(helmRouter *mux.Router)
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

func NewHelmRouter(pipelineRestHandler restHandler.PipelineTriggerRestHandler, sseChannel *sse2.SSE) *HelmRouterImpl {
	routerImpl := &HelmRouterImpl{restHandler: pipelineRestHandler}
	sse = sseChannel
	return routerImpl
}

type HelmRouterImpl struct {
	restHandler restHandler.PipelineTriggerRestHandler
}

func (router HelmRouterImpl) initHelmRouter(helmRouter *mux.Router) {
	helmRouter.Path("/cd-pipeline/trigger").HandlerFunc(router.restHandler.OverrideConfig).Methods("POST")
	helmRouter.Path("/update-release-status").HandlerFunc(router.restHandler.ReleaseStatusUpdate).Methods("POST")
	helmRouter.Path("/stop-start-app").HandlerFunc(router.restHandler.StartStopApp).Methods("POST")
	helmRouter.Path("/stop-start-dg").HandlerFunc(router.restHandler.StartStopDeploymentGroup).Methods("POST")
	helmRouter.Path("/release/").
		Handler(sse2.SubscribeHandler(sse.Broker, PollTopic, fetchReleaseData)).
		Methods("GET").
		Queries("name", "{name}")
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
