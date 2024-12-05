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

package repository

import (
	"context"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"time"
)

type PipelineStrategyHistoryRepository interface {
	CreateHistory(model *PipelineStrategyHistory) (*PipelineStrategyHistory, error)
	CreateHistoryWithTxn(model *PipelineStrategyHistory, tx *pg.Tx) (*PipelineStrategyHistory, error)
	GetHistoryForDeployedStrategyById(id, pipelineId int) (*PipelineStrategyHistory, error)
	GetDeploymentDetailsForDeployedStrategyHistory(pipelineId int) ([]*PipelineStrategyHistory, error)
	GetHistoryByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int) (*PipelineStrategyHistory, error)
	CheckIfTriggerHistoryExistsForPipelineIdOnTime(pipelineId int, deployedOn time.Time) (bool, error)
	GetDeployedHistoryList(pipelineId, baseConfigId int) ([]*PipelineStrategyHistory, error)
}

type PipelineStrategyHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPipelineStrategyHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *PipelineStrategyHistoryRepositoryImpl {
	return &PipelineStrategyHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

type PipelineStrategyHistory struct {
	TableName           struct{}                               `sql:"pipeline_strategy_history" pg:",discard_unknown_columns"`
	Id                  int                                    `sql:"id,pk"`
	PipelineId          int                                    `sql:"pipeline_id, notnull"`
	Strategy            chartRepoRepository.DeploymentStrategy `sql:"strategy,notnull"`
	Config              string                                 `sql:"config"`
	Default             bool                                   `sql:"default,notnull"`
	Deployed            bool                                   `sql:"deployed"`
	DeployedOn          time.Time                              `sql:"deployed_on"`
	DeployedBy          int32                                  `sql:"deployed_by"`
	PipelineTriggerType pipelineConfig.TriggerType             `sql:"pipeline_trigger_type"`
	sql.AuditLog
	//getting below data from cd_workflow_runner and users join
	DeploymentStatus  string `sql:"-"`
	DeployedByEmailId string `sql:"-"`
}

func (impl PipelineStrategyHistoryRepositoryImpl) CreateHistory(model *PipelineStrategyHistory) (*PipelineStrategyHistory, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.logger.Errorw("err in creating strategy history entry", "err", err)
		return model, err
	}
	return model, nil
}

func (impl PipelineStrategyHistoryRepositoryImpl) CreateHistoryWithTxn(model *PipelineStrategyHistory, tx *pg.Tx) (*PipelineStrategyHistory, error) {
	err := tx.Insert(model)
	if err != nil {
		impl.logger.Errorw("err in creating strategy history entry", "err", err)
		return model, err
	}
	return model, nil
}

func (impl PipelineStrategyHistoryRepositoryImpl) GetHistoryForDeployedStrategyById(id, pipelineId int) (*PipelineStrategyHistory, error) {
	var history PipelineStrategyHistory
	err := impl.dbConnection.Model(&history).Where("id = ?", id).
		Where("pipeline_id = ?", pipelineId).
		Where("deployed = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("error in getting strategy history", "err", err)
		return &history, err
	}
	return &history, nil
}

func (impl PipelineStrategyHistoryRepositoryImpl) GetDeploymentDetailsForDeployedStrategyHistory(pipelineId int) ([]*PipelineStrategyHistory, error) {
	var histories []*PipelineStrategyHistory
	err := impl.dbConnection.Model(&histories).Where("pipeline_id = ?", pipelineId).
		Where("deployed = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("error in getting strategy history", "err", err)
		return histories, err
	}
	return histories, nil
}

func (impl PipelineStrategyHistoryRepositoryImpl) GetHistoryByPipelineIdAndWfrId(ctx context.Context, pipelineId, wfrId int) (*PipelineStrategyHistory, error) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "PipelineStrategyHistoryRepositoryImpl.GetHistoryByPipelineIdAndWfrId")
	defer span.End()
	var history PipelineStrategyHistory
	err := impl.dbConnection.Model(&history).Join("INNER JOIN cd_workflow_runner cwr ON cwr.started_on = pipeline_strategy_history.deployed_on").
		Where("pipeline_strategy_history.pipeline_id = ?", pipelineId).
		Where("pipeline_strategy_history.deployed = ?", true).
		Where("cwr.id = ?", wfrId).
		Select()
	if err != nil {
		impl.logger.Errorw("error in getting pipeline strategy history by pipelineId & wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return &history, err
	}
	return &history, nil
}

func (impl PipelineStrategyHistoryRepositoryImpl) GetDeployedHistoryList(pipelineId, baseConfigId int) ([]*PipelineStrategyHistory, error) {
	var histories []*PipelineStrategyHistory
	query := "SELECT psh.id, psh.deployed_on, psh.deployed_by, cwr.status as deployment_status, users.email_id as deployed_by_email_id" +
		" FROM pipeline_strategy_history psh" +
		" INNER JOIN cd_workflow_runner cwr ON cwr.started_on = psh.deployed_on" +
		" INNER JOIN users ON users.id = psh.deployed_by" +
		" WHERE psh.pipeline_id = ? AND psh.deployed = true AND psh.id <= ?" +
		" ORDER BY psh.id DESC;"
	_, err := impl.dbConnection.Query(&histories, query, pipelineId, baseConfigId)
	if err != nil {
		impl.logger.Errorw("error in getting pipeline strategy history list by pipelineId", "err", err, "pipelineId", pipelineId)
		return histories, err
	}
	return histories, nil
}

func (impl PipelineStrategyHistoryRepositoryImpl) CheckIfTriggerHistoryExistsForPipelineIdOnTime(pipelineId int, deployedOn time.Time) (bool, error) {
	var history PipelineStrategyHistory
	exists, err := impl.dbConnection.Model(&history).
		Where("pipeline_strategy_history.deployed_on = ?", deployedOn).
		Where("pipeline_strategy_history.pipeline_id = ?", pipelineId).
		Where("pipeline_strategy_history.deployed = ?", true).
		Exists()
	return exists, err
}
