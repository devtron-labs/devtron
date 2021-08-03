package bulkUpdate

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strings"
)

type BulkUpdateReadme struct {
	tableName struct{} `sql:"bulk_update_readme" pg:",discard_unknown_columns"`
	Id        int      `sql:"id"`
	Resource  string   `sql:"resource"`
	Script    string   `sql:"script"`
	Readme    string   `sql:"readme"`
}

type BulkUpdateRepository interface {
	BuildAppNameQuery(appNameIncludes []string, appNameExcludes []string) string
	FindBulkUpdateReadme(operation string) (*BulkUpdateReadme, error)

	//For Deployment Template :
	FindDeploymentTemplateBulkAppNameForGlobal(appNameIncludes []string, appNameExcludes []string) ([]*pipelineConfig.App, error)
	FindDeploymentTemplateBulkAppNameForEnv(appNameIncludes []string, appNameExcludes []string, envId int) ([]*pipelineConfig.App, error)
	FindAppByChartId(chartId int) (*pipelineConfig.App, error)
	FindAppByChartEnvId(chartEnvId int) (*pipelineConfig.App, error)
	FindBulkChartsByAppNameSubstring(appNameIncludes []string, appNameExcludes []string) ([]*chartConfig.Chart, error)
	FindBulkChartsEnvByAppNameSubstring(appNameIncludes []string, appNameExcludes []string, envId int) ([]*chartConfig.EnvConfigOverride, error)
	BulkUpdateChartsValuesYamlAndGlobalOverrideById(id int, patch string) error
	BulkUpdateChartsEnvYamlOverrideById(id int, patch string) error

	//For ConfigMap & Secret :
	FindCMAndSecretBulkAppModelForGlobal(appNameIncludes []string, appNameExcludes []string) ([]*chartConfig.ConfigMapAppModel, error)
	FindCMAndSecretBulkAppModelForEnv(appNameIncludes []string, appNameExcludes []string, envId int) ([]*chartConfig.ConfigMapEnvModel, error)
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

func (repositoryImpl BulkUpdateRepositoryImpl) BuildAppNameQuery(appNameIncludes []string, appNameExcludes []string) string {
	var appNameQuery string
	appNameIncludesQuery := "app_name LIKE ANY (array["
	appNameIncludesQuery += "'" + strings.Join(appNameIncludes, "', '") + "'"
	appNameIncludesQuery += "])"
	appNameQuery = fmt.Sprintf("( %s ) ", appNameIncludesQuery)

	if appNameExcludes != nil {
		appNameExcludesQuery := "app_name NOT LIKE ALL (array["
		appNameExcludesQuery += "'" + strings.Join(appNameExcludes, "', '") + "'"
		appNameExcludesQuery += "])"
		appNameQuery += fmt.Sprintf("AND ( %s ) ", appNameExcludesQuery)
	}
	return appNameQuery
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindBulkUpdateReadme(resource string) (*BulkUpdateReadme, error) {
	bulkUpdateReadme := &BulkUpdateReadme{}
	err := repositoryImpl.dbConnection.
		Model(bulkUpdateReadme).Where("resource LIKE ?", resource).
		Select()
	return bulkUpdateReadme, err
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindDeploymentTemplateBulkAppNameForGlobal(appNameIncludes []string, appNameExcludes []string) ([]*pipelineConfig.App, error) {
	apps := []*pipelineConfig.App{}
	appNameQuery := repositoryImpl.BuildAppNameQuery(appNameIncludes, appNameExcludes)
	err := repositoryImpl.dbConnection.
		Model(&apps).Join("INNER JOIN charts ch ON app.id = ch.app_id").
		Where(appNameQuery).
		Where("app.active = ?", true).
		Where("ch.latest = ?", true).
		Select()
	return apps, err
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindDeploymentTemplateBulkAppNameForEnv(appNameIncludes []string, appNameExcludes []string, envId int) ([]*pipelineConfig.App, error) {
	apps := []*pipelineConfig.App{}
	appNameQuery := repositoryImpl.BuildAppNameQuery(appNameIncludes, appNameExcludes)
	err := repositoryImpl.dbConnection.
		Model(&apps).Join("INNER JOIN charts ch ON app.id = ch.app_id").
		Join("INNER JOIN chart_env_config_override ON ch.id = chart_env_config_override.chart_id").
		Where(appNameQuery).
		Where("app.active = ?", true).
		Where("chart_env_config_override.target_environment = ? ", envId).
		Where("chart_env_config_override.latest = ?", true).
		Select()
	return apps, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) FindCMAndSecretBulkAppModelForGlobal(appNameIncludes []string, appNameExcludes []string) ([]*chartConfig.ConfigMapAppModel, error) {
	CmAndSecretAppModel := []*chartConfig.ConfigMapAppModel{}
	appNameQuery := repositoryImpl.BuildAppNameQuery(appNameIncludes, appNameExcludes)
	err := repositoryImpl.dbConnection.
		Model(&CmAndSecretAppModel).Join("INNER JOIN app ON app.id = config_map_app_model.app_id").
		Where(appNameQuery).
		Where("app.active = ?", true).
		Select()
	return CmAndSecretAppModel, err
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindCMAndSecretBulkAppModelForEnv(appNameIncludes []string, appNameExcludes []string, envId int) ([]*chartConfig.ConfigMapEnvModel, error) {
	CmAndSecretEnvModel := []*chartConfig.ConfigMapEnvModel{}
	appNameQuery := repositoryImpl.BuildAppNameQuery(appNameIncludes, appNameExcludes)
	err := repositoryImpl.dbConnection.
		Model(&CmAndSecretEnvModel).Join("INNER JOIN app ON app.id = config_map_env_model.app_id").
		Where(appNameQuery).
		Where("app.active = ?", true).
		Where("config_map_env_model.environment_id = ? ", envId).
		Select()
	return CmAndSecretEnvModel, err
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindAppByChartId(chartId int) (*pipelineConfig.App, error) {
	app := &pipelineConfig.App{}
	err := repositoryImpl.dbConnection.
		Model(app).Join("INNER JOIN charts ch ON app.id = ch.app_id").
		Where("ch.id = ?", chartId).
		Where("app.active = ?", true).
		Where("ch.latest = ?", true).
		Select()
	return app, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) FindAppByChartEnvId(chartEnvId int) (*pipelineConfig.App, error) {
	app := &pipelineConfig.App{}
	err := repositoryImpl.dbConnection.
		Model(app).Join("INNER JOIN charts ch ON app.id = ch.app_id").
		Join("INNER JOIN chart_env_config_override ON ch.id = chart_env_config_override.chart_id").
		Where("app.active = ?", true).
		Where("chart_env_config_override.id = ? ", chartEnvId).
		Where("chart_env_config_override.latest = ?", true).
		Select()
	return app, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) FindBulkChartsByAppNameSubstring(appNameIncludes []string, appNameExcludes []string) ([]*chartConfig.Chart, error) {
	charts := []*chartConfig.Chart{}
	appNameQuery := repositoryImpl.BuildAppNameQuery(appNameIncludes, appNameExcludes)
	err := repositoryImpl.dbConnection.
		Model(&charts).Join("INNER JOIN app ON app.id=app_id ").
		Where(appNameQuery).
		Where("app.active = ?", true).
		Where("latest = ?", true).
		Select()
	return charts, err
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindBulkChartsEnvByAppNameSubstring(appNameIncludes []string, appNameExcludes []string, envId int) ([]*chartConfig.EnvConfigOverride, error) {
	charts := []*chartConfig.EnvConfigOverride{}
	appNameQuery := repositoryImpl.BuildAppNameQuery(appNameIncludes, appNameExcludes)
	err := repositoryImpl.dbConnection.
		Model(&charts).Join("INNER JOIN charts ch ON ch.id=env_config_override.chart_id").
		Join("INNER JOIN app ON app.id=ch.app_id").
		Where(appNameQuery).
		Where("app.active = ?", true).
		Where("env_config_override.target_environment = ?", envId).
		Where("env_config_override.latest = ?", true).
		Select()
	return charts, err
}
func (repositoryImpl BulkUpdateRepositoryImpl) BulkUpdateChartsValuesYamlAndGlobalOverrideById(id int, patch string) error {
	chart := &chartConfig.Chart{}
	_, err := repositoryImpl.dbConnection.
		Model(chart).
		Set("values_yaml = ?", patch).
		Set("global_override = ?", patch).
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
	SecretEnvModel := []*chartConfig.ConfigMapEnvModel{}
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
