package bean

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
)

type CiBuildType string

const (
	SELF_DOCKERFILE_BUILD_TYPE    CiBuildType = "self-dockerfile-build"
	MANAGED_DOCKERFILE_BUILD_TYPE CiBuildType = "managed-dockerfile-build"
	SKIP_BUILD_BUILD_TYPE         CiBuildType = "skip-build"
	BUILDPACK_BUILD_TYPE          CiBuildType = "buildpack-build"
)

type CiBuildConfigBean struct {
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

func ConvertBuildConfigBeanToDbEntity(templateId int, overrideTemplateId int, ciBuildConfigBean *CiBuildConfigBean) (*pipelineConfig.CiBuildConfig, error) {
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
	ciBuildConfig := &pipelineConfig.CiBuildConfig{
		Type:                 string(ciBuildType),
		CiTemplateId:         templateId,
		CiTemplateOverrideId: overrideTemplateId,
		BuildMetadata:        buildMetadata,
	}
	return ciBuildConfig, nil
}

func ConvertDbBuildConfigToBean(dbConfig *pipelineConfig.CiBuildConfig) (*CiBuildConfigBean, error) {
	var buildPackConfig *BuildPackConfig
	var dockerBuildConfig *DockerBuildConfig
	var err error
	if dbConfig == nil {
		return nil, nil
	}
	ciBuildType := CiBuildType(dbConfig.Type)
	if ciBuildType == BUILDPACK_BUILD_TYPE {
		buildPackConfig, err = convertMetadataToBuildPackConfig(dbConfig.BuildMetadata)
		if err != nil {
			return nil, err
		}
	} else if ciBuildType == SELF_DOCKERFILE_BUILD_TYPE || ciBuildType == MANAGED_DOCKERFILE_BUILD_TYPE {
		dockerBuildConfig, err = convertMetadataToDockerBuildConfig(dbConfig.BuildMetadata)
		if err != nil {
			return nil, err
		}
	}
	ciBuildConfigBean := &CiBuildConfigBean{
		Id:                dbConfig.Id,
		CiBuildType:       ciBuildType,
		BuildPackConfig:   buildPackConfig,
		DockerBuildConfig: dockerBuildConfig,
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

func OverrideCiBuildConfig(dockerfilePath string, args string, targetPlatform string, ciBuildConfigBean *CiBuildConfigBean) (*CiBuildConfigBean, error) {
	dockerArgs := map[string]string{}
	if args != "" {
		if err := json.Unmarshal([]byte(args), &dockerArgs); err != nil {
			return nil, err
		}
	}
	if ciBuildConfigBean == nil {
		ciBuildConfigBean = &CiBuildConfigBean{
			CiBuildType: SELF_DOCKERFILE_BUILD_TYPE,
			DockerBuildConfig: &DockerBuildConfig{
				DockerfilePath: dockerfilePath,
				Args:           dockerArgs,
				TargetPlatform: targetPlatform,
			},
		}
	} else if ciBuildConfigBean.CiBuildType == SELF_DOCKERFILE_BUILD_TYPE {
		dockerBuildConfig := ciBuildConfigBean.DockerBuildConfig
		dockerBuildConfig.DockerfilePath = dockerfilePath
		dockerBuildConfig.Args = dockerArgs
	}
	return ciBuildConfigBean, nil
}
