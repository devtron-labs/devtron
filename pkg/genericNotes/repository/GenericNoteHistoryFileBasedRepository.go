package repository

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type GenericNoteHistoryFileBasedRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

func NewGenericNoteHistoryFileBasedRepositoryImpl(logger *zap.SugaredLogger) *GenericNoteHistoryFileBasedRepositoryImpl {
	return &GenericNoteHistoryFileBasedRepositoryImpl{}
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) StartTx() (*pg.Tx, error) {
	return nil, nil
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) RollbackTx(tx *pg.Tx) error {
	return nil
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) CommitTx(tx *pg.Tx) error {
	return nil
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) SaveHistory(tx *pg.Tx, model *GenericNoteHistory) error {
	result := impl.dbConnection.Create(model)
	return result.Error
}

func (impl GenericNoteHistoryFileBasedRepositoryImpl) FindHistoryByNoteId(id []int) ([]GenericNoteHistory, error) {
	var clusterNoteHistories []GenericNoteHistory
	result := impl.dbConnection.
		Where("note_id =?", id).
		Find(&clusterNoteHistories)
	err := result.Error
	return clusterNoteHistories, err
}

