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

type GenericNoteHistory struct {
	tableName   struct{} `sql:"generic_note_history" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	NoteId      int      `sql:"note_id"`
	Description string   `sql:"description"`
	sql.AuditLog
}

type GenericNoteHistoryRepository interface {
	SaveHistory(tx *pg.Tx, model *GenericNoteHistory) error
	FindHistoryByNoteId(id []int) ([]GenericNoteHistory, error)
}

func NewGenericNoteHistoryRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *GenericNoteHistoryRepositoryImpl {
	return &GenericNoteHistoryRepositoryImpl{
		dbConnection: dbConnection,
		logger:       logger,
	}
}

type GenericNoteHistoryRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func (impl GenericNoteHistoryRepositoryImpl) SaveHistory(tx *pg.Tx, model *GenericNoteHistory) error {
	return tx.Insert(model)
}

func (impl GenericNoteHistoryRepositoryImpl) FindHistoryByNoteId(id []int) ([]GenericNoteHistory, error) {
	var clusterNoteHistories []GenericNoteHistory
	err := impl.dbConnection.
		Model(&clusterNoteHistories).
		Where("note_id =?", id).
		Select()
	return clusterNoteHistories, err
}
