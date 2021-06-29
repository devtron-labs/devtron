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
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
)

type WebhookListenerGithubImpl struct {
	logger *zap.SugaredLogger
	gitHostConfig    	 pipeline.GitHostConfig
	eventClient          client.EventClient
}

func NewWebhookListenerGithubImpl(logger *zap.SugaredLogger, gitHostConfig pipeline.GitHostConfig,
	eventClient client.EventClient) *WebhookListenerGithubImpl {
	return &WebhookListenerGithubImpl{
		logger: logger,
		gitHostConfig: gitHostConfig,
		eventClient: eventClient,
	}
}

func (impl WebhookListenerGithubImpl) OnWebhookEvent(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debug("github webhook event came")

	// validate secret
	secret, err := impl.gitHostConfig.GetGitHostSecretByName(pipeline.GIT_HOST_NAME_GITHUB)
	if err != nil {
		impl.logger.Errorw("Error in getting github host secret", "err", err)
		return
	}
	requestBodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		impl.logger.Errorw("Cannot read the request body:", "err", err)
		return
	}
	isValidSig := IsValidSignature(r, secret, requestBodyBytes, impl.logger)
	if !isValidSig {
		impl.logger.Error("Signature mismatch")
		return
	}

	// validate supported Event type
	eventType := r.Header.Get("X-GitHub-Event")
	impl.logger.Debugw("eventType : ", eventType)
	if !supportedGitGubEventType(eventType){
		impl.logger.Errorw("Event type not supported ", eventType)
		return
	}

	// make request to nats to handle this webhook
	webhookEvent := WebhookEvent{
		RequestPayloadJson:         string(requestBodyBytes),
		GitHostType:      			GitHostType(pipeline.GIT_HOST_NAME_GITHUB),
		WebhookEventType:        	WebhookEventType(eventType),
	}

	// write event
	err = impl.eventClient.WriteNatsEvent(pubsub.WEBHOOK_EVENT_TOPIC, webhookEvent)
	if err != nil {
		impl.logger.Errorw("Error while handling webhook in git-sensor", "err", err)
	}
}

func supportedGitGubEventType(eventTypeRequest string) bool {
	supportedEventTypes := []string {
		"pull_request",
	}
	for _, eventType := range supportedEventTypes {
		if eventType == eventTypeRequest  {
			return true
		}
	}
	return false
}

func IsValidSignature(r *http.Request, secretKey string, requestBodyBytes []byte, logger *zap.SugaredLogger) bool {
	gotHash := strings.SplitN(r.Header.Get("X-Hub-Signature"), "=", 2)
	if gotHash[0] != "sha1" {
		return false
	}

	hash := hmac.New(sha1.New, []byte(secretKey))
	if _, err := hash.Write(requestBodyBytes); err != nil {
		logger.Errorw("Cannot compute the HMAC for request:", "err", err)
		return false
	}

	expectedHash := hex.EncodeToString(hash.Sum(nil))
	return gotHash[1] == expectedHash
}