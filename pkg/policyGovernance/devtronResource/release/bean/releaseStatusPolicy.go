/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
)

type ReleaseActionCheckPolicy struct {
	Definitions []*ReleaseActionCheckPolicyDefinition `json:"definitions"`
	//not keeping selector as this will be valid on all releases
	Consequence bean2.ConsequenceAction `json:"consequence"` //always block currently
}

type ReleaseActionCheckPolicyDefinition struct {
	OperationType      PolicyReleaseOperationType      `json:"operationType"`
	OperationPaths     []PolicyReleaseOperationPath    `json:"operationPaths"`
	PossibleFromStates []*ReleaseStatusDefinitionState `json:"possibleFromStates"`
}

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
	ConfigStatus             PolicyReleaseConfigStatus      `json:"configStatus"`
	ReleaseRolloutStatus     PolicyReleaseRolloutStatus     `json:"rolloutStatus"`
	DependencyArtifactStatus PolicyDependencyArtifactStatus `json:"dependencyArtifactStatus"`
	LockStatus               PolicyLockStatus               `json:"lockStatus"`
}

type PolicyReleaseOperationType string

const (
	PolicyReleaseOperationTypePatch             PolicyReleaseOperationType = "patch"
	PolicyReleaseOperationTypeDeploymentTrigger PolicyReleaseOperationType = "deploymentTrigger"
	PolicyReleaseOperationTypeDelete            PolicyReleaseOperationType = "delete"
)

type PolicyReleaseConfigStatus string

const (
	PolicyReleaseConfigStatusDraft     PolicyReleaseConfigStatus = "draft"
	PolicyConfigStatusReadyForRelease  PolicyReleaseConfigStatus = "readyForRelease"
	PolicyReleaseConfigStatusHold      PolicyReleaseConfigStatus = "hold"
	PolicyReleaseConfigStatusRescind   PolicyReleaseConfigStatus = "rescind"
	PolicyReleaseConfigStatusCorrupted PolicyReleaseConfigStatus = "corrupted"
	PolicyReleaseConfigStatusAny       PolicyReleaseConfigStatus = "*"
)

type PolicyReleaseRolloutStatus string

const (
	PolicyReleaseRolloutStatusNotDeployed        PolicyReleaseRolloutStatus = "notDeployed"
	PolicyReleaseRolloutStatusPartiallyDeployed  PolicyReleaseRolloutStatus = "partiallyDeployed"
	PolicyReleaseRolloutStatusCompletelyDeployed PolicyReleaseRolloutStatus = "completelyDeployed"
	PolicyReleaseRolloutStatusAny                PolicyReleaseRolloutStatus = "*"
)

type PolicyDependencyArtifactStatus string

const (
	PolicyDependencyArtifactStatusNotSelected     PolicyDependencyArtifactStatus = "noImageSelected"
	PolicyDependencyArtifactStatusPartialSelected PolicyDependencyArtifactStatus = "partialImagesSelected"
	PolicyDependencyArtifactStatusAllSelected     PolicyDependencyArtifactStatus = "allImagesSelected"
	PolicyDependencyArtifactStatusAny             PolicyDependencyArtifactStatus = "*"
)

type PolicyLockStatus string

const (
	PolicyLockStatusLocked   PolicyLockStatus = "locked"
	PolicyLockStatusUnLocked PolicyLockStatus = "unLocked"
	PolicyLockStatusAny      PolicyLockStatus = "*"
)

type PolicyReleaseOperationPath string

const (
	ReleasePolicyOpPathDescription              PolicyReleaseOperationPath = bean.ResourceObjectDescriptionPath
	ReleasePolicyOpPathName                     PolicyReleaseOperationPath = bean.ResourceObjectNamePath
	ReleasePolicyOpPathTags                     PolicyReleaseOperationPath = bean.ResourceObjectTagsPath
	ReleasePolicyOpPathCatalog                  PolicyReleaseOperationPath = bean.ResourceObjectMetadataPath
	ReleasePolicyOpPathReleaseNote              PolicyReleaseOperationPath = bean.ReleaseResourceObjectReleaseNotePath
	ReleasePolicyOpPathStatus                   PolicyReleaseOperationPath = bean.ReleaseResourceConfigStatusPath
	ReleasePolicyOpPathLock                     PolicyReleaseOperationPath = bean.ReleaseResourceConfigStatusIsLockedPath
	ReleasePolicyOpPathDependencyApp            PolicyReleaseOperationPath = "dependency.applications"
	ReleasePolicyOpPathDependencyAppImage       PolicyReleaseOperationPath = "dependency.applications.image"
	ReleasePolicyOpPathDependencyAppInstruction PolicyReleaseOperationPath = "dependency.applications.instruction"
)

var PatchQueryPathReleasePolicyOpPathMap = map[bean.PatchQueryPath]PolicyReleaseOperationPath{
	bean.DescriptionQueryPath:           ReleasePolicyOpPathDescription,
	bean.NameQueryPath:                  ReleasePolicyOpPathName,
	bean.TagsQueryPath:                  ReleasePolicyOpPathTags,
	bean.CatalogQueryPath:               ReleasePolicyOpPathCatalog,
	bean.ReleaseNoteQueryPath:           ReleasePolicyOpPathReleaseNote,
	bean.ReleaseStatusQueryPath:         ReleasePolicyOpPathStatus,
	bean.ReleaseLockQueryPath:           ReleasePolicyOpPathLock,
	bean.ReleaseDepInstructionQueryPath: ReleasePolicyOpPathDependencyAppInstruction,
	bean.ReleaseDepConfigImageQueryPath: ReleasePolicyOpPathDependencyAppImage,
	bean.ReleaseDepApplicationQueryPath: ReleasePolicyOpPathDependencyApp,
}
