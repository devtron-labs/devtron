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
	context2 "context"
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	installedAppReader "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	installedAppReadBean "github.com/devtron-labs/devtron/pkg/appStore/installedApp/read/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps"
	bean2 "github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	"github.com/devtron-labs/devtron/pkg/workflow/status"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
)

type CDPipelineEventProcessorImpl struct {
	logger                  *zap.SugaredLogger
	pubSubClient            *pubsub.PubSubClientServiceImpl
	cdWorkflowCommonService cd.CdWorkflowCommonService
	workflowStatusService   status.WorkflowStatusService
	cdHandlerService        devtronApps.HandlerService
	pipelineRepository      pipelineConfig.PipelineRepository
	installedAppReadService installedAppReader.InstalledAppReadService
}

func NewCDPipelineEventProcessorImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	workflowStatusService status.WorkflowStatusService,
	cdHandlerService devtronApps.HandlerService,
	pipelineRepository pipelineConfig.PipelineRepository,
	installedAppReadService installedAppReader.InstalledAppReadService) *CDPipelineEventProcessorImpl {
	cdPipelineEventProcessorImpl := &CDPipelineEventProcessorImpl{
		logger:                  logger,
		pubSubClient:            pubSubClient,
		cdWorkflowCommonService: cdWorkflowCommonService,
		workflowStatusService:   workflowStatusService,
		cdHandlerService:        cdHandlerService,
		pipelineRepository:      pipelineRepository,
		installedAppReadService: installedAppReadService,
	}
	return cdPipelineEventProcessorImpl
}

func (impl *CDPipelineEventProcessorImpl) SubscribeCDBulkTriggerTopic() error {

	callback := func(msg *model.PubSubMsg) {
		event := &bean.BulkCDDeployEvent{}
		err := json.Unmarshal([]byte(msg.Data), event)
		if err != nil {
			impl.logger.Errorw("Error unmarshalling received event", "topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC, "msg", msg.Data, "err", err)
			return
		}
		event.ValuesOverrideRequest.UserId = event.UserId
		// trigger

		triggerContext := bean2.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
			Context:     context2.Background(),
		}
		_, _, _, err = impl.cdHandlerService.ManualCdTrigger(triggerContext, event.ValuesOverrideRequest, event.UserMetadata)
		if err != nil {
			impl.logger.Errorw("Error triggering CD", "topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC, "msg", msg.Data, "err", err)
		}
	}
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		event := &bean.BulkCDDeployEvent{}
		err := json.Unmarshal([]byte(msg.Data), event)
		if err != nil {
			return "error unmarshalling received event", []interface{}{"msg", msg.Data, "err", err}
		}
		return "got message for trigger cd in bulk", []interface{}{"pipelineId", event.ValuesOverrideRequest.PipelineId, "appId", event.ValuesOverrideRequest.AppId, "cdWorkflowType", event.ValuesOverrideRequest.CdWorkflowType, "ciArtifactId", event.ValuesOverrideRequest.CiArtifactId}
	}
	validations := impl.cdWorkflowCommonService.GetTriggerValidateFuncs()
	err := impl.pubSubClient.Subscribe(pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC, callback, loggerFunc, validations...)
	if err != nil {
		impl.logger.Error("failed to subscribe to NATS topic", "topic", pubsub.CD_BULK_DEPLOY_TRIGGER_TOPIC, "err", err)
		return err
	}
	return nil
}

func (impl *CDPipelineEventProcessorImpl) SubscribeArgoTypePipelineSyncEvent() error {
	callback := func(msg *model.PubSubMsg) {
		statusUpdateEvent := bean.ArgoPipelineStatusSyncEvent{}
		var err error
		var cdPipeline *pipelineConfig.Pipeline
		var installedApp *installedAppReadBean.InstalledAppMin

		err = json.Unmarshal([]byte(msg.Data), &statusUpdateEvent)
		if err != nil {
			impl.logger.Errorw("unmarshal error on argo pipeline status update event", "err", err)
			return
		}

		if statusUpdateEvent.IsAppStoreApplication {
			installedApp, err = impl.installedAppReadService.GetInstalledAppByInstalledAppVersionId(statusUpdateEvent.InstalledAppVersionId)
			if err != nil {
				impl.logger.Errorw("error in getting installedAppVersion by id", "err", err, "id", statusUpdateEvent.PipelineId)
				return
			}
		} else {
			cdPipeline, err = impl.pipelineRepository.FindById(statusUpdateEvent.PipelineId)
			if err != nil {
				impl.logger.Errorw("error in getting cdPipeline by id", "err", err, "id", statusUpdateEvent.PipelineId)
				return
			}
		}

		triggerContext := bean2.TriggerContext{
			ReferenceId: pointer.String(msg.MsgId),
		}

		err, _ = impl.workflowStatusService.UpdatePipelineTimelineAndStatusByLiveApplicationFetch(triggerContext, cdPipeline, installedApp, statusUpdateEvent.UserId)
		if err != nil {
			impl.logger.Errorw("error on argo pipeline status update", "err", err, "msg", msg.Data)
			return
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		statusUpdateEvent := bean.ArgoPipelineStatusSyncEvent{}
		err := json.Unmarshal([]byte(msg.Data), &statusUpdateEvent)
		if err != nil {
			return "unmarshal error on argo pipeline status update event", []interface{}{"err", err}
		}
		return "got message for argo pipeline status update", []interface{}{"pipelineId", statusUpdateEvent.PipelineId, "installedAppVersionId", statusUpdateEvent.InstalledAppVersionId, "isAppStoreApplication", statusUpdateEvent.IsAppStoreApplication}
	}

	validations := impl.cdWorkflowCommonService.GetTriggerValidateFuncs()
	err := impl.pubSubClient.Subscribe(pubsub.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, callback, loggerFunc, validations...)
	if err != nil {
		impl.logger.Errorw("error in subscribing to argo application status update topic", "err", err)
		return err
	}
	return nil
}
