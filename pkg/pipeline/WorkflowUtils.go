package pipeline

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned"
	v1alpha12 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/argoproj/argo-workflows/v3/workflow/common"
	blob_storage "github.com/devtron-labs/common-lib-private/blob-storage"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean3 "github.com/devtron-labs/devtron/pkg/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/util"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"strconv"
	"strings"
	"time"
)

var ArgoWorkflowOwnerRef = v1.OwnerReference{APIVersion: "argoproj.io/v1alpha1", Kind: "Workflow", Name: "{{workflow.name}}", UID: "{{workflow.uid}}", BlockOwnerDeletion: &[]bool{true}[0]}

type ConfigMapSecretDto struct {
	Name     string
	Data     map[string]string
	OwnerRef v1.OwnerReference
}

func ExtractVolumesFromCmCs(configMaps []bean2.ConfigSecretMap, secrets []bean2.ConfigSecretMap) []v12.Volume {
	var volumes []v12.Volume
	configMapVolumes := extractVolumesFromConfigSecretMaps(true, configMaps)
	secretVolumes := extractVolumesFromConfigSecretMaps(false, secrets)

	for _, volume := range configMapVolumes {
		volumes = append(volumes, volume)
	}
	for _, volume := range secretVolumes {
		volumes = append(volumes, volume)
	}
	return volumes
}

func extractVolumesFromConfigSecretMaps(isCm bool, configSecretMaps []bean2.ConfigSecretMap) []v12.Volume {
	var volumes []v12.Volume
	for _, configSecretMap := range configSecretMaps {
		if configSecretMap.Type != util.ConfigMapSecretUsageTypeVolume {
			// not volume type so ignoring
			continue
		}
		var volumeSource v12.VolumeSource
		if isCm {
			volumeSource = v12.VolumeSource{
				ConfigMap: &v12.ConfigMapVolumeSource{
					LocalObjectReference: v12.LocalObjectReference{
						Name: configSecretMap.Name,
					},
				},
			}
		} else {
			volumeSource = v12.VolumeSource{
				Secret: &v12.SecretVolumeSource{
					SecretName: configSecretMap.Name,
				},
			}
		}
		volumes = append(volumes, v12.Volume{
			Name:         configSecretMap.Name + "-vol",
			VolumeSource: volumeSource,
		})
	}
	return volumes
}

func UpdateContainerEnvsFromCmCs(workflowMainContainer *v12.Container, configMaps []bean2.ConfigSecretMap, secrets []bean2.ConfigSecretMap) {
	for _, configMap := range configMaps {
		updateContainerEnvs(true, workflowMainContainer, configMap)
	}

	for _, secret := range secrets {
		updateContainerEnvs(false, workflowMainContainer, secret)
	}
}

func updateContainerEnvs(isCM bool, workflowMainContainer *v12.Container, configSecretMap bean2.ConfigSecretMap) {
	if configSecretMap.Type == repository.VOLUME_CONFIG {
		workflowMainContainer.VolumeMounts = append(workflowMainContainer.VolumeMounts, v12.VolumeMount{
			Name:      configSecretMap.Name + "-vol",
			MountPath: configSecretMap.MountPath,
		})
	} else if configSecretMap.Type == repository.ENVIRONMENT_CONFIG {
		var envFrom v12.EnvFromSource
		if isCM {
			envFrom = v12.EnvFromSource{
				ConfigMapRef: &v12.ConfigMapEnvSource{
					LocalObjectReference: v12.LocalObjectReference{
						Name: configSecretMap.Name,
					},
				},
			}
		} else {
			envFrom = v12.EnvFromSource{
				SecretRef: &v12.SecretEnvSource{
					LocalObjectReference: v12.LocalObjectReference{
						Name: configSecretMap.Name,
					},
				},
			}
		}
		workflowMainContainer.EnvFrom = append(workflowMainContainer.EnvFrom, envFrom)
	}
}

