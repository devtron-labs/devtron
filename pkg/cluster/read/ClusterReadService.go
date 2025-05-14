/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package read

import (
	"github.com/devtron-labs/devtron/pkg/cluster/adapter"
	"github.com/devtron-labs/devtron/pkg/cluster/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/repository"
	"go.uber.org/zap"
)

type ClusterReadService interface {
	IsClusterReachable(clusterId int) (bool, error)
	FindById(id int) (*bean.ClusterBean, error)
	FindOne(clusterName string) (*bean.ClusterBean, error)
	FindByClusterURL(clusterURL string) (*bean.ClusterBean, error)
}

type ClusterReadServiceImpl struct {
	logger            *zap.SugaredLogger
	clusterRepository repository.ClusterRepository
}

func NewClusterReadServiceImpl(logger *zap.SugaredLogger,
	clusterRepository repository.ClusterRepository) *ClusterReadServiceImpl {
	return &ClusterReadServiceImpl{
		logger:            logger,
		clusterRepository: clusterRepository,
	}
}

func (impl *ClusterReadServiceImpl) IsClusterReachable(clusterId int) (bool, error) {
	cluster, err := impl.clusterRepository.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("error in finding cluster from clusterId", "envId", clusterId)
		return false, err
	}
	if len(cluster.ErrorInConnecting) > 0 {
		return false, nil
	}
	return true, nil

}

func (impl *ClusterReadServiceImpl) FindById(id int) (*bean.ClusterBean, error) {
	model, err := impl.clusterRepository.FindById(id)
	if err != nil {
		return nil, err
	}
	bean := adapter.GetClusterBean(*model)
	return &bean, nil
}

func (impl *ClusterReadServiceImpl) FindOne(clusterName string) (*bean.ClusterBean, error) {
	model, err := impl.clusterRepository.FindOne(clusterName)
	if err != nil {
		return nil, err
	}
	bean := adapter.GetClusterBean(*model)
	return &bean, nil
}

func (impl *ClusterReadServiceImpl) FindByClusterURL(clusterURL string) (*bean.ClusterBean, error) {
	model, err := impl.clusterRepository.FindByClusterURL(clusterURL)
	if err != nil {
		return nil, err
	}
	bean := adapter.GetClusterBean(*model)
	return &bean, nil
}
