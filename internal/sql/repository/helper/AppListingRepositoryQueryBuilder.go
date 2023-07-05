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

package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

type AppType int

const (
	CustomApp     AppType = 0 // cicd app
	ChartStoreApp AppType = 1 // helm app
	Job           AppType = 2 // jobs
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
	AppIds            []int     `json:"-"` //internal use only
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

func (impl AppListingRepositoryQueryBuilder) BuildJobListingQuery(appIDs []int, statuses []string, sortOrder string) string {
	query := "select ci_pipeline.name as ci_pipeline_name,ci_pipeline.id as ci_pipeline_id,app.id as job_id,app.display_name " +
		"as job_name,app.description,cwr.started_on,cwr.status,env1.environment_name,cem.environment_id,ci_workflow.environment_name as last_triggered_environment_name from app left join ci_pipeline on" +
		" app.id = ci_pipeline.app_id and ci_pipeline.active=true left join (select cw.ci_pipeline_id, cw.status, cw.started_on " +
		"  from ci_workflow cw inner join (select ci_pipeline_id, MAX(started_on) max_started_on from ci_workflow group by ci_pipeline_id ) " +
		"cws on cw.ci_pipeline_id = cws.ci_pipeline_id " +
		"and cw.started_on = cws.max_started_on order by cw.ci_pipeline_id) cwr on cwr.ci_pipeline_id = ci_pipeline.id " +
		"LEFT JOIN ci_env_mapping cem on cem.ci_pipeline_id = ci_pipeline.id " +
		"left join environment env on env.id = cem.environment_id" +
		" where app.active = true and app.app_type = 2 "
	if len(appIDs) > 0 {
		query += "and app.id IN (" + GetCommaSepratedString(appIDs) + ") "
	}
	if len(statuses) > 0 {
		query += "and cwr.status IN (" + util.ProcessAppStatuses(statuses) + ") "
	}
	query += " order by app.display_name"
	if sortOrder == "DESC" {
		query += " DESC "
	}
	return query
}
func (impl AppListingRepositoryQueryBuilder) OverviewCiPipelineQuery() string {
	query := "select ci_pipeline.id as ci_pipeline_id,ci_pipeline.name " +
		"as ci_pipeline_name,cwr.status,cwr.started_on,env1.environment_name,cem.last_triggered_env_id as environment_id,ci_workflow.environment_name as last_triggered_environment_name from ci_pipeline" +
		" left join (select cw.ci_pipeline_id,cw.status,cw.started_on from ci_workflow cw" +
		" inner join (SELECT  ci_pipeline_id, MAX(started_on) max_started_on FROM ci_workflow GROUP BY ci_pipeline_id)" +
		" cws on cw.ci_pipeline_id = cws.ci_pipeline_id and cw.started_on = cws.max_started_on order by cw.ci_pipeline_id)" +
		" cwr on cwr.ci_pipeline_id = ci_pipeline.id" +
		" LEFT JOIN ci_env_mapping cem on cem.ci_pipeline_id = ci_pipeline.id " +
		"left join environment env on env.id = cem.environment_id " +
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

func (impl AppListingRepositoryQueryBuilder) BuildAppListingQueryForAppIds(appListingFilter AppListingFilter) string {
	whereCondition := impl.buildAppListingWhereCondition(appListingFilter)
	orderByClause := impl.buildAppListingSortBy(appListingFilter)
	query := "SELECT a.id as app_id " + getAppListingCommonQueryString()
	if appListingFilter.DeploymentGroupId != 0 {
		query = query + " INNER JOIN deployment_group_app dga ON a.id = dga.app_id "
	}
	query = query + whereCondition + orderByClause
	return query
}

func (impl AppListingRepositoryQueryBuilder) BuildAppListingQuery(appListingFilter AppListingFilter) string {
	whereCondition := impl.buildAppListingWhereCondition(appListingFilter)
	orderByClause := impl.buildAppListingSortBy(appListingFilter)
	query := "SELECT env.id AS environment_id, env.environment_name,env.namespace as namespace ,a.id AS app_id, a.app_name, env.default,aps.status as app_status," +
		" p.id as pipeline_id, env.active, a.team_id, t.name as team_name" +
		" , cluster.cluster_name as cluster_name" + getAppListingCommonQueryString()
	if appListingFilter.DeploymentGroupId != 0 {
		query = query + " INNER JOIN deployment_group_app dga ON a.id = dga.app_id "
	}
	query = query + whereCondition + orderByClause
	return query
}

func (impl AppListingRepositoryQueryBuilder) BuildAppListingQueryLastDeploymentTime() string {
	query := "select DISTINCT ON( pco.pipeline_id) pco.pipeline_id, pco.pipeline_release_counter, pco.created_on as last_deployed_time," +
		" cia.data_source, cia.material_info as material_info_json, cia.id as ci_artifact_id" +
		" from pipeline_config_override pco" +
		" inner join ci_artifact cia on cia.id=pco.ci_artifact_id" +
		" order by pco.pipeline_id,pco.pipeline_release_counter desc;"
	return query
}

func (impl AppListingRepositoryQueryBuilder) GetQueryForAppEnvContainerss(appListingFilter AppListingFilter) string {

	query := "SELECT p.environment_id , a.id AS app_id, a.app_name,p.id as pipeline_id, a.team_id ,aps.status as app_status "

	query += impl.TestForCommonAppFilter(appListingFilter)
	return query
}

func (impl AppListingRepositoryQueryBuilder) CommonJoinSubQuery(appListingFilter AppListingFilter) string {
	whereCondition := impl.buildAppListingWhereCondition(appListingFilter)

	query := " LEFT JOIN pipeline p ON a.id=p.app_id  and p.deleted=false " +
		" LEFT JOIN app_status aps on aps.app_id = a.id and p.environment_id = aps.env_id "

	if appListingFilter.DeploymentGroupId != 0 {
		query = query + " INNER JOIN deployment_group_app dga ON a.id = dga.app_id "
	}

	query = query + whereCondition

	return query
}
func (impl AppListingRepositoryQueryBuilder) TestForCommonAppFilter(appListingFilter AppListingFilter) string {
	query := " FROM app a" + impl.CommonJoinSubQuery(appListingFilter)
	return query
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

func (impl AppListingRepositoryQueryBuilder) GetAppIdsQueryWithPaginationForLastDeployedSearch(appListingFilter AppListingFilter) string {
	join := impl.CommonJoinSubQuery(appListingFilter)
	countQuery := " (SELECT count(distinct(a.id)) as count " +
		" FROM app a " + join + ") AS total_count "

	query := "SELECT a.id as app_id,MAX(pco.id) as last_deployed_time, " + countQuery +
		" FROM pipeline p " +
		" INNER JOIN pipeline_config_override pco ON pco.pipeline_id = p.id and p.deleted=false " +
		" RIGHT JOIN ( SELECT DISTINCT(a.id) as id FROM app a " + join + " ) da on p.app_id = da.id and p.deleted=false " +
		" INNER JOIN app a ON da.id = a.id "
	query += fmt.Sprintf(" GROUP BY a.id,total_count ORDER BY last_deployed_time %s NULLS ", appListingFilter.SortOrder)
	if appListingFilter.SortOrder == "DESC" {
		query += " LAST "
	} else {
		query += " FIRST "
	}
	query += fmt.Sprintf(" LIMIT %v OFFSET %v", appListingFilter.Size, appListingFilter.Offset)
	return query
}

func (impl AppListingRepositoryQueryBuilder) GetAppIdsQueryWithPaginationForAppNameSearch(appListingFilter AppListingFilter) string {
	orderByClause := impl.buildAppListingSortBy(appListingFilter)
	join := impl.CommonJoinSubQuery(appListingFilter)
	countQuery := "( SELECT count(distinct(a.id)) as count FROM app a" + join + " ) as total_count"
	query := "SELECT DISTINCT(a.id) as app_id, a.app_name, " + countQuery +
		" FROM app a " + join
	if appListingFilter.SortBy == "appNameSort" {
		query += orderByClause
	}
	query += fmt.Sprintf("LIMIT %v OFFSET %v", appListingFilter.Size, appListingFilter.Offset)
	return query
}

func (impl AppListingRepositoryQueryBuilder) buildAppListingSortBy(appListingFilter AppListingFilter) string {
	orderByCondition := " ORDER BY a.app_name "

	if appListingFilter.SortOrder != "ASC" {
		orderByCondition += " DESC "
	}

	return orderByCondition
}

func (impl AppListingRepositoryQueryBuilder) buildJobListingSortBy(appListingFilter AppListingFilter) string {
	orderByCondition := " ORDER BY a.name"
	return orderByCondition
}

func (impl AppListingRepositoryQueryBuilder) buildJobListingWhereCondition(jobListingFilter AppListingFilter) string {
	whereCondition := "WHERE a.active = true and a.app_type = 2 "

	if len(jobListingFilter.Teams) > 0 {
		teamIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(jobListingFilter.Teams)), ","), "[]")
		whereCondition = whereCondition + "and a.team_id IN (" + teamIds + ") "
	}

	if jobListingFilter.AppNameSearch != "" {
		likeClause := "'%" + jobListingFilter.AppNameSearch + "%'"
		whereCondition = whereCondition + "and a.display_name like " + likeClause + " "
	}
	// add job stats filter here
	if len(jobListingFilter.AppStatuses) > 0 {
		appStatuses := util.ProcessAppStatuses(jobListingFilter.AppStatuses)
		whereCondition = whereCondition + "and aps.status IN (" + appStatuses + ") "
	}
	return whereCondition
}

