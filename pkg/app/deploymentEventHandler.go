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

package app

import (
	"github.com/devtron-labs/common-lib/async"
	"github.com/devtron-labs/devtron/api/bean"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	util "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
)

type DeploymentEventHandler interface {
	WriteCDNotificationEventAsync(appId int, envId int, override *chartConfig.PipelineOverride, eventType util.EventType)
}

type DeploymentEventHandlerImpl struct {
	logger        *zap.SugaredLogger
	eventFactory  client.EventFactory
	eventClient   client.EventClient
	asyncRunnable *async.Runnable
}

func NewDeploymentEventHandlerImpl(logger *zap.SugaredLogger, eventClient client.EventClient, eventFactory client.EventFactory, asyncRunnable *async.Runnable) *DeploymentEventHandlerImpl {
	deploymentEventHandlerImpl := &DeploymentEventHandlerImpl{
		logger:        logger,
		eventClient:   eventClient,
		eventFactory:  eventFactory,
		asyncRunnable: asyncRunnable,
	}
	return deploymentEventHandlerImpl
}

func (impl *DeploymentEventHandlerImpl) writeCDNotificationEvent(appId int, envId int, override *chartConfig.PipelineOverride, eventType util.EventType) {
	event, _ := impl.eventFactory.Build(eventType, &override.PipelineId, appId, &envId, util.CD)
	impl.logger.Debugw("event WriteCDNotificationEvent", "event", event, "override", override)
	event = impl.eventFactory.BuildExtraCDData(event, nil, override.Id, bean.CD_WORKFLOW_TYPE_DEPLOY)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("error in writing event", "event", event, "err", evtErr)
	}
}

// WriteCDNotificationEventAsync executes WriteCDNotificationEvent in a panic-safe goroutine
func (impl *DeploymentEventHandlerImpl) WriteCDNotificationEventAsync(appId int, envId int, override *chartConfig.PipelineOverride, eventType util.EventType) {
	impl.asyncRunnable.Execute(func() {
		impl.writeCDNotificationEvent(appId, envId, override, eventType)
	})
}
