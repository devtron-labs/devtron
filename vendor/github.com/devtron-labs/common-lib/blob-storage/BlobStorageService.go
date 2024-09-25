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

import (
	"context"
	"fmt"
	"github.com/devtron-labs/common-lib/utils"
	"go.uber.org/zap"
	"log"
	"os"
)

type BlobStorageService interface {
	PutWithCommand(request *BlobStorageRequest) error
	Get(request *BlobStorageRequest) (bool, error)
	DeleteObjectForS3(request *BlobStorageRequest) error
}

type BlobStorageServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewBlobStorageServiceImpl(logger *zap.SugaredLogger) *BlobStorageServiceImpl {
	if logger == nil {
		logger, _ = utils.NewSugardLogger()
	}
	impl := &BlobStorageServiceImpl{
		logger: logger,
	}
	return impl
}

func (impl *BlobStorageServiceImpl) PutWithCommand(request *BlobStorageRequest) error {
	var err error
	switch request.StorageType {
	case BLOB_STORAGE_S3:
		awsS3Blob := AwsS3Blob{}
		err = awsS3Blob.UploadBlob(request, err)
	case BLOB_STORAGE_AZURE:
		azureBlob := AzureBlob{}
		err = azureBlob.UploadBlob(context.Background(), request.DestinationKey, request.AzureBlobBaseConfig, request.SourceKey, request.AzureBlobBaseConfig.BlobContainerName)
	case BLOB_STORAGE_GCP:
		gcpBlob := GCPBlob{}
		err = gcpBlob.UploadBlob(request)
	default:
		return fmt.Errorf("blob-storage %s not supported", request.StorageType)
	}
	if err != nil {
		log.Println(" -----> push err", err)
	}
	return err
}

func (impl *BlobStorageServiceImpl) Get(request *BlobStorageRequest) (bool, int64, error) {

	downloadSuccess := false
	numBytes := int64(0)
	file, err := os.Create("/" + request.DestinationKey)
	defer file.Close()
	if err != nil {
		log.Println(err)
		return false, 0, err
	}
	switch request.StorageType {
	case BLOB_STORAGE_S3:
		awsS3Blob := AwsS3Blob{}
		downloadSuccess, numBytes, err = awsS3Blob.DownloadBlob(request, downloadSuccess, numBytes, err, file)
	case BLOB_STORAGE_AZURE:
		b := AzureBlob{}
		downloadSuccess, err = b.DownloadBlob(context.Background(), request.SourceKey, request.AzureBlobBaseConfig, file)
		fileInfo, _ := file.Stat()
		numBytes = fileInfo.Size()
	case BLOB_STORAGE_GCP:
		gcpBlob := &GCPBlob{}
		downloadSuccess, numBytes, err = gcpBlob.DownloadBlob(request, file)
	default:
		return downloadSuccess, numBytes, fmt.Errorf("blob-storage %s not supported", request.StorageType)
	}

	return downloadSuccess, numBytes, err
}

// TODO: Have not Tested it
func (impl *BlobStorageServiceImpl) DeleteObjectForS3(request *BlobStorageRequest) error {
	if request.StorageType == BLOB_STORAGE_S3 {
		awsS3Blob := AwsS3Blob{}
		err := awsS3Blob.DeleteObjectFromBlob(request)
		if err != nil {
			impl.logger.Errorw("error in deleting object from S3", "err", err, "DestinationKey", request.DestinationKey, "StorageType", request.StorageType)
			return err
		}
	}
	return nil
}

func (impl *BlobStorageServiceImpl) UploadToBlobWithSession(request *BlobStorageRequest) error {
	var err error
	switch request.StorageType {
	case BLOB_STORAGE_S3:
		awsS3Blob := AwsS3Blob{}
		_, err = awsS3Blob.UploadWithSession(request)
	case BLOB_STORAGE_AZURE:
		azureBlob := AzureBlob{}
		err = azureBlob.UploadBlob(context.Background(), request.DestinationKey, request.AzureBlobBaseConfig, request.SourceKey, request.AzureBlobBaseConfig.BlobContainerName)
	case BLOB_STORAGE_GCP:
		gcpBlob := GCPBlob{}
		err = gcpBlob.UploadBlob(request)
	default:
		return fmt.Errorf("blob-storage %s not supported", request.StorageType)
	}
	if err != nil {
		log.Println(" -----> push err", err)
	}
	return err
}
