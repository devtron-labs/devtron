package devtronApps

import (
	"context"
	"encoding/json"
	"fmt"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	gitSensorClient "github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	bean4 "github.com/devtron-labs/devtron/pkg/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/bean"
	"github.com/devtron-labs/devtron/pkg/imageDigestPolicy"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/plugin"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/sql"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	repository5 "github.com/devtron-labs/devtron/pkg/variables/repository"
	util4 "github.com/devtron-labs/devtron/util"
	util2 "github.com/devtron-labs/devtron/util/event"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"strconv"
	"strings"
	"time"
)

const (
	GIT_COMMIT_HASH_PREFIX       = "GIT_COMMIT_HASH"
	GIT_SOURCE_TYPE_PREFIX       = "GIT_SOURCE_TYPE"
	GIT_SOURCE_VALUE_PREFIX      = "GIT_SOURCE_VALUE"
	GIT_SOURCE_COUNT             = "GIT_SOURCE_COUNT"
	APP_LABEL_KEY_PREFIX         = "APP_LABEL_KEY"
	APP_LABEL_VALUE_PREFIX       = "APP_LABEL_VALUE"
	APP_LABEL_COUNT              = "APP_LABEL_COUNT"
	CHILD_CD_ENV_NAME_PREFIX     = "CHILD_CD_ENV_NAME"
	CHILD_CD_CLUSTER_NAME_PREFIX = "CHILD_CD_CLUSTER_NAME"
	CHILD_CD_COUNT               = "CHILD_CD_COUNT"
)

