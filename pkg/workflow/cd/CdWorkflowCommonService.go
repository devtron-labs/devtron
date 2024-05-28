package cd

import (
	"context"
	"errors"
	"fmt"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/adapter"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	util4 "github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"k8s.io/utils/strings/slices"
	"time"
)

type CdWorkflowCommonService interface {
	SupersedePreviousDeployments(cdWfrId int, pipelineId int, triggeredAt time.Time, triggeredBy int32) error
	MarkCurrentDeploymentFailed(runner *pipelineConfig.CdWorkflowRunner, releaseErr error, triggeredBy int32) error
	UpdateCDWorkflowRunnerStatus(ctx context.Context, wfrId int, userId int32, status string, options ...adapter.UpdateOptions) error

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

func (impl *CdWorkflowCommonServiceImpl) SupersedePreviousDeployments(cdWfrId int, pipelineId int, triggeredAt time.Time, triggeredBy int32) error {
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
		impl.logger.Errorw("no previous runner found in updating cd wf runner status,", "err", err, "currentRunnerId", cdWfrId)
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
		previousRunner.Message = pipelineConfig.NEW_DEPLOYMENT_INITIATED
		previousRunner.Status = pipelineConfig.WorkflowFailed
		previousRunner.UpdatedOn = time.Now()
		previousRunner.UpdatedBy = triggeredBy
		timeline := &pipelineConfig.PipelineStatusTimeline{
			CdWorkflowRunnerId: previousRunner.Id,
			Status:             pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED,
			StatusDetail:       "This deployment is superseded.",
			StatusTime:         time.Now(),
			AuditLog: sql.AuditLog{
				CreatedBy: 1,
				CreatedOn: time.Now(),
				UpdatedBy: 1,
				UpdatedOn: time.Now(),
			},
		}
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

func (impl *CdWorkflowCommonServiceImpl) MarkCurrentDeploymentFailed(runner *pipelineConfig.CdWorkflowRunner, releaseErr error, triggeredBy int32) error {
	err := impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineFailed(runner.Id, extractTimelineFailedStatusDetails(releaseErr))
	if err != nil {
		impl.logger.Errorw("error updating CdPipelineStatusTimeline", "err", err, "releaseErr", releaseErr)
		return err
	}
	//update current WF with error status
	impl.logger.Errorw("error in triggering cd WF, setting wf status as fail ", "wfId", runner.Id, "err", releaseErr)
	runner.Status = pipelineConfig.WorkflowFailed
	runner.Message = util.GetClientErrorDetailedMessage(releaseErr)
	runner.FinishedOn = time.Now()
	runner.UpdatedOn = time.Now()
	runner.UpdatedBy = triggeredBy
	err1 := impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
	if err1 != nil {
		impl.logger.Errorw("error updating cd wf runner status", "err", releaseErr, "currentRunner", runner)
		return err1
	}
	util4.TriggerCDMetrics(pipelineConfig.GetTriggerMetricsFromRunnerObj(runner), impl.config.ExposeCDMetrics)
	return nil
}

func (impl *CdWorkflowCommonServiceImpl) UpdateCDWorkflowRunnerStatus(ctx context.Context, wfrId int, userId int32, status string, options ...adapter.UpdateOptions) error {
	// In case of terminal status update finished on time
	isTerminalStatus := slices.Contains(pipelineConfig.WfrTerminalStatusList, status)
	cdWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(wfrId)
	if err != nil {
		impl.logger.Errorw("err on fetching cd workflow, UpdateCDWorkflowRunnerStatus", "err", err)
		return err
	}
	// if the current cdWfr status is already a terminal status and then don't update the status
	// e.g: Status : Failed --> Progressing (not allowed)
	if slices.Contains(pipelineConfig.WfrTerminalStatusList, cdWfr.Status) {
		impl.logger.Warnw("deployment has already been terminated for workflow runner, UpdateCDWorkflowRunnerStatus", "workflowRunnerId", cdWfr.Id, "err", err)
		return fmt.Errorf("deployment has already been terminated for workflow runner")
	}
	for _, option := range options {
		option(cdWfr)
	}
	if status == pipelineConfig.WorkflowFailed {
		err = impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineFailed(cdWfr.Id, cdWfr.Message)
		if err != nil {
			impl.logger.Errorw("error updating CdPipelineStatusTimeline", "err", err)
			return err
		}
	}
	cdWfr.Status = status
	if isTerminalStatus {
		cdWfr.FinishedOn = time.Now()
	}
	cdWfr.UpdateAuditLog(userId)
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(cdWfr)
	if err != nil {
		impl.logger.Errorw("error on update cd workflow runner, UpdateCDWorkflowRunnerStatus", "cdWfr", cdWfr, "err", err)
		return err
	}
	if isTerminalStatus {
		util4.TriggerCDMetrics(pipelineConfig.GetTriggerMetricsFromRunnerObj(cdWfr), impl.config.ExposeCDMetrics)
	}
	return nil
}

func extractTimelineFailedStatusDetails(err error) string {
	errorString := util.GetClientErrorDetailedMessage(err)
	switch errorString {
	case pipelineConfig.FOUND_VULNERABILITY:
		return pipelineConfig.TIMELINE_DESCRIPTION_VULNERABLE_IMAGE
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
