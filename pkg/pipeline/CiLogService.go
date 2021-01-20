/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package pipeline

import (
	"context"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	s32 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
	"io"
	v12 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"
)

type CiLogService interface {
	FetchRunningWorkflowLogs(ciLogRequest CiLogRequest, token string, host string, isExt bool) (io.ReadCloser, func() error, error)
	FetchLogs(ciLogRequest CiLogRequest) (*os.File, func() error, error)
}

type CiLogServiceImpl struct {
	logger     *zap.SugaredLogger
	ciService  CiService
	kubeClient kubernetes.Interface
}

type CiLogRequest struct {
	PipelineId      int
	WorkflowId      int
	WorkflowName    string
	AccessKey       string
	SecretKet       string
	Region          string
	LogsBucket      string
	LogsFilePath    string
	Namespace       string
	CloudProvider   string
	AzureBlobConfig *AzureBlobConfig
}

func NewCiLogServiceImpl(logger *zap.SugaredLogger, ciService CiService, ciConfig *CiConfig) *CiLogServiceImpl {
	clientset, err := kubernetes.NewForConfig(ciConfig.ClusterConfig)
	if err != nil {
		logger.Errorw("Can not create kubernetes client: ", "err", err)
		return nil
	}
	return &CiLogServiceImpl{
		logger:     logger,
		ciService:  ciService,
		kubeClient: clientset,
	}
}

func (impl *CiLogServiceImpl) FetchRunningWorkflowLogs(ciLogRequest CiLogRequest, token string, host string, isExt bool) (io.ReadCloser, func() error, error) {
	podLogOpts := &v12.PodLogOptions{
		Container: "main",
		Follow:    true,
	}
	var kubeClient kubernetes.Interface
	kubeClient = impl.kubeClient
	var err error
	if isExt {
		config := &rest.Config{
			Host:        host,
			BearerToken: token,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}
		kubeClient, err = kubernetes.NewForConfig(config)
		if err != nil {
			impl.logger.Errorw("Can not create kubernetes client: ", "err", err)
			return nil, nil, err
		}
	}
	req := kubeClient.CoreV1().Pods(ciLogRequest.Namespace).GetLogs(ciLogRequest.WorkflowName, podLogOpts)
	podLogs, err := req.Stream()
	if podLogs == nil || err != nil {
		impl.logger.Errorw("error in opening stream", "name", ciLogRequest.WorkflowName)
		return nil, nil, err
	}
	cleanUpFunc := func() error {
		impl.logger.Info("closing running pod log stream")
		err = podLogs.Close()
		if err != nil {
			impl.logger.Errorw("err", "err", err)
		}
		return err
	}
	return podLogs, cleanUpFunc, nil
}

func (impl *CiLogServiceImpl) FetchLogs(ciLogRequest CiLogRequest) (*os.File, func() error, error) {
	tempFile := ciLogRequest.WorkflowName + ".log"
	file, err := os.Create(tempFile)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil, nil, err
	}

	if ciLogRequest.CloudProvider == CLOUD_PROVIDER_AWS {
		sess, _ := session.NewSession(&aws.Config{
			Region: aws.String(ciLogRequest.Region),
			//Credentials: credentials.NewStaticCredentials(ciLogRequest.AccessKey, ciLogRequest.SecretKet, ""),
		})

		downloader := s3manager.NewDownloader(sess)
		_, err = downloader.Download(file,
			&s32.GetObjectInput{
				Bucket: aws.String(ciLogRequest.LogsBucket),
				Key:    aws.String(ciLogRequest.LogsFilePath),
			})
	} else if ciLogRequest.CloudProvider == CLOUD_PROVIDER_AWS {
		blobClient := AzureBlob{}
		err=blobClient.DownloadBlob(context.Background(), ciLogRequest.LogsFilePath, ciLogRequest.AzureBlobConfig, file)
		if err != nil {
			impl.logger.Errorw("azure download error", "err", err)
			return nil, nil, err
		}
	} else {
		return nil, nil, fmt.Errorf("unsupported cloud %s", ciLogRequest.CloudProvider)
	}
	//----s3 download start

	// --- s3 download end

	cleanUpFunc := func() error {
		impl.logger.Info("cleaning up log files")
		fErr := file.Close()
		if fErr != nil {
			impl.logger.Errorw("err", "err", fErr)
			return fErr
		}
		fErr = os.Remove(tempFile)
		if fErr != nil {
			impl.logger.Errorw("err", "err", fErr)
			return fErr
		}
		return fErr
	}

	if err != nil {
		impl.logger.Errorw("err", "err", err)
		_ = cleanUpFunc()
		return nil, nil, err
	}
	return file, cleanUpFunc, nil
}

