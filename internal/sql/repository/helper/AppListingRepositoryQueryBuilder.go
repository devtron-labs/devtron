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
	"strings"
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
	Environments      []int    `json:"environments"`
	Statuses          []string `json:"statutes"`
	Teams             []int    `json:"teams"`
	AppStatuses       []string `json:"appStatuses"`
	TagFilters        []TagFilter
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
type TagFilterOperator string

// TagFilter holds one row of label filter sent by UI.
// key is always required.
// value is required for EQUALS/DOES_NOT_EQUAL/CONTAINS/DOES_NOT_CONTAIN.
// value must be absent for EXISTS/DOES_NOT_EXIST.
type TagFilter struct {
	Key      string            `json:"key"`
	Operator TagFilterOperator `json:"operator"`
	Value    *string           `json:"value"`
}

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)

const (
	TagFilterOperatorEquals         TagFilterOperator = "EQUALS"
	TagFilterOperatorDoesNotEqual   TagFilterOperator = "DOES_NOT_EQUAL"
	TagFilterOperatorContains       TagFilterOperator = "CONTAINS"
	TagFilterOperatorDoesNotContain TagFilterOperator = "DOES_NOT_CONTAIN"
	TagFilterOperatorExists         TagFilterOperator = "EXISTS"
	TagFilterOperatorDoesNotExist   TagFilterOperator = "DOES_NOT_EXIST"
)

const (
	AppNameSortBy      SortBy = "appNameSort"
	LastDeployedSortBy        = "lastDeployedSort"
)

var likePatternEscaper = strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_")

func (operator TagFilterOperator) IsValid() bool {
	switch operator {
	case TagFilterOperatorEquals,
		TagFilterOperatorDoesNotEqual,
		TagFilterOperatorContains,
		TagFilterOperatorDoesNotContain,
		TagFilterOperatorExists,
		TagFilterOperatorDoesNotExist:
		return true
	default:
		return false
	}
}

