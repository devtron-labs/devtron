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
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ClusterNoteHistory struct {
	tableName   struct{} `sql:"cluster_note_history" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	NoteId      int      `sql:"note_id"`
	Description string   `sql:"description"`
	sql.AuditLog
}

type ClusterNoteHistoryRepository interface {
	SaveHistory(model *ClusterNoteHistory) error
	FindHistoryByNoteId(id []int) ([]ClusterNoteHistory, error)
}

func NewClusterNoteHistoryRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *ClusterNoteHistoryRepositoryImpl {
	return &ClusterNoteHistoryRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type ClusterNoteHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func (impl ClusterNoteHistoryRepositoryImpl) SaveHistory(model *ClusterNoteHistory) error {
	return impl.dbConnection.Insert(model)
}

func (impl ClusterNoteHistoryRepositoryImpl) FindHistoryByNoteId(id []int) ([]ClusterNoteHistory, error) {
	var clusterNoteHistories []ClusterNoteHistory
	err := impl.dbConnection.
		Model(&clusterNoteHistories).
		Where("note_id =?", id).
		Select()
	return clusterNoteHistories, err
}
