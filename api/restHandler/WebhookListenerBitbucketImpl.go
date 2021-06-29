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
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

type WebhookListenerBitbucketImpl struct {
	logger *zap.SugaredLogger
	gitHostConfig    	 pipeline.GitHostConfig
	eventClient          client.EventClient
}

func NewWebhookListenerBitbucketImpl(logger *zap.SugaredLogger, gitHostConfig pipeline.GitHostConfig, eventClient client.EventClient) *WebhookListenerBitbucketImpl {
	return &WebhookListenerBitbucketImpl{
		logger: logger,
		gitHostConfig: gitHostConfig,
		eventClient: eventClient,
	}
}

func (impl WebhookListenerBitbucketImpl) OnWebhookEvent(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debug("bitbucket webhook event came")

	vars := mux.Vars(r)
	secretInRequest := vars["secret"]

	// get secret from DB
	secret, err := impl.gitHostConfig.GetGitHostSecretByName(pipeline.GIT_HOST_NAME_BITBUCKET_CLOUD)
	if err != nil {
		impl.logger.Errorw("Error in getting bitbucket cloud host secret", "err", err)
		return
	}

	// validate secret
	if secretInRequest != secret {
		impl.logger.Error("secret is not matching")
		return
	}

	// validate supported Event type
	eventType := r.Header.Get("X-Event-Key")
	impl.logger.Debugw("eventType : ", eventType)
	if !supportedBitbucketEventType(eventType){
		impl.logger.Errorw("Event type not supported ", eventType)
		return
	}

	requestBodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		impl.logger.Errorw("Cannot read the request body:", "err", err)
		return
	}

	// make request to nats to handle this webhook
	webhookEvent := WebhookEvent{
		RequestPayloadJson:         string(requestBodyBytes),
		GitHostType:      			GitHostType(pipeline.GIT_HOST_NAME_BITBUCKET_CLOUD),
		WebhookEventType:        	WebhookEventType(eventType),
	}

	// write event
	err = impl.eventClient.WriteNatsEvent(pubsub.WEBHOOK_EVENT_TOPIC, webhookEvent)
	if err != nil {
		impl.logger.Errorw("Error while handling webhook in git-sensor", "err", err)
	}

}

func supportedBitbucketEventType(eventTypeRequest string) bool {
	supportedEventTypes := []string {
		"pullrequest:created",
		"pullrequest:updated",
		"pullrequest:changes_request_created",
		"pullrequest:changes_request_removed",
		"pullrequest:approved",
		"pullrequest:unapproved",
		"pullrequest:fulfilled",
		"pullrequest:rejected",
		"pullrequest:comment_created",
		"pullrequest:comment_updated",
		"pullrequest:comment_deleted",
	}
	for _, eventType := range supportedEventTypes {
		if eventType == eventTypeRequest  {
			return true
		}
	}
	return false
}



