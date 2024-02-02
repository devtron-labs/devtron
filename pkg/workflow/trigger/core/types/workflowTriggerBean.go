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
	"context"
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
	repository4 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/plugin"
	repository5 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/workflow/pipeline/ci/materials/types"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
	"strings"
	"time"
)

type WorkflowContext struct {
	context.Context
	// add app metadata
	// add pipeline metadata
	// add cluster metadata
	// add env metadata
}

type WorkflowTriggerRequest struct {
	WorkflowNamePrefix string `json:"workflowNamePrefix"`
	PipelineName       string `json:"pipelineName"`
	PipelineId         int    `json:"pipelineId"`
	AppId              int    `json:"appId"`
	EnvironmentId      int    `json:"environmentId"`
	OrchestratorHost   string `json:"orchestratorHost"`
	OrchestratorToken  string `json:"orchestratorToken"`
	AppName            string `json:"appName"`
	TriggerByAuthor    string `json:"triggerByAuthor"`
	WorkflowRunnerId   int    `json:"workflowRunnerId"`
	CdPipelineId       int    `json:"cdPipelineId"`
	CiPipelineType     string `json:"ciPipelineType"`
	// docker registry attributes start
	DockerImageTag     string `json:"dockerImageTag"`
	DockerRegistryId   string `json:"dockerRegistryId"`
	DockerRegistryType string `json:"dockerRegistryType"`
	DockerRegistryURL  string `json:"dockerRegistryURL"`
	DockerConnection   string `json:"dockerConnection"`
	DockerCert         string `json:"dockerCert"`
	DockerRepository   string `json:"dockerRepository"`
	DockerUsername     string `json:"dockerUsername"`
	DockerPassword     string `json:"dockerPassword"`
	AwsRegion          string `json:"awsRegion"`
	AccessKey          string `json:"accessKey"`
	SecretKey          string `json:"secretKey"`
	// docker registry attributes end
	CheckoutPath string `json:"checkoutPath"`
	// Blob storage start
	CiCacheLocation       string                            `json:"ciCacheLocation"`
	CiCacheRegion         string                            `json:"ciCacheRegion"`
	CiCacheFileName       string                            `json:"ciCacheFileName"`
	CloudProvider         blob_storage.BlobStorageType      `json:"cloudProvider"`
	BlobStorageConfigured bool                              `json:"blobStorageConfigured"`
	BlobStorageS3Config   *blob_storage.BlobStorageS3Config `json:"blobStorageS3Config"`
	AzureBlobConfig       *blob_storage.AzureBlobConfig     `json:"azureBlobConfig"`
	GcpBlobConfig         *blob_storage.GcpBlobConfig       `json:"gcpBlobConfig"`
	BlobStorageLogsKey    string                            `json:"blobStorageLogsKey"`
	CiArtifactLocation    string                            `json:"ciArtifactLocation"`
	CiArtifactBucket      string                            `json:"ciArtifactBucket"`
	CiArtifactFileName    string                            `json:"ciArtifactFileName"`
	CiArtifactRegion      string                            `json:"ciArtifactRegion"`
	CdCacheLocation       string                            `json:"cdCacheLocation"`
	CdCacheRegion         string                            `json:"cdCacheRegion"`
	// Blob storage end
	CiProjectDetails           []bean.CiProjectDetails `json:"ciProjectDetails"`
	ContainerResources         bean.ContainerResources `json:"containerResources"`
	ActiveDeadlineSeconds      int64                   `json:"activeDeadlineSeconds"`
	CiImage                    string                  `json:"ciImage"`
	Namespace                  string                  `json:"namespace"`
	WorkflowId                 int                     `json:"workflowId"`
	TriggeredBy                int32                   `json:"triggeredBy"`
	CacheLimit                 int64                   `json:"cacheLimit"`
	BeforeDockerBuildScripts   []*bean2.CiScript       `json:"beforeDockerBuildScripts"`
	AfterDockerBuildScripts    []*bean2.CiScript       `json:"afterDockerBuildScripts"`
	ScanEnabled                bool                    `json:"scanEnabled"`
	InAppLoggingEnabled        bool                    `json:"inAppLoggingEnabled"`
	DefaultAddressPoolBaseCidr string                  `json:"defaultAddressPoolBaseCidr"`
	DefaultAddressPoolSize     int                     `json:"defaultAddressPoolSize"`
	PreCiSteps                 []*bean.StepObject      `json:"preCiSteps"`
	PostCiSteps                []*bean.StepObject      `json:"postCiSteps"`
	RefPlugins                 []*bean.RefPluginObject `json:"refPlugins"`
	CiBuildConfig              *bean.CiBuildConfigBean `json:"ciBuildConfig"`
	CiBuildDockerMtuValue      int                     `json:"ciBuildDockerMtuValue"`
	IgnoreDockerCachePush      bool                    `json:"ignoreDockerCachePush"`
	IgnoreDockerCachePull      bool                    `json:"ignoreDockerCachePull"`
	CacheInvalidate            bool                    `json:"cacheInvalidate"`
	IsPvcMounted               bool                    `json:"IsPvcMounted"`
	ExtraEnvironmentVariables  map[string]string       `json:"extraEnvironmentVariables"`
	EnableBuildContext         bool                    `json:"enableBuildContext"`
	IsExtRun                   bool                    `json:"isExtRun"`
	ImageRetryCount            int                     `json:"imageRetryCount"`
	ImageRetryInterval         int                     `json:"imageRetryInterval"`
	// Data from CD Workflow service
	StageYaml                   string                                `json:"stageYaml"`
	ArtifactLocation            string                                `json:"artifactLocation"`
	CiArtifactDTO               CiArtifactDTO                         `json:"ciArtifactDTO"`
	CdImage                     string                                `json:"cdImage"`
	StageType                   string                                `json:"stageType"`
	WorkflowPrefixForLog        string                                `json:"workflowPrefixForLog"`
	DeploymentTriggeredBy       string                                `json:"deploymentTriggeredBy,omitempty"`
	DeploymentTriggerTime       time.Time                             `json:"deploymentTriggerTime,omitempty"`
	DeploymentReleaseCounter    int                                   `json:"deploymentReleaseCounter,omitempty"`
	WorkflowExecutor            pipelineConfig.WorkflowExecutorType   `json:"workflowExecutor"`
	PrePostDeploySteps          []*bean.StepObject                    `json:"prePostDeploySteps"`
	CiArtifactLastFetch         time.Time                             `json:"ciArtifactLastFetch"`
	UseExternalClusterBlob      bool                                  `json:"useExternalClusterBlob"`
	RegistryDestinationImageMap map[string][]string                   `json:"registryDestinationImageMap"`
	RegistryCredentialMap       map[string]plugin.RegistryCredentials `json:"registryCredentialMap"`
	PluginArtifactStage         string                                `json:"pluginArtifactStage"`
	PushImageBeforePostCI       bool                                  `json:"pushImageBeforePostCI"`
	Type                        bean.WorkflowPipelineType
	Pipeline                    *pipelineConfig.Pipeline
	Env                         *repository.Environment // needed cluster metadata rather than env
	AppLabels                   map[string]string
	Scope                       resourceQualifiers.Scope
	ImagePathReservationIds     []int `json:"-"`
}

