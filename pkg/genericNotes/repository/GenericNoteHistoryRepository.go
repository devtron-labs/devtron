/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type GenericNoteHistory struct {
	tableName   struct{} `sql:"generic_note_history" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	NoteId      int      `sql:"note_id"`
	Description string   `sql:"description"`
	sql.AuditLog
}

type GenericNoteHistoryRepository interface {
	sql.TransactionWrapper
	SaveHistory(tx *pg.Tx, model *GenericNoteHistory) error
	FindHistoryByNoteId(id []int) ([]GenericNoteHistory, error)
}

func NewGenericNoteHistoryRepositoryImpl(dbConnection *pg.DB, TransactionUtilImpl *sql.TransactionUtilImpl) *GenericNoteHistoryRepositoryImpl {
	return &GenericNoteHistoryRepositoryImpl{
		dbConnection:        dbConnection,
		TransactionUtilImpl: TransactionUtilImpl,
	}
}

type GenericNoteHistoryRepositoryImpl struct {
	*sql.TransactionUtilImpl
	dbConnection *pg.DB
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
