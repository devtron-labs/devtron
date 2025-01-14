package config

import (
	"fmt"
	v1 "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/go-pg/pg"
)

type InfraConfigEntClient interface{}

type configFactories struct {
	cpuConfigFactory     configFactory[float64]
	memConfigFactory     configFactory[float64]
	timeoutConfigFactory configFactory[float64]
}

type unitFactories struct {
	cpuUnitFactory  units.UnitService[float64]
	memUnitFactory  units.UnitService[float64]
	timeUnitFactory units.UnitService[float64]
}

func (impl *InfraConfigClientImpl) getEntConfigurationUnits() (map[v1.ConfigKeyStr]map[string]v1.Unit, error) {
	return make(map[v1.ConfigKeyStr]map[string]v1.Unit), nil
}

func (impl *InfraConfigClientImpl) formatTypedValueAsString(configKey v1.ConfigKeyStr, configValue any) (string, error) {
	switch configKey {
	case v1.CPU_LIMIT, v1.CPU_REQUEST:
		return impl.getCPUConfigFactory().formatTypedValueAsString(configValue)
	case v1.MEMORY_LIMIT, v1.MEMORY_REQUEST:
		return impl.getMemoryConfigFactory().formatTypedValueAsString(configValue)
	case v1.TIME_OUT:
		return impl.getTimeoutConfigFactory().formatTypedValueAsString(configValue)
	default:
		return "", fmt.Errorf("config key %q not supported", configKey)
	}
}

func (impl *InfraConfigClientImpl) validateEntConfig(supportedConfigKeyMap v1.InfraConfigKeys, platformConfigurations, defaultConfigurations []*v1.ConfigurationBean, skipError bool) (v1.InfraConfigKeys, error) {
	return supportedConfigKeyMap, nil
}

// convertValueStringToInterface converts valueString to interface{} based on key
func (impl *InfraConfigClientImpl) convertValueStringToInterface(configKey v1.ConfigKeyStr, valueString string) (any, int, error) {
	switch configKey {
	case v1.CPU_LIMIT, v1.CPU_REQUEST:
		return impl.getCPUConfigFactory().getValueFromString(valueString)
	case v1.MEMORY_LIMIT, v1.MEMORY_REQUEST:
		return impl.getMemoryConfigFactory().getValueFromString(valueString)
	case v1.TIME_OUT:
		return impl.getTimeoutConfigFactory().getValueFromString(valueString)
	// Add more cases as needed for different config keys
	default:
		// Default case, return error for an unsupported key
		return nil, 0, fmt.Errorf("config key %q not supported", configKey)
	}
}

// isConfigActive checks if the config is active based on the value count and repository. flag
func (impl *InfraConfigClientImpl) isConfigActive(configKey v1.ConfigKeyStr, valueCount int, configActive bool) bool {
	switch configKey {
	case v1.CPU_LIMIT, v1.CPU_REQUEST:
		return impl.getCPUConfigFactory().isConfigActive(valueCount, configActive)
	case v1.MEMORY_LIMIT, v1.MEMORY_REQUEST:
		return impl.getMemoryConfigFactory().isConfigActive(valueCount, configActive)
	case v1.TIME_OUT:
		return impl.getTimeoutConfigFactory().isConfigActive(valueCount, configActive)
	// Add more cases as needed for different config keys
	default:
		// Default case, return the flag configActive as is
		return configActive
	}
}

func (impl *InfraConfigClientImpl) HandlePostUpdateOperations(tx *pg.Tx, updatedInfraConfigs []*repository.InfraProfileConfigurationEntity) error {
	for _, updatedInfraConfig := range updatedInfraConfigs {
		switch util.GetConfigKeyStr(updatedInfraConfig.Key) {
		case v1.CPU_LIMIT, v1.CPU_REQUEST:
			return impl.getCPUConfigFactory().handlePostUpdateOperations(tx, updatedInfraConfig)
		case v1.MEMORY_LIMIT, v1.MEMORY_REQUEST:
			return impl.getMemoryConfigFactory().handlePostUpdateOperations(tx, updatedInfraConfig)
		case v1.TIME_OUT:
			return impl.getTimeoutConfigFactory().handlePostUpdateOperations(tx, updatedInfraConfig)
		default:
			return fmt.Errorf("config key %q not supported", updatedInfraConfig.Key)
		}
	}
	return nil
}

