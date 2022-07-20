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
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
)

type ApplicationStatusUpdateHandler interface {
	Subscribe() error
}

type ApplicationStatusUpdateHandlerImpl struct {
	logger              *zap.SugaredLogger
	appService          app.AppService
	workflowDagExecutor pipeline.WorkflowDagExecutor
	installedAppService service.InstalledAppService
}

func NewApplicationStatusUpdateHandlerImpl(logger *zap.SugaredLogger, appService app.AppService,
	workflowDagExecutor pipeline.WorkflowDagExecutor, installedAppService service.InstalledAppService) *ApplicationStatusUpdateHandlerImpl {
	appStatusUpdateHandlerImpl := &ApplicationStatusUpdateHandlerImpl{
		logger:              logger,
		appService:          appService,
		workflowDagExecutor: workflowDagExecutor,
		installedAppService: installedAppService,
	}

	return appStatusUpdateHandlerImpl
}

func (impl *ApplicationStatusUpdateHandlerImpl) Subscribe() error {
	return nil
}
