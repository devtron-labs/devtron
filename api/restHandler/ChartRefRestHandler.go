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
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type ChartRefRestHandler interface {
	ChartRefAutocomplete(w http.ResponseWriter, r *http.Request)
	ChartRefAutocompleteForApp(w http.ResponseWriter, r *http.Request)
	ChartRefAutocompleteForEnv(w http.ResponseWriter, r *http.Request)
}

type ChartRefRestHandlerImpl struct {
	chartService pipeline.ChartService
	logger       *zap.SugaredLogger
}

func NewChartRefRestHandlerImpl(chartService pipeline.ChartService, logger *zap.SugaredLogger) *ChartRefRestHandlerImpl {
	handler := &ChartRefRestHandlerImpl{chartService: chartService, logger: logger}
	return handler
}

func (handler ChartRefRestHandlerImpl) ChartRefAutocomplete(w http.ResponseWriter, r *http.Request) {
	result, err := handler.chartService.ChartRefAutocomplete()
	if err != nil {
		handler.logger.Errorw("service err, ChartRefAutocomplete", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler ChartRefRestHandlerImpl) ChartRefAutocompleteForApp(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, ChartRefAutocompleteForApp", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	result, err := handler.chartService.ChartRefAutocompleteForAppOrEnv(appId, 0)
	if err != nil {
		handler.logger.Errorw("service err, ChartRefAutocompleteForApp", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler ChartRefRestHandlerImpl) ChartRefAutocompleteForEnv(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, ChartRefAutocompleteForEnv", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	environmentId, err := strconv.Atoi(vars["environmentId"])
	if err != nil {
		handler.logger.Errorw("request err, ChartRefAutocompleteForEnv", "err", err, "appId", appId, "environmentId", environmentId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	result, err := handler.chartService.ChartRefAutocompleteForAppOrEnv(appId, environmentId)
	if err != nil {
		handler.logger.Errorw("service err, ChartRefAutocompleteForEnv", "err", err, "appId", appId, "environmentId", environmentId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, result, http.StatusOK)
}
