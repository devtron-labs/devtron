package bulkAction

type NameIncludesExcludes struct {
	Names []string `json:"names"`
}

type DeploymentTemplateSpec struct {
	PatchJson string `json:"patchJson"`
}
type DeploymentTemplateTask struct {
	Spec *DeploymentTemplateSpec `json:"spec"`
}
type CmAndSecretSpec struct {
	Names     []string `json:"names"`
	PatchJson string   `json:"patchJson"`
}
type CmAndSecretTask struct {
	Spec *CmAndSecretSpec `json:"spec"`
}
type BulkUpdatePayload struct {
	Includes           *NameIncludesExcludes   `json:"includes"`
	Excludes           *NameIncludesExcludes   `json:"excludes"`
	EnvIds             []int                   `json:"envIds"`
	Global             bool                    `json:"global"`
	DeploymentTemplate *DeploymentTemplateTask `json:"deploymentTemplate"`
	ConfigMap          *CmAndSecretTask        `json:"configMap"`
	Secret             *CmAndSecretTask        `json:"secret"`
}
type BulkUpdateScript struct {
	ApiVersion string             `json:"apiVersion" validate:"required"`
	Kind       string             `json:"kind" validate:"required"`
	Spec       *BulkUpdatePayload `json:"spec" validate:"required"`
}
type BulkUpdateSeeExampleResponse struct {
	Operation string            `json:"operation"`
	Script    *BulkUpdateScript `json:"script" validate:"required"`
	ReadMe    string            `json:"readme"`
}
type ImpactedObjectsResponse struct {
	DeploymentTemplate []*DeploymentTemplateImpactedObjectsResponseForOneApp `json:"deploymentTemplate"`
	ConfigMap          []*CmAndSecretImpactedObjectsResponseForOneApp        `json:"configMap"`
	Secret             []*CmAndSecretImpactedObjectsResponseForOneApp        `json:"secret"`
}
type DeploymentTemplateImpactedObjectsResponseForOneApp struct {
	AppId   int    `json:"appId"`
	AppName string `json:"appName"`
	EnvId   int    `json:"envId"`
}
type CmAndSecretImpactedObjectsResponseForOneApp struct {
	AppId   int      `json:"appId"`
	AppName string   `json:"appName"`
	EnvId   int      `json:"envId"`
	Names   []string `json:"names"`
}
type DeploymentTemplateBulkUpdateResponseForOneApp struct {
	AppId   int    `json:"appId"`
	AppName string `json:"appName"`
	EnvId   int    `json:"envId"`
	Message string `json:"message"`
}
type CmAndSecretBulkUpdateResponseForOneApp struct {
	AppId   int      `json:"appId"`
	AppName string   `json:"appName"`
	EnvId   int      `json:"envId"`
	Names   []string `json:"names"`
	Message string   `json:"message"`
}
type BulkUpdateResponse struct {
	DeploymentTemplate *DeploymentTemplateBulkUpdateResponse `json:"deploymentTemplate"`
	ConfigMap          *CmAndSecretBulkUpdateResponse        `json:"configMap"`
	Secret             *CmAndSecretBulkUpdateResponse        `json:"secret"`
}
type DeploymentTemplateBulkUpdateResponse struct {
	Message    []string                                         `json:"message"`
	Failure    []*DeploymentTemplateBulkUpdateResponseForOneApp `json:"failure"`
	Successful []*DeploymentTemplateBulkUpdateResponseForOneApp `json:"successful"`
}
type CmAndSecretBulkUpdateResponse struct {
	Message    []string                                  `json:"message"`
	Failure    []*CmAndSecretBulkUpdateResponseForOneApp `json:"failure"`
	Successful []*CmAndSecretBulkUpdateResponseForOneApp `json:"successful"`
}

type BulkApplicationForEnvironmentPayload struct {
	AppIdIncludes   []int `json:"appIdIncludes,omitempty"`
	AppIdExcludes   []int `json:"appIdExcludes,omitempty"`
	EnvId           int   `json:"envId"`
	UserId          int32 `json:"-"`
	InvalidateCache bool  `json:"invalidateCache"`
}

type BulkApplicationForEnvironmentResponse struct {
	BulkApplicationForEnvironmentPayload
	Response map[string]map[string]bool `json:"response"`
}

type CdBulkAction int

const (
	CD_BULK_DELETE CdBulkAction = iota
)

type CdBulkActionRequestDto struct {
	Action                CdBulkAction `json:"action"`
	EnvIds                []int        `json:"envIds"`
	EnvNames              []string     `json:"envNames"`
	AppIds                []int        `json:"appIds"`
	AppNames              []string     `json:"appNames"`
	ProjectIds            []int        `json:"projectIds"`
	ProjectNames          []string     `json:"projectNames"`
	DeleteWfAndCiPipeline bool         `json:"deleteWfAndCiPipeline"`
	ForceDelete           bool         `json:"forceDelete"`
	NonCascadeDelete      bool         `json:"nonCascadeDelete"`
	UserId                int32        `json:"-"`
}

type CdBulkActionResponseDto struct {
	PipelineName    string `json:"pipelineName"`
	AppName         string `json:"appName"`
	EnvironmentName string `json:"environmentName"`
	DeletionResult  string `json:"deletionResult,omitempty"`
}

type CiBulkActionResponseDto struct {
	PipelineName   string `json:"pipelineName"`
	DeletionResult string `json:"deletionResult,omitempty"`
}
type WfBulkActionResponseDto struct {
	WorkflowId     int    `json:"workflowId"`
	DeletionResult string `json:"deletionResult,omitempty"`
}

type PipelineAndWfBulkActionResponseDto struct {
	CdPipelinesRespDtos []*CdBulkActionResponseDto `json:"cdPipelines"`
	CiPipelineRespDtos  []*CiBulkActionResponseDto `json:"ciPipelines"`
	AppWfRespDtos       []*WfBulkActionResponseDto `json:"appWorkflows"`
}
