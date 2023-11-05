/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package pipeline

import (
	"encoding/json"
	"errors"
	"fmt"
	repository5 "github.com/devtron-labs/devtron/internal/sql/repository"
	appRepository "github.com/devtron-labs/devtron/internal/sql/repository/app"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/app"
	repository1 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/plugin"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/variables"
	repository4 "github.com/devtron-labs/devtron/pkg/variables/repository"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/common-lib/blob-storage"
	client "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	util2 "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
)

type CiService interface {
	TriggerCiPipeline(trigger types.Trigger) (int, error)
	GetCiMaterials(pipelineId int, ciMaterials []*pipelineConfig.CiPipelineMaterial) ([]*pipelineConfig.CiPipelineMaterial, error)
}

type CiServiceImpl struct {
	Logger                         *zap.SugaredLogger
	workflowService                WorkflowService
	ciPipelineMaterialRepository   pipelineConfig.CiPipelineMaterialRepository
	ciWorkflowRepository           pipelineConfig.CiWorkflowRepository
	ciConfig                       *types.CiConfig
	eventClient                    client.EventClient
	eventFactory                   client.EventFactory
	mergeUtil                      *util.MergeUtil
	ciPipelineRepository           pipelineConfig.CiPipelineRepository
	prePostCiScriptHistoryService  history.PrePostCiScriptHistoryService
	pipelineStageService           PipelineStageService
	userService                    user.UserService
	ciTemplateService              CiTemplateService
	appCrudOperationService        app.AppCrudOperationService
	envRepository                  repository1.EnvironmentRepository
	appRepository                  appRepository.AppRepository
	customTagService               CustomTagService
	variableSnapshotHistoryService variables.VariableSnapshotHistoryService
	config                         *types.CiConfig
	pluginInputVariableParser      PluginInputVariableParser
	globalPluginService            plugin.GlobalPluginService
	scopedVariableManager          variables.ScopedVariableManager
}

func NewCiServiceImpl(Logger *zap.SugaredLogger, workflowService WorkflowService,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository, eventClient client.EventClient,
	eventFactory client.EventFactory, mergeUtil *util.MergeUtil, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	prePostCiScriptHistoryService history.PrePostCiScriptHistoryService,
	pipelineStageService PipelineStageService,
	userService user.UserService,
	ciTemplateService CiTemplateService, appCrudOperationService app.AppCrudOperationService, envRepository repository1.EnvironmentRepository, appRepository appRepository.AppRepository,
	scopedVariableManager variables.ScopedVariableManager,
	customTagService CustomTagService,
	pluginInputVariableParser PluginInputVariableParser,
	globalPluginService plugin.GlobalPluginService,
) *CiServiceImpl {
	cis := &CiServiceImpl{
		Logger:                        Logger,
		workflowService:               workflowService,
		ciPipelineMaterialRepository:  ciPipelineMaterialRepository,
		ciWorkflowRepository:          ciWorkflowRepository,
		eventClient:                   eventClient,
		eventFactory:                  eventFactory,
		mergeUtil:                     mergeUtil,
		ciPipelineRepository:          ciPipelineRepository,
		prePostCiScriptHistoryService: prePostCiScriptHistoryService,
		pipelineStageService:          pipelineStageService,
		userService:                   userService,
		ciTemplateService:             ciTemplateService,
		appCrudOperationService:       appCrudOperationService,
		envRepository:                 envRepository,
		appRepository:                 appRepository,
		scopedVariableManager:         scopedVariableManager,
		customTagService:              customTagService,
		pluginInputVariableParser:     pluginInputVariableParser,
		globalPluginService:           globalPluginService,
	}
	config, err := types.GetCiConfig()
	if err != nil {
		return nil
	}
	cis.config = config
	return cis
}

func (impl *CiServiceImpl) GetCiMaterials(pipelineId int, ciMaterials []*pipelineConfig.CiPipelineMaterial) ([]*pipelineConfig.CiPipelineMaterial, error) {
	if !(len(ciMaterials) == 0) {
		return ciMaterials, nil
	} else {
		ciMaterials, err := impl.ciPipelineMaterialRepository.GetByPipelineId(pipelineId)
		if err != nil {
			impl.Logger.Errorw("err", "err", err)
			return nil, err
		}
		impl.Logger.Debug("ciMaterials for pipeline trigger ", ciMaterials)
		return ciMaterials, nil
	}
}

