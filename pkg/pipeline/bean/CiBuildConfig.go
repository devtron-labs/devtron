package bean

type CiBuildType string

const (
	SELF_DOCKERFILE_BUILD_TYPE    CiBuildType = "self-dockerfile-build"
	MANAGED_DOCKERFILE_BUILD_TYPE CiBuildType = "managed-dockerfile-build"
	SKIP_BUILD_TYPE               CiBuildType = "skip-build"
	BUILDPACK_BUILD_TYPE          CiBuildType = "buildpack-build"
)
const Main = "main"
const UniquePlaceHolderForAppName = "$etron"

const PIPELINE_NAME_ALREADY_EXISTS_ERROR = "pipeline name already exist"

type CiBuildConfigBean struct {
	Id                        int                `json:"id"`
	GitMaterialId             int                `json:"gitMaterialId,omitempty" validate:"required"`
	BuildContextGitMaterialId int                `json:"buildContextGitMaterialId,omitempty" validate:"required"`
	UseRootBuildContext       bool               `json:"useRootBuildContext"`
	CiBuildType               CiBuildType        `json:"ciBuildType"`
	DockerBuildConfig         *DockerBuildConfig `json:"dockerBuildConfig,omitempty"`
	BuildPackConfig           *BuildPackConfig   `json:"buildPackConfig"`
	PipelineType              string             `json:"pipelineType"`
}

type DockerBuildConfig struct {
	DockerfilePath         string              `json:"dockerfileRelativePath,omitempty"`
	DockerfileContent      string              `json:"dockerfileContent"`
	Args                   map[string]string   `json:"args,omitempty"`
	TargetPlatform         string              `json:"targetPlatform,omitempty"`
	Language               string              `json:"language,omitempty"`
	LanguageFramework      string              `json:"languageFramework,omitempty"`
	DockerBuildOptions     map[string]string   `json:"dockerBuildOptions,omitempty"`
	BuildContext           string              `json:"buildContext,omitempty"`
	UseBuildx              bool                `json:"useBuildx"`
	BuildxProvenanceMode   string              `json:"buildxProvenanceMode"`
	BuildxK8sDriverOptions []map[string]string `json:"buildxK8SDriverOptions,omitempty"`
}

type BuildPackConfig struct {
	BuilderId       string            `json:"builderId"`
	Language        string            `json:"language"`
	LanguageVersion string            `json:"languageVersion"`
	BuildPacks      []string          `json:"buildPacks"`
	Args            map[string]string `json:"args"`
	ProjectPath     string            `json:"projectPath,omitempty"`
}
