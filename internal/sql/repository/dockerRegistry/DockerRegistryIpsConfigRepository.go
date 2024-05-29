/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package repository

import (
	"github.com/go-pg/pg"
)

type DockerRegistryIpsConfig struct {
	tableName             struct{}                        `sql:"docker_registry_ips_config" pg:",discard_unknown_columns"`
	Id                    int                             `sql:"id,pk"`
	DockerArtifactStoreId string                          `sql:"docker_artifact_store_id,notnull"`
	CredentialType        DockerRegistryIpsCredentialType `sql:"credential_type,notnull"`
	CredentialValue       string                          `sql:"credential_value"`
	AppliedClusterIdsCsv  string                          `sql:"applied_cluster_ids_csv"` // -1 means all_cluster
	IgnoredClusterIdsCsv  string                          `sql:"ignored_cluster_ids_csv"`
	Active                bool                            `sql:"active,notnull"`
}

type DockerRegistryIpsCredentialType string

type DockerRegistryIpsConfigRepository interface {
	Save(config *DockerRegistryIpsConfig, tx *pg.Tx) error
	Update(config *DockerRegistryIpsConfig, tx *pg.Tx) error
	FindInActiveByDockerRegistryId(dockerRegistryId string) (*DockerRegistryIpsConfig, error)
	FindByDockerRegistryId(dockerRegistryId string) (*DockerRegistryIpsConfig, error)
}

type DockerRegistryIpsConfigRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewDockerRegistryIpsConfigRepositoryImpl(dbConnection *pg.DB) *DockerRegistryIpsConfigRepositoryImpl {
	return &DockerRegistryIpsConfigRepositoryImpl{dbConnection: dbConnection}
}

func (impl DockerRegistryIpsConfigRepositoryImpl) Save(config *DockerRegistryIpsConfig, tx *pg.Tx) error {
	return tx.Insert(config)
}

func (impl DockerRegistryIpsConfigRepositoryImpl) Update(config *DockerRegistryIpsConfig, tx *pg.Tx) error {
	return tx.Update(config)
}

func (impl DockerRegistryIpsConfigRepositoryImpl) FindByDockerRegistryId(dockerRegistryId string) (*DockerRegistryIpsConfig, error) {
	var dockerRegistryIpsConfig DockerRegistryIpsConfig
	//added limit 1 for fasting querying
	err := impl.dbConnection.Model(&dockerRegistryIpsConfig).
		Where("docker_artifact_store_id = ?", dockerRegistryId).
		Where("active = ?", true).
		Limit(1).Select()
	return &dockerRegistryIpsConfig, err
}

func (impl DockerRegistryIpsConfigRepositoryImpl) FindInActiveByDockerRegistryId(dockerRegistryId string) (*DockerRegistryIpsConfig, error) {
	var dockerRegistryIpsConfig DockerRegistryIpsConfig
	//added limit 1 for fasting querying
	err := impl.dbConnection.Model(&dockerRegistryIpsConfig).
		Where("docker_artifact_store_id = ?", dockerRegistryId).
		Limit(1).Select()
	return &dockerRegistryIpsConfig, err
}
