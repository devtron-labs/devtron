/*
 * Copyright (c) 2024. Devtron Inc.
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

package bulkUpdate

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
)

// DEPRECATED: Use BulkEditConfig instead of BulkUpdateReadme.
// TODO: Remove this table in future versions.
type BulkUpdateReadme struct {
	tableName struct{} `sql:"bulk_update_readme" pg:",discard_unknown_columns"`
	Id        int      `sql:"id"`
	Resource  string   `sql:"resource"`
	Script    string   `sql:"script"`
	Readme    string   `sql:"readme"`
}

// BulkEditConfig is used to store the configuration for bulk edit operations.
type BulkEditConfig struct {
	tableName  struct{} `sql:"bulk_edit_config" pg:",discard_unknown_columns"`
	Id         int      `sql:"id,pk"`
	ApiVersion string   `sql:"api_version,notnull"`
	Kind       string   `sql:"kind,notnull"`
	Readme     string   `sql:"readme"`
	Schema     string   `sql:"schema"`
}

type BulkUpdateRepository interface {
	FindBulkEditConfig(apiVersion, kind string) (*BulkEditConfig, error)

	// methods for Deployment Template :

	FindDeploymentTemplateBulkAppNameForGlobal(appNameIncludes []string, appNameExcludes []string) ([]*app.App, error)
	FindDeploymentTemplateBulkAppNameForEnv(appNameIncludes []string, appNameExcludes []string, envId int) ([]*app.App, error)
	FindAppByChartId(chartId int) (*app.App, error)
	FindAppByChartEnvId(chartEnvId int) (*app.App, error)
	FindBulkChartsByAppNameSubstring(appNameIncludes []string, appNameExcludes []string) ([]*chartRepoRepository.Chart, error)
	FindBulkChartsEnvByAppNameSubstring(appNameIncludes []string, appNameExcludes []string, envId int) ([]*chartConfig.EnvConfigOverride, error)
	BulkUpdateChartsValuesYamlAndGlobalOverrideById(id int, patchValuesYml string, patchGlobalOverrideYml string) error
	BulkUpdateChartsEnvYamlOverrideById(id int, patch string) error

	// methods for ConfigMap & Secret :

	FindCMBulkAppModelForGlobal(appNameIncludes []string, appNameExcludes []string, configMapNames []string) ([]*chartConfig.ConfigMapAppModel, error)
	FindSecretBulkAppModelForGlobal(appNameIncludes []string, appNameExcludes []string, secretNames []string) ([]*chartConfig.ConfigMapAppModel, error)
	FindCMBulkAppModelForEnv(appNameIncludes []string, appNameExcludes []string, envId int, configMapNames []string) ([]*chartConfig.ConfigMapEnvModel, error)
	FindSecretBulkAppModelForEnv(appNameIncludes []string, appNameExcludes []string, envId int, secretNames []string) ([]*chartConfig.ConfigMapEnvModel, error)
	BulkUpdateConfigMapDataForGlobalById(id int, patch string) error
	BulkUpdateSecretDataForGlobalById(id int, patch string) error
	BulkUpdateConfigMapDataForEnvById(id int, patch string) error
	BulkUpdateSecretDataForEnvById(id int, patch string) error
}

func NewBulkUpdateRepository(dbConnection *pg.DB,
	logger *zap.SugaredLogger) *BulkUpdateRepositoryImpl {
	return &BulkUpdateRepositoryImpl{dbConnection: dbConnection,
		logger: logger}
}

type BulkUpdateRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func appendBuildAppNameQuery(q *orm.Query, appNameIncludes []string, appNameExcludes []string) *orm.Query {
	if len(appNameIncludes) != 0 {
		q = q.Where("app.app_name LIKE ANY (array[?])", pg.In(appNameIncludes))
	}
	if len(appNameExcludes) != 0 {
		q = q.Where("app.app_name NOT LIKE ALL (array[?])", pg.In(appNameExcludes))
	}
	return q

}

func appendBuildCMNameQuery(q *orm.Query, configMapNames []string) *orm.Query {
	if len(configMapNames) == 0 {
		return q
	}
	//replacing configMapName with "%configMapName%"
	configMapNamesLikeClause := make([]string, len(configMapNames))
	for i := range configMapNames {
		configMapNamesLikeClause[i] = util.GetLIKEClauseQueryParam(configMapNames[i])
	}
	return q.Where("config_map_data LIKE ANY (array[?])", pg.In(configMapNamesLikeClause))
}

func appendBuildSecretNameQuery(q *orm.Query, secretNames []string) *orm.Query {
	if len(secretNames) == 0 {
		return q
	}
	//replacing secretName with "%secretName%"
	secretNamesLikeClause := make([]string, len(secretNames))
	for i := range secretNames {
		secretNamesLikeClause[i] = util.GetLIKEClauseQueryParam(secretNames[i])
	}
	return q.Where("secret_data LIKE ANY (array[?])", pg.In(secretNamesLikeClause))
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindBulkEditConfig(apiVersion, kind string) (*BulkEditConfig, error) {
	bulkEditConfig := &BulkEditConfig{}
	err := repositoryImpl.dbConnection.
		Model(bulkEditConfig).
		Where("api_version = ?", apiVersion).
		Where("kind = ?", kind).
		Select()
	return bulkEditConfig, err
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindDeploymentTemplateBulkAppNameForGlobal(appNameIncludes []string, appNameExcludes []string) ([]*app.App, error) {
	apps := []*app.App{}
	q := repositoryImpl.dbConnection.
		Model(&apps).Join("INNER JOIN charts ch ON app.id = ch.app_id").
		Where("app.active = ?", true).
		Where("ch.latest = ?", true)
	q = appendBuildAppNameQuery(q, appNameIncludes, appNameExcludes)
	err := q.Select()
	return apps, err
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindDeploymentTemplateBulkAppNameForEnv(appNameIncludes []string, appNameExcludes []string, envId int) ([]*app.App, error) {
	apps := []*app.App{}
	q := repositoryImpl.dbConnection.
		Model(&apps).Join("INNER JOIN charts ch ON app.id = ch.app_id").
		Join("INNER JOIN chart_env_config_override ON ch.id = chart_env_config_override.chart_id").
		Where("app.active = ?", true).
		Where("chart_env_config_override.target_environment = ? ", envId).
		Where("chart_env_config_override.latest = ?", true)
	q = appendBuildAppNameQuery(q, appNameIncludes, appNameExcludes)
	err := q.Select()
	return apps, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) FindCMBulkAppModelForGlobal(appNameIncludes []string, appNameExcludes []string, configMapNames []string) ([]*chartConfig.ConfigMapAppModel, error) {
	CmAndSecretAppModel := []*chartConfig.ConfigMapAppModel{}
	q := repositoryImpl.dbConnection.
		Model(&CmAndSecretAppModel).Join("INNER JOIN app ON app.id = config_map_app_model.app_id").
		Where("app.active = ?", true)
	q = appendBuildAppNameQuery(q, appNameIncludes, appNameExcludes)
	q = appendBuildCMNameQuery(q, configMapNames)
	err := q.Select()
	return CmAndSecretAppModel, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) FindSecretBulkAppModelForGlobal(appNameIncludes []string, appNameExcludes []string, secretNames []string) ([]*chartConfig.ConfigMapAppModel, error) {
	CmAndSecretAppModel := []*chartConfig.ConfigMapAppModel{}
	q := repositoryImpl.dbConnection.
		Model(&CmAndSecretAppModel).Join("INNER JOIN app ON app.id = config_map_app_model.app_id").
		Where("app.active = ?", true)
	q = appendBuildAppNameQuery(q, appNameIncludes, appNameExcludes)
	q = appendBuildSecretNameQuery(q, secretNames)
	err := q.Select()
	return CmAndSecretAppModel, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) FindCMBulkAppModelForEnv(appNameIncludes []string, appNameExcludes []string, envId int, configMapNames []string) ([]*chartConfig.ConfigMapEnvModel, error) {
	CmAndSecretEnvModel := []*chartConfig.ConfigMapEnvModel{}
	q := repositoryImpl.dbConnection.
		Model(&CmAndSecretEnvModel).Join("INNER JOIN app ON app.id = config_map_env_model.app_id").
		Where("app.active = ?", true).
		Where("config_map_env_model.environment_id = ? ", envId)
	q = appendBuildAppNameQuery(q, appNameIncludes, appNameExcludes)
	q = appendBuildCMNameQuery(q, configMapNames)
	err := q.Select()
	return CmAndSecretEnvModel, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) FindSecretBulkAppModelForEnv(appNameIncludes []string, appNameExcludes []string, envId int, secretNames []string) ([]*chartConfig.ConfigMapEnvModel, error) {
	CmAndSecretEnvModel := []*chartConfig.ConfigMapEnvModel{}
	q := repositoryImpl.dbConnection.
		Model(&CmAndSecretEnvModel).Join("INNER JOIN app ON app.id = config_map_env_model.app_id").
		Where("app.active = ?", true).
		Where("config_map_env_model.environment_id = ? ", envId)
	q = appendBuildAppNameQuery(q, appNameIncludes, appNameExcludes)
	q = appendBuildSecretNameQuery(q, secretNames)
	err := q.Select()
	return CmAndSecretEnvModel, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) FindAppByChartId(chartId int) (*app.App, error) {
	app := &app.App{}
	err := repositoryImpl.dbConnection.
		Model(app).Join("INNER JOIN charts ch ON app.id = ch.app_id").
		Where("ch.id = ?", chartId).
		Where("app.active = ?", true).
		Where("ch.latest = ?", true).
		Select()
	return app, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) FindAppByChartEnvId(chartEnvId int) (*app.App, error) {
	app := &app.App{}
	err := repositoryImpl.dbConnection.
		Model(app).Join("INNER JOIN charts ch ON app.id = ch.app_id").
		Join("INNER JOIN chart_env_config_override ON ch.id = chart_env_config_override.chart_id").
		Where("app.active = ?", true).
		Where("chart_env_config_override.id = ? ", chartEnvId).
		Where("chart_env_config_override.latest = ?", true).
		Select()
	return app, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) FindBulkChartsByAppNameSubstring(appNameIncludes []string, appNameExcludes []string) ([]*chartRepoRepository.Chart, error) {
	charts := []*chartRepoRepository.Chart{}
	q := repositoryImpl.dbConnection.
		Model(&charts).Join("INNER JOIN app ON app.id=app_id ").
		Where("app.active = ?", true).
		Where("latest = ?", true)
	q = appendBuildAppNameQuery(q, appNameIncludes, appNameExcludes)
	err := q.Select()
	return charts, err
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindBulkChartsEnvByAppNameSubstring(appNameIncludes []string, appNameExcludes []string, envId int) ([]*chartConfig.EnvConfigOverride, error) {
	charts := []*chartConfig.EnvConfigOverride{}
	q := repositoryImpl.dbConnection.
		Model(&charts).Join("INNER JOIN charts ch ON ch.id=env_config_override.chart_id").
		Join("INNER JOIN app ON app.id=ch.app_id").
		Where("app.active = ?", true).
		Where("env_config_override.target_environment = ?", envId).
		Where("env_config_override.latest = ?", true).
		Column("env_config_override.*", "Chart")
	q = appendBuildAppNameQuery(q, appNameIncludes, appNameExcludes)
	err := q.Select()
	return charts, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) BulkUpdateChartsValuesYamlAndGlobalOverrideById(id int, patchValuesYml string, patchGlobalOverrideYml string) error {
	chart := &chartRepoRepository.Chart{}
	_, err := repositoryImpl.dbConnection.
		Model(chart).
		Set("values_yaml = ?", patchValuesYml).
		Set("global_override = ?", patchGlobalOverrideYml).
		Where("id = ?", id).
		Update()
	if err != nil {
		repositoryImpl.logger.Errorw("error in bulk updating deployment template for charts", "err", err)
		return err
	}
	return nil
}
func (repositoryImpl BulkUpdateRepositoryImpl) BulkUpdateChartsEnvYamlOverrideById(id int, patch string) error {
	chartEnv := &chartConfig.EnvConfigOverride{}
	_, err := repositoryImpl.dbConnection.
		Model(chartEnv).
		Set("env_override_yaml = ?", patch).
		Where("id = ?", id).
		Update()
	if err != nil {
		repositoryImpl.logger.Errorw("error in bulk updating deployment template charts(for env)", "err", err)
		return err
	}
	return nil
}
func (repositoryImpl BulkUpdateRepositoryImpl) BulkUpdateConfigMapDataForGlobalById(id int, patch string) error {
	CmAppModel := &chartConfig.ConfigMapAppModel{}
	_, err := repositoryImpl.dbConnection.
		Model(CmAppModel).
		Set("config_map_data = ?", patch).
		Where("id = ?", id).
		Update()
	if err != nil {
		repositoryImpl.logger.Errorw("error in bulk updating config_map_data", "err", err)
		return err
	}
	return nil
}
func (repositoryImpl BulkUpdateRepositoryImpl) BulkUpdateSecretDataForGlobalById(id int, patch string) error {
	SecretAppModel := &chartConfig.ConfigMapAppModel{}
	_, err := repositoryImpl.dbConnection.
		Model(SecretAppModel).
		Set("secret_data = ?", patch).
		Where("id = ?", id).
		Update()
	if err != nil {
		repositoryImpl.logger.Errorw("error in bulk updating secret_data", "err", err)
		return err
	}
	return nil
}
func (repositoryImpl BulkUpdateRepositoryImpl) BulkUpdateConfigMapDataForEnvById(id int, patch string) error {
	CmEnvModel := &chartConfig.ConfigMapEnvModel{}
	_, err := repositoryImpl.dbConnection.
		Model(CmEnvModel).
		Set("config_map_data = ?", patch).
		Where("id = ?", id).
		Update()
	if err != nil {
		repositoryImpl.logger.Errorw("error in bulk updating config_map_data", "err", err)
		return err
	}
	return nil
}
func (repositoryImpl BulkUpdateRepositoryImpl) BulkUpdateSecretDataForEnvById(id int, patch string) error {
	SecretEnvModel := &chartConfig.ConfigMapEnvModel{}
	_, err := repositoryImpl.dbConnection.
		Model(SecretEnvModel).
		Set("secret_data = ?", patch).
		Where("id = ?", id).
		Update()
	if err != nil {
		repositoryImpl.logger.Errorw("error in bulk updating secret_data", "err", err)
		return err
	}
	return nil
}
