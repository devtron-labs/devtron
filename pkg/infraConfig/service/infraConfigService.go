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

package service

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/infraConfig/adapter"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/constants"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	util2 "github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type InfraConfigService interface {

	// GetConfigurationUnits fetches all the units for the configurations.
	GetConfigurationUnits() map[constants.ConfigKeyStr]map[string]units.Unit
	// GetProfileByName fetches the profile and its configurations matching the given profileName.
	GetProfileByName(name string) (*bean.ProfileBeanDto, error)
	// UpdateProfile updates the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	UpdateProfile(userId int32, profileName string, profileBean *bean.ProfileBeanDto) error

	GetInfraConfigurationsByScopeAndPlatform(scope bean.Scope, platform string) (*bean.InfraConfig, error)
}

type InfraConfigServiceImpl struct {
	logger           *zap.SugaredLogger
	infraProfileRepo repository.InfraConfigRepository
	appService       app.AppService
	units            *units.Units
	infraConfig      *bean.InfraConfig
}

func NewInfraConfigServiceImpl(logger *zap.SugaredLogger,
	infraProfileRepo repository.InfraConfigRepository,
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

func (impl *InfraConfigServiceImpl) GetProfileByName(name string) (*bean.ProfileBeanDto, error) {
	infraProfile, err := impl.infraProfileRepo.GetProfileByName(name)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "error", err)
		return nil, err
	}

	profileBean := adapter.ConvertToProfileBean(infraProfile)
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileId(infraProfile.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return nil, err
	}

	configurationBeans, err := adapter.ConvertToPlatformMap(infraConfigurations, profileBean.Name)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return nil, err
	}

	profileBean.Configurations = configurationBeans
	appCount, err := impl.appService.GetActiveCiCdAppsCount()
	if err != nil {
		impl.logger.Errorw("error in fetching app count for default profile", "error", err)
		return nil, err
	}
	profileBean.AppCount = appCount
	return &profileBean, nil
}

