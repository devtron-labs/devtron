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
	"encoding/json"
	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/pkg/client/clientset/versioned"
	v1alpha12 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/argoproj/argo/workflow/util"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"strconv"
)

type CdWorkflowService interface {
	SubmitWorkflow(workflowRequest *CdWorkflowRequest, pipeline *pipelineConfig.Pipeline, env *cluster.Environment) (*v1alpha1.Workflow, error)
	DeleteWorkflow(wfName string, namespace string) error
	GetWorkflow(name string, namespace string, url string, token string, isExtRun bool) (*v1alpha1.Workflow, error)
	ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error)
	UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error)
	TerminateWorkflow(name string, namespace string, url string, token string, isExtRun bool) error
}

const CD_WORKFLOW_NAME = "cd"

type CdWorkflowServiceImpl struct {
	Logger        *zap.SugaredLogger
	config        *rest.Config
	cdConfig      *CdConfig
	appService    app.AppService
	envRepository cluster.EnvironmentRepository
}

type CdWorkflowRequest struct {
	AppId                     int                `json:"appId"`
	EnvironmentId             int                `json:"envId"`
	WorkflowId                int                `json:"workflowId"`
	WorkflowRunnerId          int                `json:"workflowRunnerId"`
	CdPipelineId              int                `json:"cdPipelineId"`
	TriggeredBy               int32              `json:"triggeredBy"`
	StageYaml                 string             `json:"stageYaml"`
	ArtifactLocation          string             `json:"artifactLocation"`
	CiProjectDetails          []CiProjectDetails `json:"ciProjectDetails"`
	CiArtifactDTO             CiArtifactDTO      `json:"ciArtifactDTO"`
	Namespace                 string             `json:"namespace"`
	WorkflowNamePrefix        string             `json:"workflowNamePrefix"`
	CdImage                   string             `json:"cdImage"`
	ActiveDeadlineSeconds     int64              `json:"activeDeadlineSeconds"`
	StageType                 string             `json:"stageType"`
	DockerUsername            string             `json:"dockerUsername"`
	DockerPassword            string             `json:"dockerPassword"`
	AwsRegion                 string             `json:"awsRegion"`
	SecretKey                 string             `json:"secretKey"`
	AccessKey                 string             `json:"accessKey"`
	DockerRegistryType        string             `json:"dockerRegistryType"`
	DockerRegistryURL         string             `json:"dockerRegistryURL"`
	OrchestratorHost          string             `json:"orchestratorHost"`
	OrchestratorToken         string             `json:"orchestratorToken"`
	IsExtRun                  bool               `json:"isExtRun"`
	ExtraEnvironmentVariables map[string]string  `json:"extraEnvironmentVariables"`
}

const PRE = "PRE"
const POST = "POST"

func NewCdWorkflowServiceImpl(Logger *zap.SugaredLogger, envRepository cluster.EnvironmentRepository, cdConfig *CdConfig, appService app.AppService) *CdWorkflowServiceImpl {
	return &CdWorkflowServiceImpl{Logger: Logger, config: cdConfig.ClusterConfig,
		cdConfig: cdConfig, appService: appService, envRepository: envRepository}
}

func (impl *CdWorkflowServiceImpl) SubmitWorkflow(workflowRequest *CdWorkflowRequest, pipeline *pipelineConfig.Pipeline, env *cluster.Environment) (*v1alpha1.Workflow, error) {
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
	archiveLogs := true

	limitCpu := impl.cdConfig.LimitCpu
	limitMem := impl.cdConfig.LimitMem

	reqCpu := impl.cdConfig.ReqCpu
	reqMem := impl.cdConfig.ReqMem
	ttl := int32(300)

	var volumes []v12.Volume
	var steps [][]v1alpha1.WorkflowStep

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

	entryPoint := CD_WORKFLOW_NAME
	if len(configMaps.Maps) > 0 {
		entryPoint = "cd-stages-with-env"
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

			steps = append(steps, []v1alpha1.WorkflowStep{
				{
					Name:     "create-env-cm-" + strconv.Itoa(i),
					Template: "cm-" + strconv.Itoa(i),
				},
			})
		}
	}

	if len(secrets.Secrets) > 0 {
		entryPoint = "cd-stages-with-env"
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

			steps = append(steps, []v1alpha1.WorkflowStep{
				{
					Name:     "create-env-sec-" + strconv.Itoa(i),
					Template: "sec-" + strconv.Itoa(i),
				},
			})
		}
	}

	var templates []v1alpha1.Template
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

	steps = append(steps, []v1alpha1.WorkflowStep{
		{
			Name:     "run-wf",
			Template: CD_WORKFLOW_NAME,
		},
	})
	templates = append(templates, v1alpha1.Template{
		Name:  "cd-stages-with-env",
		Steps: steps,
	})

	templates = append(templates, v1alpha1.Template{
		Name: "cd",
		Container: &v12.Container{
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
		ActiveDeadlineSeconds: &workflowRequest.ActiveDeadlineSeconds,
		ArchiveLocation: &v1alpha1.ArtifactLocation{
			ArchiveLogs: &archiveLogs,
		},
	})

	var cdTemplate v1alpha1.Template
	for _, cm := range configMaps.Maps {
		for _, t := range templates {
			if t.Name == "cd" {
				cdTemplate = t
				break
			}
		}
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
		for i, t := range templates {
			if t.Name == "cd" {
				templates[i] = cdTemplate
				break
			}
		}
	}

	for _, s := range secrets.Secrets {
		for _, t := range templates {
			if t.Name == "cd" {
				cdTemplate = t
				break
			}
		}
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
		for i, t := range templates {
			if t.Name == "cd" {
				templates[i] = cdTemplate
				break
			}
		}
	}

	var (
		cdWorkflow = v1alpha1.Workflow{
			ObjectMeta: v1.ObjectMeta{
				GenerateName: workflowRequest.WorkflowNamePrefix + "-",
				Annotations:  map[string]string{"workflows.argoproj.io/controller-instanceid": impl.cdConfig.WfControllerInstanceID},
			},
			Spec: v1alpha1.WorkflowSpec{
				ServiceAccountName:      impl.cdConfig.WorkflowServiceAccount,
				NodeSelector:            map[string]string{impl.cdConfig.TaintKey: impl.cdConfig.TaintValue},
				Tolerations:             []v12.Toleration{{Key: impl.cdConfig.TaintKey, Value: impl.cdConfig.TaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}},
				Entrypoint:              entryPoint,
				TTLSecondsAfterFinished: &ttl,
				Templates:               templates,
				Volumes:                 volumes,
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

	createdWf, err := wfClient.Create(&cdWorkflow) // submit the hello world workflow
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
	workflow, err := wfClient.Get(name, v1.GetOptions{})
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
	err = util.TerminateWorkflow(wfClient, name)
	return err
}

func (impl *CdWorkflowServiceImpl) UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
	impl.Logger.Debugw("updating wf", "name", wf.Name)
	wfClient, err := impl.getClientInstance(wf.Namespace)
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return nil, err
	}
	updatedWf, err := wfClient.Update(wf)
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
	workflowList, err := wfClient.List(v1.ListOptions{})
	return workflowList, err
}

func (impl *CdWorkflowServiceImpl) DeleteWorkflow(wfName string, namespace string) error {
	wfClient, err := impl.getClientInstance(namespace)
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return err
	}
	err = wfClient.Delete(wfName, &v1.DeleteOptions{})
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
