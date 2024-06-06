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

package out

import (
	"context"
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	apiBean "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/adapter"
	internalUtil "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/app/status"
	eventProcessorBean "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/celEvaluator"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/out/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	globalUtil "github.com/devtron-labs/devtron/util"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

type WorkflowEventPublishService interface {
	TriggerBulkHibernateAsync(request bean.StopDeploymentGroupRequest) (interface{}, error)
	TriggerAsyncRelease(userDeploymentRequestId int, overrideRequest *apiBean.ValuesOverrideRequest,
		valuesOverrideResponse *app.ValuesOverrideResponse, ctx context.Context, triggeredBy int32) (releaseNo int, err error)
	TriggerBulkDeploymentAsync(requests []*bean.BulkTriggerRequest, UserId int32) (interface{}, error)
}

type WorkflowEventPublishServiceImpl struct {
	logger                        *zap.SugaredLogger
	pubSubClient                  *pubsub.PubSubClientServiceImpl
	cdWorkflowCommonService       cd.CdWorkflowCommonService
	pipelineStatusTimelineService status.PipelineStatusTimelineService
	config                        *types.CdConfig
	triggerEventEvaluator         celEvaluator.TriggerEventEvaluator

	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
	pipelineRepository   pipelineConfig.PipelineRepository
	groupRepository      repository.DeploymentGroupRepository
}

func NewWorkflowEventPublishServiceImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	triggerEventEvaluator celEvaluator.TriggerEventEvaluator,

	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	groupRepository repository.DeploymentGroupRepository) (*WorkflowEventPublishServiceImpl, error) {
	config, err := types.GetCdConfig()
	if err != nil {
		return nil, err
	}
	impl := &WorkflowEventPublishServiceImpl{
		logger:                        logger,
		pubSubClient:                  pubSubClient,
		cdWorkflowCommonService:       cdWorkflowCommonService,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
		config:                        config,
		triggerEventEvaluator:         triggerEventEvaluator,

		cdWorkflowRepository: cdWorkflowRepository,
		pipelineRepository:   pipelineRepository,
		groupRepository:      groupRepository,
	}
	return impl, nil
}

func (impl *WorkflowEventPublishServiceImpl) TriggerBulkHibernateAsync(request bean.StopDeploymentGroupRequest) (interface{}, error) {
	dg, err := impl.groupRepository.FindByIdWithApp(request.DeploymentGroupId)
	if err != nil {
		impl.logger.Errorw("error while fetching dg", "err", err)
		return nil, err
	}

	for _, app := range dg.DeploymentGroupApps {
		deploymentGroupAppWithEnv := &bean.DeploymentGroupAppWithEnv{
			AppId:             app.AppId,
			EnvironmentId:     dg.EnvironmentId,
			DeploymentGroupId: dg.Id,
			Active:            dg.Active,
			UserId:            request.UserId,
			RequestType:       request.RequestType,
		}

		data, err := json.Marshal(deploymentGroupAppWithEnv)
		if err != nil {
			impl.logger.Errorw("error while writing app stop event to nats ", "app", app.AppId, "deploymentGroup", app.DeploymentGroupId, "err", err)
		} else {
			err = impl.pubSubClient.Publish(pubsub.BULK_HIBERNATE_TOPIC, string(data))
			if err != nil {
				impl.logger.Errorw("Error while publishing request", "topic", pubsub.BULK_HIBERNATE_TOPIC, "error", err)
			}
		}
	}
	return nil, nil
}

// TriggerAsyncRelease will publish async Install/Upgrade request event for Devtron App releases
func (impl *WorkflowEventPublishServiceImpl) TriggerAsyncRelease(userDeploymentRequestId int, overrideRequest *apiBean.ValuesOverrideRequest,
	valuesOverrideResponse *app.ValuesOverrideResponse, ctx context.Context, triggeredBy int32) (releaseNo int, err error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "WorkflowEventPublishServiceImpl.TriggerAsyncRelease")
	defer span.End()
	topic, msg, err := impl.getAsyncDeploymentTopicAndPayload(userDeploymentRequestId, overrideRequest.DeploymentAppType, valuesOverrideResponse)
	if err != nil {
		impl.logger.Errorw("error in fetching values for trigger", "err", err)
		return releaseNo, err
	}
	// publish nats event for async installation
	err = impl.pubSubClient.Publish(topic, msg)
	if err != nil {
		impl.logger.Errorw("failed to publish trigger request event", "topic", topic, "msg", msg, "err", err)
		//update workflow runner status, used in app workflow view
		err1 := impl.cdWorkflowCommonService.UpdateCDWorkflowRunnerStatus(newCtx, overrideRequest.WfrId, overrideRequest.UserId, pipelineConfig.WorkflowFailed, adapter.WithMessage(err.Error()))
		if err1 != nil {
			impl.logger.Errorw("error in updating the workflow runner status, TriggerAsyncRelease", "err", err1)
		}
		return 0, err
	}

	//update workflow runner status, used in app workflow view
	err = impl.cdWorkflowCommonService.UpdateCDWorkflowRunnerStatus(newCtx, overrideRequest.WfrId, overrideRequest.UserId, pipelineConfig.WorkflowInQueue)
	if err != nil {
		impl.logger.Errorw("error in updating the workflow runner status, TriggerAsyncRelease", "err", err)
		return 0, err
	}
	err = impl.UpdatePreviousQueuedRunnerStatus(overrideRequest.WfrId, overrideRequest.PipelineId, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in updating the previous queued workflow runner status, TriggerAsyncRelease", "err", err)
		return 0, err
	}
	return 0, nil
}

