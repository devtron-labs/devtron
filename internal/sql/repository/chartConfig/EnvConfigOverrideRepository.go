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

package chartConfig

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/go-pg/pg"
	"github.com/juju/errors"
)

type EnvConfigOverride struct {
	tableName         struct{}           `sql:"chart_env_config_override" pg:",discard_unknown_columns"`
	Id                int                `sql:"id,pk"`
	ChartId           int                `sql:"chart_id,notnull"`
	TargetEnvironment int                `sql:"target_environment,notnull"` //target environment
	EnvOverrideValues string             `sql:"env_override_yaml,notnull"`
	Status            models.ChartStatus `sql:"status,notnull"` //new, deployment-in-progress, error, rollbacked, su
	ManualReviewed    bool               `sql:"reviewed,notnull"`
	Active            bool               `sql:"active,notnull"`
	Namespace         string             `sql:"namespace,notnull"`
	Chart             *Chart
	Environment       *cluster.Environment `sql:"-"`
	Latest            bool                 `sql:"latest,notnull"`
	Previous          bool                 `sql:"previous,notnull"`
	IsOverride        bool                 `sql:"is_override,notnull"`
	models.AuditLog
}

type EnvConfigOverrideRepository interface {
	Save(*EnvConfigOverride) error
	GetByChartAndEnvironment(chartId, targetEnvironmentId int) (*EnvConfigOverride, error)
	ActiveEnvConfigOverride(appId, environmentId int) (*EnvConfigOverride, error) //successful env config
	Get(id int) (*EnvConfigOverride, error)
	//this api updates only EnvOverrideValues, EnvMergedValues, Status, ManualReviewed, active based on id
	UpdateProperties(config *EnvConfigOverride) error
	GetByEnvironment(targetEnvironmentId int) ([]EnvConfigOverride, error)

	GetEnvConfigByChartId(chartId int) ([]EnvConfigOverride, error)
	UpdateEnvConfigStatus(config *EnvConfigOverride) error
	Delete(envConfigOverride *EnvConfigOverride) error
	FindLatestChartForAppByAppIdAndEnvId(appId, targetEnvironmentId int) (*EnvConfigOverride, error)
	FindChartByAppIdAndEnvIdAndChartRefId(appId, targetEnvironmentId int, chartRefId int) (*EnvConfigOverride, error)
	Update(envConfigOverride *EnvConfigOverride) (*EnvConfigOverride, error)
	FindChartForAppByAppIdAndEnvId(appId, targetEnvironmentId int) (*EnvConfigOverride, error)
}

type EnvConfigOverrideRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewEnvConfigOverrideRepository(dbConnection *pg.DB) *EnvConfigOverrideRepositoryImpl {
	return &EnvConfigOverrideRepositoryImpl{dbConnection: dbConnection}
}
func (r EnvConfigOverrideRepositoryImpl) Save(override *EnvConfigOverride) error {
	err := r.dbConnection.Insert(override)
	return err
}

