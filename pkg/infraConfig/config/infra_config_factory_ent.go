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
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/variables"
	"go.uber.org/zap"
)

func getConfigFactory(logger *zap.SugaredLogger,
	scopedVariableManager variables.ScopedVariableManager,
	configReadService read.ConfigReadService) *configFactories {
	return &configFactories{
		cpuConfigFactory:     newCPUClientImpl(logger),
		memConfigFactory:     newMemClientImpl(logger),
		timeoutConfigFactory: newTimeoutClientImpl(logger),
	}
}

func getUnitFactoryMap(logger *zap.SugaredLogger) *unitFactories {
	cpuUnitFactory := units.NewCPUUnitFactory(logger)
	memUnitFactory := units.NewMemoryUnitFactory(logger)
	timeUnitFactory := units.NewTimeUnitFactory(logger)
	unitFactoryMap := &unitFactories{
		cpuUnitFactory:  cpuUnitFactory,
		memUnitFactory:  memUnitFactory,
		timeUnitFactory: timeUnitFactory,
	}
	return unitFactoryMap
}
