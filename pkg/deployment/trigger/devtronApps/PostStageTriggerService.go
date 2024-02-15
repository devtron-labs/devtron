package devtronApps

import (
	"context"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	util2 "github.com/devtron-labs/devtron/util/event"
	"time"
)

func (impl *TriggerServiceImpl) TriggerPostStage(request bean.TriggerRequest) error {
	request.WorkflowType = bean2.CD_WORKFLOW_TYPE_POST
	//setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	triggeredBy := request.TriggeredBy
	pipeline := request.Pipeline
	cdWf := request.CdWf
	ctx := context.Background() //before there was only one context. To check why here we are not using ctx from request.TriggerContext
	env, namespace, err := impl.getEnvAndNsIfRunStageInEnv(ctx, request)
	if err != nil {
		impl.logger.Errorw("error, getEnvAndNsIfRunStageInEnv", "err", err, "pipeline", pipeline, "stage", request.WorkflowType)
		return nil
	}
	request.RunStageInEnvNamespace = namespace
	cdWf, runner, err := impl.createStartingWfAndRunner(request, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in creating wf starting and runner entry", "err", err, "request", request)
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

	err = impl.checkVulnerabilityStatusAndFailWfIfNeeded(ctx, cdWf.CiArtifact, pipeline, runner, triggeredBy)
	if err != nil {
		impl.logger.Errorw("error, checkVulnerabilityStatusAndFailWfIfNeeded", "err", err, "runner", runner)
		return err
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
