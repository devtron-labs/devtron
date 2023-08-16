package variables

import (
	repository2 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"go.uber.org/zap"
)

type VariableSnapshotHistoryService interface {
	SaveVariableHistoriesForTrigger(variableHistories []*repository2.VariableSnapshotHistoryBean) error
	GetVariableHistoryForReferences(references []repository2.HistoryReference) ([]*repository2.VariableSnapshotHistoryBean, error)
}

type VariableSnapshotHistoryServiceImpl struct {
	logger     *zap.SugaredLogger
	repository repository2.VariableSnapshotHistoryRepository
}

func NewVariableSnapshotHistoryServiceImpl(repository repository2.VariableSnapshotHistoryRepository, logger *zap.SugaredLogger) *VariableSnapshotHistoryServiceImpl {
	return &VariableSnapshotHistoryServiceImpl{
		repository: repository,
		logger:     logger,
	}
}

func (impl VariableSnapshotHistoryServiceImpl) SaveVariableHistoriesForTrigger(variableHistories []*repository2.VariableSnapshotHistoryBean) error {
	variableSnapshotHistoryList := make([]*repository2.VariableSnapshotHistory, 0)
	for _, history := range variableHistories {
		variableSnapshotHistoryList = append(variableSnapshotHistoryList, &repository2.VariableSnapshotHistory{
			VariableSnapshotHistoryBean: *history,
		})
	}
	err := impl.repository.SaveVariableSnapshots(variableSnapshotHistoryList)
	if err != nil {
		return err
	}
	return nil
}

func (impl VariableSnapshotHistoryServiceImpl) GetVariableHistoryForReferences(references []repository2.HistoryReference) ([]*repository2.VariableSnapshotHistoryBean, error) {
	snapshots, err := impl.repository.GetVariableSnapshots(references)
	if err != nil {
		return nil, err
	}
	variableSnapshotHistories := make([]*repository2.VariableSnapshotHistoryBean, 0)

	for _, snapshot := range snapshots {
		variableSnapshotHistories = append(variableSnapshotHistories, &repository2.VariableSnapshotHistoryBean{
			VariableSnapshot: snapshot.VariableSnapshot,
			HistoryReference: snapshot.HistoryReference,
		})
	}
	return variableSnapshotHistories, nil
}
