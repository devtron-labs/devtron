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
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	v1alpha12 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/common"
	"github.com/argoproj/argo-workflows/v3/workflow/util"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/devtron/api/bean"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/util/k8s"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"strconv"
	"strings"
	"time"
)

type CommonWorkflowService interface {
	SubmitWorkflow(workflowRequest *CommonWorkflowRequest, pipeline *pipelineConfig.Pipeline, env *repository.Environment, appLabels map[string]string, isJob bool, isCi bool) error
	//DeleteWorkflow(wfName string, namespace string) error
	GetWorkflow(name string, namespace string, isExt bool, environment *repository.Environment) (*v1alpha1.Workflow, error)
	//ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error)
	//UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error)
	//TerminateWorkflow(executorType pipelineConfig.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config, isExtRun bool) error
	TerminateWorkflow(executorType pipelineConfig.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config, isExt bool, environment *repository.Environment) error
}

type CommonWorkflowServiceImpl struct {
	Logger                 *zap.SugaredLogger
	config                 *rest.Config
	ciCdConfig             *CiCdConfig
	appService             app.AppService
	envRepository          repository.EnvironmentRepository
	globalCMCSService      GlobalCMCSService
	argoWorkflowExecutor   ArgoWorkflowExecutor
	systemWorkflowExecutor SystemWorkflowExecutor
	k8sUtil                *k8s.K8sUtil
	k8sCommonService       k8s2.K8sCommonService
}

type CiCdTriggerEvent struct {
	Type                  string                 `json:"type"`
	CommonWorkflowRequest *CommonWorkflowRequest `json:"commonWorkflowRequest"`
}

