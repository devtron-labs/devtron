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
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/git"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
)

type WebhookEventHandler interface {
	OnWebhookEvent(w http.ResponseWriter, r *http.Request)
}

type WebhookEventHandlerImpl struct {
	logger                 *zap.SugaredLogger
	gitHostConfig          pipeline.GitHostConfig
	eventClient            client.EventClient
	webhookSecretValidator git.WebhookSecretValidator
	webhookEventDataConfig pipeline.WebhookEventDataConfig
}

func NewWebhookEventHandlerImpl(logger *zap.SugaredLogger, gitHostConfig pipeline.GitHostConfig, eventClient client.EventClient,
	webhookSecretValidator git.WebhookSecretValidator, webhookEventDataConfig pipeline.WebhookEventDataConfig) *WebhookEventHandlerImpl {
	return &WebhookEventHandlerImpl{
		logger:                 logger,
		gitHostConfig:          gitHostConfig,
		eventClient:            eventClient,
		webhookSecretValidator: webhookSecretValidator,
		webhookEventDataConfig: webhookEventDataConfig,
	}
}

func (impl WebhookEventHandlerImpl) OnWebhookEvent(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debug("webhook event came")

	// get git host Id and secret from request
	vars := mux.Vars(r)
	gitHostId, err := strconv.Atoi(vars["gitHostId"])
	secretFromRequest := vars["secret"]
	if err != nil {
		impl.logger.Errorw("Error in getting git host Id from request", "err", err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	impl.logger.Debug("gitHostId", gitHostId)
	impl.logger.Debug("secretFromRequest", secretFromRequest)

	// get git host from DB
	gitHost, err := impl.gitHostConfig.GetById(gitHostId)
	if err != nil {
		impl.logger.Errorw("Error in getting git host from DB", "err", err, "gitHostId", gitHostId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// validate signature
	requestBodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		impl.logger.Errorw("Cannot read the request body:", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	isValidSig := impl.webhookSecretValidator.ValidateSecret(r, secretFromRequest, requestBodyBytes, gitHost)
	impl.logger.Debug("Secret validation result ", isValidSig)
	if !isValidSig {
		impl.logger.Error("Signature mismatch")
		writeJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	// validate event type
	eventType := r.Header.Get(gitHost.EventTypeHeader)
	impl.logger.Debugw("eventType : ", eventType)
	if len(eventType) == 0 {
		impl.logger.Errorw("Event type not known ", eventType)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// make request to handle this webhook
	webhookEvent := &pipeline.WebhookEventDataRequest{
		GitHostId:          gitHostId,
		EventType:          eventType,
		RequestPayloadJson: string(requestBodyBytes),
	}

	// save in DB
	err = impl.webhookEventDataConfig.Save(webhookEvent)
	if err != nil {
		impl.logger.Errorw("Error while saving webhook data", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// write event
	err = impl.eventClient.WriteNatsEvent(pubsub.WEBHOOK_EVENT_TOPIC, webhookEvent)
	if err != nil {
		impl.logger.Errorw("Error while handling webhook in git-sensor", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
	}
}
