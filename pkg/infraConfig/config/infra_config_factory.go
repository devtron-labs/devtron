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
	"github.com/devtron-labs/devtron/pkg/config/read"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type configFactory[T any] interface {
	validate(platformConfigurations, defaultConfigurations []*v1.ConfigurationBean) error
	getConfigKeys() []v1.ConfigKeyStr
	getSupportedUnits() map[v1.ConfigKeyStr]map[string]v1.Unit

	getInfraConfigEntities(infraConfig *v1.InfraConfig, profileId int, platformName string) ([]*repository.InfraProfileConfigurationEntity, error)
	getValueFromString(valueString string) (T, int, error)
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

type configFactories struct {
	cpuConfigFactory     configFactory[float64]
	memConfigFactory     configFactory[float64]
	timeoutConfigFactory configFactory[float64]
	configEntFactories
}

type unitFactories struct {
	cpuUnitFactory  units.UnitService[float64]
	memUnitFactory  units.UnitService[float64]
	timeUnitFactory units.UnitService[float64]
	unitEntFactories
}

func getConfigFactory(logger *zap.SugaredLogger,
	scopedVariableManager variables.ScopedVariableManager,
	configReadService read.ConfigReadService) *configFactories {
	return &configFactories{
		cpuConfigFactory:     newCPUClientImpl(logger),
		memConfigFactory:     newMemClientImpl(logger),
		timeoutConfigFactory: newTimeoutClientImpl(logger),
		configEntFactories:   newConfigEntFactories(logger, scopedVariableManager, configReadService),
	}
}

func getUnitFactoryMap(logger *zap.SugaredLogger) *unitFactories {
	cpuUnitFactory := units.NewCPUUnitFactory(logger)
	memUnitFactory := units.NewMemoryUnitFactory(logger)
	timeUnitFactory := units.NewTimeUnitFactory(logger)
	unitFactoryMap := &unitFactories{
		cpuUnitFactory:   cpuUnitFactory,
		memUnitFactory:   memUnitFactory,
		timeUnitFactory:  timeUnitFactory,
		unitEntFactories: newUnitEntFactories(logger),
	}
	return unitFactoryMap
}
