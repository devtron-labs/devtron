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
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type DeploymentGroupRepository interface {
	Create(model *DeploymentGroup) (*DeploymentGroup, error)
	GetById(id int) (*DeploymentGroup, error)
	GetAll() ([]DeploymentGroup, error)
	Update(model *DeploymentGroup) (*DeploymentGroup, error)
	Delete(model *DeploymentGroup) error
	FindByIdWithApp(id int) (*DeploymentGroup, error)
}

type DeploymentGroupRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewDeploymentGroupRepositoryImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *DeploymentGroupRepositoryImpl {
	return &DeploymentGroupRepositoryImpl{dbConnection: dbConnection, Logger: Logger}
}

type DeploymentGroup struct {
	TableName           struct{} `sql:"deployment_group" pg:",discard_unknown_columns"`
	Id                  int      `sql:"id,pk"`
	Name                string   `sql:"name,notnull"`
	AppCount            int      `sql:"app_count,notnull"`
	NoOfApps            string   `sql:"no_of_apps"`
	EnvironmentId       int      `sql:"environment_id"`
	CiPipelineId        int      `sql:"ci_pipeline_id"`
	Active              bool     `sql:"active,notnull"`
	DeploymentGroupApps []*DeploymentGroupApp
	models.AuditLog
}

func (impl *DeploymentGroupRepositoryImpl) Create(model *DeploymentGroup) (*DeploymentGroup, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl *DeploymentGroupRepositoryImpl) GetById(id int) (*DeploymentGroup, error) {
	var model DeploymentGroup
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Select()
	return &model, err
}

func (impl *DeploymentGroupRepositoryImpl) GetAll() ([]DeploymentGroup, error) {
	var models []DeploymentGroup
	err := impl.dbConnection.Model(&models).Select()
	return models, err
}

func (impl *DeploymentGroupRepositoryImpl) Update(model *DeploymentGroup) (*DeploymentGroup, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl *DeploymentGroupRepositoryImpl) Delete(model *DeploymentGroup) error {
	err := impl.dbConnection.Delete(model)
	if err != nil {
		impl.Logger.Error(err)
		return err
	}
	return nil
}

func (impl *DeploymentGroupRepositoryImpl) FindByIdWithApp(id int) (*DeploymentGroup, error) {
	deploymentGroup := &DeploymentGroup{}
	err := impl.dbConnection.Model(deploymentGroup).Column("deployment_group.*").
		Relation("DeploymentGroupApps", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("active IS TRUE"), nil
		}).
		Where("id =? ", id).Select()
	if err != nil {
		impl.Logger.Errorw("error in fetching apps", "id", id, "err", err)
		return nil, err
	}
	return deploymentGroup, err
}
