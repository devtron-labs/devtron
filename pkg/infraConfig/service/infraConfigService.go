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
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/infraConfig/adapter"
	"github.com/devtron-labs/devtron/pkg/infraConfig/bean"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	bean2 "github.com/devtron-labs/devtron/pkg/infraConfig/units/bean"
	util2 "github.com/devtron-labs/devtron/pkg/infraConfig/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type InfraConfigService interface {

	// GetConfigurationUnits fetches all the units for the configurations.
	GetConfigurationUnits() (map[bean.ConfigKeyStr]map[string]bean.Unit, error)
	// GetProfileByName fetches the profile and its configurations matching the given profileName.
	GetProfileByName(name string) (*bean.ProfileBeanDto, error)
	// UpdateProfile updates the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	UpdateProfile(userId int32, profileName string, profileBean *bean.ProfileBeanDto) error

	// GetInfraConfigurationsByScopeAndPlatform fetches the infra configurations for the given scope and platform.
	GetInfraConfigurationsByScopeAndPlatform(scope bean.Scope, platform string) (*bean.InfraConfig, error)
}

type InfraConfigServiceImpl struct {
	logger           *zap.SugaredLogger
	infraProfileRepo repository.InfraConfigRepository
	appService       app.AppService
	infraConfig      *bean.InfraConfig
	unitFactoryMap   map[units.PropertyType]units.UnitService
}

