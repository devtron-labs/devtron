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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	s32 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
	"io"
	v12 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"os"
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
	PipelineId   int
	WorkflowId   int
	WorkflowName string
	AccessKey    string
	SecretKet    string
	Region       string
	LogsBucket   string
	LogsFilePath string
	Namespace    string
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
