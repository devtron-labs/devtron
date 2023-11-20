package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
)

const EmptyLikeRegex = "%%"

func BuildQueryForParentTypeCIOrWebhook(listingFilterOpts bean.ArtifactsListFilterOptions) string {
	commonPaginatedQueryPart := fmt.Sprintf(" cia.image LIKE '%v'", listingFilterOpts.SearchString)
	orderByClause := " ORDER BY cia.id DESC"
	limitOffsetQueryPart := fmt.Sprintf(" LIMIT %v OFFSET %v", listingFilterOpts.Limit, listingFilterOpts.Offset)
	finalQuery := ""
	if listingFilterOpts.ParentStageType == bean.CI_WORKFLOW_TYPE {
		selectQuery := " SELECT cia.* "
		remainingQuery := " FROM ci_artifact cia" +
			" INNER JOIN ci_pipeline cp ON (cp.id=cia.pipeline_id or (cp.id=cia.component_id and cia.data_source='post_ci' ) )" +
			" INNER JOIN pipeline p ON p.ci_pipeline_id = cp.id and p.id=%v" +
			" WHERE "
		remainingQuery = fmt.Sprintf(remainingQuery, listingFilterOpts.PipelineId)
		if len(listingFilterOpts.ExcludeArtifactIds) > 0 {
			remainingQuery += fmt.Sprintf("cia.id NOT IN (%s) AND ", helper.GetCommaSepratedString(listingFilterOpts.ExcludeArtifactIds))
		}

		countQuery := " SELECT count(cia.id)  as total_count"
		totalCountQuery := countQuery + remainingQuery + commonPaginatedQueryPart
		selectQuery = fmt.Sprintf("%s,(%s) ", selectQuery, totalCountQuery)
		finalQuery = selectQuery + remainingQuery + commonPaginatedQueryPart + orderByClause + limitOffsetQueryPart
	} else if listingFilterOpts.ParentStageType == bean.WEBHOOK_WORKFLOW_TYPE {
		selectQuery := " SELECT cia.* "
		remainingQuery := " FROM ci_artifact cia " +
			" WHERE cia.external_ci_pipeline_id = %v AND "
		remainingQuery = fmt.Sprintf(remainingQuery, listingFilterOpts.ParentId)
		if len(listingFilterOpts.ExcludeArtifactIds) > 0 {
			remainingQuery += fmt.Sprintf("cia.id NOT IN (%s) AND ", helper.GetCommaSepratedString(listingFilterOpts.ExcludeArtifactIds))
		}

		countQuery := " SELECT count(cia.id)  as total_count"
		totalCountQuery := countQuery + remainingQuery + commonPaginatedQueryPart
		selectQuery = fmt.Sprintf("%s,(%s) ", selectQuery, totalCountQuery)
		finalQuery = selectQuery + remainingQuery + commonPaginatedQueryPart + orderByClause + limitOffsetQueryPart

	}
	return finalQuery
}

func BuildQueryForArtifactsForCdStage(listingFilterOptions bean.ArtifactsListFilterOptions) string {

	commonQuery := " from ci_artifact LEFT JOIN cd_workflow ON ci_artifact.id = cd_workflow.ci_artifact_id" +
		" LEFT JOIN cd_workflow_runner ON cd_workflow_runner.cd_workflow_id=cd_workflow.id " +
		" Where (((cd_workflow_runner.id in (select MAX(cd_workflow_runner.id) OVER (PARTITION BY cd_workflow.ci_artifact_id) FROM cd_workflow_runner inner join cd_workflow on cd_workflow.id=cd_workflow_runner.cd_workflow_id))" +
		" AND ((cd_workflow.pipeline_id= %v and cd_workflow_runner.workflow_type = '%v' ) OR (cd_workflow.pipeline_id = %v AND cd_workflow_runner.workflow_type = '%v' AND cd_workflow_runner.status IN ('Healthy','Succeeded') )))" +
		" OR (ci_artifact.component_id = %v  and ci_artifact.data_source= '%v' ))" +
		" AND (ci_artifact.image LIKE '%v' )"

	commonQuery = fmt.Sprintf(commonQuery, listingFilterOptions.PipelineId, listingFilterOptions.StageType, listingFilterOptions.ParentId, listingFilterOptions.ParentStageType, listingFilterOptions.ParentId, listingFilterOptions.PluginStage, listingFilterOptions.SearchString)
	if len(listingFilterOptions.ExcludeArtifactIds) > 0 {
		commonQuery = commonQuery + fmt.Sprintf(" AND ( ci_artifact.id NOT IN (%v))", helper.GetCommaSepratedString(listingFilterOptions.ExcludeArtifactIds))
	}

	totalCountQuery := "SELECT COUNT(DISTINCT ci_artifact.id) as total_count " + commonQuery
	selectQuery := fmt.Sprintf("SELECT DISTINCT(ci_artifact.id) , (%v) ", totalCountQuery)
	//GroupByQuery := " GROUP BY cia.id "
	limitOffSetQuery := fmt.Sprintf(" order by ci_artifact.id desc LIMIT %v OFFSET %v", listingFilterOptions.Limit, listingFilterOptions.Offset)

	//finalQuery := selectQuery + commonQuery + GroupByQuery + limitOffSetQuery
	finalQuery := selectQuery + commonQuery + limitOffSetQuery
	return finalQuery
}

func BuildQueryForArtifactsForRollback(listingFilterOptions bean.ArtifactsListFilterOptions) string {
	commonQuery := " FROM cd_workflow_runner cdwr " +
		" INNER JOIN cd_workflow cdw ON cdw.id=cdwr.cd_workflow_id " +
		" INNER JOIN ci_artifact cia ON cia.id=cdw.ci_artifact_id " +
		" WHERE cdw.pipeline_id=%v AND cdwr.workflow_type = '%v' "

	commonQuery = fmt.Sprintf(commonQuery, listingFilterOptions.PipelineId, listingFilterOptions.StageType)
	if listingFilterOptions.SearchString != EmptyLikeRegex {
		commonQuery += fmt.Sprintf(" AND cia.image LIKE '%v' ", listingFilterOptions.SearchString)
	}
	if len(listingFilterOptions.ExcludeWfrIds) > 0 {
		commonQuery = fmt.Sprintf(" %s AND cdwr.id NOT IN (%s)", commonQuery, helper.GetCommaSepratedString(listingFilterOptions.ExcludeWfrIds))
	}
	totalCountQuery := " SELECT COUNT(cia.id) as total_count " + commonQuery
	orderByQuery := " ORDER BY cdwr.id DESC "
	limitOffsetQuery := fmt.Sprintf("LIMIT %v OFFSET %v", listingFilterOptions.Limit, listingFilterOptions.Offset)
	finalQuery := fmt.Sprintf(" SELECT cdwr.id as cd_workflow_runner_id,cdwr.triggered_by,cdwr.started_on,cia.*,(%s) ", totalCountQuery) + commonQuery + orderByQuery + limitOffsetQuery
	return finalQuery
}
