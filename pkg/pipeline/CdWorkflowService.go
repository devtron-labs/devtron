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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/workflow/common"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/util/k8s"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	v1alpha12 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/util"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type CdWorkflowService interface {
	SubmitWorkflow(workflowRequest *CdWorkflowRequest, pipeline *pipelineConfig.Pipeline, env *repository.Environment) error
	DeleteWorkflow(wfName string, namespace string) error
	GetWorkflow(name string, namespace string, restConfig *rest.Config, isExtRun bool) (*v1alpha1.Workflow, error)
	ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error)
	UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error)
	TerminateWorkflow(name string, namespace string, restConfig *rest.Config, isExtRun bool) error
}

const (
	CD_WORKFLOW_NAME        = "cd"
	CD_WORKFLOW_WITH_STAGES = "cd-stages-with-env"
)

type CdWorkflowServiceImpl struct {
	Logger               *zap.SugaredLogger
	config               *rest.Config
	cdConfig             *CdConfig
	appService           app.AppService
	envRepository        repository.EnvironmentRepository
	globalCMCSService    GlobalCMCSService
	argoWorkflowExecutor ArgoWorkflowExecutor
	k8sUtil              *k8s.K8sUtil
}

type CdWorkflowRequest struct {
	AppId                      int                                 `json:"appId"`
	EnvironmentId              int                                 `json:"envId"`
	WorkflowId                 int                                 `json:"workflowId"`
	WorkflowRunnerId           int                                 `json:"workflowRunnerId"`
	CdPipelineId               int                                 `json:"cdPipelineId"`
	TriggeredBy                int32                               `json:"triggeredBy"`
	StageYaml                  string                              `json:"stageYaml"`
	ArtifactLocation           string                              `json:"artifactLocation"`
	ArtifactBucket             string                              `json:"ciArtifactBucket"`
	ArtifactFileName           string                              `json:"ciArtifactFileName"`
	ArtifactRegion             string                              `json:"ciArtifactRegion"`
	CiProjectDetails           []CiProjectDetails                  `json:"ciProjectDetails"`
	CiArtifactDTO              CiArtifactDTO                       `json:"ciArtifactDTO"`
	Namespace                  string                              `json:"namespace"`
	WorkflowNamePrefix         string                              `json:"workflowNamePrefix"`
	CdImage                    string                              `json:"cdImage"`
	ActiveDeadlineSeconds      int64                               `json:"activeDeadlineSeconds"`
	StageType                  string                              `json:"stageType"`
	DockerUsername             string                              `json:"dockerUsername"`
	DockerPassword             string                              `json:"dockerPassword"`
	AwsRegion                  string                              `json:"awsRegion"`
	SecretKey                  string                              `json:"secretKey"`
	AccessKey                  string                              `json:"accessKey"`
	DockerConnection           string                              `json:"dockerConnection"`
	DockerCert                 string                              `json:"dockerCert"`
	CdCacheLocation            string                              `json:"cdCacheLocation"`
	CdCacheRegion              string                              `json:"cdCacheRegion"`
	DockerRegistryType         string                              `json:"dockerRegistryType"`
	DockerRegistryURL          string                              `json:"dockerRegistryURL"`
	OrchestratorHost           string                              `json:"orchestratorHost"`
	OrchestratorToken          string                              `json:"orchestratorToken"`
	IsExtRun                   bool                                `json:"isExtRun"`
	ExtraEnvironmentVariables  map[string]string                   `json:"extraEnvironmentVariables"`
	BlobStorageConfigured      bool                                `json:"blobStorageConfigured"`
	BlobStorageS3Config        *blob_storage.BlobStorageS3Config   `json:"blobStorageS3Config"`
	CloudProvider              blob_storage.BlobStorageType        `json:"cloudProvider"`
	AzureBlobConfig            *blob_storage.AzureBlobConfig       `json:"azureBlobConfig"`
	GcpBlobConfig              *blob_storage.GcpBlobConfig         `json:"gcpBlobConfig"`
	BlobStorageLogsKey         string                              `json:"blobStorageLogsKey"`
	InAppLoggingEnabled        bool                                `json:"inAppLoggingEnabled"`
	WorkflowPrefixForLog       string                              `json:"workflowPrefixForLog"`
	DefaultAddressPoolBaseCidr string                              `json:"defaultAddressPoolBaseCidr"`
	DefaultAddressPoolSize     int                                 `json:"defaultAddressPoolSize"`
	DeploymentTriggeredBy      string                              `json:"deploymentTriggeredBy,omitempty"`
	DeploymentTriggerTime      time.Time                           `json:"deploymentTriggerTime,omitempty"`
	DeploymentReleaseCounter   int                                 `json:"deploymentReleaseCounter,omitempty"`
	WorkflowExecutor           pipelineConfig.WorkflowExecutorType `json:"workflowExecutor"`
}

