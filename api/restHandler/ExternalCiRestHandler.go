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
	"github.com/devtron-labs/devtron/api/router/pubsub"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
)

type ExternalCiRestHandler interface {
	HandleExternalCiWebhook(w http.ResponseWriter, r *http.Request)
}

type ExternalCiRestHandlerImpl struct {
	logger         *zap.SugaredLogger
	webhookService pipeline.WebhookService
	ciEventHandler pubsub.CiEventHandler
}

func NewExternalCiRestHandlerImpl(logger *zap.SugaredLogger, webhookService pipeline.WebhookService, ciEventHandler pubsub.CiEventHandler) *ExternalCiRestHandlerImpl {
	return &ExternalCiRestHandlerImpl{
		webhookService: webhookService,
		logger:         logger,
		ciEventHandler: ciEventHandler,
	}
}

func (impl ExternalCiRestHandlerImpl) HandleExternalCiWebhook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	vars := mux.Vars(r)
	apiKey := vars["api-key"]
	if apiKey == "" {
		impl.logger.Errorw("request err, HandleExternalCiWebhook", "apiKey", apiKey)
		writeJsonResp(w, errors.New("invalid api-key"), nil, http.StatusBadRequest)
		return
	}

	var req pubsub.CiCompleteEvent
	err := decoder.Decode(&req)
	if err != nil {
		impl.logger.Errorw("request err, HandleExternalCiWebhook", "err", err, "payload", req)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, HandleExternalCiWebhook", "payload", req)

	ciPipelineId, err := impl.webhookService.AuthenticateExternalCiWebhook(apiKey)
	if err != nil {
		impl.logger.Errorw("auth error", "err", err, "apiKey", apiKey, "payload", req)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	ciArtifactReq, err := impl.ciEventHandler.BuildCiArtifactRequest(req)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", req)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	_, err = impl.webhookService.SaveCiArtifactWebhook(ciPipelineId, ciArtifactReq)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", req)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, nil, http.StatusOK)
}
