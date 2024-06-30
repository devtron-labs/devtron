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

type K8sResourceHistory struct {
	tableName         struct{} `sql:"kubernetes_resource_history" pg:",discard_unknown_columns"`
	Id                int      `sql:"id,pk"`
	AppId             int      `sql:"app_id"`
	AppName           string   `sql:"app_name"`
	EnvId             int      `sql:"env_id"`
	Namespace         string   `sql:"namespace,omitempty"`
	ResourceName      string   `sql:"resource_name,notnull"`
	Kind              string   `sql:"kind,notnull"`
	Group             string   `sql:"group"`
	ForceDelete       bool     `sql:"force_delete, omitempty"`
	ActionType        string   `sql:"action_type"`
	DeploymentAppType string   `sql:"deployment_app_type"`
	sql.AuditLog
}

type K8sResourceHistoryRepository interface {
	SaveK8sResourceHistory(history *K8sResourceHistory) error
}

type K8sResourceHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewK8sResourceHistoryRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *K8sResourceHistoryRepositoryImpl {
	return &K8sResourceHistoryRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

func (repo K8sResourceHistoryRepositoryImpl) SaveK8sResourceHistory(k8sResourceHistory *K8sResourceHistory) error {
	return repo.dbConnection.Insert(k8sResourceHistory)
}
