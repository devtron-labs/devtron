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

package adapter

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v0"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	unitsBean "github.com/devtron-labs/devtron/pkg/infraConfig/units/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"slices"
)

func GetInfraProfileEntity(configurationBean *v1.ConfigurationBean, valueString, platform string, userId int32) *repository.InfraProfileConfigurationEntity {
	infraProfile := &repository.InfraProfileConfigurationEntity{
		Id:          configurationBean.Id,
		Key:         util.GetConfigKey(configurationBean.Key),
		ValueString: valueString,
		Unit:        util.GetUnitSuffix(configurationBean.Key, configurationBean.Unit),
		ProfileId:   configurationBean.ProfileId, // maintained for backward compatibility
		UniqueId:    repository.GetUniqueId(configurationBean.ProfileId, platform),
		Active:      configurationBean.Active,
		AuditLog:    sql.NewDefaultAuditLog(userId),
		ProfilePlatformMapping: &repository.ProfilePlatformMapping{
			ProfileId: configurationBean.ProfileId,
			Platform:  platform,
		},
	}
	if configurationBean.ProfileName == v1.GLOBAL_PROFILE_NAME {
		infraProfile.Active = true
	}
	return infraProfile
}

// Deprecated: GetV0ProfileBean is used for backward compatibility with V0.
// Only used for deprecated APIs.
func GetV0ProfileBean(profileBean *v1.ProfileBeanDto) *v0.ProfileBeanV0 {
	if profileBean == nil {
		return &v0.ProfileBeanV0{}
	}

	profileName := profileBean.GetName()
	if profileName == v1.GLOBAL_PROFILE_NAME {
		profileName = v1.DEFAULT_PROFILE_NAME
	}

	profileType := profileBean.Type
	if profileType == v1.GLOBAL {
		profileType = v1.DEFAULT
	}

	ciRunnerConfig := profileBean.GetConfigurations()[v1.RUNNER_PLATFORM]
	profileV0Bean := &v0.ProfileBeanV0{
		ProfileBeanAbstract: v1.ProfileBeanAbstract{
			Id:          profileBean.Id,
			Name:        profileName,
			Description: profileBean.GetDescription(),
			Active:      profileBean.Active,
			Type:        profileType,
			AppCount:    profileBean.AppCount,
		},

		Configurations: ConvertToV0ConfigBeans(ciRunnerConfig),
	}
	profileV0Bean.BuildxDriverType = profileBean.GetBuildxDriverType()
	return profileV0Bean
}

// ConvertToV1ProfileBean converts V0 ProfileBean to V1 ProfileBean
// Only used for deprecated APIs handling.
func ConvertToV1ProfileBean(profileBean *v0.ProfileBeanV0) *v1.ProfileBeanDto {
	if profileBean == nil {
		return nil
	}
	profileName := profileBean.GetName()
	if profileName == v1.DEFAULT_PROFILE_NAME {
		profileName = v1.GLOBAL_PROFILE_NAME
	}
	profileType := profileBean.Type
	if profileType == v1.GLOBAL {
		profileType = v1.DEFAULT
	}
	newProfileBean := &v1.ProfileBeanDto{
		ProfileBeanAbstract: v1.ProfileBeanAbstract{
			Id:          profileBean.Id,
			Name:        profileName,
			Description: profileBean.GetDescription(),
			Active:      profileBean.Active,
			Type:        profileType,
			AppCount:    profileBean.AppCount,
		},
		Configurations: map[string][]*v1.ConfigurationBean{v1.RUNNER_PLATFORM: getV1ConfigBeans(profileBean.Configurations)},
	}
	newProfileBean.BuildxDriverType = profileBean.GetBuildxDriverType()
	return newProfileBean
}

