package bean

import util5 "github.com/devtron-labs/common-lib/utils/k8s"

type PodRotateRequest struct {
	AppId               int                        `json:"appId" validate:"required"`
	EnvironmentId       int                        `json:"environmentId" validate:"required"`
	ResourceIdentifiers []util5.ResourceIdentifier `json:"resources" validate:"required"`
	UserId              int32                      `json:"-"`
}

type RequestType string

const (
	START RequestType = "START"
	STOP  RequestType = "STOP"
)

type StopAppRequest struct {
	AppId         int         `json:"appId" validate:"required"`
	EnvironmentId int         `json:"environmentId" validate:"required"`
	UserId        int32       `json:"userId"`
	RequestType   RequestType `json:"requestType" validate:"oneof=START STOP"`
	// ReferenceId is a unique identifier for the workflow runner
	// refer pipelineConfig.CdWorkflowRunner
	ReferenceId *string
}
