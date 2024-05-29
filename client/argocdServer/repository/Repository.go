/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package repository

import (
	"context"
	"errors"
	repository2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/repository"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/argoproj/argo-cd/v2/reposerver/apiclient"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/connection"
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
	logger                  *zap.SugaredLogger
	argoCDConnectionManager connection.ArgoCDConnectionManager
}

func NewServiceClientImpl(logger *zap.SugaredLogger, argoCDConnectionManager connection.ArgoCDConnectionManager) *ServiceClientImpl {
	return &ServiceClientImpl{
		logger:                  logger,
		argoCDConnectionManager: argoCDConnectionManager,
	}
}

func (r ServiceClientImpl) getService(ctx context.Context) (repository2.RepositoryServiceClient, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := r.argoCDConnectionManager.GetConnection(token)
	//defer conn.Close()
	return repository2.NewRepositoryServiceClient(conn), nil
}

func (r ServiceClientImpl) List(ctx context.Context, query *repository2.RepoQuery) (*v1alpha1.RepositoryList, error) {
	ctx, cancel := context.WithTimeout(ctx, argoApplication.TimeoutFast)
	defer cancel()
	client, err := r.getService(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListRepositories(ctx, query)
}

func (r ServiceClientImpl) ListApps(ctx context.Context, query *repository2.RepoAppsQuery) (*repository2.RepoAppsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, argoApplication.TimeoutFast)
	defer cancel()
	client, err := r.getService(ctx)
	if err != nil {
		return nil, err
	}
	return client.ListApps(ctx, query)
}

func (r ServiceClientImpl) GetAppDetails(ctx context.Context, query *repository2.RepoAppDetailsQuery) (*apiclient.RepoAppDetailsResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, argoApplication.TimeoutFast)
	defer cancel()
	client, err := r.getService(ctx)
	if err != nil {
		return nil, err
	}
	return client.GetAppDetails(ctx, query)
}

func (r ServiceClientImpl) Create(ctx context.Context, query *repository2.RepoCreateRequest) (*v1alpha1.Repository, error) {
	ctx, cancel := context.WithTimeout(ctx, argoApplication.TimeoutSlow)
	defer cancel()
	client, err := r.getService(ctx)
	if err != nil {
		return nil, err
	}
	return client.CreateRepository(ctx, query)
}

func (r ServiceClientImpl) Update(ctx context.Context, query *repository2.RepoUpdateRequest) (*v1alpha1.Repository, error) {
	ctx, cancel := context.WithTimeout(ctx, argoApplication.TimeoutSlow)
	defer cancel()
	client, err := r.getService(ctx)
	if err != nil {
		return nil, err
	}
	return client.UpdateRepository(ctx, query)
}

func (r ServiceClientImpl) Delete(ctx context.Context, query *repository2.RepoQuery) (*repository2.RepoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, argoApplication.TimeoutSlow)
	defer cancel()
	client, err := r.getService(ctx)
	if err != nil {
		return nil, err
	}
	return client.DeleteRepository(ctx, query)
}
