/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

// Package v0 implements the infra config with float64 values only.
//
// Deprecated: v0 is functionally broken and should not be used
// except for compatibility with legacy systems. Use v1 instead.
//
// This package is frozen and no new functionality will be added.
package v0

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
)

// Deprecated: ProfileBeanV0 is deprecated in favor of v1.ProfileBeanDto
type ProfileBeanV0 struct {
	v1.ProfileBeanAbstract
	Configurations []ConfigurationBeanV0 `json:"configurations" validate:"dive"`
}

func (profileBean *ProfileBeanV0) GetBuildxDriverType() v1.BuildxDriver {
	if profileBean == nil {
		return ""
	}
	return profileBean.ProfileBeanAbstract.GetBuildxDriverType()
}

func (profileBean *ProfileBeanV0) GetDescription() string {
	if profileBean == nil {
		return ""
	}
	return profileBean.ProfileBeanAbstract.GetDescription()
}

func (profileBean *ProfileBeanV0) GetName() string {
	if profileBean == nil {
		return ""
	}
	return profileBean.ProfileBeanAbstract.GetName()
}

// Deprecated: ProfileResponseV0 is deprecated in favor of v1.ProfileResponse
type ProfileResponseV0 struct {
	Profile ProfileBeanV0 `json:"profile"`
	InfraConfigMetaDataV0
}

// Deprecated: InfraConfigMetaDataV0 is deprecated in favor of v1.InfraConfigMetaData
type InfraConfigMetaDataV0 struct {
	DefaultConfigurations []ConfigurationBeanV0                  `json:"defaultConfigurations"`
	ConfigurationUnits    map[v1.ConfigKeyStr]map[string]v1.Unit `json:"configurationUnits"`
}
