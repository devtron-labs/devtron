package bean

type TaskType string

const (
	TaskTypePreDeployment  TaskType = "pre-deployment"
	TaskTypePostDeployment TaskType = "post-deployment"
	TaskTypeDeployment     TaskType = "deployment"
)

var CdPipelineAllDeploymentTaskRuns = []TaskType{TaskTypeDeployment, TaskTypePreDeployment, TaskTypePostDeployment}

type TaskRunSource struct {
	Kind             DevtronResourceKind    `json:"kind"`
	Version          DevtronResourceVersion `json:"version"`
	Id               int                    `json:"id"`
	Identifier       string                 `json:"identifier"`
	ReleaseVersion   string                 `json:"releaseVersion"`
	Name             string                 `json:"name"`
	ReleaseTrackName string                 `json:"releaseTrackName"`
}

type TaskRunIdentifier struct {
	Id                 int
	IdType             IdType
	DtResourceId       int
	DtResourceSchemaId int
}
