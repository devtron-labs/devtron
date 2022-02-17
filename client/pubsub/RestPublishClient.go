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
	"time"

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

//TODO : adhiran : We need to define stream names and subjects which will be published under those streams. Will also help in binding
func (impl *NatsPublishClientImpl) Publish(req *PublishRequest) error {
	streamInfo, strInfoErr := impl.pubSubClient.JetStrCtxt.StreamInfo(req.Topic)
	if strInfoErr != nil {
		impl.logger.Errorw("Error while getting stream info", "topic", req.Topic, "error", strInfoErr)
	}
	if streamInfo == nil {
		//Stream doesn't already exist. Create a new stream from jetStreamContext
		_, addStrError := impl.pubSubClient.JetStrCtxt.AddStream(&nats.StreamConfig{
			Name:     req.Topic,
			Subjects: []string{req.Topic + ".*"},
		})
		if addStrError != nil {
			impl.logger.Errorw("Error while creating stream", "topic", req.Topic, "error", addStrError)
		}
	}

	//Generate random string for passing as Header Id in message
	randString := "MsgHeaderId-" + util.Generate(10)
	_, publishErr := impl.pubSubClient.JetStrCtxt.PublishAsync(req.Topic, req.Payload, nats.MsgId(randString))
	if publishErr != nil {
		impl.logger.Errorw("Error while publishing asyncRequest", "topic", req.Topic, "body", string(req.Payload), "err", publishErr)
	}
	//TODO : adhiran : Need to find out why we need below select case
	select {
	case <-impl.pubSubClient.JetStrCtxt.PublishAsyncComplete():
	case <-time.After(5 * time.Second):
		impl.logger.Errorw("Did not resolve in time")
	}
	return publishErr
}