func (impl *TriggerServiceImpl) TriggerPreStage(request bean.TriggerRequest) error {
	// setting triggeredAt variable to have consistent data for various audit log places in db for deployment time
	triggeredAt := time.Now()
	triggeredBy := request.TriggeredBy
	artifact := request.Artifact
	pipeline := request.Pipeline
	ctx := request.TriggerContext.Context
	// in case of pre stage manual trigger auth is already applied and for auto triggers there is no need for auth check here
	cdWf := request.CdWf

	var env *repository2.Environment
	var err error
	_, span := otel.Tracer("orchestrator").Start(ctx, "envRepository.FindById")
	env, err = impl.envRepository.FindById(pipeline.EnvironmentId)
	span.End()
	if err != nil {
		impl.logger.Errorw(" unable to find env ", "err", err)
		return err
	}

	app, err := impl.appRepository.FindById(pipeline.AppId)
	if err != nil {
		return err
	}
	preStage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(pipeline.Id, repository4.PIPELINE_STAGE_TYPE_PRE_CD)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching CD pipeline stage", "cdPipelineId", pipeline.Id, "stage", repository4.PIPELINE_STAGE_TYPE_PRE_CD, "err", err)
		return err
	}
	// this will handle the scenario when post stage yaml is not migrated yet into pipeline stage table
	pipelineStageType := resourceFilter.PrePipelineStageYaml
	stageId := pipeline.Id
	if preStage != nil {
		pipelineStageType = resourceFilter.PipelineStage
		stageId = preStage.Id
	}

	scope := resourceQualifiers.Scope{AppId: pipeline.AppId, EnvId: pipeline.EnvironmentId, ClusterId: env.ClusterId, ProjectId: app.TeamId, IsProdEnv: env.Default}
	impl.logger.Infow("scope for auto trigger ", "scope", scope)
	filters, err := impl.resourceFilterService.GetFiltersByScope(scope)
	if err != nil {
		impl.logger.Errorw("error in getting resource filters for the pipeline", "pipelineId", pipeline.Id, "err", err)
		return err
	}
	// get releaseTags from imageTaggingService
	imageTagNames, err := impl.imageTaggingService.GetTagNamesByArtifactId(artifact.Id)
	if err != nil {
		impl.logger.Errorw("error in getting image tags for the given artifact id", "artifactId", artifact.Id, "err", err)
		return err
	}

	filterState, filterIdVsState, err := impl.resourceFilterService.CheckForResource(filters, artifact.Image, imageTagNames)
	if err != nil {
		return err
	}

	// store evaluated result
	filterEvaluationAudit, err := impl.resourceFilterService.CreateFilterEvaluationAudit(resourceFilter.Artifact, artifact.Id, pipelineStageType, stageId, filters, filterIdVsState)
	if err != nil {
		impl.logger.Errorw("error in creating filter evaluation audit data cd pre stage trigger", "err", err, "cdPipelineId", pipeline.Id, "artifactId", artifact.Id, "preStageId", preStage.Id)
		return err
	}

	// allow or block w.r.t filterState
	if filterState != resourceFilter.ALLOW {
		return fmt.Errorf("the artifact does not pass filtering condition")
	}

	if cdWf == nil {
		cdWf = &pipelineConfig.CdWorkflow{
			CiArtifactId: artifact.Id,
			PipelineId:   pipeline.Id,
			AuditLog:     sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: 1, UpdatedOn: triggeredAt, UpdatedBy: 1},
		}
		err = impl.cdWorkflowRepository.SaveWorkFlow(ctx, cdWf)
		if err != nil {
			return err
		}
	}
	cdWorkflowExecutorType := impl.config.GetWorkflowExecutorType()
	runner := &pipelineConfig.CdWorkflowRunner{
		Name:                  pipeline.Name,
		WorkflowType:          bean2.CD_WORKFLOW_TYPE_PRE,
		ExecutorType:          cdWorkflowExecutorType,
		Status:                pipelineConfig.WorkflowStarting, // starting PreStage
		TriggeredBy:           triggeredBy,
		StartedOn:             triggeredAt,
		Namespace:             impl.config.GetDefaultNamespace(),
		BlobStorageEnabled:    impl.config.BlobStorageEnabled,
		CdWorkflowId:          cdWf.Id,
		LogLocation:           fmt.Sprintf("%s/%s%s-%s/main.log", impl.config.GetDefaultBuildLogsKeyPrefix(), strconv.Itoa(cdWf.Id), string(bean2.CD_WORKFLOW_TYPE_PRE), pipeline.Name),
		AuditLog:              sql.AuditLog{CreatedOn: triggeredAt, CreatedBy: 1, UpdatedOn: triggeredAt, UpdatedBy: 1},
		RefCdWorkflowRunnerId: request.RefCdWorkflowRunnerId,
		ReferenceId:           request.TriggerContext.ReferenceId,
	}
	if pipeline.RunPreStageInEnv {
		impl.logger.Debugw("env", "env", env)
		runner.Namespace = env.Namespace
	}
	_, span1 := otel.Tracer("orchestrator").Start(ctx, "cdWorkflowRepository.SaveWorkFlowRunner")
	_, err = impl.cdWorkflowRepository.SaveWorkFlowRunner(runner)
	span1.End()
	if err != nil {
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

	// checking vulnerability for the selected image
	isVulnerable, err := impl.GetArtifactVulnerabilityStatus(artifact, pipeline, ctx)
	if err != nil {
		impl.logger.Errorw("error in getting Artifact vulnerability status, TriggerPreStage", "err", err)
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
		return fmt.Errorf("found vulnerability for image digest %s", artifact.ImageDigest)
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "buildWFRequest")
	cdStageWorkflowRequest, err := impl.buildWFRequest(runner, cdWf, pipeline, triggeredBy)
	span.End()
	if err != nil {
		return err
	}
	cdStageWorkflowRequest.StageType = types.PRE
	// handling copyContainerImage plugin specific logic
	imagePathReservationIds, err := impl.SetCopyContainerImagePluginDataInWorkflowRequest(cdStageWorkflowRequest, pipeline.Id, types.PRE, artifact)
	if err != nil {
		runner.Status = pipelineConfig.WorkflowFailed
		runner.Message = err.Error()
		_ = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
		return err
	} else {
		runner.ImagePathReservationIds = imagePathReservationIds
		_ = impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "cdWorkflowService.SubmitWorkflow")
	cdStageWorkflowRequest.Pipeline = pipeline
	cdStageWorkflowRequest.Env = env
	cdStageWorkflowRequest.Type = bean3.CD_WORKFLOW_PIPELINE_TYPE
	_, jobHelmPackagePath, err := impl.cdWorkflowService.SubmitWorkflow(cdStageWorkflowRequest)
	span.End()
	if err != nil {
		return err
	}
	if util.IsManifestDownload(pipeline.DeploymentAppType) || util.IsManifestPush(pipeline.DeploymentAppType) {
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
		deleteChart := !util.IsManifestPush(pipeline.DeploymentAppType)
		imageTag := strings.Split(artifact.Image, ":")[1]
		chartName := fmt.Sprintf("%s-%s-%s-%s", "pre", pipeline.App.AppName, pipeline.Environment.Name, imageTag)
		chartBytes, err := impl.chartTemplateService.LoadChartInBytes(jobHelmPackagePath, deleteChart, chartName, fmt.Sprint(cdWf.Id))
		if err != nil && util.IsManifestDownload(pipeline.DeploymentAppType) {
			return err
		}
		if util.IsManifestPush(pipeline.DeploymentAppType) {
			err = impl.PushPrePostCDManifest(runner.Id, triggeredBy, jobHelmPackagePath, types.PRE, pipeline, imageTag, ctx)
			if err != nil {
				runner.Status = pipelineConfig.WorkflowFailed
				runner.UpdatedBy = triggeredBy
				runner.UpdatedOn = triggeredAt
				runner.FinishedOn = time.Now()
				runnerSaveErr := impl.cdWorkflowRepository.UpdateWorkFlowRunner(runner)
				if runnerSaveErr != nil {
					impl.logger.Errorw("error in saving runner object in db", "err", runnerSaveErr)
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
			impl.logger.Errorw("error in saving runner object in db", "err", err)
			return err
		}
		// Handle auto trigger after pre stage success event
		go impl.TriggerAutoCDOnPreStageSuccess(request.TriggerContext, pipeline.Id, artifact.Id, cdWf.Id, triggeredBy)
	}

	err = impl.sendPreStageNotification(ctx, cdWf, pipeline)
	if err != nil {
		return err
	}
	// creating cd config history entry
	_, span = otel.Tracer("orchestrator").Start(ctx, "prePostCdScriptHistoryService.CreatePrePostCdScriptHistory")
	err = impl.prePostCdScriptHistoryService.CreatePrePostCdScriptHistory(pipeline, nil, repository3.PRE_CD_TYPE, true, triggeredBy, triggeredAt)
	span.End()
	if err != nil {
		impl.logger.Errorw("error in creating pre cd script entry", "err", err, "pipeline", pipeline)
		return err
	}
	return nil
}

func (impl *TriggerServiceImpl) SetCopyContainerImagePluginDataInWorkflowRequest(cdStageWorkflowRequest *types.WorkflowRequest, pipelineId int, pipelineStage string, artifact *repository.CiArtifact) ([]int, error) {
	copyContainerImagePluginId, err := impl.globalPluginService.GetRefPluginIdByRefPluginName(pipeline.COPY_CONTAINER_IMAGE)
	var imagePathReservationIds []int
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting copyContainerImage plugin id", "err", err)
		return imagePathReservationIds, err
	}
	for _, step := range cdStageWorkflowRequest.PrePostDeploySteps {
		if copyContainerImagePluginId != 0 && step.RefPluginId == copyContainerImagePluginId {
			var pipelineStageEntityType int
			if pipelineStage == types.PRE {
				pipelineStageEntityType = bean3.EntityTypePreCD
			} else {
				pipelineStageEntityType = bean3.EntityTypePostCD
			}
			customTagId := -1
			var DockerImageTag string

			customTag, err := impl.customTagService.GetActiveCustomTagByEntityKeyAndValue(pipelineStageEntityType, strconv.Itoa(pipelineId))
			if err != nil && err != pg.ErrNoRows {
				impl.logger.Errorw("error in fetching custom tag data", "err", err)
				return imagePathReservationIds, err
			}

			if !customTag.Enabled {
				// case when custom tag is not configured - source image tag will be taken as docker image tag
				pluginTriggerImageSplit := strings.Split(artifact.Image, ":")
				DockerImageTag = pluginTriggerImageSplit[len(pluginTriggerImageSplit)-1]
			} else {
				// for copyContainerImage plugin parse destination images and save its data in image path reservation table
				customTagDbObject, customDockerImageTag, err := impl.customTagService.GetCustomTag(pipelineStageEntityType, strconv.Itoa(pipelineId))
				if err != nil && err != pg.ErrNoRows {
					impl.logger.Errorw("error in fetching custom tag by entity key and value for CD", "err", err)
					return imagePathReservationIds, err
				}
				if customTagDbObject != nil && customTagDbObject.Id > 0 {
					customTagId = customTagDbObject.Id
				}
				DockerImageTag = customDockerImageTag
			}

			var sourceDockerRegistryId string
			if artifact.DataSource == repository.PRE_CD || artifact.DataSource == repository.POST_CD || artifact.DataSource == repository.POST_CI {
				if artifact.CredentialsSourceType == repository.GLOBAL_CONTAINER_REGISTRY {
					sourceDockerRegistryId = artifact.CredentialSourceValue
				}
			} else {
				sourceDockerRegistryId = cdStageWorkflowRequest.DockerRegistryId
			}
			registryDestinationImageMap, registryCredentialMap, err := impl.pluginInputVariableParser.HandleCopyContainerImagePluginInputVariables(step.InputVars, DockerImageTag, cdStageWorkflowRequest.CiArtifactDTO.Image, sourceDockerRegistryId)
			if err != nil {
				impl.logger.Errorw("error in parsing copyContainerImage input variable", "err", err)
				return imagePathReservationIds, err
			}
			var destinationImages []string
			for _, images := range registryDestinationImageMap {
				for _, image := range images {
					destinationImages = append(destinationImages, image)
				}
			}
			// fetch already saved artifacts to check if they are already present
			savedCIArtifacts, err := impl.ciArtifactRepository.FindCiArtifactByImagePaths(destinationImages)
			if err != nil {
				impl.logger.Errorw("error in fetching artifacts by image path", "err", err)
				return imagePathReservationIds, err
			}
			if len(savedCIArtifacts) > 0 {
				// if already present in ci artifact, return "image path already in use error"
				return imagePathReservationIds, bean3.ErrImagePathInUse
			}
			imagePathReservationIds, err = impl.ReserveImagesGeneratedAtPlugin(customTagId, registryDestinationImageMap)
			if err != nil {
				impl.logger.Errorw("error in reserving image", "err", err)
				return imagePathReservationIds, err
			}
			cdStageWorkflowRequest.RegistryDestinationImageMap = registryDestinationImageMap
			cdStageWorkflowRequest.RegistryCredentialMap = registryCredentialMap
			var pluginArtifactStage string
			if pipelineStage == types.PRE {
				pluginArtifactStage = repository.PRE_CD
			} else {
				pluginArtifactStage = repository.POST_CD
			}
			cdStageWorkflowRequest.PluginArtifactStage = pluginArtifactStage
		}
	}
	return imagePathReservationIds, nil
}

func (impl *TriggerServiceImpl) buildWFRequest(runner *pipelineConfig.CdWorkflowRunner, cdWf *pipelineConfig.CdWorkflow, cdPipeline *pipelineConfig.Pipeline, triggeredBy int32) (*types.WorkflowRequest, error) {
	cdWorkflowConfig, err := impl.cdWorkflowRepository.FindConfigByPipelineId(cdPipeline.Id)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}

	workflowExecutor := runner.ExecutorType

	artifact, err := impl.ciArtifactRepository.Get(cdWf.CiArtifactId)
	if err != nil {
		return nil, err
	}
	// Migration of deprecated DataSource Type
	if artifact.IsMigrationRequired() {
		migrationErr := impl.ciArtifactRepository.MigrateToWebHookDataSourceType(artifact.Id)
		if migrationErr != nil {
			impl.logger.Warnw("unable to migrate deprecated DataSource", "artifactId", artifact.Id)
		}
	}
	ciMaterialInfo, err := repository.GetCiMaterialInfo(artifact.MaterialInfo, artifact.DataSource)
	if err != nil {
		impl.logger.Errorw("parsing error", "err", err)
		return nil, err
	}

	var ciProjectDetails []bean3.CiProjectDetails
	var ciPipeline *pipelineConfig.CiPipeline
	if cdPipeline.CiPipelineId > 0 {
		ciPipeline, err = impl.ciPipelineRepository.FindById(cdPipeline.CiPipelineId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.logger.Errorw("cannot find ciPipelineRequest", "err", err)
			return nil, err
		}

		for _, m := range ciPipeline.CiPipelineMaterials {
			// git material should be active in this case
			if m == nil || m.GitMaterial == nil || !m.GitMaterial.Active {
				continue
			}
			var ciMaterialCurrent repository.CiMaterialInfo
			for _, ciMaterial := range ciMaterialInfo {
				if ciMaterial.Material.GitConfiguration.URL == m.GitMaterial.Url {
					ciMaterialCurrent = ciMaterial
					break
				}
			}
			gitMaterial, err := impl.materialRepository.FindById(m.GitMaterialId)
			if err != nil && !util.IsErrNoRows(err) {
				impl.logger.Errorw("could not fetch git materials", "err", err)
				return nil, err
			}

			ciProjectDetail := bean3.CiProjectDetails{
				GitRepository:   ciMaterialCurrent.Material.GitConfiguration.URL,
				MaterialName:    gitMaterial.Name,
				CheckoutPath:    gitMaterial.CheckoutPath,
				FetchSubmodules: gitMaterial.FetchSubmodules,
				SourceType:      m.Type,
				SourceValue:     m.Value,
				Type:            string(m.Type),
				GitOptions: bean3.GitOptions{
					UserName:      gitMaterial.GitProvider.UserName,
					Password:      gitMaterial.GitProvider.Password,
					SshPrivateKey: gitMaterial.GitProvider.SshPrivateKey,
					AccessToken:   gitMaterial.GitProvider.AccessToken,
					AuthMode:      gitMaterial.GitProvider.AuthMode,
				},
			}
			if executors.IsShallowClonePossible(m, impl.config.GitProviders, impl.config.CloningMode) {
				ciProjectDetail.CloningMode = executors.CloningModeShallow
			}
			if len(ciMaterialCurrent.Modifications) > 0 {
				ciProjectDetail.CommitHash = ciMaterialCurrent.Modifications[0].Revision
				ciProjectDetail.Author = ciMaterialCurrent.Modifications[0].Author
				ciProjectDetail.GitTag = ciMaterialCurrent.Modifications[0].Tag
				ciProjectDetail.Message = ciMaterialCurrent.Modifications[0].Message
				commitTime, err := convert(ciMaterialCurrent.Modifications[0].ModifiedTime)
				if err != nil {
					return nil, err
				}
				ciProjectDetail.CommitTime = commitTime.Format(bean4.LayoutRFC3339)
			} else if ciPipeline.PipelineType == bean3.CI_JOB {
				// This has been done to resolve unmarshalling issue in ci-runner, in case of no commit time(eg- polling container images)
				ciProjectDetail.CommitTime = time.Time{}.Format(bean4.LayoutRFC3339)
			} else {
				impl.logger.Debugw("devtronbug#1062", ciPipeline.Id, cdPipeline.Id)
				return nil, fmt.Errorf("modifications not found for %d", ciPipeline.Id)
			}

			// set webhook data
			if m.Type == pipelineConfig.SOURCE_TYPE_WEBHOOK && len(ciMaterialCurrent.Modifications) > 0 {
				webhookData := ciMaterialCurrent.Modifications[0].WebhookData
				ciProjectDetail.WebhookData = pipelineConfig.WebhookData{
					Id:              webhookData.Id,
					EventActionType: webhookData.EventActionType,
					Data:            webhookData.Data,
				}
			}

			ciProjectDetails = append(ciProjectDetails, ciProjectDetail)
		}
	}
	var stageYaml string
	var deployStageWfr pipelineConfig.CdWorkflowRunner
	var deployStageTriggeredByUserEmail string
	var pipelineReleaseCounter int
	var preDeploySteps []*bean3.StepObject
	var postDeploySteps []*bean3.StepObject
	var refPluginsData []*bean3.RefPluginObject
	//if pipeline_stage_steps present for pre-CD or post-CD then no need to add stageYaml to cdWorkflowRequest in that
	//case add PreDeploySteps and PostDeploySteps to cdWorkflowRequest, this is done for backward compatibility
	pipelineStage, err := impl.pipelineStageService.GetCdStageByCdPipelineIdAndStageType(cdPipeline.Id, runner.WorkflowType.WorkflowTypeToStageType())
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching CD pipeline stage", "cdPipelineId", cdPipeline.Id, "stage ", runner.WorkflowType.WorkflowTypeToStageType(), "err", err)
		return nil, err
	}
	env, err := impl.envRepository.FindById(cdPipeline.EnvironmentId)
	if err != nil {
		impl.logger.Errorw("error in getting environment by id", "err", err)
		return nil, err
	}

	//Scope will pick the environment of CD pipeline irrespective of in-cluster mode,
	//since user sees the environment of the CD pipeline
	scope := resourceQualifiers.Scope{
		AppId:     cdPipeline.App.Id,
		EnvId:     env.Id,
		ClusterId: env.ClusterId,
		SystemMetadata: &resourceQualifiers.SystemMetadata{
			EnvironmentName: env.Name,
			ClusterName:     env.Cluster.ClusterName,
			Namespace:       env.Namespace,
			Image:           artifact.Image,
			ImageTag:        util3.GetImageTagFromImage(artifact.Image),
		},
	}
	if pipelineStage != nil {
		var variableSnapshot map[string]string
		if runner.WorkflowType == bean2.CD_WORKFLOW_TYPE_PRE {
			//TODO: use const from pipeline.WorkflowService:95
			prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(cdPipeline.Id, "preCD", scope)
			if err != nil {
				impl.logger.Errorw("error in getting pre, post & refPlugin steps data for wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}
			preDeploySteps = prePostAndRefPluginResponse.PreStageSteps
			refPluginsData = prePostAndRefPluginResponse.RefPluginData
			variableSnapshot = prePostAndRefPluginResponse.VariableSnapshot
		} else if runner.WorkflowType == bean2.CD_WORKFLOW_TYPE_POST {
			//TODO: use const from pipeline.WorkflowService:96
			prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(cdPipeline.Id, "postCD", scope)
			if err != nil {
				impl.logger.Errorw("error in getting pre, post & refPlugin steps data for wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}
			postDeploySteps = prePostAndRefPluginResponse.PostStageSteps
			refPluginsData = prePostAndRefPluginResponse.RefPluginData
			variableSnapshot = prePostAndRefPluginResponse.VariableSnapshot
			deployStageWfr, deployStageTriggeredByUserEmail, pipelineReleaseCounter, err = impl.getDeployStageDetails(cdPipeline.Id)
			if err != nil {
				impl.logger.Errorw("error in getting deployStageWfr, deployStageTriggeredByUser and pipelineReleaseCounter wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("unsupported workflow triggerd")
		}

		//Save Scoped VariableSnapshot
		var variableSnapshotHistories = util4.GetBeansPtr(
			repository5.GetSnapshotBean(runner.Id, repository5.HistoryReferenceTypeCDWORKFLOWRUNNER, variableSnapshot))
		if len(variableSnapshotHistories) > 0 {
			err = impl.scopedVariableManager.SaveVariableHistoriesForTrigger(variableSnapshotHistories, runner.TriggeredBy)
			if err != nil {
				impl.logger.Errorf("Not able to save variable snapshot for CD trigger %s %d %s", err, runner.Id, variableSnapshot)
			}
		}
	} else {
		//in this case no plugin script is not present for this cdPipeline hence going with attaching preStage or postStage config
		if runner.WorkflowType == bean2.CD_WORKFLOW_TYPE_PRE {
			stageYaml = cdPipeline.PreStageConfig
		} else if runner.WorkflowType == bean2.CD_WORKFLOW_TYPE_POST {
			stageYaml = cdPipeline.PostStageConfig
			deployStageWfr, deployStageTriggeredByUserEmail, pipelineReleaseCounter, err = impl.getDeployStageDetails(cdPipeline.Id)
			if err != nil {
				impl.logger.Errorw("error in getting deployStageWfr, deployStageTriggeredByUser and pipelineReleaseCounter wf request", "err", err, "cdPipelineId", cdPipeline.Id)
				return nil, err
			}

		} else {
			return nil, fmt.Errorf("unsupported workflow triggerd")
		}
	}

	digestConfigurationRequest := imageDigestPolicy.DigestPolicyConfigurationRequest{
		ClusterId:     env.ClusterId,
		EnvironmentId: env.Id,
		PipelineId:    cdPipeline.Id,
	}
	digestPolicyConfigurations, err := impl.imageDigestPolicyService.GetDigestPolicyConfigurations(digestConfigurationRequest)
	if err != nil {
		impl.logger.Errorw("error in checking if isImageDigestPolicyConfiguredForPipeline", "err", err, "pipelineId", cdPipeline.Id)
		return nil, err
	}
	image := artifact.Image
	if digestPolicyConfigurations.UseDigestForTrigger() {
		image = ReplaceImageTagWithDigest(image, artifact.ImageDigest)
	}
	cdStageWorkflowRequest := &types.WorkflowRequest{
		EnvironmentId:         cdPipeline.EnvironmentId,
		AppId:                 cdPipeline.AppId,
		WorkflowId:            cdWf.Id,
		WorkflowRunnerId:      runner.Id,
		WorkflowNamePrefix:    strconv.Itoa(runner.Id) + "-" + runner.Name,
		WorkflowPrefixForLog:  strconv.Itoa(cdWf.Id) + string(runner.WorkflowType) + "-" + runner.Name,
		CdImage:               impl.config.GetDefaultImage(),
		CdPipelineId:          cdWf.PipelineId,
		TriggeredBy:           triggeredBy,
		StageYaml:             stageYaml,
		CiProjectDetails:      ciProjectDetails,
		Namespace:             runner.Namespace,
		ActiveDeadlineSeconds: impl.config.GetDefaultTimeout(),
		CiArtifactDTO: types.CiArtifactDTO{
			Id:           artifact.Id,
			PipelineId:   artifact.PipelineId,
			Image:        image,
			ImageDigest:  artifact.ImageDigest,
			MaterialInfo: artifact.MaterialInfo,
			DataSource:   artifact.DataSource,
			WorkflowId:   artifact.WorkflowId,
		},
		OrchestratorHost:  impl.config.OrchestratorHost,
		OrchestratorToken: impl.config.OrchestratorToken,
		CloudProvider:     impl.config.CloudProvider,
		WorkflowExecutor:  workflowExecutor,
		RefPlugins:        refPluginsData,
		Scope:             scope,
	}

	extraEnvVariables := make(map[string]string)
	if env != nil {
		extraEnvVariables[plugin.CD_PIPELINE_ENV_NAME_KEY] = env.Name
		if env.Cluster != nil {
			extraEnvVariables[plugin.CD_PIPELINE_CLUSTER_NAME_KEY] = env.Cluster.ClusterName
		}
	}
	ciWf, err := impl.ciWorkflowRepository.FindLastTriggeredWorkflowByArtifactId(artifact.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting ciWf by artifactId", "err", err, "artifactId", artifact.Id)
		return nil, err
	}
	var webhookAndCiData *gitSensorClient.WebhookAndCiData
	if ciWf != nil && ciWf.GitTriggers != nil {
		i := 1
		var gitCommitEnvVariables []types.GitMetadata

		for ciPipelineMaterialId, gitTrigger := range ciWf.GitTriggers {
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_COMMIT_HASH_PREFIX, i)] = gitTrigger.Commit
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_SOURCE_TYPE_PREFIX, i)] = string(gitTrigger.CiConfigureSourceType)
			extraEnvVariables[fmt.Sprintf("%s_%d", GIT_SOURCE_VALUE_PREFIX, i)] = gitTrigger.CiConfigureSourceValue

			gitCommitEnvVariables = append(gitCommitEnvVariables, types.GitMetadata{
				GitCommitHash:  gitTrigger.Commit,
				GitSourceType:  string(gitTrigger.CiConfigureSourceType),
				GitSourceValue: gitTrigger.CiConfigureSourceValue,
			})

			// CODE-BLOCK starts - store extra environment variables if webhook
			if gitTrigger.CiConfigureSourceType == pipelineConfig.SOURCE_TYPE_WEBHOOK {
				webhookDataId := gitTrigger.WebhookData.Id
				if webhookDataId > 0 {
					webhookDataRequest := &gitSensorClient.WebhookDataRequest{
						Id:                   webhookDataId,
						CiPipelineMaterialId: ciPipelineMaterialId,
					}
					webhookAndCiData, err = impl.gitSensorGrpcClient.GetWebhookData(context.Background(), webhookDataRequest)
					if err != nil {
						impl.logger.Errorw("err while getting webhook data from git-sensor", "err", err, "webhookDataRequest", webhookDataRequest)
						return nil, err
					}
					if webhookAndCiData != nil {
						for extEnvVariableKey, extEnvVariableVal := range webhookAndCiData.ExtraEnvironmentVariables {
							extraEnvVariables[extEnvVariableKey] = extEnvVariableVal
						}
					}
				}
			}
			// CODE_BLOCK ends

			i++
		}
		gitMetadata, err := json.Marshal(&gitCommitEnvVariables)
		if err != nil {
			impl.logger.Errorw("err while marshaling git metdata", "err", err)
			return nil, err
		}
		extraEnvVariables[plugin.GIT_METADATA] = string(gitMetadata)

		extraEnvVariables[GIT_SOURCE_COUNT] = strconv.Itoa(len(ciWf.GitTriggers))
	}

	childCdIds, err := impl.appWorkflowRepository.FindChildCDIdsByParentCDPipelineId(cdPipeline.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting child cdPipelineIds by parent cdPipelineId", "err", err, "parent cdPipelineId", cdPipeline.Id)
		return nil, err
	}
	if len(childCdIds) > 0 {
		childPipelines, err := impl.pipelineRepository.FindByIdsIn(childCdIds)
		if err != nil {
			impl.logger.Errorw("error in getting pipelines by ids", "err", err, "ids", childCdIds)
			return nil, err
		}
		var childCdEnvVariables []types.ChildCdMetadata
		for i, childPipeline := range childPipelines {
			extraEnvVariables[fmt.Sprintf("%s_%d", CHILD_CD_ENV_NAME_PREFIX, i+1)] = childPipeline.Environment.Name
			extraEnvVariables[fmt.Sprintf("%s_%d", CHILD_CD_CLUSTER_NAME_PREFIX, i+1)] = childPipeline.Environment.Cluster.ClusterName

			childCdEnvVariables = append(childCdEnvVariables, types.ChildCdMetadata{
				ChildCdEnvName:     childPipeline.Environment.Name,
				ChildCdClusterName: childPipeline.Environment.Cluster.ClusterName,
			})
		}
		childCdEnvVariablesMetadata, err := json.Marshal(&childCdEnvVariables)
		if err != nil {
			impl.logger.Errorw("err while marshaling childCdEnvVariables", "err", err)
			return nil, err
		}
		extraEnvVariables[plugin.CHILD_CD_METADATA] = string(childCdEnvVariablesMetadata)

		extraEnvVariables[CHILD_CD_COUNT] = strconv.Itoa(len(childPipelines))
	}
	if ciPipeline != nil && ciPipeline.Id > 0 {
		extraEnvVariables["APP_NAME"] = ciPipeline.App.AppName
		cdStageWorkflowRequest.DockerUsername = ciPipeline.CiTemplate.DockerRegistry.Username
		cdStageWorkflowRequest.DockerPassword = ciPipeline.CiTemplate.DockerRegistry.Password
		cdStageWorkflowRequest.AwsRegion = ciPipeline.CiTemplate.DockerRegistry.AWSRegion
		cdStageWorkflowRequest.DockerConnection = ciPipeline.CiTemplate.DockerRegistry.Connection
		cdStageWorkflowRequest.DockerCert = ciPipeline.CiTemplate.DockerRegistry.Cert
		cdStageWorkflowRequest.AccessKey = ciPipeline.CiTemplate.DockerRegistry.AWSAccessKeyId
		cdStageWorkflowRequest.SecretKey = ciPipeline.CiTemplate.DockerRegistry.AWSSecretAccessKey
		cdStageWorkflowRequest.DockerRegistryType = string(ciPipeline.CiTemplate.DockerRegistry.RegistryType)
		cdStageWorkflowRequest.DockerRegistryURL = ciPipeline.CiTemplate.DockerRegistry.RegistryURL
		cdStageWorkflowRequest.DockerRegistryId = ciPipeline.CiTemplate.DockerRegistry.Id
		cdStageWorkflowRequest.CiPipelineType = ciPipeline.PipelineType
	} else if cdPipeline.AppId > 0 {
		ciTemplate, err := impl.ciTemplateRepository.FindByAppId(cdPipeline.AppId)
		if err != nil {
			return nil, err
		}
		extraEnvVariables["APP_NAME"] = ciTemplate.App.AppName
		cdStageWorkflowRequest.DockerUsername = ciTemplate.DockerRegistry.Username
		cdStageWorkflowRequest.DockerPassword = ciTemplate.DockerRegistry.Password
		cdStageWorkflowRequest.AwsRegion = ciTemplate.DockerRegistry.AWSRegion
		cdStageWorkflowRequest.DockerConnection = ciTemplate.DockerRegistry.Connection
		cdStageWorkflowRequest.DockerCert = ciTemplate.DockerRegistry.Cert
		cdStageWorkflowRequest.AccessKey = ciTemplate.DockerRegistry.AWSAccessKeyId
		cdStageWorkflowRequest.SecretKey = ciTemplate.DockerRegistry.AWSSecretAccessKey
		cdStageWorkflowRequest.DockerRegistryType = string(ciTemplate.DockerRegistry.RegistryType)
		cdStageWorkflowRequest.DockerRegistryURL = ciTemplate.DockerRegistry.RegistryURL
		appLabels, err := impl.appLabelRepository.FindAllByAppId(cdPipeline.AppId)
		cdStageWorkflowRequest.DockerRegistryId = ciTemplate.DockerRegistry.Id
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error in getting labels by appId", "err", err, "appId", cdPipeline.AppId)
			return nil, err
		}
		var appLabelEnvVariables []types.AppLabelMetadata
		for i, appLabel := range appLabels {
			extraEnvVariables[fmt.Sprintf("%s_%d", APP_LABEL_KEY_PREFIX, i+1)] = appLabel.Key
			extraEnvVariables[fmt.Sprintf("%s_%d", APP_LABEL_VALUE_PREFIX, i+1)] = appLabel.Value
			appLabelEnvVariables = append(appLabelEnvVariables, types.AppLabelMetadata{
				AppLabelKey:   appLabel.Key,
				AppLabelValue: appLabel.Value,
			})
		}
		if len(appLabels) > 0 {
			extraEnvVariables[APP_LABEL_COUNT] = strconv.Itoa(len(appLabels))
			appLabelEnvVariablesMetadata, err := json.Marshal(&appLabelEnvVariables)
			if err != nil {
				impl.logger.Errorw("err while marshaling appLabelEnvVariables", "err", err)
				return nil, err
			}
			extraEnvVariables[plugin.APP_LABEL_METADATA] = string(appLabelEnvVariablesMetadata)

		}
	}
	cdStageWorkflowRequest.ExtraEnvironmentVariables = extraEnvVariables
	cdStageWorkflowRequest.DeploymentTriggerTime = deployStageWfr.StartedOn
	cdStageWorkflowRequest.DeploymentTriggeredBy = deployStageTriggeredByUserEmail

	if pipelineReleaseCounter > 0 {
		cdStageWorkflowRequest.DeploymentReleaseCounter = pipelineReleaseCounter
	}
	if cdWorkflowConfig.CdCacheRegion == "" {
		cdWorkflowConfig.CdCacheRegion = impl.config.GetDefaultCdLogsBucketRegion()
	}

	if runner.WorkflowType == bean2.CD_WORKFLOW_TYPE_PRE {
		//populate input variables of steps with extra env variables
		setExtraEnvVariableInDeployStep(preDeploySteps, extraEnvVariables, webhookAndCiData)
		cdStageWorkflowRequest.PrePostDeploySteps = preDeploySteps
	} else if runner.WorkflowType == bean2.CD_WORKFLOW_TYPE_POST {
		setExtraEnvVariableInDeployStep(postDeploySteps, extraEnvVariables, webhookAndCiData)
		cdStageWorkflowRequest.PrePostDeploySteps = postDeploySteps
	}
	cdStageWorkflowRequest.BlobStorageConfigured = runner.BlobStorageEnabled
	switch cdStageWorkflowRequest.CloudProvider {
	case types.BLOB_STORAGE_S3:
		//No AccessKey is used for uploading artifacts, instead IAM based auth is used
		cdStageWorkflowRequest.CdCacheRegion = cdWorkflowConfig.CdCacheRegion
		cdStageWorkflowRequest.CdCacheLocation = cdWorkflowConfig.CdCacheBucket
		cdStageWorkflowRequest.ArtifactLocation, cdStageWorkflowRequest.CiArtifactBucket, cdStageWorkflowRequest.CiArtifactFileName = impl.buildArtifactLocationForS3(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			AccessKey:                  impl.config.BlobStorageS3AccessKey,
			Passkey:                    impl.config.BlobStorageS3SecretKey,
			EndpointUrl:                impl.config.BlobStorageS3Endpoint,
			IsInSecure:                 impl.config.BlobStorageS3EndpointInsecure,
			CiCacheBucketName:          cdWorkflowConfig.CdCacheBucket,
			CiCacheRegion:              cdWorkflowConfig.CdCacheRegion,
			CiCacheBucketVersioning:    impl.config.BlobStorageS3BucketVersioned,
			CiArtifactBucketName:       cdStageWorkflowRequest.CiArtifactBucket,
			CiArtifactRegion:           cdWorkflowConfig.CdCacheRegion,
			CiArtifactBucketVersioning: impl.config.BlobStorageS3BucketVersioned,
			CiLogBucketName:            impl.config.GetDefaultBuildLogsBucket(),
			CiLogRegion:                impl.config.GetDefaultCdLogsBucketRegion(),
			CiLogBucketVersioning:      impl.config.BlobStorageS3BucketVersioned,
		}
	case types.BLOB_STORAGE_GCP:
		cdStageWorkflowRequest.GcpBlobConfig = &blob_storage.GcpBlobConfig{
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
			ArtifactBucketName:     impl.config.GetDefaultBuildLogsBucket(),
			LogBucketName:          impl.config.GetDefaultBuildLogsBucket(),
		}
		cdStageWorkflowRequest.ArtifactLocation = impl.buildDefaultArtifactLocation(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.CiArtifactFileName = cdStageWorkflowRequest.ArtifactLocation
	case types.BLOB_STORAGE_AZURE:
		cdStageWorkflowRequest.AzureBlobConfig = &blob_storage.AzureBlobConfig{
			Enabled:               true,
			AccountName:           impl.config.AzureAccountName,
			BlobContainerCiCache:  impl.config.AzureBlobContainerCiCache,
			AccountKey:            impl.config.AzureAccountKey,
			BlobContainerCiLog:    impl.config.AzureBlobContainerCiLog,
			BlobContainerArtifact: impl.config.AzureBlobContainerCiLog,
		}
		cdStageWorkflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			EndpointUrl:     impl.config.AzureGatewayUrl,
			IsInSecure:      impl.config.AzureGatewayConnectionInsecure,
			CiLogBucketName: impl.config.AzureBlobContainerCiLog,
			CiLogRegion:     "",
			AccessKey:       impl.config.AzureAccountName,
		}
		cdStageWorkflowRequest.ArtifactLocation = impl.buildDefaultArtifactLocation(cdWorkflowConfig, cdWf, runner)
		cdStageWorkflowRequest.CiArtifactFileName = cdStageWorkflowRequest.ArtifactLocation
	default:
		if impl.config.BlobStorageEnabled {
			return nil, fmt.Errorf("blob storage %s not supported", cdStageWorkflowRequest.CloudProvider)
		}
	}
	cdStageWorkflowRequest.DefaultAddressPoolBaseCidr = impl.config.GetDefaultAddressPoolBaseCidr()
	cdStageWorkflowRequest.DefaultAddressPoolSize = impl.config.GetDefaultAddressPoolSize()
	if util.IsManifestDownload(cdPipeline.DeploymentAppType) || util.IsManifestPush(cdPipeline.DeploymentAppType) {
		cdStageWorkflowRequest.IsDryRun = true
	}
	return cdStageWorkflowRequest, nil
}

