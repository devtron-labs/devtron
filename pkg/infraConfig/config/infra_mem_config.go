package config

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/infraConfig/adapter"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/errors"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	unitsBean "github.com/devtron-labs/devtron/pkg/infraConfig/units/bean"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
	"reflect"
	"strconv"
)

type memClientImpl struct {
	logger         *zap.SugaredLogger
	memUnitFactory units.UnitService[float64]
}

func newMemClientImpl(logger *zap.SugaredLogger) *memClientImpl {
	return &memClientImpl{
		logger:         logger,
		memUnitFactory: units.NewMemoryUnitFactory(logger),
	}
}

func (impl *memClientImpl) getMemoryClient() units.UnitService[float64] {
	return impl.memUnitFactory
}

// mergedConfiguration:
//   - If configurations are not found in profileBean,
//     then the merged memory configuration will be the global profile platform
//     (either the same platform or default platform of globalProfile).
//   - If configurations are found in profileBean,
//     then the merged memory configuration will be the profile configuration.
//
// Inputs:
//   - memLimit: *bean.ConfigurationBean
//   - memReq: *bean.ConfigurationBean
//   - defaultConfigurations: []*bean.ConfigurationBean
//
// Outputs:
//   - *bean.ConfigurationBean (merged memLimit configuration)
//   - *bean.ConfigurationBean (merged memReq configuration)
//   - error
func (impl *memClientImpl) getAppliedConfiguration(key v1.ConfigKeyStr, profileConfigBean *v1.ConfigurationBean, defaultConfigurations []*v1.ConfigurationBean) (*v1.ConfigurationBean, error) {
	profileData, err := impl.getValueFromBean(profileConfigBean)
	if err != nil {
		impl.logger.Errorw("error in getting memory config data", "error", err, "memoryConfig", profileConfigBean)
		return profileConfigBean, err
	}
	defaultConfigBean, defaultData, err := impl.getConfigBeanAndDataForKey(key, defaultConfigurations)
	if err != nil {
		impl.logger.Errorw("error in getting memory config data", "error", err, "memoryConfig", defaultConfigurations)
		return profileConfigBean, err
	}
	if profileConfigBean != nil {
		profileConfigBean.Active = impl.isConfigActive(impl.getValueCount(profileData), profileConfigBean.Active)
	}
	if defaultConfigBean != nil {
		defaultConfigBean.Active = impl.isConfigActive(impl.getValueCount(defaultData), defaultConfigBean.Active)
	}
	return impl.getInheritedConfigurations(key, profileData, defaultData, profileConfigBean, defaultConfigBean)
}

func (impl *memClientImpl) getConfigBeanAndDataForKey(key v1.ConfigKeyStr, configurations []*v1.ConfigurationBean) (configBean *v1.ConfigurationBean, configData float64, err error) {
	for _, configuration := range configurations {
		if configuration.Key == key {
			configBean = configuration
			configData, err = impl.getValueFromBean(configBean)
			if err != nil {
				impl.logger.Errorw("error in getting memory config data", "error", err, "memoryConfig", configBean)
				return configBean, configData, err
			}
			return configBean, configData, nil
		}
	}
	return configBean, configData, nil
}

func (impl *memClientImpl) getInheritedConfigurations(key v1.ConfigKeyStr, profileData, defaultData float64,
	profileConfigBean, defaultConfigBean *v1.ConfigurationBean) (*v1.ConfigurationBean, error) {
	defaultConfigBeanAbstract := v1.ConfigurationBeanAbstract{
		Key:  key,
		Unit: impl.memUnitFactory.GetDefaultUnitSuffix(),
	}
	return getInheritedConfigurations(defaultConfigBeanAbstract, profileData, defaultData, profileConfigBean, defaultConfigBean)
}