const PRE = "PRE"
const POST = "POST"

func NewCdWorkflowServiceImpl(Logger *zap.SugaredLogger, envRepository repository.EnvironmentRepository, cdConfig *CdConfig, appService app.AppService, globalCMCSService GlobalCMCSService, argoWorkflowExecutor ArgoWorkflowExecutor, k8sUtil *k8s.K8sUtil) *CdWorkflowServiceImpl {
	return &CdWorkflowServiceImpl{Logger: Logger,
		config:               cdConfig.ClusterConfig,
		cdConfig:             cdConfig,
		appService:           appService,
		envRepository:        envRepository,
		globalCMCSService:    globalCMCSService,
		argoWorkflowExecutor: argoWorkflowExecutor,
		k8sUtil:              k8sUtil,
	}
}

func (impl *CdWorkflowServiceImpl) SubmitWorkflow(workflowRequest *CdWorkflowRequest, pipeline *pipelineConfig.Pipeline, env *repository.Environment) error {

	containerEnvVariables := []v12.EnvVar{}
	if impl.cdConfig.CloudProvider == BLOB_STORAGE_S3 && impl.cdConfig.BlobStorageS3AccessKey != "" {
		miniCred := []v12.EnvVar{{Name: "AWS_ACCESS_KEY_ID", Value: impl.cdConfig.BlobStorageS3AccessKey}, {Name: "AWS_SECRET_ACCESS_KEY", Value: impl.cdConfig.BlobStorageS3SecretKey}}
		containerEnvVariables = append(containerEnvVariables, miniCred...)
	}
	if (workflowRequest.StageType == PRE && pipeline.RunPreStageInEnv) || (workflowRequest.StageType == POST && pipeline.RunPostStageInEnv) {
		workflowRequest.IsExtRun = true
	}
	ciCdTriggerEvent := CiCdTriggerEvent{
		Type:      cdStage,
		CdRequest: workflowRequest,
	}

	// key will be used for log archival through in-app logging
	ciCdTriggerEvent.CdRequest.BlobStorageLogsKey = fmt.Sprintf("%s/%s", impl.cdConfig.DefaultBuildLogsKeyPrefix, workflowRequest.WorkflowPrefixForLog)
	ciCdTriggerEvent.CdRequest.InAppLoggingEnabled = impl.cdConfig.InAppLoggingEnabled || (workflowRequest.WorkflowExecutor == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM)
	workflowJson, err := json.Marshal(&ciCdTriggerEvent)
	if err != nil {
		impl.Logger.Errorw("error occurred while marshalling ciCdTriggerEvent", "error", err)
		return err
	}

	privileged := true
	storageConfigured := workflowRequest.BlobStorageConfigured
	ttl := int32(impl.cdConfig.BuildLogTTLValue)
	workflowTemplate := bean3.WorkflowTemplate{}
	workflowTemplate.TTLValue = &ttl
	workflowTemplate.WorkflowId = workflowRequest.WorkflowId
	workflowTemplate.WorkflowRunnerId = workflowRequest.WorkflowRunnerId
	workflowTemplate.WorkflowRequestJson = string(workflowJson)

	var globalCmCsConfigs []*bean3.GlobalCMCSDto
	var workflowConfigMaps []bean.ConfigSecretMap
	var workflowSecrets []bean.ConfigSecretMap

	if !workflowRequest.IsExtRun {
		// inject global variables only if IsExtRun is false
		globalCmCsConfigs, err = impl.globalCMCSService.FindAllActiveByPipelineType(repository2.PIPELINE_TYPE_CD)
		if err != nil {
			impl.Logger.Errorw("error in getting all global cm/cs config", "err", err)
			return err
		}
		for i := range globalCmCsConfigs {
			globalCmCsConfigs[i].Name = fmt.Sprintf("%s-%s-%s", strings.ToLower(globalCmCsConfigs[i].Name), strconv.Itoa(workflowRequest.WorkflowRunnerId), CD_WORKFLOW_NAME)
		}

		workflowConfigMaps, workflowSecrets, err = GetFromGlobalCmCsDtos(globalCmCsConfigs)
		if err != nil {
			impl.Logger.Errorw("error in creating templates for global secrets", "err", err)
		}
	}

	cdPipelineLevelConfigMaps, cdPipelineLevelSecrets, err := impl.getConfiguredCmCs(pipeline, workflowRequest.StageType)
	if err != nil {
		impl.Logger.Errorw("error occurred while fetching pipeline configured cm and cs", "pipelineId", pipeline.Id, "err", err)
		return err
	}

	existingConfigMap, existingSecrets, err := impl.appService.GetCmSecretNew(workflowRequest.AppId, workflowRequest.EnvironmentId, false)
	if err != nil {
		impl.Logger.Errorw("failed to get configmap data", "err", err)
		return err
	}
	impl.Logger.Debugw("existing cm", "pipelineId", pipeline.Id, "cm", existingConfigMap)

	for _, cm := range existingConfigMap.Maps {
		if _, ok := cdPipelineLevelConfigMaps[cm.Name]; ok {
			if !cm.External {
				cm.Name = cm.Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + strconv.Itoa(workflowRequest.WorkflowRunnerId)
			}
			workflowConfigMaps = append(workflowConfigMaps, cm)
		}
	}

	for _, secret := range existingSecrets.Secrets {
		if _, ok := cdPipelineLevelSecrets[secret.Name]; ok {
			if !secret.External {
				secret.Name = secret.Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + strconv.Itoa(workflowRequest.WorkflowRunnerId)
			}
			workflowSecrets = append(workflowSecrets, *secret)
		}
	}

	workflowTemplate.ConfigMaps = workflowConfigMaps
	workflowTemplate.Secrets = workflowSecrets

	workflowTemplate.ServiceAccountName = impl.cdConfig.WorkflowServiceAccount
	workflowTemplate.NodeSelector = map[string]string{impl.cdConfig.TaintKey: impl.cdConfig.TaintValue}
	workflowTemplate.Tolerations = []v12.Toleration{{Key: impl.cdConfig.TaintKey, Value: impl.cdConfig.TaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}}
	workflowTemplate.Volumes = ExtractVolumesFromCmCs(workflowConfigMaps, workflowSecrets)
	workflowTemplate.ArchiveLogs = storageConfigured
	workflowTemplate.ArchiveLogs = workflowTemplate.ArchiveLogs && !ciCdTriggerEvent.CdRequest.InAppLoggingEnabled
	workflowTemplate.RestartPolicy = v12.RestartPolicyNever

	if len(impl.cdConfig.NodeLabel) > 0 {
		workflowTemplate.NodeSelector = impl.cdConfig.NodeLabel
	}

	limitCpu := impl.cdConfig.LimitCpu
	limitMem := impl.cdConfig.LimitMem
	reqCpu := impl.cdConfig.ReqCpu
	reqMem := impl.cdConfig.ReqMem

	eventEnv := v12.EnvVar{Name: "CI_CD_EVENT", Value: string(workflowJson)}
	inAppLoggingEnv := v12.EnvVar{Name: "IN_APP_LOGGING", Value: strconv.FormatBool(ciCdTriggerEvent.CdRequest.InAppLoggingEnabled)}
	containerEnvVariables = append(containerEnvVariables, eventEnv, inAppLoggingEnv)
	workflowMainContainer := v12.Container{
		Env:   containerEnvVariables,
		Name:  common.MainContainerName,
		Image: workflowRequest.CdImage,
		SecurityContext: &v12.SecurityContext{
			Privileged: &privileged,
		},
		Resources: v12.ResourceRequirements{
			Limits: v12.ResourceList{
				v12.ResourceCPU:    resource.MustParse(limitCpu),
				v12.ResourceMemory: resource.MustParse(limitMem),
			},
			Requests: v12.ResourceList{
				v12.ResourceCPU:    resource.MustParse(reqCpu),
				v12.ResourceMemory: resource.MustParse(reqMem),
			},
		},
	}
	UpdateContainerEnvsFromCmCs(&workflowMainContainer, workflowConfigMaps, workflowSecrets)

	impl.updateBlobStorageConfig(workflowRequest, &workflowTemplate, storageConfigured, ciCdTriggerEvent.CdRequest.BlobStorageLogsKey)
	workflowTemplate.Containers = []v12.Container{workflowMainContainer}
	workflowTemplate.WorkflowNamePrefix = workflowRequest.WorkflowNamePrefix
	workflowTemplate.WfControllerInstanceID = impl.cdConfig.WfControllerInstanceID
	workflowTemplate.ActiveDeadlineSeconds = &workflowRequest.ActiveDeadlineSeconds
	workflowTemplate.Namespace = workflowRequest.Namespace
	if workflowRequest.IsExtRun {
		configMap := env.Cluster.Config
		bearerToken := configMap[k8s.BearerToken]
		clusterConfig := &k8s.ClusterConfig{
			ClusterName:           env.Cluster.ClusterName,
			BearerToken:           bearerToken,
			Host:                  env.Cluster.ServerUrl,
			InsecureSkipTLSVerify: true,
		}
		restConfig, err2 := impl.k8sUtil.GetRestConfigByCluster(clusterConfig)
		if err2 != nil {
			impl.Logger.Errorw("error in getting rest config from cluster config", "err", err2, "appId", workflowRequest.AppId)
			return err2
		}
		workflowTemplate.ClusterConfig = restConfig
	} else {
		workflowTemplate.ClusterConfig = impl.config
	}
	workflowExecutor := impl.getWorkflowExecutor(workflowRequest.WorkflowExecutor)
	if workflowExecutor == nil {
		return errors.New("workflow executor not found")
	}
	return workflowExecutor.ExecuteWorkflow(workflowTemplate)
}

