/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package out

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/out/bean"
	"go.uber.org/zap"
)

type CIPipelineEventPublishService interface {
	PublishGitWebhookEvent(gitHostId int, eventType, requestJSON string) error
}

type CIPipelineEventPublishServiceImpl struct {
	logger       *zap.SugaredLogger
	pubSubClient *pubsub.PubSubClientServiceImpl
}

func NewCIPipelineEventPublishServiceImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl) *CIPipelineEventPublishServiceImpl {
	return &CIPipelineEventPublishServiceImpl{
		logger:       logger,
		pubSubClient: pubSubClient,
	}
}

func (impl *CIPipelineEventPublishServiceImpl) PublishGitWebhookEvent(gitHostId int, eventType, requestJSON string) error {
	event := &bean.CIPipelineGitWebhookEvent{
		GitHostId:          gitHostId,
		EventType:          eventType,
		RequestPayloadJson: requestJSON,
	}
	body, err := json.Marshal(event)
	if err != nil {
		impl.logger.Errorw("error in marshaling git webhook event", "err", err, "event", event)
		return err
	}
	err = impl.pubSubClient.Publish(pubsub.WEBHOOK_EVENT_TOPIC, string(body))
	if err != nil {
		impl.logger.Errorw("error in publishing git webhook event", "err", err, "eventBody", body)
		return err
	}
	return nil
}
