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

package cd

import (
	"context"
	"errors"
	"fmt"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/adapter/cdWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/timelineStatus"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	globalUtil "github.com/devtron-labs/devtron/util"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"k8s.io/utils/strings/slices"
	"time"
)

type CdWorkflowCommonService interface {
	SupersedePreviousDeployments(ctx context.Context, cdWfrId int, pipelineId int, triggeredAt time.Time, triggeredBy int32) error
	MarkDeploymentFailedForRunnerId(cdWfrId int, releaseErr error, triggeredBy int32) error
	MarkCurrentDeploymentFailed(runner *pipelineConfig.CdWorkflowRunner, releaseErr error, triggeredBy int32) error
	UpdateNonTerminalStatusInRunner(ctx context.Context, wfrId int, userId int32, status string) error
	UpdatePreviousQueuedRunnerStatus(cdWfrId, pipelineId int, triggeredBy int32) error

	GetTriggerValidateFuncs() []pubsub.ValidateMsg
}

type CdWorkflowCommonServiceImpl struct {
	logger                        *zap.SugaredLogger
	cdWorkflowRepository          pipelineConfig.CdWorkflowRepository
	pipelineStatusTimelineService status.PipelineStatusTimelineService

	//TODO: remove below
	config                           *types.CdConfig
	pipelineRepository               pipelineConfig.PipelineRepository
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository
}

func NewCdWorkflowCommonServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	pipelineRepository pipelineConfig.PipelineRepository,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository) (*CdWorkflowCommonServiceImpl, error) {
	config, err := types.GetCdConfig()
	if err != nil {
		return nil, err
	}
	return &CdWorkflowCommonServiceImpl{
		logger:                           logger,
		cdWorkflowRepository:             cdWorkflowRepository,
		pipelineStatusTimelineService:    pipelineStatusTimelineService,
		config:                           config,
		pipelineRepository:               pipelineRepository,
		pipelineStatusTimelineRepository: pipelineStatusTimelineRepository,
	}, nil
}

func (impl *CdWorkflowCommonServiceImpl) SupersedePreviousDeployments(ctx context.Context, cdWfrId int, pipelineId int, triggeredAt time.Time, triggeredBy int32) error {
	_, span := otel.Tracer("orchestrator").Start(ctx, "CdWorkflowCommonServiceImpl.SupersedePreviousDeployments")
	defer span.End()
	// Initiating DB transaction
	dbConnection := impl.cdWorkflowRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		impl.logger.Errorw("error on update status, txn begin failed", "err", err)
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	//update [n,n-1] statuses as failed if not terminal
	terminalStatus := []string{string(health.HealthStatusHealthy), pipelineConfig.WorkflowAborted, pipelineConfig.WorkflowFailed, pipelineConfig.WorkflowSucceeded}
	previousNonTerminalRunners, err := impl.cdWorkflowRepository.FindPreviousCdWfRunnerByStatus(pipelineId, cdWfrId, terminalStatus)
	if err != nil {
		impl.logger.Errorw("error fetching previous wf runner, updating cd wf runner status,", "err", err, "currentRunnerId", cdWfrId)
		return err
	} else if len(previousNonTerminalRunners) == 0 {
		impl.logger.Infow("no previous runner found in updating cd wf runner status,", "err", err, "currentRunnerId", cdWfrId)
		return nil
	}

	var timelines []*pipelineConfig.PipelineStatusTimeline
	for _, previousRunner := range previousNonTerminalRunners {
		if previousRunner.Status == string(health.HealthStatusHealthy) ||
			previousRunner.Status == pipelineConfig.WorkflowSucceeded ||
			previousRunner.Status == pipelineConfig.WorkflowAborted ||
			previousRunner.Status == pipelineConfig.WorkflowFailed {
			//terminal status return
			impl.logger.Infow("skip updating cd wf runner status as previous runner status is", "status", previousRunner.Status)
			continue
		}
		impl.logger.Infow("updating cd wf runner status as previous runner status is", "status", previousRunner.Status)
		previousRunner.FinishedOn = triggeredAt
		previousRunner.Message = pipelineConfig.ErrorDeploymentSuperseded.Error()
		previousRunner.Status = pipelineConfig.WorkflowFailed
		previousRunner.UpdatedOn = time.Now()
		previousRunner.UpdatedBy = triggeredBy
		timeline := &pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: previousRunner.Id,
			Status:             timelineStatus.TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED,
			StatusDetail:       timelineStatus.TIMELINE_DESCRIPTION_DEPLOYMENT_SUPERSEDED,
			StatusTime:         time.Now(),
		}
		timeline.CreateAuditLog(1)
		timelines = append(timelines, timeline)
	}

	err = impl.cdWorkflowRepository.UpdateWorkFlowRunners(previousNonTerminalRunners)
	if err != nil {
		impl.logger.Errorw("error updating cd wf runner status", "err", err, "previousNonTerminalRunners", previousNonTerminalRunners)
		return err
	}
	err = impl.pipelineStatusTimelineRepository.SaveTimelinesWithTxn(timelines, tx)
	if err != nil {
		impl.logger.Errorw("error updating pipeline status timelines", "err", err, "timelines", timelines)
		return err
	}
	//commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error in db transaction commit", "err", err)
		return err
	}
	return nil

}

