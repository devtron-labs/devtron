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
	"errors"
	"fmt"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
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
	TriggerCiPipeline(trigger Trigger) (int, error)
	GetCiMaterials(pipelineId int, ciMaterials []*pipelineConfig.CiPipelineMaterial) ([]*pipelineConfig.CiPipelineMaterial, error)
}

type CiServiceImpl struct {
	Logger                        *zap.SugaredLogger
	workflowService               WorkflowService
	ciPipelineMaterialRepository  pipelineConfig.CiPipelineMaterialRepository
	ciWorkflowRepository          pipelineConfig.CiWorkflowRepository
	ciConfig                      *CiConfig
	eventClient                   client.EventClient
	eventFactory                  client.EventFactory
	mergeUtil                     *util.MergeUtil
	ciPipelineRepository          pipelineConfig.CiPipelineRepository
	prePostCiScriptHistoryService history.PrePostCiScriptHistoryService
	pipelineStageService          PipelineStageService
	userService                   user.UserService
	ciTemplateService             CiTemplateService
	appCrudOperationService       app.AppCrudOperationService
}

func NewCiServiceImpl(Logger *zap.SugaredLogger, workflowService WorkflowService,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository, ciConfig *CiConfig, eventClient client.EventClient,
	eventFactory client.EventFactory, mergeUtil *util.MergeUtil, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	prePostCiScriptHistoryService history.PrePostCiScriptHistoryService,
	pipelineStageService PipelineStageService,
	userService user.UserService,
	ciTemplateService CiTemplateService, appCrudOperationService app.AppCrudOperationService) *CiServiceImpl {
	return &CiServiceImpl{
		Logger:                        Logger,
		workflowService:               workflowService,
		ciPipelineMaterialRepository:  ciPipelineMaterialRepository,
		ciWorkflowRepository:          ciWorkflowRepository,
		ciConfig:                      ciConfig,
		eventClient:                   eventClient,
		eventFactory:                  eventFactory,
		mergeUtil:                     mergeUtil,
		ciPipelineRepository:          ciPipelineRepository,
		prePostCiScriptHistoryService: prePostCiScriptHistoryService,
		pipelineStageService:          pipelineStageService,
		userService:                   userService,
		ciTemplateService:             ciTemplateService,
		appCrudOperationService:       appCrudOperationService,
	}
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

func (impl *CiServiceImpl) TriggerCiPipeline(trigger Trigger) (int, error) {
	impl.Logger.Debug("ci pipeline manual trigger")
	ciMaterials, err := impl.GetCiMaterials(trigger.PipelineId, trigger.CiMaterials)
	if err != nil {
		return 0, err
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
	if ciWorkflowConfig.Namespace == "" {
		ciWorkflowConfig.Namespace = impl.ciConfig.DefaultNamespace
	}
	savedCiWf, err := impl.saveNewWorkflow(pipeline, ciWorkflowConfig, trigger.CommitHashes, trigger.TriggeredBy)
	if err != nil {
		impl.Logger.Errorw("could not save new workflow", "err", err)
		return 0, err
	}

	workflowRequest, err := impl.buildWfRequestForCiPipeline(pipeline, trigger, ciMaterials, savedCiWf, ciWorkflowConfig, ciPipelineScripts)
	if err != nil {
		impl.Logger.Errorw("make workflow req", "err", err)
		return 0, err
	}

	err = impl.updateCiWorkflow(workflowRequest, savedCiWf)

	appLabels, err := impl.appCrudOperationService.GetLabelsByAppId(pipeline.AppId)
	if err != nil {
		return 0, err
	}

	createdWf, err := impl.executeCiPipeline(workflowRequest, appLabels)
	if err != nil {
		impl.Logger.Errorw("workflow error", "err", err)
		return 0, err
	}
	impl.Logger.Debugw("ci triggered", "wf name ", createdWf.Name, " pipeline ", trigger.PipelineId)

	middleware.CiTriggerCounter.WithLabelValues(pipeline.App.AppName, pipeline.Name).Inc()
	go impl.WriteCITriggerEvent(trigger, pipeline, workflowRequest)
	return savedCiWf.Id, err
}

func (impl *CiServiceImpl) WriteCITriggerEvent(trigger Trigger, pipeline *pipelineConfig.CiPipeline, workflowRequest *WorkflowRequest) {
	event := impl.eventFactory.Build(util2.Trigger, &pipeline.Id, pipeline.AppId, nil, util2.CI)
	material := &client.MaterialTriggerInfo{}

	gitTriggers := make(map[int]pipelineConfig.GitCommit)

	for k, v := range trigger.CommitHashes {
		gitCommit := pipelineConfig.GitCommit{
			Commit:  v.Commit,
			Author:  v.Author,
			Changes: v.Changes,
			Message: v.Message,
			Date:    v.Date,
		}

		// set webhook data in gitTriggers
		_webhookData := v.WebhookData
		if _webhookData != nil {
			gitCommit.WebhookData = pipelineConfig.WebhookData{
				Id:              _webhookData.Id,
				EventActionType: _webhookData.EventActionType,
				Data:            _webhookData.Data,
			}
		}

		gitTriggers[k] = gitCommit
	}

	material.GitTriggers = gitTriggers

	event.UserId = int(trigger.TriggeredBy)
	event.CiWorkflowRunnerId = workflowRequest.WorkflowId
	event = impl.eventFactory.BuildExtraCIData(event, material, workflowRequest.CiImage)
	_, evtErr := impl.eventClient.WriteNotificationEvent(event)
	if evtErr != nil {
		impl.Logger.Errorw("error in writing event", "err", evtErr)
	}
}

// TODO: Send all trigger data
func (impl *CiServiceImpl) BuildPayload(trigger Trigger, pipeline *pipelineConfig.CiPipeline, workflowRequest *WorkflowRequest) *client.Payload {
	payload := &client.Payload{}
	payload.AppName = pipeline.App.AppName
	payload.PipelineName = pipeline.Name
	return payload
}

func (impl *CiServiceImpl) saveNewWorkflow(pipeline *pipelineConfig.CiPipeline, wfConfig *pipelineConfig.CiWorkflowConfig,
	commitHashes map[int]bean.GitCommit, userId int32) (wf *pipelineConfig.CiWorkflow, error error) {
	gitTriggers := make(map[int]pipelineConfig.GitCommit)
	for k, v := range commitHashes {
		gitCommit := pipelineConfig.GitCommit{
			Commit:                 v.Commit,
			Author:                 v.Author,
			Date:                   v.Date,
			Message:                v.Message,
			Changes:                v.Changes,
			CiConfigureSourceValue: v.CiConfigureSourceValue,
			CiConfigureSourceType:  v.CiConfigureSourceType,
			GitRepoUrl:             v.GitRepoUrl,
			GitRepoName:            v.GitRepoName,
		}
		webhookData := v.WebhookData
		if webhookData != nil {
			gitCommit.WebhookData = pipelineConfig.WebhookData{
				Id:              webhookData.Id,
				EventActionType: webhookData.EventActionType,
				Data:            webhookData.Data,
			}
		}

		gitTriggers[k] = gitCommit
	}

	ciWorkflow := &pipelineConfig.CiWorkflow{
		Name:               pipeline.Name + "-" + strconv.Itoa(pipeline.Id),
		Status:             pipelineConfig.WorkflowStarting,
		Message:            "",
		StartedOn:          time.Now(),
		CiPipelineId:       pipeline.Id,
		Namespace:          wfConfig.Namespace,
		BlobStorageEnabled: impl.ciConfig.BlobStorageEnabled,
		GitTriggers:        gitTriggers,
		LogLocation:        "",
		TriggeredBy:        userId,
	}
	err := impl.ciWorkflowRepository.SaveWorkFlow(ciWorkflow)
	if err != nil {
		impl.Logger.Errorw("saving workflow error", "err", err)
		return &pipelineConfig.CiWorkflow{}, err
	}
	impl.Logger.Debugw("workflow saved ", "id", ciWorkflow.Id)
	return ciWorkflow, nil
}

func (impl *CiServiceImpl) executeCiPipeline(workflowRequest *WorkflowRequest, appLabels map[string]string) (*v1alpha1.Workflow, error) {
	createdWorkFlow, err := impl.workflowService.SubmitWorkflow(workflowRequest, appLabels)
	if err != nil {
		impl.Logger.Errorw("workflow error", "err", err)
		return nil, err
	}
	return createdWorkFlow, nil
}

func (impl *CiServiceImpl) buildS3ArtifactLocation(ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, savedWf *pipelineConfig.CiWorkflow) (string, string, string) {
	ciArtifactLocationFormat := ciWorkflowConfig.CiArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = impl.ciConfig.CiArtifactLocationFormat
	}
	ArtifactLocation := fmt.Sprintf("s3://%s/%s/"+ciArtifactLocationFormat, ciWorkflowConfig.LogsBucket, impl.ciConfig.DefaultArtifactKeyPrefix, savedWf.Id, savedWf.Id)
	artifactFileName := fmt.Sprintf(impl.ciConfig.DefaultArtifactKeyPrefix+"/"+ciArtifactLocationFormat, savedWf.Id, savedWf.Id)
	return ArtifactLocation, ciWorkflowConfig.LogsBucket, artifactFileName
}

func (impl *CiServiceImpl) buildDefaultArtifactLocation(ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, savedWf *pipelineConfig.CiWorkflow) string {
	ciArtifactLocationFormat := ciWorkflowConfig.CiArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = impl.ciConfig.CiArtifactLocationFormat
	}
	ArtifactLocation := fmt.Sprintf("%s/"+ciArtifactLocationFormat, impl.ciConfig.DefaultArtifactKeyPrefix, savedWf.Id, savedWf.Id)
	return ArtifactLocation
}

