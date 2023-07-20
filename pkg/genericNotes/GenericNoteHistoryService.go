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

package genericNotes

import (
	"github.com/devtron-labs/devtron/pkg/genericNotes/repository"
	"time"

	"go.uber.org/zap"
)

type ClusterNoteHistoryBean struct {
	Id          int       `json:"id" validate:"number"`
	NoteId      int       `json:"noteId" validate:"required"`
	Description string    `json:"description" validate:"required"`
	CreatedBy   int32     `json:"createdBy" validate:"number"`
	CreatedOn   time.Time `json:"createdOn" validate:"required"`
}

type ClusterNoteHistoryService interface {
	Save(bean *ClusterNoteHistoryBean, userId int32) (*ClusterNoteHistoryBean, error)
}

type ClusterNoteHistoryServiceImpl struct {
	clusterNoteHistoryRepository repository.ClusterNoteHistoryRepository
	logger                       *zap.SugaredLogger
}

func NewClusterNoteHistoryServiceImpl(repositoryHistory repository.ClusterNoteHistoryRepository, logger *zap.SugaredLogger) *ClusterNoteHistoryServiceImpl {
	clusterNoteHistoryService := &ClusterNoteHistoryServiceImpl{
		clusterNoteHistoryRepository: repositoryHistory,
		logger:                       logger,
	}
	return clusterNoteHistoryService
}

func (impl *ClusterNoteHistoryServiceImpl) Save(bean *ClusterNoteHistoryBean, userId int32) (*ClusterNoteHistoryBean, error) {
	clusterAudit := &repository.ClusterNoteHistory{
		NoteId:      bean.NoteId,
		Description: bean.Description,
	}
	clusterAudit.CreatedBy = bean.CreatedBy
	clusterAudit.CreatedOn = bean.CreatedOn
	clusterAudit.UpdatedBy = userId
	clusterAudit.UpdatedOn = time.Now()
	err := impl.clusterNoteHistoryRepository.SaveHistory(clusterAudit)
	if err != nil {
		impl.logger.Errorw("cluster note history save failed in db", "id", bean.NoteId)
		return nil, err
	}
	return bean, err
}
