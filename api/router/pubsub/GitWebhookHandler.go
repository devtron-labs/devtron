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
	"github.com/devtron-labs/common-lib-private/pubsub-lib/model"

	pubsub "github.com/devtron-labs/common-lib-private/pubsub-lib"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/pkg/git"

	"go.uber.org/zap"
)

type GitWebhookHandler interface {
	subscribe() error
}

type GitWebhookHandlerImpl struct {
	logger            *zap.SugaredLogger
	pubsubClient      *pubsub.PubSubClientServiceImpl
	gitWebhookService git.GitWebhookService
}

func NewGitWebhookHandler(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClientServiceImpl, gitWebhookService git.GitWebhookService) *GitWebhookHandlerImpl {
	gitWebhookHandlerImpl := &GitWebhookHandlerImpl{
		logger:            logger,
		pubsubClient:      pubsubClient,
		gitWebhookService: gitWebhookService,
	}
	err := gitWebhookHandlerImpl.subscribe()
	if err != nil {
		logger.Error("err", err)
		return nil
	}
	return gitWebhookHandlerImpl
}

func (impl *GitWebhookHandlerImpl) subscribe() error {
	callback := func(msg *model.PubSubMsg) {
		//defer msg.Ack()
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
	}
	err := impl.pubsubClient.Subscribe(pubsub.NEW_CI_MATERIAL_TOPIC, callback)
	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}
