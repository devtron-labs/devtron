package blob_storage

type BlobStorageRequest struct {
	StorageType         BlobStorageType
	SourceKey           string
	DestinationKey      string
	AwsS3BaseConfig     *AwsS3BaseConfig
	AzureBlobBaseConfig *AzureBlobBaseConfig
	GcpBlobBaseConfig   *GcpBlobBaseConfig
}

type BlobStorageS3Config struct {
	AccessKey            string `json:"accessKey"`
	Passkey              string `json:"passkey"`
	EndpointUrl          string `json:"endpointUrl"`
	IsInSecure           bool   `json:"isInSecure"`
	CiLogBucketName      string `json:"ciLogBucketName"`
	CiLogRegion          string `json:"ciLogRegion"`
	CiCacheBucketName    string `json:"ciCacheBucketName"`
	CiCacheRegion        string `json:"ciCacheRegion"`
	CiArtifactBucketName string `json:"ciArtifactBucketName"`
	CiArtifactRegion     string `json:"ciArtifactRegion"`
}

type AwsS3BaseConfig struct {
	AccessKey   string `json:"accessKey"`
	Passkey     string `json:"passkey"`
	EndpointUrl string `json:"endpointUrl"`
	IsInSecure  bool   `json:"isInSecure"`
	BucketName  string `json:"bucketName"`
	Region      string `json:"region"`
}

type AzureBlobConfig struct {
	Enabled               bool   `json:"enabled"`
	AccountName           string `json:"accountName"`
	BlobContainerCiLog    string `json:"blobContainerCiLog"`
	BlobContainerCiCache  string `json:"blobContainerCiCache"`
	BlobContainerArtifact string `json:"blobStorageArtifact"`
	AccountKey            string `json:"accountKey"`
}

type AzureBlobBaseConfig struct {
	Enabled           bool   `json:"enabled"`
	AccountName       string `json:"accountName"`
	AccountKey        string `json:"accountKey"`
	BlobContainerName string `json:"blobContainerName"`
}

type GcpBlobConfig struct {
	CredentialFileJsonData string `json:"credentialFileData"`
	CacheBucketName        string `json:"ciCacheBucketName"`
	LogBucketName          string `json:"logBucketName"`
	ArtifactBucketName     string `json:"artifactBucketName"`
}

type GcpBlobBaseConfig struct {
	BucketName             string `json:"bucketName"`
	CredentialFileJsonData string `json:"credentialFileData"`
}

type BlobStorageType string

const (
	BLOB_STORAGE_AZURE BlobStorageType = "AZURE"
	BLOB_STORAGE_S3                    = "S3"
	BLOB_STORAGE_GCP                   = "GCP"
	BLOB_STORAGE_MINIO                 = "MINIO"
)
