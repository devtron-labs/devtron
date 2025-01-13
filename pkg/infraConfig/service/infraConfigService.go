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
	GetConfigurationUnits() map[bean.ConfigKeyStr]map[string]units.Unit
	// GetProfileByName fetches the profile and its configurations matching the given profileName.
	GetProfileByName(name string) (*bean.ProfileBeanDto, error)
	// UpdateProfile updates the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	UpdateProfile(userId int32, profileName string, profileBean *bean.ProfileBeanDto) error

	GetInfraConfigurationsByScopeAndPlatform(scope *bean.Scope, platform string) (*bean.InfraConfig, error)
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
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileId([]int{infraProfile.Id})
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
	if err = impl.validate(profileToUpdate, defaultProfile); err != nil {
		impl.logger.Errorw("error occurred in validation the profile create request", "profileName", profileName, "profileCreateRequest", profileToUpdate, "error", err)
		return err
	}
	// validations end

	profileToUpdate.Id = defaultProfile.Id
	infraProfileEntity := adapter.ConvertToInfraProfileEntity(profileToUpdate)
	// user couldn't delete the profile, always set this to active
	infraProfileEntity.Active = true
	//todo make it compatible with ent

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
	adapter.SetProfilePlatform(defaultProfile, infraConfigurations)
	err = impl.infraProfileRepo.UpdateConfigurations(tx, infraConfigurations)
	if err != nil {
		impl.logger.Errorw("error in creating configurations", "error", "profileName", profileName, "profileCreateRequest", profileToUpdate, err)
		return err
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update profile", "profileCreateRequest", profileToUpdate, "error", err)
		return err
	}
	return nil
}

// loadDefaultProfile loads default configurations from environment and save them in db.
// this will only create the default profile only once if not exists in db.(container restarts won't create new default profile everytime)
// this will load the default configurations provided in InfraConfig. if db is in out of sync with InfraConfig then it will create new entries for those missing configurations in db.
func (impl *InfraConfigServiceImpl) loadDefaultProfile() error {

	profile, err := impl.infraProfileRepo.GetProfileByName(bean.GLOBAL_PROFILE_NAME)
	// make sure about no rows error
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching default profile", "error", err)
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
			Name:        bean.GLOBAL_PROFILE_NAME,
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
	var nodeselector []string
	defaultConfigurationsFromEnv, err := adapter.LoadInfraConfigInEntities(impl.infraConfig, nodeselector, "", "")
	if err != nil {
		impl.logger.Errorw("error in loading default configurations from environment", "error", err)
		return err
	}

	// get db configurations and create new entries if db is out of sync
	defaultConfigurationsFromDB, err := impl.infraProfileRepo.GetConfigurationsByProfileName(bean.GLOBAL_PROFILE_NAME)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return err
	}
	defaultConfigurationsFromDBMap := make(map[bean.ConfigKey]bool)
	for _, defaultConfigurationFromDB := range defaultConfigurationsFromDB {
		defaultConfigurationsFromDBMap[defaultConfigurationFromDB.Key] = true
	}

	creatableConfigurations := make([]*repository.InfraProfileConfigurationEntity, 0, len(defaultConfigurationsFromEnv))
	creatableProfilePlatformMapping := make([]*repository.ProfilePlatformMapping, 0)

	platformsFromDb, err := impl.infraProfileRepo.GetPlatformListByProfileName(bean.GLOBAL_PROFILE_NAME)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching platforms from db", "error", err)
		return err
	}
	runnerPlatFormMapping := &repository.ProfilePlatformMapping{
		Platform:  bean.RUNNER_PLATFORM,
		ProfileId: profile.Id,
		Active:    true,
		AuditLog:  sql.NewDefaultAuditLog(1),
	}

	//creating default platform if not found in db
	if len(platformsFromDb) == 0 {
		creatableProfilePlatformMapping = append(creatableProfilePlatformMapping, runnerPlatFormMapping)
	}
	if len(creatableProfilePlatformMapping) > 0 {
		err = impl.infraProfileRepo.CreatePlatformProfileMapping(tx, creatableProfilePlatformMapping)
		if err != nil {
			impl.logger.Errorw("error in saving default configurations", "error", err)
			return err
		}
	}

	for _, configurationFromEnv := range defaultConfigurationsFromEnv {
		if ok, exist := defaultConfigurationsFromDBMap[configurationFromEnv.Key]; !exist || !ok {
			configurationFromEnv.ProfileId = profile.Id
			configurationFromEnv.Active = true
			configurationFromEnv.ProfilePlatformMappingId = runnerPlatFormMapping.Id
			configurationFromEnv.ProfilePlatformMapping.Platform = bean.RUNNER_PLATFORM
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

func (impl *InfraConfigServiceImpl) GetInfraConfigurationsByScopeAndPlatform(scope *bean.Scope, platform string) (*bean.InfraConfig, error) {

	defaultConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileName(bean.GLOBAL_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "scope", scope, "error", err)
		return nil, err
	}

	defaultConfigurationsMap, err := adapter.ConvertToPlatformMap(defaultConfigurations, bean.GLOBAL_PROFILE_NAME)
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
		case bean.CPU_LIMIT:
			infraConfiguration.SetCiLimitCpu(impl.getResolvedValue(config).(string))
		case bean.CPU_REQUEST:
			infraConfiguration.SetCiReqCpu(impl.getResolvedValue(config).(string))
		case bean.MEMORY_LIMIT:
			infraConfiguration.SetCiLimitMem(impl.getResolvedValue(config).(string))
		case bean.MEMORY_REQUEST:
			infraConfiguration.SetCiReqMem(impl.getResolvedValue(config).(string))
		case bean.TIME_OUT:
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
	if configurationBean.Key == util2.GetConfigKeyStr(bean.TimeOutKey) {
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

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() map[bean.ConfigKeyStr]map[string]units.Unit {
	configurationUnits := make(map[bean.ConfigKeyStr]map[string]units.Unit)
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

	configurationUnits[bean.CPU_REQUEST] = cpuUnits
	configurationUnits[bean.CPU_LIMIT] = cpuUnits

	configurationUnits[bean.MEMORY_REQUEST] = memUnits
	configurationUnits[bean.MEMORY_LIMIT] = memUnits

	configurationUnits[bean.TIME_OUT] = timeUnits

	return configurationUnits
}

func (impl *InfraConfigServiceImpl) validate(profileToUpdate *bean.ProfileBeanDto, defaultProfile *bean.ProfileBeanDto) error {

	err := impl.validateInfraConfig(profileToUpdate, defaultProfile)
	if err != nil {
		err = errors.Wrap(err, bean.PayloadValidationError)
		return err
	}
	return nil
}
