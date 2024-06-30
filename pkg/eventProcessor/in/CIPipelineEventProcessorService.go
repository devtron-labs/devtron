/*
 * Copyright (c) 2024. Devtron Inc.
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

package in

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/pkg/git"
	"go.uber.org/zap"
)

type CIPipelineEventProcessorImpl struct {
	logger            *zap.SugaredLogger
	pubSubClient      *pubsub.PubSubClientServiceImpl
	gitWebhookService git.GitWebhookService
}

func NewCIPipelineEventProcessorImpl(logger *zap.SugaredLogger, pubSubClient *pubsub.PubSubClientServiceImpl,
	gitWebhookService git.GitWebhookService) *CIPipelineEventProcessorImpl {
	ciPipelineEventProcessorImpl := &CIPipelineEventProcessorImpl{
		logger:            logger,
		pubSubClient:      pubSubClient,
		gitWebhookService: gitWebhookService,
	}
	return ciPipelineEventProcessorImpl
}

func (impl *CIPipelineEventProcessorImpl) SubscribeNewCIMaterialEvent() error {
	callback := func(msg *model.PubSubMsg) {
		//defer msg.Ack()
		ciPipelineMaterial := gitSensor.CiPipelineMaterial{}
		err := json.Unmarshal([]byte(msg.Data), &ciPipelineMaterial)
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

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		ciPipelineMaterial := gitSensor.CiPipelineMaterial{}
		err := json.Unmarshal([]byte(msg.Data), &ciPipelineMaterial)
		if err != nil {
			return "error while unmarshalling json response", []interface{}{"error", err}
		}
		return "got message for about new ci material", []interface{}{"ciPipelineMaterialId", ciPipelineMaterial.Id, "gitMaterialId", ciPipelineMaterial.GitMaterialId, "type", ciPipelineMaterial.Type}
	}

	err := impl.pubSubClient.Subscribe(pubsub.NEW_CI_MATERIAL_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}
