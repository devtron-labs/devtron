package blob_storage

type BlobStorageRequest struct {
	StorageType          BlobStorageType
	Region               string
	Endpoint             string // for s3 compatible storage
	BucketName           string
	SourceKey            string
	DestinationKey       string
	AccessKey            string
	Passkey              string
	FileDownloadLocation string
	AwsS3BaseConfig      *AwsS3BaseConfig
	AzureBlobConfig      *AzureBlobConfig
}

type BlobStorageS3Config struct {
	AccessKey            string `json:"accessKey"`
	Passkey              string `json:"passkey"`
	EndpointUrl          string `json:"endpointUrl"`
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
	BucketName  string `json:"bucketName"`
	Region      string `json:"region"`
	//CiLogBucketName      string `json:"ciLogBucketName"`
	//CiLogRegion          string `json:"ciLogRegion"`
	//CiCacheBucketName    string `json:"ciCacheBucketName"`
	//CiCacheRegion        string `json:"ciCacheRegion"`
	//CiArtifactBucketName string `json:"ciArtifactBucketName"`
	//CiArtifactRegion     string `json:"ciArtifactRegion"`
}

type BlobStorageType string

const (
	BLOB_STORAGE_AZURE BlobStorageType = "AZURE"
	BLOB_STORAGE_S3                    = "S3"
	BLOB_STORAGE_GCP                   = "GCP"
	BLOB_STORAGE_MINIO                 = "MINIO"
)
