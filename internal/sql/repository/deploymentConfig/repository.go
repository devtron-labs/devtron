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

package deploymentConfig

import (
	"fmt"

	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

type DeploymentAppType string

const (
	Argo DeploymentAppType = "argo_cd"
	Helm DeploymentAppType = "helm"
)

type ConfigType string

const (
	Custom          ConfigType = "custom"
	SystemGenerated ConfigType = "system_generated"
)

type DeploymentConfig struct {
	tableName         struct{} `sql:"deployment_config" pg:",discard_unknown_columns"`
	Id                int      `sql:"id,pk"`
	AppId             int      `sql:"app_id"`
	EnvironmentId     int      `sql:"environment_id"`
	DeploymentAppType string   `sql:"deployment_app_type"`
	ConfigType        string   `sql:"config_type"`
	RepoUrl           string   `sql:"repo_url"`
	RepoName          string   `sql:"repo_name"`
	ReleaseMode       string   `sql:"release_mode"`
	ReleaseConfig     string   `sql:"release_config"`
	Active            bool     `sql:"active,notnull"`
	sql.AuditLog
}

type Repository interface {
	Save(tx *pg.Tx, config *DeploymentConfig) (*DeploymentConfig, error)
	SaveAll(tx *pg.Tx, configs []*DeploymentConfig) ([]*DeploymentConfig, error)
	Update(tx *pg.Tx, config *DeploymentConfig) (*DeploymentConfig, error)
	UpdateAll(tx *pg.Tx, config []*DeploymentConfig) ([]*DeploymentConfig, error)
	GetByAppIdAndEnvId(tx *pg.Tx, appId, envId int) (*DeploymentConfig, error)
	GetAppLevelConfigForDevtronApps(tx *pg.Tx, appId int) (*DeploymentConfig, error)
	GetAppAndEnvLevelConfigsInBulk(appIdToEnvIdsMap map[int][]int) ([]*DeploymentConfig, error)
	GetByAppIdAndEnvIdEvenIfInactive(appId, envId int) (*DeploymentConfig, error)
	GetConfigByAppIds(appIds []int) ([]*DeploymentConfig, error)
	GetAllConfigsForActiveApps() ([]*DeploymentConfig, error)
	GetAllEnvLevelConfigsWithReleaseMode(releaseMode string) ([]*DeploymentConfig, error)
	GetDeploymentAppTypeForChartStoreAppByAppId(appId int) (string, error)
	// GitOps count methods
	GetGitOpsEnabledPipelineCount() (int, error)
}

type RepositoryImpl struct {
	dbConnection *pg.DB
}

func NewRepositoryImpl(dbConnection *pg.DB) *RepositoryImpl {
	return &RepositoryImpl{dbConnection: dbConnection}
}

func (impl *RepositoryImpl) Save(tx *pg.Tx, config *DeploymentConfig) (*DeploymentConfig, error) {
	var err error
	if tx != nil {
		err = tx.Insert(config)
	} else {
		err = impl.dbConnection.Insert(config)
	}
	return config, err
}

func (impl *RepositoryImpl) SaveAll(tx *pg.Tx, configs []*DeploymentConfig) ([]*DeploymentConfig, error) {
	var err error
	if tx != nil {
		err = tx.Insert(&configs)
	} else {
		err = impl.dbConnection.Insert(&configs)
	}
	return configs, err
}

func (impl *RepositoryImpl) Update(tx *pg.Tx, config *DeploymentConfig) (*DeploymentConfig, error) {
	var err error
	if tx != nil {
		err = tx.Update(config)
	} else {
		err = impl.dbConnection.Update(config)
	}
	return config, err
}

func (impl *RepositoryImpl) UpdateAll(tx *pg.Tx, configs []*DeploymentConfig) ([]*DeploymentConfig, error) {
	var err error
	for _, config := range configs {
		if tx != nil {
			_, err = tx.Model(config).WherePK().UpdateNotNull()
		} else {
			_, err = impl.dbConnection.Model(&config).UpdateNotNull()
		}
		if err != nil {
			return nil, err
		}
	}
	return configs, err
}

func (impl *RepositoryImpl) GetByAppIdAndEnvId(tx *pg.Tx, appId, envId int) (*DeploymentConfig, error) {
	result := &DeploymentConfig{}
	var connection orm.DB
	if tx != nil {
		connection = tx
	} else {
		connection = impl.dbConnection
	}
	err := connection.Model(result).
		Join("INNER JOIN app a").
		JoinOn("deployment_config.app_id = a.id").
		Join("INNER JOIN environment e").
		JoinOn("deployment_config.environment_id = e.id").
		Where("a.active = ?", true).
		Where("e.active = ?", true).
		Where("deployment_config.app_id = ?", appId).
		Where("deployment_config.environment_id = ?", envId).
		Where("deployment_config.active = ?", true).
		Order("deployment_config.id DESC").Limit(1).
		Select()
	return result, err
}

func (impl *RepositoryImpl) GetAppLevelConfigForDevtronApps(tx *pg.Tx, appId int) (*DeploymentConfig, error) {
	result := &DeploymentConfig{}
	var connection orm.DB
	if tx != nil {
		connection = tx
	} else {
		connection = impl.dbConnection
	}
	err := connection.Model(result).
		Join("INNER JOIN app a").
		JoinOn("deployment_config.app_id = a.id").
		Where("a.active = ?", true).
		Where("deployment_config.app_id = ? ", appId).
		Where("deployment_config.environment_id is NULL").
		Where("deployment_config.active = ?", true).
		Select()
	return result, err
}

func (impl *RepositoryImpl) GetAppAndEnvLevelConfigsInBulk(appIdToEnvIdsMap map[int][]int) ([]*DeploymentConfig, error) {
	var result []*DeploymentConfig
	if len(appIdToEnvIdsMap) == 0 {
		return result, nil
	}
	err := impl.dbConnection.Model(&result).
		Join("INNER JOIN app a").
		JoinOn("deployment_config.app_id = a.id").
		Join("INNER JOIN environment e").
		JoinOn("deployment_config.environment_id = e.id").
		Where("a.active = ?", true).
		Where("e.active = ?", true).
		WhereOrGroup(func(query *orm.Query) (*orm.Query, error) {
			for appId, envIds := range appIdToEnvIdsMap {
				if len(envIds) == 0 {
					continue
				}
				query = query.Where("deployment_config.app_id = ?", appId).
					Where("deployment_config.environment_id in (?)", pg.In(envIds)).
					Where("deployment_config.active = ?", true)
			}
			return query, nil
		}).Select()
	return result, err
}

func (impl *RepositoryImpl) GetByAppIdAndEnvIdEvenIfInactive(appId, envId int) (*DeploymentConfig, error) {
	if envId == 0 {
		return impl.getByAppIdEvenIfInactive(appId)
	}
	return impl.getByAppIdAndEnvIdEvenIfInactive(appId, envId)
}

func (impl *RepositoryImpl) getByAppIdEvenIfInactive(appId int) (*DeploymentConfig, error) {
	result := &DeploymentConfig{}
	err := impl.dbConnection.Model(result).
		Join("INNER JOIN app a").
		JoinOn("deployment_config.app_id = a.id").
		Where("a.active = ?", true).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			query = query.Where("deployment_config.app_id = ?", appId).
				Where("deployment_config.environment_id is NULL")
			return query, nil
		}).
		Order("deployment_config.id DESC").
		Limit(1).
		Select()
	return result, err
}

