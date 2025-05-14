/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package certificate

import (
	"context"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/certificate"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/connection"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

type ServiceClient interface {
	CreateCertificate(ctx context.Context, grpcConfig *bean.ArgoGRPCConfig, query *certificate.RepositoryCertificateCreateRequest) (*v1alpha1.RepositoryCertificateList, error)
	DeleteCertificate(ctx context.Context, grpcConfig *bean.ArgoGRPCConfig, query *certificate.RepositoryCertificateQuery, opts ...grpc.CallOption) (*v1alpha1.RepositoryCertificateList, error)
}

type ServiceClientImpl struct {
	logger                  *zap.SugaredLogger
	argoCDConnectionManager connection.ArgoCDConnectionManager
}

func NewServiceClientImpl(
	logger *zap.SugaredLogger,
	argoCDConnectionManager connection.ArgoCDConnectionManager) *ServiceClientImpl {
	return &ServiceClientImpl{
		logger:                  logger,
		argoCDConnectionManager: argoCDConnectionManager,
	}
}

func (c *ServiceClientImpl) getService(ctx context.Context, grpcConfig *bean.ArgoGRPCConfig) (certificate.CertificateServiceClient, error) {
	conn := c.argoCDConnectionManager.GetGrpcClientConnection(grpcConfig)
	//defer conn.Close()
	return certificate.NewCertificateServiceClient(conn), nil
}

func (c *ServiceClientImpl) CreateCertificate(ctx context.Context, grpcConfig *bean.ArgoGRPCConfig, query *certificate.RepositoryCertificateCreateRequest) (*v1alpha1.RepositoryCertificateList, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := c.getService(ctx, grpcConfig)
	if err != nil {
		return nil, err
	}
	return client.CreateCertificate(ctx, query)
}

func (c *ServiceClientImpl) DeleteCertificate(ctx context.Context, grpcConfig *bean.ArgoGRPCConfig, query *certificate.RepositoryCertificateQuery, opts ...grpc.CallOption) (*v1alpha1.RepositoryCertificateList, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	client, err := c.getService(ctx, grpcConfig)
	if err != nil {
		return nil, err
	}
	return client.DeleteCertificate(ctx, query, opts...)
}
