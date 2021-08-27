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
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type WebhookDataRestHandler interface {
	GetWebhookPayloadDataForPipelineMaterialId(w http.ResponseWriter, r *http.Request)
	GetWebhookPayloadFilterDataForPipelineMaterialId(w http.ResponseWriter, r *http.Request)
}

type WebhookDataRestHandlerImpl struct {
	logger                       *zap.SugaredLogger
	userAuthService              user.UserService
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	enforcerUtil                 rbac.EnforcerUtil
	enforcer                     rbac.Enforcer
	gitSensorClient              gitSensor.GitSensorClient
	webhookEventDataConfig       pipeline.WebhookEventDataConfig
}

func NewWebhookDataRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository, enforcerUtil rbac.EnforcerUtil, enforcer rbac.Enforcer,
	gitSensorClient gitSensor.GitSensorClient, webhookEventDataConfig pipeline.WebhookEventDataConfig) *WebhookDataRestHandlerImpl {
	return &WebhookDataRestHandlerImpl{
		logger:                       logger,
		userAuthService:              userAuthService,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		enforcerUtil:                 enforcerUtil,
		enforcer:                     enforcer,
		gitSensorClient:              gitSensorClient,
		webhookEventDataConfig:       webhookEventDataConfig,
	}
}

func (impl WebhookDataRestHandlerImpl) GetWebhookPayloadDataForPipelineMaterialId(w http.ResponseWriter, r *http.Request) {

	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	pipelineMaterialId, err := strconv.Atoi(vars["pipelineMaterialId"])
	if err != nil {
		impl.logger.Error("can not get pipelineMaterialId from request")
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	impl.logger.Infow("request payload, GetWebhookPayloadDataForPipelineMaterialId", "pipelineMaterialId", pipelineMaterialId)

	ciPipelineMaterial, err := impl.ciPipelineMaterialRepository.GetById(pipelineMaterialId)
	if err != nil {
		impl.logger.Errorw("Error in fetching ciPipelineMaterial", "err", err, "pipelineMaterialId", pipelineMaterialId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//RBAC
	token := r.Header.Get("token")
	object := impl.enforcerUtil.GetAppRBACNameByAppId(ciPipelineMaterial.CiPipeline.AppId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	v := r.URL.Query()
	limit, err := strconv.Atoi(v.Get("limit"))
	if err != nil {
		impl.logger.Error("can not get limit from request")
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	offset, err := strconv.Atoi(v.Get("offset"))
	if err != nil {
		impl.logger.Error("can not get offset from request")
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	eventTimeSortOrder := v.Get("timeSort")

	webhookPayloadDataRequest := &gitSensor.WebhookPayloadDataRequest{
		CiPipelineMaterialId: pipelineMaterialId,
		Limit:                limit,
		Offset:               offset,
		EventTimeSortOrder:   eventTimeSortOrder,
	}

	response, err := impl.gitSensorClient.GetWebhookPayloadDataForPipelineMaterialId(webhookPayloadDataRequest)
	if err != nil {
		impl.logger.Errorw("service err, GetWebhookPayloadDataForPipelineMaterialId", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, nil, response, http.StatusOK)

}

func (impl WebhookDataRestHandlerImpl) GetWebhookPayloadFilterDataForPipelineMaterialId(w http.ResponseWriter, r *http.Request) {

	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	pipelineMaterialId, err := strconv.Atoi(vars["pipelineMaterialId"])
	if err != nil {
		impl.logger.Error("can not get pipelineMaterialId from request")
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	impl.logger.Infow("request payload, GetWebhookPayloadDataForPipelineMaterialId", "pipelineMaterialId", pipelineMaterialId)

	ciPipelineMaterial, err := impl.ciPipelineMaterialRepository.GetById(pipelineMaterialId)
	if err != nil {
		impl.logger.Errorw("Error in fetching ciPipelineMaterial", "err", err, "pipelineMaterialId", pipelineMaterialId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//RBAC
	token := r.Header.Get("token")
	object := impl.enforcerUtil.GetAppRBACNameByAppId(ciPipelineMaterial.CiPipeline.AppId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC

	parsedDataId, err := strconv.Atoi(vars["parsedDataId"])
	if err != nil {
		impl.logger.Error("can not get parsedDataId from request")
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	webhookPayloadFilterDataRequest := &gitSensor.WebhookPayloadFilterDataRequest{
		CiPipelineMaterialId: pipelineMaterialId,
		ParsedDataId:         parsedDataId,
	}

	response, err := impl.gitSensorClient.GetWebhookPayloadFilterDataForPipelineMaterialId(webhookPayloadFilterDataRequest)
	if err != nil {
		impl.logger.Errorw("service err, GetWebhookPayloadFilterDataForPipelineMaterialId", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// set payload json
	if response != nil && response.PayloadId != 0 {
		webhookEventData, err := impl.webhookEventDataConfig.GetById(response.PayloadId)
		if err != nil {
			impl.logger.Errorw("error in getting webhook payload data", "err", err)
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}

		if webhookEventData != nil {
			response.PayloadJson = webhookEventData.RequestPayloadJson
		}
	}

	writeJsonResp(w, nil, response, http.StatusOK)

}