func (workflowRequest *WorkflowTriggerRequest) LoadFromUserContext(userContext bean2.UserContext) {
	workflowRequest.TriggeredBy = userContext.UserId
	workflowRequest.TriggerByAuthor = userContext.EmailId
}

func (workflowRequest *WorkflowTriggerRequest) updateExternalRunMetadata() {
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

func (workflowRequest *WorkflowTriggerRequest) CheckBlobStorageConfig(config *CiCdConfig) bool {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return config.UseBlobStorageConfigInCiWorkflow
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return config.UseBlobStorageConfigInCdWorkflow
	default:
		return false
	}

}

func (workflowRequest *WorkflowTriggerRequest) updateUseExternalClusterBlob(config *CiCdConfig) {
	workflowRequest.UseExternalClusterBlob = !workflowRequest.CheckBlobStorageConfig(config) && workflowRequest.IsExtRun
}

func (workflowRequest *WorkflowTriggerRequest) GetWorkflowTemplate(workflowJson []byte, config *CiCdConfig) bean.WorkflowTemplate {

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

func (workflowRequest *WorkflowTriggerRequest) checkConfigType(config *CiCdConfig) {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		config.Type = CiConfigType
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		config.Type = CdConfigType
	}
}

func (workflowRequest *WorkflowTriggerRequest) GetBlobStorageLogsKey(config *CiCdConfig) string {
	return fmt.Sprintf("%s/%s", config.GetDefaultBuildLogsKeyPrefix(), workflowRequest.WorkflowPrefixForLog)
}

