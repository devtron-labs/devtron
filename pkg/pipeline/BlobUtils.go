package pipeline

import (
	"encoding/base64"
	"fmt"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"strconv"
)

func updateRequestWithExtClusterCmAndSecret(request *blob_storage.BlobStorageRequest, cmConfig *bean2.CmBlobStorageConfig, secretConfig *bean2.SecretBlobStorageConfig) *blob_storage.BlobStorageRequest {
	request.StorageType = cmConfig.CloudProvider

	request.AwsS3BaseConfig.AccessKey = cmConfig.S3AccessKey
	request.AwsS3BaseConfig.EndpointUrl = cmConfig.S3Endpoint
	request.AwsS3BaseConfig.Passkey = decodeSecretKey(secretConfig.S3SecretKey)
	isInSecure, _ := strconv.ParseBool(cmConfig.S3EndpointInsecure)
	request.AwsS3BaseConfig.IsInSecure = isInSecure
	request.AwsS3BaseConfig.BucketName = cmConfig.CdDefaultBuildLogsBucket
	request.AwsS3BaseConfig.Region = cmConfig.CdDefaultCdLogsBucketRegion
	s3BucketVersioned, _ := strconv.ParseBool(cmConfig.S3BucketVersioned)
	request.AwsS3BaseConfig.VersioningEnabled = s3BucketVersioned

	request.AzureBlobBaseConfig.AccountName = cmConfig.AzureAccountName
	request.AzureBlobBaseConfig.AccountKey = decodeSecretKey(secretConfig.AzureAccountKey)
	request.AzureBlobBaseConfig.BlobContainerName = cmConfig.AzureBlobContainerCiLog

	request.GcpBlobBaseConfig.CredentialFileJsonData = decodeSecretKey(secretConfig.GcpBlobStorageCredentialJson)
	request.GcpBlobBaseConfig.BucketName = cmConfig.CdDefaultBuildLogsBucket

	return request
}

func decodeSecretKey(secretKey string) string {
	decodedKey, err := base64.StdEncoding.DecodeString(secretKey)
	if err != nil {
		fmt.Println("error decoding base64 key:", err)
	}
	return string(decodedKey)
}
