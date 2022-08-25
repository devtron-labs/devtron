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
	pubsub_lib "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type PubSubClient struct {
	logger     *zap.SugaredLogger
	JetStrCtxt nats.JetStreamContext
	Conn       nats.Conn
}

type PubSubConfig struct {
	NatsServerHost string `env:"NATS_SERVER_HOST" envDefault:"nats://localhost:4222"`
}

/* #nosec */
func NewPubSubClient(puSubClientServiceImpl *pubsub_lib.PubSubClientServiceImpl) (*PubSubClient, error) {

	natsClient := &PubSubClient{
		logger:     puSubClientServiceImpl.Logger,
		JetStrCtxt: puSubClientServiceImpl.NatsClient.JetStrCtxt,
	}
	return natsClient, nil
}