func (impl *memClientImpl) validate(platformConfigurations, defaultConfigurations []*v1.ConfigurationBean) (err error) {
	var memLimit, memReq *v1.ConfigurationBean
	for _, configuration := range platformConfigurations {
		// get memory limit and req
		switch configuration.Key {
		case v1.MEMORY_LIMIT:
			memLimit = configuration
		case v1.MEMORY_REQUEST:
			memReq = configuration
		}
	}
	memLimit, err = impl.getAppliedConfiguration(v1.MEMORY_LIMIT, memLimit, defaultConfigurations)
	if err != nil {
		impl.logger.Errorw("error in merging memory limit configurations", "error", err, "memLimit", memLimit)
		return err
	}
	memReq, err = impl.getAppliedConfiguration(v1.MEMORY_REQUEST, memReq, defaultConfigurations)
	if err != nil {
		impl.logger.Errorw("error in merging memory request configurations", "error", err, "memReq", memReq)
		return err
	}
	if !memLimit.IsEmpty() && !memReq.IsEmpty() {
		err = impl.validMemoryBean(memLimit, memReq)
		if err != nil {
			impl.logger.Errorw("error in validating memory", "error", err, "memLimit", memLimit, "memReq", memReq)
			return err
		}
	}
	return nil
}

func (impl *memClientImpl) validMemoryBean(memLimit, memReq *v1.ConfigurationBean) error {
	memLimitValue, err := impl.getValueFromBean(memLimit)
	if err != nil {
		impl.logger.Errorw("error in getting memory limit value", "error", err, "memLimit", memLimit)
		return err
	}
	memLimitConfiguration := adapter.GetGenericConfigurationBean(memLimit, memLimitValue)
	memLimitConfig, err := impl.getMemoryClient().Validate(memLimitConfiguration)
	if err != nil {
		impl.logger.Errorw("error in validating memory limit", "error", err, "memLimit", memLimit)
		return err
	}
	memReqValue, err := impl.getValueFromBean(memReq)
	if err != nil {
		impl.logger.Errorw("error in getting memory request value", "error", err, "memReq", memReq)
		return err
	}
	memReqConfiguration := adapter.GetGenericConfigurationBean(memReq, memReqValue)
	memReqConfig, err := impl.getMemoryClient().Validate(memReqConfiguration)
	if err != nil {
		impl.logger.Errorw("error in validating memory request", "error", err, "memReq", memReq)
		return err
	}
	if !memLimitConfig.IsEmpty() && !memReqConfig.IsEmpty() {
		if !impl.validMemoryLimitRequest(memLimitConfig.Value, memLimitConfig.Unit.ConversionFactor, memReqConfig.Value, memReqConfig.Unit.ConversionFactor) {
			impl.logger.Errorw("error in comparing memory limit and request", "memLimit", memLimitConfig, "memReq", memReqConfig)
			return util.NewApiError(http.StatusBadRequest, errors.MEMLimReqErrorCompErr, errors.MEMLimReqErrorCompErr)
		}
	}
	return nil
}

func (impl *memClientImpl) validMemoryLimitRequest(lim, limFactor, req, reqFactor float64) bool {
	return validLimitRequestForCPUorMem(lim, limFactor, req, reqFactor)
}

func (impl *memClientImpl) getConfigKeys() []v1.ConfigKeyStr {
	return []v1.ConfigKeyStr{v1.MEMORY_LIMIT, v1.MEMORY_REQUEST}
}

func (impl *memClientImpl) getSupportedUnits() map[v1.ConfigKeyStr]map[string]v1.Unit {
	supportedUnitsMap := make(map[v1.ConfigKeyStr]map[string]v1.Unit)
	supportedUnits := impl.getMemoryClient().GetAllUnits()
	for _, configKey := range impl.getConfigKeys() {
		supportedUnitsMap[configKey] = supportedUnits
	}
	return supportedUnitsMap
}

func (impl *memClientImpl) getInfraConfigEntities(infraConfig *v1.InfraConfig, profileId int, platformName string) ([]*repository.InfraProfileConfigurationEntity, error) {
	defaultConfigurations := make([]*repository.InfraProfileConfigurationEntity, 0)
	memLimitValue, unitType, err := parseCPUorMemoryValue[unitsBean.MemoryUnitStr](infraConfig.CiLimitMem)
	if err != nil {
		return defaultConfigurations, err
	}
	memLimitParsedValue, err := impl.getMemoryClient().ParseValAndUnit(memLimitValue, unitType.GetUnitSuffix())
	if err != nil {
		return defaultConfigurations, err
	}
	memLimit := adapter.NewInfraProfileConfigEntity(v1.MEMORY_LIMIT, profileId, platformName, memLimitParsedValue)
	defaultConfigurations = append(defaultConfigurations, memLimit)
	memReqValue, unitType, err := parseCPUorMemoryValue[unitsBean.MemoryUnitStr](infraConfig.CiReqMem)
	if err != nil {
		return defaultConfigurations, err
	}
	memReqParsedValue, err := impl.getMemoryClient().ParseValAndUnit(memReqValue, unitType.GetUnitSuffix())
	if err != nil {
		return defaultConfigurations, err
	}
	memReq := adapter.NewInfraProfileConfigEntity(v1.MEMORY_REQUEST, profileId, platformName, memReqParsedValue)
	defaultConfigurations = append(defaultConfigurations, memReq)
	return defaultConfigurations, nil
}