func (impl AppListingRepositoryQueryBuilder) buildAppListingWhereCondition(appListingFilter AppListingFilter) string {
	whereCondition := "WHERE a.active = true and a.app_type = 0 "
	if len(appListingFilter.Environments) > 0 {
		envIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(appListingFilter.Environments)), ","), "[]")
		whereCondition = whereCondition + "and p.environment_id IN (" + envIds + ") "
	}

	if len(appListingFilter.Teams) > 0 {
		teamIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(appListingFilter.Teams)), ","), "[]")
		whereCondition = whereCondition + "and a.team_id IN (" + teamIds + ") "
	}

	if appListingFilter.AppNameSearch != "" {
		likeClause := "'%" + appListingFilter.AppNameSearch + "%'"
		whereCondition = whereCondition + "and a.app_name like " + likeClause + " "
	}

	if appListingFilter.DeploymentGroupId > 0 {
		whereCondition = whereCondition + "and dga.deployment_group_id = " + strconv.Itoa(appListingFilter.DeploymentGroupId) + " "
	}
	//add app-status filter here
	if len(appListingFilter.AppStatuses) > 0 {
		appStatuses := util.ProcessAppStatuses(appListingFilter.AppStatuses)
		whereCondition = whereCondition + "and aps.status IN (" + appStatuses + ") "
	}

	if len(appListingFilter.AppIds) > 0 {
		appIds := GetCommaSepratedString(appListingFilter.AppIds)
		whereCondition = whereCondition + "and a.id IN (" + appIds + ") "
	}
	return whereCondition
}
func GetCommaSepratedString(appIds []int) string {
	appIdsString := ""
	for i, appId := range appIds {
		appIdsString += fmt.Sprintf("%d", appId)
		if i != len(appIds)-1 {
			appIdsString += ","
		}
	}
	return appIdsString
}
