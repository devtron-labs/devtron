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
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/go-pg/pg"
)

type ClusterAccounts struct {
	tableName struct{} `sql:"cluster_accounts" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	Account   string   `sql:"account"`
	Config    string   `sql:"config"`
	ClusterId int      `sql:"cluster_id"`
	Cluster   Cluster
	//Namespace string `sql:"namespace"`
	Active  bool `sql:"active"`
	Default bool `sql:"is_default"`
	models.AuditLog
}

type ClusterAccountsRepository interface {
	FindOne(clusterName string) (*ClusterAccounts, error)
	FindOneByEnvironment(clusterName string) (*ClusterAccounts, error)
	Save(account *ClusterAccounts) error
	Update(account *ClusterAccounts) error
	FindById(id int) (*ClusterAccounts, error)
	FindAll() ([]ClusterAccounts, error)
}

func NewClusterAccountsRepositoryImpl(dbConnection *pg.DB) *ClusterAccountsRepositoryImpl {
	return &ClusterAccountsRepositoryImpl{dbConnection: dbConnection}
}

type ClusterAccountsRepositoryImpl struct {
	dbConnection *pg.DB
}

func (repositoryImpl ClusterAccountsRepositoryImpl) FindOne(clusterName string) (*ClusterAccounts, error) {
	account := &ClusterAccounts{}
	err := repositoryImpl.dbConnection.
		Model(account).
		Column("cluster_accounts.*", "Cluster").
		Join("inner join cluster c on cluster_accounts.cluster_id = c.id").
		Where("c.cluster_name = ?", clusterName).
		Where("cluster_accounts.is_default = ?", true).
		Where("cluster_accounts.active = ?", true).
		Limit(1).
		Select()
	return account, err
}

func (repositoryImpl ClusterAccountsRepositoryImpl) FindOneByEnvironment(environment string) (*ClusterAccounts, error) {
	account := &ClusterAccounts{}
	err := repositoryImpl.dbConnection.
		Model(account).
		Column("cluster_accounts.*").
		Join("inner join environment ecm on cluster_accounts.cluster_id = ecm.cluster_id").
		Where("ecm.environment_name = ?", environment).
		Where("ecm.active = ?", true).
		Where("cluster_accounts.is_default = ?", true).
		Where("cluster_accounts.active = ?", true).
		Limit(1).
		Select()
	return account, err
}

func (repositoryImpl ClusterAccountsRepositoryImpl) Save(account *ClusterAccounts) error {
	return repositoryImpl.dbConnection.Insert(account)
}

func (repositoryImpl ClusterAccountsRepositoryImpl) Update(account *ClusterAccounts) error {
	return repositoryImpl.dbConnection.Update(account)
}

func (repositoryImpl ClusterAccountsRepositoryImpl) FindById(id int) (*ClusterAccounts, error) {
	account := &ClusterAccounts{}
	err := repositoryImpl.dbConnection.
		Model(account).
		Column("cluster_accounts.*", "Cluster").
		Join("inner join cluster c on cluster_accounts.cluster_id = c.id").
		Where("cluster_accounts.id = ?", id).
		Where("cluster_accounts.is_default = ?", true).
		Where("cluster_accounts.active = ?", true).
		Limit(1).
		Select()
	return account, err
}

func (repositoryImpl ClusterAccountsRepositoryImpl) FindAll() ([]ClusterAccounts, error) {
	var accounts []ClusterAccounts
	err := repositoryImpl.dbConnection.
		Model(&accounts).
		Column("cluster_accounts.*", "Cluster").
		Join("inner join cluster c on cluster_accounts.cluster_id = c.id").
		Select()
	return accounts, err
}
