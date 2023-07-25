package bean

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/sql"
	"time"
)

type CiBuildType string

const (
	SELF_DOCKERFILE_BUILD_TYPE    CiBuildType = "self-dockerfile-build"
	MANAGED_DOCKERFILE_BUILD_TYPE CiBuildType = "managed-dockerfile-build"
	SKIP_BUILD_BUILD_TYPE         CiBuildType = "skip-build"
	BUILDPACK_BUILD_TYPE          CiBuildType = "buildpack-build"
)

type CiBuildConfigBean struct {
	Id                        int                `json:"id"`
	GitMaterialId             int                `json:"gitMaterialId,omitempty" validate:"required"`
	BuildContextGitMaterialId int                `json:"buildContextGitMaterialId,omitempty" validate:"required"`
	UseRootBuildContext       bool               `json:"useRootBuildContext"`
	CiBuildType               CiBuildType        `json:"ciBuildType"`
	DockerBuildConfig         *DockerBuildConfig `json:"dockerBuildConfig,omitempty"`
	BuildPackConfig           *BuildPackConfig   `json:"buildPackConfig"`
}

type DockerBuildConfig struct {
	DockerfilePath     string            `json:"dockerfileRelativePath,omitempty"`
	DockerfileContent  string            `json:"dockerfileContent"`
	Args               map[string]string `json:"args,omitempty"`
	TargetPlatform     string            `json:"targetPlatform,omitempty"`
	Language           string            `json:"language,omitempty"`
	LanguageFramework  string            `json:"languageFramework,omitempty"`
	DockerBuildOptions map[string]string `json:"dockerBuildOptions,omitempty"`
	BuildContext       string            `json:"buildContext,omitempty"`
	UseBuildx          bool              `json:"useBuildx"`
}

type BuildPackConfig struct {
	BuilderId       string            `json:"builderId"`
	Language        string            `json:"language"`
	LanguageVersion string            `json:"languageVersion"`
	BuildPacks      []string          `json:"buildPacks"`
	Args            map[string]string `json:"args"`
	ProjectPath     string            `json:"projectPath,omitempty"`
}

func ConvertBuildConfigBeanToDbEntity(templateId int, overrideTemplateId int, ciBuildConfigBean *CiBuildConfigBean, userId int32) (*pipelineConfig.CiBuildConfig, error) {
	buildMetadata := ""
	ciBuildType := ciBuildConfigBean.CiBuildType
	if ciBuildType == BUILDPACK_BUILD_TYPE {
		buildPackConfigMetadataBytes, err := json.Marshal(ciBuildConfigBean.BuildPackConfig)
		if err != nil {
			return nil, err
		}
		buildMetadata = string(buildPackConfigMetadataBytes)
	} else if ciBuildType == SELF_DOCKERFILE_BUILD_TYPE || ciBuildType == MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuildMetadataBytes, err := json.Marshal(ciBuildConfigBean.DockerBuildConfig)
		if err != nil {
			return nil, err
		}
		buildMetadata = string(dockerBuildMetadataBytes)
	}
	ciBuildConfigEntity := &pipelineConfig.CiBuildConfig{
		Id:                   ciBuildConfigBean.Id,
		Type:                 string(ciBuildType),
		CiTemplateId:         templateId,
		CiTemplateOverrideId: overrideTemplateId,
		BuildMetadata:        buildMetadata,
		AuditLog:             sql.AuditLog{UpdatedOn: time.Now(), UpdatedBy: userId},
		UseRootContext:       &ciBuildConfigBean.UseRootBuildContext,
	}
	return ciBuildConfigEntity, nil
}

