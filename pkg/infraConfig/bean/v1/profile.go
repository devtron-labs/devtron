/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package v1 implements the infra config with interface values.
package v1

import "strings"

type ProfileBeanDto struct {
	ProfileBeanAbstract
	Configurations map[string][]*ConfigurationBean `json:"configurations"`
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
	Id          int         `json:"id"`
	Name        string      `json:"name" validate:"required,min=1,max=50,global-entity-name"`
	Description string      `json:"description" validate:"max=350"`
	Active      bool        `json:"active"`
	Type        ProfileType `json:"type"`
	AppCount    int         `json:"appCount"`
	ProfileBeanAbstractEnt
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
