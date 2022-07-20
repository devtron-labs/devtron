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
	Publish(req *PublishRequest) error
}

func NewNatsPublishClientImpl(logger *zap.SugaredLogger) *NatsPublishClientImpl {
	return &NatsPublishClientImpl{
		logger: logger,
	}
}

type NatsPublishClientImpl struct {
	logger *zap.SugaredLogger
}

type PublishRequest struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

//TODO : adhiran : check the req.topic. We dont have dynamic topics listed in stream subjects arrary.So this might fail in
//subscription if the subject name passed is not listed
func (impl *NatsPublishClientImpl) Publish(req *PublishRequest) error {

	return nil
}