func (impl *CdWorkflowCommonServiceImpl) MarkDeploymentFailedForRunnerId(cdWfrId int, releaseErr error, triggeredBy int32) error {
	runner, err := impl.cdWorkflowRepository.FindBasicWorkflowRunnerById(cdWfrId)
	if err != nil {
		impl.logger.Errorw("err in FindWorkflowRunnerById", "cdWfrId", cdWfrId, "err", err)
		return err
	}
	return impl.MarkCurrentDeploymentFailed(runner, releaseErr, triggeredBy)
}

func (impl *CdWorkflowCommonServiceImpl) MarkCurrentDeploymentFailed(runner *pipelineConfig.CdWorkflowRunner, releaseErr error, triggeredBy int32) error {
	if slices.Contains(pipelineConfig.WfrTerminalStatusList, runner.Status) {
		impl.logger.Infow("cd wf runner status is already in terminal state", "err", releaseErr, "currentRunner", runner)
		return nil
	}
	//update current WF with error status
	impl.logger.Errorw("error in triggering cd WF, setting wf status as fail ", "wfId", runner.Id, "err", releaseErr)
	runner.Status = pipelineConfig.WorkflowFailed
	runner.Message = util.GetClientErrorDetailedMessage(releaseErr)
	runner.FinishedOn = time.Now()
	runner.UpdateAuditLog(triggeredBy)
	err1 := impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
	if err1 != nil {
		impl.logger.Errorw("error updating cd wf runner status", "err", releaseErr, "currentRunner", runner)
		return err1
	}
	if runner.WorkflowType.IsStageTypeDeploy() {
		if errors.Is(releaseErr, pipelineConfig.ErrorDeploymentSuperseded) {
			err := impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineSuperseded(runner.Id)
			if err != nil {
				impl.logger.Errorw("error updating CdPipelineStatusTimeline", "err", err, "releaseErr", releaseErr)
				return err
			}
		} else {
			err := impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineFailed(runner.Id, extractTimelineFailedStatusDetails(releaseErr))
			if err != nil {
				impl.logger.Errorw("error updating CdPipelineStatusTimeline", "err", err, "releaseErr", releaseErr)
				return err
			}
		}
		globalUtil.TriggerCDMetrics(cdWorkflow.GetTriggerMetricsFromRunnerObj(runner), impl.config.ExposeCDMetrics)
	}
	return nil
}

func (impl *CdWorkflowCommonServiceImpl) UpdateNonTerminalStatusInRunner(ctx context.Context, wfrId int, userId int32, status string) error {
	_, span := otel.Tracer("orchestrator").Start(ctx, "CdWorkflowCommonServiceImpl.UpdateNonTerminalStatusInRunner")
	defer span.End()
	// In case of terminal status update finished on time
	isTerminalStatus := slices.Contains(pipelineConfig.WfrTerminalStatusList, status)
	if isTerminalStatus {
		return fmt.Errorf("unsupported status %s for update operation", status)
	}
	cdWfr, err := impl.cdWorkflowRepository.FindBasicWorkflowRunnerById(wfrId)
	if err != nil {
		impl.logger.Errorw("err on fetching cd workflow, UpdateNonTerminalStatusInRunner", "err", err)
		return err
	}
	// if the current cdWfr status is already a terminal status and then don't update the status
	// e.g: Status : Failed --> Progressing (not allowed)
	if slices.Contains(pipelineConfig.WfrTerminalStatusList, cdWfr.Status) {
		impl.logger.Warnw("deployment has already been terminated for workflow runner, UpdateNonTerminalStatusInRunner", "workflowRunnerId", cdWfr.Id, "err", err)
		return nil
	}
	cdWfr.Status = status
	cdWfr.UpdateAuditLog(userId)
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(cdWfr)
	if err != nil {
		impl.logger.Errorw("error on update cd workflow runner, UpdateNonTerminalStatusInRunner", "cdWfr", cdWfr, "err", err)
		return err
	}
	return nil
}

