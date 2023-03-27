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
	"fmt"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"k8s.io/apimachinery/pkg/util/intstr"
	"net/url"
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
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type CdWorkflowService interface {
	SubmitWorkflow(workflowRequest *CdWorkflowRequest, pipeline *pipelineConfig.Pipeline, env *repository.Environment) (*v1alpha1.Workflow, error)
	DeleteWorkflow(wfName string, namespace string) error
	GetWorkflow(name string, namespace string, url string, token string, isExtRun bool) (*v1alpha1.Workflow, error)
	ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error)
	UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error)
	TerminateWorkflow(name string, namespace string, url string, token string, isExtRun bool) error
}

const (
	CD_WORKFLOW_NAME        = "cd"
	CD_WORKFLOW_WITH_STAGES = "cd-stages-with-env"
)

type CdWorkflowServiceImpl struct {
	Logger            *zap.SugaredLogger
	config            *rest.Config
	cdConfig          *CdConfig
	appService        app.AppService
	envRepository     repository.EnvironmentRepository
	globalCMCSService GlobalCMCSService
}

type CdWorkflowRequest struct {
	AppId                      int                               `json:"appId"`
	EnvironmentId              int                               `json:"envId"`
	WorkflowId                 int                               `json:"workflowId"`
	WorkflowRunnerId           int                               `json:"workflowRunnerId"`
	CdPipelineId               int                               `json:"cdPipelineId"`
	TriggeredBy                int32                             `json:"triggeredBy"`
	StageYaml                  string                            `json:"stageYaml"`
	ArtifactLocation           string                            `json:"artifactLocation"`
	ArtifactBucket             string                            `json:"ciArtifactBucket"`
	ArtifactFileName           string                            `json:"ciArtifactFileName"`
	ArtifactRegion             string                            `json:"ciArtifactRegion"`
	CiProjectDetails           []CiProjectDetails                `json:"ciProjectDetails"`
	CiArtifactDTO              CiArtifactDTO                     `json:"ciArtifactDTO"`
	Namespace                  string                            `json:"namespace"`
	WorkflowNamePrefix         string                            `json:"workflowNamePrefix"`
	CdImage                    string                            `json:"cdImage"`
	ActiveDeadlineSeconds      int64                             `json:"activeDeadlineSeconds"`
	StageType                  string                            `json:"stageType"`
	DockerUsername             string                            `json:"dockerUsername"`
	DockerPassword             string                            `json:"dockerPassword"`
	AwsRegion                  string                            `json:"awsRegion"`
	SecretKey                  string                            `json:"secretKey"`
	AccessKey                  string                            `json:"accessKey"`
	DockerConnection           string                            `json:"dockerConnection"`
	DockerCert                 string                            `json:"dockerCert"`
	CdCacheLocation            string                            `json:"cdCacheLocation"`
	CdCacheRegion              string                            `json:"cdCacheRegion"`
	DockerRegistryType         string                            `json:"dockerRegistryType"`
	DockerRegistryURL          string                            `json:"dockerRegistryURL"`
	OrchestratorHost           string                            `json:"orchestratorHost"`
	OrchestratorToken          string                            `json:"orchestratorToken"`
	IsExtRun                   bool                              `json:"isExtRun"`
	ExtraEnvironmentVariables  map[string]string                 `json:"extraEnvironmentVariables"`
	BlobStorageConfigured      bool                              `json:"blobStorageConfigured"`
	BlobStorageS3Config        *blob_storage.BlobStorageS3Config `json:"blobStorageS3Config"`
	CloudProvider              blob_storage.BlobStorageType      `json:"cloudProvider"`
	AzureBlobConfig            *blob_storage.AzureBlobConfig     `json:"azureBlobConfig"`
	GcpBlobConfig              *blob_storage.GcpBlobConfig       `json:"gcpBlobConfig"`
	DefaultAddressPoolBaseCidr string                            `json:"defaultAddressPoolBaseCidr"`
	DefaultAddressPoolSize     int                               `json:"defaultAddressPoolSize"`
	DeploymentTriggeredBy      string                            `json:"deploymentTriggeredBy,omitempty"`
	DeploymentTriggerTime      time.Time                         `json:"deploymentTriggerTime,omitempty"`
	DeploymentReleaseCounter   int                               `json:"deploymentReleaseCounter,omitempty"`
}

