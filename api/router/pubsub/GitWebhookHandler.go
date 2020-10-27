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
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/git"
	"encoding/json"
	"github.com/nats-io/stan"
	"go.uber.org/zap"
	"time"
)

type GitWebhookHandler interface {
	Subscribe() error
}

type GitWebhookHandlerImpl struct {
	logger            *zap.SugaredLogger
	pubsubClient      *pubsub.PubSubClient
	gitWebhookService git.GitWebhookService
}

const newCiMaterialTopic = "GIT-SENSOR.NEW-CI-MATERIAL"
const newCiMaterialTopicGroup = "GIT-SENSOR.NEW-CI-MATERIAL_GROUP-1"
const newCiMaterialTopicDurable = "GIT-SENSOR.NEW-CI-MATERIAL_DURABLE-1"

func NewGitWebhookHandler(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClient, gitWebhookService git.GitWebhookService) *GitWebhookHandlerImpl {
	gitWebhookHandlerImpl := &GitWebhookHandlerImpl{
		logger:            logger,
		pubsubClient:      pubsubClient,
		gitWebhookService: gitWebhookService,
	}
	err := gitWebhookHandlerImpl.Subscribe()
	if err != nil {
		logger.Error("err", err)
		return nil
	}
	return gitWebhookHandlerImpl
}

func (impl *GitWebhookHandlerImpl) Subscribe() error {
	_, err := impl.pubsubClient.Conn.QueueSubscribe(newCiMaterialTopic, newCiMaterialTopicGroup, func(msg *stan.Msg) {
		defer msg.Ack()
		ciPipelineMaterial := gitSensor.CiPipelineMaterial{}
		err := json.Unmarshal([]byte(string(msg.Data)), &ciPipelineMaterial)
		if err != nil {
			impl.logger.Error("err", err)
			return
		}
		resp, err := impl.gitWebhookService.HandleGitWebhook(ciPipelineMaterial)
		impl.logger.Debug(resp)
		if err != nil {
			impl.logger.Error("err", err)
			return
		}
	}, stan.DurableName(newCiMaterialTopicDurable), stan.StartWithLastReceived(), stan.AckWait(time.Duration(impl.pubsubClient.AckDuration)*time.Second), stan.SetManualAckMode(), stan.MaxInflight(1))

	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}
