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

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
)

type OCIRegistryConfig struct {
	tableName             struct{} `sql:"oci_registry_config" pg:",discard_unknown_columns"`
	Id                    int      `sql:"id,pk"`
	DockerArtifactStoreId string   `sql:"docker_artifact_store_id,notnull"`
	RepositoryType        string   `sql:"repository_type,notnull"`
	RepositoryAction      string   `sql:"repository_action,notnull"`
	RepositoryList        string   `sql:"repository_list"`
	IsPublic              bool     `sql:"is_public"`
	Deleted               bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

type OCIRegistryConfigRepository interface {
	Save(config *OCIRegistryConfig, tx *pg.Tx) error
	Update(config *OCIRegistryConfig, tx *pg.Tx) error
	FindByDockerRegistryId(dockerRegistryId string) ([]*OCIRegistryConfig, error)
	FindOneByDockerRegistryIdAndRepositoryType(dockerRegistryId string, repositoryType string) (*OCIRegistryConfig, error)
}

type OCIRegistryConfigRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewOCIRegistryConfigRepositoryImpl(dbConnection *pg.DB) *OCIRegistryConfigRepositoryImpl {
	return &OCIRegistryConfigRepositoryImpl{dbConnection: dbConnection}
}

func (impl OCIRegistryConfigRepositoryImpl) Save(config *OCIRegistryConfig, tx *pg.Tx) error {
	return tx.Insert(config)
}

func (impl OCIRegistryConfigRepositoryImpl) Update(config *OCIRegistryConfig, tx *pg.Tx) error {
	return tx.Update(config)
}

func (impl OCIRegistryConfigRepositoryImpl) FindByDockerRegistryId(dockerRegistryId string) ([]*OCIRegistryConfig, error) {
	var ociRegistryConfig []*OCIRegistryConfig
	err := impl.dbConnection.Model(&ociRegistryConfig).
		Where("docker_artifact_store_id = ?", dockerRegistryId).
		Where("deleted = ?", false).
		Select()
	return ociRegistryConfig, err
}

func (impl OCIRegistryConfigRepositoryImpl) FindOneByDockerRegistryIdAndRepositoryType(dockerRegistryId string, repositoryType string) (*OCIRegistryConfig, error) {
	var ociRegistryConfig OCIRegistryConfig
	err := impl.dbConnection.Model(&ociRegistryConfig).
		Where("docker_artifact_store_id = ?", dockerRegistryId).
		Where("repository_type = ?", repositoryType).
		Where("deleted = ?", false).
		Limit(1).Select()
	return &ociRegistryConfig, err
}
