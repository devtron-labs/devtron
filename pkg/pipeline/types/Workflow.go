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

package types

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/workflow/common"
	"github.com/devtron-labs/common-lib/blob-storage"
	bean3 "github.com/devtron-labs/devtron/api/bean"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
	"time"
)

type WorkflowRequest struct {
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
	CiProjectDetails           []bean.CiProjectDetails           `json:"ciProjectDetails"`
	ContainerResources         bean.ContainerResources           `json:"containerResources"`
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
	PreCiSteps                 []*bean.StepObject                `json:"preCiSteps"`
	PostCiSteps                []*bean.StepObject                `json:"postCiSteps"`
	RefPlugins                 []*bean.RefPluginObject           `json:"refPlugins"`
	AppName                    string                            `json:"appName"`
	TriggerByAuthor            string                            `json:"triggerByAuthor"`
	CiBuildConfig              *bean.CiBuildConfigBean           `json:"ciBuildConfig"`
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
	PrePostDeploySteps       []*bean.StepObject                  `json:"prePostDeploySteps"`
	CiArtifactLastFetch      time.Time                           `json:"ciArtifactLastFetch"`
	CiPipelineType           string                              `json:"ciPipelineType"`
	UseExternalClusterBlob   bool                                `json:"useExternalClusterBlob"`
	Type                     bean.WorkflowPipelineType
	Pipeline                 *pipelineConfig.Pipeline
	Env                      *repository.Environment
	AppLabels                map[string]string
}

func (workflowRequest *WorkflowRequest) updateExternalRunMetadata() {
	pipeline := workflowRequest.Pipeline
	env := workflowRequest.Env
	// Check for external in case of PRE-/POST-CD
	if (workflowRequest.StageType == PRE && pipeline.RunPreStageInEnv) || (workflowRequest.StageType == POST && pipeline.RunPostStageInEnv) {
		workflowRequest.IsExtRun = true
	}
	// Check for external in case of JOB
	if env != nil && env.Id != 0 && workflowRequest.CheckForJob() {
		workflowRequest.EnvironmentId = env.Id
		workflowRequest.IsExtRun = true
	}
}

func (workflowRequest *WorkflowRequest) CheckBlobStorageConfig(config *CiCdConfig) bool {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return config.UseBlobStorageConfigInCiWorkflow
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return config.UseBlobStorageConfigInCdWorkflow
	default:
		return false
	}

}

func (workflowRequest *WorkflowRequest) updateUseExternalClusterBlob(config *CiCdConfig) {
	workflowRequest.UseExternalClusterBlob = !workflowRequest.CheckBlobStorageConfig(config) && workflowRequest.IsExtRun
}

func (workflowRequest *WorkflowRequest) GetWorkflowTemplate(workflowJson []byte, config *CiCdConfig) bean.WorkflowTemplate {

	ttl := int32(config.BuildLogTTLValue)
	workflowTemplate := bean.WorkflowTemplate{}
	workflowTemplate.TTLValue = &ttl
	workflowTemplate.WorkflowId = workflowRequest.WorkflowId
	workflowTemplate.WorkflowRequestJson = string(workflowJson)
	workflowTemplate.RefPlugins = workflowRequest.RefPlugins
	workflowTemplate.ActiveDeadlineSeconds = &workflowRequest.ActiveDeadlineSeconds
	workflowTemplate.Namespace = workflowRequest.Namespace
	workflowTemplate.WorkflowNamePrefix = workflowRequest.WorkflowNamePrefix
	if workflowRequest.Type == bean.CD_WORKFLOW_PIPELINE_TYPE {
		workflowTemplate.WorkflowRunnerId = workflowRequest.WorkflowRunnerId
		workflowTemplate.PrePostDeploySteps = workflowRequest.PrePostDeploySteps
	}
	return workflowTemplate
}

func (workflowRequest *WorkflowRequest) checkConfigType(config *CiCdConfig) {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		config.Type = CiConfigType
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		config.Type = CdConfigType
	}
}

func (workflowRequest *WorkflowRequest) GetBlobStorageLogsKey(config *CiCdConfig) string {
	return fmt.Sprintf("%s/%s", config.GetDefaultBuildLogsKeyPrefix(), workflowRequest.WorkflowPrefixForLog)
}