type CommonWorkflowRequest struct {
	WorkflowNamePrefix         string                            `json:"workflowNamePrefix"`
	PipelineName               string                            `json:"pipelineName"`
	PipelineId                 int                               `json:"pipelineId"`
	DockerImageTag             string                            `json:"dockerImageTag"`
	DockerRegistryId           string                            `json:"dockerRegistryId"`
	DockerRegistryType         string                            `json:"dockerRegistryType"`
	DockerRegistryURL          string                            `json:"dockerRegistryURL"`
	DockerConnection           string                            `json:"dockerConnection"`
	DockerCert                 string                            `json:"dockerCert"`
	DockerRepository           string                            `json:"dockerRepository"`
	CheckoutPath               string                            `json:"checkoutPath"`
	DockerUsername             string                            `json:"dockerUsername"`
	DockerPassword             string                            `json:"dockerPassword"`
	AwsRegion                  string                            `json:"awsRegion"`
	AccessKey                  string                            `json:"accessKey"`
	SecretKey                  string                            `json:"secretKey"`
	CiCacheLocation            string                            `json:"ciCacheLocation"`
	CiCacheRegion              string                            `json:"ciCacheRegion"`
	CiCacheFileName            string                            `json:"ciCacheFileName"`
	CiProjectDetails           []CiProjectDetails                `json:"ciProjectDetails"`
	ContainerResources         ContainerResources                `json:"containerResources"`
	ActiveDeadlineSeconds      int64                             `json:"activeDeadlineSeconds"`
	CiImage                    string                            `json:"ciImage"`
	Namespace                  string                            `json:"namespace"`
	WorkflowId                 int                               `json:"workflowId"`
	TriggeredBy                int32                             `json:"triggeredBy"`
	CacheLimit                 int64                             `json:"cacheLimit"`
	BeforeDockerBuildScripts   []*bean2.CiScript                 `json:"beforeDockerBuildScripts"`
	AfterDockerBuildScripts    []*bean2.CiScript                 `json:"afterDockerBuildScripts"`
	CiArtifactLocation         string                            `json:"ciArtifactLocation"`
	CiArtifactBucket           string                            `json:"ciArtifactBucket"`
	CiArtifactFileName         string                            `json:"ciArtifactFileName"`
	CiArtifactRegion           string                            `json:"ciArtifactRegion"`
	ScanEnabled                bool                              `json:"scanEnabled"`
	CloudProvider              blob_storage.BlobStorageType      `json:"cloudProvider"`
	BlobStorageConfigured      bool                              `json:"blobStorageConfigured"`
	BlobStorageS3Config        *blob_storage.BlobStorageS3Config `json:"blobStorageS3Config"`
	AzureBlobConfig            *blob_storage.AzureBlobConfig     `json:"azureBlobConfig"`
	GcpBlobConfig              *blob_storage.GcpBlobConfig       `json:"gcpBlobConfig"`
	BlobStorageLogsKey         string                            `json:"blobStorageLogsKey"`
	InAppLoggingEnabled        bool                              `json:"inAppLoggingEnabled"`
	DefaultAddressPoolBaseCidr string                            `json:"defaultAddressPoolBaseCidr"`
	DefaultAddressPoolSize     int                               `json:"defaultAddressPoolSize"`
	PreCiSteps                 []*bean3.StepObject               `json:"preCiSteps"`
	PostCiSteps                []*bean3.StepObject               `json:"postCiSteps"`
	RefPlugins                 []*bean3.RefPluginObject          `json:"refPlugins"`
	AppName                    string                            `json:"appName"`
	TriggerByAuthor            string                            `json:"triggerByAuthor"`
	CiBuildConfig              *bean3.CiBuildConfigBean          `json:"ciBuildConfig"`
	CiBuildDockerMtuValue      int                               `json:"ciBuildDockerMtuValue"`
	IgnoreDockerCachePush      bool                              `json:"ignoreDockerCachePush"`
	IgnoreDockerCachePull      bool                              `json:"ignoreDockerCachePull"`
	CacheInvalidate            bool                              `json:"cacheInvalidate"`
	IsPvcMounted               bool                              `json:"IsPvcMounted"`
	ExtraEnvironmentVariables  map[string]string                 `json:"extraEnvironmentVariables"`
	EnableBuildContext         bool                              `json:"enableBuildContext"`
	AppId                      int                               `json:"appId"`
	EnvironmentId              int                               `json:"environmentId"`
	OrchestratorHost           string                            `json:"orchestratorHost"`
	OrchestratorToken          string                            `json:"orchestratorToken"`
	IsExtRun                   bool                              `json:"isExtRun"`
	ImageRetryCount            int                               `json:"imageRetryCount"`
	ImageRetryInterval         int                               `json:"imageRetryInterval"`
	// Data from CD Workflow service
	WorkflowRunnerId         int                                 `json:"workflowRunnerId"`
	CdPipelineId             int                                 `json:"cdPipelineId"`
	StageYaml                string                              `json:"stageYaml"`
	ArtifactLocation         string                              `json:"artifactLocation"`
	CiArtifactDTO            CiArtifactDTO                       `json:"ciArtifactDTO"`
	CdImage                  string                              `json:"cdImage"`
	StageType                string                              `json:"stageType"`
	CdCacheLocation          string                              `json:"cdCacheLocation"`
	CdCacheRegion            string                              `json:"cdCacheRegion"`
	WorkflowPrefixForLog     string                              `json:"workflowPrefixForLog"`
	DeploymentTriggeredBy    string                              `json:"deploymentTriggeredBy,omitempty"`
	DeploymentTriggerTime    time.Time                           `json:"deploymentTriggerTime,omitempty"`
	DeploymentReleaseCounter int                                 `json:"deploymentReleaseCounter,omitempty"`
	WorkflowExecutor         pipelineConfig.WorkflowExecutorType `json:"workflowExecutor"`
	PrePostDeploySteps       []*bean3.StepObject                 `json:"prePostDeploySteps"`
}