func (impl *CiServiceImpl) buildWfRequestForCiPipeline(pipeline *pipelineConfig.CiPipeline, trigger Trigger,
	ciMaterials []*pipelineConfig.CiPipelineMaterial, savedWf *pipelineConfig.CiWorkflow,
	ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, ciPipelineScripts []*pipelineConfig.CiPipelineScript) (*WorkflowRequest, error) {
	var ciProjectDetails []CiProjectDetails
	commitHashes := trigger.CommitHashes
	for _, ciMaterial := range ciMaterials {
		// ignore those materials which have inactive git material
		if ciMaterial == nil || ciMaterial.GitMaterial == nil || !ciMaterial.GitMaterial.Active {
			continue
		}
		commitHashForPipelineId := commitHashes[ciMaterial.Id]
		ciProjectDetail := CiProjectDetails{
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
			GitOptions: GitOptions{
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
	var preCiSteps []*bean2.StepObject
	var postCiSteps []*bean2.StepObject
	var refPluginsData []*bean2.RefPluginObject
	var err error
	if !(len(beforeDockerBuildScripts) == 0 && len(afterDockerBuildScripts) == 0) {
		//found beforeDockerBuildScripts/afterDockerBuildScripts
		//building preCiSteps & postCiSteps from them, refPluginsData not needed
		preCiSteps = buildCiStepsDataFromDockerBuildScripts(beforeDockerBuildScripts)
		postCiSteps = buildCiStepsDataFromDockerBuildScripts(afterDockerBuildScripts)
	} else {
		//beforeDockerBuildScripts & afterDockerBuildScripts not found
		//getting preCiStepsData, postCiStepsData & refPluginsData
		preCiSteps, postCiSteps, refPluginsData, err = impl.pipelineStageService.BuildPrePostAndRefPluginStepsDataForWfRequest(pipeline.Id)
		if err != nil {
			impl.Logger.Errorw("error in getting pre, post & refPlugin steps data for wf request", "err", err, "ciPipelineId", pipeline.Id)
			return nil, err
		}
	}
	dockerImageTag := impl.buildImageTag(commitHashes, pipeline.Id, savedWf.Id)
	if ciWorkflowConfig.CiCacheBucket == "" {
		ciWorkflowConfig.CiCacheBucket = impl.ciConfig.DefaultCacheBucket
	}

	if ciWorkflowConfig.CiCacheRegion == "" {
		ciWorkflowConfig.CiCacheRegion = impl.ciConfig.DefaultCacheBucketRegion
	}
	if ciWorkflowConfig.CiImage == "" {
		ciWorkflowConfig.CiImage = impl.ciConfig.DefaultImage
	}
	if ciWorkflowConfig.CiTimeout == 0 {
		ciWorkflowConfig.CiTimeout = impl.ciConfig.DefaultTimeout
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
	if ciBuildConfigBean.CiBuildType == bean2.SELF_DOCKERFILE_BUILD_TYPE || ciBuildConfigBean.CiBuildType == bean2.MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuildConfig := ciBuildConfigBean.DockerBuildConfig
		dockerfilePath = filepath.Join(checkoutPath, dockerBuildConfig.DockerfilePath)
		dockerBuildConfig.DockerfilePath = dockerfilePath
		checkoutPath = dockerfilePath[:strings.LastIndex(dockerfilePath, "/")+1]
	} else if ciBuildConfigBean.CiBuildType == bean2.BUILDPACK_BUILD_TYPE {
		buildPackConfig := ciBuildConfigBean.BuildPackConfig
		checkoutPath = filepath.Join(checkoutPath, buildPackConfig.ProjectPath)
	}
	workflowRequest := &WorkflowRequest{
		WorkflowNamePrefix:         strconv.Itoa(savedWf.Id) + "-" + savedWf.Name,
		PipelineName:               pipeline.Name,
		PipelineId:                 pipeline.Id,
		CiCacheFileName:            pipeline.Name + "-" + strconv.Itoa(pipeline.Id) + ".tar.gz",
		CiProjectDetails:           ciProjectDetails,
		Namespace:                  ciWorkflowConfig.Namespace,
		BlobStorageConfigured:      savedWf.BlobStorageEnabled,
		CiImage:                    ciWorkflowConfig.CiImage,
		ActiveDeadlineSeconds:      ciWorkflowConfig.CiTimeout,
		WorkflowId:                 savedWf.Id,
		TriggeredBy:                savedWf.TriggeredBy,
		CacheLimit:                 impl.ciConfig.CacheLimit,
		ScanEnabled:                pipeline.ScanEnabled,
		CloudProvider:              impl.ciConfig.CloudProvider,
		DefaultAddressPoolBaseCidr: impl.ciConfig.DefaultAddressPoolBaseCidr,
		DefaultAddressPoolSize:     impl.ciConfig.DefaultAddressPoolSize,
		PreCiSteps:                 preCiSteps,
		PostCiSteps:                postCiSteps,
		RefPlugins:                 refPluginsData,
		AppName:                    pipeline.App.AppName,
		TriggerByAuthor:            user.EmailId,
		CiBuildConfig:              ciBuildConfigBean,
		CiBuildDockerMtuValue:      impl.ciConfig.CiRunnerDockerMTUValue,
		IgnoreDockerCachePush:      impl.ciConfig.IgnoreDockerCacheForCI,
		IgnoreDockerCachePull:      impl.ciConfig.IgnoreDockerCacheForCI,
		CacheInvalidate:            trigger.InvalidateCache,
		ExtraEnvironmentVariables:  trigger.ExtraEnvironmentVariables,
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
		ciWorkflowConfig.LogsBucket = impl.ciConfig.DefaultBuildLogsBucket
	}

	switch workflowRequest.CloudProvider {
	case BLOB_STORAGE_S3:
		//No AccessKey is used for uploading artifacts, instead IAM based auth is used
		workflowRequest.CiCacheRegion = ciWorkflowConfig.CiCacheRegion
		workflowRequest.CiCacheLocation = ciWorkflowConfig.CiCacheBucket
		workflowRequest.CiArtifactLocation, workflowRequest.CiArtifactBucket, workflowRequest.CiArtifactFileName = impl.buildS3ArtifactLocation(ciWorkflowConfig, savedWf)
		workflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			AccessKey:                  impl.ciConfig.BlobStorageS3AccessKey,
			Passkey:                    impl.ciConfig.BlobStorageS3SecretKey,
			EndpointUrl:                impl.ciConfig.BlobStorageS3Endpoint,
			IsInSecure:                 impl.ciConfig.BlobStorageS3EndpointInsecure,
			CiCacheBucketName:          ciWorkflowConfig.CiCacheBucket,
			CiCacheRegion:              ciWorkflowConfig.CiCacheRegion,
			CiCacheBucketVersioning:    impl.ciConfig.BlobStorageS3BucketVersioned,
			CiArtifactBucketName:       workflowRequest.CiArtifactBucket,
			CiArtifactRegion:           impl.ciConfig.DefaultCdLogsBucketRegion,
			CiArtifactBucketVersioning: impl.ciConfig.BlobStorageS3BucketVersioned,
			CiLogBucketName:            impl.ciConfig.DefaultBuildLogsBucket,
			CiLogRegion:                impl.ciConfig.DefaultCdLogsBucketRegion,
			CiLogBucketVersioning:      impl.ciConfig.BlobStorageS3BucketVersioned,
		}
	case BLOB_STORAGE_GCP:
		workflowRequest.GcpBlobConfig = &blob_storage.GcpBlobConfig{
			CredentialFileJsonData: impl.ciConfig.BlobStorageGcpCredentialJson,
			CacheBucketName:        ciWorkflowConfig.CiCacheBucket,
			LogBucketName:          ciWorkflowConfig.LogsBucket,
			ArtifactBucketName:     ciWorkflowConfig.LogsBucket,
		}
		workflowRequest.CiArtifactLocation = impl.buildDefaultArtifactLocation(ciWorkflowConfig, savedWf)
		workflowRequest.CiArtifactFileName = workflowRequest.CiArtifactLocation
	case BLOB_STORAGE_AZURE:
		workflowRequest.AzureBlobConfig = &blob_storage.AzureBlobConfig{
			Enabled:               impl.ciConfig.CloudProvider == BLOB_STORAGE_AZURE,
			AccountName:           impl.ciConfig.AzureAccountName,
			BlobContainerCiCache:  impl.ciConfig.AzureBlobContainerCiCache,
			AccountKey:            impl.ciConfig.AzureAccountKey,
			BlobContainerCiLog:    impl.ciConfig.AzureBlobContainerCiLog,
			BlobContainerArtifact: impl.ciConfig.AzureBlobContainerCiLog,
		}
		workflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			EndpointUrl:           impl.ciConfig.AzureGatewayUrl,
			IsInSecure:            impl.ciConfig.AzureGatewayConnectionInsecure,
			CiLogBucketName:       impl.ciConfig.AzureBlobContainerCiLog,
			CiLogRegion:           impl.ciConfig.DefaultCacheBucketRegion,
			CiLogBucketVersioning: impl.ciConfig.BlobStorageS3BucketVersioned,
			AccessKey:             impl.ciConfig.AzureAccountName,
		}
		workflowRequest.CiArtifactLocation = impl.buildDefaultArtifactLocation(ciWorkflowConfig, savedWf)
		workflowRequest.CiArtifactFileName = workflowRequest.CiArtifactLocation
	default:
		if impl.ciConfig.BlobStorageEnabled {
			return nil, fmt.Errorf("blob storage %s not supported", workflowRequest.CloudProvider)
		}
	}
	return workflowRequest, nil
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

func (impl *CiServiceImpl) buildImageTag(commitHashes map[int]bean.GitCommit, id int, wfId int) string {
	dockerImageTag := ""
	for _, v := range commitHashes {
		_truncatedCommit := ""
		if v.WebhookData == nil {
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

func (impl *CiServiceImpl) updateCiWorkflow(request *WorkflowRequest, savedWf *pipelineConfig.CiWorkflow) error {
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
