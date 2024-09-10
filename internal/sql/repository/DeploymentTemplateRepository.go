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
	ChartRefId                  int                    `json:"chartRefId"`
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
		"  wfr.finished_on, wfr.status, c.chart_ref_id, c.chart_version FROM cd_workflow_runner wfr" +
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
        WITH pip AS (
		  SELECT 
		  p.id AS pipeline_id, 
		  MAX(dth.id) AS deployment_template_history_id,
		  p.environment_id,
		  e.environment_name
		  FROM pipeline AS p 
		  INNER JOIN deployment_template_history dth ON dth.pipeline_id = p.id 
		  INNER JOIN environment e ON e.id = p.environment_id 
		  WHERE p.deleted = false 
		  AND p.app_id = ? 
		  AND p.environment_id != ?
		  AND dth.deployed = true
		  GROUP BY p.id, e.environment_name
		)
		
		SELECT
				pip.pipeline_id,
				pip.environment_id, 
				pip.environment_name,
				pip.deployment_template_history_id,
				c.chart_ref_id, 
				c.chart_version
		FROM pip
		INNER JOIN pipeline_config_override pco ON pco.pipeline_id = pip.pipeline_id
		INNER JOIN chart_env_config_override ceco ON ceco.id = pco.env_config_override_id
		INNER JOIN charts c ON c.id = ceco.chart_id
		WHERE pco.id IN (
		  SELECT max(pco.id) 
		  FROM pipeline_config_override AS pco 
		  WHERE pipeline_id IN (SELECT pipeline_id FROM pip)
		  GROUP BY pipeline_id
		);  
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