const (
	BLOB_STORAGE_AZURE             = "AZURE"
	BLOB_STORAGE_S3                = "S3"
	BLOB_STORAGE_GCP               = "GCP"
	BLOB_STORAGE_MINIO             = "MINIO"
	CI_WORKFLOW_NAME               = "ci"
	CI_WORKFLOW_WITH_STAGES        = "ci-stages-with-env"
	CI_NODE_SELECTOR_APP_LABEL_KEY = "devtron.ai/node-selector"
	CI_NODE_PVC_ALL_ENV            = "devtron.ai/ci-pvc-all"
	CI_NODE_PVC_PIPELINE_PREFIX    = "devtron.ai/ci-pvc"
	PRE                            = "PRE"
	POST                           = "POST"
	ciEvent                        = "CI"
	cdStage                        = "CD"
	CD_WORKFLOW_NAME               = "cd"
	CD_WORKFLOW_WITH_STAGES        = "cd-stages-with-env"
)

type ContainerResources struct {
	MinCpu        string `json:"minCpu"`
	MaxCpu        string `json:"maxCpu"`
	MinStorage    string `json:"minStorage"`
	MaxStorage    string `json:"maxStorage"`
	MinEphStorage string `json:"minEphStorage"`
	MaxEphStorage string `json:"maxEphStorage"`
	MinMem        string `json:"minMem"`
	MaxMem        string `json:"maxMem"`
}
type CiProjectDetails struct {
	GitRepository   string `json:"gitRepository"`
	MaterialName    string `json:"materialName"`
	CheckoutPath    string `json:"checkoutPath"`
	FetchSubmodules bool   `json:"fetchSubmodules"`
	CommitHash      string `json:"commitHash"`
	GitTag          string `json:"gitTag"`
	CommitTime      string `json:"commitTime"`
	//Branch        string          `json:"branch"`
	Type        string                    `json:"type"`
	Message     string                    `json:"message"`
	Author      string                    `json:"author"`
	GitOptions  GitOptions                `json:"gitOptions"`
	SourceType  pipelineConfig.SourceType `json:"sourceType"`
	SourceValue string                    `json:"sourceValue"`
	WebhookData pipelineConfig.WebhookData
}
type GitOptions struct {
	UserName      string               `json:"userName"`
	Password      string               `json:"password"`
	SshPrivateKey string               `json:"sshPrivateKey"`
	AccessToken   string               `json:"accessToken"`
	AuthMode      repository2.AuthMode `json:"authMode"`
}

func NewCommonWorkflowServiceImpl(Logger *zap.SugaredLogger, envRepository repository.EnvironmentRepository, ciCdConfig *CiCdConfig,
	appService app.AppService, globalCMCSService GlobalCMCSService, argoWorkflowExecutor ArgoWorkflowExecutor,
	k8sUtil *k8s.K8sUtil,
	systemWorkflowExecutor SystemWorkflowExecutor, k8sCommonService k8s2.K8sCommonService) (*CommonWorkflowServiceImpl, error) {
	commonWorkflowService := &CommonWorkflowServiceImpl{Logger: Logger,
		ciCdConfig:             ciCdConfig,
		appService:             appService,
		envRepository:          envRepository,
		globalCMCSService:      globalCMCSService,
		argoWorkflowExecutor:   argoWorkflowExecutor,
		k8sUtil:                k8sUtil,
		systemWorkflowExecutor: systemWorkflowExecutor,
		k8sCommonService:       k8sCommonService,
	}
	restConfig, err := k8sUtil.GetK8sInClusterRestConfig()
	if err != nil {
		Logger.Errorw("error in getting in cluster rest config", "err", err)
		return nil, err
	}
	commonWorkflowService.config = restConfig
	return commonWorkflowService, nil
}

