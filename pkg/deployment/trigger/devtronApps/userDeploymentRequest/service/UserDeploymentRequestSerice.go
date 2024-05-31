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
	"fmt"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/adapter"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/trigger/devtronApps/userDeploymentRequest/repository"
	eventProcessorBean "github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
)

type UserDeploymentRequestService interface {
	SaveNewDeployment(ctx context.Context, asyncCdDeployRequest *eventProcessorBean.AsyncCdDeployRequest) (int, error)
	UpdateStatusForCdWfIds(ctx context.Context, status bean.UserDeploymentRequestStatus, cdWfIds ...int) (err error)
	UpdateStatusOnPipelineDelete(pipelineId int) (err error)
	GetDeployRequestStatusByCdWfId(cdWfId int) (bean.UserDeploymentRequestStatus, error)
	GetLatestAsyncCdDeployRequestForPipeline(deploymentReqId int) (*eventProcessorBean.AsyncCdDeployRequest, error)
	IsLatestForPipelineId(id, pipelineId int) (isLatest bool, err error)
	GetAllInCompleteRequests() ([]*eventProcessorBean.AsyncCdDeployRequest, error)
}

type UserDeploymentRequestServiceImpl struct {
	userDeploymentRequestRepo repository.UserDeploymentRequestRepository
	logger                    *zap.SugaredLogger
}

func NewUserDeploymentRequestServiceImpl(
	repository repository.UserDeploymentRequestRepository,
	logger *zap.SugaredLogger) *UserDeploymentRequestServiceImpl {
	userDeploymentRequestService := &UserDeploymentRequestServiceImpl{
		userDeploymentRequestRepo: repository,
		logger:                    logger,
	}
	return userDeploymentRequestService
}

func (impl *UserDeploymentRequestServiceImpl) SaveNewDeployment(ctx context.Context, asyncCdDeployRequest *eventProcessorBean.AsyncCdDeployRequest) (userDeploymentRequestId int, err error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestServiceImpl.SaveNewDeployment")
	defer span.End()
	userDeploymentRequest := adapter.NewUserDeploymentRequest(asyncCdDeployRequest)
	userDeploymentRequest.Status = bean.DeploymentRequestPending
	err = impl.userDeploymentRequestRepo.Save(newCtx, userDeploymentRequest)
	if err != nil {
		impl.logger.Errorw("error in saving userDeploymentRequest", "asyncCdDeployRequest", asyncCdDeployRequest, "err", err)
		return userDeploymentRequestId, err
	}
	asyncCdDeployRequest.UserDeploymentRequestId = userDeploymentRequest.Id
	userDeploymentRequestId = userDeploymentRequest.Id
	return userDeploymentRequestId, nil
}

func (impl *UserDeploymentRequestServiceImpl) UpdateStatusForCdWfIds(ctx context.Context, status bean.UserDeploymentRequestStatus, cdWfIds ...int) (err error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "UserDeploymentRequestServiceImpl.UpdateStatusForCdWfIds")
	defer span.End()
	models, err := impl.userDeploymentRequestRepo.FindByCdWfIds(newCtx, cdWfIds...)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in getting userDeploymentRequests by cdWfIds", "cdWfIds", cdWfIds, "err", err)
		return err
	}
	if errors.Is(err, pg.ErrNoRows) {
		return nil
	}
	tx, err := impl.userDeploymentRequestRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update userDeploymentRequest", "error", err)
		return err
	}
	defer impl.userDeploymentRequestRepo.RollbackTx(tx)
	var validCdWfIds []int
	for _, model := range models {
		if model.Status == status {
			// no change in status, skipping
			continue
		}
		// pre-condition failed
		isValid := validateStatusUpdate(model.Status, status)
		if !isValid {
			return fmt.Errorf("invalid status update request from %s to %s", model.Status, status)
		}
		validCdWfIds = append(validCdWfIds, model.CdWorkflowId)
		if status.IsCompleted() {
			_, err = impl.userDeploymentRequestRepo.MarkAllPreviousSuperseded(newCtx, tx, model.PipelineId, model.Id)
			if err != nil {
				impl.logger.Errorw("error in marking previous userDeploymentRequest superseded", "pipelineId", model.PipelineId, "userDeploymentRequestId", model.Id, "err", err)
				return err
			}
		}
	}
	_, err = impl.userDeploymentRequestRepo.UpdateStatusForCdWfIds(newCtx, tx, status, validCdWfIds...)
	if err != nil {
		impl.logger.Errorw("error in updating userDeploymentRequest status", "status", status, "cdWfIds", cdWfIds, "err", err)
		return err
	}
	err = impl.userDeploymentRequestRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update userDeploymentRequest", "error", err)
		return err
	}
	return nil
}