func (impl *InfraConfigServiceImpl) UpdateProfile(userId int32, profileName string, profileToUpdate *bean.ProfileBeanDto) error {
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

	infraProfileEntity := adapter.ConvertToInfraProfileEntity(profileToUpdate)
	// user couldn't delete the profile, always set this to active
	infraProfileEntity.Active = true

	infraConfigurations := adapter.ConvertFromPlatformMap(profileToUpdate.Configurations, defaultProfile, userId)

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update profile", "profileBean", profileToUpdate, "error", err)
		return err
	}
	defer func(infraProfileRepo repository.InfraConfigRepository, tx *pg.Tx) {
		err := infraProfileRepo.RollbackTx(tx)
		if err != nil {
			impl.logger.Errorw("error in rolling back transaction to update profile", "error", err)
		}
	}(impl.infraProfileRepo, tx)

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

	profile, err := impl.infraProfileRepo.GetProfileByName(constants.GLOBAL_PROFILE_NAME)
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
	defer func(infraProfileRepo repository.InfraConfigRepository, tx *pg.Tx) {
		err := infraProfileRepo.RollbackTx(tx)
		if err != nil {
			impl.logger.Errorw("error in rolling back transaction to save default configurations", "error", err)
		}
	}(impl.infraProfileRepo, tx)

	if profileCreationRequired {
		// if default profiles not found then create default profile
		defaultProfile := &repository.InfraProfileEntity{
			Name:        constants.GLOBAL_PROFILE_NAME,
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

	defaultConfigurationsFromEnv, err := adapter.LoadInfraConfigInEntities(impl.infraConfig)
	if err != nil {
		impl.logger.Errorw("error in loading default configurations from environment", "error", err)
		return err
	}

	// get db configurations and create new entries if db is out of sync
	defaultConfigurationsFromDB, err := impl.infraProfileRepo.GetConfigurationsByProfileName(constants.GLOBAL_PROFILE_NAME)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return err
	}
	defaultConfigurationsFromDBMap := make(map[constants.ConfigKey]bool)
	for _, defaultConfigurationFromDB := range defaultConfigurationsFromDB {
		defaultConfigurationsFromDBMap[defaultConfigurationFromDB.Key] = true
	}

	creatableConfigurations := make([]*repository.InfraProfileConfigurationEntity, 0, len(defaultConfigurationsFromEnv))
	creatableProfilePlatformMapping := make([]*repository.ProfilePlatformMapping, 0)
	for _, configurationFromEnv := range defaultConfigurationsFromEnv {
		if ok, exist := defaultConfigurationsFromDBMap[configurationFromEnv.Key]; !exist || !ok {
			configurationFromEnv.ProfileId = profile.Id
			configurationFromEnv.Active = true
			configurationFromEnv.Platform = constants.DEFAULT_PLATFORM
			configurationFromEnv.AuditLog = sql.NewDefaultAuditLog(1)
			creatableConfigurations = append(creatableConfigurations, configurationFromEnv)
		}
	}

	_, err = impl.infraProfileRepo.GetPlatformListByProfileName(constants.GLOBAL_PROFILE_NAME)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching platforms from db", "error", err)
		return err
	}
	//creating default platform if not found in db
	if errors.Is(err, pg.ErrNoRows) {
		creatableProfilePlatformMapping = append(creatableProfilePlatformMapping, &repository.ProfilePlatformMapping{
			Platform:  constants.DEFAULT_PLATFORM,
			ProfileId: profile.Id,
			Active:    true,
		})
	}

	if len(creatableConfigurations) > 0 {
		err = impl.infraProfileRepo.CreateConfigurations(tx, creatableConfigurations)
		if err != nil {
			impl.logger.Errorw("error in saving default configurations", "configurations", creatableConfigurations, "error", err)
			return err
		}
	}

	if len(creatableProfilePlatformMapping) > 0 {
		err = impl.infraProfileRepo.CreatePlatformProfileMapping(tx, creatableProfilePlatformMapping)
		if err != nil {
			impl.logger.Errorw("error in saving default configurations", "error", err)
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

	defaultConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileName(constants.GLOBAL_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "scope", scope, "error", err)
		return nil, err
	}

	defaultConfigurationsMap, err := adapter.ConvertToPlatformMap(defaultConfigurations, constants.GLOBAL_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "scope", scope, "error", err)
		return nil, err
	}
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
		case constants.CPU_LIMIT:
			infraConfiguration.SetCiLimitCpu(impl.getResolvedValue(config).(string))
		case constants.CPU_REQUEST:
			infraConfiguration.SetCiReqCpu(impl.getResolvedValue(config).(string))
		case constants.MEMORY_LIMIT:
			infraConfiguration.SetCiLimitMem(impl.getResolvedValue(config).(string))
		case constants.MEMORY_REQUEST:
			infraConfiguration.SetCiReqMem(impl.getResolvedValue(config).(string))
		case constants.TIME_OUT:
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
	if configurationBean.Key == util2.GetConfigKeyStr(constants.TimeOutKey) {
		timeout := configurationBean.Value.(float64)
		//timeout, _ := strconv.ParseFloat(configurationBean.Value.(float64), 64)
		// if user ever gives the timeout in float, after conversion to int64 it will be rounded off
		timeUnit := units.TimeUnitStr(configurationBean.Unit)
		return int64(timeout * impl.units.GetTimeUnits()[timeUnit].ConversionFactor)
	}
	if configurationBean.Unit == string(units.CORE) || configurationBean.Unit == string(units.BYTE) {
		return fmt.Sprintf("%v", configurationBean.Value.(float64))
	}
	return fmt.Sprintf("%v%v", configurationBean.Value.(float64), configurationBean.Unit)
}

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() map[constants.ConfigKeyStr]map[string]units.Unit {
	configurationUnits := make(map[constants.ConfigKeyStr]map[string]units.Unit)
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

	configurationUnits[constants.CPU_REQUEST] = cpuUnits
	configurationUnits[constants.CPU_LIMIT] = cpuUnits

	configurationUnits[constants.MEMORY_REQUEST] = memUnits
	configurationUnits[constants.MEMORY_LIMIT] = memUnits

	configurationUnits[constants.TIME_OUT] = timeUnits

	return configurationUnits
}