func (impl *InfraConfigClientImpl) HandlePostCreateOperations(tx *pg.Tx, createdInfraConfigs []*repository.InfraProfileConfigurationEntity) error {
	for _, createdInfraConfig := range createdInfraConfigs {
		switch util.GetConfigKeyStr(createdInfraConfig.Key) {
		case v1.CPU_LIMIT, v1.CPU_REQUEST:
			if err := impl.getCPUConfigFactory().handlePostCreateOperations(tx, createdInfraConfig); err != nil {
				return err
			}
		case v1.MEMORY_LIMIT, v1.MEMORY_REQUEST:
			if err := impl.getMemoryConfigFactory().handlePostCreateOperations(tx, createdInfraConfig); err != nil {
				return err
			}
		case v1.TIME_OUT:
			if err := impl.getTimeoutConfigFactory().handlePostCreateOperations(tx, createdInfraConfig); err != nil {
				return err
			}
		default:
			return fmt.Errorf("config key %q not supported", createdInfraConfig.Key)
		}
	}
	return nil
}

func (impl *InfraConfigClientImpl) getInfraConfigEntEntities(profileId int, infraConfig *v1.InfraConfig) ([]*repository.InfraProfileConfigurationEntity, error) {
	return make([]*repository.InfraProfileConfigurationEntity, 0), nil
}

func (impl *InfraConfigClientImpl) OverrideInfraConfig(infraConfiguration *v1.InfraConfig, configurationBean *v1.ConfigurationBean) (*v1.InfraConfig, error) {
	switch configurationBean.Key {
	case v1.CPU_LIMIT, v1.CPU_REQUEST:
		return impl.getCPUConfigFactory().overrideInfraConfig(infraConfiguration, configurationBean)
	case v1.MEMORY_LIMIT, v1.MEMORY_REQUEST:
		return impl.getMemoryConfigFactory().overrideInfraConfig(infraConfiguration, configurationBean)
	case v1.TIME_OUT:
		return impl.getTimeoutConfigFactory().overrideInfraConfig(infraConfiguration, configurationBean)
	default:
		return nil, fmt.Errorf("config key %q not supported", configurationBean.Key)
	}
}

func (impl *InfraConfigClientImpl) MergeInfraConfigurations(supportedConfigKey v1.ConfigKeyStr, profileConfiguration *v1.ConfigurationBean, defaultConfigurations []*v1.ConfigurationBean) (*v1.ConfigurationBean, error) {
	switch supportedConfigKey {
	case v1.CPU_LIMIT:
		return impl.getCPUConfigFactory().getAppliedConfiguration(v1.CPU_LIMIT, profileConfiguration, defaultConfigurations)
	case v1.CPU_REQUEST:
		return impl.getCPUConfigFactory().getAppliedConfiguration(v1.CPU_REQUEST, profileConfiguration, defaultConfigurations)
	case v1.MEMORY_LIMIT:
		return impl.getMemoryConfigFactory().getAppliedConfiguration(v1.MEMORY_LIMIT, profileConfiguration, defaultConfigurations)
	case v1.MEMORY_REQUEST:
		return impl.getMemoryConfigFactory().getAppliedConfiguration(v1.MEMORY_REQUEST, profileConfiguration, defaultConfigurations)
	case v1.TIME_OUT:
		return impl.getTimeoutConfigFactory().getAppliedConfiguration(v1.TIME_OUT, profileConfiguration, defaultConfigurations)
	default:
		return nil, fmt.Errorf("config key %q not supported", supportedConfigKey)
	}
}

func (impl *InfraConfigClientImpl) handleInfraConfigTriggerAudit(supportedConfigKeys v1.InfraConfigKeys, workflowId int, triggeredBy int32, infraConfig *v1.InfraConfig) error {
	if supportedConfigKeys.IsSupported(v1.CPU_LIMIT) && supportedConfigKeys.IsSupported(v1.CPU_REQUEST) {
		err := impl.getCPUConfigFactory().handleInfraConfigTriggerAudit(workflowId, triggeredBy, infraConfig)
		if err != nil {
			return err
		}
	}
	if supportedConfigKeys.IsSupported(v1.MEMORY_LIMIT) && supportedConfigKeys.IsSupported(v1.MEMORY_REQUEST) {
		err := impl.getMemoryConfigFactory().handleInfraConfigTriggerAudit(workflowId, triggeredBy, infraConfig)
		if err != nil {
			return err
		}
	}
	if supportedConfigKeys.IsSupported(v1.TIME_OUT) {
		err := impl.getTimeoutConfigFactory().handleInfraConfigTriggerAudit(workflowId, triggeredBy, infraConfig)
		if err != nil {
			return err
		}
	}
	return nil
}
