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
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"math"
	"net/http"
	"reflect"
	"strconv"
)

type timeoutClientImpl struct {
	logger          *zap.SugaredLogger
	timeUnitFactory units.UnitService[float64]
}

func newTimeoutClientImpl(logger *zap.SugaredLogger) *timeoutClientImpl {
	return &timeoutClientImpl{
		logger:          logger,
		timeUnitFactory: units.NewTimeUnitFactory(logger),
	}
}

func (impl *timeoutClientImpl) getTimeClient() units.UnitService[float64] {
	return impl.timeUnitFactory
}

// mergedConfiguration:
//   - If configurations are not found in profileBean,
//     then the merged timeout configuration will be the global profile platform
//     (either the same platform or default platform of globalProfile).
//   - If configurations are found in profileBean,
//     then the merged timeout configuration will be the profile configuration.
//
// Inputs:
//   - timeOut: *bean.ConfigurationBean
//   - defaultConfigurations: []*bean.ConfigurationBean
//
// Outputs:
//   - *bean.ConfigurationBean (merged timeout configuration)
//   - error
func (impl *timeoutClientImpl) getAppliedConfiguration(key v1.ConfigKeyStr, profileConfigBean *v1.ConfigurationBean, defaultConfigurations []*v1.ConfigurationBean) (*v1.ConfigurationBean, error) {
	profileData, err := impl.getValueFromBean(profileConfigBean)
	if err != nil {
		impl.logger.Errorw("error in getting timeout config data", "error", err, "timeoutConfig", profileConfigBean)
		return profileConfigBean, err
	}
	defaultConfigBean, defaultData, err := impl.getConfigBeanAndDataForKey(key, defaultConfigurations)
	if err != nil {
		impl.logger.Errorw("error in getting timeout config data", "error", err, "timeoutConfig", defaultConfigurations)
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

func (impl *timeoutClientImpl) getConfigBeanAndDataForKey(key v1.ConfigKeyStr, configurations []*v1.ConfigurationBean) (configBean *v1.ConfigurationBean, configData float64, err error) {
	for _, configuration := range configurations {
		if configuration.Key == key {
			configBean = configuration
			configData, err = impl.getValueFromBean(configBean)
			if err != nil {
				impl.logger.Errorw("error in getting timeout config data", "error", err, "timeoutConfig", configBean)
				return configBean, configData, err
			}
			return configBean, configData, nil
		}
	}
	return configBean, configData, nil
}

func (impl *timeoutClientImpl) getInheritedConfigurations(key v1.ConfigKeyStr, profileData, defaultData float64,
	profileConfigBean, defaultConfigBean *v1.ConfigurationBean) (*v1.ConfigurationBean, error) {
	defaultConfigBeanAbstract := v1.ConfigurationBeanAbstract{
		Key:  key,
		Unit: impl.timeUnitFactory.GetDefaultUnitSuffix(),
	}
	return getInheritedConfigurations(defaultConfigBeanAbstract, profileData, defaultData, profileConfigBean, defaultConfigBean)
}

func (impl *timeoutClientImpl) validate(platformConfigurations, defaultConfigurations []*v1.ConfigurationBean) (err error) {
	var timeOut *v1.ConfigurationBean
	for _, configuration := range platformConfigurations {
		switch configuration.Key {
		case v1.TIME_OUT:
			timeOut = configuration
		}
	}
	timeOut, err = impl.getAppliedConfiguration(v1.TIME_OUT, timeOut, defaultConfigurations)
	if err != nil {
		impl.logger.Errorw("error in merging time out configuration", "error", err, "timeOut", timeOut)
		return err
	}
	if !timeOut.IsEmpty() {
		err = impl.validateTimeOut(timeOut)
		if err != nil {
			impl.logger.Errorw("error in validating time out", "error", err, "timeOut", timeOut)
			return err
		}
	}
	return nil
}

func (impl *timeoutClientImpl) validateTimeOut(timeOut *v1.ConfigurationBean) error {
	timeOutValue, err := impl.getValueFromBean(timeOut)
	if err != nil {
		impl.logger.Errorw("error in getting time out data", "error", err, "timeOut", timeOut)
		return err
	}
	timeOutConfiguration := adapter.GetGenericConfigurationBean(timeOut, timeOutValue)
	_, err = impl.getTimeClient().Validate(timeOutConfiguration)
	if err != nil {
		impl.logger.Errorw("error in validating time out unit", "error", err, "timeOut", timeOut)
		return err
	}
	return nil
}

func (impl *timeoutClientImpl) getConfigKeys() []v1.ConfigKeyStr {
	return []v1.ConfigKeyStr{v1.TIME_OUT}
}

func (impl *timeoutClientImpl) getSupportedUnits() map[v1.ConfigKeyStr]map[string]v1.Unit {
	supportedUnitsMap := make(map[v1.ConfigKeyStr]map[string]v1.Unit)
	supportedUnits := impl.getTimeClient().GetAllUnits()
	for _, configKey := range impl.getConfigKeys() {
		supportedUnitsMap[configKey] = supportedUnits
	}
	return supportedUnitsMap
}

func (impl *timeoutClientImpl) getInfraConfigEntities(infraConfig *v1.InfraConfig, profileId int, platformName string) ([]*repository.InfraProfileConfigurationEntity, error) {
	defaultConfigurations := make([]*repository.InfraProfileConfigurationEntity, 0)
	timeoutParsedValue, err := impl.getTimeClient().ParseValAndUnit(infraConfig.CiDefaultTimeout, unitsBean.SecondStr.GetUnitSuffix())
	if err != nil {
		return defaultConfigurations, err
	}
	timeoutConfig := adapter.NewInfraProfileConfigEntity(v1.TIME_OUT, profileId, platformName, timeoutParsedValue)
	defaultConfigurations = append(defaultConfigurations, timeoutConfig)
	return defaultConfigurations, nil
}

func (impl *timeoutClientImpl) getValueFromString(valueString string) (float64, int, error) {
	// Convert string to float64 and ensure it's within integer range
	valueFloat, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0, 0, err
	}
	modifiedValue := math.Min(math.Floor(valueFloat), math.MaxInt64)
	return modifiedValue, impl.getValueCount(modifiedValue), nil // Returning float64 for timeout
}

func (impl *timeoutClientImpl) overrideInfraConfig(infraConfiguration *v1.InfraConfig, configurationBean *v1.ConfigurationBean) (*v1.InfraConfig, error) {
	timeoutData, err := impl.getValueFromBean(configurationBean)
	if err != nil {
		return infraConfiguration, err
	}
	// if a user ever gives the timeout in float, after conversion to int64 it will be rounded off
	timeUnit, ok := unitsBean.TimeUnitStr(configurationBean.Unit).GetUnit()
	if !ok {
		impl.logger.Errorw("error in getting time unit", "unit", configurationBean.Unit)
		errMsg := errors.InvalidUnitFound(configurationBean.Unit, configurationBean.Key)
		return infraConfiguration, util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	infraConfiguration = infraConfiguration.SetCiTimeout(timeoutData * timeUnit.ConversionFactor)
	return infraConfiguration, nil
}

func (impl *timeoutClientImpl) getValueFromBean(configurationBean *v1.ConfigurationBean) (float64, error) {
	if configurationBean == nil {
		return 0, nil
	}
	valueString, err := impl.formatTypedValueAsString(configurationBean.Value)
	if err != nil {
		return 0, err
	}
	timeoutData, _, err := impl.getValueFromString(valueString)
	if err != nil {
		impl.logger.Errorw("error in getting configMap data", "error", err, "configMap", configurationBean)
		return 0, err
	}
	return timeoutData, nil
}

func (impl *timeoutClientImpl) formatTypedValueAsString(configValue any) (string, error) {
	var valueFloat float64
	switch v := configValue.(type) {
	case float64:
		valueFloat = v
	default:
		errMsg := fmt.Sprintf("invalid value for timeout configuration: %v", configValue)
		return "", util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	//valueFloat, _ := strconv.ParseFloat(configValue, 64)
	modifiedValue := math.Min(math.Floor(valueFloat), math.MaxInt64)
	return strconv.FormatFloat(modifiedValue, 'f', -1, 64), nil
}

func (impl *timeoutClientImpl) getValueCount(value float64) int {
	if reflect.ValueOf(value).IsZero() {
		return 0
	}
	return 1
}

func (impl *timeoutClientImpl) isConfigActive(valueCount int, isConfigActive bool) bool {
	return isConfigActive
}

func (impl *timeoutClientImpl) handlePostCreateOperations(tx *pg.Tx, createdInfraConfig *repository.InfraProfileConfigurationEntity) error {
	return nil
}

func (impl *timeoutClientImpl) handlePostUpdateOperations(tx *pg.Tx, updatedInfraConfig *repository.InfraProfileConfigurationEntity) error {
	return nil
}

func (impl *timeoutClientImpl) handlePostDeleteOperations(tx *pg.Tx, deletedInfraConfig *repository.InfraProfileConfigurationEntity) error {
	return nil
}

func (impl *timeoutClientImpl) handleInfraConfigTriggerAudit(workflowId int, triggeredBy int32, infraConfig *v1.InfraConfig) error {
	return nil
}

func (impl *timeoutClientImpl) resolveScopeVariablesForAppliedConfiguration(scope resourceQualifiers.Scope, configuration *v1.ConfigurationBean) (*v1.ConfigurationBean, map[string]string, error) {
	return configuration, nil, nil
}
