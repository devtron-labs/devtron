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
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg/orm"
	"net/url"

	"github.com/go-pg/pg"
	"github.com/pkg/errors"
)

const (
	REGISTRYTYPE_ECR                         = "ecr"
	REGISTRYTYPE_GCR                         = "gcr"
	REGISTRYTYPE_ARTIFACT_REGISTRY           = "artifact-registry"
	REGISTRYTYPE_OTHER                       = "other"
	REGISTRYTYPE_DOCKER_HUB                  = "docker-hub"
	JSON_KEY_USERNAME                 string = "_json_key"
	STORAGE_ACTION_TYPE_PULL                 = "PULL"
	STORAGE_ACTION_TYPE_PUSH                 = "PUSH"
	STORAGE_ACTION_TYPE_PULL_AND_PUSH        = "PULL/PUSH"
	OCI_REGISRTY_REPO_TYPE_CONTAINER         = "CONTAINER"
	OCI_REGISRTY_REPO_TYPE_CHART             = "CHART"
)

type RegistryType string

var OCI_REGISRTY_REPO_TYPE_LIST = []string{OCI_REGISRTY_REPO_TYPE_CONTAINER, OCI_REGISRTY_REPO_TYPE_CHART}

type DockerArtifactStore struct {
	tableName              struct{}     `sql:"docker_artifact_store" json:",omitempty"  pg:",discard_unknown_columns"`
	Id                     string       `sql:"id,pk" json:"id,,omitempty"`
	PluginId               string       `sql:"plugin_id,notnull" json:"pluginId,omitempty"`
	RegistryURL            string       `sql:"registry_url" json:"registryUrl,omitempty"`
	RegistryType           RegistryType `sql:"registry_type,notnull" json:"registryType,omitempty"`
	IsOCICompliantRegistry bool         `sql:"is_oci_compliant_registry,notnull" json:"isOCICompliantRegistry,omitempty"`
	AWSAccessKeyId         string       `sql:"aws_accesskey_id" json:"awsAccessKeyId,omitempty" `
	AWSSecretAccessKey     string       `sql:"aws_secret_accesskey" json:"awsSecretAccessKey,omitempty"`
	AWSRegion              string       `sql:"aws_region" json:"awsRegion,omitempty"`
	Username               string       `sql:"username" json:"username,omitempty"`
	Password               string       `sql:"password" json:"password,omitempty"`
	IsDefault              bool         `sql:"is_default,notnull" json:"isDefault"`
	Connection             string       `sql:"connection" json:"connection,omitempty"`
	Cert                   string       `sql:"cert" json:"cert,omitempty"`
	Active                 bool         `sql:"active,notnull" json:"active"`
	IpsConfig              *DockerRegistryIpsConfig
	OCIRegistryConfig      []*OCIRegistryConfig
	sql.AuditLog
}

type DockerArtifactStoreExt struct {
	*DockerArtifactStore
	DeploymentCount int `sql:"deployment_count" json:"deploymentCount"`
}

type ChartDeploymentCount struct {
	OCIChartName    string `sql:"oci_chart_name" json:"ociChartName"`
	DeploymentCount int    `sql:"deployment_count" json:"deploymentCount"`
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
	GetConnection() *pg.DB
	Save(artifactStore *DockerArtifactStore, tx *pg.Tx) error
	FindActiveDefaultStore() (*DockerArtifactStore, error)
	FindAllActiveForAutocomplete() ([]DockerArtifactStore, error)
	FindAll() ([]DockerArtifactStore, error)
	FindAllChartProviders() ([]DockerArtifactStore, error)
	FindOne(storeId string) (*DockerArtifactStore, error)
	FindOneWithDeploymentCount(storeId string) (*DockerArtifactStoreExt, error)
	FindOneWithChartDeploymentCount(storeId, chartName string) (*ChartDeploymentCount, error)
	FindOneInactive(storeId string) (*DockerArtifactStore, error)
	Update(artifactStore *DockerArtifactStore, tx *pg.Tx) error
	Delete(storeId string) error
	MarkRegistryDeleted(artifactStore *DockerArtifactStore, tx *pg.Tx) error
	FindInactive(storeId string) (bool, error)
}
type DockerArtifactStoreRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewDockerArtifactStoreRepositoryImpl(dbConnection *pg.DB) *DockerArtifactStoreRepositoryImpl {
	return &DockerArtifactStoreRepositoryImpl{dbConnection: dbConnection}
}

func (impl DockerArtifactStoreRepositoryImpl) GetConnection() *pg.DB {
	return impl.dbConnection
}

func (impl DockerArtifactStoreRepositoryImpl) Save(artifactStore *DockerArtifactStore, tx *pg.Tx) error {
	if util.IsBaseStack() {
		return tx.Insert(artifactStore)
	}

	//TODO check for unique default
	//there can be only one default
	model, err := impl.FindActiveDefaultStore()
	if err == pg.ErrNoRows {
		artifactStore.IsDefault = true
	} else if err == nil && model.Id != artifactStore.Id && artifactStore.IsDefault == true {
		model.IsDefault = false
		err = impl.Update(model, tx)
		if err != nil {
			return err
		}
	}
	return tx.Insert(artifactStore)
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
		Column("docker_artifact_store.id", "registry_url", "registry_type", "is_default", "is_oci_compliant_registry", "OCIRegistryConfig").
		Where("active = ?", true).
		Relation("OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE"), nil
		}).
		Select()

	return providers, err
}

