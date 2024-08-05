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
	"bytes"
	"context"
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
	setAWSEnvironmentVariables(s3BaseConfig, command)
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

// TODO KB need to verify for versioning not enabled
func downLoadFromS3(file *os.File, request *BlobStorageRequest, sess *session.Session) (success bool, bytesSize int64, err error) {
	svc := s3.New(sess)
	s3BaseConfig := request.AwsS3BaseConfig
	var version *string
	var size int64
	if s3BaseConfig.VersioningEnabled {
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

		for _, v := range result.Versions {
			if *v.IsLatest && *v.Key == request.SourceKey {
				version = v.VersionId
				log.Println("selected version ", v.VersionId, " last modified ", v.LastModified)
				size = *v.Size
				break
			}
		}
	}

	downloader := s3manager.NewDownloader(sess)
	input := &s3.GetObjectInput{
		Bucket: aws.String(s3BaseConfig.BucketName),
		Key:    aws.String(request.SourceKey),
	}
	if version != nil {
		input.VersionId = version
	}
	numBytes, err := downloader.Download(file, input)
	if err != nil {
		log.Println("Couldn't download cache file")
		return false, 0, err
	}
	log.Println("downloaded ", file.Name(), numBytes, " bytes ")

	if version != nil && numBytes != size {
		log.Println("cache sizes don't match, skipping step ", " version cache size ", size, " downloaded size ", numBytes)
		return false, 0, nil
	}

	return true, numBytes, nil
}

func (impl *AwsS3Blob) DeleteObjectFromBlob(request *BlobStorageRequest) error {
	s3BaseConfig := request.AwsS3BaseConfig
	var cmdArgs []string
	destinationFileString := fmt.Sprintf("s3://%s/%s", s3BaseConfig.BucketName, request.DestinationKey)
	cmdArgs = append(cmdArgs, "s3", "rm", destinationFileString)
	if s3BaseConfig.EndpointUrl != "" {
		cmdArgs = append(cmdArgs, "--endpoint-url", s3BaseConfig.EndpointUrl)
	}
	if s3BaseConfig.Region != "" {
		cmdArgs = append(cmdArgs, "--region", s3BaseConfig.Region)
	}
	command := exec.Command("aws", cmdArgs...)
	err := utils.RunCommand(command)
	return err
}
func (impl *AwsS3Blob) UploadWithSession(request *BlobStorageRequest) (*s3manager.UploadOutput, error) {

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
	content, err := os.ReadFile(request.SourceKey)
	if err != nil {
		log.Println("error in reading source file", "sourceFile", request.SourceKey, "destinationKey", request.DestinationKey)
		return nil, err
	}
	s3Session := session.New(awsCfg)
	uploader := s3manager.NewUploader(s3Session)
	input := &s3manager.UploadInput{
		Bucket: aws.String(s3BaseConfig.BucketName), // bucket's name
		Key:    aws.String(request.DestinationKey),  // files destination location
		Body:   bytes.NewReader(content),            // content of the file
	}
	output, err := uploader.UploadWithContext(context.Background(), input)
	if err != nil {
		log.Println("error in uploading file to S3", "err", err, "sourceKey", request.SourceKey, "destinationKey", request.DestinationKey)
		return nil, err
	}
	return output, err

}
