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

type DeploymentTemplateComparisonMetadata struct {
	ChartId                     int                    `json:"chartRefId"`
	ChartVersion                string                 `json:"chartVersion,omitempty"`
	ChartType                   string                 `json:"chartType,omitempty"`
	EnvironmentId               int                    `json:"environmentId,omitempty"`
	EnvironmentName             string                 `json:"environmentName,omitempty"`
	DeploymentTemplateHistoryId int                    `json:"deploymentTemplateHistoryId,omitempty"`
	StartedOn                   *time.Time             `json:"startedOn,omitempty"`
	FinishedOn                  *time.Time             `json:"finishedOn,omitempty"`
	Status                      string                 `json:"status,omitempty"`
	PipelineId                  int                    `json:"pipelineId,omitempty"`
	Type                        DeploymentTemplateType `json:"type"`
}

type DeploymentTemplateRepository interface {
	FetchDeploymentHistoryWithChartRefs(appId int, envId int) ([]*DeploymentTemplateComparisonMetadata, error)
	FetchPipelineOverrideValues(id int) (string, error)
	FetchLatestDeploymentWithChartRefs(appId int, envId int) ([]*DeploymentTemplateComparisonMetadata, error)
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

func (impl DeploymentTemplateRepositoryImpl) FetchDeploymentHistoryWithChartRefs(appId int, envId int) ([]*DeploymentTemplateComparisonMetadata, error) {

	var result []*DeploymentTemplateComparisonMetadata
	limit := 15

	query := "select p.id as pipeline_id, dth.id as deployment_template_history_id," +
		"  wfr.finished_on, wfr.status, ceco.chart_id, c.chart_version FROM cd_workflow_runner wfr" +
		" JOIN cd_workflow wf ON wf.id = wfr.cd_workflow_id JOIN pipeline p ON p.id = wf.pipeline_id" +
		" JOIN deployment_template_history dth ON dth.deployed_on = wfr.started_on " +
		"JOIN pipeline_config_override pco ON pco.cd_workflow_id = wf.id " +
		"JOIN chart_env_config_override ceco ON ceco.id = pco.env_config_override_id JOIN charts c " +
		"ON c.id = ceco.chart_id where p.environment_id = ?  AND p.app_id = ? AND p.deleted = false  AND wfr.workflow_type = 'DEPLOY' " +
		"ORDER BY wfr.id DESC LIMIT ? ;"

	_, err := impl.dbConnection.Query(&result, query, envId, appId, limit)
	if err != nil {
		impl.Logger.Error("error in fetching deployment history", "error", err)
	}
	return result, err
}

func (impl DeploymentTemplateRepositoryImpl) FetchLatestDeploymentWithChartRefs(appId int, envId int) ([]*DeploymentTemplateComparisonMetadata, error) {

	var result []*DeploymentTemplateComparisonMetadata

	query := `
        WITH ranked_rows AS (
            SELECT 
                p.id as pipeline_id,
                p.environment_id, 
                dth.id as deployment_template_history_id, 
                ceco.chart_id, 
                c.chart_version, 
                ROW_NUMBER() OVER (PARTITION BY p.environment_id ORDER BY pco.id DESC) AS row_num
            FROM 
                pipeline p
            JOIN deployment_template_history dth ON dth.pipeline_id = p.id
            JOIN 
                pipeline_config_override pco ON pco.pipeline_id = p.id
            JOIN 
                chart_env_config_override ceco ON ceco.id = pco.env_config_override_id
            JOIN 
                charts c ON c.id = ceco.chart_id
            WHERE 
                p.app_id = ?
                AND p.deleted = false  
                AND p.environment_id NOT IN (?)
                AND dth.deployed = true
        )
        SELECT
            rr.pipeline_id, 
            rr.environment_id, 
            rr.deployment_template_history_id, 
            rr.chart_id, 
            rr.chart_version, 
            e.environment_name
        FROM 
            ranked_rows rr
        JOIN 
            environment e ON rr.environment_id = e.id
        WHERE 
            rr.row_num = 1;
    `

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
		impl.Logger.Error("error in fetching merged values yaml", "error", err)
	}
	return result.MergedValuesYaml, err
}
