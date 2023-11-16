package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
)

const EmptyLikeRegex = "%%"

func BuildQueryForParentTypeCIOrWebhook(listingFilterOpts bean.ArtifactsListFilterOptions) string {
	commonPaginatedQueryPart := ""
	if listingFilterOpts.SearchString != EmptyLikeRegex {
		commonPaginatedQueryPart = fmt.Sprintf(" cia.image LIKE '%v' ", listingFilterOpts.SearchString)
	}
	orderByClause := " ORDER BY cia.id DESC"
	limitOffsetQueryPart := fmt.Sprintf(" LIMIT %v OFFSET %v", listingFilterOpts.Limit, listingFilterOpts.Offset)
	finalQuery := ""
	if listingFilterOpts.ParentStageType == bean.CI_WORKFLOW_TYPE {
		selectQuery := " SELECT cia.* "
		remainingQuery := " FROM ci_artifact cia" +
			" INNER JOIN ci_pipeline cp ON cp.id=cia.pipeline_id" +
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
	commonQuery := " FROM cd_workflow_runner " +
		" INNER JOIN cd_workflow ON cd_workflow.id=cd_workflow_runner.cd_workflow_id " +
		" INNER JOIN ci_artifact cia ON cia.id = cd_workflow.ci_artifact_id " +
		" WHERE ((cd_workflow.pipeline_id = %v AND cd_workflow_runner.workflow_type = '%v') " +
		"       OR (cd_workflow.pipeline_id = %v AND cd_workflow_runner.workflow_type = '%v' AND cd_workflow_runner.status IN ('Healthy','Succeeded'))) "

	if listingFilterOptions.SearchString != EmptyLikeRegex {
		commonQuery += " AND cia.image LIKE '%v' "
	}

	commonQuery = fmt.Sprintf(commonQuery, listingFilterOptions.PipelineId, listingFilterOptions.StageType, listingFilterOptions.ParentId, listingFilterOptions.ParentStageType, listingFilterOptions.SearchString)
	if len(listingFilterOptions.ExcludeArtifactIds) > 0 {
		commonQuery = commonQuery + fmt.Sprintf(" AND cd_workflow.ci_artifact_id NOT IN (%v)", helper.GetCommaSepratedString(listingFilterOptions.ExcludeArtifactIds))
	}

	totalCountQuery := "SELECT COUNT(DISTINCT ci_artifact_id) as total_count " + commonQuery
	selectQuery := fmt.Sprintf("SELECT cia.id , (%v) ", totalCountQuery)
	GroupByQuery := " GROUP BY cia.id "
	limitOffSetQuery := fmt.Sprintf(" LIMIT %v OFFSET %v", listingFilterOptions.Limit, listingFilterOptions.Offset)

	finalQuery := selectQuery + commonQuery + GroupByQuery + limitOffSetQuery
	return finalQuery
}

func BuildQueryForArtifactsForRollback(listingFilterOptions bean.ArtifactsListFilterOptions) string {
	commonQuery := " FROM cd_workflow_runner cdwr " +
		" INNER JOIN cd_workflow cdw ON cdw.id=cdwr.cd_workflow_id " +
		" INNER JOIN ci_artifact cia ON cia.id=cdw.ci_artifact_id " +
		" WHERE cdw.pipeline_id=%v AND cdwr.workflow_type = '%v' "
	if listingFilterOptions.SearchString != EmptyLikeRegex {
		commonQuery += " AND cia.image LIKE '%v' "
	}
	commonQuery = fmt.Sprintf(commonQuery, listingFilterOptions.PipelineId, listingFilterOptions.StageType, listingFilterOptions.SearchString)
	if len(listingFilterOptions.ExcludeWfrIds) > 0 {
		commonQuery = fmt.Sprintf(" %s AND cdwr.id NOT IN (%s)", commonQuery, helper.GetCommaSepratedString(listingFilterOptions.ExcludeWfrIds))
	}
	totalCountQuery := " SELECT COUNT(cia.id) as total_count " + commonQuery
	orderByQuery := " ORDER BY cdwr.id DESC "
	limitOffsetQuery := fmt.Sprintf("LIMIT %v OFFSET %v", listingFilterOptions.Limit, listingFilterOptions.Offset)
	finalQuery := fmt.Sprintf(" SELECT cdwr.id as cd_workflow_runner_id,cdwr.triggered_by,cdwr.started_on,cia.*,(%s) ", totalCountQuery) + commonQuery + orderByQuery + limitOffsetQuery
	return finalQuery
}
