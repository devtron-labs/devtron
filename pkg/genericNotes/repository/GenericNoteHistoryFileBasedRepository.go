package repository

import (
	"github.com/go-pg/pg"
)

type GenericNoteHistoryFileBasedRepositoryImpl struct {
}

func NewGenericNoteHistoryFileBasedRepositoryImpl() *GenericNoteHistoryFileBasedRepositoryImpl {
	return &GenericNoteHistoryFileBasedRepositoryImpl{}
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) StartTx() (*pg.Tx, error) {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) RollbackTx(tx *pg.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) CommitTx(tx *pg.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) SaveHistory(tx *pg.Tx, model *GenericNoteHistory) error {
	//TODO implement me
	panic("implement me")
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) FindHistoryByNoteId(id []int) ([]GenericNoteHistory, error) {
	//TODO implement me
	panic("implement me")
}

