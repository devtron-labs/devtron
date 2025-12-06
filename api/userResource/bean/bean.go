package bean

import (
	bean2 "github.com/devtron-labs/devtron/pkg/appWorkflow/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/k8s/bean"
)

const (
	PathParamKind    = "kind"
	PathParamVersion = "version"
)

type PathParams struct {
	Kind    string
	Version string
}

type ResourceOptionsReqDto struct {
	EntityAccessType
	AppAndJobReqDto
	ClusterReqDto
	JobWorkflowReqDto
}
type EntityAccessType struct {
	Entity     string `json:"entity"`
	AccessType string `json:"accessType"`
}

type AppAndJobReqDto struct {
	TeamIds []int `json:"teamIds"`
}
type ClusterReqDto struct {
	*bean3.ResourceRequestBean
}
type JobWorkflowReqDto struct {
	*bean2.WorkflowNamesRequest
}
