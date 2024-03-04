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

package models

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type HelmValues struct {
	tableName         struct{}  `sql:"helm_values"`
	AppName           string    `sql:"app_name,pk"`
	TargetEnvironment string    `sql:"environment,pk"` //target environment
	Values            string    `sql:"values_yaml"`
	Active            bool      `sql:"active,notnull"`
	CreatedOn         time.Time `sql:"created_on"`
	CreatedBy         int32     `sql:"created_by"`
	UpdatedOn         time.Time `sql:"updated_on"`
	UpdatedBy         int32     `sql:"updated_by"`
}

type HelmValuesService interface {
	AddHelmValues(manifest *HelmValues) error
	GetHelmValues(appName, targetEnvironment string) (*HelmValues, error)
}

type HelmValuesServiceImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewHelmValuesServiceImpl(Logger *zap.SugaredLogger, dbConnection *pg.DB) *HelmValuesServiceImpl {
	return &HelmValuesServiceImpl{dbConnection: dbConnection, Logger: Logger}
}

func (impl HelmValuesServiceImpl) AddHelmValues(manifest *HelmValues) error {
	err := impl.dbConnection.Insert(manifest)
	if err != nil {
		impl.Logger.Errorw("error in db insert", "err", err)
	}
	return err
}

func (impl HelmValuesServiceImpl) GetHelmValues(appName, targetEnvironment string) (*HelmValues, error) {
	hv := &HelmValues{
		AppName:           appName,
		TargetEnvironment: targetEnvironment,
	}
	err := impl.dbConnection.Select(hv)
	return hv, err
}
