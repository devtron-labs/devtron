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
	"fmt"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/pipeline/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/user"
	"path/filepath"
	"strconv"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
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
}

func NewCiServiceImpl(Logger *zap.SugaredLogger, workflowService WorkflowService,
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository, ciConfig *CiConfig, eventClient client.EventClient,
	eventFactory client.EventFactory, mergeUtil *util.MergeUtil, ciPipelineRepository pipelineConfig.CiPipelineRepository,
	prePostCiScriptHistoryService history.PrePostCiScriptHistoryService,
	pipelineStageService PipelineStageService,
	userService user.UserService) *CiServiceImpl {
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
	}
}

const WorkflowStarting = "Starting"
const WorkflowInProgress = "Progressing"
const WorkflowAborted = "Aborted"
const WorkflowFailed = "Failed"

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

	createdWf, err := impl.executeCiPipeline(workflowRequest)
	if err != nil {
		impl.Logger.Errorw("workflow error", "err", err)
		return 0, err
	}
	impl.Logger.Debugw("ci triggered", "wf name ", createdWf.Name, " pipeline ", trigger.PipelineId)
	middleware.CiTriggerCounter.WithLabelValues(strconv.Itoa(pipeline.AppId), strconv.Itoa(trigger.PipelineId)).Inc()
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
	_, evtErr := impl.eventClient.WriteEvent(event)
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
		Name:         pipeline.Name + "-" + strconv.Itoa(pipeline.Id),
		Status:       WorkflowStarting,
		Message:      "",
		StartedOn:    time.Now(),
		CiPipelineId: pipeline.Id,
		Namespace:    wfConfig.Namespace,
		GitTriggers:  gitTriggers,
		LogLocation:  "",
		TriggeredBy:  userId,
	}
	err := impl.ciWorkflowRepository.SaveWorkFlow(ciWorkflow)
	if err != nil {
		impl.Logger.Errorw("saving workflow error", "err", err)
		return &pipelineConfig.CiWorkflow{}, err
	}
	impl.Logger.Debugw("workflow saved ", "id", ciWorkflow.Id)
	return ciWorkflow, nil
}

func (impl *CiServiceImpl) executeCiPipeline(workflowRequest *WorkflowRequest) (*v1alpha1.Workflow, error) {
	createdWorkFlow, err := impl.workflowService.SubmitWorkflow(workflowRequest)
	if err != nil {
		impl.Logger.Errorw("workflow error", "err", err)
		return nil, err
	}
	return createdWorkFlow, nil
}
func (impl *CiServiceImpl) buildArtifactLocation(ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, savedWf *pipelineConfig.CiWorkflow) string {
	if ciWorkflowConfig.LogsBucket == "" {
		ciWorkflowConfig.LogsBucket = impl.ciConfig.DefaultBuildLogsBucket
	}
	ciArtifactLocationFormat := ciWorkflowConfig.CiArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = impl.ciConfig.CiArtifactLocationFormat
	}
	ArtifactLocation := fmt.Sprintf("s3://%s/"+impl.ciConfig.DefaultArtifactKeyPrefix+"/"+ciArtifactLocationFormat, ciWorkflowConfig.LogsBucket, savedWf.Id, savedWf.Id)
	return ArtifactLocation
}

func (impl *CiServiceImpl) buildArtifactLocationAzure(ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, savedWf *pipelineConfig.CiWorkflow) string {
	ciArtifactLocationFormat := ciWorkflowConfig.CiArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = impl.ciConfig.CiArtifactLocationFormat
	}
	ArtifactLocation := fmt.Sprintf(ciArtifactLocationFormat, savedWf.Id, savedWf.Id)
	return ArtifactLocation
}

