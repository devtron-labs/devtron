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

package cluster

import (
	"fmt"
	"time"

	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ClusterNoteBean struct {
	Id          int       `json:"id" validate:"number"`
	ClusterId   int       `json:"clusterId" validate:"required"`
	Description string    `json:"description" validate:"required"`
	UpdatedBy   int32     `json:"updatedBy"`
	UpdatedOn   time.Time `json:"updatedOn"`
}

type ClusterNoteResponseBean struct {
	Id          int       `json:"id" validate:"number"`
	Description string    `json:"description" validate:"required"`
	UpdatedBy   string    `json:"updatedBy"`
	UpdatedOn   time.Time `json:"updatedOn"`
}

type ClusterNoteService interface {
	Save(bean *ClusterNoteBean, userId int32) (*ClusterNoteBean, error)
	Update(bean *ClusterNoteBean, userId int32) (*ClusterNoteBean, error)
}

type ClusterNoteServiceImpl struct {
	clusterNoteRepository     repository.ClusterNoteRepository
	clusterNoteHistoryService ClusterNoteHistoryService
	logger                    *zap.SugaredLogger
}

func NewClusterNoteServiceImpl(repository repository.ClusterNoteRepository, clusterNoteHistoryService ClusterNoteHistoryService, logger *zap.SugaredLogger) *ClusterNoteServiceImpl {
	clusterNoteService := &ClusterNoteServiceImpl{
		clusterNoteRepository:     repository,
		clusterNoteHistoryService: clusterNoteHistoryService,
		logger:                    logger,
	}
	return clusterNoteService
}

func (impl *ClusterNoteServiceImpl) Save(bean *ClusterNoteBean, userId int32) (*ClusterNoteBean, error) {
	existingModel, err := impl.clusterNoteRepository.FindByClusterId(bean.ClusterId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Error(err)
		return nil, err
	}
	if existingModel.Id > 0 {
		impl.logger.Errorw("error on fetching cluster, duplicate", "id", bean.ClusterId)
		return nil, fmt.Errorf("cluster note already exists")
	}

	model := &repository.ClusterNote{
		ClusterId:   bean.ClusterId,
		Description: bean.Description,
	}

	model.CreatedBy = userId
	model.UpdatedBy = userId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()

	err = impl.clusterNoteRepository.Save(model)
	if err != nil {
		impl.logger.Errorw("error in saving cluster note in db", "err", err)
		err = &util.ApiError{
			Code:            constants.ClusterCreateDBFailed,
			InternalMessage: "cluster note creation failed in db",
			UserMessage:     fmt.Sprintf("requested by %d", userId),
		}
		return nil, err
	}
	bean.Id = model.Id
	bean.UpdatedBy = model.UpdatedBy
	bean.UpdatedOn = model.UpdatedOn
	// audit the existing description to cluster audit history
	clusterAudit := &ClusterNoteHistoryBean{
		NoteId:      bean.Id,
		Description: bean.Description,
		CreatedOn:   model.CreatedOn,
		CreatedBy:   model.CreatedBy,
	}
	_, _ = impl.clusterNoteHistoryService.Save(clusterAudit, userId)
	return bean, err
}

func (impl *ClusterNoteServiceImpl) Update(bean *ClusterNoteBean, userId int32) (*ClusterNoteBean, error) {
	model, err := impl.clusterNoteRepository.FindByClusterId(bean.ClusterId)
	if err != nil {
		impl.logger.Error(err)
		return nil, err
	}
	if model.Id == 0 {
		impl.logger.Errorw("error on fetching cluster note, not found", "id", bean.Id)
		return nil, fmt.Errorf("cluster note not found")
	}
	// update the cluster description with new data
	model.Description = bean.Description
	model.UpdatedBy = userId
	model.UpdatedOn = time.Now()

	err = impl.clusterNoteRepository.Update(model)
	if err != nil {
		err = &util.ApiError{
			Code:            constants.ClusterUpdateDBFailed,
			InternalMessage: "cluster note update failed in db",
			UserMessage:     fmt.Sprintf("requested by %d", userId),
		}
		return nil, err
	}
	bean.Id = model.Id
	bean.UpdatedBy = model.UpdatedBy
	bean.UpdatedOn = model.UpdatedOn
	// audit the existing description to cluster audit history
	clusterAudit := &ClusterNoteHistoryBean{
		NoteId:      bean.Id,
		Description: bean.Description,
		CreatedOn:   model.CreatedOn,
		CreatedBy:   model.CreatedBy,
	}
	_, _ = impl.clusterNoteHistoryService.Save(clusterAudit, userId)
	return bean, err
}
