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

package infraConfig

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/infraConfig/adapter"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	util2 "github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type InfraConfigService interface {

	// GetConfigurationUnits fetches all the units for the configurations.
	GetConfigurationUnits() map[util2.ConfigKeyStr]map[string]units.Unit
	// GetProfileByName fetches the profile and its configurations matching the given profileName.
	GetProfileByName(name string) (*bean.ProfileBean, error)
	// UpdateProfile updates the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	UpdateProfile(userId int32, profileName string, profileBean *bean.ProfileBean) error

	GetInfraConfigurationsByScopeAndPlatform(scope bean.Scope, platform string) (*bean.InfraConfig, error)
}

type InfraConfigServiceImpl struct {
	logger           *zap.SugaredLogger
	infraProfileRepo InfraConfigRepository
	appService       app.AppService
	units            *units.Units
	infraConfig      *bean.InfraConfig
}

func NewInfraConfigServiceImpl(logger *zap.SugaredLogger,
	infraProfileRepo InfraConfigRepository,
	appService app.AppService,
	units *units.Units) (*InfraConfigServiceImpl, error) {
	infraConfiguration := &bean.InfraConfig{}
	err := env.Parse(infraConfiguration)
	if err != nil {
		return nil, err
	}
	infraProfileService := &InfraConfigServiceImpl{
		logger:           logger,
		infraProfileRepo: infraProfileRepo,
		appService:       appService,
		units:            units,
		infraConfig:      infraConfiguration,
	}
	err = infraProfileService.loadDefaultProfile()
	return infraProfileService, err
}

func (impl *InfraConfigServiceImpl) GetProfileByName(name string) (*bean.ProfileBean, error) {
	infraProfile, err := impl.infraProfileRepo.GetProfileByName(name)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "error", err)
		return nil, err
	}

	profileBean := infraProfile.ConvertToProfileBean()
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileId(infraProfile.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return nil, err
	}

	configurationBeans := adapter.ConvertToPlatformMap(infraConfigurations, profileBean.Name)

	profileBean.Configurations = configurationBeans
	appCount, err := impl.appService.GetActiveCiCdAppsCount()
	if err != nil {
		impl.logger.Errorw("error in fetching app count for default profile", "error", err)
		return nil, err
	}
	profileBean.AppCount = appCount
	return &profileBean, nil
}

func (impl *InfraConfigServiceImpl) UpdateProfile(userId int32, profileName string, profileToUpdate *bean.ProfileBean) error {
	// validation
	defaultProfile, err := impl.GetProfileByName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "profileName", profileName, "profileCreateRequest", profileToUpdate, "error", err)
		return err
	}
	if err = impl.Validate(profileToUpdate, defaultProfile); err != nil {
		impl.logger.Errorw("error occurred in validation the profile create request", "profileName", profileName, "profileCreateRequest", profileToUpdate, "error", err)
		return err
	}
	// validations end

	infraProfileEntity := profileToUpdate.ConvertToInfraProfileEntity()
	// user couldn't delete the profile, always set this to active
	infraProfileEntity.Active = true

	infraConfigurations := adapter.ConvertFromPlatformMap(profileToUpdate.Configurations, defaultProfile, userId)

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update profile", "profileBean", profileToUpdate, "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)
	infraProfileEntity.UpdatedOn = time.Now()
	infraProfileEntity.UpdatedBy = userId
	err = impl.infraProfileRepo.UpdateProfile(tx, profileName, infraProfileEntity)
	if err != nil {
		impl.logger.Errorw("error in updating profile", "error", "profileName", profileName, "profileCreateRequest", profileToUpdate, err)
		return err
	}

	err = impl.infraProfileRepo.UpdateConfigurations(tx, infraConfigurations)
	if err != nil {
		impl.logger.Errorw("error in creating configurations", "error", "profileName", profileName, "profileCreateRequest", profileToUpdate, err)
		return err
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update profile", "profileName", profileName, "profileCreateRequest", profileToUpdate, "error", err)
	}
	return err
}