func (r EnvConfigOverrideRepositoryImpl) ActiveEnvConfigOverride(appId, environmentId int) (*EnvConfigOverride, error) {
	var environmentConfig struct {
		Id                int                `sql:"id,pk"`
		ChartId           int                `sql:"chart_id,notnull"`
		TargetEnvironment int                `sql:"target_environment,notnull"` //target environment
		EnvOverrideValues string             `sql:"env_override_yaml,notnull"`
		Status            models.ChartStatus `sql:"status,notnull"` //new, deployment-in-progress, error, rollbacked, su
		ManualReviewed    bool               `sql:"reviewed,notnull"`
		Active            bool               `sql:"active,notnull"`
		Namespace         string             `sql:"namespace"`

		ChartName               string `sql:"chart_name"`
		ChartLocation           string `sql:"chart_location"`  //location within git repo where current chart is pointing
		GlobalOverride          string `sql:"global_override"` //json format
		ImageDescriptorTemplate string `sql:"image_descriptor_template"`
		EnvironmentName         string `sql:"environment_name"`
		Latest                  bool   `sql:"latest,notnull"`
		AppName                 string `sql:"app_name"`
		IsOverride              bool   `sql:"is_override"`
		ChartRefId              int    `sql:"chart_ref_id,notnull"`
		ChartVersion            string `sql:"chart_version,notnull"`
	}

	query := "SELECT " +
		" ec.id as id, ec.chart_id as chart_id," +
		" ec.target_environment as target_environment, ec.env_override_yaml as env_override_yaml, ec.status as status, ec.reviewed as reviewed," +
		" ec.active as active, ec.namespace as namespace, ec.latest as latest," +

		" ch.chart_name as chart_name," +
		" ch.chart_location as chart_location," +
		" ch.global_override as global_override, ch.chart_version as chart_version," +
		" ch.image_descriptor_template as image_descriptor_template," +
		" en.environment_name as environment_name, ec.is_override, ch.chart_ref_id" +

		" FROM chart_env_config_override ec" +
		" LEFT JOIN charts ch on ec.chart_id=ch.id" +
		" LEFT JOIN environment en on en.id=ec.target_environment" +
		" WHERE ec.target_environment=? and ec.active = ? and ch.app_id =? and ec.latest = ?;"

	_, err := r.dbConnection.Query(&environmentConfig, query, environmentId, true, appId, true)
	if err != nil {
		return nil, err
	}

	chart := &Chart{
		ChartName:               environmentConfig.ChartName,
		ChartLocation:           environmentConfig.ChartLocation,
		GlobalOverride:          environmentConfig.GlobalOverride,
		ImageDescriptorTemplate: environmentConfig.ImageDescriptorTemplate,
		ChartRefId:              environmentConfig.ChartRefId,
		ChartVersion:            environmentConfig.ChartVersion,
	}
	env := &cluster.Environment{
		Name: environmentConfig.EnvironmentName,
	}
	eco := &EnvConfigOverride{
		Id:                environmentConfig.Id,
		ChartId:           environmentConfig.ChartId,
		TargetEnvironment: environmentConfig.TargetEnvironment,
		EnvOverrideValues: environmentConfig.EnvOverrideValues,
		Status:            environmentConfig.Status,
		ManualReviewed:    environmentConfig.ManualReviewed,
		Active:            environmentConfig.Active,
		Namespace:         environmentConfig.Namespace,
		Chart:             chart,
		Environment:       env,
		Latest:            environmentConfig.Latest,
		IsOverride:        environmentConfig.IsOverride,
		//AppMetricsOverride: environmentConfig.AppMetricsOverride,
	}
	return eco, err
}

func (r EnvConfigOverrideRepositoryImpl) GetByChartAndEnvironment(chartId, targetEnvironmentId int) (*EnvConfigOverride, error) {
	eco := &EnvConfigOverride{}
	err := r.dbConnection.
		Model(eco).
		Where("env_config_override.target_environment = ?", targetEnvironmentId).
		Where("env_config_override.active = ?", true).
		Where("Chart.id =? ", chartId).
		Column("env_config_override.*", "Chart").
		Select()
	if pg.ErrNoRows == err {
		return nil, errors.NotFoundf(err.Error())
	}
	return eco, err
}

func (r EnvConfigOverrideRepositoryImpl) Get(id int) (*EnvConfigOverride, error) {
	eco := &EnvConfigOverride{}
	err := r.dbConnection.
		Model(eco).
		Where("env_config_override.id = ?", id).
		Column("env_config_override.*", "Chart").
		Select()
	return eco, err
}

//this api updates only EnvOverrideValues, EnvMergedValues, Status, ManualReviewed, active
// based on id
func (r EnvConfigOverrideRepositoryImpl) UpdateProperties(config *EnvConfigOverride) error {
	_, err := r.dbConnection.Model(config).
		Set("env_override_yaml = ?", config.EnvOverrideValues).
		Set("status =?", config.Status).
		Set("reviewed =?", config.ManualReviewed).
		Set("active =?", config.Active).
		Set("updated_by =?", config.UpdatedBy).
		Set("updated_on =? ", config.UpdatedOn).
		Set("previous =?", config.Previous).
		Set("is_override =?", config.IsOverride).
		Set("namespace =?", config.Namespace).
		Set("latest =?", config.Latest).
		//Set("app_metrics_override =?", config.AppMetricsOverride).
		WherePK().
		Update()
	return err
}