func (workflowRequest *WorkflowTriggerRequest) GetWorkflowJson(config *CiCdConfig) ([]byte, error) {
	workflowRequest.updateBlobStorageLogsKey(config)
	workflowRequest.updateExternalRunMetadata()
	workflowRequest.updateUseExternalClusterBlob(config)
	workflowJson, err := workflowRequest.getWorkflowJson()
	if err != nil {
		return nil, err
	}
	return workflowJson, err
}

func (workflowRequest *WorkflowTriggerRequest) GetEventTypeForWorkflowRequest() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return bean.CiStage
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return bean.CdStage
	default:
		return ""
	}
}

func (workflowRequest *WorkflowTriggerRequest) GetWorkflowTypeForWorkflowRequest() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return bean.CI_WORKFLOW_NAME
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return bean.CD_WORKFLOW_NAME
	default:
		return ""
	}
}

func (workflowRequest *WorkflowTriggerRequest) getContainerEnvVariables(config *CiCdConfig, workflowJson []byte) (containerEnvVariables []v1.EnvVar) {
	containerEnvVariables = []v1.EnvVar{{Name: bean.IMAGE_SCANNER_ENDPOINT, Value: config.ImageScannerEndpoint}, {Name: "NATS_SERVER_HOST", Value: config.NatsServerHost}}
	eventEnv := v1.EnvVar{Name: "CI_CD_EVENT", Value: string(workflowJson)}
	inAppLoggingEnv := v1.EnvVar{Name: "IN_APP_LOGGING", Value: strconv.FormatBool(workflowRequest.InAppLoggingEnabled)}
	containerEnvVariables = append(containerEnvVariables, eventEnv, inAppLoggingEnv)
	return containerEnvVariables
}

func (workflowRequest *WorkflowTriggerRequest) getPVCForWorkflowRequest() string {
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
		// pvc not supported for other then ci and job currently
	}
	return pvc
}

func (workflowRequest *WorkflowTriggerRequest) getDefaultBuildLogsKeyPrefix(config *CiCdConfig) string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return config.CiDefaultBuildLogsKeyPrefix
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return config.CdDefaultBuildLogsKeyPrefix
	default:
		return ""
	}
}

func (workflowRequest *WorkflowTriggerRequest) getBlobStorageLogsPrefix() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return workflowRequest.WorkflowNamePrefix
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return workflowRequest.WorkflowPrefixForLog
	default:
		return ""
	}
}

func (workflowRequest *WorkflowTriggerRequest) updateBlobStorageLogsKey(config *CiCdConfig) {
	workflowRequest.BlobStorageLogsKey = fmt.Sprintf("%s/%s", workflowRequest.getDefaultBuildLogsKeyPrefix(config), workflowRequest.getBlobStorageLogsPrefix())
	workflowRequest.InAppLoggingEnabled = config.InAppLoggingEnabled || (workflowRequest.WorkflowExecutor == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM)
}

func (workflowRequest *WorkflowTriggerRequest) getWorkflowJson() ([]byte, error) {
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

func (workflowRequest *WorkflowTriggerRequest) AddNodeConstraintsFromConfig(workflowTemplate *bean.WorkflowTemplate, config *CiCdConfig) {
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

func (workflowRequest *WorkflowTriggerRequest) GetGlobalCmCsNamePrefix() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return strconv.Itoa(workflowRequest.WorkflowId) + "-" + bean.CI_WORKFLOW_NAME
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return strconv.Itoa(workflowRequest.WorkflowRunnerId) + "-" + bean.CD_WORKFLOW_NAME
	default:
		return ""
	}
}

