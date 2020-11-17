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
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"go.uber.org/zap"
	"time"
)

type ClusterAccountsBean struct {
	Id      int             `json:"id"`
	Account string          `json:"account"`
	Config  json.RawMessage `json:"config"`
	//Environment string          `json:"environment, omitempty"`
	//ClusterName string          `json:"cluster_name, omitempty"`
	//Namespace   string          `json:"namespace, omitempty"`
	Default   bool `json:"default, omitempty"`
	ClusterId int  `json:"cluster_id"`
}

type ClusterAccountsService interface {
	Save(account *ClusterAccountsBean, userId int32) error
	FindOne(clusterName string) (*ClusterAccountsBean, error)
	FindOneByEnvironment(environment string) (*ClusterAccountsBean, error)
	Update(account *ClusterAccountsBean, userId int32) error

	FindById(id int) (*ClusterAccountsBean, error)
	FindAll() ([]ClusterAccountsBean, error)
}

type ClusterAccountsServiceImpl struct {
	clusterAccountsRepository            cluster.ClusterAccountsRepository
	environmentClusterMappingsRepository cluster.EnvironmentRepository
	clusterService                       ClusterService
	logger                               *zap.SugaredLogger
}

func NewClusterAccountsServiceImpl(repository cluster.ClusterAccountsRepository,
	environmentClusterMappingsRepository cluster.EnvironmentRepository,
	clusterService ClusterService, logger *zap.SugaredLogger) *ClusterAccountsServiceImpl {
	return &ClusterAccountsServiceImpl{
		clusterAccountsRepository:            repository,
		logger:                               logger,
		environmentClusterMappingsRepository: environmentClusterMappingsRepository,
		clusterService:                       clusterService,
	}
}

func (impl ClusterAccountsServiceImpl) Save(request *ClusterAccountsBean, userId int32) error {
	cls, err := impl.clusterService.FindById(request.Id)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by name", "err", err)
	}
	var model = cluster.ClusterAccounts{
		Account:   request.Account,
		Config:    string(request.Config),
		ClusterId: cls.Id,
		Active:    true,
		Default:   request.Default,
	}
	model.CreatedBy = userId
	model.UpdatedBy = userId
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	err = impl.clusterAccountsRepository.Save(&model)
	return err
}

func (impl ClusterAccountsServiceImpl) FindOne(clusterName string) (*ClusterAccountsBean, error) {
	account, err := impl.clusterAccountsRepository.FindOne(clusterName)
	if err != nil {
		impl.logger.Errorw("error in getting cluster account", "err", err)
		return nil, err
	}
	bean := &ClusterAccountsBean{
		Account: account.Account,
		Config:  []byte(account.Config),
		Default: account.Default,
	}
	return bean, nil
}

func (impl ClusterAccountsServiceImpl) FindOneByEnvironment(environment string) (*ClusterAccountsBean, error) {
	account, err := impl.clusterAccountsRepository.FindOneByEnvironment(environment)
	if err != nil {
		impl.logger.Errorw("error in getting cluster account", "err", err)
		return nil, err
	}
	bean := &ClusterAccountsBean{
		Account: account.Account,
		Config:  []byte(account.Config),
		Default: account.Default,
	}
	return bean, nil
}

func (impl ClusterAccountsServiceImpl) Update(request *ClusterAccountsBean, userId int32) error {
	model, err := impl.clusterAccountsRepository.FindById(request.Id)
	if err != nil {
		impl.logger.Errorw("error in getting cluster by name", "err", err)
	}
	model.Account = request.Account
	model.Config = string(request.Config)
	model.Active = true
	model.Default = request.Default

	model.UpdatedBy = userId
	model.UpdatedOn = time.Now()
	err = impl.clusterAccountsRepository.Update(model)
	return err
}

func (impl ClusterAccountsServiceImpl) FindById(id int) (*ClusterAccountsBean, error) {
	account, err := impl.clusterAccountsRepository.FindById(id)
	if err != nil {
		impl.logger.Errorw("error in getting cluster account", "err", err)
		return nil, err
	}
	bean := &ClusterAccountsBean{
		Account: account.Account,
		Config:  []byte(account.Config),
		Default: account.Default,
	}
	return bean, nil
}

func (impl ClusterAccountsServiceImpl) FindAll() ([]ClusterAccountsBean, error) {
	accounts, err := impl.clusterAccountsRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in getting cluster account", "err", err)
		return nil, err
	}
	var clusterAccountBeans []ClusterAccountsBean
	for _, account := range accounts {
		clusterAccountBeans = append(clusterAccountBeans, ClusterAccountsBean{
			Account: account.Account,
			Config:  []byte(account.Config),
			Default: account.Default,
		})
	}

	return clusterAccountBeans, nil
}
