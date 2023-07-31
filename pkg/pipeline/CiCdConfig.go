package pipeline

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os/user"
	"path/filepath"
	"strings"
)

type CiCdConfig struct {
	//from ciCdConfig
	DefaultCacheBucket               string   `env:"DEFAULT_CACHE_BUCKET" envDefault:"ci-caching"`
	DefaultCacheBucketRegion         string   `env:"DEFAULT_CACHE_BUCKET_REGION" envDefault:"us-east-2"`
	CiLogsKeyPrefix                  string   `env:"CI_LOGS_KEY_PREFIX" envDxefault:"my-artifacts"`
	CiDefaultImage                   string   `env:"DEFAULT_CI_IMAGE" envDefault:"686244538589.dkr.ecr.us-east-2.amazonaws.com/cirunner:47"`
	CiDefaultNamespace               string   `env:"DEFAULT_NAMESPACE" envDefault:"devtron-ci"`
	CiDefaultTimeout                 int64    `env:"DEFAULT_TIMEOUT" envDefault:"3600"`
	CiDefaultBuildLogsBucket         string   `env:"DEFAULT_BUILD_LOGS_BUCKET" envDefault:"devtron-pro-ci-logs"`
	CiDefaultCdLogsBucketRegion      string   `env:"DEFAULT_CD_LOGS_BUCKET_REGION" envDefault:"us-east-2"`
	CiLimitCpu                       string   `env:"LIMIT_CI_CPU" envDefault:"0.5"`
	CiLimitMem                       string   `env:"LIMIT_CI_MEM" envDefault:"3G"`
	CiReqCpu                         string   `env:"REQ_CI_CPU" envDefault:"0.5"`
	CiReqMem                         string   `env:"REQ_CI_MEM" envDefault:"3G"`
	CiTaintKey                       string   `env:"CI_NODE_TAINTS_KEY" envDefault:""`
	CiTaintValue                     string   `env:"CI_NODE_TAINTS_VALUE" envDefault:""`
	CiNodeLabelSelector              []string `env:"CI_NODE_LABEL_SELECTOR"`
	CacheLimit                       int64    `env:"CACHE_LIMIT" envDefault:"5000000000"` // TODO: Add to default db config also
	CiDefaultBuildLogsKeyPrefix      string   `env:"DEFAULT_BUILD_LOGS_KEY_PREFIX" envDefault:"arsenal-v1"`
	CiDefaultArtifactKeyPrefix       string   `env:"DEFAULT_ARTIFACT_KEY_LOCATION" envDefault:"arsenal-v1/ci-artifacts"`
	CiWorkflowServiceAccount         string   `env:"WORKFLOW_SERVICE_ACCOUNT" envDefault:"ci-runner"`
	ExternalCiApiSecret              string   `env:"EXTERNAL_CI_API_SECRET" envDefault:"devtroncd-secret"`
	ExternalCiWebhookUrl             string   `env:"EXTERNAL_CI_WEB_HOOK_URL" envDefault:""`
	ExternalCiPayload                string   `env:"EXTERNAL_CI_PAYLOAD" envDefault:"{\"ciProjectDetails\":[{\"gitRepository\":\"https://github.com/vikram1601/getting-started-nodejs.git\",\"checkoutPath\":\"./abc\",\"commitHash\":\"239077135f8cdeeccb7857e2851348f558cb53d3\",\"commitTime\":\"2022-10-30T20:00:00\",\"branch\":\"master\",\"message\":\"Update README.md\",\"author\":\"User Name \"}],\"dockerImage\":\"445808685819.dkr.ecr.us-east-2.amazonaws.com/orch:23907713-2\"}"`
	CiArtifactLocationFormat         string   `env:"CI_ARTIFACT_LOCATION_FORMAT" envDefault:"%d/%d.zip"`
	ImageScannerEndpoint             string   `env:"IMAGE_SCANNER_ENDPOINT" envDefault:"http://image-scanner-new-demo-devtroncd-service.devtroncd:80"`
	CiDefaultAddressPoolBaseCidr     string   `env:"CI_DEFAULT_ADDRESS_POOL_BASE_CIDR"`
	CiDefaultAddressPoolSize         int      `env:"CI_DEFAULT_ADDRESS_POOL_SIZE"`
	CiRunnerDockerMTUValue           int      `env:"CI_RUNNER_DOCKER_MTU_VALUE" envDefault:"-1"`
	IgnoreDockerCacheForCI           bool     `env:"CI_IGNORE_DOCKER_CACHE"`
	VolumeMountsForCiJson            string   `env:"CI_VOLUME_MOUNTS_JSON"`
	BuildPvcCachePath                string   `env:"PRE_CI_CACHE_PATH" envDefault:"/devtroncd-cache"`
	DefaultPvcCachePath              string   `env:"DOCKER_BUILD_CACHE_PATH" envDefault:"/var/lib/docker"`
	BuildxPvcCachePath               string   `env:"BUILDX_CACHE_PATH" envDefault:"/var/lib/devtron/buildx"`
	UseBlobStorageConfigInCiWorkflow bool     `env:"USE_BLOB_STORAGE_CONFIG_IN_CI_WORKFLOW" envDefault:"true"`
	DefaultTargetPlatform            string   `env:"DEFAULT_TARGET_PLATFORM" envDefault:""`
	UseBuildx                        bool     `env:"USE_BUILDX" envDefault:"false"`
	EnableBuildContext               bool     `env:"ENABLE_BUILD_CONTEXT" envDefault:"false"`
	ImageRetryCount                  int      `env:"IMAGE_RETRY_COUNT" envDefault:"0"`
	ImageRetryInterval               int      `env:"IMAGE_RETRY_INTERVAL" envDefault:"5"` //image retry interval takes value in seconds

	//from ciCdConfig
	CdLimitCpu                       string                              `env:"CD_LIMIT_CI_CPU" envDefault:"0.5"`
	CdLimitMem                       string                              `env:"CD_LIMIT_CI_MEM" envDefault:"3G"`
	CdReqCpu                         string                              `env:"CD_REQ_CI_CPU" envDefault:"0.5"`
	CdReqMem                         string                              `env:"CD_REQ_CI_MEM" envDefault:"3G"`
	CdTaintKey                       string                              `env:"CD_NODE_TAINTS_KEY" envDefault:"dedicated"`
	CdWorkflowServiceAccount         string                              `env:"CD_WORKFLOW_SERVICE_ACCOUNT" envDefault:"cd-runner"`
	CdDefaultBuildLogsKeyPrefix      string                              `env:"DEFAULT_BUILD_LOGS_KEY_PREFIX" `
	CdDefaultArtifactKeyPrefix       string                              `env:"DEFAULT_CD_ARTIFACT_KEY_LOCATION" `
	CdTaintValue                     string                              `env:"CD_NODE_TAINTS_VALUE" envDefault:"ci"`
	CdDefaultBuildLogsBucket         string                              `env:"DEFAULT_BUILD_LOGS_BUCKET" `
	CdNodeLabelSelector              []string                            `env:"CD_NODE_LABEL_SELECTOR"`
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

	//common in both ciconfig and cd config
	Mode                           string `env:"MODE" envDefault:"DEV"`
	OrchestratorHost               string `env:"ORCH_HOST" envDefault:"http://devtroncd-orchestrator-service-prod.devtroncd/webhook/msg/nats"`
	OrchestratorToken              string `env:"ORCH_TOKEN" envDefault:""`
	ClusterConfig                  *rest.Config
	NodeLabel                      map[string]string
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
}

