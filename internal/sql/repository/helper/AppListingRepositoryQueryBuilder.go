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
	"go.uber.org/zap"
	"strconv"
	"strings"
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
	AppNameSearch     string    `json:"appNameSearch"`
	SortOrder         SortOrder `json:"sortOrder"`
	SortBy            SortBy    `json:"sortBy"`
	Offset            int       `json:"offset"`
	Size              int       `json:"size"`
	DeploymentGroupId int       `json:"deploymentGroupId"`
}

type SortBy string
type SortOrder string

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)

const (
	AppNameSortBy SortBy = "appNameSort"
)

func (impl AppListingRepositoryQueryBuilder) BuildAppListingQuery(appListingFilter AppListingFilter) string {
	whereCondition := impl.buildAppListingWhereCondition(appListingFilter)
	orderByClause := impl.buildAppListingSortBy(appListingFilter)
	query := "SELECT env.id AS environment_id, env.environment_name, a.id AS app_id, a.app_name,  env.default, p.id as pipeline_id" +
		" FROM pipeline p" +
		" INNER JOIN environment env ON env.id=p.environment_id" +
		" RIGHT JOIN app a ON a.id=p.app_id  and p.deleted=false "
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

func (impl AppListingRepositoryQueryBuilder) buildAppListingSortBy(appListingFilter AppListingFilter) string {
	orderByCondition := " ORDER BY p.updated_on desc "
	return orderByCondition
}

func (impl AppListingRepositoryQueryBuilder) buildAppListingWhereCondition(appListingFilter AppListingFilter) string {
	whereCondition := "WHERE a.active = true and a.app_store is false and env.active = true "
	if len(appListingFilter.Environments) > 0 {
		envIds := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(appListingFilter.Environments)), ","), "[]")
		whereCondition = whereCondition + "and env.id IN (" + envIds + ") "
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
	return whereCondition
}
