package CiPipeline

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
const PIPELINE_TYPE_IS_NOT_VALID = "PipelineType is not valid"

type PipelineType string

const (
	CI_BUILD PipelineType = "CI_BUILD"
	LINKED   PipelineType = "LINKED"
	// CI_EXTERNAL field is been sent from the dashboard in CreateLinkedCI request and directly gets saved to Database without any validations
	CI_EXTERNAL PipelineType = "CI_EXTERNAL" // Deprecated Enum: TODO fix the PipelineTypes in code and database
	EXTERNAL    PipelineType = "EXTERNAL"
	CI_JOB      PipelineType = "CI_JOB"
	LINKED_CD   PipelineType = "LINKED_CD"
	NORMAL_JOB  PipelineType = "NORMAL_JOB"
)

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

func (pType PipelineType) IsValidPipelineType() bool {
	switch pType {
	case CI_BUILD, LINKED, EXTERNAL, CI_JOB, LINKED_CD, NORMAL_JOB:
		return true
	default:
		return false
	}
}
