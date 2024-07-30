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

package variables

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	repository2 "github.com/devtron-labs/devtron/pkg/variables/repository"
	"go.uber.org/zap"
	"time"
)

type VariableSnapshotHistoryService interface {
	SaveVariableHistoriesForTrigger(variableHistories []*repository2.VariableSnapshotHistoryBean, userId int32) error
	GetVariableHistoryForReferences(references []repository2.HistoryReference) (map[repository2.HistoryReference]*repository2.VariableSnapshotHistoryBean, error)
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

func (impl VariableSnapshotHistoryServiceImpl) SaveVariableHistoriesForTrigger(variableHistories []*repository2.VariableSnapshotHistoryBean, userId int32) error {
	variableSnapshotHistoryList := make([]*repository2.VariableSnapshotHistory, 0)
	for _, history := range variableHistories {
		exists, err := impl.repository.CheckIfVariableSnapshotExists(history.HistoryReference)
		if err != nil {
			impl.logger.Errorw("error in checking if variable snapshot exists", "historyReference", history.HistoryReference, "err", err)
			return err
		}
		if exists {
			continue
		}
		variableSnapshotHistoryList = append(variableSnapshotHistoryList, &repository2.VariableSnapshotHistory{
			VariableSnapshotHistoryBean: *history,
			AuditLog: sql.AuditLog{
				CreatedOn: time.Now(),
				CreatedBy: userId,
				UpdatedOn: time.Now(),
				UpdatedBy: userId,
			},
		})
	}
	if len(variableSnapshotHistoryList) > 0 {
		err := impl.repository.SaveVariableSnapshots(variableSnapshotHistoryList)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl VariableSnapshotHistoryServiceImpl) GetVariableHistoryForReferences(references []repository2.HistoryReference) (map[repository2.HistoryReference]*repository2.VariableSnapshotHistoryBean, error) {
	snapshots, err := impl.repository.GetVariableSnapshots(references)
	if err != nil {
		return nil, err
	}
	variableSnapshotHistories := make(map[repository2.HistoryReference]*repository2.VariableSnapshotHistoryBean)
	for _, snapshot := range snapshots {
		variableSnapshotHistories[snapshot.HistoryReference] = &repository2.VariableSnapshotHistoryBean{
			VariableSnapshot: snapshot.VariableSnapshot,
			HistoryReference: snapshot.HistoryReference,
		}
	}
	return variableSnapshotHistories, nil
}
