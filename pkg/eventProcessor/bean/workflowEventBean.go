package bean

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"time"
)

type CdStageCompleteEvent struct {
	CiProjectDetails              []bean3.CiProjectDetails     `json:"ciProjectDetails"`
	WorkflowId                    int                          `json:"workflowId"`
	WorkflowRunnerId              int                          `json:"workflowRunnerId"`
	CdPipelineId                  int                          `json:"cdPipelineId"`
	TriggeredBy                   int32                        `json:"triggeredBy"`
	StageYaml                     string                       `json:"stageYaml"`
	ArtifactLocation              string                       `json:"artifactLocation"`
	PipelineName                  string                       `json:"pipelineName"`
	CiArtifactDTO                 pipelineConfig.CiArtifactDTO `json:"ciArtifactDTO"`
	PluginRegistryArtifactDetails map[string][]string          `json:"PluginRegistryArtifactDetails"`
}

type AsyncCdDeployEvent struct {
	ValuesOverrideRequest *bean.ValuesOverrideRequest `json:"valuesOverrideRequest"`
	TriggeredAt           time.Time                   `json:"triggeredAt"`
	TriggeredBy           int32                       `json:"triggeredBy"`
}