func (impl AppListingRepositoryQueryBuilder) BuildJobListingQuery(appIDs []int, statuses []string, environmentIds []int, sortOrder string) (string, []interface{}) {
	var queryParams []interface{}
	query := `select ci_pipeline.name as ci_pipeline_name,ci_pipeline.id as ci_pipeline_id,app.id as job_id,app.display_name 
		  as job_name, app.app_name,app.description,app.created_by,app.team_id,cwr.started_on,cwr.status,cem.environment_id,cwr.environment_id as last_triggered_environment_id from app left join ci_pipeline on 
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
		queryParams = append(queryParams, pg.In(statuses))
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
func (impl AppListingRepositoryQueryBuilder) JobsLastSucceededOnTimeQuery(ciPipelineIDs []int) (string, []interface{}) {
	// use this query with atleast 1 cipipeline id
	query := "select cw.ci_pipeline_id,cw.finished_on " +
		"as last_succeeded_on from ci_workflow cw inner join " +
		"(SELECT  ci_pipeline_id, MAX(finished_on) finished_on " +
		"FROM ci_workflow WHERE ci_workflow.status = 'Succeeded'" +
		"GROUP BY ci_pipeline_id) cws on cw.ci_pipeline_id = cws.ci_pipeline_id and cw.finished_on = cws.finished_on " +
		"where cw.ci_pipeline_id IN (?); "

	return query, []interface{}{pg.In(ciPipelineIDs)}
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
	queryParams = append(queryParams, whereConditionParams...)
	return query, queryParams
}

func (impl AppListingRepositoryQueryBuilder) TestForCommonAppFilter(appListingFilter AppListingFilter) (string, []interface{}) {
	queryTemp, queryParams := impl.CommonJoinSubQuery(appListingFilter)
	query := " FROM app a " + queryTemp
	return query, queryParams
}

func (impl AppListingRepositoryQueryBuilder) BuildAppListingQueryLastDeploymentTimeV2(pipelineIDs []int) (string, []interface{}) {
	whereCondition := ""
	queryParams := []interface{}{}
	if len(pipelineIDs) > 0 {
		whereCondition += " Where pco.pipeline_id IN (?) "
		queryParams = append(queryParams, pg.In(pipelineIDs))
	}
	query := "select pco.pipeline_id , MAX(pco.created_on) as last_deployed_time" +
		" from pipeline_config_override pco" + whereCondition +
		" GROUP BY pco.pipeline_id;"
	return query, queryParams
}

func (impl AppListingRepositoryQueryBuilder) GetAppIdsQueryWithPaginationForLastDeployedSearch(appListingFilter AppListingFilter) (string, []interface{}) {
	join, queryParams := impl.CommonJoinSubQuery(appListingFilter)
	countQuery := " (SELECT count(distinct(a.id)) as count FROM app a " + join + ") AS total_count "
	// appending query params for count query as well
	queryParams = append(queryParams, queryParams...)
	query := "SELECT a.id as app_id,MAX(pco.id) as last_deployed_time, " + countQuery +
		` FROM pipeline p 
		  INNER JOIN pipeline_config_override pco ON pco.pipeline_id = p.id and p.deleted=false 
		  RIGHT JOIN ( SELECT DISTINCT(a.id) as id FROM app a ` + join +
		` ) da on p.app_id = da.id and p.deleted=false  
		  INNER JOIN app a ON da.id = a.id  `
	if appListingFilter.SortOrder == Desc {
		query += ` GROUP BY a.id,total_count ORDER BY last_deployed_time DESC NULLS `
	} else {
		query += ` GROUP BY a.id,total_count ORDER BY last_deployed_time ASC NULLS `
	}
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
	//adding queryParams two times because join query is used in countQuery and mainQuery two times
	queryParams = append(queryParams, queryParams...)
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
	whereCondition := " WHERE a.active = ? and a.app_type = ? "
	queryParams = append(queryParams, true, CustomApp)
	if len(appListingFilter.Environments) > 0 {
		whereCondition += " and p.environment_id IN (?) "
		queryParams = append(queryParams, pg.In(appListingFilter.Environments))
	}

	if len(appListingFilter.Teams) > 0 {
		whereCondition += " and a.team_id IN (?) "
		queryParams = append(queryParams, pg.In(appListingFilter.Teams))
	}

	if appListingFilter.AppNameSearch != "" {
		whereCondition += " and a.app_name like ? "
		queryParams = append(queryParams, util.GetLIKEClauseQueryParam(appListingFilter.AppNameSearch))
	}

	if appListingFilter.DeploymentGroupId > 0 {
		whereCondition += " and dga.deployment_group_id = ? "
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
	if isNotDeployedFilterApplied {
		deploymentAppType := "manifest_download"
		whereCondition += " and (p.deployment_app_created=? and (p.deployment_app_type <> ? or dc.deployment_app_type <> ? ) or a.id NOT IN (SELECT app_id from pipeline) "
		queryParams = append(queryParams, false, deploymentAppType, deploymentAppType)
		if len(appStatusExcludingNotDeployed) > 0 {
			whereCondition += " or aps.status IN (?) "
			queryParams = append(queryParams, pg.In(appStatusExcludingNotDeployed))
		}
		whereCondition += " ) "
	} else if len(appStatusExcludingNotDeployed) > 0 {
		whereCondition += " and aps.status IN (?) "
		queryParams = append(queryParams, pg.In(appStatusExcludingNotDeployed))
	}

	// Tag filters are AND-combined for now as requested by product.
	// Each row translates to a correlated EXISTS/NOT EXISTS on app_label.
	tagWhereCondition, tagQueryParams := impl.buildTagFiltersWhereConditionAND(appListingFilter.TagFilters)
	whereCondition += tagWhereCondition
	queryParams = append(queryParams, tagQueryParams...)

	// Future OR support placeholder (intentionally disabled today):
	// orTagWhereCondition, orTagQueryParams := impl.buildTagFiltersWhereConditionOR(appListingFilter.TagFilters)
	// whereCondition += orTagWhereCondition
	// queryParams = append(queryParams, orTagQueryParams...)

	if len(appListingFilter.AppIds) > 0 {
		whereCondition += " and a.id IN (?) "
		queryParams = append(queryParams, pg.In(appListingFilter.AppIds))
	}
	return whereCondition, queryParams
}

func (impl AppListingRepositoryQueryBuilder) buildTagFiltersWhereConditionAND(tagFilters []TagFilter) (string, []interface{}) {
	if len(tagFilters) == 0 {
		return "", nil
	}
	var queryBuilder strings.Builder
	queryParams := make([]interface{}, 0, len(tagFilters)*2)
	for _, tagFilter := range tagFilters {
		predicate, predicateParams := impl.buildTagFilterPredicate(tagFilter)
		queryBuilder.WriteString(" and ")
		queryBuilder.WriteString(predicate)
		queryParams = append(queryParams, predicateParams...)
	}
	return queryBuilder.String(), queryParams
}

// buildTagFiltersWhereConditionOR is intentionally unused today.
// It is kept as documented dead code so switching to OR in future is straightforward.
func (impl AppListingRepositoryQueryBuilder) buildTagFiltersWhereConditionOR(tagFilters []TagFilter) (string, []interface{}) {
	if len(tagFilters) == 0 {
		return "", nil
	}
	clauses := make([]string, 0, len(tagFilters))
	queryParams := make([]interface{}, 0, len(tagFilters)*2)
	for _, tagFilter := range tagFilters {
		predicate, predicateParams := impl.buildTagFilterPredicate(tagFilter)
		clauses = append(clauses, predicate)
		queryParams = append(queryParams, predicateParams...)
	}
	return " and (" + strings.Join(clauses, " OR ") + ") ", queryParams
}

// buildTagFilterPredicate converts one UI tag filter row into a SQL predicate.
// Operator behavior (all case-sensitive):
// - EQUALS: key exists with exact value match.
// - DOES_NOT_EQUAL: key exists with at least one value different from target.
// - CONTAINS: key exists with at least one value containing target substring.
// - DOES_NOT_CONTAIN: key exists with at least one value not containing target substring.
// - EXISTS: key exists.
// - DOES_NOT_EXIST: key does not exist.
func (impl AppListingRepositoryQueryBuilder) buildTagFilterPredicate(tagFilter TagFilter) (string, []interface{}) {
	value := ""
	if tagFilter.Value != nil {
		value = *tagFilter.Value
	}
	switch tagFilter.Operator {
	case TagFilterOperatorEquals:
		return "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value = ?)",
			[]interface{}{tagFilter.Key, value}
	case TagFilterOperatorDoesNotEqual:
		// Best-practice semantics for multi-value keys:
		// include app when key exists and at least one value is different from target.
		return "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value <> ?)",
			[]interface{}{tagFilter.Key, value}
	case TagFilterOperatorContains:
		return "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value LIKE ? ESCAPE '\\')",
			[]interface{}{tagFilter.Key, buildContainsPattern(value)}
	case TagFilterOperatorDoesNotContain:
		// Best-practice semantics for multi-value keys:
		// include app when key exists and at least one value does not contain target.
		return "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ? and al.value NOT LIKE ? ESCAPE '\\')",
			[]interface{}{tagFilter.Key, buildContainsPattern(value)}
	case TagFilterOperatorExists:
		return "EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ?)",
			[]interface{}{tagFilter.Key}
	case TagFilterOperatorDoesNotExist:
		return "NOT EXISTS (SELECT 1 FROM app_label al WHERE al.app_id = a.id and al.key = ?)",
			[]interface{}{tagFilter.Key}
	default:
		// Invalid operator should never reach here due request validation.
		// Returning false condition keeps query safe if validation is bypassed.
		return "1 = 0", nil
	}
}

func buildContainsPattern(value string) string {
	// Escape SQL LIKE wildcard chars so "contains" behaves like plain substring search.
	escaped := likePatternEscaper.Replace(value)
	return "%" + escaped + "%"
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
