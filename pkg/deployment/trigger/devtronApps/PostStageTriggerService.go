package devtronApps

import (
	"context"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/util/event"
	"strconv"
	"time"
)

func (impl *TriggerServiceImpl) TriggerPostStage(request bean.TriggerRequest) error {
	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	triggeredBy := request.TriggeredBy
	pipeline := request.Pipeline
	cdWf := request.CdWf

	runner := &pipelineConfig.CdWorkflowRunner{
		Name:                  pipeline.Name,
		WorkflowType:          bean2.CD_WORKFLOW_TYPE_POST,
		ExecutorType:          impl.config.GetWorkflowExecutorType(),
		Status:                pipelineConfig.WorkflowStarting, // starting PostStage
		TriggeredBy:           triggeredBy,
		StartedOn:             triggeredAt,
		Namespace:             impl.config.GetDefaultNamespace(),
		BlobStorageEnabled:    impl.config.BlobStorageEnabled,
		CdWorkflowId:          cdWf.Id,
		LogLocation:           fmt.Sprintf("%s/%s%s-%s/main.log", impl.config.GetDefaultBuildLogsKeyPrefix(), strconv.Itoa(cdWf.Id), string(bean2.CD_WORKFLOW_TYPE_POST), pipeline.Name),
		AuditLog:              sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: triggeredBy, UpdatedOn: triggeredAt, UpdatedBy: triggeredBy},
		RefCdWorkflowRunnerId: request.RefCdWorkflowRunnerId,
		ReferenceId:           request.TriggerContext.ReferenceId,
	}
	var env *repository2.Environment
	var err error
	if pipeline.RunPostStageInEnv {
		env, err = impl.envRepository.FindById(pipeline.EnvironmentId)
		if err != nil {
			impl.logger.Errorw(" unable to find env ", "err", err)
			return err
		}
		runner.Namespace = env.Namespace
	}

	_, err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	if err != nil {
		return err
	}

	if cdWf.CiArtifact == nil || cdWf.CiArtifact.Id == 0 {
		cdWf.CiArtifact, err = impl.ciArtifactRepository.Get(cdWf.CiArtifactId)
		if err != nil {
			impl.logger.Errorw("error fetching artifact data", "err", err)
			return err
		}
	}
	// Migration of deprecated DataSource Type
	if cdWf.CiArtifact.IsMigrationRequired() {
		migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(cdWf.CiArtifact.Id)
		if migrationErr != nil {
			impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", cdWf.CiArtifact.Id)
		}
	}
	//checking vulnerability for the selected image
	isVulnerable, err := impl.GetArtifactVulnerabilityStatus(cdWf.CiArtifact, pipeline, context.Background())
	if err != nil {
		impl.logger.Errorw("error in getting Artifact vulnerability status, TriggerPostStage", "err", err)
		return err
	}
	if isVulnerable {
		// if image vulnerable, update timeline status and return
		runner.Status = pipelineConfig.WorkflowFailed
		runner.Message = pipelineConfig.FOUND_VULNERABILITY
		runner.FinishedOn = time.Now()
		runner.UpdatedOn = time.Now()
		runner.UpdatedBy = triggeredBy
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		if err != nil {
			impl.logger.Errorw("error in updating wfr status due to vulnerable image", "err", err)
			return err
		}
		return fmt.Errorf("found vulnerability for image digest %s", cdWf.CiArtifact.ImageDigest)
	}

	cdStageWorkflowRequest, err := impl.buildWFRequest(runner, cdWf, pipeline, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error in building wfRequest", "err", err, "runner", runner, "cdWf", cdWf, "pipeline", pipeline)
		return err
	}
	cdStageWorkflowRequest.StageType = types.POST
	cdStageWorkflowRequest.Pipeline = pipeline
	cdStageWorkflowRequest.Env = env
	cdStageWorkflowRequest.Type = bean3.CD_WORKFLOW_PIPELINE_TYPE
	// handling plugin specific logic

	pluginImagePathReservationIds, err := impl.SetCopyContainerImagePluginDataInWorkflowRequest(cdStageWorkflowRequest, pipeline.Id, types.POST, cdWf.CiArtifact)
	if err != nil {
		runner.Status = pipelineConfig.WorkflowFailed
		runner.Message = err.Error()
		_ = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		return err
	}

	_, err = impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest)
	if err != nil {
		impl.logger.Errorw("error in submitting workflow", "err", err, "cdStageWorkflowRequest", cdStageWorkflowRequest, "pipeline", pipeline, "env", env)
		return err
	}

	wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(context.Background(), cdWf.Id, bean2.CD_WORKFLOW_TYPE_POST)
	if err != nil {
		impl.logger.Errorw("error in getting wfr by workflowId and runnerType", "err", err, "wfId", cdWf.Id)
		return err
	}
	wfr.ImagePathReservationIds = pluginImagePathReservationIds
	err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(&wfr)
	if err != nil {
		impl.logger.Error("error in updating image path reservation ids in cd workflow runner", "err", "err")
	}

	event := impl.eventFactory.Build(util2.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util2.CD)
	impl.logger.Debugw("event Cd Post Trigger", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, &wfr, 0, bean2.CD_WORKFLOW_TYPE_POST)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	//creating cd config history entry
	err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, nil, repository3.POST_CD_TYPE, true, triggeredBy, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in creating post cd script entry", "err", err, "pipeline", pipeline)
		return err
	}
	return nil
}