func (impl *RepositoryImpl) getByAppIdAndEnvIdEvenIfInactive(appId, envId int) (*DeploymentConfig, error) {
	if envId == 0 {
		return nil, fmt.Errorf("empty envId passed for deployment config")
	}
	result := &DeploymentConfig{}
	err := impl.dbConnection.Model(result).
		Join("INNER JOIN app a").
		JoinOn("deployment_config.app_id = a.id").
		Join("INNER JOIN environment e").
		JoinOn("deployment_config.environment_id = e.id").
		Where("a.active = ?", true).
		Where("e.active = ?", true).
		WhereGroup(func(query *orm.Query) (*orm.Query, error) {
			query = query.Where("deployment_config.app_id = ?", appId).
				Where("deployment_config.environment_id = ?", envId)
			return query, nil
		}).
		Order("deployment_config.id DESC").
		Limit(1).
		Select()
	return result, err
}

func (impl *RepositoryImpl) GetConfigByAppIds(appIds []int) ([]*DeploymentConfig, error) {
	var results []*DeploymentConfig
	if len(appIds) == 0 {
		return results, nil
	}
	err := impl.dbConnection.Model(&results).
		Join("INNER JOIN app a").
		JoinOn("deployment_config.app_id = a.id").
		Where("a.active = ?", true).
		Where("deployment_config.app_id in (?) ", pg.In(appIds)).
		Where("deployment_config.active = ?", true).
		Select()
	return results, err
}

