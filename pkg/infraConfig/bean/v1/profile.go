/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

// Package v1 implements the infra config with interface values.
package v1

import "strings"

type ProfileBeanDto struct {
	ProfileBeanAbstract
	Configurations map[string][]*ConfigurationBean `json:"configurations"`
}

func (profileBean *ProfileBeanDto) GetBuildxDriverType() BuildxDriver {
	if profileBean == nil {
		return ""
	}
	return profileBean.ProfileBeanAbstract.GetBuildxDriverType()
}

func (profileBean *ProfileBeanDto) GetDescription() string {
	if profileBean == nil {
		return ""
	}
	return profileBean.ProfileBeanAbstract.GetDescription()
}

func (profileBean *ProfileBeanDto) GetName() string {
	if profileBean == nil {
		return ""
	}
	return profileBean.ProfileBeanAbstract.GetName()
}

func (profileBean *ProfileBeanDto) DeepCopy() *ProfileBeanDto {
	if profileBean == nil {
		return nil
	}
	profile := *profileBean
	configurations := make(map[string][]*ConfigurationBean)
	for key, value := range profileBean.Configurations {
		configurations[key] = value
	}
	profile.Configurations = configurations
	return &profile
}

func (profileBean *ProfileBeanDto) GetConfigurations() map[string][]*ConfigurationBean {
	configurations := make(map[string][]*ConfigurationBean)
	if profileBean == nil || len(profileBean.Configurations) == 0 {
		return configurations
	}
	return profileBean.Configurations
}

func (profileBean *ProfileBeanDto) SetPlatformConfigurations(platform string, configurations []*ConfigurationBean) *ProfileBeanDto {
	if profileBean == nil {
		return nil
	}
	if profileBean.Configurations == nil {
		profileBean.Configurations = make(map[string][]*ConfigurationBean)
	}
	profileBean.Configurations[platform] = configurations
	return profileBean
}

type ProfileBeanAbstract struct {
	Id               int          `json:"id"`
	Name             string       `json:"name" validate:"required,min=1,max=50"`
	Description      string       `json:"description" validate:"max=350"`
	BuildxDriverType BuildxDriver `json:"buildxDriverType" default:"kubernetes"`
	Active           bool         `json:"active"`
	Type             ProfileType  `json:"type"`
	AppCount         int          `json:"appCount"`
}

func (p *ProfileBeanAbstract) GetBuildxDriverType() BuildxDriver {
	if p == nil {
		return ""
	}
	if len(p.BuildxDriverType) == 0 {
		// the default driver is k8s for new profiles
		return BuildxK8sDriver
	}
	return p.BuildxDriverType
}

func (p *ProfileBeanAbstract) GetDescription() string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(p.Description)
}

func (p *ProfileBeanAbstract) GetName() string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(strings.ToLower(p.Name))
}

type ProfileType string

const (
	GLOBAL  ProfileType = "GLOBAL"
	DEFAULT ProfileType = "DEFAULT"
	NORMAL  ProfileType = "NORMAL"
)

type BuildxDriver string

func (b BuildxDriver) String() string {
	return string(b)
}

func (b BuildxDriver) IsKubernetes() bool {
	return b == BuildxK8sDriver
}

func (b BuildxDriver) IsDockerContainer() bool {
	return b == BuildxDockerContainerDriver
}

func (b BuildxDriver) IsPlatformSupported(platform string) bool {
	if b.IsKubernetes() {
		// k8s supports all platforms
		return true
	} else {
		// docker container supports only runner platform
		return platform == RUNNER_PLATFORM
	}
}

func (b BuildxDriver) IsValid() bool {
	return b.IsKubernetes() || b.IsDockerContainer()
}

const (
	// BuildxK8sDriver is the default driver for buildx
	BuildxK8sDriver BuildxDriver = "kubernetes"
	// BuildxDockerContainerDriver is the driver for docker container
	BuildxDockerContainerDriver BuildxDriver = "docker-container"
)

type ProfileResponse struct {
	Profile ProfileBeanDto `json:"profile"`
	InfraConfigMetaData
}

type InfraConfigMetaData struct {
	DefaultConfigurations map[string][]*ConfigurationBean  `json:"defaultConfigurations,omitempty"`
	ConfigurationUnits    map[ConfigKeyStr]map[string]Unit `json:"configurationUnits,omitempty"`
}