func (impl *CiServiceImpl) TriggerCiPipeline(trigger types.Trigger) (int, error) {
	impl.Logger.Debug("ci pipeline manual trigger")
	ciMaterials, err := impl.GetCiMaterials(trigger.PipelineId, trigger.CiMaterials)
	if err != nil {
		return 0, err
	}
	if trigger.PipelineType == bean2.CI_JOB && len(ciMaterials) != 0 {
		ciMaterials = []*pipelineConfig.CiPipelineMaterial{ciMaterials[0]}
		ciMaterials[0].GitMaterial = nil
		ciMaterials[0].GitMaterialId = 0
	}
	ciPipelineScripts, err := impl.ciPipelineRepository.FindCiScriptsByCiPipelineId(trigger.PipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		return 0, err
	}

	var pipeline *pipelineConfig.CiPipeline
	for _, m := range ciMaterials {
		pipeline = m.CiPipeline
		break
	}

	ciWorkflowConfig, err := impl.ciWorkflowRepository.FindConfigByPipelineId(trigger.PipelineId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("could not fetch ci config", "pipeline", trigger.PipelineId)
		return 0, err
	}

	scope := resourceQualifiers.Scope{
		AppId: pipeline.App.Id,
	}
	env, isJob, err := impl.getEnvironmentForJob(pipeline, trigger)
	if err != nil {
		return 0, err
	}
	if isJob && env != nil {
		ciWorkflowConfig.Namespace = env.Namespace

		//This will be populated for jobs running in selected environment
		scope.EnvId = env.Id
		scope.ClusterId = env.ClusterId

		scope.SystemMetadata = &resourceQualifiers.SystemMetadata{
			EnvironmentName: env.Name,
			ClusterName:     env.Cluster.ClusterName,
			Namespace:       env.Namespace,
		}
	}
	if ciWorkflowConfig.Namespace == "" {
		ciWorkflowConfig.Namespace = impl.config.GetDefaultNamespace()
	}

	//preCiSteps, postCiSteps, refPluginsData, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(pipeline.Id, ciEvent)
	prePostAndRefPluginResponse, err := impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(pipeline.Id, bean2.CiStage, scope)
	if err != nil {
		impl.Logger.Errorw("error in getting pre steps data for wf request", "err", err, "ciPipelineId", pipeline.Id)
		return 0, err
	}
	preCiSteps := prePostAndRefPluginResponse.PreStageSteps
	postCiSteps := prePostAndRefPluginResponse.PostStageSteps
	refPluginsData := prePostAndRefPluginResponse.RefPluginData
	variableSnapshot := prePostAndRefPluginResponse.VariableSnapshot

	if len(preCiSteps) == 0 && isJob {
		return 0, &util.ApiError{
			UserMessage: "No tasks are configured in this job pipeline",
		}
	}
	savedCiWf, err := impl.saveNewWorkflow(pipeline, ciWorkflowConfig, trigger.CommitHashes, trigger.TriggeredBy, trigger.EnvironmentId, isJob, trigger.ReferenceCiWorkflowId)
	if err != nil {
		impl.Logger.Errorw("could not save new workflow", "err", err)
		return 0, err
	}

	workflowRequest, err := impl.buildWfRequestForCiPipeline(pipeline, trigger, ciMaterials, savedCiWf, ciWorkflowConfig, ciPipelineScripts, preCiSteps, postCiSteps, refPluginsData)
	if err != nil {
		impl.Logger.Errorw("make workflow req", "err", err)
		return 0, err
	}
	workflowRequest.Scope = scope

	if impl.config != nil && impl.config.BuildxK8sDriverOptions != "" {
		err = impl.setBuildxK8sDriverData(workflowRequest)
		if err != nil {
			impl.Logger.Errorw("error in setBuildxK8sDriverData", "BUILDX_K8S_DRIVER_OPTIONS", impl.config.BuildxK8sDriverOptions, "err", err)
			return 0, err
		}
	}

	//savedCiWf.LogLocation = impl.ciCdConfig.CiDefaultBuildLogsKeyPrefix + "/" + workflowRequest.WorkflowNamePrefix + "/main.log"
	savedCiWf.LogLocation = fmt.Sprintf("%s/%s/main.log", impl.config.GetDefaultBuildLogsKeyPrefix(), workflowRequest.WorkflowNamePrefix)
	err = impl.updateCiWorkflow(workflowRequest, savedCiWf)

	appLabels, err := impl.appCrudOperationService.GetLabelsByAppId(pipeline.AppId)
	if err != nil {
		return 0, err
	}
	workflowRequest.AppId = pipeline.AppId
	workflowRequest.AppLabels = appLabels
	workflowRequest.Env = env
	if isJob {
		workflowRequest.Type = bean2.JOB_WORKFLOW_PIPELINE_TYPE
	} else {
		workflowRequest.Type = bean2.CI_WORKFLOW_PIPELINE_TYPE
	}
	err = impl.executeCiPipeline(workflowRequest)
	if err != nil {
		impl.Logger.Errorw("workflow error", "err", err)
		return 0, err
	}
	impl.Logger.Debugw("ci triggered", " pipeline ", trigger.PipelineId)

	var variableSnapshotHistories = util3.GetBeansPtr(
		repository4.GetSnapshotBean(savedCiWf.Id, repository4.HistoryReferenceTypeCIWORKFLOW, variableSnapshot))
	if len(variableSnapshotHistories) > 0 {
		err = impl.scopedVariableManager.SaveVariableHistoriesForTrigger(variableSnapshotHistories, trigger.TriggeredBy)
		if err != nil {
			impl.Logger.Errorf("Not able to save variable snapshot for CI trigger %s", err)
		}
	}

	middleware.CiTriggerCounter.WithLabelValues(pipeline.App.AppName, pipeline.Name).Inc()
	go impl.WriteCITriggerEvent(trigger, pipeline, workflowRequest)
	return savedCiWf.Id, err
}

func (impl *CiServiceImpl) setBuildxK8sDriverData(workflowRequest *types.WorkflowRequest) error {
	ciBuildConfig := workflowRequest.CiBuildConfig
	if ciBuildConfig != nil {
		if dockerBuildConfig := ciBuildConfig.DockerBuildConfig; dockerBuildConfig != nil {
			buildxK8sDriverOptions := make([]map[string]string, 0)
			err := json.Unmarshal([]byte(impl.config.BuildxK8sDriverOptions), &buildxK8sDriverOptions)
			if err != nil {
				errMsg := "error in parsing BUILDX_K8S_DRIVER_OPTIONS from the devtron-cm, "
				err = errors.New(errMsg + "error : " + err.Error())
				impl.Logger.Errorw(errMsg, "err", err)
				return err
			} else {
				dockerBuildConfig.BuildxK8sDriverOptions = buildxK8sDriverOptions
			}
		}
	}
	return nil
}

