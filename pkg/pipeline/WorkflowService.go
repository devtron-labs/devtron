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
	"time"

	"github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/argoproj/argo/pkg/client/clientset/versioned"
	v1alpha12 "github.com/argoproj/argo/pkg/client/clientset/versioned/typed/workflow/v1alpha1"
	"github.com/argoproj/argo/workflow/util"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	"go.uber.org/zap"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

type WorkflowService interface {
	SubmitWorkflow(workflowRequest *WorkflowRequest) (*v1alpha1.Workflow, error)
	DeleteWorkflow(wfName string, namespace string) error
	GetWorkflow(name string, namespace string) (*v1alpha1.Workflow, error)
	ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error)
	UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error)
	TerminateWorkflow(name string, namespace string) error
}

type CiCdTriggerEvent struct {
	Type      string             `json:"type"`
	CiRequest *WorkflowRequest   `json:"ciRequest"`
	CdRequest *CdWorkflowRequest `json:"cdRequest"`
}

type WorkflowServiceImpl struct {
	Logger   *zap.SugaredLogger
	config   *rest.Config
	ciConfig *CiConfig
}

type WorkflowRequest struct {
	WorkflowNamePrefix       string             `json:"workflowNamePrefix"`
	PipelineName             string             `json:"pipelineName"`
	PipelineId               int                `json:"pipelineId"`
	DockerImageTag           string             `json:"dockerImageTag"`
	DockerRegistryType       string             `json:"dockerRegistryType"`
	DockerRegistryURL        string             `json:"dockerRegistryURL"`
	DockerConnection         string             `json:"dockerConnection"`
	DockerCert               string             `json:"dockerCert"`
	DockerBuildArgs          string             `json:"dockerBuildArgs"`
	DockerRepository         string             `json:"dockerRepository"`
	DockerFileLocation       string             `json:"dockerfileLocation"`
	DockerUsername           string             `json:"dockerUsername"`
	DockerPassword           string             `json:"dockerPassword"`
	AwsRegion                string             `json:"awsRegion"`
	AccessKey                string             `json:"accessKey"`
	SecretKey                string             `json:"secretKey"`
	CiCacheLocation          string             `json:"ciCacheLocation"`
	CiCacheRegion            string             `json:"ciCacheRegion"`
	CiCacheFileName          string             `json:"ciCacheFileName"`
	CiProjectDetails         []CiProjectDetails `json:"ciProjectDetails"`
	ContainerResources       ContainerResources `json:"containerResources"`
	ActiveDeadlineSeconds    int64              `json:"activeDeadlineSeconds"`
	CiImage                  string             `json:"ciImage"`
	Namespace                string             `json:"namespace"`
	WorkflowId               int                `json:"workflowId"`
	TriggeredBy              int32              `json:"triggeredBy"`
	CacheLimit               int64              `json:"cacheLimit"`
	BeforeDockerBuildScripts []*bean.CiScript   `json:"beforeDockerBuildScripts"`
	AfterDockerBuildScripts  []*bean.CiScript   `json:"afterDockerBuildScripts"`
	CiArtifactLocation       string             `json:"ciArtifactLocation"`
	InvalidateCache          bool               `json:"invalidateCache"`
	ScanEnabled              bool               `json:"scanEnabled"`
	CloudProvider            string             `json:"cloudProvider"`
	AzureBlobConfig          *AzureBlobConfig   `json:"azureBlobConfig"`
	MinioEndpoint            string             `json:"minioEndpoint"`
}

const BLOB_STORAGE_AZURE = "AZURE"
const BLOB_STORAGE_S3 = "S3"
const BLOB_STORAGE_GCP = "GCP"
const BLOB_STORAGE_MINIO = "MINIO"

type AzureBlobConfig struct {
	Enabled              bool   `json:"enabled"`
	AccountName          string `json:"accountName"`
	BlobContainerCiLog   string `json:"blobContainerCiLog"`
	BlobContainerCiCache string `json:"blobContainerCiCache"`
	AccountKey           string `json:"accountKey"`
}

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

// Used for default values
/*func NewContainerResources() ContainerResources {
	return ContainerResources{
		MinCpu:        "",
		MaxCpu:        "0.5",
		MinStorage:    "",
		MaxStorage:    "",
		MinEphStorage: "",
		MaxEphStorage: "",
		MinMem:        "",
		MaxMem:        "200Mi",
	}
}*/

type CiProjectDetails struct {
	GitRepository   string    `json:"gitRepository"`
	MaterialName    string    `json:"materialName"`
	CheckoutPath    string    `json:"checkoutPath"`
	FetchSubmodules bool      `json:"fetchSubmodules"`
	CommitHash      string    `json:"commitHash"`
	GitTag          string    `json:"gitTag"`
	CommitTime      time.Time `json:"commitTime"`
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
	UserName      string              `json:"userName"`
	Password      string              `json:"password"`
	SshPrivateKey string              `json:"sshPrivateKey"`
	AccessToken   string              `json:"accessToken"`
	AuthMode      repository.AuthMode `json:"authMode"`
}

func NewWorkflowServiceImpl(Logger *zap.SugaredLogger, ciConfig *CiConfig) *WorkflowServiceImpl {
	return &WorkflowServiceImpl{Logger: Logger, config: ciConfig.ClusterConfig, ciConfig: ciConfig}
}

const ciEvent = "CI"
const cdStage = "CD"

