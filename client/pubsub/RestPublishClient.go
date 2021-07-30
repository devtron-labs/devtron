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
	"go.uber.org/zap"
)

type NatsPublishClient interface {
	Publish(req *PublishRequest) (string, error)
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

func (impl *NatsPublishClientImpl) Publish(req *PublishRequest) (string, error) {
	id, err := impl.pubSubClient.Conn.PublishAsync(req.Topic, req.Payload, func(s string, err error) {
		if err != nil {
			impl.logger.Errorw("error in publishing msg ", "topic", req.Topic, "body", string(req.Payload), "err", err)
		}
	})
	if err != nil {
		impl.logger.Errorw("error in publishing msg submit", "topic", req.Topic, "body", string(req.Payload), "err", err)
	}
	return id, err
}
