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

package chartProvider

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/appStore/chartProvider"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type ChartProviderRestHandler interface {
	GetChartProviderList(w http.ResponseWriter, r *http.Request)
	ToggleChartProvider(w http.ResponseWriter, r *http.Request)
	SyncChartProvider(w http.ResponseWriter, r *http.Request)
}

type ChartProviderRestHandlerImpl struct {
	Logger               *zap.SugaredLogger
	chartProviderService chartProvider.ChartProviderService
	validator       *validator.Validate
	userAuthService user.UserService
	enforcer        casbin.Enforcer
}

func NewChartProviderRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService, validator *validator.Validate, chartProviderService chartProvider.ChartProviderService,
	enforcer casbin.Enforcer) *ChartProviderRestHandlerImpl {
	return &ChartProviderRestHandlerImpl{
		Logger:               Logger,
		validator:            validator,
		chartProviderService: chartProviderService,
		userAuthService:      userAuthService,
		enforcer:             enforcer,
	}
}

func (handler *ChartProviderRestHandlerImpl) GetChartProviderList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	handler.Logger.Infow("request payload, GetChartProviderList", "userId", userId)

	res, err := handler.chartProviderService.GetChartProviderList()
	if err != nil {
		handler.Logger.Errorw("service err, GetChartProviderList", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *ChartProviderRestHandlerImpl) ToggleChartProvider(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	var request chartProvider.ChartProviderRequestDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, ToggleChartProvider", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, ToggleChartProvider", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, ToggleChartProvider", "payload", request, "userId", userId)
	token := r.Header.Get("token")
	//RBAC starts
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		handler.Logger.Infow("user forbidden to toggle chart provider", "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC ends
	request.UserId = userId
	err = ValidateRequestObjectForChartRepoId(&request)
	if err != nil {
		handler.Logger.Errorw("request err, ToggleChartProvider", "err", err, "ChartRepoId", request.Id)
		common.WriteJsonResp(w, err, "Invalid ChartRepoId", http.StatusBadRequest)
		return
	}
	err = handler.chartProviderService.ToggleChartProvider(&request)
	if err != nil {
		handler.Logger.Errorw("service err, ToggleChartProvider", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (handler *ChartProviderRestHandlerImpl) SyncChartProvider(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	var request chartProvider.ChartProviderRequestDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, SyncChartProvider", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, SyncChartProvider", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, SyncChartProvider", "payload", request, "userId", userId)
	token := r.Header.Get("token")
	//RBAC starts
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		handler.Logger.Infow("user forbidden to sync chart provider", "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC ends
	request.UserId = userId
	err = ValidateRequestObjectForChartRepoId(&request)
	if err != nil {
		handler.Logger.Errorw("request err, ToggleChartProvider", "err", err, "ChartRepoId", request.Id)
		common.WriteJsonResp(w, err, "Invalid ChartRepoId", http.StatusBadRequest)
		return
	}
	err = handler.chartProviderService.SyncChartProvider(&request)
	if err != nil {
		handler.Logger.Errorw("service err, SyncChartProvider", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func ValidateRequestObjectForChartRepoId(request *chartProvider.ChartProviderRequestDto) error {
	if !request.IsOCIRegistry {
		chartRepoId, err := strconv.Atoi(request.Id)
		if err != nil || chartRepoId <= 0 {
			return err
		}
		request.ChartRepoId = chartRepoId
	}
	return nil
}
