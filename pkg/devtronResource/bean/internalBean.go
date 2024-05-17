package bean

import "github.com/devtron-labs/devtron/api/bean"

type IdIdentifierIndex struct {
	Ids              []int    //all ids
	Identifiers      []string //all identifiers
	IdentifiersIndex []int    //index of dependency in all dependencies array at which this identifier is stored at, will be used to replace identifier with id
}

var DevtronResourceSupportedVersionMap = map[DevtronResourceKind]map[DevtronResourceVersion]bool{
	DevtronResourceApplication: {
		DevtronResourceVersion1: true,
	},
	DevtronResourceDevtronApplication: {
		DevtronResourceVersion1: true,
	},
	DevtronResourceHelmApplication: {
		DevtronResourceVersion1: true,
	},
	DevtronResourceCluster: {
		DevtronResourceVersion1: true,
	},
	DevtronResourceJob: {
		DevtronResourceVersion1: true,
	},
	DevtronResourceCdPipeline: {
		DevtronResourceVersion1: true,
	},
	DevtronResourceReleaseTrack: {
		DevtronResourceVersionAlpha1: true,
	},
	DevtronResourceRelease: {
		DevtronResourceVersionAlpha1: true,
	},
}

type ResourceObjectRequirementRequest struct {
	ReqBean                  *DtResourceObjectInternalBean
	ObjectDataPath           string
	SkipJsonSchemaValidation bool
}

type DtResourceObjectInternalBean struct {
	*DevtronResourceObjectDescriptorBean
	//Schema            string                           `json:"schema,omitempty"`
	ObjectData   string                           `json:"objectData"`
	Dependencies []*DevtronResourceDependencyBean `json:"dependencies"`
	//ChildDependencies []*DevtronResourceDependencyBean `json:"childDependencies"`
	Overview     *ResourceOverview   `json:"overview,omitempty"`
	ConfigStatus *ConfigStatus       `json:"configStatus,omitempty"`
	ParentConfig *ResourceIdentifier `json:"parentConfig,omitempty"`
	//PatchQuery        []PatchQuery                     `json:"query,omitempty"`
	//DependencyInfo    *DependencyInfo                  `json:"DependencyInfo,omitempty"`
}

type DtResourceObjectInternalDescriptorBean struct {
	Kind                    string `json:"kind"`
	SubKind                 string `json:"subKind,omitempty"` //TODO: extract out these fields in apiBean
	Version                 string `json:"version"`
	Id                      int    `json:"id,omitempty"`
	IdType                  IdType `json:"idType,omitempty"`
	Identifier              string `json:"identifier,omitempty"`
	DevtronResourceSchemaId int    `json:"devtronResourceSchemaId"`
	Name                    string `json:"name,omitempty"`
	ReleaseVersion          string `json:"releaseVersion,omitempty"`
}

var PatchQueryPathAuditPathMap = map[PatchQueryPath]string{
	DescriptionQueryPath:   ResourceObjectDescriptionPath,
	ReleaseStatusQueryPath: ReleaseResourceConfigStatusPath,
	ReleaseNoteQueryPath:   ReleaseResourceObjectReleaseNotePath,
	TagsQueryPath:          ResourceObjectTagsPath,
	ReleaseLockQueryPath:   ReleaseResourceConfigStatusIsLockedPath,
	NameQueryPath:          ResourceObjectNamePath,
}

type ReleaseConfigStatus string

func (s ReleaseConfigStatus) ToString() string {
	return string(s)
}

const (
	DraftReleaseConfigStatus     ReleaseConfigStatus = "draft"
	ReadyForReleaseConfigStatus  ReleaseConfigStatus = "readyForRelease"
	HoldReleaseConfigStatus      ReleaseConfigStatus = "hold"
	RescindReleaseConfigStatus   ReleaseConfigStatus = "rescind"
	CorruptedReleaseConfigStatus ReleaseConfigStatus = "corrupted"
)

type ReleaseRolloutStatus string //status of release, i.e. rollout status of the release. Not to be confused with config status

func (s ReleaseRolloutStatus) ToString() string {
	return string(s)
}

const (
	NotDeployedReleaseRolloutStatus        ReleaseRolloutStatus = "notDeployed"
	PartiallyDeployedReleaseRolloutStatus  ReleaseRolloutStatus = "partiallyDeployed"
	CompletelyDeployedReleaseRolloutStatus ReleaseRolloutStatus = "completelyDeployed"
)

func (s ReleaseRolloutStatus) IsPartiallyDeployed() bool {
	return s == PartiallyDeployedReleaseRolloutStatus
}

func (s ReleaseRolloutStatus) IsCompletelyDeployed() bool {
	return s == CompletelyDeployedReleaseRolloutStatus
}

