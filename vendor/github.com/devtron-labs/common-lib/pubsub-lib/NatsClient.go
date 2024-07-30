/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package pubsub_lib

import (
	"github.com/caarlos0/env"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"sync"
	"time"
)

type NatsClient struct {
	logger     *zap.SugaredLogger
	JetStrCtxt nats.JetStreamContext
	Conn       *nats.Conn
	ConnWg     *sync.WaitGroup
}

type NatsClientConfig struct {
	NatsServerHost string `env:"NATS_SERVER_HOST" envDefault:"nats://devtron-nats.devtroncd:4222"`

	// consumer wise
	// NatsMsgProcessingBatchSize is the number of messages that will be processed in one go
	NatsMsgProcessingBatchSize int `env:"NATS_MSG_PROCESSING_BATCH_SIZE" envDefault:"1"`

	// NatsMsgBufferSize is the number of messages that will be buffered in memory (channel size)
	// it is recommended to set this value equal to NatsMsgProcessingBatchSize as we want to process maximum messages in the buffer in one go.
	// Note: if NatsMsgBufferSize is less than NatsMsgProcessingBatchSize
	// then the wait time for the unprocessed messages in the buffer will be high.(total process time = life-time in buffer + processing time)
	// NatsMsgBufferSize can be configured independently of NatsMsgProcessingBatchSize if needed by setting its value to positive value in env.
	// if NatsMsgBufferSize set to a non-positive value then it will take the value of NatsMsgProcessingBatchSize.
	// Note: always get this value by calling GetNatsMsgBufferSize method
	NatsMsgBufferSize    int `env:"NATS_MSG_BUFFER_SIZE" envDefault:"-1"`
	NatsMsgMaxAge        int `env:"NATS_MSG_MAX_AGE" envDefault:"86400"`
	NatsMsgAckWaitInSecs int `env:"NATS_MSG_ACK_WAIT_IN_SECS" envDefault:"120"`
	NatsMsgReplicas      int `env:"NATS_MSG_REPLICAS" envDefault:"0"`
}

func (ncc NatsClientConfig) GetNatsMsgBufferSize() int {
	// if NatsMsgBufferSize is set to a non-positive value then it will take the value of NatsMsgProcessingBatchSize.
	if ncc.NatsMsgBufferSize <= 0 {
		return ncc.NatsMsgProcessingBatchSize
	}
	return ncc.NatsMsgBufferSize
}

func (ncc NatsClientConfig) GetDefaultNatsConsumerConfig() NatsConsumerConfig {
	return NatsConsumerConfig{
		NatsMsgProcessingBatchSize: ncc.NatsMsgProcessingBatchSize,
		NatsMsgBufferSize:          ncc.GetNatsMsgBufferSize(),
		AckWaitInSecs:              ncc.NatsMsgAckWaitInSecs,
		//Replicas:                   ncc.Replicas,
	}
}

func (ncc NatsClientConfig) GetDefaultNatsStreamConfig() NatsStreamConfig {
	return NatsStreamConfig{
		StreamConfig: StreamConfig{
			MaxAge:   time.Duration(ncc.NatsMsgMaxAge) * time.Second,
			Replicas: ncc.NatsMsgReplicas,
		},
	}
}

type StreamConfig struct {
	MaxAge time.Duration `json:"max_age"`
	//it will show the instances created for the consumers on a particular subject(topic)
	Replicas int `json:"num_replicas"`
}

type NatsStreamConfig struct {
	StreamConfig StreamConfig `json:"streamConfig"`
}

type NatsConsumerConfig struct {
	// NatsMsgProcessingBatchSize is the number of messages that will be processed in one go
	NatsMsgProcessingBatchSize int `json:"natsMsgProcessingBatchSize"`
	// NatsMsgBufferSize is the number of messages that will be buffered in memory (channel size).
	// Note: always get this value by calling GetNatsMsgBufferSize method
	NatsMsgBufferSize int `json:"natsMsgBufferSize"`
	// AckWaitInSecs is the time in seconds for which the message can be in unacknowledged state
	AckWaitInSecs int `json:"ackWaitInSecs"`
}

func (consumerConf NatsConsumerConfig) GetNatsMsgBufferSize() int {
	// if NatsMsgBufferSize is set to a non-positive value then it will take the value of NatsMsgProcessingBatchSize.
	if consumerConf.NatsMsgBufferSize <= 0 {
		return consumerConf.NatsMsgProcessingBatchSize
	}
	return consumerConf.NatsMsgBufferSize
}

// func (consumerConf NatsConsumerConfig) GetNatsMsgProcessingBatchSize() int {
// 	if nbs := consumerConf.GetNatsMsgBufferSize(); nbs < consumerConf.NatsMsgProcessingBatchSize {
// 		return nbs
// 	}
// 	return consumerConf.NatsMsgProcessingBatchSize
// }

func NewNatsClient(logger *zap.SugaredLogger) (*NatsClient, error) {
	//connWg := new(sync.WaitGroup)
	//connWg.Add(1)
	cfg := &NatsClientConfig{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Error("error occurred while parsing nats client config", "err", err)
		return &NatsClient{}, err
	}

	// Connect to NATS
	nc, err := nats.Connect(cfg.NatsServerHost,
		// Because draining can involve messages flowing to the server, for a flush and asynchronous message processing,
		// the timeout for drain should generally be higher than the timeout for a simple message request-reply or similar.
		nats.DrainTimeout(time.Duration(cfg.NatsMsgAckWaitInSecs)*time.Second),
		nats.ReconnectWait(10*time.Second), nats.MaxReconnects(100),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			logger.Errorw("Nats Connection got disconnected!", "Reason", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			logger.Infow("Nats Connection got reconnected", "url", nc.ConnectedUrl())
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			logger.Errorw("Nats Client Connection closed!", "Reason", nc.LastError())
			//connWg.Done()
		}))
	if err != nil {
		logger.Error("err", err)
		return &NatsClient{}, err
	}

	// Create a jetstream context
	js, err := nc.JetStream()

	if err != nil {
		logger.Errorw("Error while creating jetstream context", "error", err)
	}

	natsClient := &NatsClient{
		logger:     logger,
		JetStrCtxt: js,
		Conn:       nc,
		//ConnWg:     connWg,
	}
	return natsClient, nil
}