func (impl *CdWorkflowServiceImpl) updateBlobStorageConfig(workflowRequest *CdWorkflowRequest, workflowTemplate *bean3.WorkflowTemplate, storageConfigured bool, blobStorageKey string) {
	workflowTemplate.BlobStorageConfigured = storageConfigured && (impl.cdConfig.UseBlobStorageConfigInCdWorkflow || !workflowRequest.IsExtRun)
	workflowTemplate.BlobStorageS3Config = workflowRequest.BlobStorageS3Config
	workflowTemplate.AzureBlobConfig = workflowRequest.AzureBlobConfig
	workflowTemplate.GcpBlobConfig = workflowRequest.GcpBlobConfig
	workflowTemplate.CloudStorageKey = blobStorageKey
}

func (impl *CdWorkflowServiceImpl) getWorkflowExecutor(executorType pipelineConfig.WorkflowExecutorType) WorkflowExecutor {
	if executorType == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF {
		return impl.argoWorkflowExecutor
	}
	impl.Logger.Warnw("workflow executor not found", "type", executorType)
	return nil
}

func (impl *CdWorkflowServiceImpl) getConfiguredCmCs(pipeline *pipelineConfig.Pipeline, stage string) (map[string]bool, map[string]bool, error) {

	cdPipelineLevelConfigMaps := make(map[string]bool)
	cdPipelineLevelSecrets := make(map[string]bool)

	if stage == "PRE" {
		preStageConfigMapSecretsJson := pipeline.PreStageConfigMapSecretNames
		preStageConfigmapSecrets := bean2.PreStageConfigMapSecretNames{}
		err := json.Unmarshal([]byte(preStageConfigMapSecretsJson), &preStageConfigmapSecrets)
		if err != nil {
			return cdPipelineLevelConfigMaps, cdPipelineLevelSecrets, err
		}
		for _, cm := range preStageConfigmapSecrets.ConfigMaps {
			cdPipelineLevelConfigMaps[cm] = true
		}
		for _, secret := range preStageConfigmapSecrets.Secrets {
			cdPipelineLevelSecrets[secret] = true
		}
	} else {
		postStageConfigMapSecretsJson := pipeline.PostStageConfigMapSecretNames
		postStageConfigmapSecrets := bean2.PostStageConfigMapSecretNames{}
		err := json.Unmarshal([]byte(postStageConfigMapSecretsJson), &postStageConfigmapSecrets)
		if err != nil {
			return cdPipelineLevelConfigMaps, cdPipelineLevelSecrets, err
		}
		for _, cm := range postStageConfigmapSecrets.ConfigMaps {
			cdPipelineLevelConfigMaps[cm] = true
		}
		for _, secret := range postStageConfigmapSecrets.Secrets {
			cdPipelineLevelSecrets[secret] = true
		}
	}
	return cdPipelineLevelConfigMaps, cdPipelineLevelSecrets, nil
}