func (impl *CiServiceImpl) getEnvironmentForJob(pipeline *pipelineConfig.CiPipeline, trigger types.Trigger) (*repository1.Environment, bool, error) {
	app, err := impl.appRepository.FindById(pipeline.AppId)
	if err != nil {
		impl.Logger.Errorw("could not find app", "err", err)
		return nil, false, err
	}

	var env *repository1.Environment
	isJob := false
	if app.AppType == helper.Job {
		isJob = true
		if trigger.EnvironmentId != 0 {
			env, err = impl.envRepository.FindById(trigger.EnvironmentId)
			if err != nil {
				impl.Logger.Errorw("could not find environment", "err", err)
				return nil, isJob, err
			}
			return env, isJob, nil
		}
	}
	return nil, isJob, nil
}

func (impl *CiServiceImpl) WriteCITriggerEvent(trigger types.Trigger, pipeline *pipelineConfig.CiPipeline, workflowRequest *types.WorkflowRequest) {
	event := impl.eventFactory.Build(util2.Trigger, &pipeline.Id, pipeline.AppId, nil, util2.CI)
	material := &client.MaterialTriggerInfo{}

	material.GitTriggers = trigger.CommitHashes

	event.UserId = int(trigger.TriggeredBy)
	event.CiWorkflowRunnerId = workflowRequest.WorkflowId
	event = impl.eventFactory.BuildExtraCIData(event, material, workflowRequest.CiImage)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.Logger.Errorw("error in writing event", "err", evtErr)
	}
}

// TODO: Send all trigger data
func (impl *CiServiceImpl) BuildPayload(trigger types.Trigger, pipeline *pipelineConfig.CiPipeline) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = pipeline.App.AppName
	payload.PipelineName = pipeline.Name
	return payload
}

func (impl *CiServiceImpl) saveNewWorkflow(pipeline *pipelineConfig.CiPipeline, wfConfig *pipelineConfig.CiWorkflowConfig,
	commitHashes map[int]pipelineConfig.GitCommit, userId int32, EnvironmentId int, isJob bool, refCiWorkflowId int) (wf *pipelineConfig.CiWorkflow, error error) {

	ciWorkflow := &pipelineConfig.CiWorkflow{
		Name:                  pipeline.Name + "-" + strconv.Itoa(pipeline.Id),
		Status:                pipelineConfig.WorkflowStarting,
		Message:               "",
		StartedOn:             time.Now(),
		CiPipelineId:          pipeline.Id,
		Namespace:             impl.config.GetDefaultNamespace(),
		BlobStorageEnabled:    impl.config.BlobStorageEnabled,
		GitTriggers:           commitHashes,
		LogLocation:           "",
		TriggeredBy:           userId,
		ReferenceCiWorkflowId: refCiWorkflowId,
	}
	if isJob {
		ciWorkflow.Namespace = wfConfig.Namespace
		ciWorkflow.EnvironmentId = EnvironmentId
	}
	err := impl.ciWorkflowRepository.SaveWorkFlow(ciWorkflow)
	if err != nil {
		impl.Logger.Errorw("saving workflow error", "err", err)
		return &pipelineConfig.CiWorkflow{}, err
	}
	impl.Logger.Debugw("workflow saved ", "id", ciWorkflow.Id)
	return ciWorkflow, nil
}

func (impl *CiServiceImpl) executeCiPipeline(workflowRequest *types.WorkflowRequest) error {
	_, err := impl.workflowService.SubmitWorkflow(workflowRequest)
	if err != nil {
		impl.Logger.Errorw("workflow error", "err", err)
		return err
	}
	return nil
}

func (impl *CiServiceImpl) buildS3ArtifactLocation(ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, savedWf *pipelineConfig.CiWorkflow) (string, string, string) {
	ciArtifactLocationFormat := ciWorkflowConfig.CiArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = impl.config.GetArtifactLocationFormat()
	}
	ArtifactLocation := fmt.Sprintf("s3://%s/%s/"+ciArtifactLocationFormat, ciWorkflowConfig.LogsBucket, impl.config.GetDefaultArtifactKeyPrefix(), savedWf.Id, savedWf.Id)
	artifactFileName := fmt.Sprintf(impl.config.GetDefaultArtifactKeyPrefix()+"/"+ciArtifactLocationFormat, savedWf.Id, savedWf.Id)
	return ArtifactLocation, ciWorkflowConfig.LogsBucket, artifactFileName
}

func (impl *CiServiceImpl) buildDefaultArtifactLocation(ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, savedWf *pipelineConfig.CiWorkflow) string {
	ciArtifactLocationFormat := ciWorkflowConfig.CiArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = impl.config.GetArtifactLocationFormat()
	}
	ArtifactLocation := fmt.Sprintf("%s/"+ciArtifactLocationFormat, impl.config.GetDefaultArtifactKeyPrefix(), savedWf.Id, savedWf.Id)
	return ArtifactLocation
}

