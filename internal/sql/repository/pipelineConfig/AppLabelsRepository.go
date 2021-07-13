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

package pipelineConfig

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
)

type AppLabels struct {
	tableName struct{} `sql:"app_labels" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Label     string   `sql:"label,notnull"`
	AppId     int      `sql:"app_id"`
	Active    bool     `sql:"active,notnull"`
	App       App
	models.AuditLog
}

type AppLabelsRepository interface {
	Create(model *AppLabels) (*AppLabels, error)
	Update(model *AppLabels) (*AppLabels, error)
	FindById(id int) (*AppLabels, error)
	FindAllActive() ([]AppLabels, error)
	FindByLabel(label string) (*AppLabels, error)
	FindByLabels(labels []string) ([]AppLabels, error)
	FindAllActiveByAppId(appId int) ([]AppLabels, error)
}

type AppLabelsRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewAppLabelsRepositoryImpl(dbConnection *pg.DB) *AppLabelsRepositoryImpl {
	return &AppLabelsRepositoryImpl{dbConnection: dbConnection}
}

func (impl AppLabelsRepositoryImpl) Create(model *AppLabels) (*AppLabels, error) {
	err := impl.dbConnection.Insert(model)
	if err != nil {
		return model, err
	}
	return model, nil
}
func (impl AppLabelsRepositoryImpl) Update(model *AppLabels) (*AppLabels, error) {
	err := impl.dbConnection.Update(model)
	if err != nil {
		return model, err
	}

	return model, nil
}
func (impl AppLabelsRepositoryImpl) FindById(id int) (*AppLabels, error) {
	var model AppLabels
	err := impl.dbConnection.Model(&model).Where("id = ?", id).Order("id desc").Limit(1).Select()
	return &model, err
}
func (impl AppLabelsRepositoryImpl) FindAllActive() ([]AppLabels, error) {
	var userModel []AppLabels
	err := impl.dbConnection.Model(&userModel).Where("active=?", true).Order("updated_on desc").Select()
	return userModel, err
}
func (impl AppLabelsRepositoryImpl) FindByLabel(label string) (*AppLabels, error) {
	var model AppLabels
	err := impl.dbConnection.Model(&model).Where("label = ?", label).Select()
	return &model, err
}

func (impl AppLabelsRepositoryImpl) FindByLabels(labels []string) ([]AppLabels, error) {
	if len(labels) == 0 {
		return nil, fmt.Errorf("no labels provided for search")
	}
	var models []AppLabels
	err := impl.dbConnection.Model(&models).Where("labels in (?)", pg.In(labels)).Select()
	return models, err
}

func (impl AppLabelsRepositoryImpl) FindAllActiveByAppId(appId int) ([]AppLabels, error) {
	var models []AppLabels
	err := impl.dbConnection.Model(&models).Where("app_id=?", appId).
		Where("active=?", true).Select()
	return models, err
}
