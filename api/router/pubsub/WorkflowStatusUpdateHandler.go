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
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
)

type WorkflowStatusUpdateHandler interface {
	Subscribe() error
}

type WorkflowStatusUpdateHandlerImpl struct {
	logger               *zap.SugaredLogger
	pubsubClient         *pubsub.PubSubClient
	ciHandler            pipeline.CiHandler
	cdHandler            pipeline.CdHandler
	eventFactory         client.EventFactory
	eventClient          client.EventClient
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
}

const workflowStatusUpdate = "WORKFLOW_STATUS_UPDATE"
const workflowStatusUpdateGroup = "WORKFLOW_STATUS_UPDATE_GROUP-1"
const workflowStatusUpdateDurable = "WORKFLOW_STATUS_UPDATE_DURABLE-1"

const cdWorkflowStatusUpdate = "CD_WORKFLOW_STATUS_UPDATE"
const cdWorkflowStatusUpdateGroup = "CD_WORKFLOW_STATUS_UPDATE_GROUP-1"
const cdWorkflowStatusUpdateDurable = "CD_WORKFLOW_STATUS_UPDATE_DURABLE-1"

func NewWorkflowStatusUpdateHandlerImpl(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClient, ciHandler pipeline.CiHandler, cdHandler pipeline.CdHandler,
	eventFactory client.EventFactory, eventClient client.EventClient, cdWorkflowRepository pipelineConfig.CdWorkflowRepository) *WorkflowStatusUpdateHandlerImpl {
	workflowStatusUpdateHandlerImpl := &WorkflowStatusUpdateHandlerImpl{
		logger:               logger,
		pubsubClient:         pubsubClient,
		ciHandler:            ciHandler,
		cdHandler:            cdHandler,
		eventFactory:         eventFactory,
		eventClient:          eventClient,
		cdWorkflowRepository: cdWorkflowRepository,
	}
	err := workflowStatusUpdateHandlerImpl.Subscribe()
	if err != nil {
		logger.Error("err", err)
		return nil
	}
	err = workflowStatusUpdateHandlerImpl.SubscribeCD()
	if err != nil {
		logger.Error("err", err)
		return nil
	}
	return workflowStatusUpdateHandlerImpl
}

func (impl *WorkflowStatusUpdateHandlerImpl) Subscribe() error {

	return nil
}

func (impl *WorkflowStatusUpdateHandlerImpl) SubscribeCD() error {

	return nil
}
