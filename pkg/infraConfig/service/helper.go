/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"slices"
)

func prepareConfigurationsAndMappings(creatableConfigs []*repository.InfraProfileConfigurationEntity, dbConfigs []*repository.InfraProfileConfigurationEntity,
	existingPlatforms []string, profileId int, migrationRequired bool) ([]*repository.InfraProfileConfigurationEntity, []*repository.ProfilePlatformMapping) {
	defaultDbConfigMap := util.CreateConfigKeyPlatformMap(dbConfigs)
	existingPlatformMap := sliceUtil.GetMapOf(existingPlatforms, true)
	var creatableConfigurations []*repository.InfraProfileConfigurationEntity
	var platformMappings []*repository.ProfilePlatformMapping
	platformsAdded := make(map[string]bool)

	for _, config := range creatableConfigs {
		platform := config.ProfilePlatformMapping.Platform
		if platform == "" {
			platform = v1.RUNNER_PLATFORM
		}
		keyPlatform := v1.ConfigKeyPlatformKey{Key: config.Key, Platform: platform}

		if !defaultDbConfigMap[keyPlatform] {
			config.Active = true
			config.UniqueId = repository.GetUniqueId(profileId, platform)
			config.ProfileId = profileId // maintained for backward compatibility
			config.ProfilePlatformMapping = &repository.ProfilePlatformMapping{
				ProfileId: profileId,
				Platform:  platform,
			}
			config.AuditLog = sql.NewDefaultAuditLog(1)
			creatableConfigurations = append(creatableConfigurations, config)
		}

		if migrationRequired && !existingPlatformMap[platform] && !platformsAdded[platform] {
			platformsAdded[platform] = true
			platformMappings = append(platformMappings, &repository.ProfilePlatformMapping{
				ProfileId: profileId,
				Platform:  platform,
				Active:    true,
				AuditLog:  sql.NewDefaultAuditLog(1),
				UniqueId:  repository.GetUniqueId(profileId, platform),
			})
		}
	}
	return creatableConfigurations, platformMappings
}

func getAppliedConfigForProfileV0(profile *v1.ProfileBeanDto, defaultConfigurationsMap map[string][]*v1.ConfigurationBean) *v1.ProfileBeanDto {
	if len(profile.GetConfigurations()) == 0 {
		profile.Configurations = defaultConfigurationsMap
		return profile
	}
	for platform, defaultConfigurations := range defaultConfigurationsMap {
		extraConfigurations := make([]*v1.ConfigurationBean, 0)
		for _, defaultConfiguration := range defaultConfigurations {
			if !slices.ContainsFunc(profile.GetConfigurations()[platform], func(config *v1.ConfigurationBean) bool {
				return config.Key == defaultConfiguration.Key
			}) {
				extraConfigurations = append(extraConfigurations, defaultConfiguration)
			}
		}
		// if the profile doesn't have the default configuration, add it to the profile
		profile.GetConfigurations()[platform] = append(profile.GetConfigurations()[platform], extraConfigurations...)
	}
	return profile
}

func filterCreatableConfigForDefaultAndOverrideEnvConfigs(envConfigs, dbConfigs []*repository.InfraProfileConfigurationEntity) ([]*repository.InfraProfileConfigurationEntity, []*repository.InfraProfileConfigurationEntity) {
	creatableConfigs := make([]*repository.InfraProfileConfigurationEntity, 0)
	// Create a map for faster lookups of DB configurations by key
	dbConfigMap := make(map[v1.ConfigKeyStr]*repository.InfraProfileConfigurationEntity)
	for _, dbConfig := range dbConfigs {
		if dbConfig.ProfilePlatformMapping.Platform == v1.RUNNER_PLATFORM {
			dbConfigMap[util.GetConfigKeyStr(dbConfig.Key)] = dbConfig
		}
	}
	// Override environment configurations with database configurations
	for i, envConfig := range envConfigs {
		if dbConfig, exists := dbConfigMap[util.GetConfigKeyStr(envConfig.Key)]; exists {
			// Create a copy of dbConfig to avoid mutating the original
			copiedConfig := *dbConfig
			envConfigs[i] = &copiedConfig
		} else {
			creatableConfigs = append(creatableConfigs, envConfigs[i])
		}
	}
	return creatableConfigs, envConfigs
}

func getConfiguredInfraConfigKeys(platform string, configurations []*v1.ConfigurationBean) v1.InfraConfigKeys {
	// Get the supported keys for the platform,
	// and mark the ones that are present
	supportedConfigKeys := util.GetConfigKeysMapForPlatform(platform)
	// Mark the keys that are present
	for _, config := range configurations {
		if supportedConfigKeys.IsSupported(config.Key) && config.Active {
			supportedConfigKeys = supportedConfigKeys.MarkConfigured(config.Key)
		}
	}
	return supportedConfigKeys
}

func getDefaultInfraConfigFromEnv(envConfig *types.CiConfig) (*v1.InfraConfig, error) {
	infraConfiguration := &v1.InfraConfig{}
	err := env.Parse(infraConfiguration)
	if err != nil {
		return infraConfiguration, err
	}
	infraConfiguration, err = updateEntInfraConfigFromEnv(infraConfiguration, envConfig)
	if err != nil {
		return infraConfiguration, err
	}
	return infraConfiguration, nil
}
