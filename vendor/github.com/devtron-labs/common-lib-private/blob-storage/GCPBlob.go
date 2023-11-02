package blob_storage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"log"
	"os"
)

type GCPBlob struct {
}

func (impl *GCPBlob) UploadBlob(request *BlobStorageRequest) error {
	ctx := context.Background()
	config := request.GcpBlobBaseConfig
	storageClient, err := impl.createGcpClient(ctx, config)
	if err != nil {
		return err
	}
	file, err := os.Open(request.SourceKey)
	if err != nil {
		return err
	}
	defer file.Close()
	gcpObject := impl.getGcpObject(storageClient, config, request.DestinationKey)
	if err != nil {
		return err
	}
	objectWriter := gcpObject.NewWriter(ctx)
	_, err = io.Copy(objectWriter, file)
	if err := objectWriter.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	return err
}

func (impl *GCPBlob) DownloadBlob(request *BlobStorageRequest, file *os.File) (bool, int64, error) {
	ctx := context.Background()
	config := request.GcpBlobBaseConfig
	storageClient, err := impl.createGcpClient(ctx, config)
	if err != nil {
		return false, 0, err
	}
	latestGeneration, err := impl.getLatestVersion(storageClient, request, ctx, config)
	if err != nil {
		return false, 0, err
	}
	gcpObject := impl.getGcpObject(storageClient, config, request.SourceKey)
	if err != nil {
		return false, 0, err
	}
	objectReader, err := gcpObject.If(storage.Conditions{GenerationMatch: latestGeneration}).NewReader(ctx)
	if err != nil {
		return false, 0, err
	}
	writtenBytes, err := io.Copy(file, objectReader)
	err = objectReader.Close()
	if err != nil {
		fmt.Println("error occurred while downloading blob", err)
	}
	return err == nil, writtenBytes, err
}

func (impl *GCPBlob) getLatestVersion(storageClient *storage.Client, request *BlobStorageRequest, ctx context.Context, config *GcpBlobBaseConfig) (int64, error) {
	fileName := request.SourceKey
	objects := storageClient.Bucket(config.BucketName).Objects(ctx, &storage.Query{
		Versions: true,
		Prefix:   fileName,
	})
	var latestGeneration int64 = 0
	var latestTimestampInMillis int64 = 0
	for {
		objectAttrs, err := objects.Next()
		if err == iterator.Done {
			break
		}
		objectName := objectAttrs.Name
		if objectName != fileName {
			continue
		}
		updatedTime := objectAttrs.Updated
		generation := objectAttrs.Generation
		fileTimestampInMillis := updatedTime.UnixMilli()
		if latestTimestampInMillis == 0 {
			latestTimestampInMillis = fileTimestampInMillis
			latestGeneration = generation
		}
		if fileTimestampInMillis > latestTimestampInMillis {
			latestTimestampInMillis = fileTimestampInMillis
			latestGeneration = generation
		}
	}
	return latestGeneration, nil
}

func (impl *GCPBlob) getGcpObject(storageClient *storage.Client, config *GcpBlobBaseConfig, fileKey string) *storage.ObjectHandle {
	gcpObject := storageClient.Bucket(config.BucketName).Object(fileKey)
	return gcpObject
}

func (impl *GCPBlob) createGcpClient(ctx context.Context, config *GcpBlobBaseConfig) (*storage.Client, error) {
	log.Println("going to create gcp client")
	var storageClient *storage.Client
	var err error
	if config.CredentialFileJsonData != "" {
		storageClient, err = storage.NewClient(ctx, option.WithCredentialsJSON([]byte(config.CredentialFileJsonData)))
	} else {
		storageClient, err = storage.NewClient(ctx)
	}
	if err != nil {
		return nil, err
	}
	return storageClient, err
}
