package adapter

import (
	bean2 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/devtronResource/release/bean"
	"github.com/tidwall/gjson"
	"k8s.io/utils/pointer"
)

func GetPolicyDefinitionStateFromReleaseObject(objectData string) *bean.ReleaseStatusDefinitionState {
	state := &bean.ReleaseStatusDefinitionState{
		ConfigStatus:  bean2.Status(gjson.Get(objectData, bean2.ResourceConfigStatusStatusPath).String()),
		ReleaseStatus: bean2.ReleaseStatus(gjson.Get(objectData, bean2.ResourceReleaseRolloutStatusPath).String()),
		LockStatus:    pointer.Bool(gjson.Get(objectData, bean2.ResourceConfigStatusIsLockedPath).Bool()),
	}
	//TODO: with dependency object get, after db dep get object creation
	upstreamDep := gjson.Get(objectData, `dependencies.#(typeOfDependency=="upstream")#`)
	upstreamDepLen := len(upstreamDep.Array())
	artifactLen := len(gjson.Get(upstreamDep.String(), `config.artifactConfig.#(artifactId>0)#`).Array())
	var depArtifactState bean2.DependencyArtifactStatus
	if artifactLen == 0 {
		depArtifactState = bean2.NotSelectedDependencyArtifactStatus
	} else if artifactLen < upstreamDepLen {
		depArtifactState = bean2.PartialSelectedDependencyArtifactStatus
	} else if artifactLen == upstreamDepLen {
		depArtifactState = bean2.AllSelectedDependencyArtifactStatus
	}
	state.DependencyArtifactStatus = depArtifactState
	return state
}
