package blob_storage

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/devtron-labs/common-lib/utils"
	"log"
	"os"
	"os/exec"
)

type AwsS3Blob struct{}

func (impl *AwsS3Blob) UploadBlob(request *BlobStorageRequest, err error) error {
	s3BaseConfig := request.AwsS3BaseConfig
	var cmdArgs []string
	destinationFileString := fmt.Sprintf("s3://%s/%s", s3BaseConfig.BucketName, request.DestinationKey)
	cmdArgs = append(cmdArgs, "s3", "cp", request.SourceKey, destinationFileString)
	if s3BaseConfig.EndpointUrl != "" {
		cmdArgs = append(cmdArgs, "--endpoint-url", s3BaseConfig.EndpointUrl)
	}
	if s3BaseConfig.Region != "" {
		cmdArgs = append(cmdArgs, "--region", s3BaseConfig.Region)
	}
	command := exec.Command("aws", cmdArgs...)
	err = utils.RunCommand(command)
	return err
}

func (impl *AwsS3Blob) DownloadBlob(request *BlobStorageRequest, downloadSuccess bool, numBytes int64, err error, file *os.File) (bool, int64, error) {
	s3BaseConfig := request.AwsS3BaseConfig
	awsCfg := &aws.Config{
		Region: aws.String(s3BaseConfig.Region),
	}
	if s3BaseConfig.AccessKey != "" {
		awsCfg.Credentials = credentials.NewStaticCredentials(s3BaseConfig.AccessKey, s3BaseConfig.Passkey, "")
	}

	if s3BaseConfig.EndpointUrl != "" { // to handle s3 compatible storage
		awsCfg.Endpoint = aws.String(s3BaseConfig.EndpointUrl)
		awsCfg.DisableSSL = aws.Bool(s3BaseConfig.IsInSecure)
		awsCfg.S3ForcePathStyle = aws.Bool(true)
	}
	sess := session.Must(session.NewSession(awsCfg))
	downloadSuccess, numBytes, err = downLoadFromS3(file, request, sess)
	return downloadSuccess, numBytes, err
}

//TODO KB need to verify for versioning not enabled
func downLoadFromS3(file *os.File, request *BlobStorageRequest, sess *session.Session) (success bool, bytesSize int64, err error) {
	svc := s3.New(sess)
	s3BaseConfig := request.AwsS3BaseConfig
	input := &s3.ListObjectVersionsInput{
		Bucket: aws.String(s3BaseConfig.BucketName),
		Prefix: aws.String(request.SourceKey),
	}
	result, err := svc.ListObjectVersions(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Println(aerr.Error())
			}
		} else {
			log.Println(err.Error())
		}
		return false, 0, err
	}

	var version *string
	var size int64
	for _, v := range result.Versions {
		if *v.IsLatest && *v.Key == request.SourceKey {
			version = v.VersionId
			log.Println("selected version ", v.VersionId, " last modified ", v.LastModified)
			size = *v.Size
			break
		}
	}

	downloader := s3manager.NewDownloader(sess)
	numBytes, err := downloader.Download(file,
		&s3.GetObjectInput{
			Bucket:    aws.String(s3BaseConfig.BucketName),
			Key:       aws.String(request.SourceKey),
			VersionId: version,
		})
	if err != nil {
		log.Println("Couldn't download cache file")
		return false, 0, nil
	}
	log.Println("downloaded ", file.Name(), numBytes, " bytes ")

	if numBytes != size {
		log.Println("cache sizes don't match, skipping step ", " version cache size ", size, " downloaded size ", numBytes)
		return false, 0, nil
	}

	return true, numBytes, nil
}