func (impl DockerArtifactStoreRepositoryImpl) FindAll() ([]DockerArtifactStore, error) {
	var providers []DockerArtifactStore
	err := impl.dbConnection.Model(&providers).
		Column("docker_artifact_store.*", "IpsConfig", "OCIRegistryConfig").
		Where("docker_artifact_store.active = ?", true).
		Relation("OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE"), nil
		}).
		Relation("IpsConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.JoinOn("(ips_config.active=true or ips_config is null)"), nil
		}).
		Select()
	return providers, err
}

func (impl DockerArtifactStoreRepositoryImpl) FindAllChartProviders() ([]DockerArtifactStore, error) {
	var providers []DockerArtifactStore
	err := impl.dbConnection.Model(&providers).
		Column("docker_artifact_store.*", "OCIRegistryConfig").
		Where("docker_artifact_store.active = ?", true).
		Relation("OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE and " +
				"repository_type='CHART' and " +
				"(repository_action='PULL' or repository_action='PULL/PUSH')"), nil
		}).
		Select()
	return providers, err
}

func (impl DockerArtifactStoreRepositoryImpl) FindOne(storeId string) (*DockerArtifactStore, error) {
	var provider DockerArtifactStore
	err := impl.dbConnection.Model(&provider).
		Column("docker_artifact_store.*", "IpsConfig", "OCIRegistryConfig").
		Where("docker_artifact_store.id = ?", storeId).
		Where("docker_artifact_store.active = ?", true).
		Relation("OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE"), nil
		}).
		Relation("IpsConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.JoinOn("(ips_config.active=true or ips_config is null)"), nil
		}).
		Select()
	return &provider, err
}

func (impl DockerArtifactStoreRepositoryImpl) FindOneWithDeploymentCount(storeId string) (*DockerArtifactStoreExt, error) {
	var provider DockerArtifactStoreExt
	query := "SELECT docker_artifact_store.*, count(jq.ia_id) as deployment_count FROM docker_artifact_store" +
		fmt.Sprintf(" LEFT JOIN oci_registry_config orc on (docker_artifact_store.id = orc.docker_artifact_store_id and orc.is_chart_pull_active = true and orc.deleted = false and orc.repository_type = '%s' and (orc.repository_action = '%s' or orc.repository_action = '%s'))", OCI_REGISRTY_REPO_TYPE_CHART, STORAGE_ACTION_TYPE_PULL, STORAGE_ACTION_TYPE_PULL_AND_PUSH) +
		" LEFT JOIN (SELECT aps.docker_artifact_store_id as das_id ,ia.id as ia_id FROM installed_app_versions iav INNER JOIN installed_apps ia on iav.installed_app_id = ia.id INNER JOIN app_store_application_version asav on iav.app_store_application_version_id = asav.id INNER JOIN app_store aps on asav.app_store_id = aps.id WHERE ia.active=true and iav.active=true) jq on jq.das_id = docker_artifact_store.id" +
		" WHERE docker_artifact_store.id = ? and docker_artifact_store.active = true Group by docker_artifact_store.id;"
	_, err := impl.dbConnection.Query(&provider, query, storeId)
	return &provider, err
}

func (impl DockerArtifactStoreRepositoryImpl) FindOneWithChartDeploymentCount(storeId, chartName string) (*ChartDeploymentCount, error) {
	var provider ChartDeploymentCount
	query := "SELECT aps.name as oci_chart_name, COUNT(ia.id) as deployment_count FROM installed_app_versions iav" +
		" INNER JOIN installed_apps ia on iav.installed_app_id = ia.id" +
		" INNER JOIN app_store_application_version asav on iav.app_store_application_version_id = asav.id" +
		" INNER JOIN app_store aps on asav.app_store_id = aps.id" +
		" WHERE ia.active=true and iav.active=true and aps.docker_artifact_store_id = ? and aps.name = ?" +
		" GROUP BY oci_chart_name,aps.docker_artifact_store_id;"
	_, err := impl.dbConnection.Query(&provider, query, storeId, chartName)
	return &provider, err
}

func (impl DockerArtifactStoreRepositoryImpl) FindOneInactive(storeId string) (*DockerArtifactStore, error) {
	var provider DockerArtifactStore
	err := impl.dbConnection.Model(&provider).
		Column("docker_artifact_store.*", "IpsConfig", "OCIRegistryConfig").
		Where("docker_artifact_store.id = ?", storeId).
		Where("docker_artifact_store.active = ?", false).
		Relation("OCIRegistryConfig", func(q *orm.Query) (query *orm.Query, err error) {
			return q.Where("deleted IS FALSE"), nil
		}).
		Select()
	return &provider, err
}

func (impl DockerArtifactStoreRepositoryImpl) Update(artifactStore *DockerArtifactStore, tx *pg.Tx) error {
	//TODO check for unique default
	//there can be only one default

	if artifactStore.IsDefault == true {
		model, err := impl.FindActiveDefaultStore()
		if err == nil && model.Id != artifactStore.Id {
			model.IsDefault = false
			_ = impl.Update(model, tx)
		}
	}
	return tx.Update(artifactStore)
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

func (impl DockerArtifactStoreRepositoryImpl) MarkRegistryDeleted(deleteReq *DockerArtifactStore, tx *pg.Tx) error {
	if deleteReq.IsDefault {
		return errors.New("default registry can't be deleted")
	}
	deleteReq.Active = false
	return tx.Update(deleteReq)
}

func (impl DockerArtifactStoreRepositoryImpl) FindInactive(storeId string) (bool, error) {
	var provider DockerArtifactStore
	exist, err := impl.dbConnection.Model(&provider).
		Where("docker_artifact_store.id = ?", storeId).
		Where("active = ?", false).
		Exists()
	return exist, err
}