type CiVolumeMount struct {
	Name               string `json:"name"`
	HostMountPath      string `json:"hostMountPath"`
	ContainerMountPath string `json:"containerMountPath"`
}

const ExternalCiWebhookPath = "orchestrator/webhook/ext-ci"

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
	if err != nil {
		return nil, err
	}

	//validation for supported cloudproviders
	if cfg.BlobStorageEnabled && cfg.CloudProvider != BLOB_STORAGE_S3 && cfg.CloudProvider != BLOB_STORAGE_AZURE &&
		cfg.CloudProvider != BLOB_STORAGE_GCP && cfg.CloudProvider != BLOB_STORAGE_MINIO {
		return nil, fmt.Errorf("unsupported blob storage provider: %s", cfg.CloudProvider)
	}

	return cfg, err
}
func getNodeLabel(cfg *CiCdConfig, isCi bool) (map[string]string, error) {
	node := cfg.CdNodeLabelSelector
	if isCi {
		node = cfg.CiNodeLabelSelector
	}
	cfg.NodeLabel = make(map[string]string)
	for _, l := range node {
		if l == "" {
			continue
		}
		kv := strings.Split(l, "=")
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid ci node label selector %s, it must be in form key=value, key2=val2", kv)
		}
		cfg.NodeLabel[kv[0]] = kv[1]
	}
	return cfg.NodeLabel, nil
}
