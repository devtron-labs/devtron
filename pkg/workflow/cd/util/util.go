/*
 * Copyright (c) 2024. Devtron Inc.
 */

package util

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
)

func GetKeyForPipelineIdAndArtifact(pipelineId, artifactId int) string {
	return fmt.Sprintf("%v-%v", pipelineId, artifactId)
}

func GetCdWorkflowIdVsRunnerIdMap(dbObjects []*pipelineConfig.CdWorkflowRunner) map[int]int {
	cdWorkflowIdVsRunnerIdMap := make(map[int]int, len(dbObjects))
	for _, dbObject := range dbObjects {
		cdWorkflowIdVsRunnerIdMap[dbObject.CdWorkflowId] = dbObject.Id
	}
	return cdWorkflowIdVsRunnerIdMap
}

func GetPipelineArtifactIdKeyVsCdWorkflowIdMap(dbObjects []*pipelineConfig.CdWorkflow) map[string]int {
	pipelineArtifactIdKeyVsCdWorkflowIdMap := make(map[string]int, len(dbObjects))
	for _, dbObject := range dbObjects {
		pipelineArtifactIdKeyVsCdWorkflowIdMap[GetKeyForPipelineIdAndArtifact(dbObject.PipelineId, dbObject.CiArtifactId)] = dbObject.Id
	}
	return pipelineArtifactIdKeyVsCdWorkflowIdMap
}
