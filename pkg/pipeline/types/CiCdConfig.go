package types

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/caarlos0/env"
	blob_storage "github.com/devtron-labs/common-lib-private/blob-storage"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	v12 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type CiCdConfig struct {
	//from ciConfig
	DefaultCacheBucket               string                              `env:"DEFAULT_CACHE_BUCKET" envDefault:"ci-caching"`
	DefaultCacheBucketRegion         string                              `env:"DEFAULT_CACHE_BUCKET_REGION" envDefault:"us-east-2"`
	CiLogsKeyPrefix                  string                              `env:"CI_LOGS_KEY_PREFIX" envDxefault:"my-artifacts"`
	CiDefaultImage                   string                              `env:"DEFAULT_CI_IMAGE" envDefault:"686244538589.dkr.ecr.us-east-2.amazonaws.com/cirunner:47"`
	CiDefaultNamespace               string                              `env:"DEFAULT_NAMESPACE" envDefault:"devtron-ci"`
	CiDefaultTimeout                 int64                               `env:"DEFAULT_TIMEOUT" envDefault:"3600"`
	CiDefaultBuildLogsBucket         string                              `env:"DEFAULT_BUILD_LOGS_BUCKET" envDefault:"devtron-pro-ci-logs"`
	CiDefaultCdLogsBucketRegion      string                              `env:"DEFAULT_CD_LOGS_BUCKET_REGION" envDefault:"us-east-2"`
	CiLimitCpu                       string                              `env:"LIMIT_CI_CPU" envDefault:"0.5"`
	CiLimitMem                       string                              `env:"LIMIT_CI_MEM" envDefault:"3G"`
	CiReqCpu                         string                              `env:"REQ_CI_CPU" envDefault:"0.5"`
	CiReqMem                         string                              `env:"REQ_CI_MEM" envDefault:"3G"`
	CiTaintKey                       string                              `env:"CI_NODE_TAINTS_KEY" envDefault:""`
	CiTaintValue                     string                              `env:"CI_NODE_TAINTS_VALUE" envDefault:""`
	CiNodeLabelSelector              []string                            `env:"CI_NODE_LABEL_SELECTOR"`
	CacheLimit                       int64                               `env:"CACHE_LIMIT" envDefault:"5000000000"` // TODO: Add to default db config also
	CiDefaultBuildLogsKeyPrefix      string                              `env:"DEFAULT_BUILD_LOGS_KEY_PREFIX" envDefault:"arsenal-v1"`
	CiDefaultArtifactKeyPrefix       string                              `env:"DEFAULT_ARTIFACT_KEY_LOCATION" envDefault:"arsenal-v1/ci-artifacts"`
	CiWorkflowServiceAccount         string                              `env:"WORKFLOW_SERVICE_ACCOUNT" envDefault:"ci-runner"`
	ExternalCiApiSecret              string                              `env:"EXTERNAL_CI_API_SECRET" envDefault:"devtroncd-secret"`
	ExternalCiWebhookUrl             string                              `env:"EXTERNAL_CI_WEB_HOOK_URL" envDefault:""`
	ExternalCiPayload                string                              `env:"EXTERNAL_CI_PAYLOAD" envDefault:"{\"ciProjectDetails\":[{\"gitRepository\":\"https://github.com/vikram1601/getting-started-nodejs.git\",\"checkoutPath\":\"./abc\",\"commitHash\":\"239077135f8cdeeccb7857e2851348f558cb53d3\",\"commitTime\":\"2022-10-30T20:00:00\",\"branch\":\"master\",\"message\":\"Update README.md\",\"author\":\"User Name \"}],\"dockerImage\":\"445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2\"}"`
	CiArtifactLocationFormat         string                              `env:"CI_ARTIFACT_LOCATION_FORMAT" envDefault:"%d/%d.zip"`
	ImageScannerEndpoint             string                              `env:"IMAGE_SCANNER_ENDPOINT" envDefault:"http://image-scanner-new-demo-devtroncd-service.devtroncd:80"`
	CiDefaultAddressPoolBaseCidr     string                              `env:"CI_DEFAULT_ADDRESS_POOL_BASE_CIDR"`
	CiDefaultAddressPoolSize         int                                 `env:"CI_DEFAULT_ADDRESS_POOL_SIZE"`
	CiRunnerDockerMTUValue           int                                 `env:"CI_RUNNER_DOCKER_MTU_VALUE" envDefault:"-1"`
	IgnoreDockerCacheForCI           bool                                `env:"CI_IGNORE_DOCKER_CACHE"`
	VolumeMountsForCiJson            string                              `env:"CI_VOLUME_MOUNTS_JSON"`
	BuildPvcCachePath                string                              `env:"PRE_CI_CACHE_PATH" envDefault:"/devtroncd-cache"`
	DefaultPvcCachePath              string                              `env:"DOCKER_BUILD_CACHE_PATH" envDefault:"/var/lib/docker"`
	BuildxPvcCachePath               string                              `env:"BUILDX_CACHE_PATH" envDefault:"/var/lib/devtron/buildx"`
	UseBlobStorageConfigInCiWorkflow bool                                `env:"USE_BLOB_STORAGE_CONFIG_IN_CI_WORKFLOW" envDefault:"true"`
	DefaultTargetPlatform            string                              `env:"DEFAULT_TARGET_PLATFORM" envDefault:""`
	UseBuildx                        bool                                `env:"USE_BUILDX" envDefault:"false"`
	EnableBuildContext               bool                                `env:"ENABLE_BUILD_CONTEXT" envDefault:"false"`
	ImageRetryCount                  int                                 `env:"IMAGE_RETRY_COUNT" envDefault:"0"`
	ImageRetryInterval               int                                 `env:"IMAGE_RETRY_INTERVAL" envDefault:"5"` //image retry interval takes value in seconds
	CiWorkflowExecutorType           pipelineConfig.WorkflowExecutorType `env:"CI_WORKFLOW_EXECUTOR_TYPE" envDefault:"AWF"`
	BuildxK8sDriverOptions           string                              `env:"BUILDX_K8S_DRIVER_OPTIONS" envDefault:""`
	CIAutoTriggerBatchSize           int                                 `env:"CI_SUCCESS_AUTO_TRIGGER_BATCH_SIZE" envDefault:"1"`
	SkipCreatingEcrRepo              bool                                `env:"SKIP_CREATING_ECR_REPO" envDefault:"false"`
	MaxCiWorkflowRetries             int                                 `env:"MAX_CI_WORKFLOW_RETRIES" envDefault:"0"`

	//from CdConfig
	CdLimitCpu                       string                              `env:"CD_LIMIT_CI_CPU" envDefault:"0.5"`
	CdLimitMem                       string                              `env:"CD_LIMIT_CI_MEM" envDefault:"3G"`
	CdReqCpu                         string                              `env:"CD_REQ_CI_CPU" envDefault:"0.5"`
	CdReqMem                         string                              `env:"CD_REQ_CI_MEM" envDefault:"3G"`
	CdTaintKey                       string                              `env:"CD_NODE_TAINTS_KEY" envDefault:"dedicated"`
	ExternalCdTaintKey               string                              `env:"EXTERNAL_CD_NODE_TAINTS_KEY" envDefault:"dedicated"`
	UseExternalNode                  bool                                `env:"USE_EXTERNAL_NODE" envDefault:"false"`
	CdWorkflowServiceAccount         string                              `env:"CD_WORKFLOW_SERVICE_ACCOUNT" envDefault:"cd-runner"`
	CdDefaultBuildLogsKeyPrefix      string                              `env:"DEFAULT_BUILD_LOGS_KEY_PREFIX" `
	CdDefaultArtifactKeyPrefix       string                              `env:"DEFAULT_CD_ARTIFACT_KEY_LOCATION" `
	CdTaintValue                     string                              `env:"CD_NODE_TAINTS_VALUE" envDefault:"ci"`
	ExternalCdTaintValue             string                              `env:"EXTERNAL_CD_NODE_TAINTS_VALUE" envDefault:"ci"`
	CdDefaultBuildLogsBucket         string                              `env:"DEFAULT_BUILD_LOGS_BUCKET" `
	CdNodeLabelSelector              []string                            `env:"CD_NODE_LABEL_SELECTOR"`
	ExternalCdNodeLabelSelector      []string                            `env:"EXTERNAL_CD_NODE_LABEL_SELECTOR"`
	CdArtifactLocationFormat         string                              `env:"CD_ARTIFACT_LOCATION_FORMAT" envDefault:"%d/%d.zip"`
	CdDefaultNamespace               string                              `env:"DEFAULT_CD_NAMESPACE"`
	CdDefaultImage                   string                              `env:"DEFAULT_CI_IMAGE"`
	CdDefaultTimeout                 int64                               `env:"DEFAULT_CD_TIMEOUT" envDefault:"3600"`
	CdDefaultCdLogsBucketRegion      string                              `env:"DEFAULT_CD_LOGS_BUCKET_REGION" `
	WfControllerInstanceID           string                              `env:"WF_CONTROLLER_INSTANCE_ID" envDefault:"devtron-runner"`
	CdDefaultAddressPoolBaseCidr     string                              `env:"CD_DEFAULT_ADDRESS_POOL_BASE_CIDR"`
	CdDefaultAddressPoolSize         int                                 `env:"CD_DEFAULT_ADDRESS_POOL_SIZE"`
	ExposeCDMetrics                  bool                                `env:"EXPOSE_CD_METRICS" envDefault:"false"`
	UseBlobStorageConfigInCdWorkflow bool                                `env:"USE_BLOB_STORAGE_CONFIG_IN_CD_WORKFLOW" envDefault:"true"`
	CdWorkflowExecutorType           pipelineConfig.WorkflowExecutorType `env:"CD_WORKFLOW_EXECUTOR_TYPE" envDefault:"AWF"`
	TerminationGracePeriod           int                                 `env:"TERMINATION_GRACE_PERIOD_SECS" envDefault:"180"`
	CloningMode                      string                              `env:"CLONING_MODE" envDefault:"SHALLOW"`
	GitProviders                     string                              `env:"GIT_PROVIDERS" envDefault:"github,gitlab"`
	MaxCdWorkflowRunnerRetries       int                                 `env:"MAX_CD_WORKFLOW_RUNNER_RETRIES" envDefault:"0"`

	//common in both ciconfig and cd config
	Type                           string
	Mode                           string `env:"MODE" envDefault:"DEV"`
	OrchestratorHost               string `env:"ORCH_HOST" envDefault:"http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats"`
	OrchestratorToken              string `env:"ORCH_TOKEN" envDefault:""`
	ClusterConfig                  *rest.Config
	CloudProvider                  blob_storage.BlobStorageType `env:"BLOB_STORAGE_PROVIDER" envDefault:"S3"`
	BlobStorageEnabled             bool                         `env:"BLOB_STORAGE_ENABLED" envDefault:"false"`
	BlobStorageS3AccessKey         string                       `env:"BLOB_STORAGE_S3_ACCESS_KEY"`
	BlobStorageS3SecretKey         string                       `env:"BLOB_STORAGE_S3_SECRET_KEY"`
	BlobStorageS3Endpoint          string                       `env:"BLOB_STORAGE_S3_ENDPOINT"`
	BlobStorageS3EndpointInsecure  bool                         `env:"BLOB_STORAGE_S3_ENDPOINT_INSECURE" envDefault:"false"`
	BlobStorageS3BucketVersioned   bool                         `env:"BLOB_STORAGE_S3_BUCKET_VERSIONED" envDefault:"true"`
	BlobStorageGcpCredentialJson   string                       `env:"BLOB_STORAGE_GCP_CREDENTIALS_JSON"`
	AzureAccountName               string                       `env:"AZURE_ACCOUNT_NAME"`
	AzureGatewayUrl                string                       `env:"AZURE_GATEWAY_URL" envDefault:"http://devtron-minio.devtroncd:9000"`
	AzureGatewayConnectionInsecure bool                         `env:"AZURE_GATEWAY_CONNECTION_INSECURE" envDefault:"true"`
	AzureBlobContainerCiLog        string                       `env:"AZURE_BLOB_CONTAINER_CI_LOG"`
	AzureBlobContainerCiCache      string                       `env:"AZURE_BLOB_CONTAINER_CI_CACHE"`
	AzureAccountKey                string                       `env:"AZURE_ACCOUNT_KEY"`
	BuildLogTTLValue               int                          `env:"BUILD_LOG_TTL_VALUE_IN_SECS" envDefault:"3600"`
	BaseLogLocationPath            string                       `env:"BASE_LOG_LOCATION_PATH" envDefault:"/home/devtron/"`
	InAppLoggingEnabled            bool                         `env:"IN_APP_LOGGING_ENABLED" envDefault:"false"`
	BuildxProvenanceMode           string                       `env:"BUILDX_PROVENANCE_MODE" envDefault:""` //provenance is set to false if this flag is not set
}

