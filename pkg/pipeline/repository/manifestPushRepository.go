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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ManifestPushConfig struct {
	tableName         struct{} `sql:"manifest_push_config" pg:",discard_unknown_columns"`
	Id                int      `sql:"id,pk"`
	AppId             int      `sql:"app_id"`
	EnvId             int      `sql:"env_id"`
	CredentialsConfig string   `sql:"credentials_config"`
	ChartName         string   `sql:"chart_name"`
	ChartBaseVersion  string   `sql:"chart_base_version"`
	StorageType       string   `sql:"storage_type"`
	Deleted           bool     `sql:"deleted, notnull"`
	sql.AuditLog
}

type ManifestPushConfigRepository interface {
	SaveConfig(manifestPushConfig *ManifestPushConfig) (*ManifestPushConfig, error)
	GetManifestPushConfigByAppIdAndEnvId(appId, envId int) (*ManifestPushConfig, error)
}

type ManifestPushConfigRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewManifestPushConfigRepository(logger *zap.SugaredLogger,
	dbConnection *pg.DB,
) *ManifestPushConfigRepositoryImpl {
	return &ManifestPushConfigRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

func (impl ManifestPushConfigRepositoryImpl) SaveConfig(manifestPushConfig *ManifestPushConfig) (*ManifestPushConfig, error) {
	err := impl.dbConnection.Insert(manifestPushConfig)
	if err != nil {
		return manifestPushConfig, err
	}
	return manifestPushConfig, err
}

func (impl ManifestPushConfigRepositoryImpl) GetManifestPushConfigByAppIdAndEnvId(appId, envId int) (*ManifestPushConfig, error) {
	var manifestPushConfig *ManifestPushConfig
	err := impl.dbConnection.Model(manifestPushConfig).
		Where("app_id = ? ", appId).
		Where("env_id = ? ", envId).
		Select()
	if err != nil && err != pg.ErrNoRows {
		return manifestPushConfig, err
	}
	return manifestPushConfig, nil
}
