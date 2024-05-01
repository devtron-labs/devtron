package bean

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

var PatchQueryPathAuditPathMap = map[PatchQueryPath]string{
	DescriptionQueryPath: ResourceObjectDescriptionPath,
	StatusQueryPath:      ResourceConfigStatusPath,
	NoteQueryPath:        ResourceObjectReleaseNotePath,
	TagsQueryPath:        ResourceObjectTagsPath,
	LockQueryPath:        ResourceConfigStatusIsLockedPath,
	NameQueryPath:        ResourceObjectNamePath,
}
