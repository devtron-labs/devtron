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

package restHandler

import (
	"github.com/devtron-labs/devtron/internal/util/ArgoUtil"
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type CDRestHandler interface {
	FetchResourceTree(w http.ResponseWriter, r *http.Request)

	FetchPodContainerLogs(w http.ResponseWriter, r *http.Request)
}

type CDRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	resourceService ArgoUtil.ResourceService
}

func NewCDRestHandlerImpl(logger *zap.SugaredLogger, resourceService ArgoUtil.ResourceService) *CDRestHandlerImpl {
	cdRestHandler := &CDRestHandlerImpl{logger: logger, resourceService: resourceService}
	return cdRestHandler
}

func (handler CDRestHandlerImpl) FetchResourceTree(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app-name"]

	res, err := handler.resourceService.FetchResourceTree(appName)
	if err != nil {
		handler.logger.Errorw("request err, FetchResourceTree", "err", err, "appName", appName)
	}
	resJson, err := json.Marshal(res)
	_, err = w.Write(resJson)
	if err != nil {
		handler.logger.Errorw("request err, FetchResourceTree", "err", err, "appName", appName, "response", res)
	}
}

func (handler CDRestHandlerImpl) FetchPodContainerLogs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appName := vars["app-name"]
	podName := vars["pod-name"]

	res, err := handler.resourceService.FetchPodContainerLogs(appName, podName, ArgoUtil.PodContainerLogReq{})
	if err != nil {
		handler.logger.Errorw("service err, FetchPodContainerLogs", "err", err, "appName", appName, "podName", podName)
	}
	resJson, err := json.Marshal(res)
	_, err = w.Write(resJson)
	if err != nil {
		handler.logger.Errorw("service err, FetchPodContainerLogs", "err", err, "appName", appName, "podName", podName)
	}
}
