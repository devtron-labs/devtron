package out

import (
	"context"
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest"
	bean7 "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/out/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/workflow/cd"
	util4 "github.com/devtron-labs/devtron/util"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

type WorkflowEventPublishService interface {
	PublishDeployStageSuccessEvent(request bean.DeployStageSuccessEventReq) error
	TriggerBulkHibernateAsync(request bean.StopDeploymentGroupRequest) (interface{}, error)
	TriggerHelmAsyncRelease(overrideRequest *bean3.ValuesOverrideRequest, ctx context.Context, triggeredAt time.Time,
		triggeredBy int32) (releaseNo int, manifest []byte, err error)
	TriggerBulkDeploymentAsync(requests []*bean.BulkTriggerRequest, UserId int32) (interface{}, error)
}

type WorkflowEventPublishServiceImpl struct {
	logger                              *zap.SugaredLogger
	pubSubClient                        *pubsub.PubSubClientServiceImpl
	cdWorkflowCommonService             cd.CdWorkflowCommonService
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService
	manifestCreationService             manifest.ManifestCreationService
	pipelineStatusTimelineService       status.PipelineStatusTimelineService
	config                              *types.CdConfig

	cdWorkflowRepository pipelineConfig.CdWorkflowRepository
	pipelineRepository   pipelineConfig.PipelineRepository
	groupRepository      repository.DeploymentGroupRepository
}

func NewWorkflowEventPublishServiceImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl,
	cdWorkflowCommonService cd.CdWorkflowCommonService,
	deployedConfigurationHistoryService history.DeployedConfigurationHistoryService,
	manifestCreationService manifest.ManifestCreationService,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,

	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineRepository pipelineConfig.PipelineRepository,
	groupRepository repository.DeploymentGroupRepository) (*WorkflowEventPublishServiceImpl, error) {
	config, err := types.GetCdConfig()
	if err != nil {
		return nil, err
	}
	impl := &WorkflowEventPublishServiceImpl{
		logger:                              logger,
		pubSubClient:                        pubSubClient,
		cdWorkflowCommonService:             cdWorkflowCommonService,
		deployedConfigurationHistoryService: deployedConfigurationHistoryService,
		manifestCreationService:             manifestCreationService,
		pipelineStatusTimelineService:       pipelineStatusTimelineService,
		config:                              config,

		cdWorkflowRepository: cdWorkflowRepository,
		pipelineRepository:   pipelineRepository,
		groupRepository:      groupRepository,
	}
	return impl, nil
}

func (impl *WorkflowEventPublishServiceImpl) PublishDeployStageSuccessEvent(request bean.DeployStageSuccessEventReq) error {
	reqInBytes, err := json.Marshal(request)
	if err != nil {
		impl.logger.Errorw("error in marshaling  HandleDeployStageSuccessEvent request", "err", err, "request", request)
		return err
	}
	err = impl.pubSubClient.Publish(pubsub.CD_STAGE_SUCCESS_EVENT_TOPIC, string(reqInBytes))
	if err != nil {
		impl.logger.Errorw("Error while publishing request", "topic", pubsub.CD_STAGE_SUCCESS_EVENT_TOPIC, "error", err)
		return err
	}
	return nil
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

// TriggerHelmAsyncRelease will publish async helm Install/Upgrade request event for Devtron App releases
func (impl *WorkflowEventPublishServiceImpl) TriggerHelmAsyncRelease(overrideRequest *bean3.ValuesOverrideRequest, ctx context.Context, triggeredAt time.Time, triggeredBy int32) (releaseNo int, manifest []byte, err error) {
	// build merged values and save PCO history for the release
	valuesOverrideResponse, err := impl.manifestCreationService.GetValuesOverrideForTrigger(overrideRequest, triggeredAt, ctx)
	_, span := otel.Tracer("orchestrator").Start(ctx, "CreateHistoriesForDeploymentTrigger")
	// save triggered deployment history
	err1 := impl.deployedConfigurationHistoryService.CreateHistoriesForDeploymentTrigger(valuesOverrideResponse.Pipeline, valuesOverrideResponse.PipelineStrategy, valuesOverrideResponse.EnvOverride, triggeredAt, triggeredBy)
	if err1 != nil {
		impl.logger.Errorw("error in saving histories for trigger", "err", err1, "pipelineId", valuesOverrideResponse.Pipeline.Id, "wfrId", overrideRequest.WfrId)
	}
	span.End()
	if err != nil {
		impl.logger.Errorw("error in fetching values for trigger", "err", err)
		return releaseNo, manifest, err
	}

	event := &bean7.AsyncCdDeployEvent{
		ValuesOverrideRequest: overrideRequest,
		TriggeredAt:           triggeredAt,
		TriggeredBy:           triggeredBy,
	}
	payload, err := json.Marshal(event)
	if err != nil {
		impl.logger.Errorw("failed to marshal helm async CD deploy event request", "request", event, "err", err)
		return 0, manifest, err
	}

	// publish nats event for async installation
	err = impl.pubSubClient.Publish(pubsub.DEVTRON_CHART_INSTALL_TOPIC, string(payload))
	if err != nil {
		impl.logger.Errorw("failed to publish trigger request event", "topic", pubsub.DEVTRON_CHART_INSTALL_TOPIC, "payload", payload, "err", err)
		// update workflow runner status, used in app workflow view
		err1 = impl.cdWorkflowCommonService.UpdateCDWorkflowRunnerStatus(ctx, overrideRequest, triggeredAt, pipelineConfig.WorkflowFailed, err.Error())
		if err1 != nil {
			impl.logger.Errorw("error in updating the workflow runner status, TriggerHelmAsyncRelease", "err", err1)
		}
		return 0, manifest, err
	}

	// update workflow runner status, used in app workflow view
	err = impl.cdWorkflowCommonService.UpdateCDWorkflowRunnerStatus(ctx, overrideRequest, triggeredAt, pipelineConfig.WorkflowInQueue, "")
	if err != nil {
		impl.logger.Errorw("error in updating the workflow runner status, TriggerHelmAsyncRelease", "err", err)
		return 0, manifest, err
	}
	err = impl.UpdatePreviousQueuedRunnerStatus(overrideRequest.WfrId, overrideRequest.PipelineId, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in updating the previous queued workflow runner status, TriggerHelmAsyncRelease", "err", err)
		return 0, manifest, err
	}
	return 0, manifest, nil
}

func (impl *WorkflowEventPublishServiceImpl) UpdatePreviousQueuedRunnerStatus(cdWfrId, pipelineId int, triggeredBy int32) error {
	cdWfrs, err := impl.cdWorkflowRepository.UpdatePreviousQueuedRunnerStatus(cdWfrId, pipelineId, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error on update previous queued cd workflow runner, UpdatePreviousQueuedRunnerStatus", "cdWfrId", cdWfrId, "err", err)
		return err
	}
	for _, cdWfr := range cdWfrs {
		err = impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineFailed(cdWfr.Id, pipelineConfig.NEW_DEPLOYMENT_INITIATED)
		if err != nil {
			impl.logger.Errorw("error updating CdPipelineStatusTimeline, UpdatePreviousQueuedRunnerStatus", "err", err)
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
		util4.TriggerCDMetrics(pipelineConfig.GetTriggerMetricsFromRunnerObj(cdWfr), impl.config.ExposeCDMetrics)
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