func (workflowRequest *WorkflowRequest) GetWorkflowJson(config *CiCdConfig) ([]byte, error) {
	workflowRequest.updateBlobStorageLogsKey(config)
	workflowRequest.updateExternalRunMetadata()
	workflowRequest.updateUseExternalClusterBlob(config)
	workflowJson, err := workflowRequest.getWorkflowJson()
	if err != nil {
		return nil, err
	}
	return workflowJson, err
}

func (workflowRequest *WorkflowRequest) GetEventTypeForWorkflowRequest() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return bean.CiStage
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return bean.CdStage
	default:
		return ""
	}
}

func (workflowRequest *WorkflowRequest) GetWorkflowTypeForWorkflowRequest() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return bean.CI_WORKFLOW_NAME
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return bean.CD_WORKFLOW_NAME
	default:
		return ""
	}
}

func (workflowRequest *WorkflowRequest) getContainerEnvVariables(config *CiCdConfig, workflowJson []byte) (containerEnvVariables []v1.EnvVar) {
	if workflowRequest.Type == bean.CI_WORKFLOW_PIPELINE_TYPE ||
		workflowRequest.Type == bean.JOB_WORKFLOW_PIPELINE_TYPE {
		containerEnvVariables = []v1.EnvVar{{Name: "IMAGE_SCANNER_ENDPOINT", Value: config.ImageScannerEndpoint}}
	}
	eventEnv := v1.EnvVar{Name: "CI_CD_EVENT", Value: string(workflowJson)}
	inAppLoggingEnv := v1.EnvVar{Name: "IN_APP_LOGGING", Value: strconv.FormatBool(workflowRequest.InAppLoggingEnabled)}
	containerEnvVariables = append(containerEnvVariables, eventEnv, inAppLoggingEnv)
	return containerEnvVariables
}

func (workflowRequest *WorkflowRequest) getPVCForWorkflowRequest() string {
	var pvc string
	workflowRequestType := workflowRequest.Type
	if workflowRequestType == bean.CI_WORKFLOW_PIPELINE_TYPE ||
		workflowRequestType == bean.JOB_WORKFLOW_PIPELINE_TYPE {
		pvc = workflowRequest.AppLabels[strings.ToLower(fmt.Sprintf("%s-%s", CI_NODE_PVC_PIPELINE_PREFIX, workflowRequest.PipelineName))]
		if len(pvc) == 0 {
			pvc = workflowRequest.AppLabels[CI_NODE_PVC_ALL_ENV]
		}
		if len(pvc) != 0 {
			workflowRequest.IsPvcMounted = true
			workflowRequest.IgnoreDockerCachePush = true
			workflowRequest.IgnoreDockerCachePull = true
		}
	} else {
		//pvc not supported for other then ci and job currently
	}
	return pvc
}

func (workflowRequest *WorkflowRequest) getDefaultBuildLogsKeyPrefix(config *CiCdConfig) string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return config.CiDefaultBuildLogsKeyPrefix
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return config.CdDefaultBuildLogsKeyPrefix
	default:
		return ""
	}
}

func (workflowRequest *WorkflowRequest) getBlobStorageLogsPrefix() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return workflowRequest.WorkflowNamePrefix
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return workflowRequest.WorkflowPrefixForLog
	default:
		return ""
	}
}

func (workflowRequest *WorkflowRequest) updateBlobStorageLogsKey(config *CiCdConfig) {
	workflowRequest.BlobStorageLogsKey = fmt.Sprintf("%s/%s", workflowRequest.getDefaultBuildLogsKeyPrefix(config), workflowRequest.getBlobStorageLogsPrefix())
	workflowRequest.InAppLoggingEnabled = config.InAppLoggingEnabled || (workflowRequest.WorkflowExecutor == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM)
}

func (workflowRequest *WorkflowRequest) getWorkflowJson() ([]byte, error) {
	eventType := workflowRequest.GetEventTypeForWorkflowRequest()
	ciCdTriggerEvent := CiCdTriggerEvent{
		Type:                  eventType,
		CommonWorkflowRequest: workflowRequest,
	}
	workflowJson, err := json.Marshal(&ciCdTriggerEvent)
	if err != nil {
		return nil, err
	}
	return workflowJson, err
}

