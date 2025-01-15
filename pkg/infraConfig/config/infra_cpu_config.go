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

type cpuClientImpl struct {
	logger         *zap.SugaredLogger
	cpuUnitFactory units.UnitService[float64]
}

func newCPUClientImpl(logger *zap.SugaredLogger) *cpuClientImpl {
	return &cpuClientImpl{
		logger:         logger,
		cpuUnitFactory: units.NewCPUUnitFactory(logger),
	}
}

func (impl *cpuClientImpl) getCPUClient() units.UnitService[float64] {
	return impl.cpuUnitFactory
}

// mergedConfiguration:
//   - If configurations are not found in profileBean,
//     then the merged cpu configuration will be the global profile platform
//     (either the same platform or default platform of globalProfile).
//   - If configurations are found in profileBean,
//     then the merged cpu configuration will be the profile configuration.
//
// Inputs:
//   - cpuLimit/ cpuReq: *bean.ConfigurationBean
//   - defaultConfigurations: []*bean.ConfigurationBean
//
// Outputs:
//   - *bean.ConfigurationBean (merged cpuLimit/ cpuReq configuration)
//   - error
func (impl *cpuClientImpl) getAppliedConfiguration(key v1.ConfigKeyStr, profileConfigBean *v1.ConfigurationBean, defaultConfigurations []*v1.ConfigurationBean) (*v1.ConfigurationBean, error) {
	profileData, err := impl.getValueFromBean(profileConfigBean)
	if err != nil {
		impl.logger.Errorw("error in getting cpu config data", "error", err, "cpuConfig", profileConfigBean)
		return profileConfigBean, err
	}
	defaultConfigBean, defaultData, err := impl.getConfigBeanAndDataForKey(key, defaultConfigurations)
	if err != nil {
		impl.logger.Errorw("error in getting cpu default config data", "error", err, "cpuConfig", defaultConfigBean)
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

func (impl *cpuClientImpl) getConfigBeanAndDataForKey(key v1.ConfigKeyStr, configurations []*v1.ConfigurationBean) (configBean *v1.ConfigurationBean, configData float64, err error) {
	for _, configuration := range configurations {
		if configuration.Key == key {
			configBean = configuration
			configData, err = impl.getValueFromBean(configBean)
			if err != nil {
				impl.logger.Errorw("error in getting cpu config data", "error", err, "cpuConfig", configBean)
				return configBean, configData, err
			}
			return configBean, configData, nil
		}
	}
	return configBean, configData, nil
}

func (impl *cpuClientImpl) getInheritedConfigurations(key v1.ConfigKeyStr, profileData, defaultData float64,
	profileConfigBean, defaultConfigBean *v1.ConfigurationBean) (*v1.ConfigurationBean, error) {
	defaultConfigBeanAbstract := v1.ConfigurationBeanAbstract{
		Key:  key,
		Unit: impl.cpuUnitFactory.GetDefaultUnitSuffix(),
	}
	return getInheritedConfigurations(defaultConfigBeanAbstract, profileData, defaultData, profileConfigBean, defaultConfigBean)
}

func (impl *cpuClientImpl) validate(platformConfigurations, defaultConfigurations []*v1.ConfigurationBean) (err error) {
	var cpuLimit, cpuReq *v1.ConfigurationBean
	for _, configuration := range platformConfigurations {
		// get cpu limit and req
		switch configuration.Key {
		case v1.CPU_LIMIT:
			cpuLimit = configuration
		case v1.CPU_REQUEST:
			cpuReq = configuration
		}
	}
	cpuLimit, err = impl.getAppliedConfiguration(v1.CPU_LIMIT, cpuLimit, defaultConfigurations)
	if err != nil {
		impl.logger.Errorw("error in merging cpu limit configurations", "error", err, "cpuLimit", cpuLimit)
		return err
	}
	cpuReq, err = impl.getAppliedConfiguration(v1.CPU_REQUEST, cpuReq, defaultConfigurations)
	if err != nil {
		impl.logger.Errorw("error in merging cpu request configurations", "error", err, "cpuReq", cpuReq)
		return err
	}
	if !cpuLimit.IsEmpty() && !cpuReq.IsEmpty() {
		err = impl.validCPUConfig(cpuLimit, cpuReq)
		if err != nil {
			return err
		}
	}
	return nil
}

func (impl *cpuClientImpl) validCPUConfig(cpuLimit, cpuReq *v1.ConfigurationBean) error {
	cpuLimitValue, err := impl.getValueFromBean(cpuLimit)
	if err != nil {
		impl.logger.Errorw("error in getting cpu limit value", "error", err, "cpuLimit", cpuLimit)
		return err
	}
	cpuLimitConfiguration := adapter.GetGenericConfigurationBean(cpuLimit, cpuLimitValue)
	cpuLimitConfig, err := impl.getCPUClient().Validate(cpuLimitConfiguration)
	if err != nil {
		impl.logger.Errorw("error in validating cpu limit", "error", err, "cpuLimit", cpuLimit)
		return err
	}
	cpuReqValue, err := impl.getValueFromBean(cpuReq)
	if err != nil {
		impl.logger.Errorw("error in getting cpu request value", "error", err, "cpuReq", cpuReq)
		return err
	}
	cpuReqConfiguration := adapter.GetGenericConfigurationBean(cpuReq, cpuReqValue)
	cpuReqConfig, err := impl.getCPUClient().Validate(cpuReqConfiguration)
	if err != nil {
		impl.logger.Errorw("error in validating cpu request", "error", err, "cpuReq", cpuReq)
		return err
	}
	if !cpuLimitConfig.IsEmpty() && !cpuReqConfig.IsEmpty() {
		if !impl.validCPULimitRequest(cpuLimitConfig.Value, cpuLimitConfig.Unit.ConversionFactor, cpuReqConfig.Value, cpuReqConfig.Unit.ConversionFactor) {
			impl.logger.Errorw("error in comparing cpu limit and request", "cpuLimit", cpuLimitConfig, "cpuReq", cpuReqConfig)
			return util.NewApiError(http.StatusBadRequest, errors.CPULimReqErrorCompErr, errors.CPULimReqErrorCompErr)
		}
	}
	return nil
}

func (impl *cpuClientImpl) validCPULimitRequest(lim, limFactor, req, reqFactor float64) bool {
	return validLimitRequestForCPUorMem(lim, limFactor, req, reqFactor)
}

func (impl *cpuClientImpl) getConfigKeys() []v1.ConfigKeyStr {
	return []v1.ConfigKeyStr{v1.CPU_LIMIT, v1.CPU_REQUEST}
}

func (impl *cpuClientImpl) getSupportedUnits() map[v1.ConfigKeyStr]map[string]v1.Unit {
	supportedUnitsMap := make(map[v1.ConfigKeyStr]map[string]v1.Unit)
	supportedUnits := impl.getCPUClient().GetAllUnits()
	for _, configKey := range impl.getConfigKeys() {
		supportedUnitsMap[configKey] = supportedUnits
	}
	return supportedUnitsMap
}

func (impl *cpuClientImpl) getInfraConfigEntities(infraConfig *v1.InfraConfig, profileId int, platformName string) ([]*repository.InfraProfileConfigurationEntity, error) {
	defaultConfigurations := make([]*repository.InfraProfileConfigurationEntity, 0)
	cpuLimitValue, unitType, err := parseCPUorMemoryValue[unitsBean.CPUUnitStr](infraConfig.CiLimitCpu)
	if err != nil {
		return defaultConfigurations, err
	}
	cpuLimitParsedValue, err := impl.getCPUClient().ParseValAndUnit(cpuLimitValue, unitType.GetUnitSuffix())
	if err != nil {
		return defaultConfigurations, err
	}
	cpuLimit := adapter.NewInfraProfileConfigEntity(v1.CPU_LIMIT, profileId, platformName, cpuLimitParsedValue)
	defaultConfigurations = append(defaultConfigurations, cpuLimit)

	cpuReqValue, unitType, err := parseCPUorMemoryValue[unitsBean.CPUUnitStr](infraConfig.CiReqCpu)
	if err != nil {
		return defaultConfigurations, err
	}
	cpuReqParsedValue, err := impl.getCPUClient().ParseValAndUnit(cpuReqValue, unitType.GetUnitSuffix())
	if err != nil {
		return defaultConfigurations, err
	}
	cpuReq := adapter.NewInfraProfileConfigEntity(v1.CPU_REQUEST, profileId, platformName, cpuReqParsedValue)
	defaultConfigurations = append(defaultConfigurations, cpuReq)
	return defaultConfigurations, nil
}

func (impl *cpuClientImpl) getValueFromString(valueString string) (float64, int, error) {
	// Convert string to float64 and truncate to 2 decimal places
	valueFloat, err := strconv.ParseFloat(valueString, 64)
	if err != nil {
		return 0, 0, err
	}
	truncateValue := globalUtil.TruncateFloat(valueFloat, 2)
	return truncateValue, impl.getValueCount(truncateValue), nil // Returning float64 for resource values
}

func (impl *cpuClientImpl) overrideInfraConfig(infraConfiguration *v1.InfraConfig, configurationBean *v1.ConfigurationBean) (*v1.InfraConfig, error) {
	cpuConfigData, err := impl.getValueFromBean(configurationBean)
	if err != nil {
		return infraConfiguration, err
	}
	var cpuInfraConfigData string
	if configurationBean.Unit == unitsBean.CORE.String() {
		cpuInfraConfigData = fmt.Sprintf("%v", cpuConfigData)
	} else {
		cpuInfraConfigData = fmt.Sprintf("%v%v", cpuConfigData, configurationBean.Unit)
	}
	if configurationBean.Key == v1.CPU_REQUEST {
		infraConfiguration = infraConfiguration.SetCiReqCpu(cpuInfraConfigData)
	} else if configurationBean.Key == v1.CPU_LIMIT {
		infraConfiguration = infraConfiguration.SetCiLimitCpu(cpuInfraConfigData)
	} else {
		errMsg := fmt.Sprintf("invalid key %q for cpu configuration", configurationBean.Key)
		return infraConfiguration, util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	return infraConfiguration, nil
}

func (impl *cpuClientImpl) getValueFromBean(configurationBean *v1.ConfigurationBean) (float64, error) {
	if configurationBean == nil {
		return 0, nil
	}
	valueString, err := impl.formatTypedValueAsString(configurationBean.Value)
	if err != nil {
		return 0, err
	}
	cpuConfigData, _, err := impl.getValueFromString(valueString)
	if err != nil {
		impl.logger.Errorw("error in getting config data", "error", err, "config", configurationBean)
		return 0, err
	}
	return cpuConfigData, nil
}

func (impl *cpuClientImpl) formatTypedValueAsString(configValue any) (string, error) {
	var valueFloat float64
	// Handle string input or directly as float64
	switch v := configValue.(type) {
	case float64:
		valueFloat = v
	default:
		errMsg := fmt.Sprintf("invalid value for cpu configuration: %v", configValue)
		return "", util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
	}
	// Truncate and format the float value
	truncateValue := globalUtil.TruncateFloat(valueFloat, 2)
	return strconv.FormatFloat(truncateValue, 'f', -1, 64), nil
}

func (impl *cpuClientImpl) getValueCount(value float64) int {
	if reflect.ValueOf(value).IsZero() {
		return 0
	}
	return 1
}

func (impl *cpuClientImpl) isConfigActive(valueCount int, isConfigActive bool) bool {
	return isConfigActive
}

func (impl *cpuClientImpl) handlePostCreateOperations(tx *pg.Tx, createdInfraConfig *repository.InfraProfileConfigurationEntity) error {
	return nil
}

func (impl *cpuClientImpl) handlePostUpdateOperations(tx *pg.Tx, updatedInfraConfig *repository.InfraProfileConfigurationEntity) error {
	return nil
}

func (impl *cpuClientImpl) handlePostDeleteOperations(tx *pg.Tx, deletedInfraConfig *repository.InfraProfileConfigurationEntity) error {
	return nil
}

func (impl *cpuClientImpl) handleInfraConfigTriggerAudit(workflowId int, triggeredBy int32, infraConfig *v1.InfraConfig) error {
	return nil
}

func (impl *cpuClientImpl) resolveScopeVariablesForAppliedConfiguration(scope resourceQualifiers.Scope, configuration *v1.ConfigurationBean) (*v1.ConfigurationBean, map[string]string, error) {
	return configuration, nil, nil
}
