package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
)

func BuildQueryForParentTypeCIOrWebhook(listingFilterOpts bean.ArtifactsListFilterOptions, isApprovalNode bool) string {
	commonPaginatedQueryPart := fmt.Sprintf(" cia.image LIKE '%v'", listingFilterOpts.SearchString)
	orderByClause := " ORDER BY cia.id DESC"
	limitOffsetQueryPart := fmt.Sprintf(" LIMIT %v OFFSET %v", listingFilterOpts.Limit, listingFilterOpts.Offset)
	finalQuery := ""
	commonApprovalNodeSubQueryPart := fmt.Sprintf("cia.id NOT IN "+
		" ( "+
		" SELECT DISTINCT dar.ci_artifact_id "+
		" FROM deployment_approval_request dar "+
		" WHERE dar.pipeline_id = %v "+
		" AND dar.active=true "+
		" AND dar.artifact_deployment_triggered = false"+
		" ) AND ", listingFilterOpts.PipelineId)

	if listingFilterOpts.ParentStageType == bean.CI_WORKFLOW_TYPE {
		selectQuery := " SELECT cia.* "
		remainingQuery := " FROM ci_artifact cia" +
			" INNER JOIN ci_pipeline cp ON cp.id=cia.pipeline_id" +
			" INNER JOIN pipeline p ON p.ci_pipeline_id = cp.id and p.id=%v" +
			" WHERE "
		remainingQuery = fmt.Sprintf(remainingQuery, listingFilterOpts.PipelineId)
		if isApprovalNode {
			remainingQuery += commonApprovalNodeSubQueryPart
		} else if len(listingFilterOpts.ExcludeArtifactIds) > 0 {
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
		if isApprovalNode {
			remainingQuery += commonApprovalNodeSubQueryPart
		} else if len(listingFilterOpts.ExcludeArtifactIds) > 0 {
			remainingQuery += fmt.Sprintf("cia.id NOT IN (%s) AND ", helper.GetCommaSepratedString(listingFilterOpts.ExcludeArtifactIds))
		}

		countQuery := " SELECT count(cia.id)  as total_count"
		totalCountQuery := countQuery + remainingQuery + commonPaginatedQueryPart
		selectQuery = fmt.Sprintf("%s,(%s) ", selectQuery, totalCountQuery)
		finalQuery = selectQuery + remainingQuery + commonPaginatedQueryPart + orderByClause + limitOffsetQueryPart

	}
	return finalQuery
}

func BuildQueryForArtifactsForCdStage(listingFilterOptions bean.ArtifactsListFilterOptions, isApprovalNode bool) string {
	commonQuery := " FROM cd_workflow_runner " +
		" INNER JOIN cd_workflow ON cd_workflow.id=cd_workflow_runner.cd_workflow_id " +
		" INNER JOIN ci_artifact cia ON cia.id = cd_workflow.ci_artifact_id " +
		" WHERE ((cd_workflow.pipeline_id = %v AND cd_workflow_runner.workflow_type = '%v') " +
		"       OR (cd_workflow.pipeline_id = %v AND cd_workflow_runner.workflow_type = '%v' AND cd_workflow_runner.status IN ('Healthy','Succeeded')))" +
		" AND cia.image LIKE '%v' "
	commonQuery = fmt.Sprintf(commonQuery, listingFilterOptions.PipelineId, listingFilterOptions.StageType, listingFilterOptions.ParentId, listingFilterOptions.ParentStageType, listingFilterOptions.SearchString)
	if isApprovalNode {
		commonQuery = commonQuery + fmt.Sprintf(" AND cd_workflow.ci_artifact_id NOT IN (SELECT DISTINCT dar.ci_artifact_id FROM deployment_approval_request dar WHERE dar.pipeline_id = %v AND dar.active=true AND dar.artifact_deployment_triggered = false)", listingFilterOptions.PipelineId)
	} else if len(listingFilterOptions.ExcludeArtifactIds) > 0 {
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
		" WHERE cdw.pipeline_id=%v AND cdwr.workflow_type = '%v' AND cia.image LIKE '%v'"
	commonQuery = fmt.Sprintf(commonQuery, listingFilterOptions.PipelineId, listingFilterOptions.StageType, listingFilterOptions.SearchString)
	if len(listingFilterOptions.ExcludeArtifactIds) > 0 {
		commonQuery = fmt.Sprintf(" %s AND cdwr.id NOT IN (%s)", commonQuery, helper.GetCommaSepratedString(listingFilterOptions.ExcludeWfrIds))
	}
	totalCountQuery := " SELECT COUNT(cia.id) as total_count " + commonQuery
	orderByQuery := " ORDER BY cdwr.id DESC "
	limitOffsetQuery := fmt.Sprintf("LIMIT %v OFFSET %v", listingFilterOptions.Limit, listingFilterOptions.Offset)
	finalQuery := fmt.Sprintf(" SELECT cdwr.id as cd_workflow_runner_id,cdwr.triggered_by,cdwr.started_on,cia.*,(%s) ", totalCountQuery) + commonQuery + orderByQuery + limitOffsetQuery
	return finalQuery
}

func BuildApprovedOnlyArtifactsWithFilter(listingFilterOpts bean.ArtifactsListFilterOptions) string {
	withQuery := "WITH " +
		" approved_images AS " +
		" ( " +
		" SELECT approval_request_id,count(approval_request_id) AS approval_count " +
		" FROM deployment_approval_user_data daud " +
		" WHERE user_response is NULL " +
		" GROUP BY approval_request_id " +
		" ) "
	countQuery := " SELECT count(cia.created_on) as total_count"

	commonQueryPart := " FROM deployment_approval_request dar " +
		" INNER JOIN approved_images ai ON ai.approval_request_id=dar.id AND ai.approval_count >= %v " +
		" INNER JOIN ci_artifact cia ON cia.id = dar.ci_artifact_id " +
		" WHERE dar.active=true AND dar.artifact_deployment_triggered = false AND dar.pipeline_id = %v AND " +
		" cia.image LIKE '%v' "
	commonQueryPart = fmt.Sprintf(commonQueryPart, listingFilterOpts.ApproversCount, listingFilterOpts.PipelineId, listingFilterOpts.SearchString)
	if len(listingFilterOpts.ExcludeArtifactIds) > 0 {
		commonQueryPart += fmt.Sprintf(" AND cia.id NOT IN (%s) ", helper.GetCommaSepratedString(listingFilterOpts.ExcludeArtifactIds))
	}

	orderByClause := " ORDER BY cia.created_on "
	limitOffsetQueryPart := fmt.Sprintf(" LIMIT %v OFFSET %v ", listingFilterOpts.Limit, listingFilterOpts.Offset)
	totalCountQuery := withQuery + countQuery + commonQueryPart
	selectQuery := fmt.Sprintf(" SELECT cia.*,(%s)", totalCountQuery)
	finalQuery := withQuery + selectQuery + commonQueryPart + orderByClause + limitOffsetQueryPart
	return finalQuery
}

func BuildQueryForApprovedArtifactsForRollback(listingFilterOpts bean.ArtifactsListFilterOptions) string {
	subQuery := "WITH approved_requests AS " +
		" (SELECT approval_request_id,count(approval_request_id) AS approval_count " +
		" FROM deployment_approval_user_data " +
		" WHERE user_response is NULL " +
		" GROUP BY approval_request_id ) " +
		" SELECT approval_request_id " +
		" FROM approved_requests WHERE approval_count >= %v "
	subQuery = fmt.Sprintf(subQuery, listingFilterOpts.ApproversCount)
	commonQuery := " FROM cd_workflow_runner cdwr " +
		"   INNER JOIN cd_workflow cdw ON cdw.id=cdwr.cd_workflow_id" +
		"	INNER JOIN ci_artifact cia ON cia.id=cdw.ci_artifact_id" +
		"	INNER JOIN deployment_approval_request dar ON dar.ci_artifact_id = cdw.ci_artifact_id" +
		"   WHERE dar.id IN (%s) AND cdw.pipeline_id = %v" +
		"   AND cdwr.workflow_type = '%v'" +
		"   AND cia.image LIKE '%v'"
	commonQuery = fmt.Sprintf(commonQuery, subQuery, listingFilterOpts.PipelineId, listingFilterOpts.StageType, listingFilterOpts.SearchString)
	if len(listingFilterOpts.ExcludeArtifactIds) > 0 {
		commonQuery = fmt.Sprintf(" %s AND cia.id NOT IN (%s)", commonQuery, helper.GetCommaSepratedString(listingFilterOpts.ExcludeArtifactIds))
	}

	totalCountQuery := " SELECT COUNT(cia.id) as total_count " + commonQuery
	orderByQuery := " ORDER BY cdwr.id DESC "
	limitOffsetQuery := fmt.Sprintf("LIMIT %v OFFSET %v ", listingFilterOpts.Limit, listingFilterOpts.Offset)
	finalQuery := fmt.Sprintf(" SELECT cdwr.id as cd_workflow_runner_id,cdwr.triggered_by,cdwr.started_on,cia.*,(%s) ", totalCountQuery) + commonQuery + orderByQuery + limitOffsetQuery
	return finalQuery
}