func (impl *TriggerServiceImpl) GetArtifactVulnerabilityStatus(artifact *repository.CiArtifact, cdPipeline *pipelineConfig.Pipeline, ctx context.Context) (bool, error) {
	isVulnerable := false
	if len(artifact.ImageDigest) > 0 {
		var cveStores []*security.CveStore
		_, span := otel.Tracer("orchestrator").Start(ctx, "scanResultRepository.FindByImageDigest")
		imageScanResult, err := impl.scanResultRepository.FindByImageDigest(artifact.ImageDigest)
		span.End()
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error fetching image digest", "digest", artifact.ImageDigest, "err", err)
			return false, err
		}
		for _, item := range imageScanResult {
			cveStores = append(cveStores, &item.CveStore)
		}
		_, span = otel.Tracer("orchestrator").Start(ctx, "cvePolicyRepository.GetBlockedCVEList")
		if cdPipeline.Environment.ClusterId == 0 {
			envDetails, err := impl.envRepository.FindById(cdPipeline.EnvironmentId)
			if err != nil {
				impl.logger.Errorw("error fetching cluster details by env, GetArtifactVulnerabilityStatus", "envId", cdPipeline.EnvironmentId, "err", err)
				return false, err
			}
			cdPipeline.Environment = *envDetails
		}
		blockCveList, err := impl.cvePolicyRepository.GetBlockedCVEList(cveStores, cdPipeline.Environment.ClusterId, cdPipeline.EnvironmentId, cdPipeline.AppId, false)
		span.End()
		if err != nil {
			impl.logger.Errorw("error while fetching env", "err", err)
			return false, err
		}
		if len(blockCveList) > 0 {
			isVulnerable = true
		}
	}
	return isVulnerable, nil
}

