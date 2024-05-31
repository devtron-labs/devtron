/*
 * Copyright (c) 2024. Devtron Inc.
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