type AzureBlob struct {
}

func (impl *AzureBlob) getSharedCredentials(accountName, accountKey string) (*azblob.SharedKeyCredential, error) {
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}
	return credential, err
}

func (impl *AzureBlob) getTokenCredentials() (azblob.TokenCredential, error) {
	msiEndpoint, err := adal.GetMSIEndpoint()
	if err != nil {
		return nil, fmt.Errorf("failed to get the managed service identity endpoint: %v", err)
	}

	token, err := adal.NewServicePrincipalTokenFromMSI(msiEndpoint, azure.PublicCloud.ResourceIdentifiers.Storage)
	if err != nil {
		return nil, fmt.Errorf("failed to create the managed service identity token: %v", err)
	}
	err = token.Refresh()
	if err != nil {
		return nil, fmt.Errorf("failure refreshing token from MSI endpoint %w", err)
	}

	credential := azblob.NewTokenCredential(token.Token().AccessToken, impl.defaultTokenRefreshFunction(token))
	return credential, err
}

func (impl *AzureBlob) buildContainerUrl(config *AzureBlobConfig) (*azblob.ContainerURL, error) {
	var credential azblob.Credential
	var err error
	if len(config.AccountKey) > 0 {
		credential, err = impl.getSharedCredentials(config.AccountName, config.AccountKey)
		if err != nil {
			return nil, fmt.Errorf("failed in getting credentials: %v", err)
		}
	} else {
		credential, err = impl.getTokenCredentials()
		if err != nil {
			return nil, fmt.Errorf("failed in getting credentials: %v", err)
		}
	}
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// From the Azure portal, get your storage account blob service URL endpoint.
	URL, _ := url.Parse(
		fmt.Sprintf("https://%s.blob.core.windows.net/%s", config.AccountName, config.BlobContainer))

	// Create a ContainerURL object that wraps the container URL and a request
	// pipeline to make requests.
	containerURL := azblob.NewContainerURL(*URL, p)
	return &containerURL, nil
}

func (impl *AzureBlob) DownloadBlob(context context.Context, blobName string, config *AzureBlobConfig, file *os.File) error {
	containerURL, err := impl.buildContainerUrl(config)
	if err != nil {
		return err
	}
	blobURL := containerURL.NewBlobURL(blobName)
	err = azblob.DownloadBlobToFile(context, blobURL, 0, azblob.CountToEnd, file, azblob.DownloadFromBlobOptions{})
	return err
}

func (impl *AzureBlob) UploadBlob(context context.Context, blobName string, config *AzureBlobConfig, cacheFileName string) error {
	containerURL, err := impl.buildContainerUrl(config)
	if err != nil {
		return err
	}
	blobURL := containerURL.NewBlockBlobURL(blobName)
	file, err := os.Open(cacheFileName)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = azblob.UploadFileToBlockBlob(context, file, blobURL, azblob.UploadToBlockBlobOptions{})
	return err
}

func (impl *AzureBlob) defaultTokenRefreshFunction(spToken *adal.ServicePrincipalToken) func(credential azblob.TokenCredential) time.Duration {
	return func(credential azblob.TokenCredential) time.Duration {
		err := spToken.Refresh()
		if err != nil {
			return 0
		}
		expiresIn, err := strconv.ParseInt(string(spToken.Token().ExpiresIn), 10, 64)
		if err != nil {
			return 0
		}
		credential.SetToken(spToken.Token().AccessToken)
		return time.Duration(expiresIn-300) * time.Second
	}
}
