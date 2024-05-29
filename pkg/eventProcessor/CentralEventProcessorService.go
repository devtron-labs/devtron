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

package eventProcessor

import (
	"github.com/devtron-labs/devtron/pkg/eventProcessor/in"
	"go.uber.org/zap"
)

type CentralEventProcessor struct {
	logger                                *zap.SugaredLogger
	workflowEventProcessor                *in.WorkflowEventProcessorImpl
	ciPipelineEventProcessor              *in.CIPipelineEventProcessorImpl
	cdPipelineEventProcessor              *in.CDPipelineEventProcessorImpl
	deployedApplicationEventProcessorImpl *in.DeployedApplicationEventProcessorImpl
	appStoreAppsEventProcessorImpl        *in.AppStoreAppsEventProcessorImpl
}

func NewCentralEventProcessor(logger *zap.SugaredLogger,
	workflowEventProcessor *in.WorkflowEventProcessorImpl,
	ciPipelineEventProcessor *in.CIPipelineEventProcessorImpl,
	cdPipelineEventProcessor *in.CDPipelineEventProcessorImpl,
	deployedApplicationEventProcessorImpl *in.DeployedApplicationEventProcessorImpl,
	appStoreAppsEventProcessorImpl *in.AppStoreAppsEventProcessorImpl) (*CentralEventProcessor, error) {
	cep := &CentralEventProcessor{
		logger:                                logger,
		workflowEventProcessor:                workflowEventProcessor,
		ciPipelineEventProcessor:              ciPipelineEventProcessor,
		cdPipelineEventProcessor:              cdPipelineEventProcessor,
		deployedApplicationEventProcessorImpl: deployedApplicationEventProcessorImpl,
		appStoreAppsEventProcessorImpl:        appStoreAppsEventProcessorImpl,
	}
	err := cep.SubscribeAll()
	if err != nil {
		return nil, err
	}
	return cep, nil
}

func (impl *CentralEventProcessor) SubscribeAll() error {
	var err error

	//CI pipeline event starts
	err = impl.ciPipelineEventProcessor.SubscribeNewCIMaterialEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeNewCIMaterialEvent", "err", err)
		return err
	}
	//CI pipeline event ends

	//CD pipeline event starts

	err = impl.cdPipelineEventProcessor.SubscribeCDBulkTriggerTopic()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCDBulkTriggerTopic", "err", err)
		return err
	}

	err = impl.cdPipelineEventProcessor.SubscribeArgoTypePipelineSyncEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeArgoTypePipelineSyncEvent", "err", err)
		return err
	}

	//CD pipeline event ends

	//Workflow event starts

	err = impl.workflowEventProcessor.SubscribeCDStageCompleteEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCDStageCompleteEvent", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeTriggerBulkAction()
	if err != nil {
		impl.logger.Errorw("error, SubscribeTriggerBulkAction", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeHibernateBulkAction()
	if err != nil {
		impl.logger.Errorw("error, SubscribeHibernateBulkAction", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeCIWorkflowStatusUpdate()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCIWorkflowStatusUpdate", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeCDWorkflowStatusUpdate()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCDWorkflowStatusUpdate", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeCICompleteEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCICompleteEvent", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeDevtronAsyncInstallRequest()
	if err != nil {
		impl.logger.Errorw("error, SubscribeDevtronAsyncInstallRequest", "err", err)
		return err
	}
	err = impl.workflowEventProcessor.SubscribeCDPipelineDeleteEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeCDPipelineDeleteEvent", "err", err)
		return err
	}

	//Workflow event ends

	//Deployed application status event starts (currently only argo)

	err = impl.deployedApplicationEventProcessorImpl.SubscribeArgoAppUpdate()
	if err != nil {
		impl.logger.Errorw("error, SubscribeArgoAppUpdate", "err", err)
		return err
	}
	err = impl.deployedApplicationEventProcessorImpl.SubscribeArgoAppDeleteStatus()
	if err != nil {
		impl.logger.Errorw("error, SubscribeArgoAppDeleteStatus", "err", err)
		return err
	}

	//Deployed application status event ends (currently only argo)

	//AppStore apps event starts

	err = impl.appStoreAppsEventProcessorImpl.SubscribeAppStoreAppsBulkDeployEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeAppStoreAppsBulkDeployEvent", "err", err)
		return err
	}

	err = impl.appStoreAppsEventProcessorImpl.SubscribeHelmInstallStatusEvent()
	if err != nil {
		impl.logger.Errorw("error, SubscribeHelmInstallStatusEvent", "err", err)
		return err
	}

	//AppStore apps event ends

	return nil
}
