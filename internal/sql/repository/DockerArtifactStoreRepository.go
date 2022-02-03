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
	"net/url"

	"github.com/go-pg/pg"
	"github.com/pkg/errors"
)

const REGISTRYTYPE_ECR = "ecr"
const REGISTRYTYPE_OTHER = "other"
const REGISTRYTYPE_DOCKER_HUB = "docker-hub"

type RegistryType string

type DockerArtifactStore struct {
	tableName          struct{}     `sql:"docker_artifact_store" json:",omitempty"  pg:",discard_unknown_columns"`
	Id                 string       `sql:"id,pk" json:"id,,omitempty"`
	PluginId           string       `sql:"plugin_id,notnull" json:"pluginId,omitempty"`
	RegistryURL        string       `sql:"registry_url" json:"registryUrl,omitempty"`
	RegistryType       RegistryType `sql:"registry_type,notnull" json:"registryType,omitempty"`
	AWSAccessKeyId     string       `sql:"aws_accesskey_id" json:"awsAccessKeyId,omitempty" `
	AWSSecretAccessKey string       `sql:"aws_secret_accesskey" json:"awsSecretAccessKey,omitempty"`
	AWSRegion          string       `sql:"aws_region" json:"awsRegion,omitempty"`
	Username           string       `sql:"username" json:"username,omitempty"`
	Password           string       `sql:"password" json:"password,omitempty"`
	IsDefault          bool         `sql:"is_default,notnull" json:"isDefault"`
	Connection         string       `sql:"connection" json:"connection,omitempty"`
	Cert               string       `sql:"cert" json:"cert,omitempty"`
	Active             bool         `sql:"active,notnull" json:"active"`
	sql.AuditLog
}

func (store *DockerArtifactStore) GetRegistryLocation() (registryLocation string, err error) {
	u, err := url.Parse(registryLocation)
	if err != nil {
		return "", err
	} else {
		return u.Host, nil
	}
}

type DockerArtifactStoreRepository interface {
	Save(artifactStore *DockerArtifactStore) error
	FindActiveDefaultStore() (*DockerArtifactStore, error)
	FindAllActiveForAutocomplete() ([]DockerArtifactStore, error)
	FindAll() ([]DockerArtifactStore, error)
	FindOne(storeId string) (*DockerArtifactStore, error)
	Update(artifactStore *DockerArtifactStore) error
	Delete(storeId string) error
	MarkRegistryDeleted(artifactStore *DockerArtifactStore) error
}
type DockerArtifactStoreRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewDockerArtifactStoreRepositoryImpl(dbConnection *pg.DB) *DockerArtifactStoreRepositoryImpl {
	return &DockerArtifactStoreRepositoryImpl{dbConnection: dbConnection}
}

func (impl DockerArtifactStoreRepositoryImpl) Save(artifactStore *DockerArtifactStore) error {
	//TODO check for unique default
	//there can be only one default
	model, err := impl.FindActiveDefaultStore()
	if err == pg.ErrNoRows {
		artifactStore.IsDefault = true
	} else if err == nil && model.Id != artifactStore.Id && artifactStore.IsDefault == true {
		model.IsDefault = false
		err = impl.Update(model)
		if err != nil {
			return err
		}
	}
	return impl.dbConnection.Insert(artifactStore)
}

func (impl DockerArtifactStoreRepositoryImpl) FindActiveDefaultStore() (*DockerArtifactStore, error) {
	store := &DockerArtifactStore{}
	err := impl.dbConnection.Model(store).
		Where("is_default = ?", true).
		Where("active = ?", true).Select()
	return store, err
}

func (impl DockerArtifactStoreRepositoryImpl) FindAllActiveForAutocomplete() ([]DockerArtifactStore, error) {
	var providers []DockerArtifactStore
	err := impl.dbConnection.Model(&providers).
		Where("active = ?", true).
		Column("id", "registry_url", "is_default").
		Select()
	return providers, err
}

func (impl DockerArtifactStoreRepositoryImpl) FindAll() ([]DockerArtifactStore, error) {
	var providers []DockerArtifactStore
	err := impl.dbConnection.Model(&providers).
		Where("active = ?", true).
		//Column("id", "plugin_id","registry_url", "registry_type","aws_accesskey_id","aws_secret_accesskey","aws_region","username","password","is_default","active").
		Select()
	return providers, err
}

func (impl DockerArtifactStoreRepositoryImpl) FindOne(storeId string) (*DockerArtifactStore, error) {
	var provider DockerArtifactStore
	err := impl.dbConnection.Model(&provider).
		Where("id = ?", storeId).
		Where("active = ?", true).
		//Column("id", "plugin_id","registry_url", "registry_type","aws_accesskey_id","aws_secret_accesskey","aws_region","username","password","is_default","active").
		Select()
	return &provider, err
}

func (impl DockerArtifactStoreRepositoryImpl) Update(artifactStore *DockerArtifactStore) error {
	//TODO check for unique default
	//there can be only one default

	if artifactStore.IsDefault == true {
		model, err := impl.FindActiveDefaultStore()
		if err == nil && model.Id != artifactStore.Id {
			model.IsDefault = false
			_ = impl.Update(model)
		}
	}
	return impl.dbConnection.Update(artifactStore)
}

func (impl DockerArtifactStoreRepositoryImpl) Delete(storeId string) error {

	artifactStore, err := impl.FindOne(storeId)
	if err != nil {
		return err
	}
	if artifactStore.IsDefault {
		return errors.New("default registry can't be delete")
	}
	return impl.dbConnection.Delete(artifactStore)
}

func (impl DockerArtifactStoreRepositoryImpl) MarkRegistryDeleted(deleteReq *DockerArtifactStore) error{
	if deleteReq.IsDefault {
		return errors.New("default registry can't be deleted")
	}
	deleteReq.Active = false
	return impl.dbConnection.Update(deleteReq)
}