func (impl *TriggerServiceImpl) ReserveImagesGeneratedAtPlugin(customTagId int, registryImageMap map[string][]string) ([]int, error) {
	var imagePathReservationIds []int
	for _, images := range registryImageMap {
		for _, image := range images {
			imagePathReservationData, err := impl.customTagService.ReserveImagePath(image, customTagId)
			if err != nil {
				impl.logger.Errorw("Error in marking custom tag reserved", "err", err)
				return imagePathReservationIds, err
			}
			if imagePathReservationData != nil {
				imagePathReservationIds = append(imagePathReservationIds, imagePathReservationData.Id)
			}
		}
	}
	return imagePathReservationIds, nil
}

func setExtraEnvVariableInDeployStep(deploySteps []*bean3.StepObject, extraEnvVariables map[string]string, webhookAndCiData *gitSensorClient.WebhookAndCiData) {
	for _, deployStep := range deploySteps {
		for variableKey, variableValue := range extraEnvVariables {
			if isExtraVariableDynamic(variableKey, webhookAndCiData) && deployStep.StepType == "INLINE" {
				extraInputVar := &bean3.VariableObject{
					Name:                  variableKey,
					Format:                "STRING",
					Value:                 variableValue,
					VariableType:          bean3.VARIABLE_TYPE_REF_GLOBAL,
					ReferenceVariableName: variableKey,
				}
				deployStep.InputVars = append(deployStep.InputVars, extraInputVar)
			}
		}
	}
}

