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
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"go.uber.org/zap"
	"time"
)

type ClusterHelmConfigBean struct {
	ClusterName string `json:"cluster_name"`
	Environment string `json:"environment"`
	TillerUrl   string `json:"tiller_url"`
	TillerCert  string `json:"tiller_cert"`
	TillerKey   string `json:"tiller_key"`
	Active      bool   `json:"active"`
}

type ClusterHelmConfigService interface {
	Save(clusterHelmConfigBean *ClusterHelmConfigBean, userId int32) error
	FindOneByEnvironment(environment string) (*ClusterHelmConfigBean, error)
}

type ClusterHelmConfigServiceImpl struct {
	clusterHelmConfigRepository cluster.ClusterHelmConfigRepository
	logger                      *zap.SugaredLogger
	clusterService              ClusterService
}

func NewClusterHelmConfigServiceImpl(repository cluster.ClusterHelmConfigRepository, clusterService ClusterService, logger *zap.SugaredLogger) *ClusterHelmConfigServiceImpl {
	return &ClusterHelmConfigServiceImpl{
		clusterHelmConfigRepository: repository,
		logger:                      logger,
		clusterService:              clusterService,
	}
}

func (impl ClusterHelmConfigServiceImpl) Save(bean *ClusterHelmConfigBean, userId int32) error {
	cls, err := impl.clusterService.FindOne(bean.ClusterName)
	if err != nil {
		impl.logger.Errorw("error finding cluster", "err", err)
		return err
	}
	model := &cluster.ClusterHelmConfig{
		TillerUrl:  bean.TillerUrl,
		TillerCert: bean.TillerCert,
		TillerKey:  bean.TillerKey,
		ClusterId:  cls.Id,
		Active:     bean.Active,
	}
	model.CreatedBy = userId
	model.UpdatedBy = userId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	err = impl.clusterHelmConfigRepository.Save(model)
	if err != nil {
		impl.logger.Errorw("error saving helm config", "err", err)
		return err
	}
	return nil
}

func (impl ClusterHelmConfigServiceImpl) FindOneByEnvironment(environment string) (*ClusterHelmConfigBean, error) {
	model, err := impl.clusterHelmConfigRepository.FindOneByEnvironment(environment)
	if err != nil {
		impl.logger.Errorw("error finding helm config by environment", "err", err)
		return nil, err
	}
	bean := &ClusterHelmConfigBean{
		TillerKey:   model.TillerKey,
		TillerCert:  model.TillerCert,
		TillerUrl:   model.TillerUrl,
		Environment: environment,
		Active:      model.Active,
	}
	return bean, nil
}