func (impl *CdWorkflowServiceImpl) GetWorkflow(name string, namespace string, restConfig *rest.Config, isExtRun bool) (*v1alpha1.Workflow, error) {
	impl.Logger.Debugw("getting wf", "name", name)
	var wfClient v1alpha12.WorkflowInterface
	var err error
	if isExtRun {
		wfClient, err = impl.getRuntimeEnvClientInstance(namespace, restConfig)

	} else {
		wfClient, err = impl.getClientInstance(namespace)
	}
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return nil, err
	}
	workflow, err := wfClient.Get(context.Background(), name, v1.GetOptions{})
	return workflow, err
}

func (impl *CdWorkflowServiceImpl) TerminateWorkflow(name string, namespace string, restConfig *rest.Config, isExtRun bool) error {
	impl.Logger.Debugw("terminating wf", "name", name)
	var wfClient v1alpha12.WorkflowInterface
	var err error
	if isExtRun {
		wfClient, err = impl.getRuntimeEnvClientInstance(namespace, restConfig)

	} else {
		wfClient, err = impl.getClientInstance(namespace)
	}
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return err
	}
	err = util.TerminateWorkflow(context.Background(), wfClient, name)
	return err
}

func (impl *CdWorkflowServiceImpl) UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
	impl.Logger.Debugw("updating wf", "name", wf.Name)
	wfClient, err := impl.getClientInstance(wf.Namespace)
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return nil, err
	}
	updatedWf, err := wfClient.Update(context.Background(), wf, v1.UpdateOptions{})
	if err != nil {
		impl.Logger.Errorw("cannot update wf ", "err", err)
		return nil, err
	}
	return updatedWf, err
}

