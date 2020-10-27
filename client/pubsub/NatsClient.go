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
	"github.com/caarlos0/env"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan"
	"go.uber.org/zap"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"
)

type PubSubClient struct {
	logger      *zap.SugaredLogger
	Conn        stan.Conn
	AckDuration int
}

type PubSubConfig struct {
	NatsServerHost string `env:"NATS_SERVER_HOST" envDefault:"nats://localhost:4222"`
	ClusterId      string `env:"CLUSTER_ID" envDefault:"devtron-stan"`
	ClientId       string `env:"CLIENT_ID" envDefault:"orchestrator"`
	AckDuration    string `env:"ACK_DURATION" envDefault:"30"`
}

const CD_SUCCESS = "ORCHESTRATOR.CD.TRIGGER"
/* #nosec */
func NewPubSubClient(logger *zap.SugaredLogger) (*PubSubClient, error) {

	cfg := &PubSubConfig{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Error("err", err)
		return &PubSubClient{}, err
	}

	nc, err := nats.Connect(cfg.NatsServerHost, nats.ReconnectWait(10*time.Second), nats.MaxReconnects(100))
	if err != nil {
		logger.Error("err", err)
		return &PubSubClient{}, err
	}


	s := rand.NewSource(time.Now().UnixNano())
	uuid := rand.New(s)
	uniqueClienId := "orchestrator-" + strconv.Itoa(uuid.Int())

	sc, err := stan.Connect(cfg.ClusterId, uniqueClienId, stan.NatsConn(nc))
	if err != nil {
		log.Println("err", err)
		os.Exit(1)
	}
	ack, err := strconv.Atoi(cfg.AckDuration)
	if err != nil {
		log.Println("err", err)
		os.Exit(1)
	}
	natsClient := &PubSubClient{
		logger:      logger,
		Conn:        sc,
		AckDuration: ack,
	}
	return natsClient, nil
}