func (impl *CommonWorkflowServiceImpl) SubmitWorkflow(workflowRequest *CommonWorkflowRequest, pipeline *pipelineConfig.Pipeline, env *repository.Environment, appLabels map[string]string, isJob bool, isCi bool) error {

	containerEnvVariables := []v12.EnvVar{}
	getNodeLabel(impl.ciCdConfig, isCi)
	if isCi {
		containerEnvVariables = []v12.EnvVar{{Name: "IMAGE_SCANNER_ENDPOINT", Value: impl.ciCdConfig.ImageScannerEndpoint}}
	}
	if impl.ciCdConfig.CloudProvider == BLOB_STORAGE_S3 && impl.ciCdConfig.BlobStorageS3AccessKey != "" {
		miniCred := []v12.EnvVar{{Name: "AWS_ACCESS_KEY_ID", Value: impl.ciCdConfig.BlobStorageS3AccessKey}, {Name: "AWS_SECRET_ACCESS_KEY", Value: impl.ciCdConfig.BlobStorageS3SecretKey}}
		containerEnvVariables = append(containerEnvVariables, miniCred...)
	}
	if (workflowRequest.StageType == PRE && pipeline.RunPreStageInEnv) || (workflowRequest.StageType == POST && pipeline.RunPostStageInEnv) {
		workflowRequest.IsExtRun = true
	}
	pvc := appLabels[strings.ToLower(fmt.Sprintf("%s-%s", CI_NODE_PVC_PIPELINE_PREFIX, workflowRequest.PipelineName))]
	if len(pvc) == 0 {
		pvc = appLabels[CI_NODE_PVC_ALL_ENV]
	}
	if len(pvc) != 0 {
		workflowRequest.IsPvcMounted = true
		workflowRequest.IgnoreDockerCachePush = true
		workflowRequest.IgnoreDockerCachePull = true
	}
	eventType := cdStage
	if isCi {
		eventType = ciEvent
	}
	ciCdTriggerEvent := CiCdTriggerEvent{
		Type:                  eventType,
		CommonWorkflowRequest: workflowRequest,
	}
	if env != nil && env.Id != 0 && isCi {
		workflowRequest.IsExtRun = true
	}
	// key will be used for log archival through in-app logging
	if isCi {
		ciCdTriggerEvent.CommonWorkflowRequest.BlobStorageLogsKey = fmt.Sprintf("%s/%s", impl.ciCdConfig.CiDefaultBuildLogsKeyPrefix, workflowRequest.WorkflowPrefixForLog)
	} else {
		ciCdTriggerEvent.CommonWorkflowRequest.BlobStorageLogsKey = fmt.Sprintf("%s/%s", impl.ciCdConfig.CdDefaultBuildLogsKeyPrefix, workflowRequest.WorkflowPrefixForLog)
	}
	ciCdTriggerEvent.CommonWorkflowRequest.InAppLoggingEnabled = impl.ciCdConfig.InAppLoggingEnabled || (workflowRequest.WorkflowExecutor == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM)
	workflowJson, err := json.Marshal(&ciCdTriggerEvent)
	if err != nil {
		impl.Logger.Errorw("error occurred while marshalling ciCdTriggerEvent", "error", err)
		return err
	}

	privileged := true
	storageConfigured := workflowRequest.BlobStorageConfigured
	ttl := int32(impl.ciCdConfig.BuildLogTTLValue)
	workflowTemplate := bean3.WorkflowTemplate{}
	workflowTemplate.TTLValue = &ttl
	workflowTemplate.WorkflowId = workflowRequest.WorkflowId
	if !isCi {
		workflowTemplate.WorkflowRunnerId = workflowRequest.WorkflowRunnerId
		workflowTemplate.PrePostDeploySteps = workflowRequest.PrePostDeploySteps
	}
	workflowTemplate.WorkflowRequestJson = string(workflowJson)
	workflowTemplate.RefPlugins = workflowRequest.RefPlugins

	var globalCmCsConfigs []*bean3.GlobalCMCSDto
	var workflowConfigMaps []bean.ConfigSecretMap
	var workflowSecrets []bean.ConfigSecretMap

	if !workflowRequest.IsExtRun {
		// inject global variables only if IsExtRun is false
		if isCi {
			globalCmCsConfigs, err = impl.globalCMCSService.FindAllActiveByPipelineType(repository2.PIPELINE_TYPE_CI)
		} else {
			globalCmCsConfigs, err = impl.globalCMCSService.FindAllActiveByPipelineType(repository2.PIPELINE_TYPE_CD)
		}
		if err != nil {
			impl.Logger.Errorw("error in getting all global cm/cs config", "err", err)
			return err
		}
		if isCi {
			for i := range globalCmCsConfigs {
				globalCmCsConfigs[i].Name = strings.ToLower(globalCmCsConfigs[i].Name) + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + CI_WORKFLOW_NAME
			}
		} else {
			for i := range globalCmCsConfigs {
				globalCmCsConfigs[i].Name = fmt.Sprintf("%s-%s-%s", strings.ToLower(globalCmCsConfigs[i].Name), strconv.Itoa(workflowRequest.WorkflowRunnerId), CD_WORKFLOW_NAME)
			}
		}
		workflowConfigMaps, workflowSecrets, err = GetFromGlobalCmCsDtos(globalCmCsConfigs)
		if err != nil {
			impl.Logger.Errorw("error in creating templates for global secrets", "err", err)
		}
	}
	var cdPipelineLevelConfigMaps, cdPipelineLevelSecrets map[string]bool
	if !isCi {
		cdPipelineLevelConfigMaps, cdPipelineLevelSecrets, err = impl.getConfiguredCmCs(pipeline, workflowRequest.StageType)
		if err != nil {
			impl.Logger.Errorw("error occurred while fetching pipeline configured cm and cs", "pipelineId", pipeline.Id, "err", err)
			return err
		}
	}
	var existingConfigMap *bean.ConfigMapJson
	var existingSecrets *bean.ConfigSecretJson
	if !isCi || isJob {
		existingConfigMap, existingSecrets, err = impl.appService.GetCmSecretNew(workflowRequest.AppId, workflowRequest.EnvironmentId, isJob)
		if err != nil {
			impl.Logger.Errorw("failed to get configmap data", "err", err)
			return err
		}
		impl.Logger.Debugw("existing cm", "cm", existingConfigMap, "secrets", existingSecrets)
	}

	if isCi && isJob {
		for _, cm := range existingConfigMap.Maps {
			if !cm.External {
				cm.Name = cm.Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + CI_WORKFLOW_NAME
			}
			workflowConfigMaps = append(workflowConfigMaps, cm)
		}

		for _, secret := range existingSecrets.Secrets {
			if !secret.External {
				secret.Name = secret.Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + CI_WORKFLOW_NAME
			}
			workflowSecrets = append(workflowSecrets, *secret)
		}
	} else if !isCi {

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
	}
	workflowTemplate.ConfigMaps = workflowConfigMaps
	workflowTemplate.Secrets = workflowSecrets
	if isCi {
		workflowTemplate.ServiceAccountName = impl.ciCdConfig.CiWorkflowServiceAccount
		if impl.ciCdConfig.CiTaintKey != "" || impl.ciCdConfig.CiTaintValue != "" {
			workflowTemplate.Tolerations = []v12.Toleration{{Key: impl.ciCdConfig.CiTaintKey, Value: impl.ciCdConfig.CiTaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}}
		}
		// In the future, we will give support for NodeSelector for job currently we need to have a node without dedicated NodeLabel to run job
		if len(impl.ciCdConfig.NodeLabel) > 0 && !(isJob && workflowRequest.IsExtRun) {
			workflowTemplate.NodeSelector = impl.ciCdConfig.NodeLabel
		}
	} else {
		workflowTemplate.ServiceAccountName = impl.ciCdConfig.CdWorkflowServiceAccount
		workflowTemplate.NodeSelector = map[string]string{impl.ciCdConfig.CdTaintKey: impl.ciCdConfig.CdTaintValue}
		workflowTemplate.Tolerations = []v12.Toleration{{Key: impl.ciCdConfig.CdTaintKey, Value: impl.ciCdConfig.CdTaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}}
	}
	workflowTemplate.Volumes = ExtractVolumesFromCmCs(workflowConfigMaps, workflowSecrets)
	workflowTemplate.ArchiveLogs = storageConfigured
	workflowTemplate.ArchiveLogs = workflowTemplate.ArchiveLogs && !ciCdTriggerEvent.CommonWorkflowRequest.InAppLoggingEnabled
	workflowTemplate.RestartPolicy = v12.RestartPolicyNever

	if len(impl.ciCdConfig.NodeLabel) > 0 {
		workflowTemplate.NodeSelector = impl.ciCdConfig.NodeLabel
	}
	var limitCpu, limitMem, reqCpu, reqMem string
	if isCi {
		limitCpu = impl.ciCdConfig.CiLimitCpu
		limitMem = impl.ciCdConfig.CiLimitMem
		reqCpu = impl.ciCdConfig.CiReqCpu
		reqMem = impl.ciCdConfig.CiReqMem
	} else {
		limitCpu = impl.ciCdConfig.CdLimitCpu
		limitMem = impl.ciCdConfig.CdLimitMem
		reqCpu = impl.ciCdConfig.CdReqCpu
		reqMem = impl.ciCdConfig.CdReqMem
	}

	eventEnv := v12.EnvVar{Name: "CI_CD_EVENT", Value: string(workflowJson)}
	inAppLoggingEnv := v12.EnvVar{Name: "IN_APP_LOGGING", Value: strconv.FormatBool(ciCdTriggerEvent.CommonWorkflowRequest.InAppLoggingEnabled)}
	containerEnvVariables = append(containerEnvVariables, eventEnv, inAppLoggingEnv)
	workflowImage := workflowRequest.CdImage
	if isCi {
		workflowImage = workflowRequest.CiImage
	}
	workflowMainContainer := v12.Container{
		Env:   containerEnvVariables,
		Name:  common.MainContainerName,
		Image: workflowImage,
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
	if len(pvc) != 0 && isCi {
		buildPvcCachePath := impl.ciCdConfig.BuildPvcCachePath
		buildxPvcCachePath := impl.ciCdConfig.BuildxPvcCachePath
		defaultPvcCachePath := impl.ciCdConfig.DefaultPvcCachePath

		workflowTemplate.Volumes = append(workflowTemplate.Volumes, v12.Volume{
			Name: "root-vol",
			VolumeSource: v12.VolumeSource{
				PersistentVolumeClaim: &v12.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc,
					ReadOnly:  false,
				},
			},
		})
		workflowMainContainer.VolumeMounts = append(workflowMainContainer.VolumeMounts,
			v12.VolumeMount{
				Name:      "root-vol",
				MountPath: buildPvcCachePath,
			},
			v12.VolumeMount{
				Name:      "root-vol",
				MountPath: buildxPvcCachePath,
			},
			v12.VolumeMount{
				Name:      "root-vol",
				MountPath: defaultPvcCachePath,
			})
	}
	UpdateContainerEnvsFromCmCs(&workflowMainContainer, workflowConfigMaps, workflowSecrets)

	impl.updateBlobStorageConfig(workflowRequest, &workflowTemplate, storageConfigured, ciCdTriggerEvent.CommonWorkflowRequest.BlobStorageLogsKey)
	workflowTemplate.Containers = []v12.Container{workflowMainContainer}
	workflowTemplate.WorkflowNamePrefix = workflowRequest.WorkflowNamePrefix
	if !isCi {
		workflowTemplate.WfControllerInstanceID = impl.ciCdConfig.WfControllerInstanceID
		workflowTemplate.TerminationGracePeriod = impl.ciCdConfig.TerminationGracePeriod
	}
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
	if isCi {
		workflowTemplate.WorkflowType = CI_WORKFLOW_NAME
	} else {
		workflowTemplate.WorkflowType = CD_WORKFLOW_NAME
	}
	_, err = workflowExecutor.ExecuteWorkflow(workflowTemplate)
	return err
}

