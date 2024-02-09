package bean

import (
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	"time"
)

const (
	RefreshTypeNormal    = "normal"
	TargetRevisionMaster = "master"
	PatchTypeMerge       = "merge"
)

type ArgoCdAppPatchReqDto struct {
	ArgoAppName    string
	ChartLocation  string
	GitRepoUrl     string
	TargetRevision string
	PatchType      string
}

const (
	Degraded    = "Degraded"
	Healthy     = "Healthy"
	Progressing = "Progressing"
	Suspended   = "Suspended"
	TimeoutFast = 10 * time.Second
	TimeoutSlow = 30 * time.Second
	TimeoutLazy = 60 * time.Second
	HIBERNATING = "HIBERNATING"
	SUCCEEDED   = "Succeeded"
)

type Result struct {
	Response *application.ApplicationResourceResponse
	Error    error
	Request  *application.ApplicationResourceRequest
}

type ResourceTreeResponse struct {
	*v1alpha1.ApplicationTree
	NewGenerationReplicaSets []string                        `json:"newGenerationReplicaSets"`
	Status                   string                          `json:"status"`
	RevisionHash             string                          `json:"revisionHash"`
	PodMetadata              []*PodMetadata                  `json:"podMetadata"`
	Conditions               []v1alpha1.ApplicationCondition `json:"conditions"`
}

type PodMetadata struct {
	Name           string    `json:"name"`
	UID            string    `json:"uid"`
	Containers     []*string `json:"containers"`
	InitContainers []*string `json:"initContainers"`
	IsNew          bool      `json:"isNew"`
	// EphemeralContainers are set for Pod kind manifest response only
	// will always contain running ephemeral containers
	// +optional
	EphemeralContainers []*k8sObjectsUtil.EphemeralContainerData `json:"ephemeralContainers"`
}

type ErrUnauthorized struct {
	message string
}

func NewErrUnauthorized(message string) *ErrUnauthorized {
	return &ErrUnauthorized{
		message: message,
	}
}

func (e *ErrUnauthorized) Error() string {
	return e.message
}
