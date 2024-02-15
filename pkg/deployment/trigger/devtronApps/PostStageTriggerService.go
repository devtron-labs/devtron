package devtronApps

import (
	"context"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	bean9 "github.com/devtron-labs/devtron/pkg/eventProcessor/out/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"strings"
	"time"
)

func (impl *TriggerServiceImpl) TriggerPostStage(request bean.TriggerRequest) error {
	request.WorkflowType = bean2.CD_WORKFLOW_TYPE_POST
	// setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	triggeredBy := request.TriggeredBy
	pipeline := request.Pipeline
	cdWf := request.CdWf
	ctx := context.Background() //before there was only one context. To check why here we are not using ctx from request.TriggerContext
	var env *repository2.Environment
	var err error
	env, err = impl.envRepository.FindById(pipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorw(" unable to find env ", "err", err)
		return err
	}
	env, namespace, err := impl.getEnvAndNsIfRunStageInEnv(ctx, request)
	if err != nil {
		impl.logger.Errorw("error, getEnvAndNsIfRunStageInEnv", "err", err, "pipeline", pipeline, "stage", request.WorkflowType)
		return nil
	}
	request.RunStageInEnvNamespace = namespace

	// Todo - optimize
	app, err := impl.appRepository.FindById(pipeline.AppId)
	if err != nil {
		return err
	}
	postStage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(pipeline.Id, repository4.PIPELINE_STAGE_TYPE_POST_CD)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching CD pipeline stage", "cdPipelineId", pipeline.Id, "stage", repository4.PIPELINE_STAGE_TYPE_POST_CD, "err", err)
		return err
	}
	// this will handle the scenario when post stage yaml is not migrated yet into pipeline stage table
	pipelineStageType := resourceFilter.PostPipelineStageYaml
	stageId := pipeline.Id
	if postStage != nil {
		pipelineStageType = resourceFilter.PipelineStage
		stageId = postStage.Id
	}

	scope := resourceQualifiers.Scope{AppId: pipeline.AppId, EnvId: pipeline.EnvironmentId, ClusterId: env.ClusterId, ProjectId: app.TeamId, IsProdEnv: env.Default}
	impl.logger.Infow("scope for auto trigger ", "scope", scope)
	filters, err := impl.resourceFilterService.GetFiltersByScope(scope)
	if err != nil {
		impl.logger.Errorw("error in getting resource filters for the pipeline", "pipelineId", pipeline.Id, "err", err)
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
	// get releaseTags from imageTaggingService
	imageTagNames, err := impl.imageTaggingService.GetTagNamesByArtifactId(cdWf.CiArtifactId)
	if err != nil {
		impl.logger.Errorw("error in getting image tags for the given artifact id", "artifactId", cdWf.CiArtifactId, "err", err)
		return err
	}

	// evaluate filters
	filterState, filterIdVsState, err := impl.resourceFilterService.CheckForResource(filters, cdWf.CiArtifact.Image, imageTagNames)
	if err != nil {
		return err
	}
	// store evaluated result
	filterEvaluationAudit, err := impl.resourceFilterService.CreateFilterEvaluationAudit(resourceFilter.Artifact, cdWf.CiArtifact.Id, pipelineStageType, stageId, filters, filterIdVsState)
	if err != nil {
		impl.logger.Errorw("error in creating filter evaluation audit data cd post stage trigger", "err", err, "cdPipelineId", pipeline.Id, "artifactId", cdWf.CiArtifact.Id)
		return err
	}

	// allow or block w.r.t filterState
	if filterState != resourceFilter.ALLOW {
		return fmt.Errorf("the artifact does not pass filtering condition")
	}
	cdWf, runner, err := impl.createStartingWfAndRunner(request, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in creating wf starting and runner entry", "err", err, "request", request)
		return err
	}

	if filterEvaluationAudit != nil {
		// update resource_filter_evaluation entry with wfrId and type
		err = impl.resourceFilterService.UpdateFilterEvaluationAuditRef(filterEvaluationAudit.Id, resourceFilter.CdWorkflowRunner, runner.Id)
		if err != nil {
			impl.logger.Errorw("error in updating filter evaluation audit reference", "filterEvaluationAuditId", filterEvaluationAudit.Id, "err", err)
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
	// checking vulnerability for the selected image
	err = impl.checkVulnerabilityStatusAndFailWfIfNeeded(context.Background(), cdWf.CiArtifact, pipeline, runner, triggeredBy)
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

	_, jobHelmPackagePath, err := impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest)
	if err != nil {
		impl.logger.Errorw("error in submitting workflow", "err", err, "cdStageWorkflowRequest", cdStageWorkflowRequest, "pipeline", pipeline, "env", env)
		return err
	}
	if pipeline.App.Id == 0 {
		appDbObject, err := impl.appRepository.FindById(pipeline.AppId)
		if err != nil {
			impl.logger.Errorw("error in getting app by appId", "err", err)
			return err
		}
		pipeline.App = *appDbObject
	}
	if pipeline.Environment.Id == 0 {
		envDbObject, err := impl.envRepository.FindById(pipeline.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("error in getting env by envId", "err", err)
			return err
		}
		pipeline.Environment = *envDbObject
	}
	imageTag := strings.Split(cdStageWorkflowRequest.CiArtifactDTO.Image, ":")[1]
	chartName := fmt.Sprintf("%s-%s-%s-%s", "post", pipeline.App.AppName, pipeline.Environment.Name, imageTag)

	if util.IsManifestDownload(pipeline.DeploymentAppType) || util.IsManifestPush(pipeline.DeploymentAppType) {
		chartBytes, err := impl.chartTemplateService.LoadChartInBytes(jobHelmPackagePath, false, chartName, fmt.Sprint(cdWf.Id))
		if err != nil {
			return err
		}
		if util.IsManifestPush(pipeline.DeploymentAppType) {
			err = impl.PushPrePostCDManifest(runner.Id, triggeredBy, jobHelmPackagePath, types.POST, pipeline, imageTag, context.Background())
			if err != nil {
				runner.Status = pipelineConfig.WorkflowFailed
				runner.UpdatedBy = triggeredBy
				runner.UpdatedOn = triggeredAt
				runner.FinishedOn = time.Now()
				saveRunnerErr := impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
				if saveRunnerErr != nil {
					impl.logger.Errorw("error in saving runner object in db", "err", saveRunnerErr)
				}
				impl.logger.Errorw("error in pushing manifest to helm repo", "err", err)
				return err
			}
		}
		runner.Status = pipelineConfig.WorkflowSucceeded
		runner.UpdatedBy = triggeredBy
		runner.UpdatedOn = triggeredAt
		runner.FinishedOn = time.Now()
		runner.HelmReferenceChart = chartBytes
		err = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		if err != nil {
			impl.logger.Errorw("error in saving runner object in DB", "err", err)
			return err
		}
		// Auto Trigger after Post Stage Success Event
		//TODO: update
		cdSuccessEvent := bean9.DeployStageSuccessEventReq{
			CdWorkflowId:               runner.CdWorkflowId,
			PipelineId:                 pipeline.CiPipelineId,
			PluginRegistryImageDetails: nil,
		}
		go impl.workflowEventPublishService.PublishDeployStageSuccessEvent(cdSuccessEvent)
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
	// creating cd config history entry
	err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, nil, repository3.POST_CD_TYPE, true, triggeredBy, triggeredAt)
	if err != nil {
		impl.logger.Errorw("error in creating post cd script entry", "err", err, "pipeline", pipeline)
		return err
	}
	return nil
}