type CiConfig struct {
	*CiCdConfig
}

type CdConfig struct {
	*CiCdConfig
}

type CiVolumeMount struct {
	Name               string `json:"name"`
	HostMountPath      string `json:"hostMountPath"`
	ContainerMountPath string `json:"containerMountPath"`
}

const (
	ExternalCiWebhookPath = "orchestrator/webhook/ext-ci"
	DevMode               = "DEV"
	ProdMode              = "PROD"
	CiConfigType          = "CiConfig"
	CdConfigType          = "CdConfig"
)

func GetCiConfig() (*CiConfig, error) {
	ciCdConfig := &CiCdConfig{}
	err := env.Parse(ciCdConfig)
	if err != nil {
		return nil, err
	}
	ciConfig := CiConfig{ciCdConfig}
	ciConfig.Type = CiConfigType
	return &ciConfig, nil
}
func GetCdConfig() (*CdConfig, error) {
	ciCdConfig := &CiCdConfig{}
	err := env.Parse(ciCdConfig)
	if err != nil {
		return nil, err
	}
	cdConfig := CdConfig{ciCdConfig}
	ciCdConfig.Type = CdConfigType
	return &cdConfig, nil

}

func GetCiCdConfig() (*CiCdConfig, error) {
	cfg := &CiCdConfig{}
	err := env.Parse(cfg)
	if cfg.Mode == DevMode {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		kubeconfig := flag.String("kubeconfig", filepath.Join(usr.HomeDir, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
		flag.Parse()
		cfg.ClusterConfig, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
		if err != nil {
			return nil, err
		}
	} else {
		cfg.ClusterConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}
	//validation for supported cloudproviders
	if cfg.BlobStorageEnabled && cfg.CloudProvider != BLOB_STORAGE_S3 && cfg.CloudProvider != BLOB_STORAGE_AZURE &&
		cfg.CloudProvider != BLOB_STORAGE_GCP && cfg.CloudProvider != BLOB_STORAGE_MINIO {
		return nil, fmt.Errorf("unsupported blob storage provider: %s", cfg.CloudProvider)
	}

	return cfg, err
}

func GetNodeLabel(cfg *CiCdConfig, pipelineType bean.WorkflowPipelineType, isExt bool) (map[string]string, error) {
	node := []string{}
	if pipelineType == bean.CI_WORKFLOW_PIPELINE_TYPE || pipelineType == bean.JOB_WORKFLOW_PIPELINE_TYPE {
		node = cfg.CiNodeLabelSelector
	}
	if pipelineType == bean.CD_WORKFLOW_PIPELINE_TYPE {
		if isExt && cfg.UseExternalNode {
			node = cfg.ExternalCdNodeLabelSelector
		} else {
			node = cfg.CdNodeLabelSelector
		}
	}

	nodeLabel := make(map[string]string)
	for _, l := range node {
		if l == "" {
			continue
		}
		kv := strings.Split(l, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid ci node label selector %s, it must be in form key=value, key2=val2", kv)
		}
		nodeLabel[kv[0]] = kv[1]
	}
	return nodeLabel, nil
}
func (impl *CiCdConfig) GetDefaultImage() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiDefaultImage
	case CdConfigType:
		return impl.CdDefaultImage
	default:
		return ""
	}

}
func (impl *CiCdConfig) GetDefaultNamespace() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiDefaultNamespace
	case CdConfigType:
		return impl.CdDefaultNamespace
	default:
		return ""
	}
}
func (impl *CiCdConfig) GetDefaultTimeout() int64 {
	switch impl.Type {
	case CiConfigType:
		return impl.CiDefaultTimeout
	case CdConfigType:
		return impl.CdDefaultTimeout
	default:
		return 0
	}
}
func (impl *CiCdConfig) GetDefaultBuildLogsBucket() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiDefaultBuildLogsBucket
	case CdConfigType:
		return impl.CdDefaultBuildLogsBucket
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetDefaultCdLogsBucketRegion() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiDefaultCdLogsBucketRegion
	case CdConfigType:
		return impl.CdDefaultCdLogsBucketRegion
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetLimitCpu() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiLimitCpu
	case CdConfigType:
		return impl.CdLimitCpu
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetLimitMem() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiLimitMem
	case CdConfigType:
		return impl.CdLimitMem
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetReqCpu() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiReqCpu
	case CdConfigType:
		return impl.CdReqCpu
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetReqMem() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiReqMem
	case CdConfigType:
		return impl.CdReqMem
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetTaintKey() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiTaintKey
	case CdConfigType:
		return impl.CdTaintKey
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetTaintValue() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiTaintValue
	case CdConfigType:
		return impl.CdTaintValue
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetDefaultBuildLogsKeyPrefix() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiDefaultBuildLogsKeyPrefix
	case CdConfigType:
		return impl.CdDefaultBuildLogsKeyPrefix
	default:
		return ""
	}
}
func (impl *CiCdConfig) GetDefaultArtifactKeyPrefix() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiDefaultArtifactKeyPrefix
	case CdConfigType:
		return impl.CdDefaultArtifactKeyPrefix
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetWorkflowServiceAccount() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiWorkflowServiceAccount
	case CdConfigType:
		return impl.CdWorkflowServiceAccount
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetArtifactLocationFormat() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiArtifactLocationFormat
	case CdConfigType:
		return impl.CdArtifactLocationFormat
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetDefaultAddressPoolBaseCidr() string {
	switch impl.Type {
	case CiConfigType:
		return impl.CiDefaultAddressPoolBaseCidr
	case CdConfigType:
		return impl.CdDefaultAddressPoolBaseCidr
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetDefaultAddressPoolSize() int {
	switch impl.Type {
	case CiConfigType:
		return impl.CiDefaultAddressPoolSize
	case CdConfigType:
		return impl.CdDefaultAddressPoolSize
	default:
		return 0
	}
}

func (impl *CiCdConfig) GetWorkflowExecutorType() pipelineConfig.WorkflowExecutorType {
	switch impl.Type {
	case CiConfigType:
		return impl.CiWorkflowExecutorType
	case CdConfigType:
		return impl.CdWorkflowExecutorType
	default:
		return ""
	}
}

func (impl *CiCdConfig) WorkflowRetriesEnabled() bool {
	switch impl.Type {
	case CiConfigType:
		return impl.MaxCiWorkflowRetries > 0
	case CdConfigType:
		return impl.MaxCdWorkflowRunnerRetries > 0
	default:
		return false
	}
}

func (impl *CiCdConfig) GetWorkflowVolumeAndVolumeMounts() ([]v12.Volume, []v12.VolumeMount, error) {
	var volumes []v12.Volume
	var volumeMounts []v12.VolumeMount
	volumeMountsForCiJson := impl.VolumeMountsForCiJson
	if len(volumeMountsForCiJson) > 0 {
		var volumeMountsForCi []CiVolumeMount
		// Unmarshal or Decode the JSON to the interface.
		err := json.Unmarshal([]byte(volumeMountsForCiJson), &volumeMountsForCi)
		if err != nil {
			return nil, nil, err
		}

		for _, volumeMountForCi := range volumeMountsForCi {
			volumes = append(volumes, getWorkflowVolume(volumeMountForCi))
			volumeMounts = append(volumeMounts, getWorkflowVolumeMounts(volumeMountForCi))
		}
	}
	return volumes, volumeMounts, nil
}

func getWorkflowVolume(volumeMountForCi CiVolumeMount) v12.Volume {
	hostPathDirectoryOrCreate := v12.HostPathDirectoryOrCreate

	return v12.Volume{
		Name: volumeMountForCi.Name,
		VolumeSource: v12.VolumeSource{
			HostPath: &v12.HostPathVolumeSource{
				Path: volumeMountForCi.HostMountPath,
				Type: &hostPathDirectoryOrCreate,
			},
		},
	}

}

func getWorkflowVolumeMounts(volumeMountForCi CiVolumeMount) v12.VolumeMount {
	return v12.VolumeMount{
		Name:      volumeMountForCi.Name,
		MountPath: volumeMountForCi.ContainerMountPath,
	}
}

const BLOB_STORAGE_AZURE = "AZURE"

const BLOB_STORAGE_S3 = "S3"

const BLOB_STORAGE_GCP = "GCP"

const BLOB_STORAGE_MINIO = "MINIO"

type ArtifactsForCiJob struct {
	Artifacts []string `json:"artifacts"`
}

type GitTriggerInfoResponse struct {
	CiMaterials      []pipelineConfig.CiPipelineMaterialResponse `json:"ciMaterials"`
	TriggeredByEmail string                                      `json:"triggeredByEmail"`
	LastDeployedTime string                                      `json:"lastDeployedTime,omitempty"`
	AppId            int                                         `json:"appId"`
	AppName          string                                      `json:"appName"`
	EnvironmentId    int                                         `json:"environmentId"`
	EnvironmentName  string                                      `json:"environmentName"`
	Default          bool                                        `json:"default,omitempty"`
	ImageTaggingData ImageTaggingResponseDTO                     `json:"imageTaggingData"`
	Image            string                                      `json:"image"`
}

type Trigger struct {
	PipelineId                int
	CommitHashes              map[int]pipelineConfig.GitCommit
	CiMaterials               []*pipelineConfig.CiPipelineMaterial
	TriggeredBy               int32
	InvalidateCache           bool
	ExtraEnvironmentVariables map[string]string // extra env variables which will be used for CI
	EnvironmentId             int
	PipelineType              string
	CiArtifactLastFetch       time.Time
	ReferenceCiWorkflowId     int
}

func (obj Trigger) BuildTriggerObject(refCiWorkflow *pipelineConfig.CiWorkflow,
	ciMaterials []*pipelineConfig.CiPipelineMaterial, triggeredBy int32,
	invalidateCache bool, extraEnvironmentVariables map[string]string,
	pipelineType string) {

	obj.PipelineId = refCiWorkflow.CiPipelineId
	obj.CommitHashes = refCiWorkflow.GitTriggers
	obj.CiMaterials = ciMaterials
	obj.TriggeredBy = triggeredBy
	obj.InvalidateCache = invalidateCache
	obj.EnvironmentId = refCiWorkflow.EnvironmentId
	obj.ReferenceCiWorkflowId = refCiWorkflow.Id
	obj.InvalidateCache = invalidateCache
	obj.ExtraEnvironmentVariables = extraEnvironmentVariables
	obj.PipelineType = pipelineType

}

type BuildLogRequest struct {
	PipelineId        int
	WorkflowId        int
	PodName           string
	LogsFilePath      string
	Namespace         string
	CloudProvider     blob_storage.BlobStorageType
	AwsS3BaseConfig   *blob_storage.AwsS3BaseConfig
	AzureBlobConfig   *blob_storage.AzureBlobBaseConfig
	GcpBlobBaseConfig *blob_storage.GcpBlobBaseConfig
	MinioEndpoint     string
}