func (impl *CiServiceImpl) buildWfRequestForCiPipeline(pipeline *pipelineConfig.CiPipeline, trigger types.Trigger,
	ciMaterials []*pipelineConfig.CiPipelineMaterial, savedWf *pipelineConfig.CiWorkflow,
	ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, ciPipelineScripts []*pipelineConfig.CiPipelineScript,
	preCiSteps []*bean2.StepObject, postCiSteps []*bean2.StepObject, refPluginsData []*bean2.RefPluginObject) (*types.WorkflowRequest, error) {
	var ciProjectDetails []bean2.CiProjectDetails
	commitHashes := trigger.CommitHashes
	for _, ciMaterial := range ciMaterials {
		// ignore those materials which have inactive git material
		if ciMaterial == nil || ciMaterial.GitMaterial == nil || !ciMaterial.GitMaterial.Active {
			continue
		}
		commitHashForPipelineId := commitHashes[ciMaterial.Id]
		ciProjectDetail := bean2.CiProjectDetails{
			GitRepository:   ciMaterial.GitMaterial.Url,
			MaterialName:    ciMaterial.GitMaterial.Name,
			CheckoutPath:    ciMaterial.GitMaterial.CheckoutPath,
			FetchSubmodules: ciMaterial.GitMaterial.FetchSubmodules,
			CommitHash:      commitHashForPipelineId.Commit,
			Author:          commitHashForPipelineId.Author,
			SourceType:      ciMaterial.Type,
			SourceValue:     ciMaterial.Value,
			GitTag:          ciMaterial.GitTag,
			Message:         commitHashForPipelineId.Message,
			Type:            string(ciMaterial.Type),
			CommitTime:      commitHashForPipelineId.Date.Format(bean.LayoutRFC3339),
			GitOptions: bean2.GitOptions{
				UserName:      ciMaterial.GitMaterial.GitProvider.UserName,
				Password:      ciMaterial.GitMaterial.GitProvider.Password,
				SshPrivateKey: ciMaterial.GitMaterial.GitProvider.SshPrivateKey,
				AccessToken:   ciMaterial.GitMaterial.GitProvider.AccessToken,
				AuthMode:      ciMaterial.GitMaterial.GitProvider.AuthMode,
			},
		}

		if ciMaterial.Type == pipelineConfig.SOURCE_TYPE_WEBHOOK {
			webhookData := commitHashForPipelineId.WebhookData
			ciProjectDetail.WebhookData = pipelineConfig.WebhookData{
				Id:              webhookData.Id,
				EventActionType: webhookData.EventActionType,
				Data:            webhookData.Data,
			}
		}

		ciProjectDetails = append(ciProjectDetails, ciProjectDetail)
	}

	var beforeDockerBuildScripts []*bean.CiScript
	var afterDockerBuildScripts []*bean.CiScript
	for _, ciPipelineScript := range ciPipelineScripts {
		ciTask := &bean.CiScript{
			Id:             ciPipelineScript.Id,
			Index:          ciPipelineScript.Index,
			Name:           ciPipelineScript.Name,
			Script:         ciPipelineScript.Script,
			OutputLocation: ciPipelineScript.OutputLocation,
		}

		if ciPipelineScript.Stage == BEFORE_DOCKER_BUILD {
			beforeDockerBuildScripts = append(beforeDockerBuildScripts, ciTask)
		} else if ciPipelineScript.Stage == AFTER_DOCKER_BUILD {
			afterDockerBuildScripts = append(afterDockerBuildScripts, ciTask)
		}
	}

	if !(len(beforeDockerBuildScripts) == 0 && len(afterDockerBuildScripts) == 0) {
		//found beforeDockerBuildScripts/afterDockerBuildScripts
		//building preCiSteps & postCiSteps from them, refPluginsData not needed
		preCiSteps = buildCiStepsDataFromDockerBuildScripts(beforeDockerBuildScripts)
		postCiSteps = buildCiStepsDataFromDockerBuildScripts(afterDockerBuildScripts)
		refPluginsData = []*bean2.RefPluginObject{}
	}

	var dockerImageTag string
	customTag, err := impl.customTagService.GetActiveCustomTagByEntityKeyAndValue(bean2.EntityTypeCiPipelineId, strconv.Itoa(pipeline.Id))
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if customTag.Id != 0 && customTag.Enabled == true {
		imagePathReservation, err := impl.customTagService.GenerateImagePath(bean2.EntityTypeCiPipelineId, strconv.Itoa(pipeline.Id), pipeline.CiTemplate.DockerRegistry.RegistryURL, pipeline.CiTemplate.DockerRepository)
		if err != nil {
			if errors.Is(err, bean2.ErrImagePathInUse) {
				savedWf.Status = pipelineConfig.WorkflowFailed
				savedWf.Message = bean2.ImageTagUnavailableMessage
				err1 := impl.ciWorkflowRepository.UpdateWorkFlow(savedWf)
				if err1 != nil {
					impl.Logger.Errorw("could not save workflow, after failing due to conflicting image tag")
				}
				return nil, err
			}
			return nil, err
		}
		savedWf.ImagePathReservationIds = []int{imagePathReservation.Id}
		//imagePath = docker.io/avd0/dashboard:fd23414b
		dockerImageTag = strings.Split(imagePathReservation.ImagePath, ":")[1]
	} else {
		dockerImageTag = impl.buildImageTag(commitHashes, pipeline.Id, savedWf.Id)
	}

	// skopeo plugin specific logic
	registryDestinationImageMap, registryCredentialMap, pluginArtifactStage, imageReservationIds, err := impl.GetWorkflowRequestVariablesForSkopeoPlugin(
		preCiSteps, postCiSteps, dockerImageTag, customTag.Id,
		fmt.Sprintf(bean2.ImagePathPattern, pipeline.CiTemplate.DockerRegistry.RegistryURL, pipeline.CiTemplate.DockerRepository, dockerImageTag), pipeline.CiTemplate.DockerRegistry.Id)

	if err != nil {
		impl.Logger.Errorw("error in getting env variables for skopeo plugin")
		return nil, err
	}

	savedWf.ImagePathReservationIds = append(savedWf.ImagePathReservationIds, imageReservationIds...)
	// skopeo plugin logic ends

	if ciWorkflowConfig.CiCacheBucket == "" {
		ciWorkflowConfig.CiCacheBucket = impl.config.DefaultCacheBucket
	}

	if ciWorkflowConfig.CiCacheRegion == "" {
		ciWorkflowConfig.CiCacheRegion = impl.config.DefaultCacheBucketRegion
	}

	if ciWorkflowConfig.CiImage == "" {
		ciWorkflowConfig.CiImage = impl.config.GetDefaultImage()
	}
	if ciWorkflowConfig.CiTimeout == 0 {
		ciWorkflowConfig.CiTimeout = impl.config.GetDefaultTimeout()
	}

	ciTemplate := pipeline.CiTemplate
	ciLevelArgs := pipeline.DockerArgs

	if ciLevelArgs == "" {
		ciLevelArgs = "{}"
	}

	if pipeline.CiTemplate.DockerBuildOptions == "" {
		pipeline.CiTemplate.DockerBuildOptions = "{}"
	}
	user, err := impl.userService.GetById(trigger.TriggeredBy)
	if err != nil {
		impl.Logger.Errorw("unable to find user by id", "err", err, "id", trigger.TriggeredBy)
		return nil, err
	}
	var dockerfilePath string
	var dockerRepository string
	var checkoutPath string
	var ciBuildConfigBean *bean2.CiBuildConfigBean
	dockerRegistry := &repository3.DockerArtifactStore{}
	if !pipeline.IsExternal && pipeline.IsDockerConfigOverridden {
		templateOverrideBean, err := impl.ciTemplateService.FindTemplateOverrideByCiPipelineId(pipeline.Id)
		if err != nil {
			return nil, err
		}
		ciBuildConfigBean = templateOverrideBean.CiBuildConfig
		templateOverride := templateOverrideBean.CiTemplateOverride
		checkoutPath = templateOverride.GitMaterial.CheckoutPath
		dockerfilePath = templateOverride.DockerfilePath
		dockerRepository = templateOverride.DockerRepository
		dockerRegistry = templateOverride.DockerRegistry
	} else {
		checkoutPath = ciTemplate.GitMaterial.CheckoutPath
		dockerfilePath = ciTemplate.DockerfilePath
		dockerRegistry = ciTemplate.DockerRegistry
		dockerRepository = ciTemplate.DockerRepository
		ciBuildConfigEntity := ciTemplate.CiBuildConfig
		ciBuildConfigBean, err = bean2.ConvertDbBuildConfigToBean(ciBuildConfigEntity)
		if ciBuildConfigBean != nil {
			ciBuildConfigBean.BuildContextGitMaterialId = ciTemplate.BuildContextGitMaterialId
		}
		if err != nil {
			impl.Logger.Errorw("error occurred while converting buildconfig dbEntity to configBean", "ciBuildConfigEntity", ciBuildConfigEntity, "err", err)
			return nil, errors.New("error while parsing ci build config")
		}
	}
	if checkoutPath == "" {
		checkoutPath = "./"
	}
	//mergedArgs := string(merged)
	oldArgs := ciTemplate.Args
	ciBuildConfigBean, err = bean2.OverrideCiBuildConfig(dockerfilePath, oldArgs, ciLevelArgs, ciTemplate.DockerBuildOptions, ciTemplate.TargetPlatform, ciBuildConfigBean)
	if err != nil {
		impl.Logger.Errorw("error occurred while overriding ci build config", "oldArgs", oldArgs, "ciLevelArgs", ciLevelArgs, "error", err)
		return nil, errors.New("error while parsing ci build config")
	}
	buildContextCheckoutPath, err := impl.ciPipelineMaterialRepository.GetCheckoutPath(ciBuildConfigBean.BuildContextGitMaterialId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error occurred while getting checkout path from git material", "gitMaterialId", ciBuildConfigBean.BuildContextGitMaterialId, "error", err)
		return nil, err
	}
	if buildContextCheckoutPath == "" {
		buildContextCheckoutPath = checkoutPath
	}
	if ciBuildConfigBean.UseRootBuildContext {
		//use root build context i.e '.'
		buildContextCheckoutPath = "."
	}

	ciBuildConfigBean.PipelineType = trigger.PipelineType

	if ciBuildConfigBean.CiBuildType == bean2.SELF_DOCKERFILE_BUILD_TYPE || ciBuildConfigBean.CiBuildType == bean2.MANAGED_DOCKERFILE_BUILD_TYPE {
		ciBuildConfigBean.DockerBuildConfig.BuildContext = filepath.Join(buildContextCheckoutPath, ciBuildConfigBean.DockerBuildConfig.BuildContext)
		dockerBuildConfig := ciBuildConfigBean.DockerBuildConfig
		dockerfilePath = filepath.Join(checkoutPath, dockerBuildConfig.DockerfilePath)
		dockerBuildConfig.DockerfilePath = dockerfilePath
		checkoutPath = dockerfilePath[:strings.LastIndex(dockerfilePath, "/")+1]
	} else if ciBuildConfigBean.CiBuildType == bean2.BUILDPACK_BUILD_TYPE {
		buildPackConfig := ciBuildConfigBean.BuildPackConfig
		checkoutPath = filepath.Join(checkoutPath, buildPackConfig.ProjectPath)
	}

	defaultTargetPlatform := impl.config.DefaultTargetPlatform
	useBuildx := impl.config.UseBuildx

	if ciBuildConfigBean.DockerBuildConfig != nil {
		if ciBuildConfigBean.DockerBuildConfig.TargetPlatform == "" && useBuildx {
			ciBuildConfigBean.DockerBuildConfig.TargetPlatform = defaultTargetPlatform
			ciBuildConfigBean.DockerBuildConfig.UseBuildx = useBuildx
		}
		ciBuildConfigBean.DockerBuildConfig.BuildxProvenanceMode = impl.config.BuildxProvenanceMode
	}

	workflowRequest := &types.WorkflowRequest{
		WorkflowNamePrefix:          strconv.Itoa(savedWf.Id) + "-" + savedWf.Name,
		PipelineName:                pipeline.Name,
		PipelineId:                  pipeline.Id,
		CiCacheFileName:             pipeline.Name + "-" + strconv.Itoa(pipeline.Id) + ".tar.gz",
		CiProjectDetails:            ciProjectDetails,
		Namespace:                   ciWorkflowConfig.Namespace,
		BlobStorageConfigured:       savedWf.BlobStorageEnabled,
		CiImage:                     ciWorkflowConfig.CiImage,
		ActiveDeadlineSeconds:       ciWorkflowConfig.CiTimeout,
		WorkflowId:                  savedWf.Id,
		TriggeredBy:                 savedWf.TriggeredBy,
		CacheLimit:                  impl.config.CacheLimit,
		ScanEnabled:                 pipeline.ScanEnabled,
		CloudProvider:               impl.config.CloudProvider,
		DefaultAddressPoolBaseCidr:  impl.config.GetDefaultAddressPoolBaseCidr(),
		DefaultAddressPoolSize:      impl.config.GetDefaultAddressPoolSize(),
		PreCiSteps:                  preCiSteps,
		PostCiSteps:                 postCiSteps,
		RefPlugins:                  refPluginsData,
		AppName:                     pipeline.App.AppName,
		TriggerByAuthor:             user.EmailId,
		CiBuildConfig:               ciBuildConfigBean,
		CiBuildDockerMtuValue:       impl.config.CiRunnerDockerMTUValue,
		IgnoreDockerCachePush:       impl.config.IgnoreDockerCacheForCI,
		IgnoreDockerCachePull:       impl.config.IgnoreDockerCacheForCI,
		CacheInvalidate:             trigger.InvalidateCache,
		ExtraEnvironmentVariables:   trigger.ExtraEnvironmentVariables,
		EnableBuildContext:          impl.config.EnableBuildContext,
		OrchestratorHost:            impl.config.OrchestratorHost,
		OrchestratorToken:           impl.config.OrchestratorToken,
		ImageRetryCount:             impl.config.ImageRetryCount,
		ImageRetryInterval:          impl.config.ImageRetryInterval,
		WorkflowExecutor:            impl.config.GetWorkflowExecutorType(),
		Type:                        bean2.CI_WORKFLOW_PIPELINE_TYPE,
		CiArtifactLastFetch:         trigger.CiArtifactLastFetch,
		RegistryDestinationImageMap: registryDestinationImageMap,
		RegistryCredentialMap:       registryCredentialMap,
		PluginArtifactStage:         pluginArtifactStage,
	}

	if dockerRegistry != nil {

		workflowRequest.DockerRegistryId = dockerRegistry.Id
		workflowRequest.DockerRegistryType = string(dockerRegistry.RegistryType)
		workflowRequest.DockerImageTag = dockerImageTag
		workflowRequest.DockerRegistryURL = dockerRegistry.RegistryURL
		workflowRequest.DockerRepository = dockerRepository
		workflowRequest.CheckoutPath = checkoutPath
		workflowRequest.DockerUsername = dockerRegistry.Username
		workflowRequest.DockerPassword = dockerRegistry.Password
		workflowRequest.AwsRegion = dockerRegistry.AWSRegion
		workflowRequest.AccessKey = dockerRegistry.AWSAccessKeyId
		workflowRequest.SecretKey = dockerRegistry.AWSSecretAccessKey
		workflowRequest.DockerConnection = dockerRegistry.Connection
		workflowRequest.DockerCert = dockerRegistry.Cert

	}
	if ciWorkflowConfig.LogsBucket == "" {
		ciWorkflowConfig.LogsBucket = impl.config.GetDefaultBuildLogsBucket()
	}
	if len(registryDestinationImageMap) > 0 {
		workflowRequest.PushImageBeforePostCI = true
	}
	switch workflowRequest.CloudProvider {
	case types.BLOB_STORAGE_S3:
		//No AccessKey is used for uploading artifacts, instead IAM based auth is used
		workflowRequest.CiCacheRegion = ciWorkflowConfig.CiCacheRegion
		workflowRequest.CiCacheLocation = ciWorkflowConfig.CiCacheBucket
		workflowRequest.CiArtifactLocation, workflowRequest.CiArtifactBucket, workflowRequest.CiArtifactFileName = impl.buildS3ArtifactLocation(ciWorkflowConfig, savedWf)
		workflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			AccessKey:                  impl.config.BlobStorageS3AccessKey,
			Passkey:                    impl.config.BlobStorageS3SecretKey,
			EndpointUrl:                impl.config.BlobStorageS3Endpoint,
			IsInSecure:                 impl.config.BlobStorageS3EndpointInsecure,
			CiCacheBucketName:          ciWorkflowConfig.CiCacheBucket,
			CiCacheRegion:              ciWorkflowConfig.CiCacheRegion,
			CiCacheBucketVersioning:    impl.config.BlobStorageS3BucketVersioned,
			CiArtifactBucketName:       workflowRequest.CiArtifactBucket,
			CiArtifactRegion:           impl.config.GetDefaultCdLogsBucketRegion(),
			CiArtifactBucketVersioning: impl.config.BlobStorageS3BucketVersioned,
			CiLogBucketName:            impl.config.GetDefaultBuildLogsBucket(),
			CiLogRegion:                impl.config.GetDefaultCdLogsBucketRegion(),
			CiLogBucketVersioning:      impl.config.BlobStorageS3BucketVersioned,
		}
	case types.BLOB_STORAGE_GCP:
		workflowRequest.GcpBlobConfig = &blob_storage.GcpBlobConfig{
			CredentialFileJsonData: impl.config.BlobStorageGcpCredentialJson,
			CacheBucketName:        ciWorkflowConfig.CiCacheBucket,
			LogBucketName:          ciWorkflowConfig.LogsBucket,
			ArtifactBucketName:     ciWorkflowConfig.LogsBucket,
		}
		workflowRequest.CiArtifactLocation = impl.buildDefaultArtifactLocation(ciWorkflowConfig, savedWf)
		workflowRequest.CiArtifactFileName = workflowRequest.CiArtifactLocation
	case types.BLOB_STORAGE_AZURE:
		workflowRequest.AzureBlobConfig = &blob_storage.AzureBlobConfig{
			Enabled:               impl.config.CloudProvider == types.BLOB_STORAGE_AZURE,
			AccountName:           impl.config.AzureAccountName,
			BlobContainerCiCache:  impl.config.AzureBlobContainerCiCache,
			AccountKey:            impl.config.AzureAccountKey,
			BlobContainerCiLog:    impl.config.AzureBlobContainerCiLog,
			BlobContainerArtifact: impl.config.AzureBlobContainerCiLog,
		}
		workflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			EndpointUrl:           impl.config.AzureGatewayUrl,
			IsInSecure:            impl.config.AzureGatewayConnectionInsecure,
			CiLogBucketName:       impl.config.AzureBlobContainerCiLog,
			CiLogRegion:           impl.config.DefaultCacheBucketRegion,
			CiLogBucketVersioning: impl.config.BlobStorageS3BucketVersioned,
			AccessKey:             impl.config.AzureAccountName,
		}
		workflowRequest.CiArtifactLocation = impl.buildDefaultArtifactLocation(ciWorkflowConfig, savedWf)
		workflowRequest.CiArtifactFileName = workflowRequest.CiArtifactLocation
	default:
		if impl.config.BlobStorageEnabled {
			return nil, fmt.Errorf("blob storage %s not supported", workflowRequest.CloudProvider)
		}
	}
	return workflowRequest, nil
}