func (workflowRequest *WorkflowRequest) AddNodeConstraintsFromConfig(workflowTemplate *bean.WorkflowTemplate, config *CiCdConfig) {
	nodeConstraints := workflowRequest.GetNodeConstraints(config)
	if workflowRequest.Type == bean.CD_WORKFLOW_PIPELINE_TYPE && nodeConstraints.TaintKey != "" {
		workflowTemplate.NodeSelector = map[string]string{nodeConstraints.TaintKey: nodeConstraints.TaintValue}
	}
	workflowTemplate.ServiceAccountName = nodeConstraints.ServiceAccount
	if nodeConstraints.TaintKey != "" || nodeConstraints.TaintValue != "" {
		workflowTemplate.Tolerations = []v1.Toleration{{Key: nodeConstraints.TaintKey, Value: nodeConstraints.TaintValue, Operator: v1.TolerationOpEqual, Effect: v1.TaintEffectNoSchedule}}
	}
	// In the future, we will give support for NodeSelector for job currently we need to have a node without dedicated NodeLabel to run job
	if len(nodeConstraints.NodeLabel) > 0 && !(nodeConstraints.SkipNodeSelector) {
		workflowTemplate.NodeSelector = nodeConstraints.NodeLabel
	}
	workflowTemplate.ArchiveLogs = workflowRequest.BlobStorageConfigured && !workflowRequest.InAppLoggingEnabled
	workflowTemplate.RestartPolicy = v1.RestartPolicyNever

}

func (workflowRequest *WorkflowRequest) GetGlobalCmCsNamePrefix() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return strconv.Itoa(workflowRequest.WorkflowId) + "-" + bean.CI_WORKFLOW_NAME
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return strconv.Itoa(workflowRequest.WorkflowRunnerId) + "-" + bean.CD_WORKFLOW_NAME
	default:
		return ""
	}
}

func (workflowRequest *WorkflowRequest) GetConfiguredCmCs() (map[string]bool, map[string]bool, error) {

	cdPipelineLevelConfigMaps := make(map[string]bool)
	cdPipelineLevelSecrets := make(map[string]bool)

	if workflowRequest.StageType == "PRE" {
		preStageConfigMapSecretsJson := workflowRequest.Pipeline.PreStageConfigMapSecretNames
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
	}
	if workflowRequest.StageType == "POST" {
		postStageConfigMapSecretsJson := workflowRequest.Pipeline.PostStageConfigMapSecretNames
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

func (workflowRequest *WorkflowRequest) GetExistingCmCsNamePrefix() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE:
		return strconv.Itoa(workflowRequest.WorkflowId) + "-" + bean.CI_WORKFLOW_NAME
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return strconv.Itoa(workflowRequest.WorkflowRunnerId) + "-" + strconv.Itoa(workflowRequest.WorkflowRunnerId)
	case bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return strconv.Itoa(workflowRequest.WorkflowId) + "-" + bean.CI_WORKFLOW_NAME
	default:
		return ""
	}
}

func (workflowRequest *WorkflowRequest) CheckForJob() bool {
	return workflowRequest.Type == bean.JOB_WORKFLOW_PIPELINE_TYPE
}

func (workflowRequest *WorkflowRequest) GetNodeConstraints(config *CiCdConfig) *bean.NodeConstraints {
	nodeLabel, err := GetNodeLabel(config, workflowRequest.Type, workflowRequest.IsExtRun)
	if err != nil {
		return nil
	}
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return &bean.NodeConstraints{
			ServiceAccount:   config.CiWorkflowServiceAccount,
			TaintKey:         config.CiTaintKey,
			TaintValue:       config.CiTaintValue,
			NodeLabel:        nodeLabel,
			SkipNodeSelector: workflowRequest.IsExtRun,
		}
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return &bean.NodeConstraints{
			ServiceAccount:   config.CdWorkflowServiceAccount,
			TaintKey:         config.CdTaintKey,
			TaintValue:       config.CdTaintValue,
			NodeLabel:        nodeLabel,
			SkipNodeSelector: false,
		}
	default:
		return nil
	}
}

func (workflowRequest *WorkflowRequest) GetLimitReqCpuMem(config *CiCdConfig) v1.ResourceRequirements {
	limitReqCpuMem := &bean.LimitReqCpuMem{}
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		limitReqCpuMem = &bean.LimitReqCpuMem{
			LimitCpu: config.CiLimitCpu,
			LimitMem: config.CiLimitMem,
			ReqCpu:   config.CiReqCpu,
			ReqMem:   config.CiReqMem,
		}
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		limitReqCpuMem = &bean.LimitReqCpuMem{
			LimitCpu: config.CdLimitCpu,
			LimitMem: config.CdLimitMem,
			ReqCpu:   config.CdReqCpu,
			ReqMem:   config.CdReqMem,
		}
	}
	return v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(limitReqCpuMem.LimitCpu),
			v1.ResourceMemory: resource.MustParse(limitReqCpuMem.LimitMem),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse(limitReqCpuMem.ReqCpu),
			v1.ResourceMemory: resource.MustParse(limitReqCpuMem.ReqMem),
		},
	}
}

