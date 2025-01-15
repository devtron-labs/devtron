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
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/go-pg/pg"
)

type configFactory[T any] interface {
	validate(platformConfigurations, defaultConfigurations []*v1.ConfigurationBean) error
	getConfigKeys() []v1.ConfigKeyStr
	getSupportedUnits() map[v1.ConfigKeyStr]map[string]v1.Unit

	getInfraConfigEntities(infraConfig *v1.InfraConfig, profileId int, platformName string) ([]*repository.InfraProfileConfigurationEntity, error)
	getValueFromString(valueString string) (T, int, error)
	isConfigActive(valueCount int, configActive bool) bool
	getValueFromBean(configurationBean *v1.ConfigurationBean) (T, error)
	formatTypedValueAsString(configValue any) (string, error)
	overrideInfraConfig(infraConfiguration *v1.InfraConfig, configurationBean *v1.ConfigurationBean) (*v1.InfraConfig, error)
	getAppliedConfiguration(key v1.ConfigKeyStr, infraConfiguration *v1.ConfigurationBean, defaultConfigurations []*v1.ConfigurationBean) (*v1.ConfigurationBean, error)
	handlePostCreateOperations(tx *pg.Tx, createdInfraConfig *repository.InfraProfileConfigurationEntity) error
	handlePostUpdateOperations(tx *pg.Tx, updatedInfraConfig *repository.InfraProfileConfigurationEntity) error
	handlePostDeleteOperations(tx *pg.Tx, deletedInfraConfig *repository.InfraProfileConfigurationEntity) error
	handleInfraConfigTriggerAudit(workflowId int, triggeredBy int32, infraConfig *v1.InfraConfig) error
	resolveScopeVariablesForAppliedConfiguration(scope resourceQualifiers.Scope, configuration *v1.ConfigurationBean) (*v1.ConfigurationBean, map[string]string, error)
}