func (impl *CiServiceImpl) GetWorkflowRequestVariablesForSkopeoPlugin(preCiSteps []*bean2.StepObject, postCiSteps []*bean2.StepObject, customTag string, customTagId int, buildImagePath string, buildImagedockerRegistryId string) (map[string][]string, map[string]plugin.RegistryCredentials, string, []int, error) {
	var registryDestinationImageMap map[string][]string
	var registryCredentialMap map[string]plugin.RegistryCredentials
	var pluginArtifactStage string
	var imagePathReservationIds []int
	skopeoRefPluginId, err := impl.globalPluginService.GetRefPluginIdByRefPluginName(SKOPEO)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in getting skopeo plugin id", "err", err)
		return registryDestinationImageMap, registryCredentialMap, pluginArtifactStage, imagePathReservationIds, err
	}
	for _, step := range preCiSteps {
		if skopeoRefPluginId != 0 && step.RefPluginId == skopeoRefPluginId {
			// for Skopeo plugin parse destination images and save its data in image path reservation table
			registryDestinationImageMap, registryCredentialMap, err = impl.pluginInputVariableParser.HandleSkopeoPluginInputVariable(step.InputVars, customTag, buildImagePath, buildImagedockerRegistryId)
			if err != nil {
				impl.Logger.Errorw("error in parsing skopeo input variable", "err", err)
				return registryDestinationImageMap, registryCredentialMap, pluginArtifactStage, imagePathReservationIds, err
			}
			pluginArtifactStage = repository5.PRE_CI
		}
	}
	for _, step := range postCiSteps {
		if skopeoRefPluginId != 0 && step.RefPluginId == skopeoRefPluginId {
			// for Skopeo plugin parse destination images and save its data in image path reservation table
			registryDestinationImageMap, registryCredentialMap, err = impl.pluginInputVariableParser.HandleSkopeoPluginInputVariable(step.InputVars, customTag, buildImagePath, buildImagedockerRegistryId)
			if err != nil {
				impl.Logger.Errorw("error in parsing skopeo input variable", "err", err)
				return registryDestinationImageMap, registryCredentialMap, pluginArtifactStage, imagePathReservationIds, err
			}
			pluginArtifactStage = repository5.POST_CI
		}
	}
	imagePathReservationIds, err = impl.ReserveImagesGeneratedAtPlugin(customTagId, registryDestinationImageMap)
	if err != nil {
		return nil, nil, pluginArtifactStage, imagePathReservationIds, nil
	}
	return registryDestinationImageMap, registryCredentialMap, pluginArtifactStage, imagePathReservationIds, nil
}

