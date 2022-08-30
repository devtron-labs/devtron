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
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	devtron_integration_manager "github.com/devtron-labs/devtron/pkg/devtron-integration-manager"
	"go.uber.org/zap"
	"net/http"
)

type IntegrationManagerRestHandler interface {
	InstallModule(w http.ResponseWriter, r *http.Request)
	GetAllModules(w http.ResponseWriter, r *http.Request)
	GetModulesStatus(w http.ResponseWriter, r *http.Request)
}

type IntegrationManagerRestHandlerImpl struct {
	logger                    *zap.SugaredLogger
	integrationManagerService devtron_integration_manager.IntegrationManagerService
}

func NewAttributesRestHandlerImpl(logger *zap.SugaredLogger, integrationManagerService devtron_integration_manager.IntegrationManagerService) *IntegrationManagerRestHandlerImpl {
	integrationManagerHandler := &IntegrationManagerRestHandlerImpl{
		logger:                    logger,
		integrationManagerService: integrationManagerService,
	}
	return integrationManagerHandler
}

func (handler IntegrationManagerRestHandlerImpl) InstallModule(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var payload map[string]interface{}
	err := decoder.Decode(&payload)
	if err != nil {
		handler.logger.Errorw("request err, InstallModule", "err", err, "payload", payload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	moduleNameVal := payload["moduleName"]
	userIdVal := payload["userId"]

	moduleName := moduleNameVal.(string)
	userId := userIdVal.(int32)

	handler.logger.Infow("request payload, InstallModule", "payload", payload)
	err = handler.integrationManagerService.InstallModule(userId, moduleName)
	if err != nil {
		handler.logger.Errorw("service err, InstallModule", "err", err, "payload", payload)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "success", http.StatusOK)
}

func (handler IntegrationManagerRestHandlerImpl) GetAllModules(w http.ResponseWriter, r *http.Request) {

	resp, err := handler.integrationManagerService.GetAllModules()
	if err != nil {
		handler.logger.Errorw("service err, GetAllModules", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler IntegrationManagerRestHandlerImpl) GetModulesStatus(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var payload map[string]interface{}
	err := decoder.Decode(&payload)
	if err != nil {
		handler.logger.Errorw("request err, GetModulesStatus", "err", err, "payload", payload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	moduleNamesVal := payload["moduleNames"]
	moduleNames := moduleNamesVal.([]string)

	res, err := handler.integrationManagerService.GetModulesStatus(moduleNames)
	if err != nil {
		handler.logger.Errorw("service err, GetModulesStatus", "err", err, "payload", payload)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
