/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package repository

import (
	"context"
	"errors"
	repocreds "github.com/argoproj/argo-cd/v2/pkg/apiclient/repocreds"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/connection"
	"go.uber.org/zap"
)

type ServiceClient interface {
	CreateRepoCreds(ctx context.Context, query *repocreds.RepoCredsCreateRequest) (*v1alpha1.RepoCreds, error)
}

type ServiceClientImpl struct {
	logger                  *zap.SugaredLogger
	argoCDConnectionManager connection.ArgoCDConnectionManager
}

func NewServiceClientImpl(logger *zap.SugaredLogger, argoCDConnectionManager connection.ArgoCDConnectionManager) *ServiceClientImpl {
	return &ServiceClientImpl{
		logger:                  logger,
		argoCDConnectionManager: argoCDConnectionManager,
	}
}

func (r ServiceClientImpl) getService(ctx context.Context) (repocreds.RepoCredsServiceClient, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := r.argoCDConnectionManager.GetConnection(token)
	//defer conn.Close()
	return repocreds.NewRepoCredsServiceClient(conn), nil
}

func (r ServiceClientImpl) CreateRepoCreds(ctx context.Context, query *repocreds.RepoCredsCreateRequest) (*v1alpha1.RepoCreds, error) {
	ctx, cancel := context.WithTimeout(ctx, argoApplication.TimeoutSlow)
	defer cancel()
	client, err := r.getService(ctx)
	if err != nil {
		return nil, err
	}
	return client.CreateRepositoryCredentials(ctx, query)
}
