package bean

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/resourceFilter"
	"github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"time"
)

type CdWorkflowWithArtifact struct {
	Id                      int                                         `json:"id"`
	CdWorkflowId            int                                         `json:"cd_workflow_id"`
	Name                    string                                      `json:"name"`
	Status                  string                                      `json:"status"`
	PodStatus               string                                      `json:"pod_status"`
	Message                 string                                      `json:"message"`
	StartedOn               time.Time                                   `json:"started_on"`
	FinishedOn              time.Time                                   `json:"finished_on"`
	PipelineId              int                                         `json:"pipeline_id"`
	Namespace               string                                      `json:"namespace"`
	LogFilePath             string                                      `json:"log_file_path"`
	TriggeredBy             int32                                       `json:"triggered_by"`
	EmailId                 string                                      `json:"email_id"`
	Image                   string                                      `json:"image"`
	MaterialInfo            string                                      `json:"material_info,omitempty"`
	DataSource              string                                      `json:"data_source,omitempty"`
	CiArtifactId            int                                         `json:"ci_artifact_id,omitempty"`
	WorkflowType            string                                      `json:"workflow_type,omitempty"`
	ExecutorType            string                                      `json:"executor_type,omitempty"`
	BlobStorageEnabled      bool                                        `json:"blobStorageEnabled"`
	UserApprovalMetadata    *pipelineConfig.UserApprovalMetadata        `json:"userApprovalMetadata"`
	GitTriggers             map[int]pipelineConfig.GitCommit            `json:"gitTriggers"`
	CiMaterials             []pipelineConfig.CiPipelineMaterialResponse `json:"ciMaterials"`
	ImageReleaseTags        []*repository.ImageTag                      `json:"imageReleaseTags"`
	ImageComment            *repository.ImageComment                    `json:"imageComment"`
	RefCdWorkflowRunnerId   int                                         `json:"referenceCdWorkflowRunnerId"`
	AppliedFilters          []*resourceFilter.FilterMetaDataBean        `json:"appliedFilters"`
	AppliedFiltersState     resourceFilter.FilterState                  `json:"appliedFiltersState"`
	AppliedFiltersTimestamp time.Time                                   `json:"appliedFiltersTimestamp"`
}
