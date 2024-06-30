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

type PipelineConfigEventPublishService interface {
	PublishCDPipelineDelete(pipelineId int, triggeredBy int32) error
}

type PipelineConfigEventPublishServiceImpl struct {
	logger       *zap.SugaredLogger
	pubSubClient *pubsub.PubSubClientServiceImpl
}

func NewPipelineConfigEventPublishServiceImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl) *PipelineConfigEventPublishServiceImpl {
	return &PipelineConfigEventPublishServiceImpl{
		logger:       logger,
		pubSubClient: pubSubClient,
	}

}

func (impl *PipelineConfigEventPublishServiceImpl) PublishCDPipelineDelete(pipelineId int, triggeredBy int32) error {
	impl.logger.Infow("cd pipeline delete event handle", "pipelineId", pipelineId, "triggeredBy", triggeredBy)
	req := &bean.CdPipelineDeleteEvent{
		PipelineId:  pipelineId,
		TriggeredBy: triggeredBy,
	}
	data, err := json.Marshal(req)
	if err != nil {
		impl.logger.Errorw("error while writing cd pipeline delete event to nats", "err", err, "req", req)
		return err
	} else {
		err = impl.pubSubClient.Publish(pubsub.CD_PIPELINE_DELETE_EVENT_TOPIC, string(data))
		if err != nil {
			impl.logger.Errorw("Error while publishing request", "topic", pubsub.CD_PIPELINE_DELETE_EVENT_TOPIC, "error", err)
			return err
		}
	}
	return nil
}
