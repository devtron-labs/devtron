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

package service

import (
	"context"
	"errors"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/repository"
	eventProcessorBean "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type UserDeploymentRequestService interface {
	SaveNewDeployment(ctx context.Context, tx *pg.Tx, deploymentRequest *eventProcessorBean.UserDeploymentRequest) (int, error)
	GetLatestAsyncCdDeployRequestForPipeline(ctx context.Context, deploymentReqId int) (*eventProcessorBean.UserDeploymentRequest, error)
	IsLatestForPipelineId(id, pipelineId int) (isLatest bool, err error)
	GetAllInCompleteRequests(ctx context.Context) ([]*eventProcessorBean.UserDeploymentRequest, error)
}

type UserDeploymentRequestServiceImpl struct {
	userDeploymentRequestRepo repository.UserDeploymentRequestRepository
	logger                    *zap.SugaredLogger
}

func NewUserDeploymentRequestServiceImpl(logger *zap.SugaredLogger,
	userDeploymentRequestRepo repository.UserDeploymentRequestRepository) *UserDeploymentRequestServiceImpl {
	userDeploymentRequestService := &UserDeploymentRequestServiceImpl{
		logger:                    logger,
		userDeploymentRequestRepo: userDeploymentRequestRepo,
	}
	return userDeploymentRequestService
}

func (impl *UserDeploymentRequestServiceImpl) SaveNewDeployment(ctx context.Context, tx *pg.Tx, deploymentRequest *eventProcessorBean.UserDeploymentRequest) (userDeploymentRequestId int, err error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestServiceImpl.SaveNewDeployment")
	defer span.End()
	userDeploymentRequest := adapter.NewUserDeploymentRequest(deploymentRequest)
	err = impl.userDeploymentRequestRepo.Save(newCtx, tx, userDeploymentRequest)
	if err != nil {
		impl.logger.Errorw("error in saving userDeploymentRequest", "asyncCdDeployRequest", deploymentRequest, "err", err)
		return userDeploymentRequestId, err
	}
	deploymentRequest.Id = userDeploymentRequest.Id
	userDeploymentRequestId = userDeploymentRequest.Id
	return userDeploymentRequestId, nil
}

func (impl *UserDeploymentRequestServiceImpl) GetAllInCompleteRequests(ctx context.Context) ([]*eventProcessorBean.UserDeploymentRequest, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestServiceImpl.GetAllInCompleteRequests")
	defer span.End()
	models, err := impl.userDeploymentRequestRepo.GetAllInCompleteRequests(newCtx)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in getting all incomplete userDeploymentRequests", "err", err)
		return nil, err
	}
	response := make([]*eventProcessorBean.UserDeploymentRequest, 0, len(models))
	for _, model := range models {
		response = append(response, adapter.NewAsyncCdDeployRequest(&model.UserDeploymentRequest).
			WithCdWorkflowRunnerId(model.CdWorkflowRunnerId).
			WithPipelineOverrideId(model.PipelineOverrideId))
	}
	return response, nil
}

func (impl *UserDeploymentRequestServiceImpl) GetLatestAsyncCdDeployRequestForPipeline(ctx context.Context, deploymentReqId int) (*eventProcessorBean.UserDeploymentRequest, error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestServiceImpl.GetAllInCompleteRequests")
	defer span.End()
	latestDeploymentReqId, err := impl.userDeploymentRequestRepo.GetLatestIdForPipeline(newCtx, deploymentReqId)
	if err != nil {
		impl.logger.Errorw("error in getting latestDeploymentReqId by previous id", "id", deploymentReqId, "err", err)
		return nil, err
	}
	model, err := impl.userDeploymentRequestRepo.FindById(newCtx, latestDeploymentReqId)
	if err != nil {
		impl.logger.Errorw("error in getting userDeploymentRequest by id", "latestDeploymentReqId", latestDeploymentReqId, "err", err)
		return nil, err
	}
	return adapter.NewAsyncCdDeployRequest(&model.UserDeploymentRequest).
		WithCdWorkflowRunnerId(model.CdWorkflowRunnerId).
		WithPipelineOverrideId(model.PipelineOverrideId), nil
}

func (impl *UserDeploymentRequestServiceImpl) IsLatestForPipelineId(id, pipelineId int) (isLatest bool, err error) {
	isLatest, err = impl.userDeploymentRequestRepo.IsLatestForPipelineId(id, pipelineId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("err in checking latest userDeploymentRequest", "err", err)
		return false, err
	}
	return isLatest, nil
}
