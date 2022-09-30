package bean

type CiBuildType string

const (
	SELF_DOCKERFILE_BUILD_TYPE    CiBuildType = "self-dockerfile-build"
	MANAGED_DOCKERFILE_BUILD_TYPE CiBuildType = "managed-dockerfile-build"
	SKIP_BUILD_BUILD_TYPE         CiBuildType = "skip-build"
	BUILDPACK_BUILD_TYPE          CiBuildType = "buildpack-build"
)

type CiBuildConfig struct {
	Id                int                `json:"id"`
	GitMaterialId     int                `json:"gitMaterialId,omitempty" validate:"required"`
	CiBuildType       CiBuildType        `json:"ciBuildType"`
	DockerBuildConfig *DockerBuildConfig `json:"dockerBuildConfig,omitempty" validate:"required,dive"`
	BuildPackConfig   *BuildPackConfig   `json:"buildPackConfig"`
}

type DockerBuildConfig struct {
	DockerfilePath    string            `json:"dockerfileRelativePath,omitempty" validate:"required"`
	DockerfileContent string            `json:"DockerfileContent"`
	Args              map[string]string `json:"args,omitempty"`
	TargetPlatform    string            `json:"targetPlatform,omitempty"`
}

type BuildPackConfig struct {
	BuilderId       string            `json:"builderId"`
	Language        string            `json:"language"`
	LanguageVersion string            `json:"languageVersion"`
	BuildPacks      []string          `json:"buildPacks"`
	Args            map[string]string `json:"args"`
}