func (impl *CommonWorkflowServiceImpl) getConfiguredCmCs(pipeline *pipelineConfig.Pipeline, stage string) (map[string]bool, map[string]bool, error) {

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

func (impl *CommonWorkflowServiceImpl) updateBlobStorageConfig(workflowRequest *CommonWorkflowRequest, workflowTemplate *bean3.WorkflowTemplate, storageConfigured bool, blobStorageKey string) {
	workflowTemplate.BlobStorageConfigured = storageConfigured && (impl.ciCdConfig.UseBlobStorageConfigInCdWorkflow || !workflowRequest.IsExtRun)
	workflowTemplate.BlobStorageS3Config = workflowRequest.BlobStorageS3Config
	workflowTemplate.AzureBlobConfig = workflowRequest.AzureBlobConfig
	workflowTemplate.GcpBlobConfig = workflowRequest.GcpBlobConfig
	workflowTemplate.CloudStorageKey = blobStorageKey
}

func (impl *CommonWorkflowServiceImpl) getWorkflowExecutor(executorType pipelineConfig.WorkflowExecutorType) WorkflowExecutor {
	if executorType == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF {
		return impl.argoWorkflowExecutor
	} else if executorType == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM {
		return impl.systemWorkflowExecutor
	}
	impl.Logger.Warnw("workflow executor not found", "type", executorType)
	return nil
}
func (impl *CommonWorkflowServiceImpl) GetWorkflow(name string, namespace string, isExt bool, environment *repository.Environment) (*v1alpha1.Workflow, error) {
	impl.Logger.Debug("getting wf", name)
	wfClient, err := impl.getWfClient(environment, namespace, isExt)

	if err != nil {
		return nil, err
	}

	workflow, err := wfClient.Get(context.Background(), name, v1.GetOptions{})
	return workflow, err
}

func (impl *CommonWorkflowServiceImpl) TerminateWorkflow(executorType pipelineConfig.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config, isExt bool, environment *repository.Environment) error {
	impl.Logger.Debugw("terminating wf", "name", name)
	var err error
	if executorType != "" {
		workflowExecutor := impl.getWorkflowExecutor(executorType)
		err = workflowExecutor.TerminateWorkflow(name, namespace, restConfig)
	} else {
		wfClient, err := impl.getWfClient(environment, namespace, isExt)
		if err != nil {
			return err
		}
		err = util.TerminateWorkflow(context.Background(), wfClient, name)
	}
	return err
}
func (impl *CommonWorkflowServiceImpl) getRuntimeEnvClientInstance(environment *repository.Environment) (v1alpha12.WorkflowInterface, error) {
	restConfig, err, _ := impl.k8sCommonService.GetRestConfigByClusterId(context.Background(), environment.ClusterId)
	if err != nil {
		impl.Logger.Errorw("error in getting rest config by cluster id", "err", err)
		return nil, err
	}
	wfClient, err := GetClientInstance(restConfig, environment.Namespace)
	if err != nil {
		impl.Logger.Errorw("error in getting wfClient", "err", err)
		return nil, err
	}
	return wfClient, nil
}

//func (impl *CommonWorkflowServiceImpl) getClientInstance(namespace string) (v1alpha12.WorkflowInterface, error) {
//	clientSet, err := versioned.NewForConfig(impl.config)
//	if err != nil {
//		impl.Logger.Errorw("err on get client instance", "err", err)
//		return nil, err
//	}
//	wfClient := clientSet.ArgoprojV1alpha1().Workflows(namespace) // create the workflow client
//	return wfClient, nil
//}

func (impl *CommonWorkflowServiceImpl) getWfClient(environment *repository.Environment, namespace string, isExt bool) (v1alpha12.WorkflowInterface, error) {
	var wfClient v1alpha12.WorkflowInterface
	var err error
	if isExt {
		wfClient, err = impl.getRuntimeEnvClientInstance(environment)
		if err != nil {
			impl.Logger.Errorw("cannot build wf client", "err", err)
			return nil, err
		}
	} else {
		wfClient, err = GetClientInstance(impl.config, namespace)
		if err != nil {
			impl.Logger.Errorw("cannot build wf client", "err", err)
			return nil, err
		}
	}
	return wfClient, nil
}
