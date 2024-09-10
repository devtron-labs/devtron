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
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ConfigType string

const (
	CONFIGMAP_TYPE ConfigType = "CONFIGMAP"
	SECRET_TYPE    ConfigType = "SECRET"
)

type ConfigMapHistoryRepository interface {
	sql.TransactionWrapper
	CreateHistory(tx *pg.Tx, model *ConfigmapAndSecretHistory) (*ConfigmapAndSecretHistory, error)
	GetHistoryForDeployedCMCSById(id, pipelineId int, configType ConfigType) (*ConfigmapAndSecretHistory, error)
	GetDeploymentDetailsForDeployedCMCSHistory(pipelineId int, configType ConfigType) ([]*ConfigmapAndSecretHistory, error)
	GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId int, configType ConfigType) (*ConfigmapAndSecretHistory, error)
	GetDeployedHistoryForPipelineIdOnTime(pipelineId int, deployedOn time.Time, configType ConfigType) (*ConfigmapAndSecretHistory, error)
	GetDeployedHistoryList(pipelineId, baseConfigId int, configType ConfigType, componentName string) ([]*ConfigmapAndSecretHistory, error)
}

type ConfigMapHistoryRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewConfigMapHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB, transactionUtilImpl *sql.TransactionUtilImpl) *ConfigMapHistoryRepositoryImpl {
	return &ConfigMapHistoryRepositoryImpl{
		logger:              logger,
		dbConnection:        dbConnection,
		TransactionUtilImpl: transactionUtilImpl,
	}
}

type ConfigmapAndSecretHistory struct {
	TableName  struct{}   `sql:"config_map_history" pg:",discard_unknown_columns"`
	Id         int        `sql:"id,pk"`
	PipelineId int        `sql:"pipeline_id"`
	AppId      int        `sql:"app_id"`
	DataType   ConfigType `sql:"data_type"`
	Data       string     `sql:"data"`
	Deployed   bool       `sql:"deployed"`
	DeployedOn time.Time  `sql:"deployed_on"`
	DeployedBy int32      `sql:"deployed_by"`
	sql.AuditLog
	//getting below data from cd_workflow_runner join
	DeploymentStatus  string `sql:"-"`
	DeployedByEmailId string `sql:"-"`
}

func (impl ConfigMapHistoryRepositoryImpl) CreateHistory(tx *pg.Tx, model *ConfigmapAndSecretHistory) (*ConfigmapAndSecretHistory, error) {
	var err error
	if tx != nil {
		err = tx.Insert(model)
	} else {
		err = impl.dbConnection.Insert(model)
	}
	if err != nil {
		impl.logger.Errorw("err in creating env config map/secret history entry", "err", err)
		return model, err
	}
	return model, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetHistoryForDeployedCMCSById(id, pipelineId int, configType ConfigType) (*ConfigmapAndSecretHistory, error) {
	var history ConfigmapAndSecretHistory
	err := impl.dbConnection.Model(&history).Where("id = ?", id).
		Where("pipeline_id = ?", pipelineId).
		Where("data_type = ?", configType).
		Where("deployed = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("error in getting CM/CS history", "err", err)
		return &history, err
	}
	return &history, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetDeploymentDetailsForDeployedCMCSHistory(pipelineId int, configType ConfigType) ([]*ConfigmapAndSecretHistory, error) {
	var histories []*ConfigmapAndSecretHistory
	err := impl.dbConnection.Model(&histories).Where("pipeline_id = ?", pipelineId).
		Where("data_type = ?", configType).
		Where("deployed = ?", true).Select()
	if err != nil {
		impl.logger.Errorw("error in getting deployed CM/CS history", "err", err)
		return histories, err
	}
	return histories, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetHistoryByPipelineIdAndWfrId(pipelineId, wfrId int, configType ConfigType) (*ConfigmapAndSecretHistory, error) {
	var history ConfigmapAndSecretHistory
	query := "SELECT cmh.* FROM config_map_history cmh" +
		" INNER JOIN cd_workflow_runner cwr ON cwr.started_on = cmh.deployed_on" +
		" WHERE cmh.pipeline_id = ? AND cmh.deployed = true AND cmh.data_type = ? AND cwr.id = ?;"
	_, err := impl.dbConnection.Query(&history, query, pipelineId, configType, wfrId)
	if err != nil {
		impl.logger.Errorw("error in getting configmap/secret history by pipelineId & wfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		return &history, err
	}
	return &history, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetDeployedHistoryList(pipelineId, baseConfigId int, configType ConfigType, componentName string) ([]*ConfigmapAndSecretHistory, error) {
	var histories []*ConfigmapAndSecretHistory
	query := "SELECT cmh.id, cmh.deployed_on, cmh.deployed_by, cwr.status as deployment_status, users.email_id as deployed_by_email_id" +
		" FROM config_map_history cmh" +
		" INNER JOIN cd_workflow_runner cwr ON cwr.started_on = cmh.deployed_on" +
		" INNER JOIN users ON users.id = cmh.deployed_by" +
		" WHERE cmh.pipeline_id = ? AND cmh.deployed = true AND cmh.id <= ? AND cmh.data_type = ? AND cmh.data LIKE ?" +
		" ORDER BY cmh.id DESC;"
	_, err := impl.dbConnection.Query(&histories, query, pipelineId, baseConfigId, configType, "%"+fmt.Sprintf("\"name\":\"%s\"", componentName)+"%")
	if err != nil {
		impl.logger.Errorw("error in getting configmap/secret history list by pipelineId", "err", err, "pipelineId", pipelineId)
		return histories, err
	}
	return histories, nil
}

func (impl ConfigMapHistoryRepositoryImpl) GetDeployedHistoryForPipelineIdOnTime(pipelineId int, deployedOn time.Time, configType ConfigType) (*ConfigmapAndSecretHistory, error) {
	var history ConfigmapAndSecretHistory
	err := impl.dbConnection.Model(&history).
		Where("pipeline_id = ?", pipelineId).
		Where("data_type = ?", configType).
		Where("deployed_on = ?", deployedOn).
		Where("deployed = ?", true).
		Select()
	return &history, err
}