func (impl *memClientImpl) getValueFromString(valueString string) (float64, int, error) {
	// Convert string to float64 and truncate to 2 decimal places
	valueFloat, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0, 0, err
	}
	truncateValue := globalUtil.TruncateFloat(valueFloat, 2)
	return truncateValue, impl.getValueCount(truncateValue), nil // Returning float64 for resource values

}

func (impl *memClientImpl) overrideInfraConfig(infraConfiguration *v1.InfraConfig, configurationBean *v1.ConfigurationBean) (*v1.InfraConfig, error) {
	memConfigData, err := impl.getValueFromBean(configurationBean)
	if err != nil {
		return infraConfiguration, err
	}
	var memInfraConfigData string
	if configurationBean.Unit == unitsBean.BYTE.String() {
		memInfraConfigData = fmt.Sprintf("%v", memConfigData)
	} else {
		memInfraConfigData = fmt.Sprintf("%v%v", memConfigData, configurationBean.Unit)
	}
	if configurationBean.Key == v1.MEMORY_REQUEST {
		infraConfiguration = infraConfiguration.SetCiReqMem(memInfraConfigData)
	} else if configurationBean.Key == v1.MEMORY_LIMIT {
		infraConfiguration = infraConfiguration.SetCiLimitMem(memInfraConfigData)
	} else {
		errMsg := fmt.Sprintf("invalid key %q for memory configuration", configurationBean.Key)
		return infraConfiguration, util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	return infraConfiguration, nil
}

func (impl *memClientImpl) getValueFromBean(configurationBean *v1.ConfigurationBean) (float64, error) {
	if configurationBean == nil {
		return 0, nil
	}
	valueString, err := impl.formatTypedValueAsString(configurationBean.Value)
	if err != nil {
		return 0, err
	}
	memConfigData, _, err := impl.getValueFromString(valueString)
	if err != nil {
		impl.logger.Errorw("error in getting configMap data", "error", err, "configMap", configurationBean)
		return 0, err
	}
	return memConfigData, nil
}

func (impl *memClientImpl) formatTypedValueAsString(configValue any) (string, error) {
	var valueFloat float64
	// Handle string input or directly as float64
	switch v := configValue.(type) {
	case float64:
		valueFloat = v
	default:
		errMsg := fmt.Sprintf("invalid value for memory configuration: %v", configValue)
		return "", util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	// Truncate and format the float value
	truncateValue := globalUtil.TruncateFloat(valueFloat, 2)
	return strconv.FormatFloat(truncateValue, 'f', -1, 64), nil
}

func (impl *memClientImpl) getValueCount(value float64) int {
	if reflect.ValueOf(value).IsZero() {
		return 0
	}
	return 1
}

func (impl *memClientImpl) isConfigActive(valueCount int, isConfigActive bool) bool {
	return isConfigActive
}

func (impl *memClientImpl) handlePostCreateOperations(tx *pg.Tx, createdInfraConfig *repository.InfraProfileConfigurationEntity) error {
	return nil
}

func (impl *memClientImpl) handlePostUpdateOperations(tx *pg.Tx, updatedInfraConfig *repository.InfraProfileConfigurationEntity) error {
	return nil
}

func (impl *memClientImpl) handlePostDeleteOperations(tx *pg.Tx, deletedInfraConfig *repository.InfraProfileConfigurationEntity) error {
	return nil
}

func (impl *memClientImpl) handleInfraConfigTriggerAudit(workflowId int, triggeredBy int32, infraConfig *v1.InfraConfig) error {
	return nil
}

func (impl *memClientImpl) resolveScopeVariablesForAppliedConfiguration(scope resourceQualifiers.Scope, configuration *v1.ConfigurationBean) (*v1.ConfigurationBean, map[string]string, error) {
	return configuration, nil, nil
}