func (impl *CiServiceImpl) ReserveImagesGeneratedAtPlugin(customTagId int, registryImageMap map[string][]string) ([]int, error) {
	var imagePathReservationIds []int
	for _, images := range registryImageMap {
		for _, image := range images {
			imagePathReservationData, err := impl.customTagService.ReserveImagePath(image, customTagId)
			if err != nil {
				impl.Logger.Errorw("Error in marking custom tag reserved", "err", err)
				return imagePathReservationIds, err
			}
			imagePathReservationIds = append(imagePathReservationIds, imagePathReservationData.Id)
		}
	}
	return imagePathReservationIds, nil
}

func buildCiStepsDataFromDockerBuildScripts(dockerBuildScripts []*bean.CiScript) []*bean2.StepObject {
	//before plugin support, few variables were set as env vars in ci-runner
	//these variables are now moved to global vars in plugin steps, but to avoid error in old scripts adding those variables in payload
	inputVars := []*bean2.VariableObject{
		{
			Name:                  "DOCKER_IMAGE_TAG",
			Format:                "STRING",
			VariableType:          bean2.VARIABLE_TYPE_REF_GLOBAL,
			ReferenceVariableName: "DOCKER_IMAGE_TAG",
		},
		{
			Name:                  "DOCKER_REPOSITORY",
			Format:                "STRING",
			VariableType:          bean2.VARIABLE_TYPE_REF_GLOBAL,
			ReferenceVariableName: "DOCKER_REPOSITORY",
		},
		{
			Name:                  "DOCKER_REGISTRY_URL",
			Format:                "STRING",
			VariableType:          bean2.VARIABLE_TYPE_REF_GLOBAL,
			ReferenceVariableName: "DOCKER_REGISTRY_URL",
		},
		{
			Name:                  "DOCKER_IMAGE",
			Format:                "STRING",
			VariableType:          bean2.VARIABLE_TYPE_REF_GLOBAL,
			ReferenceVariableName: "DOCKER_IMAGE",
		},
	}
	var ciSteps []*bean2.StepObject
	for _, dockerBuildScript := range dockerBuildScripts {
		ciStep := &bean2.StepObject{
			Name:          dockerBuildScript.Name,
			Index:         dockerBuildScript.Index,
			Script:        dockerBuildScript.Script,
			ArtifactPaths: []string{dockerBuildScript.OutputLocation},
			StepType:      string(repository.PIPELINE_STEP_TYPE_INLINE),
			ExecutorType:  string(repository2.SCRIPT_TYPE_SHELL),
			InputVars:     inputVars,
		}
		ciSteps = append(ciSteps, ciStep)
	}
	return ciSteps
}