func (workflowRequest *WorkflowTriggerRequest) GetConfiguredCmCs() (map[string]bool, map[string]bool, error) {

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

func (workflowRequest *WorkflowTriggerRequest) GetExistingCmCsNamePrefix() string {
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

func (workflowRequest *WorkflowTriggerRequest) CheckForJob() bool {
	return workflowRequest.Type == bean.JOB_WORKFLOW_PIPELINE_TYPE
}

func (workflowRequest *WorkflowTriggerRequest) GetNodeConstraints(config *CiCdConfig) *bean.NodeConstraints {
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

func (workflowRequest *WorkflowTriggerRequest) GetLimitReqCpuMem(config *CiCdConfig) v1.ResourceRequirements {
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

func (workflowRequest *WorkflowTriggerRequest) getWorkflowImage() string {
	switch workflowRequest.Type {
	case bean.CI_WORKFLOW_PIPELINE_TYPE, bean.JOB_WORKFLOW_PIPELINE_TYPE:
		return workflowRequest.CiImage
	case bean.CD_WORKFLOW_PIPELINE_TYPE:
		return workflowRequest.CdImage
	default:
		return ""
	}
}

func (workflowRequest *WorkflowTriggerRequest) GetWorkflowMainContainer(config *CiCdConfig, workflowJson []byte, workflowTemplate *bean.WorkflowTemplate, workflowConfigMaps []bean3.ConfigSecretMap, workflowSecrets []bean3.ConfigSecretMap) (v1.Container, error) {
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
			// exposed for user specific data from ci container
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

func (workflowRequest *WorkflowTriggerRequest) updateVolumeMountsForCi(config *CiCdConfig, workflowTemplate *bean.WorkflowTemplate, workflowMainContainer *v1.Container) error {
	volume, volumeMounts, err := config.GetWorkflowVolumeAndVolumeMounts()
	if err != nil {
		return err
	}
	workflowTemplate.Volumes = volume
	workflowMainContainer.VolumeMounts = volumeMounts
	return nil
}

func (workflowRequest *WorkflowTriggerRequest) UpdateProjectMaterials(commitHashes map[int]pipelineConfig.GitCommit, materialModels []*types.CiPipelineMaterialModel) {

	var ciProjectDetails []bean.CiProjectDetails
	for _, ciMaterial := range materialModels {
		// ignore those materials which have inactive git material
		if ciMaterial == nil || ciMaterial.GitMaterial == nil || !ciMaterial.GitMaterial.Active {
			continue
		}
		commitHashForPipelineId := commitHashes[ciMaterial.Id]
		ciProjectDetail := bean.CiProjectDetails{
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
	workflowRequest.CiProjectDetails = ciProjectDetails
}

func (workflowRequest *WorkflowTriggerRequest) UpdatePipelineScripts(pipelineStepsData *bean.PrePostAndRefPluginStepsResponse, beforeScripts []*bean2.CiScript, afterScripts []*bean2.CiScript) {

	preCiSteps := pipelineStepsData.PreStageSteps
	postCiSteps := pipelineStepsData.PostStageSteps
	refPluginsData := pipelineStepsData.RefPluginData
	if !(len(beforeScripts) == 0 && len(afterScripts) == 0) {
		//found beforeDockerBuildScripts/afterDockerBuildScripts
		//building preCiSteps & postCiSteps from them, refPluginsData not needed
		preCiSteps = buildCiStepsDataFromDockerBuildScripts(beforeScripts)
		postCiSteps = buildCiStepsDataFromDockerBuildScripts(afterScripts)
		refPluginsData = []*bean.RefPluginObject{}
	}
	workflowRequest.PreCiSteps = preCiSteps
	workflowRequest.PostCiSteps = postCiSteps
	workflowRequest.RefPlugins = refPluginsData
}

func (workflowRequest *WorkflowTriggerRequest) UpdateFromEnvAndWorkflowConfig(uniqueId int, ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, ciCdConfig CiCdConfig) error {

	if ciWorkflowConfig.CiCacheBucket == "" {
		ciWorkflowConfig.CiCacheBucket = ciCdConfig.DefaultCacheBucket
	}
	if ciWorkflowConfig.CiCacheRegion == "" {
		ciWorkflowConfig.CiCacheRegion = ciCdConfig.DefaultCacheBucketRegion
	}
	if ciWorkflowConfig.CiImage == "" {
		ciWorkflowConfig.CiImage = ciCdConfig.GetDefaultImage()
	}
	if ciWorkflowConfig.CiTimeout == 0 {
		ciWorkflowConfig.CiTimeout = ciCdConfig.GetDefaultTimeout()
	}
	if ciWorkflowConfig.LogsBucket == "" {
		ciWorkflowConfig.LogsBucket = ciCdConfig.GetDefaultBuildLogsBucket()
	}

	workflowRequest.Namespace = ciWorkflowConfig.Namespace
	workflowRequest.CiImage = ciWorkflowConfig.CiImage
	workflowRequest.ActiveDeadlineSeconds = ciWorkflowConfig.CiTimeout
	workflowRequest.CiBuildDockerMtuValue = ciCdConfig.CiRunnerDockerMTUValue
	workflowRequest.IgnoreDockerCachePush = ciCdConfig.IgnoreDockerCacheForCI
	workflowRequest.IgnoreDockerCachePull = ciCdConfig.IgnoreDockerCacheForCI
	workflowRequest.CacheLimit = ciCdConfig.CacheLimit

	switch ciCdConfig.CloudProvider {
	case BLOB_STORAGE_S3:
		//No AccessKey is used for uploading artifacts, instead IAM based auth is used
		workflowRequest.CiCacheRegion = ciWorkflowConfig.CiCacheRegion
		workflowRequest.CiCacheLocation = ciWorkflowConfig.CiCacheBucket
		workflowRequest.CiArtifactLocation, workflowRequest.CiArtifactBucket, workflowRequest.CiArtifactFileName = workflowRequest.buildS3ArtifactLocation(ciWorkflowConfig, ciCdConfig, uniqueId)
		workflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			AccessKey:                  ciCdConfig.BlobStorageS3AccessKey,
			Passkey:                    ciCdConfig.BlobStorageS3SecretKey,
			EndpointUrl:                ciCdConfig.BlobStorageS3Endpoint,
			IsInSecure:                 ciCdConfig.BlobStorageS3EndpointInsecure,
			CiCacheBucketName:          ciWorkflowConfig.CiCacheBucket,
			CiCacheRegion:              ciWorkflowConfig.CiCacheRegion,
			CiCacheBucketVersioning:    ciCdConfig.BlobStorageS3BucketVersioned,
			CiArtifactBucketName:       workflowRequest.CiArtifactBucket,
			CiArtifactRegion:           ciCdConfig.GetDefaultCdLogsBucketRegion(),
			CiArtifactBucketVersioning: ciCdConfig.BlobStorageS3BucketVersioned,
			CiLogBucketName:            ciCdConfig.GetDefaultBuildLogsBucket(),
			CiLogRegion:                ciCdConfig.GetDefaultCdLogsBucketRegion(),
			CiLogBucketVersioning:      ciCdConfig.BlobStorageS3BucketVersioned,
		}
	case BLOB_STORAGE_GCP:
		workflowRequest.GcpBlobConfig = &blob_storage.GcpBlobConfig{
			CredentialFileJsonData: ciCdConfig.BlobStorageGcpCredentialJson,
			CacheBucketName:        ciWorkflowConfig.CiCacheBucket,
			LogBucketName:          ciWorkflowConfig.LogsBucket,
			ArtifactBucketName:     ciWorkflowConfig.LogsBucket,
		}
		workflowRequest.CiArtifactLocation = workflowRequest.buildDefaultArtifactLocation(ciWorkflowConfig, ciCdConfig, uniqueId)
		workflowRequest.CiArtifactFileName = workflowRequest.CiArtifactLocation
	case BLOB_STORAGE_AZURE:
		workflowRequest.AzureBlobConfig = &blob_storage.AzureBlobConfig{
			Enabled:               ciCdConfig.CloudProvider == BLOB_STORAGE_AZURE,
			AccountName:           ciCdConfig.AzureAccountName,
			BlobContainerCiCache:  ciCdConfig.AzureBlobContainerCiCache,
			AccountKey:            ciCdConfig.AzureAccountKey,
			BlobContainerCiLog:    ciCdConfig.AzureBlobContainerCiLog,
			BlobContainerArtifact: ciCdConfig.AzureBlobContainerCiLog,
		}
		workflowRequest.BlobStorageS3Config = &blob_storage.BlobStorageS3Config{
			EndpointUrl:           ciCdConfig.AzureGatewayUrl,
			IsInSecure:            ciCdConfig.AzureGatewayConnectionInsecure,
			CiLogBucketName:       ciCdConfig.AzureBlobContainerCiLog,
			CiLogRegion:           ciCdConfig.DefaultCacheBucketRegion,
			CiLogBucketVersioning: ciCdConfig.BlobStorageS3BucketVersioned,
			AccessKey:             ciCdConfig.AzureAccountName,
		}
		workflowRequest.CiArtifactLocation = workflowRequest.buildDefaultArtifactLocation(ciWorkflowConfig, ciCdConfig, uniqueId)
		workflowRequest.CiArtifactFileName = workflowRequest.CiArtifactLocation
	default:
		if ciCdConfig.BlobStorageEnabled {
			return fmt.Errorf("blob storage %s not supported", workflowRequest.CloudProvider)
		}
	}
	return nil
}

func (workflowRequest *WorkflowTriggerRequest) buildS3ArtifactLocation(ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, config CiCdConfig, uniqueId int) (string, string, string) {
	ciArtifactLocationFormat := ciWorkflowConfig.CiArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = config.GetArtifactLocationFormat()
	}
	ArtifactLocation := fmt.Sprintf("s3://%s/%s/"+ciArtifactLocationFormat, ciWorkflowConfig.LogsBucket, config.GetDefaultArtifactKeyPrefix(), uniqueId, uniqueId)
	artifactFileName := fmt.Sprintf(config.GetDefaultArtifactKeyPrefix()+"/"+ciArtifactLocationFormat, uniqueId, uniqueId)
	return ArtifactLocation, ciWorkflowConfig.LogsBucket, artifactFileName
}

func (workflowRequest *WorkflowTriggerRequest) buildDefaultArtifactLocation(ciWorkflowConfig *pipelineConfig.CiWorkflowConfig, config CiCdConfig, uniqueId int) string {
	ciArtifactLocationFormat := ciWorkflowConfig.CiArtifactLocationFormat
	if ciArtifactLocationFormat == "" {
		ciArtifactLocationFormat = config.GetArtifactLocationFormat()
	}
	ArtifactLocation := fmt.Sprintf("%s/"+ciArtifactLocationFormat, config.GetDefaultArtifactKeyPrefix(), uniqueId, uniqueId)
	return ArtifactLocation
}

func buildCiStepsDataFromDockerBuildScripts(dockerBuildScripts []*bean2.CiScript) []*bean.StepObject {
	//before plugin support, few variables were set as env vars in ci-runner
	//these variables are now moved to global vars in plugin steps, but to avoid error in old scripts adding those variables in payload
	inputVars := []*bean.VariableObject{
		{
			Name:                  "DOCKER_IMAGE_TAG",
			Format:                "STRING",
			VariableType:          bean.VARIABLE_TYPE_REF_GLOBAL,
			ReferenceVariableName: "DOCKER_IMAGE_TAG",
		},
		{
			Name:                  "DOCKER_REPOSITORY",
			Format:                "STRING",
			VariableType:          bean.VARIABLE_TYPE_REF_GLOBAL,
			ReferenceVariableName: "DOCKER_REPOSITORY",
		},
		{
			Name:                  "DOCKER_REGISTRY_URL",
			Format:                "STRING",
			VariableType:          bean.VARIABLE_TYPE_REF_GLOBAL,
			ReferenceVariableName: "DOCKER_REGISTRY_URL",
		},
		{
			Name:                  "DOCKER_IMAGE",
			Format:                "STRING",
			VariableType:          bean.VARIABLE_TYPE_REF_GLOBAL,
			ReferenceVariableName: "DOCKER_IMAGE",
		},
	}
	var ciSteps []*bean.StepObject
	for _, dockerBuildScript := range dockerBuildScripts {
		ciStep := &bean.StepObject{
			Name:          dockerBuildScript.Name,
			Index:         dockerBuildScript.Index,
			Script:        dockerBuildScript.Script,
			ArtifactPaths: []string{dockerBuildScript.OutputLocation},
			StepType:      string(repository4.PIPELINE_STEP_TYPE_INLINE), //TODO Fix this import from repository
			ExecutorType:  string(repository5.SCRIPT_TYPE_SHELL),         //TODO Fix this import from repository
			InputVars:     inputVars,
		}
		ciSteps = append(ciSteps, ciStep)
	}
	return ciSteps
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
	PipelineId           int    `json:"pipelineId"` // id of the ci pipeline from which this webhook was triggered
	Image                string `json:"image"`
	ImageDigest          string `json:"imageDigest"`
	MaterialInfo         string `json:"materialInfo"` // git material metadata json array string
	DataSource           string `json:"dataSource"`
	WorkflowId           *int   `json:"workflowId"`
	ciArtifactRepository repository2.CiArtifactRepository
}

type CiCdTriggerEvent struct {
	Type                  string                  `json:"type"`
	CommonWorkflowRequest *WorkflowTriggerRequest `json:"commonWorkflowRequest"`
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
