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
	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	util2 "github.com/devtron-labs/devtron/util/event"
	"go.uber.org/zap"
	"path/filepath"
	"strconv"
	"time"
)

type CiService interface {
	TriggerCiPipeline(trigger Trigger) (int, error)
	GetCiMaterials(pipelineId int, ciMaterials []*pipelineConfig.CiPipelineMaterial) ([]*pipelineConfig.CiPipelineMaterial, error)
}

type CiServiceImpl struct {
	Logger                       *zap.SugaredLogger
	workflowService              WorkflowService
	ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository
	ciWorkflowRepository         pipelineConfig.CiWorkflowRepository
	ciConfig                     *CiConfig
	eventClient                  client.EventClient
	eventFactory                 client.EventFactory
	mergeUtil                    *util.MergeUtil
	ciPipelineRepository         pipelineConfig.CiPipelineRepository
}

func NewCiServiceImpl(Logger *zap.SugaredLogger, workflowService WorkflowService, ciPipelineMaterialRepository pipelineConfig.CiPipelineMaterialRepository,
	ciWorkflowRepository pipelineConfig.CiWorkflowRepository, ciConfig *CiConfig, eventClient client.EventClient, eventFactory client.EventFactory, mergeUtil *util.MergeUtil, ciPipelineRepository pipelineConfig.CiPipelineRepository) *CiServiceImpl {
	return &CiServiceImpl{
		Logger:                       Logger,
		workflowService:              workflowService,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		ciWorkflowRepository:         ciWorkflowRepository,
		ciConfig:                     ciConfig,
		eventClient:                  eventClient,
		eventFactory:                 eventFactory,
		mergeUtil:                    mergeUtil,
		ciPipelineRepository:         ciPipelineRepository,
	}
}

const WorkflowStarting = "Starting"
const WorkflowAborted = "Aborted"
const WorkflowFailed = "Failed"

