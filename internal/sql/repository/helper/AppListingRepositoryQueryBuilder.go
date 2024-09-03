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

package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type AppType int

const (
	CustomApp     AppType = 0 // cicd app
	ChartStoreApp AppType = 1 // helm app
	Job           AppType = 2 // jobs
	// ExternalChartStoreApp app-type is not stored in db
	ExternalChartStoreApp AppType = 3 // external helm app
)

type AppListingRepositoryQueryBuilder struct {
	logger *zap.SugaredLogger
}

func NewAppListingRepositoryQueryBuilder(logger *zap.SugaredLogger) AppListingRepositoryQueryBuilder {
	return AppListingRepositoryQueryBuilder{
		logger: logger,
	}
}

type AppListingFilter struct {
	Environments      []int     `json:"environments"`
	Statuses          []string  `json:"statutes"`
	Teams             []int     `json:"teams"`
	AppStatuses       []string  `json:"appStatuses"`
	AppNameSearch     string    `json:"appNameSearch"`
	SortOrder         SortOrder `json:"sortOrder"`
	SortBy            SortBy    `json:"sortBy"`
	Offset            int       `json:"offset"`
	Size              int       `json:"size"`
	DeploymentGroupId int       `json:"deploymentGroupId"`
	AppIds            []int     `json:"-"` // internal use only
}

type SortBy string
type SortOrder string

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)

const (
	AppNameSortBy      SortBy = "appNameSort"
	LastDeployedSortBy        = "lastDeployedSort"
)

func (impl AppListingRepositoryQueryBuilder) BuildJobListingQuery(appIDs []int, statuses []string, environmentIds []int, sortOrder string) (string, []interface{}) {
	var queryParams []interface{}
	query := `select ci_pipeline.name as ci_pipeline_name,ci_pipeline.id as ci_pipeline_id,app.id as job_id,app.display_name 
		  as job_name, app.app_name,app.description,app.team_id,cwr.started_on,cwr.status,cem.environment_id,cwr.environment_id as last_triggered_environment_id from app left join ci_pipeline on 
		  app.id = ci_pipeline.app_id and ci_pipeline.active=true left join (select cw.ci_pipeline_id, cw.status, cw.started_on, cw.environment_id 
		  from ci_workflow cw inner join (select ci_pipeline_id, MAX(started_on) max_started_on from ci_workflow group by ci_pipeline_id ) 
		  cws on cw.ci_pipeline_id = cws.ci_pipeline_id 
		  and cw.started_on = cws.max_started_on order by cw.ci_pipeline_id) cwr on cwr.ci_pipeline_id = ci_pipeline.id 
		  LEFT JOIN ci_env_mapping cem on cem.ci_pipeline_id = ci_pipeline.id 
		  where app.active = true and app.app_type = 2 `
	if len(appIDs) > 0 {
		query += " and app.id IN (?) "
		queryParams = append(queryParams, pg.In(appIDs))
	}
	if len(statuses) > 0 {
		query += " and cwr.status IN (?) "
		queryParams = append(queryParams, util.ProcessAppStatuses(statuses))
	}
	if len(environmentIds) > 0 {
		query += " and cwr.environment_id IN (?) "
		queryParams = append(queryParams, pg.In(environmentIds))
	}
	query += " order by app.display_name"
	if sortOrder == "DESC" {
		query += " DESC "
	}
	return query, queryParams
}
func (impl AppListingRepositoryQueryBuilder) OverviewCiPipelineQuery() string {
	query := "select ci_pipeline.id as ci_pipeline_id,ci_pipeline.name " +
		"as ci_pipeline_name,cwr.status,cwr.started_on,cem.environment_id,cwr.environment_id as last_triggered_environment_id from ci_pipeline" +
		" left join (select cw.ci_pipeline_id,cw.status,cw.started_on,cw.environment_id from ci_workflow cw" +
		" inner join (SELECT  ci_pipeline_id, MAX(started_on) max_started_on FROM ci_workflow GROUP BY ci_pipeline_id)" +
		" cws on cw.ci_pipeline_id = cws.ci_pipeline_id and cw.started_on = cws.max_started_on order by cw.ci_pipeline_id)" +
		" cwr on cwr.ci_pipeline_id = ci_pipeline.id" +
		" LEFT JOIN ci_env_mapping cem on cem.ci_pipeline_id = ci_pipeline.id " +
		"where ci_pipeline.active = true and ci_pipeline.app_id = ? ;"
	return query
}

