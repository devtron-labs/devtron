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
	bean2 "github.com/devtron-labs/devtron/pkg/cluster/environment/bean"
	"github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	"go.uber.org/zap"
)

type EnvironmentReadService interface {
	GetClusterIdByEnvId(envId int) (int, error)
	GetAll() ([]bean2.EnvironmentBean, error)
}

type EnvironmentReadServiceImpl struct {
	logger                *zap.SugaredLogger
	environmentRepository repository.EnvironmentRepository
}

func NewEnvironmentReadServiceImpl(logger *zap.SugaredLogger,
	environmentRepository repository.EnvironmentRepository) *EnvironmentReadServiceImpl {
	return &EnvironmentReadServiceImpl{
		logger:                logger,
		environmentRepository: environmentRepository,
	}
}

func (impl *EnvironmentReadServiceImpl) GetClusterIdByEnvId(envId int) (int, error) {
	model, err := impl.environmentRepository.FindById(envId)
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err, "envId", envId)
		return 0, err
	}
	return model.ClusterId, nil
}

func (impl *EnvironmentReadServiceImpl) GetAll() ([]bean2.EnvironmentBean, error) {
	models, err := impl.environmentRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetching environment", "err", err)
	}
	var beans []bean2.EnvironmentBean
	for _, model := range models {
		beans = append(beans, bean2.EnvironmentBean{
			Id:                    model.Id,
			Environment:           model.Name,
			ClusterId:             model.Cluster.Id,
			ClusterName:           model.Cluster.ClusterName,
			Active:                model.Active,
			PrometheusEndpoint:    model.Cluster.PrometheusEndpoint,
			Namespace:             model.Namespace,
			Default:               model.Default,
			CdArgoSetup:           model.Cluster.CdArgoSetup,
			EnvironmentIdentifier: model.EnvironmentIdentifier,
			Description:           model.Description,
			IsVirtualEnvironment:  model.IsVirtualEnvironment,
		})
	}
	return beans, nil
}