func (impl *CdWorkflowCommonServiceImpl) UpdatePreviousQueuedRunnerStatus(cdWfrId, pipelineId int, triggeredBy int32) error {
	queuedRunners, err := impl.cdWorkflowRepository.GetPreviousQueuedRunners(cdWfrId, pipelineId)
	if err != nil {
		impl.logger.Errorw("error on getting previous queued cd workflow runner, UpdatePreviousQueuedRunnerStatus", "cdWfrId", cdWfrId, "err", err)
		return err
	}
	var queuedRunnerIds []int
	for _, queuedRunner := range queuedRunners {
		err = impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineSuperseded(queuedRunner.Id)
		if err != nil {
			impl.logger.Errorw("error updating pipeline status timelines", "err", err, "cdWfrId", queuedRunner.Id)
			return err
		}
		if queuedRunner.CdWorkflow == nil {
			pipeline, err := impl.pipelineRepository.FindById(pipelineId)
			if err != nil {
				impl.logger.Errorw("error in fetching cd pipeline, UpdatePreviousQueuedRunnerStatus", "pipelineId", pipelineId, "err", err)
				return err
			}
			queuedRunner.CdWorkflow = &pipelineConfig.CdWorkflow{
				Pipeline: pipeline,
			}
		}
		globalUtil.TriggerCDMetrics(cdWorkflow.GetTriggerMetricsFromRunnerObj(queuedRunner), impl.config.ExposeCDMetrics)
		queuedRunnerIds = append(queuedRunnerIds, queuedRunner.Id)
	}
	err = impl.cdWorkflowRepository.UpdateRunnerStatusToFailedForIds(pipelineConfig.ErrorDeploymentSuperseded.Error(), triggeredBy, queuedRunnerIds...)
	if err != nil {
		impl.logger.Errorw("error on update previous queued cd workflow runner, UpdatePreviousQueuedRunnerStatus", "cdWfrId", cdWfrId, "err", err)
		return err
	}
	return nil
}

func extractTimelineFailedStatusDetails(err error) string {
	errorString := util.GetClientErrorDetailedMessage(err)
	switch errorString {
	case pipelineConfig.FOUND_VULNERABILITY:
		return timelineStatus.TIMELINE_DESCRIPTION_VULNERABLE_IMAGE
	default:
		return util.GetTruncatedMessage(fmt.Sprintf("Deployment failed: %s", errorString), 255)
	}
}

// GetTriggerValidateFuncs gets all the required validation funcs
func (impl *CdWorkflowCommonServiceImpl) GetTriggerValidateFuncs() []pubsub.ValidateMsg {
	var duplicateTriggerValidateFunc pubsub.ValidateMsg = func(msg model.PubSubMsg) bool {
		if msg.MsgDeliverCount == 1 {
			// first time message got delivered, always validate this.
			return true
		}
		// message is redelivered, check if the message is already processed.
		if ok, err := impl.canInitiateTrigger(msg.MsgId); !ok || err != nil {
			impl.logger.Warnw("duplicate trigger condition, duplicate message", "msgId", msg.MsgId, "err", err)
			return false
		}
		return true
	}
	return []pubsub.ValidateMsg{duplicateTriggerValidateFunc}
}

// canInitiateTrigger checks if the current trigger request with natsMsgId haven't already initiated the trigger.
// throws error if the request is already processed.
func (impl *CdWorkflowCommonServiceImpl) canInitiateTrigger(natsMsgId string) (bool, error) {
	if natsMsgId == "" {
		return true, nil
	}
	exists, err := impl.cdWorkflowRepository.CheckWorkflowRunnerByReferenceId(natsMsgId)
	if err != nil {
		impl.logger.Errorw("error in fetching cd workflow runner using reference_id", "referenceId", natsMsgId, "err", err)
		return false, errors.New("error in fetching cd workflow runner")
	}

	if exists {
		impl.logger.Errorw("duplicate pre stage trigger request as there is already a workflow runner object created by this message")
		return false, errors.New("duplicate pre stage trigger request, this request was already processed")
	}
	return true, nil
}