func (impl *CiServiceImpl) buildImageTag(commitHashes map[int]pipelineConfig.GitCommit, id int, wfId int) string {
	dockerImageTag := ""
	for _, v := range commitHashes {
		_truncatedCommit := ""
		if v.WebhookData.Id == 0 {
			if v.Commit == "" {
				continue
			}
			_truncatedCommit = _getTruncatedImageTag(v.Commit)
		} else {
			_targetCheckout := v.WebhookData.Data[bean.WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME]
			if _targetCheckout == "" {
				continue
			}
			_truncatedCommit = _getTruncatedImageTag(_targetCheckout)
			if v.WebhookData.EventActionType == bean.WEBHOOK_EVENT_MERGED_ACTION_TYPE {
				_sourceCheckout := v.WebhookData.Data[bean.WEBHOOK_SELECTOR_SOURCE_CHECKOUT_NAME]
				if len(_sourceCheckout) > 0 {
					_truncatedCommit = _truncatedCommit + "-" + _getTruncatedImageTag(_sourceCheckout)
				}
			}
		}

		if dockerImageTag == "" {
			dockerImageTag = _truncatedCommit
		} else {
			dockerImageTag = dockerImageTag + "-" + _truncatedCommit
		}
	}
	if dockerImageTag != "" {
		dockerImageTag = dockerImageTag + "-" + strconv.Itoa(id) + "-" + strconv.Itoa(wfId)
	}

	// replace / with underscore, as docker image tag doesn't support slash. it gives error
	dockerImageTag = strings.ReplaceAll(dockerImageTag, "/", "_")

	return dockerImageTag
}

func (impl *CiServiceImpl) updateCiWorkflow(request *types.WorkflowRequest, savedWf *pipelineConfig.CiWorkflow) error {
	ciBuildConfig := request.CiBuildConfig
	ciBuildType := string(ciBuildConfig.CiBuildType)
	savedWf.CiBuildType = ciBuildType
	return impl.ciWorkflowRepository.UpdateWorkFlow(savedWf)
}

func _getTruncatedImageTag(imageTag string) string {
	_length := len(imageTag)
	if _length == 0 {
		return imageTag
	}

	_truncatedLength := 8

	if _length < _truncatedLength {
		return imageTag
	} else {
		return imageTag[:_truncatedLength]
	}
}
