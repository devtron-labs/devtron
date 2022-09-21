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
	blob_storage "github.com/devtron-labs/common-lib/blob-storage"
	"go.uber.org/zap"
	"io"
	v12 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
)

type CiLogService interface {
	FetchRunningWorkflowLogs(ciLogRequest BuildLogRequest, token string, host string, isExt bool) (io.ReadCloser, func() error, error)
	FetchLogs(ciLogRequest BuildLogRequest) (*os.File, func() error, error)
}

type CiLogServiceImpl struct {
	logger     *zap.SugaredLogger
	ciService  CiService
	kubeClient kubernetes.Interface
}

type BuildLogRequest struct {
	PipelineId        int
	WorkflowId        int
	WorkflowName      string
	LogsFilePath      string
	Namespace         string
	CloudProvider     blob_storage.BlobStorageType
	AwsS3BaseConfig   *blob_storage.AwsS3BaseConfig
	AzureBlobConfig   *blob_storage.AzureBlobBaseConfig
	GcpBlobBaseConfig *blob_storage.GcpBlobBaseConfig
	MinioEndpoint     string
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

func (impl *CiLogServiceImpl) FetchRunningWorkflowLogs(ciLogRequest BuildLogRequest, token string, host string, isExt bool) (io.ReadCloser, func() error, error) {
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
	podLogs, err := req.Stream(context.Background())
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

func (impl *CiLogServiceImpl) FetchLogs(logRequest BuildLogRequest) (*os.File, func() error, error) {

	tempFile := logRequest.WorkflowName + ".log"
	blobStorageService := blob_storage.NewBlobStorageServiceImpl(nil)
	request := &blob_storage.BlobStorageRequest{
		StorageType:         logRequest.CloudProvider,
		SourceKey:           logRequest.LogsFilePath,
		DestinationKey:      tempFile,
		AzureBlobBaseConfig: logRequest.AzureBlobConfig,
		AwsS3BaseConfig:     logRequest.AwsS3BaseConfig,
		GcpBlobBaseConfig:   logRequest.GcpBlobBaseConfig,
	}

	_, _, err := blobStorageService.Get(request)
	if err != nil {
		impl.logger.Errorw("err occurred while downloading logs file", "request", request, "err", err)
		return nil, nil, err
	}

	file, err := os.Open(tempFile)
	if err != nil {
		impl.logger.Errorw("err", "err", err)
		return nil, nil, err
	}

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
