package bean

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