func GetConfigMapJson(configMapSecretDto ConfigMapSecretDto) (string, error) {
	configMapBody := GetConfigMapBody(configMapSecretDto)
	configMapJson, err := json.Marshal(configMapBody)
	if err != nil {
		return "", err
	}
	return string(configMapJson), err
}

func GetConfigMapBody(configMapSecretDto ConfigMapSecretDto) v12.ConfigMap {
	return v12.ConfigMap{
		TypeMeta: v1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            configMapSecretDto.Name,
			OwnerReferences: []v1.OwnerReference{configMapSecretDto.OwnerRef},
		},
		Data: configMapSecretDto.Data,
	}
}

func GetSecretJson(configMapSecretDto ConfigMapSecretDto) (string, error) {
	secretBody := GetSecretBody(configMapSecretDto)
	secretJson, err := json.Marshal(secretBody)
	if err != nil {
		return "", err
	}
	return string(secretJson), err
}

func GetSecretBody(configMapSecretDto ConfigMapSecretDto) v12.Secret {
	secretDataMap := make(map[string][]byte)

	// adding handling to get base64 decoded value in map value
	cmsDataMarshaled, _ := json.Marshal(configMapSecretDto.Data)
	json.Unmarshal(cmsDataMarshaled, &secretDataMap)

	return v12.Secret{
		TypeMeta: v1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:            configMapSecretDto.Name,
			OwnerReferences: []v1.OwnerReference{configMapSecretDto.OwnerRef},
		},
		Data: secretDataMap,
		Type: "Opaque",
	}
}

func GetFromGlobalCmCsDtos(globalCmCsConfigs []*bean.GlobalCMCSDto) ([]bean2.ConfigSecretMap, []bean2.ConfigSecretMap, error) {
	workflowConfigMaps := make([]bean2.ConfigSecretMap, 0, len(globalCmCsConfigs))
	workflowSecrets := make([]bean2.ConfigSecretMap, 0, len(globalCmCsConfigs))

	for _, config := range globalCmCsConfigs {
		configSecretMap, err := config.ConvertToConfigSecretMap()
		if err != nil {
			return workflowConfigMaps, workflowSecrets, err
		}
		if config.ConfigType == repository.CM_TYPE_CONFIG {
			workflowConfigMaps = append(workflowConfigMaps, configSecretMap)
		} else {
			workflowSecrets = append(workflowSecrets, configSecretMap)
		}
	}
	return workflowConfigMaps, workflowSecrets, nil
}

func AddTemplatesForGlobalSecretsInWorkflowTemplate(globalCmCsConfigs []*bean.GlobalCMCSDto, steps *[]v1alpha1.ParallelSteps, volumes *[]v12.Volume, templates *[]v1alpha1.Template) error {

	cmIndex := 0
	csIndex := 0
	for _, config := range globalCmCsConfigs {
		if config.ConfigType == repository.CM_TYPE_CONFIG {
			cmJson, err := GetConfigMapJson(ConfigMapSecretDto{Name: config.Name, Data: config.Data, OwnerRef: ArgoWorkflowOwnerRef})
			if err != nil {
				return err
			}
			if config.Type == repository.VOLUME_CONFIG {
				*volumes = append(*volumes, v12.Volume{
					Name: config.Name + "-vol",
					VolumeSource: v12.VolumeSource{
						ConfigMap: &v12.ConfigMapVolumeSource{
							LocalObjectReference: v12.LocalObjectReference{
								Name: config.Name,
							},
						},
					},
				})
			}
			*steps = append(*steps, v1alpha1.ParallelSteps{
				Steps: []v1alpha1.WorkflowStep{
					{
						Name:     "create-env-cm-gb-" + strconv.Itoa(cmIndex),
						Template: "cm-gb-" + strconv.Itoa(cmIndex),
					},
				},
			})
			*templates = append(*templates, v1alpha1.Template{
				Name: "cm-gb-" + strconv.Itoa(cmIndex),
				Resource: &v1alpha1.ResourceTemplate{
					Action:            "create",
					SetOwnerReference: true,
					Manifest:          string(cmJson),
				},
			})
			cmIndex++
		} else if config.ConfigType == repository.CS_TYPE_CONFIG {

			// special handling for secret data since GetSecretJson expects encoded values in data map
			encodedSecretData, err := bean.ConvertToEncodedForm(config.Data)
			if err != nil {
				return err
			}
			var encodedSecretDataMap = make(map[string]string)
			err = json.Unmarshal(encodedSecretData, &encodedSecretDataMap)
			if err != nil {
				return err
			}

			secretJson, err := GetSecretJson(ConfigMapSecretDto{Name: config.Name, Data: encodedSecretDataMap, OwnerRef: ArgoWorkflowOwnerRef})
			if err != nil {
				return err
			}
			if config.Type == repository.VOLUME_CONFIG {
				*volumes = append(*volumes, v12.Volume{
					Name: config.Name + "-vol",
					VolumeSource: v12.VolumeSource{
						Secret: &v12.SecretVolumeSource{
							SecretName: config.Name,
						},
					},
				})
			}
			*steps = append(*steps, v1alpha1.ParallelSteps{
				Steps: []v1alpha1.WorkflowStep{
					{
						Name:     "create-env-sec-gb-" + strconv.Itoa(csIndex),
						Template: "sec-gb-" + strconv.Itoa(csIndex),
					},
				},
			})
			*templates = append(*templates, v1alpha1.Template{
				Name: "sec-gb-" + strconv.Itoa(csIndex),
				Resource: &v1alpha1.ResourceTemplate{
					Action:            "create",
					SetOwnerReference: true,
					Manifest:          string(secretJson),
				},
			})
			csIndex++
		}
	}

	return nil
}