func (impl *TriggerServiceImpl) getDeployStageDetails(pipelineId int) (pipelineConfig.CdWorkflowRunner, string, int, error) {
	deployStageWfr := pipelineConfig.CdWorkflowRunner{}
	//getting deployment pipeline latest wfr by pipelineId
	deployStageWfr, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipelineId, bean2.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in getting latest status of deploy type wfr by pipelineId", "err", err, "pipelineId", pipelineId)
		return deployStageWfr, "", 0, err
	}
	deployStageTriggeredByUserEmail, err := impl.userService.GetEmailById(deployStageWfr.TriggeredBy)
	if err != nil {
		impl.logger.Errorw("error in getting user email by id", "err", err, "userId", deployStageWfr.TriggeredBy)
		return deployStageWfr, "", 0, err
	}
	pipelineReleaseCounter, err := impl.pipelineOverrideRepository.GetCurrentPipelineReleaseCounter(pipelineId)
	if err != nil {
		impl.logger.Errorw("error occurred while fetching latest release counter for pipeline", "pipelineId", pipelineId, "err", err)
		return deployStageWfr, "", 0, err
	}
	return deployStageWfr, deployStageTriggeredByUserEmail, pipelineReleaseCounter, nil
}

func (impl *TriggerServiceImpl) buildArtifactLocationForS3(cdWorkflowConfig *pipelineConfig.CdWorkflowConfig, cdWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner) (string, string, string) {
	cdArtifactLocationFormat := cdWorkflowConfig.CdArtifactLocationFormat
	if cdArtifactLocationFormat == "" {
		cdArtifactLocationFormat = impl.config.GetArtifactLocationFormat()
	}
	if cdWorkflowConfig.LogsBucket == "" {
		cdWorkflowConfig.LogsBucket = impl.config.GetDefaultBuildLogsBucket()
	}
	ArtifactLocation := fmt.Sprintf("s3://%s/"+impl.config.GetDefaultArtifactKeyPrefix()+"/"+cdArtifactLocationFormat, cdWorkflowConfig.LogsBucket, cdWf.Id, runner.Id)
	artifactFileName := fmt.Sprintf(impl.config.GetDefaultArtifactKeyPrefix()+"/"+cdArtifactLocationFormat, cdWf.Id, runner.Id)
	return ArtifactLocation, cdWorkflowConfig.LogsBucket, artifactFileName
}

