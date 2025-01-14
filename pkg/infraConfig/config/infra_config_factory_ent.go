package config

import (
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/variables"
	"go.uber.org/zap"
)

// TODO: Add read.ConfigReadService to the function signature after creating one.

func getConfigFactory(logger *zap.SugaredLogger,
	scopedVariableManager variables.ScopedVariableManager) *configFactories {
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
