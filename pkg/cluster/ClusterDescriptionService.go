/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package cluster

import (
	"time"

	apiBean "github.com/devtron-labs/devtron/api/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ClusterDescriptionBean struct {
	ClusterId        int                              `json:"clusterId" validate:"number"`
	ClusterName      string                           `json:"clusterName" validate:"required"`
	Description      string                           `json:"description"`
	ServerUrl        string                           `json:"serverUrl"`
	ClusterCreatedBy string                           `json:"clusterCreatedBy" validate:"number"`
	ClusterCreatedOn time.Time                        `json:"clusterCreatedOn" validate:"required"`
	ClusterNote      *apiBean.GenericNoteResponseBean `json:"clusterNote,omitempty"`
}

type ClusterDescriptionService interface {
	FindByClusterIdWithClusterDetails(id int) (*ClusterDescriptionBean, error)
}

type ClusterDescriptionServiceImpl struct {
	clusterDescriptionRepository repository.ClusterDescriptionRepository
	userRepository               repository2.UserRepository
	logger                       *zap.SugaredLogger
}

func NewClusterDescriptionServiceImpl(repository repository.ClusterDescriptionRepository, userRepository repository2.UserRepository, logger *zap.SugaredLogger) *ClusterDescriptionServiceImpl {
	clusterDescriptionService := &ClusterDescriptionServiceImpl{
		clusterDescriptionRepository: repository,
		userRepository:               userRepository,
		logger:                       logger,
	}
	return clusterDescriptionService
}

func (impl *ClusterDescriptionServiceImpl) FindByClusterIdWithClusterDetails(clusterId int) (*ClusterDescriptionBean, error) {
	model, err := impl.clusterDescriptionRepository.FindByClusterIdWithClusterDetails(clusterId)
	if err != nil {
		return nil, err
	}
	clusterCreatedByUser, err := impl.userRepository.GetById(model.ClusterCreatedBy)
	if err != nil {
		impl.logger.Errorw("error in fetching user", "error", err)
		return nil, err
	}
	noteUpdatedByUser, err := impl.userRepository.GetById(model.UpdatedBy)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching user", "error", err)
		return nil, err
	}
	bean := &ClusterDescriptionBean{
		ClusterId:        model.ClusterId,
		ClusterName:      model.ClusterName,
		Description:      model.ClusterDescription,
		ServerUrl:        model.ServerUrl,
		ClusterCreatedBy: clusterCreatedByUser.EmailId,
		ClusterCreatedOn: model.ClusterCreatedOn,
	}
	if model.NoteId > 0 {
		clusterNote := &apiBean.GenericNoteResponseBean{
			Id:          model.NoteId,
			Description: model.Note,
			UpdatedBy:   noteUpdatedByUser.EmailId,
			UpdatedOn:   model.UpdatedOn,
		}
		bean.ClusterNote = clusterNote
	}
	return bean, nil
}
