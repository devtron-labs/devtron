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
	"github.com/devtron-labs/devtron/api/router/pubsub"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type ExternalCiRestHandler interface {
	HandleExternalCiWebhook(w http.ResponseWriter, r *http.Request)
	HandleExternalCiWebhookByApiToken(w http.ResponseWriter, r *http.Request)
}

type ExternalCiRestHandlerImpl struct {
	logger         *zap.SugaredLogger
	webhookService pipeline.WebhookService
	ciEventHandler pubsub.CiEventHandler
	validator      *validator.Validate
}

func NewExternalCiRestHandlerImpl(logger *zap.SugaredLogger, webhookService pipeline.WebhookService,
	ciEventHandler pubsub.CiEventHandler, validator *validator.Validate) *ExternalCiRestHandlerImpl {
	return &ExternalCiRestHandlerImpl{
		webhookService: webhookService,
		logger:         logger,
		ciEventHandler: ciEventHandler,
		validator:      validator,
	}
}

func (impl ExternalCiRestHandlerImpl) HandleExternalCiWebhook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	vars := mux.Vars(r)
	apiKey := vars["api-key"]
	if apiKey == "" {
		impl.logger.Errorw("request err, HandleExternalCiWebhook", "apiKey", apiKey)
		common.WriteJsonResp(w, errors.New("invalid api-key"), nil, http.StatusBadRequest)
		return
	}

	var req pubsub.CiCompleteEvent
	err := decoder.Decode(&req)
	if err != nil {
		impl.logger.Errorw("request err, HandleExternalCiWebhook", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, HandleExternalCiWebhook", "payload", req)
	ciPipelineId, err := impl.webhookService.AuthenticateExternalCiWebhook(apiKey)
	if err != nil {
		impl.logger.Errorw("auth error", "err", err, "apiKey", apiKey, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	ciArtifactReq, err := impl.ciEventHandler.BuildCiArtifactRequest(req)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	_, err = impl.webhookService.SaveCiArtifactWebhook(ciPipelineId, ciArtifactReq)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (impl ExternalCiRestHandlerImpl) HandleExternalCiWebhookByApiToken(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	vars := mux.Vars(r)
	externalCiId, err := strconv.Atoi(vars["externalCiId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var req pubsub.CiCompleteEvent
	err = decoder.Decode(&req)
	if err != nil {
		impl.logger.Errorw("request err, HandleExternalCiWebhookByApiToken", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, HandleExternalCiWebhookByApiToken", "payload", req)

	ciArtifactReq, err := impl.ciEventHandler.BuildCiArtifactRequestForWebhook(req)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhookByApiToken", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	err = impl.validator.Struct(req)
	if err != nil {
		impl.logger.Errorw("validation err, Create", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	_, err = impl.webhookService.SaveCiArtifactWebhookExternalCi(externalCiId, ciArtifactReq)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, nil, http.StatusOK)
}