type DependencyArtifactStatus string

func (s DependencyArtifactStatus) ToString() string {
	return string(s)
}

const (
	NotSelectedDependencyArtifactStatus     DependencyArtifactStatus = "noImageSelected"
	PartialSelectedDependencyArtifactStatus DependencyArtifactStatus = "partialImagesSelected"
	AllSelectedDependencyArtifactStatus     DependencyArtifactStatus = "allImagesSelected"
)

const (
	ReleaseConfigStatusAutoChangeComment = "Automatic status change in line with compliance."
)

const (
	ResourceSchemaMetadataPath      = "properties.overview.properties.metadata"
	ResourceObjectMetadataPath      = "overview.metadata"
	ResourceObjectOverviewPath      = "overview"
	ResourceObjectIdPath            = "overview.id"
	ResourceObjectNamePath          = "overview.name"
	ResourceObjectIdentifierPath    = "overview.identifier"
	ResourceObjectDescriptionPath   = "overview.description"
	ResourceObjectCreatedOnPath     = "overview.createdOn"
	ResourceObjectCreatedByPath     = "overview.createdBy"
	ResourceObjectTagsPath          = "overview.tags"
	ResourceObjectIdTypePath        = "overview.idType"
	ResourceObjectCreatedByIdPath   = "overview.createdBy.id"
	ResourceObjectCreatedByNamePath = "overview.createdBy.name"
	ResourceObjectCreatedByIconPath = "overview.createdBy.icon"

	ResourceObjectDependenciesPath = "dependencies"
	DependencyChildInheritanceKey  = "childInheritance"
)

// release specific keys
const (
	ReleaseResourceObjectReleaseNotePath     = "overview.releaseNote"
	ReleaseResourceObjectReleaseVersionPath  = "overview.releaseVersion"
	ReleaseResourceObjectFirstReleasedOnPath = "overview.firstReleasedOn"

	ReleaseResourceConfigStatusPath         = "status.config"
	ReleaseResourceConfigStatusStatusPath   = "status.config.status"
	ReleaseResourceConfigStatusCommentPath  = "status.config.comment"
	ReleaseResourceConfigStatusIsLockedPath = "status.config.lock"
	ReleaseResourceRolloutStatusPath        = "status.rollout.status"

	ReleaseResourceDependencyConfigImageKey              = "artifactConfig.image"
	ReleaseResourceDependencyConfigArtifactIdKey         = "artifactConfig.artifactId"
	ReleaseResourceDependencyConfigRegistryNameKey       = "artifactConfig.registryName"
	ReleaseResourceDependencyConfigRegistryTypeKey       = "artifactConfig.registryType"
	ReleaseResourceArtifactSourceAppWfIdPath             = "artifactConfig.artifactSourceAppWorkflowId"
	ReleaseResourceArtifactSourceReleaseConfigPath       = "artifactConfig.sourceReleaseConfiguration"
	ReleaseResourceDependencyConfigCiWorkflowKey         = "ciWorkflowId"
	ReleaseResourceDependencyConfigCommitSourceKey       = "commitSource"
	ReleaseResourceDependencyConfigReleaseInstructionKey = "releaseInstruction"
)

//taskRun specific keys

const (
	ResourceObjectRunSourcePath = "overview.runSource"
	ResourceTaskRunActionPath   = "action"
)

var DefaultConfigStatus = &ConfigStatus{
	Status:   DraftReleaseConfigStatus,
	IsLocked: false,
}

var DefaultRolloutStatus = NotDeployedReleaseRolloutStatus

type FilterConditionInternalBean struct {
	AppIds                    []int
	EnvIds                    []int
	StageWiseDeploymentStatus map[bean.WorkflowType][]string
	ReleaseDeploymentStatus   []string
	RequestWithoutFilters     bool
}

func NewFilterConditionInternalBean() *FilterConditionInternalBean {
	return &FilterConditionInternalBean{}

}

type IdAndSchemaIdFilter struct {
	Id                      int
	DevtronResourceSchemaId int
}

type FilterResource string

const (
	DevtronApplicationFilter             FilterResource = "application/devtron-application"
	EnvironmentFilter                    FilterResource = "environment"
	StageWiseDeploymentStatusFilter      FilterResource = "stageWiseDeploymentStatus"
	ReleaseDeploymentRolloutStatusFilter FilterResource = "releaseDeploymentRolloutStatus"
)

func (f FilterResource) ToString() string {
	return string(f)
}

const (
	RunningStatus = "Running"
	Error         = "Error"
	Missing       = "Missing"
	Unknown       = "Unknown"
	Others        = "Others"
)