func (impl *UserDeploymentRequestServiceImpl) UpdateStatusOnPipelineDelete(pipelineId int) (err error) {
	_, err = impl.userDeploymentRequestRepo.TerminateForPipelineId(nil, pipelineId)
	if err != nil {
		impl.logger.Errorw("error in updating terminated status for deleted pipeline", "pipelineId", pipelineId, "err", err)
		return err
	}
	return nil
}

func (impl *UserDeploymentRequestServiceImpl) GetAllInCompleteRequests() ([]*eventProcessorBean.AsyncCdDeployRequest, error) {
	models, err := impl.userDeploymentRequestRepo.GetAllInCompleteRequests()
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in getting all incomplete userDeploymentRequests", "err", err)
		return nil, err
	}
	response := make([]*eventProcessorBean.AsyncCdDeployRequest, 0, len(models))
	for _, model := range models {
		response = append(response, adapter.NewAsyncCdDeployRequest(&model.UserDeploymentRequest).
			WithCdWorkflowRunnerId(model.CdWorkflowRunnerId).
			WithPipelineOverrideId(model.PipelineOverrideId))
	}
	return response, nil
}

func validateStatusUpdate(curr, dest bean.UserDeploymentRequestStatus) (isAllowed bool) {
	if curr == dest {
		return true
	}
	switch curr {
	case bean.DeploymentRequestPending:
		if !slices.Contains([]bean.UserDeploymentRequestStatus{bean.DeploymentRequestTriggerAuditCompleted, bean.DeploymentRequestTriggered, bean.DeploymentRequestCompleted, bean.DeploymentRequestSuperseded}, dest) {
			return false
		}
	case bean.DeploymentRequestTriggerAuditCompleted:
		if !slices.Contains([]bean.UserDeploymentRequestStatus{bean.DeploymentRequestTriggered, bean.DeploymentRequestCompleted, bean.DeploymentRequestSuperseded}, dest) {
			return false
		}
	case bean.DeploymentRequestTriggered:
		if !slices.Contains([]bean.UserDeploymentRequestStatus{bean.DeploymentRequestCompleted, bean.DeploymentRequestSuperseded}, dest) {
			return false
		}
	case bean.DeploymentRequestCompleted:
	case bean.DeploymentRequestSuperseded:
	case bean.DeploymentRequestFailed:
	case bean.DeploymentRequestTerminated:
	default:
		return false
	}
	return true
}

func (impl *UserDeploymentRequestServiceImpl) GetDeployRequestStatusByCdWfId(cdWfId int) (bean.UserDeploymentRequestStatus, error) {
	model, err := impl.userDeploymentRequestRepo.FindByCdWfId(cdWfId)
	if err != nil {
		impl.logger.Errorw("error in getting userDeploymentRequest by cdWfId", "cdWfId", cdWfId, "err", err)
		return "", err
	}
	return model.Status, nil
}

func (impl *UserDeploymentRequestServiceImpl) GetLatestAsyncCdDeployRequestForPipeline(deploymentReqId int) (*eventProcessorBean.AsyncCdDeployRequest, error) {
	latestDeploymentReqId, err := impl.userDeploymentRequestRepo.GetLatestIdForPipeline(deploymentReqId)
	if err != nil {
		impl.logger.Errorw("error in getting latestDeploymentReqId by previous id", "id", deploymentReqId, "err", err)
		return nil, err
	}
	model, err := impl.userDeploymentRequestRepo.FindById(latestDeploymentReqId)
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