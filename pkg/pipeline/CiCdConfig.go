package pipeline

import (
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"k8s.io/client-go/rest"
)

type CiCdConfig struct {
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