func (impl *CiServiceImpl) GetCiMaterials(pipelineId int, ciMaterials []*pipelineConfig.CiPipelineMaterial) ([]*pipelineConfig.CiPipelineMaterial, error) {
	if !(ciMaterials == nil || len(ciMaterials) == 0) {
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
	gitT := make(map[int]pipelineConfig.GitCommit)
	for k, v := range trigger.CommitHashes {
		gitT[k] = pipelineConfig.GitCommit{
			Commit:  v.Commit,
			Author:  v.Author,
			Changes: v.Changes,
			Message: v.Message,
			Date:    v.Date,
		}
	}
	material.GitTriggers = gitT
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
		gitTriggers[k] = pipelineConfig.GitCommit{
			Commit:  v.Commit,
			Author:  v.Author,
			Date:    v.Date,
			Message: v.Message,
			Changes: v.Changes,
		}
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

func (impl *CiServiceImpl) buildWfRequestForCiPipeline(pipeline *pipelineConfig.CiPipeline, trigger Trigger,
	ciMaterials []*pipelineConfig.CiPipelineMaterial, savedWf *pipelineConfig.CiWorkflow,
	ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, ciPipelineScripts []*pipelineConfig.CiPipelineScript) (*WorkflowRequest, error) {
	var ciProjectDetails []CiProjectDetails
	commitHashes := trigger.CommitHashes
	for _, ciMaterial := range ciMaterials {
		ciProjectDetail := CiProjectDetails{
			GitRepository: ciMaterial.GitMaterial.Url,
			MaterialName:  ciMaterial.GitMaterial.Name,
			CheckoutPath:  ciMaterial.GitMaterial.CheckoutPath,
			CommitHash:    commitHashes[ciMaterial.Id].Commit,
			Author:        commitHashes[ciMaterial.Id].Author,
			SourceType:    ciMaterial.Type,
			SourceValue:   ciMaterial.Value,
			GitTag:        ciMaterial.GitTag,
			Message:       commitHashes[ciMaterial.Id].Message,
			Type:          string(ciMaterial.Type),
			CommitTime:    commitHashes[ciMaterial.Id].Date,
			GitOptions: GitOptions{
				UserName:    ciMaterial.GitMaterial.GitProvider.UserName,
				Password:    ciMaterial.GitMaterial.GitProvider.Password,
				SSHKey:      ciMaterial.GitMaterial.GitProvider.SshKey,
				AccessToken: ciMaterial.GitMaterial.GitProvider.AccessToken,
				AuthMode:    ciMaterial.GitMaterial.GitProvider.AuthMode,
			},
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

	dockerImageTag := impl.buildImageTag(commitHashes, pipeline.Id, savedWf.Id)
	if ciWorkflowConfig.CiCacheBucket == "" {
		ciWorkflowConfig.CiCacheBucket = impl.ciConfig.DefaultCacheBucket
	}

	if ciWorkflowConfig.LogsBucket == "" {
		ciWorkflowConfig.LogsBucket = impl.ciConfig.DefaultBuildLogsBucket
	}

	ciArtifactLocationFormat := ciWorkflowConfig.CiArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = impl.ciConfig.CiArtifactLocationFormat
	}
	ciArtifactLocation := fmt.Sprintf("s3://%s/"+impl.ciConfig.DefaultArtifactKeyPrefix+"/"+ciArtifactLocationFormat, ciWorkflowConfig.LogsBucket, savedWf.Id, savedWf.Id)

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
	if len(checkoutPath) == 0 {
		checkoutPath = "./"
	}
	dockerfilePath := filepath.Join(pipeline.CiTemplate.GitMaterial.CheckoutPath, pipeline.CiTemplate.DockerfilePath)
	workflowRequest := &WorkflowRequest{
		WorkflowNamePrefix:       strconv.Itoa(savedWf.Id) + "-" + savedWf.Name,
		PipelineName:             pipeline.Name,
		PipelineId:               pipeline.Id,
		DockerRegistryType:       string(pipeline.CiTemplate.DockerRegistry.RegistryType),
		DockerImageTag:           dockerImageTag,
		DockerRegistryURL:        pipeline.CiTemplate.DockerRegistry.RegistryURL,
		DockerRepository:         pipeline.CiTemplate.DockerRepository,
		DockerBuildArgs:          string(merged),
		DockerFileLocation:       dockerfilePath,
		DockerUsername:           pipeline.CiTemplate.DockerRegistry.Username,
		DockerPassword:           pipeline.CiTemplate.DockerRegistry.Password,
		AwsRegion:                pipeline.CiTemplate.DockerRegistry.AWSRegion,
		AccessKey:                pipeline.CiTemplate.DockerRegistry.AWSAccessKeyId,
		SecretKey:                pipeline.CiTemplate.DockerRegistry.AWSSecretAccessKey,
		CiCacheFileName:          pipeline.Name + "-" + strconv.Itoa(pipeline.Id) + ".tar.gz",
		CiProjectDetails:         ciProjectDetails,
		Namespace:                ciWorkflowConfig.Namespace,
		CiImage:                  ciWorkflowConfig.CiImage,
		ActiveDeadlineSeconds:    ciWorkflowConfig.CiTimeout,
		WorkflowId:               savedWf.Id,
		TriggeredBy:              savedWf.TriggeredBy,
		CacheLimit:               impl.ciConfig.CacheLimit,
		BeforeDockerBuildScripts: beforeDockerBuildScripts,
		AfterDockerBuildScripts:  afterDockerBuildScripts,
		CiArtifactLocation:       ciArtifactLocation,
		InvalidateCache:          trigger.InvalidateCache,
		ScanEnabled:              pipeline.ScanEnabled,
		CloudProvider:            impl.ciConfig.CloudProvider,
	}

	switch workflowRequest.CloudProvider {
	case CLOUD_PROVIDER_AWS:
		workflowRequest.CiCacheRegion = ciWorkflowConfig.CiCacheRegion
		workflowRequest.CiCacheLocation = ciWorkflowConfig.CiCacheBucket
	case CLOUD_PROVIDER_AZURE:
		workflowRequest.AzureBlobConfig = &AzureBlobConfig{
			Enabled:       true,
			AccountName:   impl.ciConfig.AzureAccountName,
			BlobContainer: impl.ciConfig.AzureBlobContainer,
			AccountKey:    impl.ciConfig.AzureAccountKey,
		}
	default:
		return nil, fmt.Errorf("cloudprovider %s not supported", workflowRequest.CloudProvider)
	}
	return workflowRequest, nil
}

func (impl *CiServiceImpl) buildImageTag(commitHashes map[int]bean.GitCommit, id int, wfId int) string {
	dockerImageTag := ""
	for _, v := range commitHashes {
		if v.Commit == "" {
			continue
		}
		if dockerImageTag == "" {
			dockerImageTag = v.Commit[:8]
		} else {
			dockerImageTag = dockerImageTag + "-" + v.Commit[:8]
		}
	}
	if dockerImageTag != "" {
		dockerImageTag = dockerImageTag + "-" + strconv.Itoa(id) + "-" + strconv.Itoa(wfId)
	}
	return dockerImageTag
}