func IsShallowClonePossible(ciMaterial *pipelineConfig.CiPipelineMaterial, gitProviders, cloningMode string) bool {
	gitProvidersList := strings.Split(gitProviders, ",")
	for _, gitProvider := range gitProvidersList {
		if strings.Contains(strings.ToLower(ciMaterial.GitMaterial.Url), strings.ToLower(gitProvider)) && cloningMode == CloningModeShallow {
			return true
		}
	}
	return false
}

func GetClientInstance(config *rest.Config, namespace string) (v1alpha12.WorkflowInterface, error) {
	clientSet, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	wfClient := clientSet.ArgoprojV1alpha1().Workflows(namespace) // create the workflow client
	return wfClient, nil
}

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
	BeforeDockerBuildScripts   []*bean3.CiScript                 `json:"beforeDockerBuildScripts"`
	AfterDockerBuildScripts    []*bean3.CiScript                 `json:"afterDockerBuildScripts"`
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
	Type                     bean.WorkflowPipelineType
	Pipeline                 *pipelineConfig.Pipeline
	Env                      *repository2.Environment
	AppLabels                map[string]string
	IsDryRun                 bool `json:"isDryRun"`
}

type CiCdTriggerEvent struct {
	Type                  string           `json:"type"`
	CommonWorkflowRequest *WorkflowRequest `json:"commonWorkflowRequest"`
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

func (workflowRequest *WorkflowRequest) getContainerEnvVariables(config *CiCdConfig, workflowJson []byte) (containerEnvVariables []v12.EnvVar) {
	containerEnvVariables = []v12.EnvVar{{Name: "IMAGE_SCANNER_ENDPOINT", Value: config.ImageScannerEndpoint}}

	if config.CloudProvider == BLOB_STORAGE_S3 && config.BlobStorageS3AccessKey != "" {
		miniCred := []v12.EnvVar{{Name: "AWS_ACCESS_KEY_ID", Value: config.BlobStorageS3AccessKey}, {Name: "AWS_SECRET_ACCESS_KEY", Value: config.BlobStorageS3SecretKey}}
		containerEnvVariables = append(containerEnvVariables, miniCred...)
	}
	eventEnv := v12.EnvVar{Name: "CI_CD_EVENT", Value: string(workflowJson)}
	inAppLoggingEnv := v12.EnvVar{Name: "IN_APP_LOGGING", Value: strconv.FormatBool(workflowRequest.InAppLoggingEnabled)}
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
		workflowTemplate.Tolerations = []v12.Toleration{{Key: nodeConstraints.TaintKey, Value: nodeConstraints.TaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}}
	}
	// In the future, we will give support for NodeSelector for job currently we need to have a node without dedicated NodeLabel to run job
	if len(nodeConstraints.NodeLabel) > 0 && !(nodeConstraints.SkipNodeSelector) {
		workflowTemplate.NodeSelector = nodeConstraints.NodeLabel
	}
	workflowTemplate.ArchiveLogs = workflowRequest.BlobStorageConfigured && !workflowRequest.InAppLoggingEnabled
	workflowTemplate.RestartPolicy = v12.RestartPolicyNever

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
		preStageConfigmapSecrets := bean3.PreStageConfigMapSecretNames{}
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
		postStageConfigmapSecrets := bean3.PostStageConfigMapSecretNames{}
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
	nodeLabel, err := getNodeLabel(config, workflowRequest.Type, workflowRequest.IsExtRun)
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
			SkipNodeSelector: false,
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

func (workflowRequest *WorkflowRequest) GetLimitReqCpuMem(config *CiCdConfig) v12.ResourceRequirements {
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
	return v12.ResourceRequirements{
		Limits: v12.ResourceList{
			v12.ResourceCPU:    resource.MustParse(limitReqCpuMem.LimitCpu),
			v12.ResourceMemory: resource.MustParse(limitReqCpuMem.LimitMem),
		},
		Requests: v12.ResourceList{
			v12.ResourceCPU:    resource.MustParse(limitReqCpuMem.ReqCpu),
			v12.ResourceMemory: resource.MustParse(limitReqCpuMem.ReqMem),
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
func (workflowRequest *WorkflowRequest) GetWorkflowMainContainer(config *CiCdConfig, workflowJson []byte, workflowTemplate *bean.WorkflowTemplate, workflowConfigMaps []bean2.ConfigSecretMap, workflowSecrets []bean2.ConfigSecretMap) (v12.Container, error) {
	privileged := true
	pvc := workflowRequest.getPVCForWorkflowRequest()
	containerEnvVariables := workflowRequest.getContainerEnvVariables(config, workflowJson)
	workflowMainContainer := v12.Container{
		Env:   containerEnvVariables,
		Name:  common.MainContainerName,
		Image: workflowRequest.getWorkflowImage(),
		SecurityContext: &v12.SecurityContext{
			Privileged: &privileged,
		},
		Resources: workflowRequest.GetLimitReqCpuMem(config),
	}
	if workflowRequest.Type == bean.CI_WORKFLOW_PIPELINE_TYPE || workflowRequest.Type == bean.JOB_WORKFLOW_PIPELINE_TYPE {
		workflowMainContainer.Name = ""
		workflowMainContainer.Ports = []v12.ContainerPort{{
			//exposed for user specific data from ci container
			Name:          "app-data",
			ContainerPort: 9102,
		}}
		err := updateVolumeMountsForCi(config, workflowTemplate, &workflowMainContainer)
		if err != nil {
			return workflowMainContainer, err
		}
	}

	if len(pvc) != 0 {
		buildPvcCachePath := config.BuildPvcCachePath
		buildxPvcCachePath := config.BuildxPvcCachePath
		defaultPvcCachePath := config.DefaultPvcCachePath

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
	return workflowMainContainer, nil
}

func CheckIfReTriggerRequired(status, message, workflowRunnerStatus string) bool {
	return ((status == string(v1alpha1.NodeError) || status == string(v1alpha1.NodeFailed)) &&
		message == POD_DELETED_MESSAGE) && workflowRunnerStatus != WorkflowCancel

}

func updateVolumeMountsForCi(config *CiCdConfig, workflowTemplate *bean.WorkflowTemplate, workflowMainContainer *v12.Container) error {
	volume, volumeMounts, err := config.GetWorkflowVolumeAndVolumeMounts()
	if err != nil {
		return err
	}
	workflowTemplate.Volumes = volume
	workflowMainContainer.VolumeMounts = volumeMounts
	return nil
}
