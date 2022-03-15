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

package pubsub

import (
	"encoding/json"

	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/util"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type NatsPublishClient interface {
	Publish(req *PublishRequest) error
}

func NewNatsPublishClientImpl(logger *zap.SugaredLogger, pubSubClient *PubSubClient) *NatsPublishClientImpl {
	return &NatsPublishClientImpl{
		logger:       logger,
		pubSubClient: pubSubClient,
	}
}

type NatsPublishClientImpl struct {
	logger       *zap.SugaredLogger
	pubSubClient *PubSubClient
}

type PublishRequest struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

func (impl *NatsPublishClientImpl) Publish(req *PublishRequest) error {
	streamInfo, err := impl.pubSubClient.JetStrCtxt.StreamInfo(constants.ORCHESTRATOR_STREAM)
	if err != nil {
		impl.logger.Errorw("Error while getting stream info", "stream name", constants.ORCHESTRATOR_STREAM, "error", err)
	}
	if streamInfo == nil {
		//Stream doesn't already exist. Create a new stream from jetStreamContext
		_, err = impl.pubSubClient.JetStrCtxt.AddStream(&nats.StreamConfig{
			Name:     constants.ORCHESTRATOR_STREAM,
			Subjects: []string{constants.ORCHESTRATOR_STREAM + ".*"},
		})
		if err != nil {
			impl.logger.Errorw("Error while creating stream", "Stream name", constants.ORCHESTRATOR_STREAM, "error", err)
			return err
		}
	}

	//Generate random string for passing as Header Id in message
	randString := "MsgHeaderId-" + util.Generate(10)
	_, err = impl.pubSubClient.JetStrCtxt.Publish(req.Topic, req.Payload, nats.MsgId(randString))
	if err != nil {
		impl.logger.Errorw("Error while publishing Request", "topic", req.Topic, "body", string(req.Payload), "err", err)
	}

	return err
}