// GetAllConfigsForActiveApps returns all deployment configs for active apps
// INNER JOIN app a is used to filter out inactive apps
// NOTE: earlier we were not deleting the deployment configs on app delete,
// so we need to filter out inactive deployment configs
func (impl *RepositoryImpl) GetAllConfigsForActiveApps() ([]*DeploymentConfig, error) {
	result := make([]*DeploymentConfig, 0)
	err := impl.dbConnection.Model(&result).
		Join("INNER JOIN app a").
		JoinOn("deployment_config.app_id = a.id").
		Where("a.active = ?", true).
		Where("deployment_config.active = ?", true).
		Select()
	return result, err
}

// GetAllEnvLevelConfigsWithReleaseMode returns all deployment configs for active apps and envs
// INNER JOIN app a is used to filter out inactive apps
// INNER JOIN environment e is used to filter out inactive envs
// NOTE: earlier we were not deleting the deployment configs on app delete,
// so we need to filter out inactive deployment configs
func (impl *RepositoryImpl) GetAllEnvLevelConfigsWithReleaseMode(releaseMode string) ([]*DeploymentConfig, error) {
	result := make([]*DeploymentConfig, 0)
	err := impl.dbConnection.Model(&result).
		Join("INNER JOIN app a").
		JoinOn("deployment_config.app_id = a.id").
		Join("INNER JOIN environment e").
		JoinOn("deployment_config.environment_id = e.id").
		Where("a.active = ?", true).
		Where("e.active = ?", true).
		Where("deployment_config.active = ?", true).
		Where("deployment_config.release_mode = ?", releaseMode).
		Select()
	return result, err
}

func (impl *RepositoryImpl) GetDeploymentAppTypeForChartStoreAppByAppId(appId int) (string, error) {
	result := &DeploymentConfig{}
	err := impl.dbConnection.Model(result).
		Join("inner join app a").
		JoinOn("deployment_config.app_id = a.id").
		Where("deployment_config.app_id = ? ", appId).
		Where("deployment_config.active = ?", true).
		Where("a.active = ?", true).
		Where("a.app_type = ? ", helper.ChartStoreApp).
		Select()
	return result.DeploymentAppType, err
}

// GetGitOpsEnabledPipelineCount returns count of GitOps enabled pipelines
// This handles lazy migration from pipeline table to deployment_config table
func (impl *RepositoryImpl) GetGitOpsEnabledPipelineCount() (int, error) {
	var count int
	// Complex query to handle lazy migration:
	// 1. Count pipelines that have deployment_config entry with argo_cd
	// 2. Count pipelines that don't have deployment_config entry but have argo_cd in pipeline table
	query := `
		SELECT COUNT(DISTINCT p.id)
		FROM pipeline p
		JOIN environment e ON p.environment_id = e.id
		JOIN app a ON p.app_id = a.id
		LEFT JOIN deployment_config dc ON dc.app_id = p.app_id
			AND dc.environment_id = p.environment_id
			AND dc.active = true
		WHERE p.deleted = false
			AND e.active = true
			AND a.active = true
			AND (
				-- Case 1: deployment_config exists and is argo_cd
				dc.deployment_app_type = 'argo_cd'
				OR
				-- Case 2: no deployment_config entry, fallback to pipeline table
				(dc.id IS NULL AND p.deployment_app_type = 'argo_cd')
			)
	`

	_, err := impl.dbConnection.Query(&count, query)
	if err != nil {
		return 0, fmt.Errorf("error getting GitOps enabled pipeline count: %w", err)
	}
	return count, nil
}