func (workflowRequest *WorkflowRequest) getWorkflowImage() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return workflowRequest.CiImage
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return workflowRequest.CdImage
	default:
		return ""
	}
}

func (workflowRequest *WorkflowRequest) GetWorkflowMainContainer(config *CiCdConfig, workflowJson []byte, workflowTemplate *bean.WorkflowTemplate, workflowConfigMaps []bean3.ConfigSecretMap, workflowSecrets []bean3.ConfigSecretMap) (v1.Container, error) {
	privileged := true
	pvc := workflowRequest.getPVCForWorkflowRequest()
	containerEnvVariables := workflowRequest.getContainerEnvVariables(config, workflowJson)
	workflowMainContainer := v1.Container{
		Env:   containerEnvVariables,
		Name:  common.MainContainerName,
		Image: workflowRequest.getWorkflowImage(),
		SecurityContext: &v1.SecurityContext{
			Privileged: &privileged,
		},
		Resources: workflowRequest.GetLimitReqCpuMem(config),
	}
	if workflowRequest.Type == bean.CI_WORKFLOW_PIPELINE_TYPE || workflowRequest.Type == bean.JOB_WORKFLOW_PIPELINE_TYPE {
		workflowMainContainer.Ports = []v1.ContainerPort{{
			//exposed for user specific data from ci container
			Name:          "app-data",
			ContainerPort: 9102,
		}}
		err := workflowRequest.updateVolumeMountsForCi(config, workflowTemplate, &workflowMainContainer)
		if err != nil {
			return workflowMainContainer, err
		}
	}

	if len(pvc) != 0 {
		buildPvcCachePath := config.BuildPvcCachePath
		buildxPvcCachePath := config.BuildxPvcCachePath
		defaultPvcCachePath := config.DefaultPvcCachePath

		workflowTemplate.Volumes = append(workflowTemplate.Volumes, v1.Volume{
			Name: "root-vol",
			VolumeSource: v1.VolumeSource{
				PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
					ClaimName: pvc,
					ReadOnly:  false,
				},
			},
		})
		workflowMainContainer.VolumeMounts = append(workflowMainContainer.VolumeMounts,
			v1.VolumeMount{
				Name:      "root-vol",
				MountPath: buildPvcCachePath,
			},
			v1.VolumeMount{
				Name:      "root-vol",
				MountPath: buildxPvcCachePath,
			},
			v1.VolumeMount{
				Name:      "root-vol",
				MountPath: defaultPvcCachePath,
			})
	}
	UpdateContainerEnvsFromCmCs(&workflowMainContainer, workflowConfigMaps, workflowSecrets)
	return workflowMainContainer, nil
}

func (workflowRequest *WorkflowRequest) updateVolumeMountsForCi(config *CiCdConfig, workflowTemplate *bean.WorkflowTemplate, workflowMainContainer *v1.Container) error {
	volume, volumeMounts, err := config.GetWorkflowVolumeAndVolumeMounts()
	if err != nil {
		return err
	}
	workflowTemplate.Volumes = volume
	workflowMainContainer.VolumeMounts = volumeMounts
	return nil
}

func UpdateContainerEnvsFromCmCs(workflowMainContainer *v1.Container, configMaps []bean3.ConfigSecretMap, secrets []bean3.ConfigSecretMap) {
	for _, configMap := range configMaps {
		updateContainerEnvs(true, workflowMainContainer, configMap)
	}

	for _, secret := range secrets {
		updateContainerEnvs(false, workflowMainContainer, secret)
	}
}