// loadDefaultProfile loads default configurations from environment and save them in db.
// this will only create the default profile only once if not exists in db.(container restarts won't create new default profile everytime)
// this will load the default configurations provided in InfraConfig. if db is in out of sync with InfraConfig then it will create new entries for those missing configurations in db.
func (impl *InfraConfigServiceImpl) loadDefaultProfile() error {

	profile, err := impl.infraProfileRepo.GetProfileByName(util2.DEFAULT_PROFILE_NAME)
	// make sure about no rows error
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		return err
	}
	profileCreationRequired := errors.Is(err, pg.ErrNoRows)
	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)
	if profileCreationRequired {
		// if default profiles not found then create default profile
		defaultProfile := &bean.InfraProfileEntity{
			Name:        util2.DEFAULT_PROFILE_NAME,
			Description: "",
			Active:      true,
			AuditLog:    sql.NewDefaultAuditLog(1),
		}

		err = impl.infraProfileRepo.CreateProfile(tx, defaultProfile)
		if err != nil {
			impl.logger.Errorw("error in saving default profile", "error", err)
			return err
		}
		profile = defaultProfile
	}

	defaultConfigurationsFromEnv, err := impl.infraConfig.LoadInfraConfigInEntities()
	if err != nil {
		impl.logger.Errorw("error in loading default configurations from environment", "error", err)
		return err
	}

	// get db configurations and create new entries if db is out of sync
	defaultConfigurationsFromDB, err := impl.infraProfileRepo.GetConfigurationsByProfileName(util2.DEFAULT_PROFILE_NAME)
	// todo: check the error logic here
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return err
	}
	defaultConfigurationsFromDBMap := make(map[util2.ConfigKey]bool)
	updatableConfigurations := make([]*bean.InfraProfileConfigurationEntity, 0, len(defaultConfigurationsFromEnv))
	for _, defaultConfigurationFromDB := range defaultConfigurationsFromDB {
		defaultConfigurationsFromDBMap[defaultConfigurationFromDB.Key] = true
		if defaultConfigurationFromDB.Platform == "" {
			defaultConfigurationFromDB.Platform = util2.CI_RUNNER_PLATFORM
			updatableConfigurations = append(updatableConfigurations, defaultConfigurationFromDB)
		}
	}
	if len(updatableConfigurations) > 0 {
		err = impl.infraProfileRepo.UpdateConfigurations(tx, updatableConfigurations)
		if err != nil {
			impl.logger.Errorw("error in updating default configurations", "configurations", updatableConfigurations, "error", err)
			return err
		}
	}

	creatableConfigurations := make([]*bean.InfraProfileConfigurationEntity, 0, len(defaultConfigurationsFromEnv))
	for _, configurationFromEnv := range defaultConfigurationsFromEnv {
		if !defaultConfigurationsFromDBMap[configurationFromEnv.Key] {
			configurationFromEnv.ProfileId = profile.Id
			configurationFromEnv.Active = true
			configurationFromEnv.AuditLog = sql.NewDefaultAuditLog(1)
			creatableConfigurations = append(creatableConfigurations, configurationFromEnv)
		}
	}

	if len(creatableConfigurations) > 0 {
		err = impl.infraProfileRepo.CreateConfigurations(tx, creatableConfigurations)
		if err != nil {
			impl.logger.Errorw("error in saving default configurations", "configurations", creatableConfigurations, "error", err)
			return err
		}
	}

	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to save default configurations", "error", err)
	}
	return err
}