func getV1ConfigBeans(configBeans []v0.ConfigurationBeanV0) []*v1.ConfigurationBean {
	if len(configBeans) == 0 {
		return nil
	}
	resp := make([]*v1.ConfigurationBean, 0)
	for _, configBean := range configBeans {
		profileName := configBean.ProfileName
		if profileName == v1.GLOBAL_PROFILE_NAME {
			profileName = v1.DEFAULT_PROFILE_NAME
		}
		configBeanV1 := &v1.ConfigurationBean{
			ConfigurationBeanAbstract: v1.ConfigurationBeanAbstract{
				Id:          configBean.Id,
				Key:         configBean.Key,
				Unit:        configBean.Unit,
				Active:      configBean.Active,
				ProfileId:   configBean.ProfileId,
				ProfileName: profileName,
			},
			Value: configBean.Value,
		}
		resp = append(resp, configBeanV1)
	}
	return resp
}

// ConvertToV0ConfigBeans converts V1 ConfigurationBean to V0 ConfigurationBean
// Only used for deprecated APIs handling.
func ConvertToV0ConfigBeans(configBeans []*v1.ConfigurationBean) []v0.ConfigurationBeanV0 {
	if len(configBeans) == 0 {
		return []v0.ConfigurationBeanV0{}
	}
	resp := make([]v0.ConfigurationBeanV0, 0)
	for _, configBean := range configBeans {
		if !slices.Contains(v1.V0ConfigKeys, configBean.Key) {
			// here skipping the value for the NodeSelectors and TolerationsKey
			continue
		}
		// Cast the returned value to float64 for supported keys
		valueFloat, ok := configBean.Value.(float64)
		if !ok {
			continue
		}
		profileName := configBean.ProfileName
		if profileName == v1.GLOBAL_PROFILE_NAME {
			profileName = v1.DEFAULT_PROFILE_NAME
		}

		// Construct the V0 bean
		beanv0 := v0.ConfigurationBeanV0{
			ConfigurationBeanAbstract: v1.ConfigurationBeanAbstract{
				Id:          configBean.Id,
				Key:         configBean.Key,
				Unit:        configBean.Unit,
				Active:      configBean.Active,
				ProfileId:   configBean.ProfileId,
				ProfileName: profileName,
			},
			Value: valueFloat,
		}
		resp = append(resp, beanv0)
	}

	return resp
}

// ConvertToProfileBean converts *repository.InfraProfileEntity to *bean.ProfileBeanDto
func ConvertToProfileBean(infraProfile *repository.InfraProfileEntity) *v1.ProfileBeanDto {
	profileType := v1.GLOBAL
	if infraProfile.Name != v1.GLOBAL_PROFILE_NAME {
		profileType = v1.NORMAL
	}
	newProfileBean := &v1.ProfileBeanDto{
		ProfileBeanAbstract: v1.ProfileBeanAbstract{
			Id:          infraProfile.Id,
			Name:        infraProfile.Name,
			Type:        profileType,
			Description: infraProfile.Description,
			Active:      infraProfile.Active,
		},
	}
	newProfileBean.BuildxDriverType = infraProfile.BuildxDriverType
	return newProfileBean
}

// ConvertToInfraProfileEntity converts *bean.ProfileBeanDto to *repository.InfraProfileEntity
func ConvertToInfraProfileEntity(profileBean *v1.ProfileBeanDto) *repository.InfraProfileEntity {
	return &repository.InfraProfileEntity{
		Id:               profileBean.Id,
		Name:             profileBean.GetName(),
		Description:      profileBean.GetDescription(),
		BuildxDriverType: profileBean.GetBuildxDriverType(),
	}
}

// NewInfraProfileConfigEntity creates a new instance of repository.InfraProfileConfigurationEntity
// Used for creating new configuration entity for migration.
func NewInfraProfileConfigEntity(key v1.ConfigKeyStr, profileId int, platform string, parsedValue *unitsBean.ParsedValue) *repository.InfraProfileConfigurationEntity {
	// Create the DB entity
	return &repository.InfraProfileConfigurationEntity{
		Key:         util.GetConfigKey(key),
		UniqueId:    repository.GetUniqueId(profileId, platform),
		Unit:        parsedValue.GetUnitType(),
		ValueString: parsedValue.GetValueString(),
		ProfilePlatformMapping: &repository.ProfilePlatformMapping{
			ProfileId: profileId,
			Platform:  platform,
		},
	}
}

