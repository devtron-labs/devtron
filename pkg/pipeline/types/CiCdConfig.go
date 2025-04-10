/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/caarlos0/env"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	bean2 "github.com/devtron-labs/common-lib/utils/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/pkg/bean/common"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/bean"
	v12 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type CancelWfRequestDto struct {
	ExecutorType         cdWorkflow.WorkflowExecutorType
	WorkflowName         string
	Namespace            string
	RestConfig           *rest.Config
	IsExt                bool
	Environment          *repository.Environment
	ForceAbort           bool
	WorkflowGenerateName string
}

// build infra configurations like ciTimeout,ciCpuLimit,ciMemLimit,ciCpuReq,ciMemReq are being managed by infraConfig service

// CATEGORY=CI_RUNNER
type CiCdConfig struct {
	// from ciConfig
	DefaultCacheBucket           string   `env:"DEFAULT_CACHE_BUCKET" envDefault:"ci-caching" description:"Bucket name for build cache" `
	DefaultCacheBucketRegion     string   `env:"DEFAULT_CACHE_BUCKET_REGION" envDefault:"us-east-2" description:"Build Cache bucket region" `
	CiLogsKeyPrefix              string   `env:"CI_LOGS_KEY_PREFIX" envDxefault:"my-artifacts" description:"Prefix for build logs"`
	CiDefaultImage               string   `env:"DEFAULT_CI_IMAGE" envDefault:"686244538589.dkr.ecr.us-east-2.amazonaws.com/cirunner:47" description:"To pass the ci-runner image"`
	CiDefaultNamespace           string   `env:"DEFAULT_NAMESPACE" envDefault:"devtron-ci" description:"Timeout for CI to be completed"`
	CiDefaultBuildLogsBucket     string   `env:"DEFAULT_BUILD_LOGS_BUCKET" envDefault:"devtron-pro-ci-logs"`
	CiDefaultCdLogsBucketRegion  string   `env:"DEFAULT_CD_LOGS_BUCKET_REGION" envDefault:"us-east-2"`
	CiTaintKey                   string   `env:"CI_NODE_TAINTS_KEY" envDefault:"" description:"Toleration key for CI"`
	CiTaintValue                 string   `env:"CI_NODE_TAINTS_VALUE" envDefault:"" description:"Toleration value for CI" `
	CiNodeLabelSelector          []string `env:"CI_NODE_LABEL_SELECTOR" description:"Node label selector for  CI"`
	CacheLimit                   int64    `env:"CACHE_LIMIT" envDefault:"5000000000" description:"Cache limit."` // TODO: Add to default db config also
	CiDefaultBuildLogsKeyPrefix  string   `env:"DEFAULT_BUILD_LOGS_KEY_PREFIX" envDefault:"arsenal-v1" description:"Bucket prefix for build logs"`
	CiDefaultArtifactKeyPrefix   string   `env:"DEFAULT_ARTIFACT_KEY_LOCATION" envDefault:"arsenal-v1/ci-artifacts" description:"Key location for artifacts being created"`
	CiWorkflowServiceAccount     string   `env:"WORKFLOW_SERVICE_ACCOUNT" envDefault:"ci-runner"`
	ExternalCiApiSecret          string   `env:"EXTERNAL_CI_API_SECRET" envDefault:"devtroncd-secret" description:"External CI API secret."`
	ExternalCiWebhookUrl         string   `env:"EXTERNAL_CI_WEB_HOOK_URL" envDefault:"" description:"default is {{HOST_URL}}/orchestrator/webhook/ext-ci. It is used for external ci."`
	ExternalCiPayload            string   `env:"EXTERNAL_CI_PAYLOAD" envDefault:"{\"ciProjectDetails\":[{\"gitRepository\":\"https://github.com/vikram1601/getting-started-nodejs.git\",\"checkoutPath\":\"./abc\",\"commitHash\":\"239077135f8cdeeccb7857e2851348f558cb53d3\",\"commitTime\":\"2022-10-30T20:00:00\",\"branch\":\"master\",\"message\":\"Update README.md\",\"author\":\"User Name \"}],\"dockerImage\":\"445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2\"}" description:"External CI payload with project details."`
	ImageScannerEndpoint         string   `env:"IMAGE_SCANNER_ENDPOINT" envDefault:"http://image-scanner-new-demo-devtroncd-service.devtroncd:80" description:"Image-scanner micro-service URL"`
	CiDefaultAddressPoolBaseCidr string   `env:"CI_DEFAULT_ADDRESS_POOL_BASE_CIDR" description:"To pass the IP cidr for CI"`
	CiDefaultAddressPoolSize     int      `env:"CI_DEFAULT_ADDRESS_POOL_SIZE" description:"The subnet size to allocate from the base pool for CI"`
	CiRunnerDockerMTUValue       int      `env:"CI_RUNNER_DOCKER_MTU_VALUE" envDefault:"-1" description:"this is to control the bytes of inofrmation passed in a network packet in ci-runner.  default is -1 (defaults to the underlying node mtu value)"`
	//Deprecated: use WorkflowCacheConfig instead
	IgnoreDockerCacheForCI           bool                            `env:"CI_IGNORE_DOCKER_CACHE" description:"Ignoring docker cache "`
	WorkflowCacheConfig              string                          `env:"WORKFLOW_CACHE_CONFIG" envDefault:"{}" description:"flag is used to configure how Docker caches are handled during a CI/CD "`
	VolumeMountsForCiJson            string                          `env:"CI_VOLUME_MOUNTS_JSON" description:"additional volume mount data for CI and JOB"`
	BuildPvcCachePath                string                          `env:"PRE_CI_CACHE_PATH" envDefault:"/devtroncd-cache" description:"Cache path for Pre CI tasks"`
	DefaultPvcCachePath              string                          `env:"DOCKER_BUILD_CACHE_PATH" envDefault:"/var/lib/docker" description:"Path to store cache of docker build  (/var/lib/docker-> for legacy docker build, /var/lib/devtron-> for buildx)"`
	BuildxPvcCachePath               string                          `env:"BUILDX_CACHE_PATH" envDefault:"/var/lib/devtron/buildx" description:"Path for the buildx cache"`
	UseBlobStorageConfigInCiWorkflow bool                            `env:"USE_BLOB_STORAGE_CONFIG_IN_CI_WORKFLOW" envDefault:"true" description:"To enable blob storage in pre and post ci"`
	DefaultTargetPlatform            string                          `env:"DEFAULT_TARGET_PLATFORM" envDefault:"" description:"Default architecture for buildx"`
	UseBuildx                        bool                            `env:"USE_BUILDX" envDefault:"false" description:"To enable buildx feature globally"`
	EnableBuildContext               bool                            `env:"ENABLE_BUILD_CONTEXT" envDefault:"false" description:"To Enable build context in Devtron."`
	ImageRetryCount                  int                             `env:"IMAGE_RETRY_COUNT" envDefault:"0" description:"push artifact(image) in ci retry count "`
	ImageRetryInterval               int                             `env:"IMAGE_RETRY_INTERVAL" envDefault:"5" description:"image retry interval takes value in seconds"` // image retry interval takes value in seconds
	CiWorkflowExecutorType           cdWorkflow.WorkflowExecutorType `env:"CI_WORKFLOW_EXECUTOR_TYPE" envDefault:"AWF" description:"Executor type for CI(AWF,System)"`
	BuildxK8sDriverOptions           string                          `env:"BUILDX_K8S_DRIVER_OPTIONS" envDefault:"" description:"To enable the k8s driver and pass args for k8s driver in buildx"`
	CIAutoTriggerBatchSize           int                             `env:"CI_SUCCESS_AUTO_TRIGGER_BATCH_SIZE" envDefault:"1" description:"this is to control the no of linked pipelines should be hanled in one go when a ci-success event of an parent ci is received"`
	SkipCreatingEcrRepo              bool                            `env:"SKIP_CREATING_ECR_REPO" envDefault:"false" description:"By disabling this ECR repo won't get created if it's not available on ECR from build configuration"`
	MaxCiWorkflowRetries             int                             `env:"MAX_CI_WORKFLOW_RETRIES" envDefault:"0" description:"Maximum time CI-workflow create pod if it fails to complete"`
	NatsServerHost                   string                          `env:"NATS_SERVER_HOST" envDefault:"nats://devtron-nats.devtroncd:4222"`
	ImageScanMaxRetries              int                             `env:"IMAGE_SCAN_MAX_RETRIES" envDefault:"3" description:"Max retry count for image-scanning"`
	ImageScanRetryDelay              int                             `env:"IMAGE_SCAN_RETRY_DELAY" envDefault:"5" description:"Delay for the image-scaning to start"` 
	ShowDockerBuildCmdInLogs         bool                            `env:"SHOW_DOCKER_BUILD_ARGS" envDefault:"true" description:"To enable showing the args passed for CI in build logs"`
	IgnoreCmCsInCiJob                bool                            `env:"IGNORE_CM_CS_IN_CI_JOB" envDefault:"false" description:"Ignore CM/CS in CI-pipeline as Job"`
	//Deprecated: use WorkflowCacheConfig instead
	SkipCiJobBuildCachePushPull bool `env:"SKIP_CI_JOB_BUILD_CACHE_PUSH_PULL" envDefault:"false" description:"To skip cache Push/Pull for ci job"`
	// from CdConfig
	CdLimitCpu                       string                          `env:"CD_LIMIT_CI_CPU" envDefault:"0.5" description:"CPU Resource Limit Pre/Post CD"`
	CdLimitMem                       string                          `env:"CD_LIMIT_CI_MEM" envDefault:"3G" description:"Memory Resource Limit Pre/Post CD"`
	CdReqCpu                         string                          `env:"CD_REQ_CI_CPU" envDefault:"0.5" description:"CPU Resource Rquest Pre/Post CD"`
	CdReqMem                         string                          `env:"CD_REQ_CI_MEM" envDefault:"3G" description:"Memory Resource Rquest Pre/Post CD"`
	CdTaintKey                       string                          `env:"CD_NODE_TAINTS_KEY" envDefault:"dedicated" description:"Toleration key for Pre/Post CD"`
	ExternalCdTaintKey               string                          `env:"EXTERNAL_CD_NODE_TAINTS_KEY" envDefault:"dedicated"`
	UseExternalNode                  bool                            `env:"USE_EXTERNAL_NODE" envDefault:"false" description:"It is used in case of Pre/ Post Cd with run in application mode. If enabled the node lebels are read from EXTERNAL_CD_NODE_LABEL_SELECTOR else from CD_NODE_LABEL_SELECTOR MODE: if the vale is DEV, it will read the local kube config file or else from the cluser location."`
	CdWorkflowServiceAccount         string                          `env:"CD_WORKFLOW_SERVICE_ACCOUNT" envDefault:"cd-runner" description:"Service account to be used in Pre/Post CD pod"`
	CdDefaultBuildLogsKeyPrefix      string                          `env:"DEFAULT_BUILD_LOGS_KEY_PREFIX" `
	CdDefaultArtifactKeyPrefix       string                          `env:"DEFAULT_CD_ARTIFACT_KEY_LOCATION" description:"Bucket prefix for build cache"`
	CdTaintValue                     string                          `env:"CD_NODE_TAINTS_VALUE" envDefault:"ci" description:"Toleration value for Pre/Post CD"`
	ExternalCdTaintValue             string                          `env:"EXTERNAL_CD_NODE_TAINTS_VALUE" envDefault:"ci"`
	CdDefaultBuildLogsBucket         string                          `env:"DEFAULT_BUILD_LOGS_BUCKET" description:"Bucket name for build logs"`
	CdNodeLabelSelector              []string                        `env:"CD_NODE_LABEL_SELECTOR" description:"Node label selector for  Pre/Post CD"`
	ExternalCdNodeLabelSelector      []string                        `env:"EXTERNAL_CD_NODE_LABEL_SELECTOR" description:"This is an array of strings used when submitting a workflow for pre or post-CD execution. If the "Run in Target Env" option is selected (indicating execution in an external cluster) and the USE_EXTERNAL_NODE flag is set to true, the final node constraints will be determined by the values specified in this environment variable."`
	CdDefaultNamespace               string                          `env:"DEFAULT_CD_NAMESPACE" description:"Namespace for devtron stack"`
	CdDefaultImage                   string                          `env:"DEFAULT_CI_IMAGE"`
	CdDefaultTimeout                 int64                           `env:"DEFAULT_CD_TIMEOUT" envDefault:"3600" description:"Timeout for Pre/Post-Cd to be completed"`
	CdDefaultCdLogsBucketRegion      string                          `env:"DEFAULT_CD_LOGS_BUCKET_REGION" `
	WfControllerInstanceID           string                          `env:"WF_CONTROLLER_INSTANCE_ID" envDefault:"devtron-runner" description:"Workflow controller instance ID."`
	CdDefaultAddressPoolBaseCidr     string                          `env:"CD_DEFAULT_ADDRESS_POOL_BASE_CIDR" description:"To pass the IP cidr for Pre/Post cd "`
	CdDefaultAddressPoolSize         int                             `env:"CD_DEFAULT_ADDRESS_POOL_SIZE" description:"The subnet size to allocate from the base pool for CD"`
	ExposeCDMetrics                  bool                            `env:"EXPOSE_CD_METRICS" envDefault:"false" description:"To expose CD metrics"`
	UseBlobStorageConfigInCdWorkflow bool                            `env:"USE_BLOB_STORAGE_CONFIG_IN_CD_WORKFLOW" envDefault:"true" description:"To enable blob storage in pre and post cd"`
	CdWorkflowExecutorType           cdWorkflow.WorkflowExecutorType `env:"CD_WORKFLOW_EXECUTOR_TYPE" envDefault:"AWF" description:"Executor type for Pre/Post CD(AWF,System)"`
	TerminationGracePeriod           int                             `env:"TERMINATION_GRACE_PERIOD_SECS" envDefault:"180" description:"this is the time given to workflow pods to shutdown. (grace full termination time)"`
	MaxCdWorkflowRunnerRetries       int                             `env:"MAX_CD_WORKFLOW_RUNNER_RETRIES" envDefault:"0" description:"Maximum time pre/post-cd-workflow create pod if it fails to complete"`

	// common in both ciconfig and cd config
	Type                                       string
	Mode                                       string `env:"MODE" envDefault:"DEV"`
	OrchestratorHost                           string `env:"ORCH_HOST" envDefault:"http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats" description:"Orchestrator micro-service URL "`
	OrchestratorToken                          string `env:"ORCH_TOKEN" envDefault:"" description:"Orchestrator token"`
	ClusterConfig                              *rest.Config
	CloudProvider                              blob_storage.BlobStorageType `env:"BLOB_STORAGE_PROVIDER" envDefault:"S3" description:"Blob storage provider name(AWS/GCP/Azure)"`
	BlobStorageEnabled                         bool                         `env:"BLOB_STORAGE_ENABLED" envDefault:"false" description:"To enable blob storage"`
	BlobStorageS3AccessKey                     string                       `env:"BLOB_STORAGE_S3_ACCESS_KEY" description:"S3 access key for s3 blob storage"`
	BlobStorageS3SecretKey                     string                       `env:"BLOB_STORAGE_S3_SECRET_KEY" description:"Secret key for s3 blob storage"`
	BlobStorageS3Endpoint                      string                       `env:"BLOB_STORAGE_S3_ENDPOINT" description:"S3 endpoint URL for s3 blob storage"`
	BlobStorageS3EndpointInsecure              bool                         `env:"BLOB_STORAGE_S3_ENDPOINT_INSECURE" envDefault:"false" description:"To use insecure s3 endpoint"`
	BlobStorageS3BucketVersioned               bool                         `env:"BLOB_STORAGE_S3_BUCKET_VERSIONED" envDefault:"true" description:"To enable buctet versioning for blob storage"`
	BlobStorageGcpCredentialJson               string                       `env:"BLOB_STORAGE_GCP_CREDENTIALS_JSON" description:"GCP cred json for GCS blob storage"`
	AzureAccountName                           string                       `env:"AZURE_ACCOUNT_NAME" description:"Account name for azure blob storage"`
	AzureGatewayUrl                            string                       `env:"AZURE_GATEWAY_URL" envDefault:"http://devtron-minio.devtroncd:9000" description:"Sent to CI runner for blob"`
	AzureGatewayConnectionInsecure             bool                         `env:"AZURE_GATEWAY_CONNECTION_INSECURE" envDefault:"true" description:"Azure gateway connection allows insecure if true"`
	AzureBlobContainerCiLog                    string                       `env:"AZURE_BLOB_CONTAINER_CI_LOG" description:"Log bucket for azure blob storage"`
	AzureBlobContainerCiCache                  string                       `env:"AZURE_BLOB_CONTAINER_CI_CACHE" description:"Cache bucket name for azure blob storage"`
	AzureAccountKey                            string                       `env:"AZURE_ACCOUNT_KEY" description:"If blob storage is bieng used of azure then pass the secret key to access the bucket"`
	BuildLogTTLValue                           int                          `env:"BUILD_LOG_TTL_VALUE_IN_SECS" envDefault:"3600" description:"This is the time that the pods of ci/pre-cd/post-cd live after completion state."`
	BaseLogLocationPath                        string                       `env:"BASE_LOG_LOCATION_PATH" envDefault:"/home/devtron/" description:"Used to store, download logs of ci workflow, artifact"`
	InAppLoggingEnabled                        bool                         `env:"IN_APP_LOGGING_ENABLED" envDefault:"false" description:"Used in case of argo workflow is enabled. If enabled logs push will be managed by us, else will be managed by argo workflow."`
	BuildxProvenanceMode                       string                       `env:"BUILDX_PROVENANCE_MODE" envDefault:"" description:"provinance is set to true by default by docker. this will add some build related data in generated build manifest.it also adds some unknown:unknown key:value pair which may not be compatible by some container registries. with buildx k8s driver , provinenance=true is causing issue when push manifest to quay registry, so setting it to false"` // provenance is set to false if this flag is not set
	ExtBlobStorageCmName                       string                       `env:"EXTERNAL_BLOB_STORAGE_CM_NAME" envDefault:"blob-storage-cm" description:"name of the config map(contains bucket name, etc.) in external cluster when there is some operation related to external cluster, for example:-downloading cd artifact pushed in external cluster's env and we need to download from there, downloads ci logs pushed in external cluster's blob"`
	ExtBlobStorageSecretName                   string                       `env:"EXTERNAL_BLOB_STORAGE_SECRET_NAME" envDefault:"blob-storage-secret" description:"name of the secret(contains password, accessId,passKeys, etc.) in external cluster when there is some operation related to external cluster, for example:-downloading cd artifact pushed in external cluster's env and we need to download from there, downloads ci logs pushed in external cluster's blob"`
	UseArtifactListingQueryV2                  bool                         `env:"USE_ARTIFACT_LISTING_QUERY_V2" envDefault:"true" description:"To use the V2 query for listing artifacts"`
	UseImageTagFromGitProviderForTagBasedBuild bool                         `env:"USE_IMAGE_TAG_FROM_GIT_PROVIDER_FOR_TAG_BASED_BUILD" envDefault:"false" description:"To use the same tag in container image as that of git tag"` // this is being done for https://github.com/devtron-labs/devtron/issues/4263
	UseDockerApiToGetDigest                    bool                         `env:"USE_DOCKER_API_TO_GET_DIGEST" envDefault:"false" description:"when user do not pass the digest  then this flag controls , finding the image digest using docker API or not. if set to true we get the digest from docker API call else use docker pull command. [logic in ci-runner]"`
	EnableWorkflowExecutionStage               bool                         `env:"ENABLE_WORKFLOW_EXECUTION_STAGE" envDefault:"true" description:"if enabled then we will display build stages separately for CI/Job/Pre-Post CD" example:"true"`
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

const (
	CiArtifactLocationFormat = "%d/%d.zip"
	CdArtifactLocationFormat = "%d/%d.zip"
)

func GetCiConfig() (*CiConfig, error) {
	ciCdConfig := &CiCdConfig{}
	err := env.Parse(ciCdConfig)
	if err != nil {
		return nil, err
	}
	ciConfig := CiConfig{
		CiCdConfig: ciCdConfig,
	}
	ciConfig.Type = CiConfigType
	return &ciConfig, nil
}

func GetCiConfigWithWorkflowCacheConfig() (*CiConfig, WorkflowCacheConfig, error) {
	ciConfig, err := GetCiConfig()
	if err != nil {
		return nil, WorkflowCacheConfig{}, err
	}
	workflowCacheConfig, err := getWorkflowCacheConfig(ciConfig.WorkflowCacheConfig)
	if err != nil {
		return nil, WorkflowCacheConfig{}, err
	}
	return ciConfig, workflowCacheConfig, nil
}

func getWorkflowCacheConfig(workflowCacheConfigEnv string) (WorkflowCacheConfig, error) {
	workflowCacheConfig := WorkflowCacheConfig{}
	err := json.Unmarshal([]byte(workflowCacheConfigEnv), &workflowCacheConfig)
	if err != nil {
		return WorkflowCacheConfig{}, err
	}
	return workflowCacheConfig, nil
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
	// validation for supported cloudproviders
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
	case CdConfigType:
		return impl.CdLimitCpu
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetLimitMem() string {
	switch impl.Type {
	case CdConfigType:
		return impl.CdLimitMem
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetReqCpu() string {
	switch impl.Type {
	case CdConfigType:
		return impl.CdReqCpu
	default:
		return ""
	}
}

func (impl *CiCdConfig) GetReqMem() string {
	switch impl.Type {
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
func (impl *CiCdConfig) getDefaultArtifactKeyPrefix() string {
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
		ciArtifactLocationFormat := CiArtifactLocationFormat
		if len(impl.getDefaultArtifactKeyPrefix()) != 0 {
			ciArtifactLocationFormat = path.Join(impl.getDefaultArtifactKeyPrefix(), ciArtifactLocationFormat)
		}
		return ciArtifactLocationFormat
	case CdConfigType:
		cdArtifactLocationFormat := CdArtifactLocationFormat
		if len(impl.getDefaultArtifactKeyPrefix()) != 0 {
			cdArtifactLocationFormat = path.Join(impl.getDefaultArtifactKeyPrefix(), cdArtifactLocationFormat)
		}
		return cdArtifactLocationFormat
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

func (impl *CiCdConfig) GetWorkflowExecutorType() cdWorkflow.WorkflowExecutorType {
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
	TargetPlatforms  []*bean2.TargetPlatform                     `json:"targetPlatforms"`
}

type Trigger struct {
	PipelineId            int
	CommitHashes          map[int]pipelineConfig.GitCommit
	CiMaterials           []*pipelineConfig.CiPipelineMaterial
	TriggeredBy           int32
	InvalidateCache       bool
	RuntimeParameters     *common.RuntimeParameters // extra env variables which will be used for CI
	EnvironmentId         int
	PipelineType          string
	CiArtifactLastFetch   time.Time
	ReferenceCiWorkflowId int
}

func (obj *Trigger) BuildTriggerObject(refCiWorkflow *pipelineConfig.CiWorkflow,
	ciMaterials []*pipelineConfig.CiPipelineMaterial, triggeredBy int32,
	invalidateCache bool, runtimeParameters *common.RuntimeParameters,
	pipelineType string) {

	obj.PipelineId = refCiWorkflow.CiPipelineId
	obj.CommitHashes = refCiWorkflow.GitTriggers
	obj.CiMaterials = ciMaterials
	obj.TriggeredBy = triggeredBy
	obj.InvalidateCache = invalidateCache
	obj.EnvironmentId = refCiWorkflow.EnvironmentId
	obj.ReferenceCiWorkflowId = refCiWorkflow.Id
	obj.InvalidateCache = invalidateCache
	obj.RuntimeParameters = runtimeParameters
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

func (r *BuildLogRequest) SetBuildLogRequest(cmConfig *bean.CmBlobStorageConfig, secretConfig *bean.SecretBlobStorageConfig) {
	r.CloudProvider = cmConfig.CloudProvider
	r.AzureBlobConfig.AccountName = cmConfig.AzureAccountName
	r.AzureBlobConfig.AccountKey = DecodeSecretKey(secretConfig.AzureAccountKey)
	r.AzureBlobConfig.BlobContainerName = cmConfig.AzureBlobContainerCiLog

	r.GcpBlobBaseConfig.CredentialFileJsonData = DecodeSecretKey(secretConfig.GcpBlobStorageCredentialJson)
	r.GcpBlobBaseConfig.BucketName = cmConfig.CdDefaultBuildLogsBucket

	r.AwsS3BaseConfig.AccessKey = cmConfig.S3AccessKey
	r.AwsS3BaseConfig.EndpointUrl = cmConfig.S3Endpoint
	r.AwsS3BaseConfig.Passkey = DecodeSecretKey(secretConfig.S3SecretKey)
	isEndpointInSecure, _ := strconv.ParseBool(cmConfig.S3EndpointInsecure)
	r.AwsS3BaseConfig.IsInSecure = isEndpointInSecure
	r.AwsS3BaseConfig.BucketName = cmConfig.CdDefaultBuildLogsBucket
	r.AwsS3BaseConfig.Region = cmConfig.CdDefaultCdLogsBucketRegion
	s3BucketVersioned, _ := strconv.ParseBool(cmConfig.S3BucketVersioned)
	r.AwsS3BaseConfig.VersioningEnabled = s3BucketVersioned
}

func DecodeSecretKey(secretKey string) string {
	decodedKey, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		fmt.Println("error decoding base64 key:", err)
	}
	return string(decodedKey)
}

type WorkflowCacheConfig struct {
	IgnoreCI    bool `json:"ignoreCI"`
	IgnoreCIJob bool `json:"ignoreCIJob"`
	IgnoreJob   bool `json:"ignoreJob"`
}
