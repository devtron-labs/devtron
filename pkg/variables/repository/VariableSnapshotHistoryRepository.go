package repository

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

type VariableSnapshotHistoryRepository interface {
	SaveVariableSnapshots(variableSnapshotHistories []*VariableSnapshotHistory) error
	GetVariableSnapshots(historyReferences []HistoryReference) ([]*VariableSnapshotHistory, error)
}

func (impl VariableSnapshotHistoryRepositoryImpl) SaveVariableSnapshots(variableSnapshotHistories []*VariableSnapshotHistory) error {
	err := impl.dbConnection.Insert(variableSnapshotHistories)
	if err != nil {
		impl.logger.Errorw("err in saving variable snapshot history", "err", err)
		return err
	}
	return nil
}

func (impl VariableSnapshotHistoryRepositoryImpl) GetVariableSnapshots(historyReferences []HistoryReference) ([]*VariableSnapshotHistory, error) {
	var variableSnapshotHistories []*VariableSnapshotHistory

	err := impl.dbConnection.Model(&variableSnapshotHistories).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			for _, historyReference := range historyReferences {
				q = q.WhereOr("history_reference_id = ? AND history_reference_type = ?", historyReference.HistoryReferenceId, historyReference.HistoryReferenceType)
			}
			return q, nil
		}).Select()
	if err != nil {
		impl.logger.Errorw("err in getting variables for entities", "err", err)
		return nil, err
	}
	return variableSnapshotHistories, nil
}

func NewVariableSnapshotHistoryRepository(logger *zap.SugaredLogger, dbConnection *pg.DB) *VariableSnapshotHistoryRepositoryImpl {
	return &VariableSnapshotHistoryRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type VariableSnapshotHistoryRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}
