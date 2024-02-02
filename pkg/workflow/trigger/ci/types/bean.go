package types

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/workflow/pipeline/ci/materials/types"
	"time"
)

type RefPluginName = string

const COPY_CONTAINER_IMAGE RefPluginName = "Copy container image"

type CiTriggerRequest struct {
	PipelineId                int
	CommitHashes              map[int]pipelineConfig.GitCommit
	CiMaterials               []*types.CiPipelineMaterialModel
	TriggeredBy               int32
	InvalidateCache           bool
	ExtraEnvironmentVariables map[string]string // extra env variables which will be used for CI
	EnvironmentId             int
	PipelineType              string
	CiArtifactLastFetch       time.Time
	ReferenceCiWorkflowId     int
}

func (obj CiTriggerRequest) BuildTriggerObject(refCiWorkflow *pipelineConfig.CiWorkflow,
	ciMaterials []*types.CiPipelineMaterialModel, triggeredBy int32,
	invalidateCache bool, extraEnvironmentVariables map[string]string,
	pipelineType string) {

	obj.PipelineId = refCiWorkflow.CiPipelineId
	obj.CommitHashes = refCiWorkflow.GitTriggers
	obj.CiMaterials = ciMaterials
	obj.TriggeredBy = triggeredBy
	obj.InvalidateCache = invalidateCache
	obj.EnvironmentId = refCiWorkflow.EnvironmentId
	obj.ReferenceCiWorkflowId = refCiWorkflow.Id
	obj.InvalidateCache = invalidateCache
	obj.ExtraEnvironmentVariables = extraEnvironmentVariables
	obj.PipelineType = pipelineType
}

type PipelineTriggerMetadata struct {
	AppId       int
	AppName     string
	AppType     helper.AppType
	EnvId       int
	EnvName     string
	ClusterId   int
	ClusterName string
	Namespace   string
	AppLabels   map[string]string
}
