package blob_storage

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"os"
)

type GCPBlob struct {
}

func (impl *GCPBlob) UploadBlob(request *BlobStorageRequest) error {
	ctx := context.Background()
	file, err := os.Open(request.SourceKey)
	if err != nil {
		return err
	}
	defer file.Close()
	err, gcpObject := getGcpObject(request, ctx)
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

func getGcpObject(request *BlobStorageRequest, ctx context.Context) (error, *storage.ObjectHandle) {
	config := request.GcpBlobBaseConfig
	storageClient, err := createGcpClient(ctx, request)
	if err != nil {
		return err, nil
	}
	gcpObject := storageClient.Bucket(config.BucketName).Object(request.DestinationKey)
	return err, gcpObject
}

func createGcpClient(ctx context.Context, request *BlobStorageRequest) (*storage.Client, error) {
	config := request.GcpBlobBaseConfig
	fmt.Println("going to create gcp client")
	storageClient, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(config.CredentialFileJsonData)))
	if err != nil {
		return nil, err
	}
	return storageClient, err
}

func (impl *GCPBlob) DownloadBlob(request *BlobStorageRequest, file *os.File) (bool, int64, error) {
	ctx := context.Background()
	config := request.GcpBlobBaseConfig
	storageClient, err := createGcpClient(ctx, request)
	if err != nil {
		return false, 0, err
	}
	objects := storageClient.Bucket(config.BucketName).Objects(ctx, &storage.Query{
		Versions: true,
		Prefix:   request.DestinationKey,
	})
	var latestGeneration int64 = 0
	var latestTimestampInMillis int64 = 0
	for {
		objectAttrs, err := objects.Next()
		if err == iterator.Done {
			break
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
	err, gcpObject := getGcpObject(request, ctx)
	if err != nil {
		return false, 0, err
	}
	objectReader, err := gcpObject.If(storage.Conditions{GenerationMatch: latestGeneration}).NewReader(ctx)
	if err != nil {
		return false, 0, err
	}
	defer objectReader.Close()
	writtenBytes, err := io.Copy(file, objectReader)
	return err != nil, writtenBytes, err
}
