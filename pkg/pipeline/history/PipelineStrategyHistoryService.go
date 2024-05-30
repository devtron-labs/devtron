/*
 * Copyright (c) 2024. Devtron Inc.
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

package history

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel"
	"time"

	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type PipelineStrategyHistoryService interface {
	CreatePipelineStrategyHistory(pipelineStrategy *chartConfig.PipelineStrategy, pipelineTriggerType pipelineConfig.TriggerType, tx *pg.Tx) (historyModel *repository.PipelineStrategyHistory, err error)
	CreateStrategyHistoryForDeploymentTrigger(strategy *chartConfig.PipelineStrategy, deployedOn time.Time, deployedBy int32, pipelineTriggerType pipelineConfig.TriggerType) error
	GetDeploymentDetailsForDeployedStrategyHistory(pipelineId int) ([]*PipelineStrategyHistoryDto, error)

	GetHistoryForDeployedStrategyById(id, pipelineId int) (*HistoryDetailDto, error)
	CheckIfHistoryExistsForPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int) (historyId int, exists bool, err error)
	GetDeployedHistoryList(pipelineId, baseConfigId int) ([]*DeployedHistoryComponentMetadataDto, error)
	GetLatestDeployedHistoryByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int) (*HistoryDetailDto, error)
}

type PipelineStrategyHistoryServiceImpl struct {
	logger                            *zap.SugaredLogger
	pipelineStrategyHistoryRepository repository.PipelineStrategyHistoryRepository
	userService                       user.UserService
}

func NewPipelineStrategyHistoryServiceImpl(logger *zap.SugaredLogger,
	pipelineStrategyHistoryRepository repository.PipelineStrategyHistoryRepository,
	userService user.UserService) *PipelineStrategyHistoryServiceImpl {
	return &PipelineStrategyHistoryServiceImpl{
		logger:                            logger,
		pipelineStrategyHistoryRepository: pipelineStrategyHistoryRepository,
		userService:                       userService,
	}
}

func (impl PipelineStrategyHistoryServiceImpl) CreatePipelineStrategyHistory(pipelineStrategy *chartConfig.PipelineStrategy, pipelineTriggerType pipelineConfig.TriggerType, tx *pg.Tx) (historyModel *repository.PipelineStrategyHistory, err error) {
	//creating new entry
	historyModel = &repository.PipelineStrategyHistory{
		PipelineId:          pipelineStrategy.PipelineId,
		Strategy:            pipelineStrategy.Strategy,
		Config:              pipelineStrategy.Config,
		Default:             pipelineStrategy.Default,
		Deployed:            false,
		PipelineTriggerType: pipelineTriggerType,
		AuditLog: sql.AuditLog{
			CreatedOn: pipelineStrategy.CreatedOn,
			CreatedBy: pipelineStrategy.CreatedBy,
			UpdatedOn: pipelineStrategy.UpdatedOn,
			UpdatedBy: pipelineStrategy.UpdatedBy,
		},
	}
	if tx != nil {
		_, err = impl.pipelineStrategyHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
	} else {
		_, err = impl.pipelineStrategyHistoryRepository.CreateHistory(historyModel)
	}
	if err != nil {
		impl.logger.Errorw("err in creating history entry for pipeline strategy", "err", err)
		return nil, err
	}
	return historyModel, err
}

func (impl PipelineStrategyHistoryServiceImpl) CreateStrategyHistoryForDeploymentTrigger(pipelineStrategy *chartConfig.PipelineStrategy, deployedOn time.Time, deployedBy int32, pipelineTriggerType pipelineConfig.TriggerType) error {
	//creating new entry
	historyModel := &repository.PipelineStrategyHistory{
		PipelineId:          pipelineStrategy.PipelineId,
		Strategy:            pipelineStrategy.Strategy,
		Config:              pipelineStrategy.Config,
		Default:             pipelineStrategy.Default,
		Deployed:            true,
		DeployedBy:          deployedBy,
		DeployedOn:          deployedOn,
		PipelineTriggerType: pipelineTriggerType,
		AuditLog: sql.AuditLog{
			CreatedOn: deployedOn,
			CreatedBy: deployedBy,
			UpdatedOn: deployedOn,
			UpdatedBy: deployedBy,
		},
	}
	_, err := impl.pipelineStrategyHistoryRepository.CreateHistory(historyModel)
	if err != nil {
		impl.logger.Errorw("err in creating history entry for pipeline strategy", "err", err)
		return err
	}
	return err
}

func (impl PipelineStrategyHistoryServiceImpl) GetDeploymentDetailsForDeployedStrategyHistory(pipelineId int) ([]*PipelineStrategyHistoryDto, error) {
	histories, err := impl.pipelineStrategyHistoryRepository.GetDeploymentDetailsForDeployedStrategyHistory(pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting history for strategy", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historiesDto []*PipelineStrategyHistoryDto
	for _, history := range histories {
		emailId, err := impl.userService.GetEmailById(history.DeployedBy)
		if err != nil {
			impl.logger.Errorw("unable to find user email by id", "err", err, "userId", history.DeployedBy)
			return nil, err
		}
		historyDto := &PipelineStrategyHistoryDto{
			Id:         history.Id,
			PipelineId: history.PipelineId,
			Deployed:   history.Deployed,
			DeployedOn: history.DeployedOn,
			DeployedBy: history.DeployedBy,
			EmailId:    emailId,
		}
		historiesDto = append(historiesDto, historyDto)
	}
	return historiesDto, nil
}

func (impl PipelineStrategyHistoryServiceImpl) CheckIfHistoryExistsForPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int) (historyId int, exists bool, err error) {
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "PipelineStrategyHistoryServiceImpl.CheckIfHistoryExistsForPipelineIdAndWfrId")
	defer span.End()
	impl.logger.Debugw("received request, CheckIfHistoryExistsForPipelineIdAndWfrId", "pipelineId", pipelineId, "wfrId", wfrId)
	// checking if history exists for pipelineId and wfrId
	history, err := impl.pipelineStrategyHistoryRepository.GetHistoryByPipelineIdAndWfrId(newCtx, pipelineId, wfrId)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in checking if history exists for pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return 0, false, err
	} else if errors.Is(err, pg.ErrNoRows) {
		return 0, false, nil
	}
	return history.Id, true, nil
}

func (impl PipelineStrategyHistoryServiceImpl) GetDeployedHistoryList(pipelineId, baseConfigId int) ([]*DeployedHistoryComponentMetadataDto, error) {
	impl.logger.Debugw("received request, GetDeployedHistoryList", "pipelineId", pipelineId, "baseConfigId", baseConfigId)

	//checking if history exists for pipelineId and wfrId
	histories, err := impl.pipelineStrategyHistoryRepository.GetDeployedHistoryList(pipelineId, baseConfigId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting history list for pipelineId and baseConfigId", "err", err, "pipelineId", pipelineId)
		return nil, err
	}
	var historyList []*DeployedHistoryComponentMetadataDto
	for _, history := range histories {
		historyList = append(historyList, &DeployedHistoryComponentMetadataDto{
			Id:               history.Id,
			DeployedOn:       history.DeployedOn,
			DeployedBy:       history.DeployedByEmailId,
			DeploymentStatus: history.DeploymentStatus,
		})
	}
	return historyList, nil
}

func (impl PipelineStrategyHistoryServiceImpl) GetHistoryForDeployedStrategyById(id, pipelineId int) (*HistoryDetailDto, error) {
	history, err := impl.pipelineStrategyHistoryRepository.GetHistoryForDeployedStrategyById(id, pipelineId)
	if err != nil {
		impl.logger.Errorw("error in getting history for strategy", "err", err, "id", id, "pipelineId", pipelineId)
		return nil, err
	}
	historyDto := &HistoryDetailDto{
		Strategy: string(history.Strategy),
		CodeEditorValue: &HistoryDetailConfig{
			DisplayName: "Strategy configuration",
			Value:       history.Config,
		},
	}
	if len(history.PipelineTriggerType) > 0 {
		historyDto.PipelineTriggerType = history.PipelineTriggerType
	}
	return historyDto, nil
}

func (impl PipelineStrategyHistoryServiceImpl) GetLatestDeployedHistoryByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int) (*HistoryDetailDto, error) {
	impl.logger.Debugw("received request, GetLatestDeployedHistoryByPipelineIdAndWfrId", "pipelineId", pipelineId, "wfrId", wfrId)
	newCtx, span := otel.Tracer("orchestrator").Start(ctx, "PipelineStrategyHistoryServiceImpl.GetLatestDeployedHistoryByPipelineIdAndWfrId")
	defer span.End()
	//checking if history exists for pipelineId and wfrId
	history, err := impl.pipelineStrategyHistoryRepository.GetHistoryByPipelineIdAndWfrId(newCtx, pipelineId, wfrId)
	if err != nil {
		impl.logger.Errorw("error in checking if history exists for pipelineId and wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return nil, err
	}
	historyDto := &HistoryDetailDto{
		Strategy: string(history.Strategy),
		CodeEditorValue: &HistoryDetailConfig{
			DisplayName: "Strategy configuration",
			Value:       history.Config,
		},
	}
	if len(history.PipelineTriggerType) > 0 {
		historyDto.PipelineTriggerType = history.PipelineTriggerType
	}
	return historyDto, nil
}
