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
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
)

type ApplicationStatusUpdateHandler interface {
	Subscribe() error
}

type ApplicationStatusUpdateHandlerImpl struct {
	logger              *zap.SugaredLogger
	pubsubClient        *pubsub.PubSubClient
	appService          app.AppService
	workflowDagExecutor pipeline.WorkflowDagExecutor
}

const applicationStatusUpdate = "APPLICATION_STATUS_UPDATE"
const applicationStatusUpdateGroup = "APPLICATION_STATUS_UPDATE_GROUP-1"
const applicationStatusUpdateDurable = "APPLICATION_STATUS_UPDATE_DURABLE-1"

func NewApplicationStatusUpdateHandlerImpl(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClient, appService app.AppService,
	workflowDagExecutor pipeline.WorkflowDagExecutor) *ApplicationStatusUpdateHandlerImpl {
	appStatusUpdateHandlerImpl := &ApplicationStatusUpdateHandlerImpl{
		logger:              logger,
		pubsubClient:        pubsubClient,
		appService:          appService,
		workflowDagExecutor: workflowDagExecutor,
	}
	err := appStatusUpdateHandlerImpl.Subscribe()
	if err != nil {
		logger.Error("err", err)
		return nil
	}
	return appStatusUpdateHandlerImpl
}

func (impl *ApplicationStatusUpdateHandlerImpl) Subscribe() error {

	return nil
}
