package repository

import (
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type DeploymentTemplateType int

const (
	DefaultVersions            DeploymentTemplateType = 1
	PublishedOnEnvironments    DeploymentTemplateType = 2
	DeployedOnSelfEnvironment  DeploymentTemplateType = 3
	DeployedOnOtherEnvironment DeploymentTemplateType = 4
)

type FetchTemplateComparisonList struct {
	ChartId                  int                    `json:"chartRefId"`
	ChartVersion             string                 `json:"chartVersion,omitempty"`
	ChartType                string                 `json:"chartType,omitempty"`
	EnvironmentId            int                    `json:"environmentId,omitempty"`
	EnvironmentName          string                 `json:"environmentName,omitempty"`
	PipelineConfigOverrideId int                    `json:"pipelineConfigOverrideId,omitempty"`
	StartedOn                *time.Time             `json:"startedOn,omitempty"`
	FinishedOn               *time.Time             `json:"finishedOn,omitempty"`
	Status                   string                 `json:"status,omitempty"`
	Type                     DeploymentTemplateType `json:"type"`
}

type DeploymentTemplateRepository interface {
	FetchDeploymentHistoryWithChartRefs(appId int, envId int) ([]*FetchTemplateComparisonList, error)
	FetchPipelineOverrideValues(id int) (string, error)
	FetchLatestDeploymentWithChartRefs(appId int, envId int) ([]*FetchTemplateComparisonList, error)
}

type DeploymentTemplateRepositoryImpl struct {
	dbConnection *pg.DB
	Logger       *zap.SugaredLogger
}

func NewDeploymentTemplateRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *DeploymentTemplateRepositoryImpl {
	return &DeploymentTemplateRepositoryImpl{
		Logger:       logger,
		dbConnection: dbConnection,
	}
}

func (impl DeploymentTemplateRepositoryImpl) FetchDeploymentHistoryWithChartRefs(appId int, envId int) ([]*FetchTemplateComparisonList, error) {

	var result []*FetchTemplateComparisonList

	query := "SELECT pco.id as pipeline_config_override_id, wfr.started_on,   wfr.finished_on, wfr.status, ceco.chart_id, c.chart_version " +
		"FROM cd_workflow_runner wfr JOIN cd_workflow wf ON wf.id = wfr.cd_workflow_id " +
		"JOIN pipeline p ON p.id = wf.pipeline_id JOIN pipeline_config_override pco ON pco.cd_workflow_id = wf.id " +
		"JOIN chart_env_config_override ceco ON ceco.id = pco.env_config_override_id JOIN charts c ON c.id = ceco.chart_id " +
		" WHERE p.environment_id = ?  AND p.app_id = ?  AND " +
		"p.deleted = false  AND wfr.workflow_type = 'DEPLOY' ORDER BY" +
		" wfr.id DESC LIMIT 15;"

	_, err := impl.dbConnection.Query(&result, query, envId, appId)
	if err != nil {
		impl.Logger.Error("error in fetching deployment history", "error", err)
	}
	return result, err
}

func (impl DeploymentTemplateRepositoryImpl) FetchLatestDeploymentWithChartRefs(appId int, envId int) ([]*FetchTemplateComparisonList, error) {

	var result []*FetchTemplateComparisonList

	query := "WITH ranked_rows AS ( SELECT p.environment_id, pco.id as pipeline_config_override_id, ceco.chart_id, " +
		"c.chart_version, ROW_NUMBER() OVER (PARTITION BY p.environment_id ORDER BY pco.id DESC) AS row_num FROM pipeline p " +
		"JOIN pipeline_config_override pco ON pco.pipeline_id = p.id JOIN chart_env_config_override ceco ON ceco.id = pco.env_config_override_id" +
		" JOIN charts c ON c.id = ceco.chart_id WHERE  p.app_id = ? AND p.deleted = false AND p.environment_id NOT IN (?)) " +
		"SELECT environment_id, pipeline_config_override_id, chart_id,  chart_version" +
		" FROM ranked_rows " +
		"WHERE row_num = 1;"

	_, err := impl.dbConnection.Query(&result, query, appId, envId)
	if err != nil {
		impl.Logger.Error("error in fetching deployment history", "error", err)
	}
	return result, err
}

func (impl DeploymentTemplateRepositoryImpl) FetchPipelineOverrideValues(id int) (string, error) {

	type value struct {
		MergedValuesYaml string `sql:"merged_values_yaml"`
	}

	var result value

	query := "select merged_values_yaml from pipeline_config_override where id = ? ; "
	_, err := impl.dbConnection.Query(&result, query, id)
	if err != nil {
		impl.Logger.Error("error", "error", err)
	}
	return result.MergedValuesYaml, err
}
