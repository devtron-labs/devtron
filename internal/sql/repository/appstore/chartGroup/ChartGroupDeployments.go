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

package chartGroup

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ChartGroupDeployment struct {
	TableName           struct{} `sql:"chart_group_deployment" pg:",discard_unknown_columns"`
	Id                  int      `sql:"id,pk"`
	ChartGroupId        int      `sql:"chart_group_id"`
	ChartGroupEntryId   int      `sql:"chart_group_entry_id"`
	InstalledAppId      int      `sql:"installed_app_id"`
	GroupInstallationId string   `sql:"group_installation_id"`
	Deleted             bool     `sql:"deleted,notnull"`
	models.AuditLog
}

type ChartGroupDeploymentRepository interface {
	Save(tx *pg.Tx, chartGroupDeployment *ChartGroupDeployment) error
	FindByChartGroupId(chartGroupId int) ([]*ChartGroupDeployment, error)
	Update(model *ChartGroupDeployment) (*ChartGroupDeployment, error)
	FindByInstalledAppId(installedAppId int) (*ChartGroupDeployment, error)
}

type ChartGroupDeploymentRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func (impl *ChartGroupDeploymentRepositoryImpl) Save(tx *pg.Tx, chartGroupDeployment *ChartGroupDeployment) error {
	_, err := tx.Model(chartGroupDeployment).Insert()
	return err
}

func NewChartGroupDeploymentRepositoryImpl(
	dbConnection *pg.DB,
	Logger *zap.SugaredLogger) *ChartGroupDeploymentRepositoryImpl {
	return &ChartGroupDeploymentRepositoryImpl{
		dbConnection: dbConnection,
		Logger:       Logger,
	}
}

func (impl *ChartGroupDeploymentRepositoryImpl) FindByChartGroupId(chartGroupId int) ([]*ChartGroupDeployment, error) {
	var chartGroupDeployments []*ChartGroupDeployment
	err := impl.dbConnection.
		Model(&chartGroupDeployments).
		Column("chart_group_deployment.*").
		Where("chart_group_id = ?", chartGroupId).
		Where("chart_group_deployment.deleted = false").
		Select()
	return chartGroupDeployments, err
}

func (impl *ChartGroupDeploymentRepositoryImpl) Update(model *ChartGroupDeployment) (*ChartGroupDeployment, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}

func (impl *ChartGroupDeploymentRepositoryImpl) FindByInstalledAppId(installedAppId int) (*ChartGroupDeployment, error) {
	var chartGroupDeployments ChartGroupDeployment
	err := impl.dbConnection.
		Model(&chartGroupDeployments).
		Where("installed_app_id = ?", installedAppId).
		Select()
	return &chartGroupDeployments, err
}
