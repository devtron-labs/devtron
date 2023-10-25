package bean

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/blob-storage"
)

type CmBlobStorageConfig struct {
	//AWS credentials
	CloudProvider               blob_storage.BlobStorageType `json:"BLOB_STORAGE_PROVIDER"`
	S3AccessKey                 string                       `json:"BLOB_STORAGE_S3_ACCESS_KEY"`
	S3Endpoint                  string                       `json:"BLOB_STORAGE_S3_ENDPOINT"`
	S3EndpointInsecure          string                       `json:"BLOB_STORAGE_S3_ENDPOINT_INSECURE"`
	S3BucketVersioned           string                       `json:"BLOB_STORAGE_S3_BUCKET_VERSIONED"`
	CdDefaultBuildLogsBucket    string                       `json:"DEFAULT_BUILD_LOGS_BUCKET" `
	CdDefaultCdLogsBucketRegion string                       `json:"DEFAULT_CD_LOGS_BUCKET_REGION" `
	DefaultCacheBucket          string                       `json:"DEFAULT_CACHE_BUCKET"`
	DefaultCacheBucketRegion    string                       `json:"DEFAULT_CACHE_BUCKET_REGION"`

	//Azure credentials
	AzureAccountName               string `json:"AZURE_ACCOUNT_NAME"`
	AzureGatewayUrl                string `json:"AZURE_GATEWAY_URL"`
	AzureGatewayConnectionInsecure string `json:"AZURE_GATEWAY_CONNECTION_INSECURE"`
	AzureBlobContainerCiLog        string `json:"AZURE_BLOB_CONTAINER_CI_LOG"`
	AzureBlobContainerCiCache      string `json:"AZURE_BLOB_CONTAINER_CI_CACHE"`
}

func (c *CmBlobStorageConfig) PopulateWithK8sExtBlobCmData(cm map[string]string) error {
	cmDataJson, err := json.Marshal(cm)
	if err != nil {
		fmt.Println("error marshalling external blob storage cm data to json:", err)
		return err
	}
	err = json.Unmarshal(cmDataJson, &c)
	if err != nil {
		fmt.Println("error unmarshalling external blob storage cm json to struct:", err)
		return err
	}
	return nil
}

type SecretBlobStorageConfig struct {
	//aws
	S3SecretKey string `json:"BLOB_STORAGE_S3_SECRET_KEY"`
	//gcp
	GcpBlobStorageCredentialJson string `json:"BLOB_STORAGE_GCP_CREDENTIALS_JSON"`
	//azure
	AzureAccountKey string `json:"AZURE_ACCOUNT_KEY"`
}

func (s *SecretBlobStorageConfig) PopulateWithK8sExtBlobSecretData(secret map[string][]byte) error {
	cmDataJson, err := json.Marshal(secret)
	if err != nil {
		fmt.Println("error marshalling external blob storage secret data to json:", err)
		return err
	}
	err = json.Unmarshal(cmDataJson, &s)
	if err != nil {
		fmt.Println("error unmarshalling external blob storage secret json to struct:", err)
		return err
	}
	return nil
}
