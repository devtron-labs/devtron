package pipeline

import (
	"context"
	"encoding/base64"
	"fmt"
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"github.com/devtron-labs/common-lib/utils/k8s"
	bean2 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"go.uber.org/zap"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

type BlobStorageConfigService interface {
	FetchCmAndSecretBlobConfigFromExternalCluster(clusterConfig *k8s.ClusterConfig, namespace string) (*bean2.CmBlobStorageConfig, *bean2.SecretBlobStorageConfig, error)
}
type BlobStorageConfigServiceImpl struct {
	Logger     *zap.SugaredLogger
	k8sUtil    *k8s.K8sUtil
	ciCdConfig *CiCdConfig
}

func NewBlobStorageConfigServiceImpl(Logger *zap.SugaredLogger, k8sUtil *k8s.K8sUtil, ciCdConfig *CiCdConfig) *BlobStorageConfigServiceImpl {
	return &BlobStorageConfigServiceImpl{
		Logger:     Logger,
		k8sUtil:    k8sUtil,
		ciCdConfig: ciCdConfig,
	}
}

func (impl *BlobStorageConfigServiceImpl) FetchCmAndSecretBlobConfigFromExternalCluster(clusterConfig *k8s.ClusterConfig, namespace string) (*bean2.CmBlobStorageConfig, *bean2.SecretBlobStorageConfig, error) {
	cmConfig := &bean2.CmBlobStorageConfig{}
	secretConfig := &bean2.SecretBlobStorageConfig{}
	ctx := context.Background()
	opts := v12.GetOptions{}
	_, _, kubeClient, err := impl.k8sUtil.GetK8sConfigAndClients(&k8s.ClusterConfig{})
	if err != nil {
		impl.Logger.Errorw("FetchCmAndSecretBlobConfigFromExternalCluster, error in getting kubeClient by cluster config", "err", err)
		return cmConfig, secretConfig, err
	}
	cm, err := kubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, impl.ciCdConfig.ExtBlobStorageCmName, opts)
	if err != nil {
		impl.Logger.Errorw("error in getting config map in external cluster", "err", err, "blobStorageCmName", impl.ciCdConfig.ExtBlobStorageCmName, "clusterName", clusterConfig.ClusterName)
		return cmConfig, secretConfig, err
	}
	secret, err := kubeClient.CoreV1().Secrets(namespace).Get(ctx, impl.ciCdConfig.ExtBlobStorageSecretName, opts)
	if err != nil {
		impl.Logger.Errorw("error in getting secret in external cluster", "err", err, "blobStorageSecretName", impl.ciCdConfig.ExtBlobStorageSecretName, "clusterName", clusterConfig.ClusterName)
		return cmConfig, secretConfig, err
	}
	//for IAM configured in S3 in external cluster, get logs/artifact will not work
	if cm.Data != nil && secret.Data != nil {
		err = cmConfig.SetCmBlobStorageConfig(cm.Data)
		if err != nil {
			fmt.Println("error marshalling external blob storage cm data to struct:", err)
			return cmConfig, secretConfig, err
		}
		err = secretConfig.SetSecretBlobStorageConfig(secret.Data)
		if err != nil {
			fmt.Println("error marshalling external blob storage secret data to struct:", err)
			return cmConfig, secretConfig, err
		}
	}
	if cm.Data == nil {
		fmt.Println("Data field not found in config map")
	}
	if secret.Data == nil {
		fmt.Println("Data field not found in secret")
	}
	impl.Logger.Infow("fetching cm and secret from external cluster cloud provider", "ext cluster config: ", cmConfig)
	return cmConfig, secretConfig, nil
}

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