// NewInheritedEntityForPlatform creates a new instance of repository.InfraProfileConfigurationEntity from existing entity, but for a different platform.
// Used for creating missing configuration entities for migration flow.
func NewInheritedEntityForPlatform(entity *repository.InfraProfileConfigurationEntity, platform string, userId int32) *repository.InfraProfileConfigurationEntity {
	return &repository.InfraProfileConfigurationEntity{
		Key:         entity.Key,
		Unit:        entity.Unit,
		Value:       entity.Value, // maintained for backward compatibility
		ValueString: entity.ValueString,
		ProfileId:   entity.ProfilePlatformMapping.ProfileId, // maintained for backward compatibility
		UniqueId:    repository.GetUniqueId(entity.ProfilePlatformMapping.ProfileId, platform),
		Active:      entity.Active,
		AuditLog:    sql.NewDefaultAuditLog(userId),
	}
}

// UpdatePlatformMappingInConfigEntities
//   - updates the ProfilePlatformMappingId in the repository.InfraProfileConfigurationEntity
func UpdatePlatformMappingInConfigEntities(infraConfigurations []*repository.InfraProfileConfigurationEntity,
	platformMappings []*repository.ProfilePlatformMapping) []*repository.InfraProfileConfigurationEntity {
	platformMappingId := make(map[string]int)
	for _, platformMapping := range platformMappings {
		if len(platformMapping.UniqueId) == 0 {
			platformMapping.UniqueId = repository.GetUniqueId(platformMapping.ProfileId, platformMapping.Platform)
		}
		platformMappingId[platformMapping.UniqueId] = platformMapping.Id
	}

	for _, infraConfiguration := range infraConfigurations {
		if profilePlatformMappingId, ok := platformMappingId[infraConfiguration.UniqueId]; ok {
			infraConfiguration.ProfilePlatformMappingId = profilePlatformMappingId
			if infraConfiguration.ProfilePlatformMapping != nil {
				infraConfiguration.ProfilePlatformMapping.Id = profilePlatformMappingId
			}
		}
	}
	return infraConfigurations
}

func GetGenericConfigurationBean[T any](configurationBean *v1.ConfigurationBean, typedValue T) *v1.GenericConfigurationBean[T] {
	return &v1.GenericConfigurationBean[T]{
		ConfigurationBeanAbstract: configurationBean.ConfigurationBeanAbstract,
		Value:                     typedValue,
	}
}

// FillMissingConfigurationsForThePayloadV0 - This function is used to fill the missing configurations in the payload
// after the k8sBuildXDriverOpts Migration => need for handling the updated of default / global profile
func FillMissingConfigurationsForThePayloadV0(profileToUpdate *v1.ProfileBeanDto, platformMapConfigs map[string][]*v1.ConfigurationBean) {
	for platform, configBeans := range platformMapConfigs {
		if existingConfig, exists := profileToUpdate.GetConfigurations()[platform]; exists {
			// If the platform already exists in the payloadConfig, update missing NodeSelectors and TolerationsKey
			defaultKeys := util.GetDefaultConfigKeysMapV0()
			for _, beans := range existingConfig {
				defaultKeys[beans.Key] = false
			}

			for _, configBean := range configBeans {
				// Add missing in case of NodeSelectors and TolerationsKey only
				if (configBean.Key == v1.NODE_SELECTOR || configBean.Key == v1.TOLERATIONS) && defaultKeys[configBean.Key] {
					profileToUpdate.GetConfigurations()[platform] = append(profileToUpdate.GetConfigurations()[platform], configBean)
				}
			}
		} else {
			// If the platform do not found in the update request, add all its configurations corresponding to its platform
			profileToUpdate.GetConfigurations()[platform] = configBeans
		}
	}
}