func ConvertDbBuildConfigToBean(dbBuildConfig *pipelineConfig.CiBuildConfig) (*CiBuildConfigBean, error) {
	var buildPackConfig *BuildPackConfig
	var dockerBuildConfig *DockerBuildConfig
	var err error
	if dbBuildConfig == nil {
		return nil, nil
	}
	ciBuildType := CiBuildType(dbBuildConfig.Type)
	if ciBuildType == BUILDPACK_BUILD_TYPE {
		buildPackConfig, err = convertMetadataToBuildPackConfig(dbBuildConfig.BuildMetadata)
		if err != nil {
			return nil, err
		}
	} else if ciBuildType == SELF_DOCKERFILE_BUILD_TYPE || ciBuildType == MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuildConfig, err = convertMetadataToDockerBuildConfig(dbBuildConfig.BuildMetadata)
		if err != nil {
			return nil, err
		}
	}
	useRootBuildContext := false
	//dbBuildConfig.UseRootContext will be nil if the entry in db never updated before
	if dbBuildConfig.UseRootContext == nil || *(dbBuildConfig.UseRootContext) {
		useRootBuildContext = true
	}
	ciBuildConfigBean := &CiBuildConfigBean{
		Id:                  dbBuildConfig.Id,
		CiBuildType:         ciBuildType,
		BuildPackConfig:     buildPackConfig,
		DockerBuildConfig:   dockerBuildConfig,
		UseRootBuildContext: useRootBuildContext,
	}
	return ciBuildConfigBean, nil
}

func convertMetadataToBuildPackConfig(buildConfMetadata string) (*BuildPackConfig, error) {
	buildPackConfig := &BuildPackConfig{}
	err := json.Unmarshal([]byte(buildConfMetadata), buildPackConfig)
	return buildPackConfig, err
}

func convertMetadataToDockerBuildConfig(dockerBuildMetadata string) (*DockerBuildConfig, error) {
	dockerBuildConfig := &DockerBuildConfig{}
	err := json.Unmarshal([]byte(dockerBuildMetadata), dockerBuildConfig)
	return dockerBuildConfig, err
}

func OverrideCiBuildConfig(dockerfilePath string, oldArgs string, ciLevelArgs string, dockerBuildOptions string, targetPlatform string, ciBuildConfigBean *CiBuildConfigBean) (*CiBuildConfigBean, error) {
	oldDockerArgs := map[string]string{}
	ciLevelDockerArgs := map[string]string{}
	dockerBuildOptionsMap := map[string]string{}
	if oldArgs != "" {
		if err := json.Unmarshal([]byte(oldArgs), &oldDockerArgs); err != nil {
			return nil, err
		}
	}
	if ciLevelArgs != "" {
		if err := json.Unmarshal([]byte(ciLevelArgs), &ciLevelDockerArgs); err != nil {
			return nil, err
		}
	}
	if dockerBuildOptions != "" {
		if err := json.Unmarshal([]byte(dockerBuildOptions), &dockerBuildOptionsMap); err != nil {
			return nil, err
		}
	}
	//no entry found in ci_build_config table, construct with requested data
	if ciBuildConfigBean == nil {
		dockerArgs := mergeMap(oldDockerArgs, ciLevelDockerArgs)
		ciBuildConfigBean = &CiBuildConfigBean{
			CiBuildType: SELF_DOCKERFILE_BUILD_TYPE,
			DockerBuildConfig: &DockerBuildConfig{
				DockerfilePath:     dockerfilePath,
				Args:               dockerArgs,
				TargetPlatform:     targetPlatform,
				DockerBuildOptions: dockerBuildOptionsMap,
				BuildContext:       "",
			},
			//setting true as default
			UseRootBuildContext: true,
		}
	} else if ciBuildConfigBean.CiBuildType == SELF_DOCKERFILE_BUILD_TYPE || ciBuildConfigBean.CiBuildType == MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuildConfig := ciBuildConfigBean.DockerBuildConfig
		dockerArgs := mergeMap(dockerBuildConfig.Args, ciLevelDockerArgs)
		//dockerBuildConfig.DockerfilePath = dockerfilePath
		dockerBuildConfig.Args = dockerArgs
	}
	return ciBuildConfigBean, nil
}

func mergeMap(oldDockerArgs map[string]string, ciLevelDockerArgs map[string]string) map[string]string {
	dockerArgs := make(map[string]string)
	for key, value := range oldDockerArgs {
		dockerArgs[key] = value
	}
	for key, value := range ciLevelDockerArgs {
		dockerArgs[key] = value
	}
	return dockerArgs
}
