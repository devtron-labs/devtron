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
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeploymentGroupAppRepository interface {
	Create(model *DeploymentGroupApp) (*DeploymentGroupApp, error)
	GetById(id int) (*DeploymentGroupApp, error)
	GetAll() ([]*DeploymentGroupApp, error)
	Update(model *DeploymentGroupApp) (*DeploymentGroupApp, error)
	Delete(model *DeploymentGroupApp) error
	GetByDeploymentGroup(deploymentGroupId int) ([]*DeploymentGroupApp, error)
}

type DeploymentGroupAppRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewDeploymentGroupAppRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *DeploymentGroupAppRepositoryImpl {
	return &DeploymentGroupAppRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type DeploymentGroupApp struct {
	TableName         struct{} `sql:"deployment_group_app" pg:",discard_unknown_columns"`
	Id                int      `sql:"id,pk"`
	DeploymentGroupId int      `sql:"deployment_group_id"`
	AppId             int      `sql:"app_id"`
	Active            bool     `sql:"active,notnull"`
	models.AuditLog
}

func (impl DeploymentGroupAppRepositoryImpl) Create(model *DeploymentGroupApp) (*DeploymentGroupApp, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl DeploymentGroupAppRepositoryImpl) GetById(id int) (*DeploymentGroupApp, error) {
	var model DeploymentGroupApp
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Select()
	return &model, err
}

func (impl DeploymentGroupAppRepositoryImpl) GetAll() ([]*DeploymentGroupApp, error) {
	var models []*DeploymentGroupApp
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}

func (impl DeploymentGroupAppRepositoryImpl) Update(model *DeploymentGroupApp) (*DeploymentGroupApp, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl DeploymentGroupAppRepositoryImpl) Delete(model *DeploymentGroupApp) error {
	err := impl.dbConnection.Delete(model)
	if err != nil {
		impl.Logger.Error(err)
		return err
	}
	return nil
}

func (impl DeploymentGroupAppRepositoryImpl) GetByDeploymentGroup(deploymentGroupId int) ([]*DeploymentGroupApp, error) {
	var models []*DeploymentGroupApp
	err := impl.dbConnection.Model(&models).
		Where("deployment_group_id = ?", deploymentGroupId).
		Select()
	return models, err
}
