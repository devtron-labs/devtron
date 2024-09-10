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
	"time"
)

type PrePostCiScriptHistory struct {
	tableName           struct{}  `sql:"pre_post_ci_script_history" pg:",discard_unknown_columns"`
	Id                  int       `sql:"id,pk"`
	CiPipelineScriptsId int       `sql:"ci_pipeline_scripts_id, notnull"`
	Script              string    `sql:"script"`
	Stage               string    `sql:"stage"`
	Name                string    `sql:"name"`
	OutputLocation      string    `sql:"output_location"`
	Built               bool      `sql:"built"`
	BuiltOn             time.Time `sql:"built_on"`
	BuiltBy             int32     `sql:"built_by"`
	sql.AuditLog
}

type PrePostCiScriptHistoryRepository interface {
	CreateHistoryWithTxn(history *PrePostCiScriptHistory, tx *pg.Tx) (*PrePostCiScriptHistory, error)
	CreateHistory(history *PrePostCiScriptHistory) (*PrePostCiScriptHistory, error)
}

type PrePostCiScriptHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewPrePostCiScriptHistoryRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *PrePostCiScriptHistoryRepositoryImpl {
	return &PrePostCiScriptHistoryRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl PrePostCiScriptHistoryRepositoryImpl) CreateHistoryWithTxn(history *PrePostCiScriptHistory, tx *pg.Tx) (*PrePostCiScriptHistory, error) {
	err := tx.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating ci script history entry", "err", err)
		return nil, err
	}
	return history, nil
}
func (impl PrePostCiScriptHistoryRepositoryImpl) CreateHistory(history *PrePostCiScriptHistory) (*PrePostCiScriptHistory, error) {
	err := impl.dbConnection.Insert(history)
	if err != nil {
		impl.logger.Errorw("err in creating ci script history entry", "err", err)
		return nil, err
	}
	return history, nil
}
