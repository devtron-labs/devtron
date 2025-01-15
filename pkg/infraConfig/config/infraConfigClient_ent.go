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

package config

import (
	"fmt"
	v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/go-pg/pg"
)

type InfraConfigEntClient interface{}

func (impl *InfraConfigClientImpl) getEntConfigurationUnits() (map[v1.ConfigKeyStr]map[string]v1.Unit, error) {
	return make(map[v1.ConfigKeyStr]map[string]v1.Unit), nil
}

func (impl *InfraConfigClientImpl) formatTypedValueAsStringEnt(configKey v1.ConfigKeyStr, configValue any) (string, error) {
	// Default case, return error for an unsupported key
	return "", fmt.Errorf("config key %q not supported", configKey)
}

func (impl *InfraConfigClientImpl) validateEntConfig(supportedConfigKeyMap v1.InfraConfigKeys, platformConfigurations, defaultConfigurations []*v1.ConfigurationBean, skipError bool) (v1.InfraConfigKeys, error) {
	return supportedConfigKeyMap, nil
}

func (impl *InfraConfigClientImpl) convertValueStringToInterfaceEnt(configKey v1.ConfigKeyStr, valueString string) (any, int, error) {
	// Default case, return error for an unsupported key
	return nil, 0, fmt.Errorf("config key %q not supported", configKey)
}

func (impl *InfraConfigClientImpl) isConfigActiveEnt(configKey v1.ConfigKeyStr, valueCount int, configActive bool) bool {
	// Default case, return the flag configActive as is
	return configActive
}

func (impl *InfraConfigClientImpl) handlePostUpdateOperationEnt(tx *pg.Tx, updatedInfraConfig *repository.InfraProfileConfigurationEntity) error {
	// Default case, return error for an unsupported key
	return fmt.Errorf("config key %q not supported", updatedInfraConfig.Key)
}

func (impl *InfraConfigClientImpl) handlePostCreateOperationEnt(tx *pg.Tx, createdInfraConfig *repository.InfraProfileConfigurationEntity) error {
	return fmt.Errorf("config key %q not supported", createdInfraConfig.Key)
}

func (impl *InfraConfigClientImpl) getInfraConfigEntEntities(profileId int, infraConfig *v1.InfraConfig) ([]*repository.InfraProfileConfigurationEntity, error) {
	return make([]*repository.InfraProfileConfigurationEntity, 0), nil
}

func (impl *InfraConfigClientImpl) overrideInfraConfigEnt(infraConfiguration *v1.InfraConfig, configurationBean *v1.ConfigurationBean) (*v1.InfraConfig, error) {
	return nil, fmt.Errorf("config key %q not supported", configurationBean.Key)
}

func (impl *InfraConfigClientImpl) mergeInfraConfigurationsEnt(supportedConfigKey v1.ConfigKeyStr, profileConfiguration *v1.ConfigurationBean, defaultConfigurations []*v1.ConfigurationBean) (*v1.ConfigurationBean, error) {
	return nil, fmt.Errorf("config key %q not supported", supportedConfigKey)
}

func (impl *InfraConfigClientImpl) handleInfraConfigTriggerAuditEnt(supportedConfigKeys v1.InfraConfigKeys, workflowId int, triggeredBy int32, infraConfig *v1.InfraConfig) error {
	return nil
}