func updateContainerEnvs(isCM bool, workflowMainContainer *v1.Container, configSecretMap bean3.ConfigSecretMap) {
	if configSecretMap.Type == repository2.VOLUME_CONFIG {
		workflowMainContainer.VolumeMounts = append(workflowMainContainer.VolumeMounts, v1.VolumeMount{
			Name:      configSecretMap.Name + "-vol",
			MountPath: configSecretMap.MountPath,
		})
	} else if configSecretMap.Type == repository2.ENVIRONMENT_CONFIG {
		var envFrom v1.EnvFromSource
		if isCM {
			envFrom = v1.EnvFromSource{
				ConfigMapRef: &v1.ConfigMapEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: configSecretMap.Name,
					},
				},
			}
		} else {
			envFrom = v1.EnvFromSource{
				SecretRef: &v1.SecretEnvSource{
					LocalObjectReference: v1.LocalObjectReference{
						Name: configSecretMap.Name,
					},
				},
			}
		}
		workflowMainContainer.EnvFrom = append(workflowMainContainer.EnvFrom, envFrom)
	}
}

const PRE = "PRE"

const POST = "POST"

const CI_NODE_PVC_ALL_ENV = "devtron.ai/ci-pvc-all"

const CI_NODE_PVC_PIPELINE_PREFIX = "devtron.ai/ci-pvc"

type CiArtifactDTO struct {
	Id                   int    `json:"id"`
	PipelineId           int    `json:"pipelineId"` //id of the ci pipeline from which this webhook was triggered
	Image                string `json:"image"`
	ImageDigest          string `json:"imageDigest"`
	MaterialInfo         string `json:"materialInfo"` //git material metadata json array string
	DataSource           string `json:"dataSource"`
	WorkflowId           *int   `json:"workflowId"`
	ciArtifactRepository repository2.CiArtifactRepository
}

type CiCdTriggerEvent struct {
	Type                  string           `json:"type"`
	CommonWorkflowRequest *WorkflowRequest `json:"commonWorkflowRequest"`
}

type GitMetadata struct {
	GitCommitHash  string `json:"GIT_COMMIT_HASH"`
	GitSourceType  string `json:"GIT_SOURCE_TYPE"`
	GitSourceValue string `json:"GIT_SOURCE_VALUE"`
}

type AppLabelMetadata struct {
	AppLabelKey   string `json:"APP_LABEL_KEY"`
	AppLabelValue string `json:"APP_LABEL_VALUE"`
}

type ChildCdMetadata struct {
	ChildCdEnvName     string `json:"CHILD_CD_ENV_NAME"`
	ChildCdClusterName string `json:"CHILD_CD_CLUSTER_NAME"`
}

type WorkflowResponse struct {
	Id                   int                                         `json:"id"`
	Name                 string                                      `json:"name"`
	Status               string                                      `json:"status"`
	PodStatus            string                                      `json:"podStatus"`
	Message              string                                      `json:"message"`
	StartedOn            time.Time                                   `json:"startedOn"`
	FinishedOn           time.Time                                   `json:"finishedOn"`
	CiPipelineId         int                                         `json:"ciPipelineId"`
	Namespace            string                                      `json:"namespace"`
	LogLocation          string                                      `json:"logLocation"`
	BlobStorageEnabled   bool                                        `json:"blobStorageEnabled"`
	GitTriggers          map[int]pipelineConfig.GitCommit            `json:"gitTriggers"`
	CiMaterials          []pipelineConfig.CiPipelineMaterialResponse `json:"ciMaterials"`
	TriggeredBy          int32                                       `json:"triggeredBy"`
	Artifact             string                                      `json:"artifact"`
	TriggeredByEmail     string                                      `json:"triggeredByEmail"`
	Stage                string                                      `json:"stage"`
	ArtifactId           int                                         `json:"artifactId"`
	IsArtifactUploaded   bool                                        `json:"isArtifactUploaded"`
	IsVirtualEnvironment bool                                        `json:"isVirtualEnvironment"`
	PodName              string                                      `json:"podName"`
	EnvironmentId        int                                         `json:"environmentId"`
	EnvironmentName      string                                      `json:"environmentName"`
	ImageReleaseTags     []*repository3.ImageTag                     `json:"imageReleaseTags"`
	ImageComment         *repository3.ImageComment                   `json:"imageComment"`
	AppWorkflowId        int                                         `json:"appWorkflowId"`
	CustomTag            *bean3.CustomTagErrorResponse               `json:"customTag,omitempty"`
	PipelineType         string                                      `json:"pipelineType"`
	ReferenceWorkflowId  int                                         `json:"referenceWorkflowId"`
}

type ConfigMapSecretDto struct {
	Name     string
	Data     map[string]string
	OwnerRef v12.OwnerReference
}

type WorkflowStatus struct {
	WorkflowName, Status, PodStatus, Message, LogLocation, PodName string
}