func (impl *InfraConfigServiceImpl) GetInfraConfigurationsByScopeAndPlatform(scope bean.Scope, platform string) (*bean.InfraConfig, error) {

	defaultConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileName(util2.DEFAULT_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "scope", scope, "error", err)
		return nil, err
	}

	defaultConfigurationsMap := adapter.ConvertToPlatformMap(defaultConfigurations, util2.DEFAULT_PROFILE_NAME)

	platformConfigurationBean := defaultConfigurationsMap[platform]
	if platformConfigurationBean == nil {
		return &bean.InfraConfig{}, nil
	}

	return impl.getInfraConfigForConfigBean(platformConfigurationBean), nil
}

func (impl *InfraConfigServiceImpl) getInfraConfigForConfigBean(platformConfigurationBean []*bean.ConfigurationBean) *bean.InfraConfig {
	infraConfiguration := &bean.InfraConfig{}
	overrideInfraConfigFunc := func(config bean.ConfigurationBean) {
		switch config.Key {
		case util2.CPU_LIMIT:
			infraConfiguration.SetCiLimitCpu(impl.getResolvedValue(config).(string))
		case util2.CPU_REQUEST:
			infraConfiguration.SetCiReqCpu(impl.getResolvedValue(config).(string))
		case util2.MEMORY_LIMIT:
			infraConfiguration.SetCiLimitMem(impl.getResolvedValue(config).(string))
		case util2.MEMORY_REQUEST:
			infraConfiguration.SetCiReqMem(impl.getResolvedValue(config).(string))
		case util2.TIME_OUT:
			infraConfiguration.SetCiDefaultTimeout(impl.getResolvedValue(config).(int64))
		}
	}
	for _, defaultConfigurationBean := range platformConfigurationBean {
		overrideInfraConfigFunc(*defaultConfigurationBean)
	}
	return infraConfiguration
}

func (impl *InfraConfigServiceImpl) getResolvedValue(configurationBean bean.ConfigurationBean) interface{} {
	// for timeout we need to get the value in seconds
	if configurationBean.Key == util2.GetConfigKeyStr(util2.TimeOut) {
		timeout, _ := strconv.ParseFloat(configurationBean.Value, 64)
		// if user ever gives the timeout in float, after conversion to int64 it will be rounded off
		timeUnit := units.TimeUnitStr(configurationBean.Unit)
		return int64(timeout * impl.units.GetTimeUnits()[timeUnit].ConversionFactor)
	}
	if configurationBean.Unit == string(units.CORE) || configurationBean.Unit == string(units.BYTE) {
		return fmt.Sprintf("%v", configurationBean.Value)
	}
	return fmt.Sprintf("%v%v", configurationBean.Value, configurationBean.Unit)
}

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() map[util2.ConfigKeyStr]map[string]units.Unit {
	configurationUnits := make(map[util2.ConfigKeyStr]map[string]units.Unit)
	cpuUnits := make(map[string]units.Unit)
	memUnits := make(map[string]units.Unit)
	timeUnits := make(map[string]units.Unit)
	for key, val := range impl.units.GetCpuUnits() {
		cpuUnits[string(key)] = val
	}
	for key, val := range impl.units.GetMemoryUnits() {
		memUnits[string(key)] = val
	}
	for key, val := range impl.units.GetTimeUnits() {
		timeUnits[string(key)] = val
	}

	configurationUnits[util2.CPU_REQUEST] = cpuUnits
	configurationUnits[util2.CPU_LIMIT] = cpuUnits

	configurationUnits[util2.MEMORY_REQUEST] = memUnits
	configurationUnits[util2.MEMORY_LIMIT] = memUnits

	configurationUnits[util2.TIME_OUT] = timeUnits

	return configurationUnits
}