func (impl *WorkflowEventPublishServiceImpl) UpdatePreviousQueuedRunnerStatus(cdWfrId, pipelineId int, triggeredBy int32) error {
	cdWfrs, err := impl.cdWorkflowRepository.UpdatePreviousQueuedRunnerStatus(cdWfrId, pipelineId, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error on update previous queued cd workflow runner, UpdatePreviousQueuedRunnerStatus", "cdWfrId", cdWfrId, "err", err)
		return err
	}
	var timelines []*pipelineConfig.PipelineStatusTimeline
	for _, cdWfr := range cdWfrs {
		err = impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineSuperseded(cdWfr.Id)
		if err != nil {
			impl.logger.Errorw("error updating pipeline status timelines", "err", err, "timelines", timelines)
			return err
		}
		if cdWfr.CdWorkflow == nil {
			pipeline, err := impl.pipelineRepository.FindById(pipelineId)
			if err != nil {
				impl.logger.Errorw("error in fetching cd pipeline, UpdatePreviousQueuedRunnerStatus", "pipelineId", pipelineId, "err", err)
				return err
			}
			cdWfr.CdWorkflow = &pipelineConfig.CdWorkflow{
				Pipeline: pipeline,
			}
		}
		globalUtil.TriggerCDMetrics(pipelineConfig.GetTriggerMetricsFromRunnerObj(cdWfr), impl.config.ExposeCDMetrics)
	}
	return nil
}

func (impl *WorkflowEventPublishServiceImpl) TriggerBulkDeploymentAsync(requests []*bean.BulkTriggerRequest, UserId int32) (interface{}, error) {
	var cdWorkflows []*pipelineConfig.CdWorkflow
	for _, request := range requests {
		cdWf := &pipelineConfig.CdWorkflow{
			CiArtifactId:   request.CiArtifactId,
			PipelineId:     request.PipelineId,
			AuditLog:       sql.AuditLog{CreatedOn: time.Now(), CreatedBy: UserId, UpdatedOn: time.Now(), UpdatedBy: UserId},
			WorkflowStatus: pipelineConfig.REQUEST_ACCEPTED,
		}
		cdWorkflows = append(cdWorkflows, cdWf)
	}
	err := impl.cdWorkflowRepository.SaveWorkFlows(cdWorkflows...)
	if err != nil {
		impl.logger.Errorw("error in saving wfs", "req", requests, "err", err)
		return nil, err
	}
	impl.triggerNatsEventForBulkAction(cdWorkflows)
	return nil, nil
}

func (impl *WorkflowEventPublishServiceImpl) triggerNatsEventForBulkAction(cdWorkflows []*pipelineConfig.CdWorkflow) {
	for _, wf := range cdWorkflows {
		data, err := json.Marshal(wf)
		if err != nil {
			wf.WorkflowStatus = pipelineConfig.QUE_ERROR
		} else {
			err = impl.pubSubClient.Publish(pubsub.BULK_DEPLOY_TOPIC, string(data))
			if err != nil {
				wf.WorkflowStatus = pipelineConfig.QUE_ERROR
			} else {
				wf.WorkflowStatus = pipelineConfig.ENQUEUED
			}
		}
		err = impl.cdWorkflowRepository.UpdateWorkFlow(wf)
		if err != nil {
			impl.logger.Errorw("error in publishing wf msg", "wf", wf, "err", err)
		}
	}
}

func (impl *WorkflowEventPublishServiceImpl) getAsyncDeploymentTopicAndPayload(userDeploymentRequestId int,
	deploymentAppType string, valuesOverrideResponse *app.ValuesOverrideResponse) (topic string, msg string, err error) {
	isPriorityEvent, err := impl.triggerEventEvaluator.IsPriorityDeployment(valuesOverrideResponse)
	if err != nil {
		impl.logger.Errorw("error while CEL expression evaluation", "err", err)
		return topic, msg, err
	}
	if internalUtil.IsAcdApp(deploymentAppType) {
		topic = pubsub.DEVTRON_CHART_GITOPS_INSTALL_TOPIC
		if isPriorityEvent {
			topic = pubsub.DEVTRON_CHART_GITOPS_PRIORITY_INSTALL_TOPIC
		}
	}
	if internalUtil.IsHelmApp(deploymentAppType) {
		topic = pubsub.DEVTRON_CHART_INSTALL_TOPIC
		if isPriorityEvent {
			topic = pubsub.DEVTRON_CHART_PRIORITY_INSTALL_TOPIC
		}
	}
	event := &eventProcessorBean.UserDeploymentRequest{
		Id: userDeploymentRequestId,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		impl.logger.Errorw("failed to marshal async CD deploy event request", "request", event, "err", err)
		return topic, msg, err
	}
	msg = string(payload)
	return topic, msg, nil
}
