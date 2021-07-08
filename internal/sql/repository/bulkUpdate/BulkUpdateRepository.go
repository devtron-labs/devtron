package bulkUpdate

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type BulkUpdateReadme struct {
	tableName struct{} `sql:"bulk_update_readme" pg:",discard_unknown_columns"`
	Id        int      `sql:"id"`
	Operation string   `sql:"operation"`
	Script    string   `sql:"script"`
	Readme    string   `sql:"readme"`
}

type BulkUpdateRepository interface {
	BuildAppNameQuery(appNameIncludes []string, appNameExcludes []string) string
	FindBulkUpdateReadme(operation string) (*BulkUpdateReadme, error)
	FindBulkAppNameForGlobal(appNameIncludes []string, appNameExcludes []string) ([]*pipelineConfig.App, error)
	FindBulkAppNameForEnv(appNameIncludes []string, appNameExcludes []string, envId int) ([]*pipelineConfig.App, error)
	FindBulkChartsByAppNameSubstring(appNameIncludes []string, appNameExcludes []string) ([]*chartConfig.Chart, error)
	FindBulkChartsEnvByAppNameSubstring(appNameIncludes []string, appNameExcludes []string, envId int) ([]*chartConfig.EnvConfigOverride, error)
	BulkUpdateChartsValuesYamlAndGlobalOverrideById(final map[int]string) error
	BulkUpdateChartsEnvYamlOverrideById(final map[int]string) error
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
	var appNameIncludesQuery string
	for i, appNameInclude := range appNameIncludes {
		if i == 0 {
			appNameIncludesQuery += fmt.Sprintf("app_name LIKE '%s' ", appNameInclude)
		} else {
			appNameIncludesQuery += fmt.Sprintf("OR app_name LIKE '%s' ", appNameInclude)
		}
	}
	var appNameExcludesQuery string
	for i, appNameExclude := range appNameExcludes {
		if i == 0 {
			appNameExcludesQuery += fmt.Sprintf("app_name NOT LIKE '%s' ", appNameExclude)
		} else {
			appNameExcludesQuery += fmt.Sprintf("AND app_name NOT LIKE '%s' ", appNameExclude)
		}
	}
	appNameQuery := fmt.Sprintf("( %s ) AND ( %s )", appNameIncludesQuery, appNameExcludesQuery)
	return appNameQuery
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindBulkUpdateReadme(operation string) (*BulkUpdateReadme, error) {
	bulkUpdateReadme := &BulkUpdateReadme{}
	err := repositoryImpl.dbConnection.
		Model(bulkUpdateReadme).Where("operation LIKE ?", operation).
		Select()
	return bulkUpdateReadme, err
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindBulkAppNameForGlobal(appNameIncludes []string, appNameExcludes []string) ([]*pipelineConfig.App, error) {
	apps := []*pipelineConfig.App{}
	if len(appNameIncludes) == 0 || len(appNameExcludes) == 0 {
		return apps, nil
	}
	appNameQuery := repositoryImpl.BuildAppNameQuery(appNameIncludes, appNameExcludes)
	err := repositoryImpl.dbConnection.
		Model(&apps).Join("INNER JOIN charts ch ON app.id = ch.app_id").
		Where(appNameQuery).
		Where("app.active = ?", true).
		Where("ch.latest = ?", true).
		Select()
	return apps, err
}

func (repositoryImpl BulkUpdateRepositoryImpl) FindBulkAppNameForEnv(appNameIncludes []string, appNameExcludes []string, envId int) ([]*pipelineConfig.App, error) {
	apps := []*pipelineConfig.App{}
	if len(appNameIncludes) == 0 || len(appNameExcludes) == 0 {
		return apps, nil
	}
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
func (repositoryImpl BulkUpdateRepositoryImpl) FindBulkChartsByAppNameSubstring(appNameIncludes []string, appNameExcludes []string) ([]*chartConfig.Chart, error) {
	charts := []*chartConfig.Chart{}
	if len(appNameIncludes) == 0 || len(appNameExcludes) == 0 {
		return charts, nil
	}
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
	if len(appNameIncludes) == 0 || len(appNameExcludes) == 0 {
		return charts, nil
	}
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
func (repositoryImpl BulkUpdateRepositoryImpl) BulkUpdateChartsValuesYamlAndGlobalOverrideById(final map[int]string) error {
	chart := &chartConfig.Chart{}
	for id, patch := range final {
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
	}
	return nil
}
func (repositoryImpl BulkUpdateRepositoryImpl) BulkUpdateChartsEnvYamlOverrideById(final map[int]string) error {
	chartEnv := &chartConfig.EnvConfigOverride{}
	for id, patch := range final {
		_, err := repositoryImpl.dbConnection.
			Model(chartEnv).
			Set("env_override_yaml = ?", patch).
			Where("id = ?", id).
			Update()
		if err != nil {
			repositoryImpl.logger.Errorw("error in bulk updating deployment template charts(for env)", "err", err)
			return err
		}
	}
	return nil
}
