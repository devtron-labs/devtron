// /*
// * Copyright (c) 2020 Devtron Labs
// *
// * Licensed under the Apache License, Version 2.0 (the "License");
// * you may not use this file except in compliance with the License.
// * You may obtain a copy of the License at
// *
// *    http://www.apache.org/licenses/LICENSE-2.0
// *
// * Unless required by applicable law or agreed to in writing, software
// * distributed under the License is distributed on an "AS IS" BASIS,
// * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// * See the License for the specific language governing permissions and
// * limitations under the License.
// *
// */
package pipeline

//
//import (
//	"context"
//	"encoding/json"
//	"fmt"
//	"github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
//	v1alpha12 "github.com/argoproj/argo-workflows/v3/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
//	"github.com/argoproj/argo-workflows/v3/workflow/util"
//	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
//	bean3 "github.com/devtron-labs/devtron/api/bean"
//	"github.com/devtron-labs/devtron/internal/sql/repository"
//	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
//	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
//	"github.com/devtron-labs/devtron/pkg/app"
//	"github.com/devtron-labs/devtron/pkg/bean"
//	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
//	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
//	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
//	"github.com/devtron-labs/devtron/util/k8s"
//	"go.uber.org/zap"
//	v12 "k8s.io/api/core/v1"
//	"k8s.io/apimachinery/pkg/api/resource"
//	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
//	"k8s.io/apimachinery/pkg/util/intstr"
//	"k8s.io/client-go/rest"
//	"net/url"
//	"strconv"
//	"strings"
//	"time"
//)
//
//type WorkflowService interface {
//	SubmitWorkflow(workflowRequest *WorkflowRequest, appLabels map[string]string, env *repository2.Environment, isJob bool) (*v1alpha1.Workflow, error)
//	DeleteWorkflow(wfName string, namespace string) error
//	GetWorkflow(name string, namespace string, isExt bool, environment *repository2.Environment) (*v1alpha1.Workflow, error)
//	ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error)
//	UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error)
//	TerminateWorkflow(name string, namespace string, isExt bool, environment *repository2.Environment) error
//	//TerminateWorkflowForCiCd(executorType pipelineConfig.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config, isExt bool, environment *repository2.Environment) error
//}
//
////type CiCdTriggerEvent struct {
////	Type      string             `json:"type"`
////	CiRequest *WorkflowRequest   `json:"ciRequest"`
////	CdRequest *CommonWorkflowRequest `json:"cdRequest"`
////}
//
//type WorkflowServiceImpl struct {
//	Logger                 *zap.SugaredLogger
//	config                 *rest.Config
//	ciConfig               *CiCdConfig
//	globalCMCSService      GlobalCMCSService
//	appService             app.AppService
//	configMapRepository    chartConfig.ConfigMapRepository
//	k8sUtil                *k8s.K8sUtil
//	k8sCommonService       k8s2.K8sCommonService
//	argoWorkflowExecutor   ArgoWorkflowExecutor
//	systemWorkflowExecutor SystemWorkflowExecutor
//}
//
//type WorkflowRequest struct {
//	WorkflowNamePrefix         string                            `json:"workflowNamePrefix"`
//	PipelineName               string                            `json:"pipelineName"`
//	PipelineId                 int                               `json:"pipelineId"`
//	DockerImageTag             string                            `json:"dockerImageTag"`
//	DockerRegistryId           string                            `json:"dockerRegistryId"`
//	DockerRegistryType         string                            `json:"dockerRegistryType"`
//	DockerRegistryURL          string                            `json:"dockerRegistryURL"`
//	DockerConnection           string                            `json:"dockerConnection"`
//	DockerCert                 string                            `json:"dockerCert"`
//	DockerRepository           string                            `json:"dockerRepository"`
//	CheckoutPath               string                            `json:"checkoutPath"`
//	DockerUsername             string                            `json:"dockerUsername"`
//	DockerPassword             string                            `json:"dockerPassword"`
//	AwsRegion                  string                            `json:"awsRegion"`
//	AccessKey                  string                            `json:"accessKey"`
//	SecretKey                  string                            `json:"secretKey"`
//	CiCacheLocation            string                            `json:"ciCacheLocation"`
//	CiCacheRegion              string                            `json:"ciCacheRegion"`
//	CiCacheFileName            string                            `json:"ciCacheFileName"`
//	CiProjectDetails           []CiProjectDetails                `json:"ciProjectDetails"`
//	ContainerResources         ContainerResources                `json:"containerResources"`
//	ActiveDeadlineSeconds      int64                             `json:"activeDeadlineSeconds"`
//	CiImage                    string                            `json:"ciImage"`
//	Namespace                  string                            `json:"namespace"`
//	WorkflowId                 int                               `json:"workflowId"`
//	TriggeredBy                int32                             `json:"triggeredBy"`
//	CacheLimit                 int64                             `json:"cacheLimit"`
//	BeforeDockerBuildScripts   []*bean.CiScript                  `json:"beforeDockerBuildScripts"`
//	AfterDockerBuildScripts    []*bean.CiScript                  `json:"afterDockerBuildScripts"`
//	CiArtifactLocation         string                            `json:"ciArtifactLocation"`
//	CiArtifactBucket           string                            `json:"ciArtifactBucket"`
//	CiArtifactFileName         string                            `json:"ciArtifactFileName"`
//	CiArtifactRegion           string                            `json:"ciArtifactRegion"`
//	ScanEnabled                bool                              `json:"scanEnabled"`
//	CloudProvider              blob_storage.BlobStorageType      `json:"cloudProvider"`
//	BlobStorageConfigured      bool                              `json:"blobStorageConfigured"`
//	BlobStorageS3Config        *blob_storage.BlobStorageS3Config `json:"blobStorageS3Config"`
//	AzureBlobConfig            *blob_storage.AzureBlobConfig     `json:"azureBlobConfig"`
//	GcpBlobConfig              *blob_storage.GcpBlobConfig       `json:"gcpBlobConfig"`
//	BlobStorageLogsKey         string                            `json:"blobStorageLogsKey"`
//	InAppLoggingEnabled        bool                              `json:"inAppLoggingEnabled"`
//	DefaultAddressPoolBaseCidr string                            `json:"defaultAddressPoolBaseCidr"`
//	DefaultAddressPoolSize     int                               `json:"defaultAddressPoolSize"`
//	PreCiSteps                 []*bean2.StepObject               `json:"preCiSteps"`
//	PostCiSteps                []*bean2.StepObject               `json:"postCiSteps"`
//	RefPlugins                 []*bean2.RefPluginObject          `json:"refPlugins"`
//	AppName                    string                            `json:"appName"`
//	TriggerByAuthor            string                            `json:"triggerByAuthor"`
//	CiBuildConfig              *bean2.CiBuildConfigBean          `json:"ciBuildConfig"`
//	CiBuildDockerMtuValue      int                               `json:"ciBuildDockerMtuValue"`
//	IgnoreDockerCachePush      bool                              `json:"ignoreDockerCachePush"`
//	IgnoreDockerCachePull      bool                              `json:"ignoreDockerCachePull"`
//	CacheInvalidate            bool                              `json:"cacheInvalidate"`
//	IsPvcMounted               bool                              `json:"IsPvcMounted"`
//	ExtraEnvironmentVariables  map[string]string                 `json:"extraEnvironmentVariables"`
//	EnableBuildContext         bool                              `json:"enableBuildContext"`
//	AppId                      int                               `json:"appId"`
//	EnvironmentId              int                               `json:"environmentId"`
//	OrchestratorHost           string                            `json:"orchestratorHost"`
//	OrchestratorToken          string                            `json:"orchestratorToken"`
//	IsExtRun                   bool                              `json:"isExtRun"`
//	ImageRetryCount            int                               `json:"imageRetryCount"`
//	ImageRetryInterval         int                               `json:"imageRetryInterval"`
//	// Data from CD Workflow service
//	WorkflowRunnerId         int                                 `json:"workflowRunnerId"`
//	CdPipelineId             int                                 `json:"cdPipelineId"`
//	StageYaml                string                              `json:"stageYaml"`
//	ArtifactLocation         string                              `json:"artifactLocation"`
//	CiArtifactDTO            CiArtifactDTO                       `json:"ciArtifactDTO"`
//	CdImage                  string                              `json:"cdImage"`
//	StageType                string                              `json:"stageType"`
//	CdCacheLocation          string                              `json:"cdCacheLocation"`
//	CdCacheRegion            string                              `json:"cdCacheRegion"`
//	WorkflowPrefixForLog     string                              `json:"workflowPrefixForLog"`
//	DeploymentTriggeredBy    string                              `json:"deploymentTriggeredBy,omitempty"`
//	DeploymentTriggerTime    time.Time                           `json:"deploymentTriggerTime,omitempty"`
//	DeploymentReleaseCounter int                                 `json:"deploymentReleaseCounter,omitempty"`
//	WorkflowExecutor         pipelineConfig.WorkflowExecutorType `json:"workflowExecutor"`
//	PrePostDeploySteps       []*bean2.StepObject                 `json:"prePostDeploySteps"`
//}
//
//const (
//	BLOB_STORAGE_AZURE             = "AZURE"
//	BLOB_STORAGE_S3                = "S3"
//	BLOB_STORAGE_GCP               = "GCP"
//	BLOB_STORAGE_MINIO             = "MINIO"
//	CI_WORKFLOW_NAME               = "ci"
//	CI_WORKFLOW_WITH_STAGES        = "ci-stages-with-env"
//	CI_NODE_SELECTOR_APP_LABEL_KEY = "devtron.ai/node-selector"
//	CI_NODE_PVC_ALL_ENV            = "devtron.ai/ci-pvc-all"
//	CI_NODE_PVC_PIPELINE_PREFIX    = "devtron.ai/ci-pvc"
//	CD_WORKFLOW_NAME               = "cd"
//	CD_WORKFLOW_WITH_STAGES        = "cd-stages-with-env"
//)
//
//type ContainerResources struct {
//	MinCpu        string `json:"minCpu"`
//	MaxCpu        string `json:"maxCpu"`
//	MinStorage    string `json:"minStorage"`
//	MaxStorage    string `json:"maxStorage"`
//	MinEphStorage string `json:"minEphStorage"`
//	MaxEphStorage string `json:"maxEphStorage"`
//	MinMem        string `json:"minMem"`
//	MaxMem        string `json:"maxMem"`
//}
//
//// Used for default values
///*func NewContainerResources() ContainerResources {
//	return ContainerResources{
//		MinCpu:        "",
//		MaxCpu:        "0.5",
//		MinStorage:    "",
//		MaxStorage:    "",
//		MinEphStorage: "",
//		MaxEphStorage: "",
//		MinMem:        "",
//		MaxMem:        "200Mi",
//	}
//}*/
//
//type CiProjectDetails struct {
//	GitRepository   string `json:"gitRepository"`
//	MaterialName    string `json:"materialName"`
//	CheckoutPath    string `json:"checkoutPath"`
//	FetchSubmodules bool   `json:"fetchSubmodules"`
//	CommitHash      string `json:"commitHash"`
//	GitTag          string `json:"gitTag"`
//	CommitTime      string `json:"commitTime"`
//	//Branch        string          `json:"branch"`
//	Type        string                    `json:"type"`
//	Message     string                    `json:"message"`
//	Author      string                    `json:"author"`
//	GitOptions  GitOptions                `json:"gitOptions"`
//	SourceType  pipelineConfig.SourceType `json:"sourceType"`
//	SourceValue string                    `json:"sourceValue"`
//	WebhookData pipelineConfig.WebhookData
//}
//
//type GitOptions struct {
//	UserName      string              `json:"userName"`
//	Password      string              `json:"password"`
//	SshPrivateKey string              `json:"sshPrivateKey"`
//	AccessToken   string              `json:"accessToken"`
//	AuthMode      repository.AuthMode `json:"authMode"`
//}
//
//func NewWorkflowServiceImpl(Logger *zap.SugaredLogger, ciConfig *CiCdConfig, globalCMCSService GlobalCMCSService,
//	appService app.AppService, configMapRepository chartConfig.ConfigMapRepository,
//	k8sUtil *k8s.K8sUtil, k8sCommonService k8s2.K8sCommonService, argoWorkflowExecutor ArgoWorkflowExecutor, systemWorkflowExecutor SystemWorkflowExecutor) (*WorkflowServiceImpl, error) {
//	workflowService := &WorkflowServiceImpl{
//		Logger:                 Logger,
//		ciConfig:               ciConfig,
//		globalCMCSService:      globalCMCSService,
//		appService:             appService,
//		configMapRepository:    configMapRepository,
//		k8sCommonService:       k8sCommonService,
//		argoWorkflowExecutor:   argoWorkflowExecutor,
//		systemWorkflowExecutor: systemWorkflowExecutor,
//	}
//	restConfig, err := k8sUtil.GetK8sInClusterRestConfig()
//	if err != nil {
//		Logger.Errorw("error in getting in cluster rest config", "err", err)
//		return nil, err
//	}
//	workflowService.config = restConfig
//	return workflowService, nil
//}
//
////
////const ciEvent = "CI"
////const cdStage = "CD"
//
//func (impl *WorkflowServiceImpl) SubmitWorkflow(workflowRequest *WorkflowRequest, appLabels map[string]string, env *repository2.Environment, isJob bool) (*v1alpha1.Workflow, error) {
//	containerEnvVariables := []v12.EnvVar{{Name: "IMAGE_SCANNER_ENDPOINT", Value: impl.ciConfig.ImageScannerEndpoint}}
//	if impl.ciConfig.CloudProvider == BLOB_STORAGE_S3 && impl.ciConfig.BlobStorageS3AccessKey != "" {
//		miniCred := []v12.EnvVar{{Name: "AWS_ACCESS_KEY_ID", Value: impl.ciConfig.BlobStorageS3AccessKey}, {Name: "AWS_SECRET_ACCESS_KEY", Value: impl.ciConfig.BlobStorageS3SecretKey}}
//		containerEnvVariables = append(containerEnvVariables, miniCred...)
//	}
//	pvc := appLabels[strings.ToLower(fmt.Sprintf("%s-%s", CI_NODE_PVC_PIPELINE_PREFIX, workflowRequest.PipelineName))]
//	if len(pvc) == 0 {
//		pvc = appLabels[CI_NODE_PVC_ALL_ENV]
//	}
//	if len(pvc) != 0 {
//		workflowRequest.IsPvcMounted = true
//		workflowRequest.IgnoreDockerCachePush = true
//		workflowRequest.IgnoreDockerCachePull = true
//	}
//	ciCdTriggerEvent := CiCdTriggerEvent{
//		Type:      ciEvent,
//		CiRequest: workflowRequest,
//	}
//	if env != nil && env.Id != 0 {
//		workflowRequest.IsExtRun = true
//	}
//	ciCdTriggerEvent.CiRequest.BlobStorageLogsKey = fmt.Sprintf("%s/%s", impl.ciConfig.CiDefaultBuildLogsKeyPrefix, workflowRequest.WorkflowNamePrefix)
//	ciCdTriggerEvent.CiRequest.InAppLoggingEnabled = impl.ciConfig.InAppLoggingEnabled
//	workflowJson, err := json.Marshal(&ciCdTriggerEvent)
//	if err != nil {
//		impl.Logger.Errorw("err", err)
//		return nil, err
//	}
//	impl.Logger.Debugw("workflowRequest ---->", "workflowJson", string(workflowJson))
//
//	wfClient, err := impl.getWfClient(env, workflowRequest.Namespace, workflowRequest.IsExtRun)
//
//	if err != nil {
//		return nil, err
//	}
//
//	privileged := true
//	blobStorageConfigured := workflowRequest.BlobStorageConfigured
//	archiveLogs := blobStorageConfigured && !impl.ciConfig.InAppLoggingEnabled
//
//	limitCpu := impl.ciConfig.CiLimitCpu
//	limitMem := impl.ciConfig.CiLimitMem
//
//	reqCpu := impl.ciConfig.CiReqCpu
//	reqMem := impl.ciConfig.CiReqMem
//	ttl := int32(impl.ciConfig.BuildLogTTLValue)
//
//	entryPoint := CI_WORKFLOW_NAME // template name from where worklow execution will start
//	//getting all cm/cs to be used by default
//	steps := make([]v1alpha1.ParallelSteps, 0)
//	volumes := make([]v12.Volume, 0)
//	templates := make([]v1alpha1.Template, 0)
//	var globalCmCsConfigs []*bean2.GlobalCMCSDto
//	if !workflowRequest.IsExtRun {
//		globalCmCsConfigs, err = impl.globalCMCSService.FindAllActiveByPipelineType(repository.PIPELINE_TYPE_CI)
//		if err != nil {
//			impl.Logger.Errorw("error in getting all global cm/cs config", "err", err)
//			return nil, err
//		}
//		if len(globalCmCsConfigs) > 0 {
//			entryPoint = CI_WORKFLOW_WITH_STAGES
//		}
//		for i := range globalCmCsConfigs {
//			globalCmCsConfigs[i].Name = strings.ToLower(globalCmCsConfigs[i].Name) + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + CI_WORKFLOW_NAME
//		}
//		err = AddTemplatesForGlobalSecretsInWorkflowTemplate(globalCmCsConfigs, &steps, &volumes, &templates)
//		if err != nil {
//			impl.Logger.Errorw("error in creating templates for global secrets", "err", err)
//		}
//	}
//
//	var configMaps bean3.ConfigMapJson
//	var secrets bean3.ConfigSecretJson
//	var existingConfigMap *bean3.ConfigMapJson
//	var existingSecrets *bean3.ConfigSecretJson
//	if isJob {
//		existingConfigMap, existingSecrets, err = impl.appService.GetCmSecretNew(workflowRequest.AppId, workflowRequest.EnvironmentId, isJob)
//		if err != nil {
//			impl.Logger.Errorw("failed to get configmap data", "err", err)
//			return nil, err
//		}
//		impl.Logger.Debugw("existing cm sec", "cm", existingConfigMap, "sec", existingSecrets)
//		configMaps, secrets, err = getConfigMapsAndSecrets(workflowRequest, existingConfigMap, existingSecrets)
//		if err != nil {
//			impl.Logger.Errorw("fail to get config map and secret with new name", err)
//		}
//
//		err = processConfigMapsAndSecrets(impl, &configMaps, &secrets, &entryPoint, &steps, &volumes, &templates)
//		if err != nil {
//			impl.Logger.Errorw("fail to append cm/cs", err)
//		}
//	}
//	steps = append(steps, v1alpha1.ParallelSteps{
//		Steps: []v1alpha1.WorkflowStep{
//			{
//				Name:     "run-wf",
//				Template: CI_WORKFLOW_NAME,
//			},
//		},
//	})
//	templates = append(templates, v1alpha1.Template{
//		Name:  CI_WORKFLOW_WITH_STAGES,
//		Steps: steps,
//	})
//
//	eventEnv := v12.EnvVar{Name: "CI_CD_EVENT", Value: string(workflowJson)}
//	inAppLoggingEnv := v12.EnvVar{Name: "IN_APP_LOGGING", Value: strconv.FormatBool(impl.ciConfig.InAppLoggingEnabled)}
//	containerEnvVariables = append(containerEnvVariables, eventEnv, inAppLoggingEnv)
//	ciTemplate := v1alpha1.Template{
//		Name: CI_WORKFLOW_NAME,
//		Container: &v12.Container{
//			Env:   containerEnvVariables,
//			Image: workflowRequest.CiImage, //TODO need to check whether trigger buildX image or normal image
//			SecurityContext: &v12.SecurityContext{
//				Privileged: &privileged,
//			},
//			Resources: v12.ResourceRequirements{
//				Limits: v12.ResourceList{
//					"cpu":    resource.MustParse(limitCpu),
//					"memory": resource.MustParse(limitMem),
//				},
//				Requests: v12.ResourceList{
//					"cpu":    resource.MustParse(reqCpu),
//					"memory": resource.MustParse(reqMem),
//				},
//			},
//			Ports: []v12.ContainerPort{{
//				//exposed for user specific data from ci container
//				Name:          "app-data",
//				ContainerPort: 9102,
//			}},
//		},
//		ActiveDeadlineSeconds: &intstr.IntOrString{
//			IntVal: int32(workflowRequest.ActiveDeadlineSeconds),
//		},
//		ArchiveLocation: &v1alpha1.ArtifactLocation{
//			ArchiveLogs: &archiveLogs,
//		},
//	}
//	if isJob {
//		ciTemplate, err = getCiTemplateWithConfigMapsAndSecrets(&configMaps, &secrets, ciTemplate, existingConfigMap, existingSecrets)
//	}
//	if impl.ciConfig.UseBlobStorageConfigInCiWorkflow || !workflowRequest.IsExtRun {
//		gcpBlobConfig := workflowRequest.GcpBlobConfig
//		blobStorageS3Config := workflowRequest.BlobStorageS3Config
//		cloudStorageKey := impl.ciConfig.CiDefaultBuildLogsKeyPrefix + "/" + workflowRequest.WorkflowNamePrefix
//		var s3Artifact *v1alpha1.S3Artifact
//		var gcsArtifact *v1alpha1.GCSArtifact
//		if blobStorageConfigured && blobStorageS3Config != nil {
//			s3CompatibleEndpointUrl := blobStorageS3Config.EndpointUrl
//			if s3CompatibleEndpointUrl == "" {
//				s3CompatibleEndpointUrl = "s3.amazonaws.com"
//			} else {
//				parsedUrl, err := url.Parse(s3CompatibleEndpointUrl)
//				if err != nil {
//					impl.Logger.Errorw("error occurred while parsing s3CompatibleEndpointUrl, ", "s3CompatibleEndpointUrl", s3CompatibleEndpointUrl, "err", err)
//				} else {
//					s3CompatibleEndpointUrl = parsedUrl.Host
//				}
//			}
//			isInsecure := blobStorageS3Config.IsInSecure
//
//			var accessKeySelector *v12.SecretKeySelector
//			var secretKeySelector *v12.SecretKeySelector
//			if blobStorageS3Config.AccessKey != "" {
//				accessKeySelector = &v12.SecretKeySelector{
//					Key: "accessKey",
//					LocalObjectReference: v12.LocalObjectReference{
//						Name: "workflow-minio-cred",
//					},
//				}
//				secretKeySelector = &v12.SecretKeySelector{
//					Key: "secretKey",
//					LocalObjectReference: v12.LocalObjectReference{
//						Name: "workflow-minio-cred",
//					},
//				}
//			}
//			s3Artifact = &v1alpha1.S3Artifact{
//				Key: cloudStorageKey,
//				S3Bucket: v1alpha1.S3Bucket{
//					Endpoint:        s3CompatibleEndpointUrl,
//					AccessKeySecret: accessKeySelector,
//					SecretKeySecret: secretKeySelector,
//					Bucket:          blobStorageS3Config.CiLogBucketName,
//					Region:          blobStorageS3Config.CiLogRegion,
//					Insecure:        &isInsecure,
//				},
//			}
//		} else if blobStorageConfigured && gcpBlobConfig != nil {
//			gcsArtifact = &v1alpha1.GCSArtifact{
//				Key: cloudStorageKey,
//				GCSBucket: v1alpha1.GCSBucket{
//					Bucket: gcpBlobConfig.LogBucketName,
//					ServiceAccountKeySecret: &v12.SecretKeySelector{
//						Key: "secretKey",
//						LocalObjectReference: v12.LocalObjectReference{
//							Name: "workflow-minio-cred",
//						},
//					},
//				},
//			}
//		}
//
//		// set in ArchiveLocation
//		ciTemplate.ArchiveLocation.S3 = s3Artifact
//		ciTemplate.ArchiveLocation.GCS = gcsArtifact
//	}
//
//	for _, config := range globalCmCsConfigs {
//		if config.Type == repository.VOLUME_CONFIG {
//			ciTemplate.Container.VolumeMounts = append(ciTemplate.Container.VolumeMounts, v12.VolumeMount{
//				Name:      config.Name + "-vol",
//				MountPath: config.MountPath,
//			})
//		} else if config.Type == repository.ENVIRONMENT_CONFIG {
//			if config.ConfigType == repository.CM_TYPE_CONFIG {
//				ciTemplate.Container.EnvFrom = append(ciTemplate.Container.EnvFrom, v12.EnvFromSource{
//					ConfigMapRef: &v12.ConfigMapEnvSource{
//						LocalObjectReference: v12.LocalObjectReference{
//							Name: config.Name,
//						},
//					},
//				})
//			} else if config.ConfigType == repository.CS_TYPE_CONFIG {
//				ciTemplate.Container.EnvFrom = append(ciTemplate.Container.EnvFrom, v12.EnvFromSource{
//					SecretRef: &v12.SecretEnvSource{
//						LocalObjectReference: v12.LocalObjectReference{
//							Name: config.Name,
//						},
//					},
//				})
//			}
//		}
//	}
//
//	// volume mount
//	volumeMountsForCiJson := impl.ciConfig.VolumeMountsForCiJson
//	if len(volumeMountsForCiJson) > 0 {
//		var volumeMountsForCi []CiVolumeMount
//		// Unmarshal or Decode the JSON to the interface.
//		err = json.Unmarshal([]byte(volumeMountsForCiJson), &volumeMountsForCi)
//		if err != nil {
//			impl.Logger.Errorw("err in unmarshalling volumeMountsForCiJson", "err", err, "val", volumeMountsForCiJson)
//			return nil, err
//		}
//
//		for _, volumeMountsForCi := range volumeMountsForCi {
//			hostPathDirectoryOrCreate := v12.HostPathDirectoryOrCreate
//			ciTemplate.Volumes = append(ciTemplate.Volumes, v12.Volume{
//				Name: volumeMountsForCi.Name,
//				VolumeSource: v12.VolumeSource{
//					HostPath: &v12.HostPathVolumeSource{
//						Path: volumeMountsForCi.HostMountPath,
//						Type: &hostPathDirectoryOrCreate,
//					},
//				},
//			})
//			ciTemplate.Container.VolumeMounts = append(ciTemplate.Container.VolumeMounts, v12.VolumeMount{
//				Name:      volumeMountsForCi.Name,
//				MountPath: volumeMountsForCi.ContainerMountPath,
//			})
//		}
//	}
//
//	// pvc mounting starts
//	if len(pvc) != 0 {
//		buildPvcCachePath := impl.ciConfig.BuildPvcCachePath
//		buildxPvcCachePath := impl.ciConfig.BuildxPvcCachePath
//		defaultPvcCachePath := impl.ciConfig.DefaultPvcCachePath
//
//		ciTemplate.Volumes = append(ciTemplate.Volumes, v12.Volume{
//			Name: "root-vol",
//			VolumeSource: v12.VolumeSource{
//				PersistentVolumeClaim: &v12.PersistentVolumeClaimVolumeSource{
//					ClaimName: pvc,
//					ReadOnly:  false,
//				},
//			},
//		})
//		ciTemplate.Container.VolumeMounts = append(ciTemplate.Container.VolumeMounts,
//			v12.VolumeMount{
//				Name:      "root-vol",
//				MountPath: buildPvcCachePath,
//			},
//			v12.VolumeMount{
//				Name:      "root-vol",
//				MountPath: buildxPvcCachePath,
//			},
//			v12.VolumeMount{
//				Name:      "root-vol",
//				MountPath: defaultPvcCachePath,
//			})
//	}
//
//	// node selector
//	if val, ok := appLabels[CI_NODE_SELECTOR_APP_LABEL_KEY]; ok && !(isJob && workflowRequest.IsExtRun) {
//		var nodeSelectors map[string]string
//		// Unmarshal or Decode the JSON to the interface.
//		err = json.Unmarshal([]byte(val), &nodeSelectors)
//		if err != nil {
//			impl.Logger.Errorw("err in unmarshalling nodeSelectors", "err", err, "val", val)
//			return nil, err
//		}
//		ciTemplate.NodeSelector = nodeSelectors
//	}
//
//	templates = append(templates, ciTemplate)
//	var (
//		ciWorkflow = v1alpha1.Workflow{
//			ObjectMeta: v1.ObjectMeta{
//				GenerateName: workflowRequest.WorkflowNamePrefix + "-",
//				Labels:       map[string]string{"devtron.ai/workflow-purpose": "ci"},
//			},
//			Spec: v1alpha1.WorkflowSpec{
//				ServiceAccountName: impl.ciConfig.CiWorkflowServiceAccount,
//				//NodeSelector:            map[string]string{impl.ciCdConfig.CiTaintKey: impl.ciCdConfig.CiTaintValue},
//				//Tolerations:             []v12.Toleration{{Key: impl.ciCdConfig.CiTaintKey, Value: impl.ciCdConfig.CiTaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}},
//				Entrypoint: entryPoint,
//				TTLStrategy: &v1alpha1.TTLStrategy{
//					SecondsAfterCompletion: &ttl,
//				},
//				Templates: templates,
//				Volumes:   volumes,
//			},
//		}
//	)
//
//	if impl.ciConfig.CiTaintKey != "" || impl.ciConfig.CiTaintValue != "" {
//		ciWorkflow.Spec.Tolerations = []v12.Toleration{{Key: impl.ciConfig.CiTaintKey, Value: impl.ciConfig.CiTaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}}
//	}
//
//	// In the future, we will give support for NodeSelector for job currently we need to have a node without dedicated NodeLabel to run job
//	if len(impl.ciConfig.NodeLabel) > 0 && !(isJob && workflowRequest.IsExtRun) {
//		ciWorkflow.Spec.NodeSelector = impl.ciConfig.NodeLabel
//	}
//	wfTemplate, err := json.Marshal(ciWorkflow)
//	if err != nil {
//		impl.Logger.Errorw("marshal error", "err", err)
//	}
//	impl.Logger.Debug("---->", string(wfTemplate))
//
//	createdWf, err := wfClient.Create(context.Background(), &ciWorkflow, v1.CreateOptions{}) // submit the hello world workflow
//	impl.Logger.Debug("workflow submitted: " + createdWf.Name)
//	impl.checkErr(err)
//	return createdWf, err
//}
//
//func getConfigMapsAndSecrets(workflowRequest *WorkflowRequest, existingConfigMap *bean3.ConfigMapJson, existingSecrets *bean3.ConfigSecretJson) (bean3.ConfigMapJson, bean3.ConfigSecretJson, error) {
//	configMaps := bean3.ConfigMapJson{}
//	secrets := bean3.ConfigSecretJson{}
//	for _, cm := range existingConfigMap.Maps {
//		if cm.External {
//			continue
//		}
//		configMaps.Maps = append(configMaps.Maps, cm)
//	}
//	for i := range configMaps.Maps {
//		configMaps.Maps[i].Name = configMaps.Maps[i].Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + CI_WORKFLOW_NAME
//	}
//
//	for _, s := range existingSecrets.Secrets {
//		if s.External {
//			continue
//		}
//		secrets.Secrets = append(secrets.Secrets, s)
//	}
//	for i := range secrets.Secrets {
//		secrets.Secrets[i].Name = secrets.Secrets[i].Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + CI_WORKFLOW_NAME
//	}
//	return configMaps, secrets, nil
//}
//
//func processConfigMapsAndSecrets(impl *WorkflowServiceImpl, configMaps *bean3.ConfigMapJson, secrets *bean3.ConfigSecretJson, entryPoint *string, steps *[]v1alpha1.ParallelSteps, volumes *[]v12.Volume, templates *[]v1alpha1.Template) error {
//
//	var configsMapping, secretsMapping map[string]string
//	var err error
//
//	if len(configMaps.Maps) > 0 {
//		configsMapping, err = processConfigMap(impl, configMaps, entryPoint, steps)
//		if err != nil {
//			return err
//		}
//	}
//
//	if len(secrets.Secrets) > 0 {
//		secretsMapping, err = processSecrets(impl, entryPoint, secrets, steps)
//		if err != nil {
//			return err
//		}
//	}
//	var secretMap []bean3.ConfigSecretMap
//	for _, cs := range secrets.Secrets {
//		secretMap = append(secretMap, *cs)
//	}
//
//	*volumes = ExtractVolumesFromCmCs(configMaps.Maps, secretMap)
//	if len(configsMapping) > 0 {
//		for i, cm := range configMaps.Maps {
//			*templates = append(*templates, getResourceTemplate("cm-"+strconv.Itoa(i), configsMapping[cm.Name]))
//		}
//	}
//
//	if len(secretsMapping) > 0 {
//		for i, s := range secrets.Secrets {
//			*templates = append(*templates, getResourceTemplate("sec-"+strconv.Itoa(i), secretsMapping[s.Name]))
//		}
//	}
//	return nil
//}
//
//func getResourceTemplate(prefix string, manifestName string) v1alpha1.Template {
//	return v1alpha1.Template{
//		Name: prefix,
//		Resource: &v1alpha1.ResourceTemplate{
//			Action:            "create",
//			SetOwnerReference: true,
//			Manifest:          manifestName,
//		},
//	}
//}
//
//func processSecrets(impl *WorkflowServiceImpl, entryPoint *string, secrets *bean3.ConfigSecretJson, steps *[]v1alpha1.ParallelSteps) (map[string]string, error) {
//	secretsMapping := make(map[string]string)
//	*entryPoint = CI_WORKFLOW_WITH_STAGES
//	for i, s := range secrets.Secrets {
//		var datamap map[string]string
//		if err := json.Unmarshal(s.Data, &datamap); err != nil {
//			impl.Logger.Errorw("error while unmarshal data", "err", err)
//			return secretsMapping, err
//		}
//		ownerDelete := true
//		configMapSecretDto := ConfigMapSecretDto{
//			Name: s.Name,
//			Data: datamap,
//			OwnerRef: v1.OwnerReference{
//				APIVersion:         "argoproj.io/v1alpha1",
//				Kind:               "Workflow",
//				Name:               "{{workflow.name}}",
//				UID:                "{{workflow.uid}}",
//				BlockOwnerDeletion: &ownerDelete,
//			},
//		}
//		secretObject := GetSecretBody(configMapSecretDto)
//		secretJson, err := json.Marshal(secretObject)
//		if err != nil {
//			impl.Logger.Errorw("error in building json", "err", err)
//			return secretsMapping, err
//		}
//		secretsMapping[s.Name] = string(secretJson)
//		*steps = append(*steps, v1alpha1.ParallelSteps{
//			Steps: []v1alpha1.WorkflowStep{
//				{
//					Name:     "create-env-sec-" + strconv.Itoa(i),
//					Template: "sec-" + strconv.Itoa(i),
//				},
//			},
//		})
//	}
//	return secretsMapping, nil
//}
//
//func processConfigMap(impl *WorkflowServiceImpl, configMaps *bean3.ConfigMapJson, entryPoint *string, steps *[]v1alpha1.ParallelSteps) (map[string]string, error) {
//	configsMapping := make(map[string]string)
//	*entryPoint = CI_WORKFLOW_WITH_STAGES
//	for i, cm := range configMaps.Maps {
//		var dataMap map[string]string
//		if err := json.Unmarshal(cm.Data, &dataMap); err != nil {
//			impl.Logger.Errorw("error while unmarshal data", "err", err)
//			return configsMapping, err
//		}
//		ownerDelete := true
//		configMapSecretDto := ConfigMapSecretDto{
//			Name: cm.Name,
//			Data: dataMap,
//			OwnerRef: v1.OwnerReference{
//				APIVersion:         "argoproj.io/v1alpha1",
//				Kind:               "Workflow",
//				Name:               "{{workflow.name}}",
//				UID:                "{{workflow.uid}}",
//				BlockOwnerDeletion: &ownerDelete,
//			},
//		}
//		cmBody := GetConfigMapBody(configMapSecretDto)
//		cmJson, err := json.Marshal(cmBody)
//		if err != nil {
//			impl.Logger.Errorw("error in building json", "err", err)
//			return configsMapping, err
//		}
//		configsMapping[cm.Name] = string(cmJson)
//
//		*steps = append(*steps, v1alpha1.ParallelSteps{
//			Steps: []v1alpha1.WorkflowStep{
//				{
//					Name:     "create-env-cm-" + strconv.Itoa(i),
//					Template: "cm-" + strconv.Itoa(i),
//				},
//			},
//		})
//	}
//	return configsMapping, nil
//}
//func getCiTemplateWithConfigMapsAndSecrets(configMaps *bean3.ConfigMapJson, secrets *bean3.ConfigSecretJson, ciTemplate v1alpha1.Template, existingConfigMap *bean3.ConfigMapJson, existingSecrets *bean3.ConfigSecretJson) (v1alpha1.Template, error) {
//	var secretMap []bean3.ConfigSecretMap
//	for _, cs := range secrets.Secrets {
//		secretMap = append(secretMap, *cs)
//	}
//	UpdateContainerEnvsFromCmCs(ciTemplate.Container, configMaps.Maps, secretMap)
//	return ciTemplate, nil
//}
//
////func (impl *WorkflowServiceImpl) getRuntimeEnvClientInstance(environment *repository2.Environment) (v1alpha12.WorkflowInterface, error) {
////	restConfig, err, _ := impl.k8sCommonService.GetRestConfigByClusterId(context.Background(), environment.ClusterId)
////	if err != nil {
////		impl.Logger.Errorw("error in getting rest config by cluster id", "err", err)
////		return nil, err
////	}
////	clientSet, err := versioned.NewForConfig(restConfig)
////	if err != nil {
////		impl.Logger.Errorw("err", "err", err)
////		return nil, err
////	}
////	wfClient := clientSet.ArgoprojV1alpha1().Workflows(environment.Namespace) // create the workflow client
////	return wfClient, nil
////}
//
////func (impl *WorkflowServiceImpl) getClientInstance(namespace string) (v1alpha12.WorkflowInterface, error) {
////	clientSet, err := versioned.NewForConfig(impl.config)
////	if err != nil {
////		impl.Logger.Errorw("err on get client instance", "err", err)
////		return nil, err
////	}
////	wfClient := clientSet.ArgoprojV1alpha1().Workflows(namespace) // create the workflow client
////	return wfClient, nil
////}
//
//func (impl *WorkflowServiceImpl) GetWorkflow(name string, namespace string, isExt bool, environment *repository2.Environment) (*v1alpha1.Workflow, error) {
//	impl.Logger.Debug("getting wf", name)
//	wfClient, err := impl.getWfClient(environment, namespace, isExt)
//
//	if err != nil {
//		return nil, err
//	}
//
//	workflow, err := wfClient.Get(context.Background(), name, v1.GetOptions{})
//	return workflow, err
//}
//
//func (impl *WorkflowServiceImpl) TerminateWorkflow(name string, namespace string, isExt bool, environment *repository2.Environment) error {
//	impl.Logger.Debugw("terminating wf", "name", name)
//
//	wfClient, err := impl.getWfClient(environment, namespace, isExt)
//	if err != nil {
//		return err
//	}
//	err = util.TerminateWorkflow(context.Background(), wfClient, name)
//	return err
//}
//func (impl *WorkflowServiceImpl) getWorkflowExecutor(executorType pipelineConfig.WorkflowExecutorType) WorkflowExecutor {
//	if executorType == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_AWF {
//		return impl.argoWorkflowExecutor
//	} else if executorType == pipelineConfig.WORKFLOW_EXECUTOR_TYPE_SYSTEM {
//		return impl.systemWorkflowExecutor
//	}
//	impl.Logger.Warnw("workflow executor not found", "type", executorType)
//	return nil
//}
//
//func (impl *WorkflowServiceImpl) TerminateWorkflowForCiCd(executorType pipelineConfig.WorkflowExecutorType, name string, namespace string, restConfig *rest.Config, isExt bool, environment *repository2.Environment) error {
//	impl.Logger.Debugw("terminating wf", "name", name)
//	var err error
//	if executorType != "" {
//		workflowExecutor := impl.getWorkflowExecutor(executorType)
//		err = workflowExecutor.TerminateWorkflow(name, namespace, restConfig)
//	} else {
//		wfClient, err := impl.getWfClient(environment, namespace, isExt)
//		if err != nil {
//			return err
//		}
//		err = util.TerminateWorkflow(context.Background(), wfClient, name)
//	}
//	return err
//}
//
//func (impl *WorkflowServiceImpl) UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
//	impl.Logger.Debugw("updating wf", "name", wf.Name)
//	wfClient, err := impl.getClientInstance(wf.Namespace)
//	if err != nil {
//		impl.Logger.Errorw("cannot build wf client", "err", err)
//		return nil, err
//	}
//	updatedWf, err := wfClient.Update(context.Background(), wf, v1.UpdateOptions{})
//	if err != nil {
//		impl.Logger.Errorw("cannot update wf ", "err", err)
//		return nil, err
//	}
//	impl.Logger.Debugw("updated wf", "name", wf.Name)
//	return updatedWf, err
//}
//
//func (impl *WorkflowServiceImpl) ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error) {
//	impl.Logger.Debug("listing all wfs")
//	wfClient, err := impl.getClientInstance(namespace)
//	if err != nil {
//		impl.Logger.Errorw("cannot build wf client", "err", err)
//		return nil, err
//	}
//	workflowList, err := wfClient.List(context.Background(), v1.ListOptions{})
//	return workflowList, err
//}
//
//func (impl *WorkflowServiceImpl) DeleteWorkflow(wfName string, namespace string) error {
//	impl.Logger.Debugw("deleting wf", "name", wfName)
//	wfClient, err := impl.getClientInstance(namespace)
//	if err != nil {
//		impl.Logger.Errorw("cannot build wf client", "err", err)
//		return err
//	}
//	err = wfClient.Delete(context.Background(), wfName, v1.DeleteOptions{})
//	return err
//}
//
//func (impl *WorkflowServiceImpl) checkErr(err error) {
//	if err != nil {
//		impl.Logger.Errorw("error", "error:", err)
//	}
//}
//
//func (impl *WorkflowServiceImpl) getWfClient(environment *repository2.Environment, namespace string, isExt bool) (v1alpha12.WorkflowInterface, error) {
//	var wfClient v1alpha12.WorkflowInterface
//	var err error
//	if isExt {
//		wfClient, err = impl.getRuntimeEnvClientInstance(environment)
//		if err != nil {
//			impl.Logger.Errorw("cannot build wf client", "err", err)
//			return nil, err
//		}
//	} else {
//		wfClient, err = impl.getClientInstance(namespace)
//		if err != nil {
//			impl.Logger.Errorw("cannot build wf client", "err", err)
//			return nil, err
//		}
//	}
//	return wfClient, nil
//}
