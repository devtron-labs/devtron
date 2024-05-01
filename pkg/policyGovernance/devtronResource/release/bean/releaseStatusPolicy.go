package bean

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
)

type ReleaseStatusPolicy struct {
	Definitions []*ReleaseStatusPolicyDefinition `json:"definitions"`
	//not keeping selector as this will be valid on all releases
	Consequence bean2.ConsequenceAction `json:"consequence"` //always block currently
}

type ReleaseStatusPolicyDefinition struct {
	StateTo            *ReleaseStatusDefinitionState   `json:"stateTo"`
	PossibleFromStates []*ReleaseStatusDefinitionState `json:"possibleFromStates"`
	AutoAction         *ReleaseStatusDefinitionState   `json:"autoAction"` //if to and from matches, then automatically change state to this configuration
}

type ReleaseStatusDefinitionState struct {
	ConfigStatus             bean.Status                   `json:"configStatus"`
	ReleaseStatus            bean.ReleaseStatus            `json:"releaseStatus"`
	DependencyArtifactStatus bean.DependencyArtifactStatus `json:"dependencyArtifactStatus"`
	LockStatus               *bool                         `json:"lockStatus"` //setting as pointer because at places in policy it might be possible that this field is not present, in that case to differ pointer will is used
}