func (impl *WorkflowServiceImpl) SubmitWorkflow(workflowRequest *WorkflowRequest) (*v1alpha1.Workflow, error) {
	containerEnvVariables := []v12.EnvVar{{Name: "IMAGE_SCANNER_ENDPOINT", Value: impl.ciConfig.ImageScannerEndpoint}}
	if impl.ciConfig.CloudProvider == BLOB_STORAGE_MINIO {
		miniCred := []v12.EnvVar{{Name: "AWS_ACCESS_KEY_ID", Value: impl.ciConfig.MinioAccessKey}, {Name: "AWS_SECRET_ACCESS_KEY", Value: impl.ciConfig.MinioSecretKey}}
		containerEnvVariables = append(containerEnvVariables, miniCred...)
	}

	ciCdTriggerEvent := CiCdTriggerEvent{
		Type:      ciEvent,
		CiRequest: workflowRequest,
	}

	workflowJson, err := json.Marshal(&ciCdTriggerEvent)
	if err != nil {
		impl.Logger.Errorw("err", err)
		return nil, err
	}
	impl.Logger.Debugw("workflowRequest ---->", "workflowJson", string(workflowJson))

	wfClient, err := impl.getClientInstance(workflowRequest.Namespace)
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return nil, err
	}

	privileged := true
	archiveLogs := true

	limitCpu := impl.ciConfig.LimitCpu
	limitMem := impl.ciConfig.LimitMem

	reqCpu := impl.ciConfig.ReqCpu
	reqMem := impl.ciConfig.ReqMem
	ttl := int32(600)

	var (
		ciWorkflow = v1alpha1.Workflow{
			ObjectMeta: v1.ObjectMeta{
				GenerateName: workflowRequest.WorkflowNamePrefix + "-",
			},
			Spec: v1alpha1.WorkflowSpec{
				ServiceAccountName: impl.ciConfig.WorkflowServiceAccount,
				//NodeSelector:            map[string]string{impl.ciConfig.TaintKey: impl.ciConfig.TaintValue},
				//Tolerations:             []v12.Toleration{{Key: impl.ciConfig.TaintKey, Value: impl.ciConfig.TaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}},
				Entrypoint:              "ci",
				TTLSecondsAfterFinished: &ttl,
				Templates: []v1alpha1.Template{
					{
						Name: "ci",
						Container: &v12.Container{
							Env:   containerEnvVariables,
							Image: workflowRequest.CiImage,
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
							Ports: []v12.ContainerPort{{
								//exposed for user specific data from ci container
								Name:          "app-data",
								ContainerPort: 9102,
							}},
						},
						ActiveDeadlineSeconds: &workflowRequest.ActiveDeadlineSeconds,
						ArchiveLocation: &v1alpha1.ArtifactLocation{
							ArchiveLogs: &archiveLogs,
						},
					},
				},
			},
		}
	)
	if impl.ciConfig.TaintKey != "" || impl.ciConfig.TaintValue != "" {
		ciWorkflow.Spec.Tolerations = []v12.Toleration{{Key: impl.ciConfig.TaintKey, Value: impl.ciConfig.TaintValue, Operator: v12.TolerationOpEqual, Effect: v12.TaintEffectNoSchedule}}
	}
	if len(impl.ciConfig.NodeLabel) > 0 {
		ciWorkflow.Spec.NodeSelector = impl.ciConfig.NodeLabel
	}

	wfTemplate, err := json.Marshal(ciWorkflow)
	if err != nil {
		impl.Logger.Errorw("marshal error", "err", err)
	}
	impl.Logger.Debug("---->", string(wfTemplate))

	createdWf, err := wfClient.Create(&ciWorkflow) // submit the hello world workflow
	impl.Logger.Debug("workflow submitted: " + createdWf.Name)
	impl.checkErr(err)
	return createdWf, err
}

func (impl *WorkflowServiceImpl) getClientInstance(namespace string) (v1alpha12.WorkflowInterface, error) {
	clientSet, err := versioned.NewForConfig(impl.config)
	if err != nil {
		impl.Logger.Errorw("err on get client instance", "err", err)
		return nil, err
	}
	wfClient := clientSet.ArgoprojV1alpha1().Workflows(namespace) // create the workflow client
	return wfClient, nil
}

func (impl *WorkflowServiceImpl) GetWorkflow(name string, namespace string) (*v1alpha1.Workflow, error) {
	impl.Logger.Debug("getting wf", name)
	wfClient, err := impl.getClientInstance(namespace)
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return nil, err
	}
	workflow, err := wfClient.Get(name, v1.GetOptions{})
	return workflow, err
}

func (impl *WorkflowServiceImpl) TerminateWorkflow(name string, namespace string) error {
	impl.Logger.Debugw("terminating wf", "name", name)
	wfClient, err := impl.getClientInstance(namespace)
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return err
	}
	err = util.TerminateWorkflow(wfClient, name)
	return err
}

func (impl *WorkflowServiceImpl) UpdateWorkflow(wf *v1alpha1.Workflow) (*v1alpha1.Workflow, error) {
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
	impl.Logger.Debugw("updated wf", "name", wf.Name)
	return updatedWf, err
}

func (impl *WorkflowServiceImpl) ListAllWorkflows(namespace string) (*v1alpha1.WorkflowList, error) {
	impl.Logger.Debug("listing all wfs")
	wfClient, err := impl.getClientInstance(namespace)
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return nil, err
	}
	workflowList, err := wfClient.List(v1.ListOptions{})
	return workflowList, err
}

func (impl *WorkflowServiceImpl) DeleteWorkflow(wfName string, namespace string) error {
	impl.Logger.Debugw("deleting wf", "name", wfName)
	wfClient, err := impl.getClientInstance(namespace)
	if err != nil {
		impl.Logger.Errorw("cannot build wf client", "err", err)
		return err
	}
	err = wfClient.Delete(wfName, &v1.DeleteOptions{})
	return err
}

func (impl *WorkflowServiceImpl) checkErr(err error) {
	if err != nil {
		impl.Logger.Errorw("error", "error:", err)
	}
}