func (impl *CdWorkflowServiceImpl) ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error) {
	wfClient, err := impl.getClientInstance(namespace)
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return nil, err
	}
	workflowList, err := wfClient.List(context.Background(), v1.ListOptions{})
	return workflowList, err
}

func (impl *CdWorkflowServiceImpl) DeleteWorkflow(wfName string, namespace string) error {
	wfClient, err := impl.getClientInstance(namespace)
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return err
	}
	err = wfClient.Delete(context.Background(), wfName, v1.DeleteOptions{})
	return err
}

func (impl *CdWorkflowServiceImpl) getClientInstance(namespace string) (v1alpha12.WorkflowInterface, error) {
	clientSet, err := versioned.NewForConfig(impl.config)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return nil, err
	}
	wfClient := clientSet.ArgoprojV1alpha1().Workflows(namespace) // create the workflow client
	return wfClient, nil
}

func (impl *CdWorkflowServiceImpl) getRuntimeEnvClientInstance(namespace string, restConfig *rest.Config) (v1alpha12.WorkflowInterface, error) {
	clientSet, err := versioned.NewForConfig(restConfig)
	if err != nil {
		impl.Logger.Errorw("err", "err", err)
		return nil, err
	}
	wfClient := clientSet.ArgoprojV1alpha1().Workflows(namespace) // create the workflow client
	return wfClient, nil
}

func (impl *CdWorkflowServiceImpl) checkErr(err error) {
	if err != nil {
		impl.Logger.Errorw("error", "error:", err)
	}
}