// use this query with atleast 1 cipipeline id
func (impl AppListingRepositoryQueryBuilder) JobsLastSucceededOnTimeQuery(ciPipelineIDs []int) string {
	// use this query with atleast 1 cipipeline id
	query := "select cw.ci_pipeline_id,cw.finished_on " +
		"as last_succeeded_on from ci_workflow cw inner join " +
		"(SELECT  ci_pipeline_id, MAX(finished_on) finished_on " +
		"FROM ci_workflow WHERE ci_workflow.status = 'Succeeded'" +
		"GROUP BY ci_pipeline_id) cws on cw.ci_pipeline_id = cws.ci_pipeline_id and cw.finished_on = cws.finished_on " +
		"where cw.ci_pipeline_id IN (" + GetCommaSepratedString(ciPipelineIDs) + "); "

	return query
}

func getAppListingCommonQueryString() string {
	return " FROM pipeline p" +
		" INNER JOIN environment env ON env.id=p.environment_id" +
		" INNER JOIN cluster cluster ON cluster.id=env.cluster_id" +
		" RIGHT JOIN app a ON a.id=p.app_id  and p.deleted=false" +
		" RIGHT JOIN team t ON t.id=a.team_id " +
		" LEFT JOIN app_status aps on aps.app_id = a.id and p.environment_id = aps.env_id "
}

func (impl AppListingRepositoryQueryBuilder) GetQueryForAppEnvContainers(appListingFilter AppListingFilter) (string, []interface{}) {
	query := "SELECT p.environment_id , a.id AS app_id, a.app_name,p.id as pipeline_id, a.team_id ,aps.status as app_status "
	queryTemp, queryParams := impl.TestForCommonAppFilter(appListingFilter)
	query += queryTemp
	return query, queryParams
}

func (impl AppListingRepositoryQueryBuilder) CommonJoinSubQuery(appListingFilter AppListingFilter) (string, []interface{}) {
	var queryParams []interface{}
	query := ` LEFT JOIN pipeline p ON a.id=p.app_id  and p.deleted=? 
		       LEFT JOIN deployment_config dc ON ( p.app_id=dc.app_id and p.environment_id=dc.environment_id and dc.active=? ) 
			   LEFT JOIN app_status aps on aps.app_id = a.id and p.environment_id = aps.env_id `
	queryParams = append(queryParams, false, true)
	if appListingFilter.DeploymentGroupId != 0 {
		query = query + " INNER JOIN deployment_group_app dga ON a.id = dga.app_id "
	}
	whereCondition, whereConditionParams := impl.buildAppListingWhereCondition(appListingFilter)
	query = query + whereCondition
	queryParams = append(queryParams, whereConditionParams)
	return query, queryParams
}

func (impl AppListingRepositoryQueryBuilder) TestForCommonAppFilter(appListingFilter AppListingFilter) (string, []interface{}) {
	queryTemp, queryParams := impl.CommonJoinSubQuery(appListingFilter)
	query := " FROM app a " + queryTemp
	return query, queryParams
}

func (impl AppListingRepositoryQueryBuilder) BuildAppListingQueryLastDeploymentTimeV2(pipelineIDs []int) string {
	whereCondition := ""
	if len(pipelineIDs) > 0 {
		whereCondition += fmt.Sprintf(" Where pco.pipeline_id IN (%s) ", GetCommaSepratedString(pipelineIDs))
	}
	query := "select pco.pipeline_id , MAX(pco.created_on) as last_deployed_time" +
		" from pipeline_config_override pco" + whereCondition +
		" GROUP BY pco.pipeline_id;"
	return query
}

func (impl AppListingRepositoryQueryBuilder) GetAppIdsQueryWithPaginationForLastDeployedSearch(appListingFilter AppListingFilter) (string, []interface{}) {
	join, queryParams := impl.CommonJoinSubQuery(appListingFilter)
	countQuery := " (SELECT count(distinct(a.id)) as count FROM app a " + join + ") AS total_count "
	query := "SELECT a.id as app_id,MAX(pco.id) as last_deployed_time, " + countQuery +
		` FROM pipeline p 
		  INNER JOIN pipeline_config_override pco ON pco.pipeline_id = p.id and p.deleted=false 
		  RIGHT JOIN ( SELECT DISTINCT(a.id) as id FROM app a ` + join +
		` ) da on p.app_id = da.id and p.deleted=false  
		  INNER JOIN app a ON da.id = a.id  
	      GROUP BY a.id,total_count ORDER BY last_deployed_time ? NULLS `
	queryParams = append(queryParams, appListingFilter.SortOrder)
	if appListingFilter.SortOrder == "DESC" {
		query += " LAST "
	} else {
		query += " FIRST "
	}
	query += " LIMIT ? OFFSET ? "
	queryParams = append(queryParams, appListingFilter.Size, appListingFilter.Offset)
	return query, queryParams
}

