package adapter

import (
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/devtronResource/release/bean"
	"github.com/tidwall/gjson"
	"net/http"
)

func GetPatchQueryForPolicyAutoAction(autoAction, stateTo *bean.ReleaseStatusDefinitionState) ([]bean2.PatchQuery, error) {
	patchQueries := make([]bean2.PatchQuery, 0, 2) //keeping max cap as 2 since only configStatus and lock supported currently
	//policy valid, apply auto action if needed
	if autoAction.ConfigStatus != stateTo.ConfigStatus {
		configStatus, err := getReleaseConfigStatusFromPolicyConfigStatus(autoAction.ConfigStatus)
		if err != nil {
			return nil, err
		}
		patchQueries = append(patchQueries, bean2.PatchQuery{
			Path: bean2.ReleaseStatusQueryPath,
			Value: &bean2.ConfigStatus{
				Status: configStatus,
			},
		})
	}
	if autoAction.LockStatus != stateTo.LockStatus {
		lockStatus, err := getReleaseLockStatusFromPolicyLockStatus(autoAction.LockStatus)
		if err != nil {
			return nil, err
		}
		patchQueries = append(patchQueries, bean2.PatchQuery{
			Path:  bean2.ReleaseLockQueryPath,
			Value: lockStatus,
		})
	}
	return patchQueries, nil
}

func GetPolicyDefinitionStateFromReleaseObject(objectData string) (*bean.ReleaseStatusDefinitionState, error) {
	policyConfigStatus, err := getReleasePolicyConfigStatusFromResourceObjData(objectData)
	if err != nil {
		return nil, err
	}
	policyRolloutStatus, err := getReleasePolicyRolloutStatusFromResourceObjData(objectData)
	if err != nil {
		return nil, err
	}
	policyDepArtifactStatus := getReleasePolicyDepArtifactStatusFromResourceObjData(objectData)
	policyLockStatus := getReleasePolicyLockStatusFromResourceObjData(objectData)
	state := &bean.ReleaseStatusDefinitionState{
		ConfigStatus:             policyConfigStatus,
		ReleaseRolloutStatus:     policyRolloutStatus,
		DependencyArtifactStatus: policyDepArtifactStatus,
		LockStatus:               policyLockStatus,
	}
	return state, nil
}
func GetReleaseAppCountAndDepArtifactStatusFromResourceObjData(objectData string) (appCount int, depArtifactStatus bean2.DependencyArtifactStatus) {
	upstreamDep := gjson.Get(objectData, `dependencies.#(typeOfDependency=="upstream")#`) //assuming only apps are upstream of release
	upstreamDepLen := len(upstreamDep.Array())
	artifactLen := len(gjson.Get(upstreamDep.String(), `#(config.artifactConfig.artifactId>0)#`).Array())
	if artifactLen == 0 {
		depArtifactStatus = bean2.NotSelectedDependencyArtifactStatus
	} else if artifactLen < upstreamDepLen {
		depArtifactStatus = bean2.PartialSelectedDependencyArtifactStatus
	} else if artifactLen == upstreamDepLen {
		depArtifactStatus = bean2.AllSelectedDependencyArtifactStatus
	}
	return upstreamDepLen, depArtifactStatus
}

func getReleasePolicyConfigStatusFromResourceObjData(objectData string) (bean.PolicyReleaseConfigStatus, error) {
	var policyConfigStatus bean.PolicyReleaseConfigStatus
	objDataConfigStatus := gjson.Get(objectData, bean2.ReleaseResourceConfigStatusStatusPath).String()
	switch objDataConfigStatus {
	case bean2.DraftReleaseConfigStatus.ToString():
		policyConfigStatus = bean.PolicyReleaseConfigStatusDraft
	case bean2.ReadyForReleaseConfigStatus.ToString():
		policyConfigStatus = bean.PolicyConfigStatusReadyForRelease
	case bean2.HoldReleaseConfigStatus.ToString():
		policyConfigStatus = bean.PolicyReleaseConfigStatusHold
	case bean2.RescindReleaseConfigStatus.ToString():
		policyConfigStatus = bean.PolicyReleaseConfigStatusRescind
	case bean2.CorruptedReleaseConfigStatus.ToString():
		policyConfigStatus = bean.PolicyReleaseConfigStatusCorrupted
	default:
		if len(objDataConfigStatus) == 0 {
			//setting default as draft in case it is not set at creation time, needs to be check in future if this needs to be replaced with error
			policyConfigStatus = bean.PolicyReleaseConfigStatusDraft
		}
	}
	return policyConfigStatus, nil
}

