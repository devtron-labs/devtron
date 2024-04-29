package bean

type DtResourceObjectCatalogReqBean struct {
	*DevtronResourceObjectDescriptorBean
	ObjectData string `json:"objectData"`
}

type DtResourceObjectCreateReqBean struct {
	*DevtronResourceObjectDescriptorBean
	Overview     *ResourceOverview   `json:"overview,omitempty"`
	ParentConfig *ResourceIdentifier `json:"parentConfig,omitempty"`
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