func NewInfraConfigServiceImpl(logger *zap.SugaredLogger,
	infraProfileRepo repository.InfraConfigRepository,
	appService app.AppService) (*InfraConfigServiceImpl, error) {
	infraConfiguration := &bean.InfraConfig{}
	err := env.Parse(infraConfiguration)
	if err != nil {
		return nil, err
	}
	cpuUnitFactory, err := units.NewUnitService(units.CPU, logger)
	if err != nil {
		logger.Errorw("error in creating cpu unit factory", "error", err)
		return nil, err
	}
	memUnitFactory, err := units.NewUnitService(units.MEMORY, logger)
	if err != nil {
		logger.Errorw("error in creating memory unit factory", "error", err)
		return nil, err
	}
	timeUnitFactory, err := units.NewUnitService(units.TIME, logger)
	if err != nil {
		logger.Errorw("error in creating time unit factory", "error", err)
		return nil, err
	}
	unitFactoryMap := make(map[units.PropertyType]units.UnitService)
	unitFactoryMap[units.CPU] = cpuUnitFactory
	unitFactoryMap[units.MEMORY] = memUnitFactory
	unitFactoryMap[units.TIME] = timeUnitFactory
	infraProfileService := &InfraConfigServiceImpl{
		logger:           logger,
		infraProfileRepo: infraProfileRepo,
		appService:       appService,
		infraConfig:      infraConfiguration,
		unitFactoryMap:   unitFactoryMap,
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
	if err = impl.validate(profileToUpdate, defaultProfile); err != nil {
		impl.logger.Errorw("error occurred in validation the profile create request", "profileName", profileName, "profileCreateRequest", profileToUpdate, "error", err)
		return err
	}
	// validations end

	infraProfileEntity := adapter.ConvertToInfraProfileEntity(profileToUpdate)
	// user couldn't delete the profile, always set this to active
	infraProfileEntity.Active = true
	// set default values, user can't change these values
	profileToUpdate.Id = defaultProfile.Id
	profileToUpdate.Name = defaultProfile.Name
	infraConfigurations := adapter.ConvertFromPlatformMap(profileToUpdate, userId)

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

	defaultConfigurationsFromEnv, err := impl.loadInfraConfigInEntities(impl.infraConfig)
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
	for _, configurationFromEnv := range defaultConfigurationsFromEnv {
		if ok, exist := defaultConfigurationsFromDBMap[configurationFromEnv.Key]; !exist || !ok {
			configurationFromEnv.ProfileId = profile.Id
			configurationFromEnv.Active = true
			configurationFromEnv.Platform = bean.DEFAULT_PLATFORM
			configurationFromEnv.AuditLog = sql.NewDefaultAuditLog(1)
			creatableConfigurations = append(creatableConfigurations, configurationFromEnv)
		}
	}

	_, err = impl.infraProfileRepo.GetPlatformListByProfileName(bean.GLOBAL_PROFILE_NAME)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error in fetching platforms from db", "error", err)
		return err
	}
	//creating default platform if not found in db
	if errors.Is(err, pg.ErrNoRows) {
		creatableProfilePlatformMapping = append(creatableProfilePlatformMapping, &repository.ProfilePlatformMapping{
			Platform:  bean.DEFAULT_PLATFORM,
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
	defaultConfigurationsDB, err := impl.infraProfileRepo.GetConfigurationsByProfileName(bean.GLOBAL_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "scope", scope, "error", err)
		return nil, err
	}
	defaultConfigurationBeansMap, err := adapter.ConvertToPlatformMap(defaultConfigurationsDB, bean.GLOBAL_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in converting default configurations into platform map", "defaultConfigurationsDB", defaultConfigurationsDB, "error", err)
		return nil, err
	}
	if platformConfigurationBean, ok := defaultConfigurationBeansMap[platform]; !ok {
		impl.logger.Debugw("platform not found in default configurations", "platform", platform)
		return &bean.InfraConfig{}, nil
	} else if platformConfigurationBean != nil {
		return impl.getInfraConfigForConfigBean(platformConfigurationBean)
	} else {
		impl.logger.Debugw("platform found with null configurations", "platform", platform)
		return &bean.InfraConfig{}, nil
	}
}

func (impl *InfraConfigServiceImpl) getInfraConfigForConfigBean(platformConfigurationBean []*bean.ConfigurationBean) (infraConfiguration *bean.InfraConfig, err error) {
	infraConfiguration = &bean.InfraConfig{}
	overrideInfraConfigFunc := func(config bean.ConfigurationBean, infraConfiguration *bean.InfraConfig) (*bean.InfraConfig, error) {
		switch config.Key {
		case bean.CPU_LIMIT:
			cpuLimit, err := impl.getResolvedValue(config)
			if err != nil {
				impl.logger.Errorw("error in getting cpu limit value", "config", config, "error", err)
				return infraConfiguration, util.NewApiError(http.StatusBadRequest, err.Error(), err.Error())
			}
			infraConfiguration.SetCiLimitCpu(cpuLimit.(string))
		case bean.CPU_REQUEST:
			cpuReq, err := impl.getResolvedValue(config)
			if err != nil {
				impl.logger.Errorw("error in getting cpu request value", "config", config, "error", err)
				return infraConfiguration, util.NewApiError(http.StatusBadRequest, err.Error(), err.Error())
			}
			infraConfiguration.SetCiReqCpu(cpuReq.(string))
		case bean.MEMORY_LIMIT:
			memoryLimit, err := impl.getResolvedValue(config)
			if err != nil {
				impl.logger.Errorw("error in getting memory limit value", "config", config, "error", err)
				return infraConfiguration, util.NewApiError(http.StatusBadRequest, err.Error(), err.Error())
			}
			infraConfiguration.SetCiLimitMem(memoryLimit.(string))
		case bean.MEMORY_REQUEST:
			memoryReq, err := impl.getResolvedValue(config)
			if err != nil {
				impl.logger.Errorw("error in getting memory request value", "config", config, "error", err)
				return infraConfiguration, util.NewApiError(http.StatusBadRequest, err.Error(), err.Error())
			}
			infraConfiguration.SetCiReqMem(memoryReq.(string))
		case bean.TIME_OUT:
			timeout, err := impl.getResolvedValue(config)
			if err != nil {
				impl.logger.Errorw("error in getting timeout value", "config", config, "error", err)
				return infraConfiguration, util.NewApiError(http.StatusBadRequest, err.Error(), err.Error())
			}
			timeoutInt, ok := timeout.(int64)
			if !ok {
				errMsg := fmt.Sprintf("invalid timeout value '%v'", timeout)
				return infraConfiguration, util.NewApiError(http.StatusBadRequest, errMsg, errMsg)
			}
			infraConfiguration.SetCiDefaultTimeout(timeoutInt)
		}
		return infraConfiguration, nil
	}
	for _, defaultConfigurationBean := range platformConfigurationBean {
		infraConfiguration, err = overrideInfraConfigFunc(*defaultConfigurationBean, infraConfiguration)
		if err != nil {
			return nil, err
		}
	}
	return infraConfiguration, nil
}

func (impl *InfraConfigServiceImpl) getResolvedValue(configurationBean bean.ConfigurationBean) (interface{}, error) {
	timeout, ok := configurationBean.Value.(float64)
	if !ok {
		impl.logger.Errorw("error in getting timeout value", "key", configurationBean.Key, "value", configurationBean.Value)
		errMsg := fmt.Sprintf(bean.InvalidTypeValue, configurationBean.Key, configurationBean.Value)
		return int64(0), errors.New(errMsg)
	}
	// for timeout, we need to get the value in seconds
	if configurationBean.Key == util2.GetConfigKeyStr(bean.TimeOutKey) {
		// timeout, _ := strconv.ParseFloat(configurationBean.Value.(float64), 64)
		//  if a user ever gives the timeout in float, after conversion to int64 it will be rounded off
		timeUnit, ok := bean2.TimeUnitStr(configurationBean.Unit).GetUnit()
		if !ok {
			impl.logger.Errorw("error in getting time unit", "unit", configurationBean.Unit)
			errMsg := fmt.Sprintf(bean.InvalidUnit, configurationBean.Unit, configurationBean.Key)
			return int64(timeout), errors.New(errMsg)
		}
		return int64(timeout * timeUnit.ConversionFactor), nil
	}
	if configurationBean.Unit == bean2.CORE.String() || configurationBean.Unit == bean2.BYTE.String() {
		return fmt.Sprintf("%v", configurationBean.Value.(float64)), nil
	}
	return fmt.Sprintf("%v%v", timeout, configurationBean.Unit), nil
}

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() (map[bean.ConfigKeyStr]map[string]bean.Unit, error) {
	configurationUnits := make(map[bean.ConfigKeyStr]map[string]bean.Unit)
	cpuUnits := impl.unitFactoryMap[units.CPU].GetAllUnits()
	memUnits := impl.unitFactoryMap[units.MEMORY].GetAllUnits()
	timeUnits := impl.unitFactoryMap[units.TIME].GetAllUnits()

	configurationUnits[bean.CPU_REQUEST] = cpuUnits
	configurationUnits[bean.CPU_LIMIT] = cpuUnits

	configurationUnits[bean.MEMORY_REQUEST] = memUnits
	configurationUnits[bean.MEMORY_LIMIT] = memUnits

	configurationUnits[bean.TIME_OUT] = timeUnits

	return configurationUnits, nil
}

func (impl *InfraConfigServiceImpl) validate(profileToUpdate *bean.ProfileBeanDto, defaultProfile *bean.ProfileBeanDto) error {
	err := util2.ValidatePayloadConfig(profileToUpdate)
	if err != nil {
		return err
	}

	err = impl.validateInfraConfig(profileToUpdate, defaultProfile)
	if err != nil {
		err = errors.Wrap(err, bean.PayloadValidationError)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) loadInfraConfigInEntities(infraConfig *bean.InfraConfig) ([]*repository.InfraProfileConfigurationEntity, error) {
	cpuLimitParsedValue, err := impl.unitFactoryMap[units.CPU].ParseValAndUnit(infraConfig.CiLimitCpu)
	if err != nil {
		return nil, err
	}
	cpuLimit, err := adapter.LoadCiLimitCpu(cpuLimitParsedValue)
	if err != nil {
		return nil, err
	}
	cpuReqParsedValue, err := impl.unitFactoryMap[units.CPU].ParseValAndUnit(infraConfig.CiReqCpu)
	if err != nil {
		return nil, err
	}
	cpuReq, err := adapter.LoadCiReqCpu(cpuReqParsedValue)
	if err != nil {
		return nil, err
	}
	memLimitParsedValue, err := impl.unitFactoryMap[units.MEMORY].ParseValAndUnit(infraConfig.CiLimitMem)
	if err != nil {
		return nil, err
	}
	memLimit, err := adapter.LoadCiLimitMem(memLimitParsedValue)
	if err != nil {
		return nil, err
	}
	memReqParsedValue, err := impl.unitFactoryMap[units.MEMORY].ParseValAndUnit(infraConfig.CiReqMem)
	if err != nil {
		return nil, err
	}
	memReq, err := adapter.LoadCiReqMem(memReqParsedValue)
	if err != nil {
		return nil, err
	}
	timeoutParsedValue, err := impl.unitFactoryMap[units.TIME].ParseValAndUnit(strconv.FormatInt(infraConfig.CiDefaultTimeout, 10))
	if err != nil {
		return nil, err
	}
	timeout, err := adapter.LoadDefaultTimeout(timeoutParsedValue)
	if err != nil {
		return nil, err
	}
	defaultConfigurations := []*repository.InfraProfileConfigurationEntity{cpuLimit, memLimit, cpuReq, memReq, timeout}
	return defaultConfigurations, nil
}