func getReleasePolicyRolloutStatusFromResourceObjData(objectData string) (bean.PolicyReleaseRolloutStatus, error) {
	var policyRolloutStatus bean.PolicyReleaseRolloutStatus
	objDatRolloutStatus := gjson.Get(objectData, bean2.ReleaseResourceRolloutStatusPath).String()
	switch objDatRolloutStatus {
	case bean2.NotDeployedReleaseRolloutStatus.ToString():
		policyRolloutStatus = bean.PolicyReleaseRolloutStatusNotDeployed
	case bean2.PartiallyDeployedReleaseRolloutStatus.ToString():
		policyRolloutStatus = bean.PolicyReleaseRolloutStatusPartiallyDeployed
	case bean2.CompletelyDeployedReleaseRolloutStatus.ToString():
		policyRolloutStatus = bean.PolicyReleaseRolloutStatusCompletelyDeployed
	default:
		if len(objDatRolloutStatus) == 0 {
			//setting default as not deployed in case it is not set at creation time, needs to be check in future if this needs to be replaced with error
			policyRolloutStatus = bean.PolicyReleaseRolloutStatusNotDeployed
		}
	}
	return policyRolloutStatus, nil
}

func getReleasePolicyDepArtifactStatusFromResourceObjData(objectData string) bean.PolicyDependencyArtifactStatus {
	var policyDepArtifactStatus bean.PolicyDependencyArtifactStatus
	upstreamDep := gjson.Get(objectData, `dependencies.#(typeOfDependency=="upstream")#`)
	upstreamDepLen := len(upstreamDep.Array())
	artifactLen := len(gjson.Get(upstreamDep.String(), `#(config.artifactConfig.artifactId>0)#`).Array())
	if artifactLen == 0 {
		policyDepArtifactStatus = bean.PolicyDependencyArtifactStatusNotSelected
	} else if artifactLen < upstreamDepLen {
		policyDepArtifactStatus = bean.PolicyDependencyArtifactStatusPartialSelected
	} else if artifactLen == upstreamDepLen {
		policyDepArtifactStatus = bean.PolicyDependencyArtifactStatusAllSelected
	}
	return policyDepArtifactStatus
}

func getReleasePolicyLockStatusFromResourceObjData(objectData string) bean.PolicyLockStatus {
	var policyLockStatus bean.PolicyLockStatus
	isLocked := gjson.Get(objectData, bean2.ReleaseResourceConfigStatusIsLockedPath).Bool()
	if isLocked {
		policyLockStatus = bean.PolicyLockStatusLocked
	} else {
		policyLockStatus = bean.PolicyLockStatusUnLocked
	}
	return policyLockStatus
}

func getReleaseConfigStatusFromPolicyConfigStatus(policyConfigStatus bean.PolicyReleaseConfigStatus) (bean2.ReleaseConfigStatus, error) {
	var releaseConfigStatus bean2.ReleaseConfigStatus
	switch policyConfigStatus {
	case bean.PolicyReleaseConfigStatusDraft:
		releaseConfigStatus = bean2.DraftReleaseConfigStatus
	case bean.PolicyConfigStatusReadyForRelease:
		releaseConfigStatus = bean2.ReadyForReleaseConfigStatus
	case bean.PolicyReleaseConfigStatusHold:
		releaseConfigStatus = bean2.HoldReleaseConfigStatus
	case bean.PolicyReleaseConfigStatusRescind:
		releaseConfigStatus = bean2.RescindReleaseConfigStatus
	case bean.PolicyReleaseConfigStatusCorrupted:
		releaseConfigStatus = bean2.CorruptedReleaseConfigStatus
	default:
		return releaseConfigStatus, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean2.PatchValueNotSupportedError, bean2.PatchValueNotSupportedError)
	}
	return releaseConfigStatus, nil
}

func getReleaseLockStatusFromPolicyLockStatus(policyLockStatus bean.PolicyLockStatus) (bool, error) {
	var lockStatus bool
	switch policyLockStatus {
	case bean.PolicyLockStatusLocked:
		lockStatus = true
	case bean.PolicyLockStatusUnLocked:
		lockStatus = false
	default:
		return lockStatus, util.GetApiErrorAdapter(http.StatusBadRequest, "400", bean2.PatchValueNotSupportedError, bean2.PatchValueNotSupportedError)
	}
	return lockStatus, nil
}

func GetLastReleaseTaskRunInfo(response []bean2.DtReleaseTaskRunInfo) *bean2.DtReleaseTaskRunInfo {
	if len(response) == 0 {
		return nil
	}
	return &response[len(response)-1]
}
