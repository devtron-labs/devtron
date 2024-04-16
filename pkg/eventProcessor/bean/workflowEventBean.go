package bean

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	bean3 "github.com/devtron-labs/devtron/pkg/pipeline/bean"
	"github.com/devtron-labs/devtron/util"
	eventUtil "github.com/devtron-labs/devtron/util/event"
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

type ImageDetailsFromCR struct {
	ImageDetails []types.ImageDetail `json:"imageDetails"`
	Region       string              `json:"region"`
}

type ImageScanningEvent struct {
	CiPipelineId int                    `json:"ciPipelineId"`
	CdPipelineId int                    `json:"cdPipelineId"`
	TriggerBy    int                    `json:"triggeredBy" validate:"required"`
	Image        string                 `json:"image" validate:"required"`
	Digest       string                 `json:"digest" validate:"required"`
	PipelineType eventUtil.PipelineType `json:"PipelineType" validate:"oneof=CI PRE-CD POST-CD"`
}

type CiCompleteEvent struct {
	CiProjectDetails              []bean3.CiProjectDetails `json:"ciProjectDetails"`
	DockerImage                   string                   `json:"dockerImage" validate:"required,image-validator"`
	Digest                        string                   `json:"digest"`
	PipelineId                    int                      `json:"pipelineId"`
	WorkflowId                    *int                     `json:"workflowId"`
	TriggeredBy                   int32                    `json:"triggeredBy"`
	PipelineName                  string                   `json:"pipelineName"`
	DataSource                    string                   `json:"dataSource"`
	MaterialType                  string                   `json:"materialType"`
	Metrics                       util.CIMetrics           `json:"metrics"`
	AppName                       string                   `json:"appName"`
	IsArtifactUploaded            bool                     `json:"isArtifactUploaded"`
	FailureReason                 string                   `json:"failureReason"`
	ImageDetailsFromCR            *ImageDetailsFromCR      `json:"imageDetailsFromCR"`
	PluginRegistryArtifactDetails map[string][]string      `json:"PluginRegistryArtifactDetails"`
	PluginArtifactStage           string                   `json:"pluginArtifactStage"`
	DockerRegistryId              string                   `json:"dockerRegistryId"`
}

type DevtronAppReleaseContextType struct {
	CancelContext context.CancelFunc
	RunnerId      int
}
