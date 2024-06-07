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
	AccessKey                  string `json:"accessKey"`
	Passkey                    string `json:"passkey"`
	EndpointUrl                string `json:"endpointUrl"`
	IsInSecure                 bool   `json:"isInSecure"`
	CiLogBucketName            string `json:"ciLogBucketName"`
	CiLogRegion                string `json:"ciLogRegion"`
	CiLogBucketVersioning      bool   `json:"ciLogBucketVersioning"`
	CiCacheBucketName          string `json:"ciCacheBucketName"`
	CiCacheRegion              string `json:"ciCacheRegion"`
	CiCacheBucketVersioning    bool   `json:"ciCacheBucketVersioning"`
	CiArtifactBucketName       string `json:"ciArtifactBucketName"`
	CiArtifactRegion           string `json:"ciArtifactRegion"`
	CiArtifactBucketVersioning bool   `json:"ciArtifactBucketVersioning"`
}

type AwsS3BaseConfig struct {
	AccessKey         string `json:"accessKey"`
	Passkey           string `json:"passkey"`
	EndpointUrl       string `json:"endpointUrl"`
	IsInSecure        bool   `json:"isInSecure"`
	BucketName        string `json:"bucketName"`
	Region            string `json:"region"`
	VersioningEnabled bool   `json:"versioningEnabled"`
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
