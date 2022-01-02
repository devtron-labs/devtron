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
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"net/http"
)

type KubeCapacityRestHandler interface {
	KubeCapacityDefault(w http.ResponseWriter, r *http.Request)
	KubeCapacityPods(w http.ResponseWriter, r *http.Request)
	KubeCapacityUtilization(w http.ResponseWriter, r *http.Request)
	AvailableResources(w http.ResponseWriter, r *http.Request)
	PodsAndUtil(w http.ResponseWriter, r *http.Request)
}
type KubeCapacityRestHandlerImpl struct {
	logger             *zap.SugaredLogger
}

func NewKubeCapacityRestHandlerImpl(logger *zap.SugaredLogger) *KubeCapacityRestHandlerImpl {
	return &KubeCapacityRestHandlerImpl{
		logger:             logger,
	}
}

const (
	KubeCapacity = "kube-capacity --output json"
	KubeCapacityPods = "kube-capacity --pods --output json"
	KubeCapacityUtilization = "kube-capacity --util --output json"
	AvailableResources = "kube-capacity --available --output json"
	PodsAndUtil = "kube-capacity --pods --util --output json"
)

func (impl KubeCapacityRestHandlerImpl) KubeCapacityDefault(w http.ResponseWriter, r *http.Request) {
	res, err := util.KubeCapacity(KubeCapacity)
	if err != nil {
		impl.logger.Errorw("err in execute command, PodsAndUtil", "err: ", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (impl KubeCapacityRestHandlerImpl) KubeCapacityPods(w http.ResponseWriter, r *http.Request) {
	res, err := util.KubeCapacity(KubeCapacityPods)
	if err != nil {
		impl.logger.Errorw("err in execute command, PodsAndUtil", "err: ", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}


func (impl KubeCapacityRestHandlerImpl) KubeCapacityUtilization(w http.ResponseWriter, r *http.Request) {
	res, err := util.KubeCapacity(KubeCapacityUtilization)
	if err != nil {
		impl.logger.Errorw("err in execute command, PodsAndUtil", "err: ", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (impl KubeCapacityRestHandlerImpl) AvailableResources(w http.ResponseWriter, r *http.Request) {
	res, err := util.KubeCapacity(AvailableResources)
	if err != nil {
		impl.logger.Errorw("err in execute command, PodsAndUtil", "err: ", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (impl KubeCapacityRestHandlerImpl) PodsAndUtil(w http.ResponseWriter, r *http.Request){
	res, err := util.KubeCapacity(PodsAndUtil)
	if err != nil {
		impl.logger.Errorw("err in execute command, PodsAndUtil", "err: ", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