func (impl *CiServiceImpl) buildWfRequestForCiPipeline(pipeline *pipelineConfig.CiPipeline, trigger Trigger,
	ciMaterials []*pipelineConfig.CiPipelineMaterial, savedWf *pipelineConfig.CiWorkflow,
	ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, ciPipelineScripts []*pipelineConfig.CiPipelineScript) (*WorkflowRequest, error) {
	var ciProjectDetails []CiProjectDetails
	commitHashes := trigger.CommitHashes
	for _, ciMaterial := range ciMaterials {
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
			CommitTime:      commitHashForPipelineId.Date,
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

	args := pipeline.CiTemplate.Args
	ciLevelArgs := pipeline.DockerArgs

	if ciLevelArgs == "" {
		ciLevelArgs = "{}"
	}

	merged, err := impl.mergeUtil.JsonPatch([]byte(args), []byte(ciLevelArgs))
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return nil, err
	}

	checkoutPath := pipeline.CiTemplate.GitMaterial.CheckoutPath
	if checkoutPath == "" {
		checkoutPath = "./"
	}
	user, err := impl.userService.GetById(trigger.TriggeredBy)
	if err != nil {
		impl.Logger.Errorw("unable to find user by id", "err", err, "id", trigger.TriggeredBy)
		return nil, err
	}
	dockerfilePath := filepath.Join(pipeline.CiTemplate.GitMaterial.CheckoutPath, pipeline.CiTemplate.DockerfilePath)
	workflowRequest := &WorkflowRequest{
		WorkflowNamePrefix:         strconv.Itoa(savedWf.Id) + "-" + savedWf.Name,
		PipelineName:               pipeline.Name,
		PipelineId:                 pipeline.Id,
		DockerRegistryId:           pipeline.CiTemplate.DockerRegistry.Id,
		DockerRegistryType:         string(pipeline.CiTemplate.DockerRegistry.RegistryType),
		DockerImageTag:             dockerImageTag,
		DockerRegistryURL:          pipeline.CiTemplate.DockerRegistry.RegistryURL,
		DockerRepository:           pipeline.CiTemplate.DockerRepository,
		DockerBuildArgs:            string(merged),
		DockerBuildTargetPlatform:  pipeline.CiTemplate.TargetPlatform,
		DockerFileLocation:         dockerfilePath,
		DockerUsername:             pipeline.CiTemplate.DockerRegistry.Username,
		DockerPassword:             pipeline.CiTemplate.DockerRegistry.Password,
		AwsRegion:                  pipeline.CiTemplate.DockerRegistry.AWSRegion,
		AccessKey:                  pipeline.CiTemplate.DockerRegistry.AWSAccessKeyId,
		SecretKey:                  pipeline.CiTemplate.DockerRegistry.AWSSecretAccessKey,
		DockerConnection:           pipeline.CiTemplate.DockerRegistry.Connection,
		DockerCert:                 pipeline.CiTemplate.DockerRegistry.Cert,
		CiCacheFileName:            pipeline.Name + "-" + strconv.Itoa(pipeline.Id) + ".tar.gz",
		CiProjectDetails:           ciProjectDetails,
		Namespace:                  ciWorkflowConfig.Namespace,
		CiImage:                    ciWorkflowConfig.CiImage,
		ActiveDeadlineSeconds:      ciWorkflowConfig.CiTimeout,
		WorkflowId:                 savedWf.Id,
		TriggeredBy:                savedWf.TriggeredBy,
		CacheLimit:                 impl.ciConfig.CacheLimit,
		InvalidateCache:            trigger.InvalidateCache,
		ScanEnabled:                pipeline.ScanEnabled,
		CloudProvider:              impl.ciConfig.CloudProvider,
		DefaultAddressPoolBaseCidr: impl.ciConfig.DefaultAddressPoolBaseCidr,
		DefaultAddressPoolSize:     impl.ciConfig.DefaultAddressPoolSize,
		PreCiSteps:                 preCiSteps,
		PostCiSteps:                postCiSteps,
		RefPlugins:                 refPluginsData,
		AppName:                    pipeline.App.AppName,
		TriggerByAuthor:            user.EmailId,
	}

	switch workflowRequest.CloudProvider {
	case BLOB_STORAGE_S3:
		//No AccessKey is used for uploading artifacts, instead IAM based auth is used
		workflowRequest.CiCacheRegion = ciWorkflowConfig.CiCacheRegion
		workflowRequest.CiCacheLocation = ciWorkflowConfig.CiCacheBucket
		workflowRequest.CiArtifactLocation = impl.buildArtifactLocation(ciWorkflowConfig, savedWf)
	case BLOB_STORAGE_AZURE:
		workflowRequest.AzureBlobConfig = &AzureBlobConfig{
			Enabled:              impl.ciConfig.CloudProvider == BLOB_STORAGE_AZURE,
			AccountName:          impl.ciConfig.AzureAccountName,
			BlobContainerCiCache: impl.ciConfig.AzureBlobContainerCiCache,
			AccountKey:           impl.ciConfig.AzureAccountKey,
			BlobContainerCiLog:   impl.ciConfig.AzureBlobContainerCiLog,
		}
		workflowRequest.CiArtifactLocation = impl.buildArtifactLocationAzure(ciWorkflowConfig, savedWf)
	case BLOB_STORAGE_MINIO:
		//For MINIO type blob storage, AccessKey & SecretAccessKey are injected through EnvVar
		workflowRequest.CiCacheLocation = ciWorkflowConfig.CiCacheBucket
		workflowRequest.CiArtifactLocation = impl.buildArtifactLocation(ciWorkflowConfig, savedWf)
		workflowRequest.MinioEndpoint = impl.ciConfig.MinioEndpoint
	default:
		return nil, fmt.Errorf("cloudprovider %s not supported", workflowRequest.CloudProvider)
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
	return dockerImageTag
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