func (impl *TriggerServiceImpl) buildDefaultArtifactLocation(cdWorkflowConfig *pipelineConfig.CdWorkflowConfig, savedWf *pipelineConfig.CdWorkflow, runner *pipelineConfig.CdWorkflowRunner) string {
	cdArtifactLocationFormat := cdWorkflowConfig.CdArtifactLocationFormat
	if cdArtifactLocationFormat == "" {
		cdArtifactLocationFormat = impl.config.GetArtifactLocationFormat()
	}
	ArtifactLocation := fmt.Sprintf("%s/"+cdArtifactLocationFormat, impl.config.GetDefaultArtifactKeyPrefix(), savedWf.Id, runner.Id)
	return ArtifactLocation
}

func ReplaceImageTagWithDigest(image, digest string) string {
	imageWithoutTag := strings.Split(image, ":")[0]
	imageWithDigest := fmt.Sprintf("%s@%s", imageWithoutTag, digest)
	return imageWithDigest
}

func (impl *TriggerServiceImpl) sendPreStageNotification(ctx context.Context, cdWf *pipelineConfig.CdWorkflow, pipeline *pipelineConfig.Pipeline) error {
	wfr, err := impl.cdWorkflowRepository.FindByWorkflowIdAndRunnerType(ctx, cdWf.Id, bean2.CD_WORKFLOW_TYPE_PRE)
	if err != nil {
		return err
	}

	event := impl.eventFactory.Build(util2.Trigger, &pipeline.Id, pipeline.AppId, &pipeline.EnvironmentId, util2.CD)
	impl.logger.Debugw("event PreStageTrigger", "event", event)
	event = impl.eventFactory.BuildExtraCDData(event, &wfr, 0, bean2.CD_WORKFLOW_TYPE_PRE)
	_, span := otel.Tracer("orchestrator").Start(ctx, "eventClient.WriteNotificationEvent")
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	span.End()
	if evtErr != nil {
		impl.logger.Errorw("CD trigger event not sent", "error", evtErr)
	}
	return nil
}

func isExtraVariableDynamic(variableName string, webhookAndCiData *gitSensorClient.WebhookAndCiData) bool {
	if strings.Contains(variableName, GIT_COMMIT_HASH_PREFIX) || strings.Contains(variableName, GIT_SOURCE_TYPE_PREFIX) || strings.Contains(variableName, GIT_SOURCE_VALUE_PREFIX) ||
		strings.Contains(variableName, APP_LABEL_VALUE_PREFIX) || strings.Contains(variableName, APP_LABEL_KEY_PREFIX) ||
		strings.Contains(variableName, CHILD_CD_ENV_NAME_PREFIX) || strings.Contains(variableName, CHILD_CD_CLUSTER_NAME_PREFIX) ||
		strings.Contains(variableName, CHILD_CD_COUNT) || strings.Contains(variableName, APP_LABEL_COUNT) || strings.Contains(variableName, GIT_SOURCE_COUNT) ||
		webhookAndCiData != nil {

		return true
	}
	return false
}
func convert(ts string) (*time.Time, error) {
	t, err := time.Parse(bean4.LayoutRFC3339, ts)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