func (impl *InfraConfigServiceImpl) Validate(profileToUpdate *bean.ProfileBeanDto, defaultProfile *bean.ProfileBeanDto) error {
	err := util2.ValidatePayloadConfig(profileToUpdate)
	if err != nil {
		return err
	}

	err = impl.validateInfraConfig(profileToUpdate, defaultProfile)
	if err != nil {
		err = errors.Wrap(err, constants.PayloadValidationError)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validateInfraConfig(profileBean *bean.ProfileBeanDto, defaultProfile *bean.ProfileBeanDto) error {

	// currently validating cpu and memory limits and reqs only
	var (
		cpuLimit *bean.ConfigurationBean
		cpuReq   *bean.ConfigurationBean
		memLimit *bean.ConfigurationBean
		memReq   *bean.ConfigurationBean
		timeout  *bean.ConfigurationBean
	)

	for _, platformConfigurations := range profileBean.Configurations {
		for _, configuration := range platformConfigurations {
			// get cpu limit and req
			switch configuration.Key {
			case constants.CPU_LIMIT:
				cpuLimit = configuration
			case constants.CPU_REQUEST:
				cpuReq = configuration
			case constants.MEMORY_LIMIT:
				memLimit = configuration
			case constants.MEMORY_REQUEST:
				memReq = configuration
			case constants.TIME_OUT:
				timeout = configuration
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

	err = impl.validateTimeOut(timeout)
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
		return errors.New(fmt.Sprintf(constants.InvalidUnit, cpuLimit.Unit, cpuLimit.Key))
	}
	cpuReqUnit, ok := cpuUnits[cpuReqUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(constants.InvalidUnit, cpuReq.Unit, cpuReq.Key))
	}

	cpuLimitInterfaceVal, err := util2.GetTypedValue(cpuLimit.Key, cpuLimit.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(constants.InvalidTypeValue, cpuLimit.Key, cpuLimit.Value))
	}
	cpuLimitVal, ok := cpuLimitInterfaceVal.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(constants.InvalidTypeValue, cpuLimit.Key, cpuLimit.Value))
	}

	cpuReqInterfaceVal, err := util2.GetTypedValue(cpuReq.Key, cpuReq.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(constants.InvalidTypeValue, cpuReq.Key, cpuReq.Value))
	}
	cpuReqVal, ok := cpuReqInterfaceVal.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(constants.InvalidTypeValue, cpuReq.Key, cpuReq.Value))
	}
	if !validLimReq(cpuLimitVal, cpuLimitUnit.ConversionFactor, cpuReqVal, cpuReqUnit.ConversionFactor) {
		return errors.New(constants.CPULimReqErrorCompErr)
	}
	return nil
}
func (impl *InfraConfigServiceImpl) validateTimeOut(timeOut *bean.ConfigurationBean) error {
	if timeOut == nil {
		return nil
	}
	timeoutUnitSuffix := units.TimeUnitStr(timeOut.Unit)
	timeUnits := impl.units.GetTimeUnits()
	_, ok := timeUnits[timeoutUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(constants.InvalidUnit, timeOut.Unit, timeOut.Key))
	}
	timeout, err := util2.GetTypedValue(timeOut.Key, timeOut.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(constants.InvalidTypeValue, timeOut.Key, timeOut.Value))
	}
	_, ok = timeout.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(constants.InvalidTypeValue, timeOut.Key, timeOut.Value))
	}
	return nil
}
func (impl *InfraConfigServiceImpl) validateMEM(memLimit, memReq *bean.ConfigurationBean) error {
	memLimitUnitSuffix := units.MemoryUnitStr(memLimit.Unit)
	memReqUnitSuffix := units.MemoryUnitStr(memReq.Unit)
	memUnits := impl.units.GetMemoryUnits()
	memLimitUnit, ok := memUnits[memLimitUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(constants.InvalidUnit, memLimit.Unit, memLimit.Key))
	}
	memReqUnit, ok := memUnits[memReqUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(constants.InvalidUnit, memReq.Unit, memReq.Key))
	}

	// Use getTypedValue to retrieve appropriate types
	memLimitInterfaceVal, err := util2.GetTypedValue(memLimit.Key, memLimit.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(constants.InvalidTypeValue, memLimit.Key, memLimit.Value))
	}
	memLimitVal, ok := memLimitInterfaceVal.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(constants.InvalidTypeValue, memLimit.Key, memLimit.Value))
	}

	memReqInterfaceVal, err := util2.GetTypedValue(memReq.Key, memReq.Value)
	if err != nil {
		return errors.New(fmt.Sprintf(constants.InvalidTypeValue, memReq.Key, memReq.Value))
	}

	memReqVal, ok := memReqInterfaceVal.(float64)
	if !ok {
		return errors.New(fmt.Sprintf(constants.InvalidTypeValue, memReq.Key, memReq.Value))
	}

	if !validLimReq(memLimitVal, memLimitUnit.ConversionFactor, memReqVal, memReqUnit.ConversionFactor) {
		return errors.New(constants.MEMLimReqErrorCompErr)
	}
	return nil
}

func validLimReq(lim, limFactor, req, reqFactor float64) bool {
	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRatio := lim / req
	convFactor := limFactor / reqFactor
	return limitToReqRatio*convFactor >= 1
}