func (r EnvConfigOverrideRepositoryImpl) GetByEnvironment(targetEnvironmentId int) ([]EnvConfigOverride, error) {
	var envConfigs []EnvConfigOverride
	err := r.dbConnection.
		Model(&envConfigs).
		Where("env_config_override.target_environment = ?", targetEnvironmentId).
		Where("env_config_override.active = ?", true).
		Column("env_config_override.*").
		Select()
	if pg.ErrNoRows == err {
		return nil, errors.NotFoundf(err.Error())
	}
	return envConfigs, err
}

func (r EnvConfigOverrideRepositoryImpl) GetEnvConfigByChartId(chartId int) ([]EnvConfigOverride, error) {
	var envConfigs []EnvConfigOverride
	err := r.dbConnection.
		Model(&envConfigs).
		Where("chart_id = ?", chartId).
		Where("active = ?", true).
		Select()
	if pg.ErrNoRows == err {
		return nil, errors.NotFoundf(err.Error())
	}
	return envConfigs, err
}

func (r EnvConfigOverrideRepositoryImpl) UpdateEnvConfigStatus(config *EnvConfigOverride) error {
	_, err := r.dbConnection.Model(config).
		Set("latest =?", config.Latest).
		Set("status =?", config.Status).
		Set("reviewed =?", config.ManualReviewed).
		Set("active =?", config.Active).
		Set("updated_by =?", config.UpdatedBy).
		Set("updated_on =? ", config.UpdatedOn).
		Set("previous =?", config.Previous).
		WherePK().
		Update()
	return err
}

func (r EnvConfigOverrideRepositoryImpl) Delete(envConfigOverride *EnvConfigOverride) error {
	err := r.dbConnection.Delete(envConfigOverride)
	return err
}

func (r EnvConfigOverrideRepositoryImpl) FindLatestChartForAppByAppIdAndEnvId(appId, targetEnvironmentId int) (*EnvConfigOverride, error) {
	eco := &EnvConfigOverride{}
	err := r.dbConnection.
		Model(eco).
		Where("env_config_override.target_environment = ?", targetEnvironmentId).
		Where("env_config_override.latest = ?", true).
		Where("Chart.app_id =? ", appId).
		Column("env_config_override.*", "Chart").
		Select()
	if pg.ErrNoRows == err {
		return nil, errors.NotFoundf(err.Error())
	}
	return eco, err
}

func (r EnvConfigOverrideRepositoryImpl) FindChartByAppIdAndEnvIdAndChartRefId(appId, targetEnvironmentId int, chartRefId int) (*EnvConfigOverride, error) {
	eco := &EnvConfigOverride{}
	err := r.dbConnection.
		Model(eco).
		Where("env_config_override.target_environment = ?", targetEnvironmentId).
		//Where("env_config_override.latest = ?", true).
		Where("Chart.app_id =? ", appId).
		Where("Chart.chart_ref_id =? ", chartRefId).
		Column("env_config_override.*", "Chart").
		Select()
	if pg.ErrNoRows == err {
		return nil, errors.NotFoundf(err.Error())
	}
	return eco, err
}

func (r EnvConfigOverrideRepositoryImpl) Update(envConfigOverride *EnvConfigOverride) (*EnvConfigOverride, error) {
	err := r.dbConnection.Update(envConfigOverride)
	return envConfigOverride, err
}

func (r EnvConfigOverrideRepositoryImpl) FindChartForAppByAppIdAndEnvId(appId, targetEnvironmentId int) (*EnvConfigOverride, error) {
	eco := &EnvConfigOverride{}
	err := r.dbConnection.
		Model(eco).
		Where("env_config_override.target_environment = ?", targetEnvironmentId).
		Where("env_config_override.active = ?", true).
		Where("Chart.app_id =? ", appId).
		Column("env_config_override.*", "Chart").
		Select()
	if pg.ErrNoRows == err {
		return nil, errors.NotFoundf(err.Error())
	}
	return eco, err
}
