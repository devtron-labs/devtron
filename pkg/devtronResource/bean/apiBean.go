package bean

type DevtronResourceObjectGetAPIBean struct {
	*DevtronResourceObjectDescriptorBean
	*DevtronResourceObjectBasicDataBean
	ChildObjects []*DevtronResourceObjectGetAPIBean `json:"childObjects,omitempty"`
}

type DevtronResourceObjectBasicDataBean struct {
	Schema        string                `json:"schema,omitempty"`
	CatalogData   string                `json:"objectData,omitempty"` //json key not changed for backward compatibility
	Overview      *ResourceOverview     `json:"overview,omitempty"`
	ReleaseStatus *ReleaseStatusApiBean `json:"configStatus,omitempty"`
	ParentConfig  *ResourceIdentifier   `json:"parentConfig,omitempty"`
}

type ReleaseStatusApiBean struct {
	Status   ReleaseStatus `json:"status"`
	Comment  string        `json:"comment,omitempty"`
	IsLocked bool          `json:"isLocked"`
}

type DtResourceObjectCatalogReqBean struct {
	*DevtronResourceObjectDescriptorBean
	ObjectData string `json:"objectData"`
}

type DtResourceObjectCreateReqBean struct {
	*DevtronResourceObjectDescriptorBean
	Overview     *ResourceOverview   `json:"overview,omitempty"`
	ParentConfig *ResourceIdentifier `json:"parentConfig,omitempty"`
}

type DtResourceObjectCloneReqBean struct {
	*DevtronResourceObjectDescriptorBean
	Overview  *ResourceOverview   `json:"overview,omitempty"`
	CloneFrom *ResourceIdentifier `json:"cloneFrom"`
}

type DtResourceObjectDependenciesReqBean struct {
	*DevtronResourceObjectDescriptorBean
	Dependencies      []*DevtronResourceDependencyBean `json:"dependencies"`
	ChildDependencies []*DevtronResourceDependencyBean `json:"childDependencies"`
}

type DtResourceObjectPatchReqBean struct {
	*DevtronResourceObjectDescriptorBean
	PatchQuery []PatchQuery `json:"query,omitempty"`
}

type DtResourceObjectOverviewDescriptorBean struct {
	*DevtronResourceObjectDescriptorBean
	*ResourceOverview
}

type ReleaseStatus string

func (s ReleaseStatus) ToString() string {
	return string(s)
}

const (
	DraftReleaseStatus                     ReleaseStatus = "draft"
	ReadyForReleaseStatus                  ReleaseStatus = "readyForRelease"
	HoldReleaseStatus                      ReleaseStatus = "hold"
	RescindReleaseStatus                   ReleaseStatus = "rescind"
	CorruptedReleaseStatus                 ReleaseStatus = "corrupted"
	PartiallyReleasedReleaseStatus         ReleaseStatus = "partiallyReleased"
	CompletelyReleasedReleaseRolloutStatus ReleaseStatus = "completelyReleased"
)

type DevtronResourceUIComponent string

func (d DevtronResourceUIComponent) ToString() string {
	return string(d)
}

const (
	UIComponentAll          DevtronResourceUIComponent = "*"
	UIComponentCatalog      DevtronResourceUIComponent = "catalog"
	UIComponentOverview     DevtronResourceUIComponent = "overview"
	UIComponentConfigStatus DevtronResourceUIComponent = "configStatus"
	UIComponentNote         DevtronResourceUIComponent = "note"
)

type PatchQueryPath string
type PatchQueryOperation string

func (n PatchQueryPath) ToString() string {
	return string(n)
}

func (n PatchQueryOperation) ToString() string {
	return string(n)
}

const (
	Replace PatchQueryOperation = "replace"
	Add     PatchQueryOperation = "add"
	Remove  PatchQueryOperation = "remove"
)

const (
	// common query paths
	DescriptionQueryPath PatchQueryPath = "description"
	NameQueryPath        PatchQueryPath = "name"
	TagsQueryPath        PatchQueryPath = "tags"
	CatalogQueryPath     PatchQueryPath = "catalog"

	//release specific query paths
	CommitQueryPath                PatchQueryPath = "commit"
	ReleaseNoteQueryPath           PatchQueryPath = "note"
	ReleaseStatusQueryPath         PatchQueryPath = "status"
	ReleaseLockQueryPath           PatchQueryPath = "lock"
	ReleaseDepInstructionQueryPath PatchQueryPath = "releaseInstruction"
	ReleaseDepConfigImageQueryPath PatchQueryPath = "image"
	ReleaseDepApplicationQueryPath PatchQueryPath = "application"
)

type FilterKeyObject = string

type SearchPropertyBy string

const (
	ArtifactTag SearchPropertyBy = "artifactTag"
	ImageTag    SearchPropertyBy = "imageTag"
)

type DtReleaseTaskRunInfo struct {
	Level          int                      `json:"level"`
	TaskRunAllowed *bool                    `json:"taskRunAllowed,omitempty"`
	Dependencies   []*CdPipelineReleaseInfo `json:"dependencies,omitempty"`
}

func (res DtReleaseTaskRunInfo) IsTaskRunAllowed() bool {
	if res.TaskRunAllowed == nil {
		return false
	}
	return *res.TaskRunAllowed
}

const (
	ReleaseLockStatusChangeSuccessMessage          = "Requirement is locked."
	ReleaseUnLockStatusChangeSuccessMessage        = "Requirement is unlocked."
	ReleaseHoldStatusChangeSuccessDetailMessage    = "No deployments can be triggered in 'On Hold' state"
	ReleaseRescindStatusChangeSuccessDetailMessage = "This release is no longer usable."

	ReleaseStatusPatchErrMessage                           = "Cannot change status"
	ReleaseStatusReadyForReleaseNoAppErrMessage            = "Please add applications and images first."
	ReleaseStatusReadyForReleaseNoOrPartialImageErrMessage = "To mark it ready for release, all apps should have respective images added."
	ReleaseStatusHoldOrRescindPatchNoCommentErrMessage     = "Comment is required for updating this status."
)
