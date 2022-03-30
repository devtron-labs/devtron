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

	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/git"
	"github.com/devtron-labs/devtron/util"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type GitWebhookHandler interface {
	Subscribe() error
}

type GitWebhookHandlerImpl struct {
	logger            *zap.SugaredLogger
	pubsubClient      *pubsub.PubSubClient
	gitWebhookService git.GitWebhookService
}

func NewGitWebhookHandler(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClient, gitWebhookService git.GitWebhookService) *GitWebhookHandlerImpl {
	gitWebhookHandlerImpl := &GitWebhookHandlerImpl{
		logger:            logger,
		pubsubClient:      pubsubClient,
		gitWebhookService: gitWebhookService,
	}
	err := util.AddStream(gitWebhookHandlerImpl.pubsubClient.JetStrCtxt, util.GIT_SENSOR_STREAM)
	if err != nil {
		logger.Error("err", err)
		return nil
	}
	err = gitWebhookHandlerImpl.Subscribe()
	if err != nil {
		logger.Error("err", err)
		return nil
	}
	return gitWebhookHandlerImpl
}

func (impl *GitWebhookHandlerImpl) Subscribe() error {
	_, err := impl.pubsubClient.JetStrCtxt.QueueSubscribe(util.NEW_CI_MATERIAL_TOPIC, util.NEW_CI_MATERIAL_TOPIC_GROUP, func(msg *nats.Msg) {
		defer msg.Ack()
		ciPipelineMaterial := gitSensor.CiPipelineMaterial{}
		err := json.Unmarshal([]byte(string(msg.Data)), &ciPipelineMaterial)
		if err != nil {
			impl.logger.Error("Error while unmarshalling json response", "error", err)
			return
		}
		resp, err := impl.gitWebhookService.HandleGitWebhook(ciPipelineMaterial)
		impl.logger.Debug(resp)
		if err != nil {
			impl.logger.Error("err", err)
			return
		}
	}, nats.Durable(util.NEW_CI_MATERIAL_TOPIC_DURABLE), nats.DeliverLast(), nats.ManualAck(), nats.BindStream(util.GIT_SENSOR_STREAM))

	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}
