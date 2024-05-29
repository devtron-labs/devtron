/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package genericNotes

import (
	"github.com/devtron-labs/devtron/pkg/genericNotes/repository"
	"github.com/go-pg/pg"
	"time"

	"go.uber.org/zap"
)

type GenericNoteHistoryBean struct {
	Id          int       `json:"id" validate:"number"`
	NoteId      int       `json:"noteId" validate:"required"`
	Description string    `json:"description" validate:"required"`
	CreatedBy   int32     `json:"createdBy" validate:"number"`
	CreatedOn   time.Time `json:"createdOn" validate:"required"`
}

type GenericNoteHistoryService interface {
	Save(tx *pg.Tx, bean *GenericNoteHistoryBean, userId int32) (*GenericNoteHistoryBean, error)
}

type GenericNoteHistoryServiceImpl struct {
	genericNoteHistoryRepository repository.GenericNoteHistoryRepository
	logger                       *zap.SugaredLogger
}

func NewGenericNoteHistoryServiceImpl(repositoryHistory repository.GenericNoteHistoryRepository, logger *zap.SugaredLogger) *GenericNoteHistoryServiceImpl {
	clusterNoteHistoryService := &GenericNoteHistoryServiceImpl{
		genericNoteHistoryRepository: repositoryHistory,
		logger:                       logger,
	}
	return clusterNoteHistoryService
}

func (impl *GenericNoteHistoryServiceImpl) Save(tx *pg.Tx, bean *GenericNoteHistoryBean, userId int32) (*GenericNoteHistoryBean, error) {
	clusterAudit := &repository.GenericNoteHistory{
		NoteId:      bean.NoteId,
		Description: bean.Description,
	}
	clusterAudit.CreatedBy = userId
	clusterAudit.CreatedOn = time.Now()
	clusterAudit.UpdatedBy = userId
	clusterAudit.UpdatedOn = time.Now()
	err := impl.genericNoteHistoryRepository.SaveHistory(tx, clusterAudit)
	if err != nil {
		impl.logger.Errorw("cluster note history save failed in db", "id", bean.NoteId)
		return nil, err
	}
	return bean, err
}