func (impl *InfraConfigServiceImpl) Validate(profileToUpdate *bean.ProfileBean, defaultProfile *bean.ProfileBean) error {
	var err error = nil
	defaultConfigurationsKeyMap := util2.GetDefaultConfigKeysMap()
	for _, platformConfiguration := range profileToUpdate.Configurations {
		// validate configurations only contain default configurations types.(cpu_limit,cpu_request,mem_limit,mem_request,timeout)
		for _, propertyConfig := range platformConfiguration {
			if _, ok := defaultConfigurationsKeyMap[propertyConfig.Key]; !ok {
				errorMsg := fmt.Sprintf("invalid configuration property \"%s\"", propertyConfig.Key)
				if err == nil {
					err = errors.New(errorMsg)
				}
				err = errors.Wrap(err, errorMsg)
			}
		}
	}

	if err != nil {
		err = errors.Wrap(err, util2.PayloadValidationError)
		return err
	}

	err = impl.validateCpuMem(profileToUpdate, defaultProfile)
	if err != nil {
		err = errors.Wrap(err, util2.PayloadValidationError)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validateCpuMem(profileBean *bean.ProfileBean, defaultProfile *bean.ProfileBean) error {

	// currently validating cpu and memory limits and reqs only
	var (
		cpuLimit *bean.ConfigurationBean
		cpuReq   *bean.ConfigurationBean
		memLimit *bean.ConfigurationBean
		memReq   *bean.ConfigurationBean
	)

	for _, platformConfigurations := range profileBean.Configurations {
		for _, configuration := range platformConfigurations {
			// get cpu limit and req
			switch configuration.Key {
			case util2.CPU_LIMIT:
				cpuLimit = configuration
			case util2.CPU_REQUEST:
				cpuReq = configuration
			case util2.MEMORY_LIMIT:
				memLimit = configuration
			case util2.MEMORY_REQUEST:
				memReq = configuration
			}
		}
	}

	// validate cpu
	err := impl.validateCPU(cpuLimit, cpuReq)
	if err != nil {
		return err
	}
	// validate mem
	err = impl.validateMEM(memLimit, memReq)
	if err != nil {
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validateCPU(cpuLimit, cpuReq *bean.ConfigurationBean) error {
	cpuLimitUnitSuffix := units.CPUUnitStr(cpuLimit.Unit)
	cpuReqUnitSuffix := units.CPUUnitStr(cpuReq.Unit)
	cpuUnits := impl.units.GetCpuUnits()
	cpuLimitUnit, ok := cpuUnits[cpuLimitUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(util2.InvalidUnit, cpuLimit.Unit, cpuLimit.Key))
	}
	cpuReqUnit, ok := cpuUnits[cpuReqUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(util2.InvalidUnit, cpuReq.Unit, cpuReq.Key))
	}

	cpuLimitVal, _ := strconv.ParseFloat(cpuLimit.Value, 64)
	cpuReqVal, _ := strconv.ParseFloat(cpuReq.Value, 64)
	if !validLimReq(cpuLimitVal, cpuLimitUnit.ConversionFactor, cpuReqVal, cpuReqUnit.ConversionFactor) {
		return errors.New(util2.CPULimReqErrorCompErr)
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validateMEM(memLimit, memReq *bean.ConfigurationBean) error {
	memLimitUnitSuffix := units.MemoryUnitStr(memLimit.Unit)
	memReqUnitSuffix := units.MemoryUnitStr(memReq.Unit)
	memUnits := impl.units.GetMemoryUnits()
	memLimitUnit, ok := memUnits[memLimitUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(util2.InvalidUnit, memLimit.Unit, memLimit.Key))
	}
	memReqUnit, ok := memUnits[memReqUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(util2.InvalidUnit, memReq.Unit, memReq.Key))
	}

	memLimitVal, _ := strconv.ParseFloat(memLimit.Value, 64)
	memReqVal, _ := strconv.ParseFloat(memReq.Value, 64)
	if !validLimReq(memLimitVal, memLimitUnit.ConversionFactor, memReqVal, memReqUnit.ConversionFactor) {
		return errors.New(util2.MEMLimReqErrorCompErr)
	}
	return nil
}

func validLimReq(lim, limFactor, req, reqFactor float64) bool {
	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRatio := lim / req
	convFactor := limFactor / reqFactor
	return limitToReqRatio*convFactor >= 1
}
