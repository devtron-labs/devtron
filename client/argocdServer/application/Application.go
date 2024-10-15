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

package application

import (
	"context"
	"errors"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/connection"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type ServiceClient interface {
	ResourceTree(ctxt context.Context, query *application.ResourcesQuery) (*v1alpha1.ApplicationTree, error)

	// GetArgoClient return argo connection client
	GetArgoClient(ctxt context.Context) (application.ApplicationServiceClient, *grpc.ClientConn, error)

	// Patch an ArgoCd application
	Patch(ctx context.Context, query *application.ApplicationPatchRequest) (*v1alpha1.Application, error)

	// GetResource returns single application resource
	GetResource(ctxt context.Context, query *application.ApplicationResourceRequest) (*application.ApplicationResourceResponse, error)

	// Get returns an application by name
	Get(ctx context.Context, query *application.ApplicationQuery) (*v1alpha1.Application, error)

	// Update updates an application
	Update(ctx context.Context, query *application.ApplicationUpdateRequest) (*v1alpha1.Application, error)

	// Sync syncs an application to its target state
	Sync(ctx context.Context, query *application.ApplicationSyncRequest) (*v1alpha1.Application, error)

	// Delete deletes an application
	Delete(ctx context.Context, query *application.ApplicationDeleteRequest) (*application.ApplicationResponse, error)

	TerminateOperation(ctx context.Context, query *application.OperationTerminateRequest) (*application.OperationTerminateResponse, error)
}

type ServiceClientImpl struct {
	logger                  *zap.SugaredLogger
	argoCDConnectionManager connection.ArgoCDConnectionManager
}

func NewApplicationClientImpl(
	logger *zap.SugaredLogger,
	argoCDConnectionManager connection.ArgoCDConnectionManager) *ServiceClientImpl {
	return &ServiceClientImpl{
		logger:                  logger,
		argoCDConnectionManager: argoCDConnectionManager,
	}
}

func (c *ServiceClientImpl) GetArgoClient(ctxt context.Context) (application.ApplicationServiceClient, *grpc.ClientConn, error) {
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, nil, errors.New("unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	asc := application.NewApplicationServiceClient(conn)
	return asc, conn, nil
}

func (c *ServiceClientImpl) ResourceTree(ctxt context.Context, query *application.ResourcesQuery) (*v1alpha1.ApplicationTree, error) {
	asc, conn, err := c.GetArgoClient(ctxt)
	if err != nil {
		c.logger.Errorw("error getting ArgoCD client", "error", err)
		return nil, err
	}
	defer util.Close(conn, c.logger)
	c.logger.Debugw("GRPC_GET_RESOURCETREE", "req", query)
	resp, err := asc.ResourceTree(ctxt, query)
	if err != nil {
		c.logger.Errorw("GRPC_GET_RESOURCETREE", "req", query, "err", err)
		return nil, err
	}
	return resp, nil
}

func (c ServiceClientImpl) Patch(ctxt context.Context, query *application.ApplicationPatchRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutLazy)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Patch(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) Get(ctx context.Context, query *application.ApplicationQuery) (*v1alpha1.Application, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	newCtx, cancel := context.WithTimeout(ctx, argoApplication.TimeoutFast)
	defer cancel()
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Get(newCtx, query)
	return resp, err
}

func (c ServiceClientImpl) Update(ctxt context.Context, query *application.ApplicationUpdateRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Update(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) Sync(ctxt context.Context, query *application.ApplicationSyncRequest) (*v1alpha1.Application, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, argoApplication.NewErrUnauthorized("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.Sync(ctx, query)
	return resp, err
}

func (c ServiceClientImpl) GetResource(ctxt context.Context, query *application.ApplicationResourceRequest) (*application.ApplicationResourceResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	return asc.GetResource(ctx, query)
}

func (c ServiceClientImpl) Delete(ctxt context.Context, query *application.ApplicationDeleteRequest) (*application.ApplicationResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutSlow)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, errors.New("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	return asc.Delete(ctx, query)
}
func (c ServiceClientImpl) TerminateOperation(ctxt context.Context, query *application.OperationTerminateRequest) (*application.OperationTerminateResponse, error) {
	ctx, cancel := context.WithTimeout(ctxt, argoApplication.TimeoutFast)
	defer cancel()
	token, ok := ctxt.Value("token").(string)
	if !ok {
		return nil, argoApplication.NewErrUnauthorized("Unauthorized")
	}
	conn := c.argoCDConnectionManager.GetConnection(token)
	defer util.Close(conn, c.logger)
	asc := application.NewApplicationServiceClient(conn)
	resp, err := asc.TerminateOperation(ctx, query)
	return resp, err
}
