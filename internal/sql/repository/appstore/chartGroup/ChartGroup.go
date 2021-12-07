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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type ChartGroup struct {
	TableName   struct{} `sql:"chart_group" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	Name        string   `sql:"name"`
	Description string   `sql:"description,notnull"`
	sql.AuditLog
	ChartGroupEntries []*ChartGroupEntry
}

type ChartGroupReposotoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewChartGroupReposotoryImpl(dbConnection *pg.DB, Logger *zap.SugaredLogger) *ChartGroupReposotoryImpl {
	return &ChartGroupReposotoryImpl{
		dbConnection: dbConnection,
		Logger:       Logger,
	}
}

type ChartGroupReposotory interface {
	Save(model *ChartGroup) (*ChartGroup, error)
	Update(model *ChartGroup) (*ChartGroup, error)
	FindByIdWithEntries(chertGroupId int) (*ChartGroup, error)
	FindById(chartGroupId int) (*ChartGroup, error)
	GetAll(max int) ([]*ChartGroup, error)
}

func (impl *ChartGroupReposotoryImpl) Save(model *ChartGroup) (*ChartGroup, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		impl.Logger.Error(err)
		return model, err
	}
	return model, nil
}
func (impl *ChartGroupReposotoryImpl) Update(model *ChartGroup) (*ChartGroup, error) {
	_, err := impl.dbConnection.Model(model).WherePK().UpdateNotNull()
	return model, err
}
func (impl *ChartGroupReposotoryImpl) FindByIdWithEntries(chertGroupId int) (*ChartGroup, error) {
	var ChartGroup ChartGroup
	err := impl.dbConnection.Model(&ChartGroup).
		Column("chart_group.*").
		Relation("ChartGroupEntries", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS false"), nil
		}).
		Where("id = ?", chertGroupId).
		Select()
	return &ChartGroup, err
}

func (impl *ChartGroupReposotoryImpl) FindById(chartGroupId int) (*ChartGroup, error) {
	var ChartGroup ChartGroup
	err := impl.dbConnection.Model(&ChartGroup).Where("id = ?", chartGroupId).Select()
	return &ChartGroup, err
}

func (impl *ChartGroupReposotoryImpl) GetAll(max int) ([]*ChartGroup, error) {
	var chartGroups []*ChartGroup
	query := impl.dbConnection.Model(&chartGroups)
	if max > 0 {
		query = query.Limit(max)
	}
	err := query.Select()
	return chartGroups, err
}