const PRE = "PRE"
const POST = "POST"

func NewCdWorkflowServiceImpl(Logger *zap.SugaredLogger,
	envRepository repository.EnvironmentRepository,
	cdConfig *CdConfig,
	appService app.AppService,
	globalCMCSService GlobalCMCSService) *CdWorkflowServiceImpl {
	return &CdWorkflowServiceImpl{Logger: Logger,
		config:            cdConfig.ClusterConfig,
		cdConfig:          cdConfig,
		appService:        appService,
		envRepository:     envRepository,
		globalCMCSService: globalCMCSService}
}

func (impl *CdWorkflowServiceImpl) SubmitWorkflow(workflowRequest *CdWorkflowRequest, pipeline *pipelineConfig.Pipeline, env *repository.Environment) (*v1alpha1.Workflow, error) {
	containerEnvVariables := []v12.EnvVar{}
	if impl.cdConfig.CloudProvider == BLOB_STORAGE_S3 && impl.cdConfig.BlobStorageS3AccessKey != "" {
		miniCred := []v12.EnvVar{{Name: "AWS_ACCESS_KEY_ID", Value: impl.cdConfig.BlobStorageS3AccessKey}, {Name: "AWS_SECRET_ACCESS_KEY", Value: impl.cdConfig.BlobStorageS3SecretKey}}
		containerEnvVariables = append(containerEnvVariables, miniCred...)
	}
	if (workflowRequest.StageType == PRE && pipeline.RunPreStageInEnv) || (workflowRequest.StageType == POST && pipeline.RunPostStageInEnv) {
		workflowRequest.IsExtRun = true
	}
	ciCdTriggerEvent := CiCdTriggerEvent{
		CdRequest: workflowRequest,
	}
	workflowJson, err := json.Marshal(&ciCdTriggerEvent)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return nil, err
	}

	privileged := true
	storageConfigured := workflowRequest.BlobStorageConfigured
	archiveLogs := storageConfigured

	limitCpu := impl.cdConfig.LimitCpu
	limitMem := impl.cdConfig.LimitMem

	reqCpu := impl.cdConfig.ReqCpu
	reqMem := impl.cdConfig.ReqMem
	ttl := int32(impl.cdConfig.BuildLogTTLValue)

	entryPoint := CD_WORKFLOW_NAME

	steps := make([]v1alpha1.ParallelSteps, 0)
	volumes := make([]v12.Volume, 0)
	templates := make([]v1alpha1.Template, 0)

	var globalCmCsConfigs []*GlobalCMCSDto

	if !workflowRequest.IsExtRun {
		globalCmCsConfigs, err = impl.globalCMCSService.FindAllActiveByPipelineType(repository2.PIPELINE_TYPE_CD)
		// inject global variables only if IsExtRun is false
		if err != nil {
			impl.Logger.Errorw("error in getting all global cm/cs config", "err", err)
			return nil, err
		}
		if len(globalCmCsConfigs) > 0 {
			entryPoint = CD_WORKFLOW_WITH_STAGES
		}
		for i := range globalCmCsConfigs {
			globalCmCsConfigs[i].Name = fmt.Sprintf("%s-%s-%s", strings.ToLower(globalCmCsConfigs[i].Name), strconv.Itoa(workflowRequest.WorkflowRunnerId), CD_WORKFLOW_NAME)
		}

		err = impl.globalCMCSService.AddTemplatesForGlobalSecretsInWorkflowTemplate(globalCmCsConfigs, &steps, &volumes, &templates)
		if err != nil {
			impl.Logger.Errorw("error in creating templates for global secrets", "err", err)
		}
	}

	preStageConfigMapSecretsJson := pipeline.PreStageConfigMapSecretNames
	postStageConfigMapSecretsJson := pipeline.PostStageConfigMapSecretNames

	existingConfigMap, existingSecrets, err := impl.appService.GetCmSecretNew(workflowRequest.AppId, workflowRequest.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("failed to get configmap data", "err", err)
		return nil, err
	}
	impl.Logger.Debugw("existing cm sec", "cm", existingConfigMap, "sec", existingSecrets)

	preStageConfigmapSecrets := bean2.PreStageConfigMapSecretNames{}
	err = json.Unmarshal([]byte(preStageConfigMapSecretsJson), &preStageConfigmapSecrets)
	if err != nil {
		impl.Logger.Error(err)
		return nil, err
	}
	postStageConfigmapSecrets := bean2.PostStageConfigMapSecretNames{}
	err = json.Unmarshal([]byte(postStageConfigMapSecretsJson), &postStageConfigmapSecrets)
	if err != nil {
		impl.Logger.Error(err)
		return nil, err
	}

	cdPipelineLevelConfigMaps := make(map[string]bool)
	cdPipelineLevelSecrets := make(map[string]bool)
	//cdPipelineLevelSecrets := make(map[string]bool)

	if workflowRequest.StageType == PRE {
		for _, cm := range preStageConfigmapSecrets.ConfigMaps {
			cdPipelineLevelConfigMaps[cm] = true
		}
		for _, secret := range preStageConfigmapSecrets.Secrets {
			cdPipelineLevelSecrets[secret] = true
		}
	} else {
		for _, cm := range postStageConfigmapSecrets.ConfigMaps {
			cdPipelineLevelConfigMaps[cm] = true
		}
		for _, secret := range postStageConfigmapSecrets.Secrets {
			cdPipelineLevelSecrets[secret] = true
		}
	}

	configMaps := bean.ConfigMapJson{}
	for _, cm := range existingConfigMap.Maps {
		if cm.External {
			continue
		}
		if _, ok := cdPipelineLevelConfigMaps[cm.Name]; ok {
			configMaps.Maps = append(configMaps.Maps, cm)
		}
	}
	for i := range configMaps.Maps {
		configMaps.Maps[i].Name = configMaps.Maps[i].Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + strconv.Itoa(workflowRequest.WorkflowRunnerId)
	}

	secrets := bean.ConfigSecretJson{}
	for _, s := range existingSecrets.Secrets {
		if s.External {
			continue
		}
		if _, ok := cdPipelineLevelSecrets[s.Name]; ok {
			secrets.Secrets = append(secrets.Secrets, s)
		}
	}
	for i := range configMaps.Maps {
		configMaps.Maps[i].Name = configMaps.Maps[i].Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + strconv.Itoa(workflowRequest.WorkflowRunnerId)
	}
	for i := range secrets.Secrets {
		secrets.Secrets[i].Name = secrets.Secrets[i].Name + "-" + strconv.Itoa(workflowRequest.WorkflowId) + "-" + strconv.Itoa(workflowRequest.WorkflowRunnerId)
	}

	configsMapping := make(map[string]string)
	secretsMapping := make(map[string]string)

	if len(configMaps.Maps) > 0 {
		entryPoint = CD_WORKFLOW_WITH_STAGES
		for i, cm := range configMaps.Maps {
			var datamap map[string]string
			if err := json.Unmarshal(cm.Data, &datamap); err != nil {
				impl.Logger.Errorw("error while unmarshal data", "err", err)
				return nil, err
			}
			ownerDelete := true
			cmBody := v12.ConfigMap{
				TypeMeta: v1.TypeMeta{
					Kind:       "ConfigMap",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: cm.Name,
					OwnerReferences: []v1.OwnerReference{{
						APIVersion:         "argoproj.io/v1alpha1",
						Kind:               "Workflow",
						Name:               "{{workflow.name}}",
						UID:                "{{workflow.uid}}",
						BlockOwnerDeletion: &ownerDelete,
					}},
				},
				Data: datamap,
			}
			cmJson, err := json.Marshal(cmBody)
			if err != nil {
				impl.Logger.Errorw("error in building json", "err", err)
				return nil, err
			}
			configsMapping[cm.Name] = string(cmJson)

			if cm.Type == "volume" {
				volumes = append(volumes, v12.Volume{
					Name: cm.Name + "-vol",
					VolumeSource: v12.VolumeSource{
						ConfigMap: &v12.ConfigMapVolumeSource{
							LocalObjectReference: v12.LocalObjectReference{
								Name: cm.Name,
							},
						},
					},
				})
			}
			steps = append(steps, v1alpha1.ParallelSteps{
				Steps: []v1alpha1.WorkflowStep{
					{
						Name:     "create-env-cm-" + strconv.Itoa(i),
						Template: "cm-" + strconv.Itoa(i),
					},
				},
			})
		}
	}

	if len(secrets.Secrets) > 0 {
		entryPoint = CD_WORKFLOW_WITH_STAGES
		for i, s := range secrets.Secrets {
			var datamap map[string][]byte
			if err := json.Unmarshal(s.Data, &datamap); err != nil {
				impl.Logger.Errorw("error while unmarshal data", "err", err)
				return nil, err
			}
			ownerDelete := true
			secretObject := v12.Secret{
				TypeMeta: v1.TypeMeta{
					Kind:       "Secret",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name: s.Name,
					OwnerReferences: []v1.OwnerReference{{
						APIVersion:         "argoproj.io/v1alpha1",
						Kind:               "Workflow",
						Name:               "{{workflow.name}}",
						UID:                "{{workflow.uid}}",
						BlockOwnerDeletion: &ownerDelete,
					}},
				},
				Data: datamap,
				Type: "Opaque",
			}
			secretJson, err := json.Marshal(secretObject)
			if err != nil {
				impl.Logger.Errorw("error in building json", "err", err)
				return nil, err
			}
			secretsMapping[s.Name] = string(secretJson)
			if s.Type == "volume" {
				volumes = append(volumes, v12.Volume{
					Name: s.Name + "-vol",
					VolumeSource: v12.VolumeSource{
						Secret: &v12.SecretVolumeSource{
							SecretName: s.Name,
						},
					},
				})
			}
			steps = append(steps, v1alpha1.ParallelSteps{
				Steps: []v1alpha1.WorkflowStep{
					{
						Name:     "create-env-sec-" + strconv.Itoa(i),
						Template: "sec-" + strconv.Itoa(i),
					},
				},
			})
		}
	}

	if len(configsMapping) > 0 {
		for i, cm := range configMaps.Maps {
			templates = append(templates, v1alpha1.Template{
				Name: "cm-" + strconv.Itoa(i),
				Resource: &v1alpha1.ResourceTemplate{
					Action:            "create",
					SetOwnerReference: true,
					Manifest:          configsMapping[cm.Name],
				},
			})
		}
	}
	if len(secretsMapping) > 0 {
		for i, s := range secrets.Secrets {
			templates = append(templates, v1alpha1.Template{
				Name: "sec-" + strconv.Itoa(i),
				Resource: &v1alpha1.ResourceTemplate{
					Action:            "create",
					SetOwnerReference: true,
					Manifest:          secretsMapping[s.Name],
				},
			})
		}
	}
	steps = append(steps, v1alpha1.ParallelSteps{
		Steps: []v1alpha1.WorkflowStep{
			{
				Name:     "run-wf",
				Template: CD_WORKFLOW_NAME,
			},
		},
	})
	templates = append(templates, v1alpha1.Template{
		Name:  CD_WORKFLOW_WITH_STAGES,
		Steps: steps,
	})

	cdTemplate := v1alpha1.Template{
		Name: "cd",
		Container: &v12.Container{
			Env:   containerEnvVariables,
			Image: workflowRequest.CdImage,
			Args:  []string{string(workflowJson)},
			SecurityContext: &v12.SecurityContext{
				Privileged: &privileged,
			},
			Resources: v12.ResourceRequirements{
				Limits: v12.ResourceList{
					"cpu":    resource.MustParse(limitCpu),
					"memory": resource.MustParse(limitMem),
				},
				Requests: v12.ResourceList{
					"cpu":    resource.MustParse(reqCpu),
					"memory": resource.MustParse(reqMem),
				},
			},
		},
		ActiveDeadlineSeconds: &intstr.IntOrString{
			IntVal: int32(workflowRequest.ActiveDeadlineSeconds),
		},
		ArchiveLocation: &v1alpha1.ArtifactLocation{
			ArchiveLogs: &archiveLogs,
		},
	}

	if impl.cdConfig.UseBlobStorageConfigInCdWorkflow || !workflowRequest.IsExtRun {
		var s3Artifact *v1alpha1.S3Artifact
		var gcsArtifact *v1alpha1.GCSArtifact
		blobStorageS3Config := workflowRequest.BlobStorageS3Config
		gcpBlobConfig := workflowRequest.GcpBlobConfig
		cloudStorageKey := impl.cdConfig.DefaultBuildLogsKeyPrefix + "/" + workflowRequest.WorkflowNamePrefix
		if storageConfigured && blobStorageS3Config != nil {
			s3CompatibleEndpointUrl := blobStorageS3Config.EndpointUrl
			if s3CompatibleEndpointUrl == "" {
				s3CompatibleEndpointUrl = "s3.amazonaws.com"
			} else {
				parsedUrl, err := url.Parse(s3CompatibleEndpointUrl)
				if err != nil {
					impl.Logger.Errorw("error occurred while parsing s3CompatibleEndpointUrl, ", "s3CompatibleEndpointUrl", s3CompatibleEndpointUrl, "err", err)
				} else {
					s3CompatibleEndpointUrl = parsedUrl.Host
				}
			}
			isInsecure := blobStorageS3Config.IsInSecure
			var accessKeySelector *v12.SecretKeySelector
			var secretKeySelector *v12.SecretKeySelector
			if blobStorageS3Config.AccessKey != "" {
				accessKeySelector = &v12.SecretKeySelector{
					Key: "accessKey",
					LocalObjectReference: v12.LocalObjectReference{
						Name: "workflow-minio-cred",
					},
				}
				secretKeySelector = &v12.SecretKeySelector{
					Key: "secretKey",
					LocalObjectReference: v12.LocalObjectReference{
						Name: "workflow-minio-cred",
					},
				}
			}
			s3Artifact = &v1alpha1.S3Artifact{
				Key: cloudStorageKey,
				S3Bucket: v1alpha1.S3Bucket{
					Endpoint:        s3CompatibleEndpointUrl,
					AccessKeySecret: accessKeySelector,
					SecretKeySecret: secretKeySelector,
					Bucket:          blobStorageS3Config.CiLogBucketName,
					Insecure:        &isInsecure,
				},
			}
			if blobStorageS3Config.CiLogRegion != "" {
				//TODO checking for Azure
				s3Artifact.Region = blobStorageS3Config.CiLogRegion
			}
		} else if storageConfigured && gcpBlobConfig != nil {
			gcsArtifact = &v1alpha1.GCSArtifact{
				Key: cloudStorageKey,
				GCSBucket: v1alpha1.GCSBucket{
					Bucket: gcpBlobConfig.LogBucketName,
					ServiceAccountKeySecret: &v12.SecretKeySelector{
						Key: "secretKey",
						LocalObjectReference: v12.LocalObjectReference{
							Name: "workflow-minio-cred",
						},
					},
				},
			}
		}

		// set in ArchiveLocation
		cdTemplate.ArchiveLocation.S3 = s3Artifact
		cdTemplate.ArchiveLocation.GCS = gcsArtifact
	}

	if !workflowRequest.IsExtRun {
		for _, config := range globalCmCsConfigs {
			if config.Type == repository2.VOLUME_CONFIG {
				cdTemplate.Container.VolumeMounts = append(cdTemplate.Container.VolumeMounts, v12.VolumeMount{
					Name:      config.Name + "-vol",
					MountPath: config.MountPath,
				})
			} else if config.Type == repository2.ENVIRONMENT_CONFIG {
				if config.ConfigType == repository2.CM_TYPE_CONFIG {
					cdTemplate.Container.EnvFrom = append(cdTemplate.Container.EnvFrom, v12.EnvFromSource{
						ConfigMapRef: &v12.ConfigMapEnvSource{
							LocalObjectReference: v12.LocalObjectReference{
								Name: config.Name,
							},
						},
					})
				} else if config.ConfigType == repository2.CS_TYPE_CONFIG {
					cdTemplate.Container.EnvFrom = append(cdTemplate.Container.EnvFrom, v12.EnvFromSource{
						SecretRef: &v12.SecretEnvSource{
							LocalObjectReference: v12.LocalObjectReference{
								Name: config.Name,
							},
						},
					})
				}
			}
		}
	}

	for _, cm := range configMaps.Maps {
		if cm.Type == "environment" {
			cdTemplate.Container.EnvFrom = append(cdTemplate.Container.EnvFrom, v12.EnvFromSource{
				ConfigMapRef: &v12.ConfigMapEnvSource{
					LocalObjectReference: v12.LocalObjectReference{
						Name: cm.Name,
					},
				},
			})
		} else if cm.Type == "volume" {
			cdTemplate.Container.VolumeMounts = append(cdTemplate.Container.VolumeMounts, v12.VolumeMount{
				Name:      cm.Name + "-vol",
				MountPath: cm.MountPath,
			})
		}
	}

	// Adding external config map reference in workflow template
	for _, cm := range existingConfigMap.Maps {
		if _, ok := cdPipelineLevelConfigMaps[cm.Name]; ok {
			if cm.External {
				if cm.Type == "environment" {
					cdTemplate.Container.EnvFrom = append(cdTemplate.Container.EnvFrom, v12.EnvFromSource{
						ConfigMapRef: &v12.ConfigMapEnvSource{
							LocalObjectReference: v12.LocalObjectReference{
								Name: cm.Name,
							},
						},
					})
				} else if cm.Type == "volume" {
					cdTemplate.Container.VolumeMounts = append(cdTemplate.Container.VolumeMounts, v12.VolumeMount{
						Name:      cm.Name,
						MountPath: cm.MountPath,
					})
				}
			}
		}
	}

	for _, s := range secrets.Secrets {
		if s.Type == "environment" {
			cdTemplate.Container.EnvFrom = append(cdTemplate.Container.EnvFrom, v12.EnvFromSource{
				SecretRef: &v12.SecretEnvSource{
					LocalObjectReference: v12.LocalObjectReference{
						Name: s.Name,
					},
				},
			})
		} else if s.Type == "volume" {
			cdTemplate.Container.VolumeMounts = append(cdTemplate.Container.VolumeMounts, v12.VolumeMount{
				Name:      s.Name + "-vol",
				MountPath: s.MountPath,
			})
		}
	}

	// Adding external secret reference in workflow template
	for _, s := range existingSecrets.Secrets {
		if _, ok := cdPipelineLevelSecrets[s.Name]; ok {
			if s.External {
				if s.Type == "environment" {
					cdTemplate.Container.EnvFrom = append(cdTemplate.Container.EnvFrom, v12.EnvFromSource{
						SecretRef: &v12.SecretEnvSource{
							LocalObjectReference: v12.LocalObjectReference{
								Name: s.Name,
							},
						},
					})
				} else if s.Type == "volume" {
					cdTemplate.Container.VolumeMounts = append(cdTemplate.Container.VolumeMounts, v12.VolumeMount{
						Name:      s.Name,
						MountPath: s.MountPath,
					})
				}
			}
		}
	}

	templates = append(templates, cdTemplate)
	var (
		cdWorkflow = v1alpha1.Workflow{
			ObjectMeta: v1.ObjectMeta{
				GenerateName: workflowRequest.WorkflowNamePrefix + "-",
				Annotations:  map[string]string{"workflows.argoproj.io/controller-instanceid": impl.cdConfig.WfControllerInstanceID},
				Labels:       map[string]string{"devtron.ai/workflow-purpose": "cd"},
			},
			Spec: v1alpha1.WorkflowSpec{
				ServiceAccountName: impl.cdConfig.WorkflowServiceAccount,
				NodeSelector:       map[string]string{impl.cdConfig.TaintKey: impl.cdConfig.TaintValue},
				Tolerations:        []v12.Toleration{{Key: impl.cdConfig.TaintKey, Value: impl.cdConfig.TaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}},
				Entrypoint:         entryPoint,
				TTLStrategy: &v1alpha1.TTLStrategy{
					SecondsAfterCompletion: &ttl,
				},
				Templates: templates,
				Volumes:   volumes,
			},
		}
	)

	//
	if len(impl.cdConfig.NodeLabel) > 0 {
		cdWorkflow.Spec.NodeSelector = impl.cdConfig.NodeLabel
	}
	//

	wfTemplate, err := json.Marshal(cdWorkflow)
	if err != nil {
		impl.Logger.Error(err)
	}
	impl.Logger.Debug("---->", string(wfTemplate))

	var wfClient v1alpha12.WorkflowInterface

	if workflowRequest.IsExtRun {
		serverUrl := env.Cluster.ServerUrl
		configMap := env.Cluster.Config
		bearerToken := configMap["bearer_token"]
		wfClient, err = impl.getRuntimeEnvClientInstance(workflowRequest.Namespace, bearerToken, serverUrl)
	}
	if wfClient == nil {
		wfClient, err = impl.getClientInstance(workflowRequest.Namespace)
		if err != nil {
			impl.Logger.Errorw("cannot build wf client", "err", err)
			return nil, err
		}
	}

	createdWf, err := wfClient.Create(context.Background(), &cdWorkflow, v1.CreateOptions{}) // submit the hello world workflow
	if err != nil {
		impl.Logger.Errorw("error in wf trigger", "err", err)
		return nil, err
	}
	impl.Logger.Debugw("workflow submitted: ", "name", createdWf.Name)
	impl.checkErr(err)
	return createdWf, err
}

func (impl *CdWorkflowServiceImpl) GetWorkflow(name string, namespace string, url string, token string, isExtRun bool) (*v1alpha1.Workflow, error) {
	impl.Logger.Debugw("getting wf", "name", name)
	var wfClient v1alpha12.WorkflowInterface
	var err error
	if isExtRun {
		wfClient, err = impl.getRuntimeEnvClientInstance(namespace, token, url)

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

func (impl *CdWorkflowServiceImpl) TerminateWorkflow(name string, namespace string, url string, token string, isExtRun bool) error {
	impl.Logger.Debugw("terminating wf", "name", name)
	var wfClient v1alpha12.WorkflowInterface
	var err error
	if isExtRun {
		wfClient, err = impl.getRuntimeEnvClientInstance(namespace, token, url)

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

func (impl *CdWorkflowServiceImpl) getRuntimeEnvClientInstance(namespace string, token string, host string) (v1alpha12.WorkflowInterface, error) {
	config := &rest.Config{
		Host:        host,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true,
		},
	}
	clientSet, err := versioned.NewForConfig(config)
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