func (impl AppListingRepositoryQueryBuilder) GetAppIdsQueryWithPaginationForAppNameSearch(appListingFilter AppListingFilter) (string, []interface{}) {
	orderByClause := impl.buildAppListingSortBy(appListingFilter)
	join, queryParams := impl.CommonJoinSubQuery(appListingFilter)
	countQuery := "( SELECT count(distinct(a.id)) as count FROM app a" + join + " ) as total_count"
	query := "SELECT DISTINCT(a.id) as app_id, a.app_name, " + countQuery +
		" FROM app a " + join
	if appListingFilter.SortBy == "appNameSort" {
		query += orderByClause
	}
	query += " LIMIT ? OFFSET ? "
	queryParams = append(queryParams, appListingFilter.Size, appListingFilter.Offset)
	return query, queryParams
}

func (impl AppListingRepositoryQueryBuilder) buildAppListingSortBy(appListingFilter AppListingFilter) string {
	orderByCondition := " ORDER BY a.app_name "

	if appListingFilter.SortOrder != "ASC" {
		orderByCondition += " DESC "
	}

	return orderByCondition
}

func (impl AppListingRepositoryQueryBuilder) buildAppListingWhereCondition(appListingFilter AppListingFilter) (string, []interface{}) {
	var queryParams []interface{}
	whereCondition := "WHERE a.active = ? and a.app_type = ? "
	queryParams = append(queryParams, true, 0)
	if len(appListingFilter.Environments) > 0 {
		whereCondition += "and p.environment_id IN (?) "
		queryParams = append(queryParams, pg.In(appListingFilter.Environments))
	}

	if len(appListingFilter.Teams) > 0 {
		whereCondition += "and a.team_id IN (?) "
		queryParams = append(queryParams, pg.In(appListingFilter.Teams))
	}

	if appListingFilter.AppNameSearch != "" {
		whereCondition += "and a.app_name like ? "
		queryParams = append(queryParams, util.GetLIKEClauseQueryParam(appListingFilter.AppNameSearch))
	}

	if appListingFilter.DeploymentGroupId > 0 {
		whereCondition += "and dga.deployment_group_id = ? "
		queryParams = append(queryParams, appListingFilter.DeploymentGroupId)
	}
	// add app-status filter here
	var appStatusExcludingNotDeployed []string
	var isNotDeployedFilterApplied bool
	if len(appListingFilter.AppStatuses) > 0 {
		for _, status := range appListingFilter.AppStatuses {
			if status == "NOT DEPLOYED" {
				isNotDeployedFilterApplied = true
			} else {
				appStatusExcludingNotDeployed = append(appStatusExcludingNotDeployed, status)
			}
		}
	}
	appStatuses := util.ProcessAppStatuses(appStatusExcludingNotDeployed)
	if isNotDeployedFilterApplied {
		deploymentAppType := "manifest_download"
		whereCondition += " and (p.deployment_app_created=? and (p.deployment_app_type != '?' || dc.deployment_app_type != '?' ) or a.id NOT IN (SELECT app_id from pipeline) "
		queryParams = append(queryParams, false, deploymentAppType, deploymentAppType)
		if len(appStatuses) > 0 {
			whereCondition += " or aps.status IN (?) "
			queryParams = append(queryParams, appStatuses)
		}
		whereCondition += " ) "
	} else if len(appStatuses) > 0 {
		whereCondition += " and aps.status IN (?) "
		queryParams = append(queryParams, appStatuses)
	}
	if len(appListingFilter.AppIds) > 0 {
		whereCondition += " and a.id IN (?) "
		queryParams = append(queryParams, pg.In(appListingFilter.AppIds))
	}
	return whereCondition, queryParams
}

func GetCommaSepratedString[T int | string](request []T) string {
	respString := ""
	for i, item := range request {
		respString += fmt.Sprintf("%v", item)
		if i != len(request)-1 {
			respString += ","
		}
	}
	return respString
}

func GetCommaSepratedStringWithComma[T int | string](appIds []T) string {
	appIdsString := ""
	for i, appId := range appIds {
		appIdsString += fmt.Sprintf("'%v'", appId)
		if i != len(appIds)-1 {
			appIdsString += ","
		}
	}
	return appIdsString
}
