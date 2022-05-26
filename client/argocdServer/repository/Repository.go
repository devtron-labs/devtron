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

package repository

import (
	"context"
	"errors"
	repository2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	"github.com/argoproj/argo-cd/v2/util/settings"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"go.uber.org/zap"
)

type ServiceClient interface {
	// List returns list of repos
	List(ctx context.Context, query *repository2.RepoQuery) (*v1alpha1.RepositoryList, error)
	// ListApps returns list of apps in the repo
	ListApps(ctx context.Context, query *repository2.RepoAppsQuery) (*repository2.RepoAppsResponse, error)
	// GetAppDetails returns application details by given path
	GetAppDetails(ctx context.Context, query *repository2.RepoAppDetailsQuery) (*apiclient.RepoAppDetailsResponse, error)
	// Create creates a repo
	Create(ctx context.Context, query *repository2.RepoCreateRequest) (*v1alpha1.Repository, error)
	// Update updates a repo
	Update(ctx context.Context, query *repository2.RepoUpdateRequest) (*v1alpha1.Repository, error)
	// Delete deletes a repo
	Delete(ctx context.Context, query *repository2.RepoQuery) (*repository2.RepoResponse, error)
}

type ServiceClientImpl struct {
	settings *settings.ArgoCDSettings
	logger   *zap.SugaredLogger
}

func NewServiceClientImpl(
	settings *settings.ArgoCDSettings,
	logger *zap.SugaredLogger,
) *ServiceClientImpl {
	return &ServiceClientImpl{
		settings: settings,
		logger:   logger,
	}
}

func getService(ctx context.Context, settings *settings.ArgoCDSettings) (repository2.RepositoryServiceClient, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := argocdServer.GetConnection(token, settings)
	//defer conn.Close()
	return repository2.NewRepositoryServiceClient(conn), nil
}

func (r ServiceClientImpl) List(ctx context.Context, query *repository2.RepoQuery) (*v1alpha1.RepositoryList, error) {
	ctx, cancel := context.WithTimeout(ctx, application.TimeoutFast)
	defer cancel()
	client, err := getService(ctx, r.settings)
	if err != nil {
		return nil, err
	}
	return client.ListRepositories(ctx, query)
}

func (r ServiceClientImpl) ListApps(ctx context.Context, query *repository2.RepoAppsQuery) (*repository2.RepoAppsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, application.TimeoutFast)
	defer cancel()
	client, err := getService(ctx, r.settings)
	if err != nil {
		return nil, err
	}
	return client.ListApps(ctx, query)
}

func (r ServiceClientImpl) GetAppDetails(ctx context.Context, query *repository2.RepoAppDetailsQuery) (*apiclient.RepoAppDetailsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, application.TimeoutFast)
	defer cancel()
	client, err := getService(ctx, r.settings)
	if err != nil {
		return nil, err
	}
	return client.GetAppDetails(ctx, query)
}

func (r ServiceClientImpl) Create(ctx context.Context, query *repository2.RepoCreateRequest) (*v1alpha1.Repository, error) {
	ctx, cancel := context.WithTimeout(ctx, application.TimeoutSlow)
	defer cancel()
	client, err := getService(ctx, r.settings)
	if err != nil {
		return nil, err
	}
	return client.CreateRepository(ctx, query)
}

func (r ServiceClientImpl) Update(ctx context.Context, query *repository2.RepoUpdateRequest) (*v1alpha1.Repository, error) {
	ctx, cancel := context.WithTimeout(ctx, application.TimeoutSlow)
	defer cancel()
	client, err := getService(ctx, r.settings)
	if err != nil {
		return nil, err
	}
	return client.UpdateRepository(ctx, query)
}

func (r ServiceClientImpl) Delete(ctx context.Context, query *repository2.RepoQuery) (*repository2.RepoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, application.TimeoutSlow)
	defer cancel()
	client, err := getService(ctx, r.settings)
	if err != nil {
		return nil, err
	}
	return client.DeleteRepository(ctx, query)
}
