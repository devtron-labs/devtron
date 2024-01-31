package cd

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	util4 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"slices"
	"time"
)

type CdWorkflowCommonService interface {
	UpdateCDWorkflowRunnerStatus(ctx context.Context, overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time, status, message string) error
}

type CdWorkflowCommonServiceImpl struct {
	logger                        *zap.SugaredLogger
	cdWorkflowRepository          pipelineConfig.CdWorkflowRepository
	pipelineStatusTimelineService status.PipelineStatusTimelineService

	//TODO: remove below
	config             *types.CdConfig
	pipelineRepository pipelineConfig.PipelineRepository
}

func NewCdWorkflowCommonServiceImpl(logger *zap.SugaredLogger,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	pipelineRepository pipelineConfig.PipelineRepository) (*CdWorkflowCommonServiceImpl, error) {
	config, err := types.GetCdConfig()
	if err != nil {
		return nil, err
	}
	return &CdWorkflowCommonServiceImpl{
		logger:                        logger,
		cdWorkflowRepository:          cdWorkflowRepository,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
		config:                        config,
		pipelineRepository:            pipelineRepository,
	}, nil
}

// TODO: remove bean.ValuesOverrideRequest
func (impl *CdWorkflowCommonServiceImpl) UpdateCDWorkflowRunnerStatus(ctx context.Context, overrideRequest *bean.ValuesOverrideRequest, triggeredAt time.Time, status, message string) error {
	// In case of terminal status update finished on time
	isTerminalStatus := slices.Contains(pipelineConfig.WfrTerminalStatusList, status)
	cdWfr, err := impl.cdWorkflowRepository.FindWorkflowRunnerById(overrideRequest.WfrId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err on fetching cd workflow, UpdateCDWorkflowRunnerStatus", "err", err)
		return err
	}
	cdWorkflowId := cdWfr.CdWorkflowId

	if cdWorkflowId == 0 {
		cdWf := &pipelineConfig.CdWorkflow{
			CiArtifactId: overrideRequest.CiArtifactId,
			PipelineId:   overrideRequest.PipelineId,
			AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
		}
		err := impl.cdWorkflowRepository.SaveWorkFlow(ctx, cdWf)
		if err != nil {
			impl.logger.Errorw("err on updating cd workflow for status update, UpdateCDWorkflowRunnerStatus", "err", err)
			return err
		}
		cdWorkflowId = cdWf.Id
		runner := &pipelineConfig.CdWorkflowRunner{
			Id:           cdWf.Id,
			Name:         overrideRequest.PipelineName,
			WorkflowType: bean.CD_WORKFLOW_TYPE_DEPLOY,
			ExecutorType: pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF,
			Status:       status,
			TriggeredBy:  overrideRequest.UserId,
			StartedOn:    triggeredAt,
			CdWorkflowId: cdWorkflowId,
			AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: overrideRequest.UserId, UpdatedOn: triggeredAt, UpdatedBy: overrideRequest.UserId},
		}
		if isTerminalStatus {
			runner.FinishedOn = time.Now()
		}
		_, err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
		if err != nil {
			impl.logger.Errorw("err on updating cd workflow runner for status update, UpdateCDWorkflowRunnerStatus", "err", err)
			return err
		}
	} else {
		// if the current cdWfr status is already a terminal status and then don't update the status
		// e.g: Status : Failed --> Progressing (not allowed)
		if slices.Contains(pipelineConfig.WfrTerminalStatusList, cdWfr.Status) {
			impl.logger.Warnw("deployment has already been terminated for workflow runner, UpdateCDWorkflowRunnerStatus", "workflowRunnerId", cdWfr.Id, "err", err)
			return fmt.Errorf("deployment has already been terminated for workflow runner")
		}
		if status == pipelineConfig.WorkflowFailed {
			err = impl.pipelineStatusTimelineService.MarkPipelineStatusTimelineFailed(cdWfr.Id, message)
			if err != nil {
				impl.logger.Errorw("error updating CdPipelineStatusTimeline", "err", err)
				return err
			}
		}
		cdWfr.Status = status
		if isTerminalStatus {
			cdWfr.FinishedOn = time.Now()
			cdWfr.Message = message
		}
		cdWfr.UpdatedBy = overrideRequest.UserId
		cdWfr.UpdatedOn = time.Now()
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(cdWfr)
		if err != nil {
			impl.logger.Errorw("error on update cd workflow runner, UpdateCDWorkflowRunnerStatus", "cdWfr", cdWfr, "err", err)
			return err
		}
	}
	if isTerminalStatus {
		if cdWfr.CdWorkflow == nil {
			pipeline, err := impl.pipelineRepository.FindById(overrideRequest.PipelineId)
			if err != nil {
				impl.logger.Errorw("error in fetching cd pipeline", "pipelineId", overrideRequest.PipelineId, "err", err)
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
