package infraConfig

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type InfraConfigService interface {
	GetConfigurationUnits() map[ConfigKeyStr]map[string]units.Unit
	GetProfileByName(name string) (*ProfileBean, error)
	UpdateProfile(userId int32, profileName string, profileBean *ProfileBean) error
}

type InfraConfigServiceImpl struct {
	logger           *zap.SugaredLogger
	infraProfileRepo InfraConfigRepository
	appService       app.AppService
	units            *units.Units
	infraConfig      *InfraConfig
}

func NewInfraConfigServiceImpl(logger *zap.SugaredLogger,
	infraProfileRepo InfraConfigRepository,
	appService app.AppService,
	units *units.Units) (*InfraConfigServiceImpl, error) {
	infraConfiguration := &InfraConfig{}
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

func (impl *InfraConfigServiceImpl) GetProfileByName(name string) (*ProfileBean, error) {
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

	configurationBeans := util.Transform(infraConfigurations, func(config *InfraProfileConfigurationEntity) ConfigurationBean {
		configBean := config.ConvertToConfigurationBean()
		configBean.ProfileName = profileBean.Name
		return configBean
	})

	profileBean.Configurations = configurationBeans
	appCount, err := impl.appService.GetActiveCiCdAppsCount()
	if err != nil {
		impl.logger.Errorw("error in fetching app count for default profile", "error", err)
		return nil, err
	}
	profileBean.AppCount = appCount
	return &profileBean, nil
}

func (impl *InfraConfigServiceImpl) UpdateProfile(userId int32, profileName string, profileToUpdate *ProfileBean) error {
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
	infraConfigurations := util.Transform(profileToUpdate.Configurations, func(config ConfigurationBean) *InfraProfileConfigurationEntity {
		config.ProfileId = defaultProfile.Id
		// user couldn't delete the configuration for default profile, always set this to active
		if infraProfileEntity.Name == DEFAULT_PROFILE_NAME {
			config.Active = true
		}
		configuration := config.ConvertToInfraProfileConfigurationEntity()
		configuration.UpdatedOn = time.Now()
		configuration.UpdatedBy = userId
		return configuration
	})

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update profile", "profileName", profileName, "profileCreateRequest", profileToUpdate, "error", err)
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

	profile, err := impl.infraProfileRepo.GetProfileByName(DEFAULT_PROFILE_NAME)
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
		defaultProfile := &InfraProfileEntity{
			Name:        DEFAULT_PROFILE_NAME,
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
	defaultConfigurationsFromDB, err := impl.infraProfileRepo.GetConfigurationsByProfileName(DEFAULT_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return err
	}
	defaultConfigurationsFromDBMap := make(map[ConfigKey]bool)
	for _, defaultConfigurationFromDB := range defaultConfigurationsFromDB {
		defaultConfigurationsFromDBMap[defaultConfigurationFromDB.Key] = true
	}

	creatableConfigurations := make([]*InfraProfileConfigurationEntity, 0, len(defaultConfigurationsFromEnv))
	util.Transform(defaultConfigurationsFromEnv, func(config *InfraProfileConfigurationEntity) *InfraProfileConfigurationEntity {
		if !defaultConfigurationsFromDBMap[config.Key] {
			config.ProfileId = profile.Id
			config.Active = true
			config.AuditLog = sql.NewDefaultAuditLog(1)
			creatableConfigurations = append(creatableConfigurations, config)
		}
		return config
	})

	if len(creatableConfigurations) > 0 {
		err = impl.infraProfileRepo.CreateConfigurations(tx, creatableConfigurations)
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

func (impl *InfraConfigServiceImpl) getInfraConfigurationsByScope(scope Scope) (*InfraConfig, error) {
	infraConfiguration := &InfraConfig{}
	overrideInfraConfigFunc := func(config ConfigurationBean) {
		switch config.Key {
		case CPU_LIMIT:
			infraConfiguration.setCiLimitCpu(impl.getResolvedValue(config).(string))
		case CPU_REQUEST:
			infraConfiguration.setCiReqCpu(impl.getResolvedValue(config).(string))
		case MEMORY_LIMIT:
			infraConfiguration.setCiLimitMem(impl.getResolvedValue(config).(string))
		case MEMORY_REQUEST:
			infraConfiguration.setCiReqMem(impl.getResolvedValue(config).(string))
		case TIME_OUT:
			infraConfiguration.setCiDefaultTimeout(impl.getResolvedValue(config).(int64))
		}
	}
	defaultConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileName(DEFAULT_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "scope", scope, "error", err)
		return nil, err
	}

	for _, defaultConfiguration := range defaultConfigurations {
		defaultConfigurationBean := defaultConfiguration.ConvertToConfigurationBean()
		overrideInfraConfigFunc(defaultConfigurationBean)
	}
	return infraConfiguration, nil
}

func (impl *InfraConfigServiceImpl) getResolvedValue(configurationBean ConfigurationBean) interface{} {
	// for timeout we need to get the value in seconds
	if configurationBean.Key == GetConfigKeyStr(TimeOut) {
		// if user ever gives the timeout in float, after conversion to int64 it will be rounded off
		return int64(configurationBean.Value * impl.units.GetTimeUnits()[configurationBean.Unit].ConversionFactor)
	}
	if configurationBean.Unit == string(units.CORE) || configurationBean.Unit == string(units.BYTE) {
		return fmt.Sprintf("%v", configurationBean.Value)
	}
	return fmt.Sprintf("%v%v", configurationBean.Value, configurationBean.Unit)
}

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() map[ConfigKeyStr]map[string]units.Unit {
	configurationUnits := make(map[ConfigKeyStr]map[string]units.Unit)
	configurationUnits[CPU_REQUEST] = impl.units.GetCpuUnits()
	configurationUnits[CPU_LIMIT] = impl.units.GetCpuUnits()

	configurationUnits[MEMORY_REQUEST] = impl.units.GetMemoryUnits()
	configurationUnits[MEMORY_LIMIT] = impl.units.GetMemoryUnits()

	configurationUnits[TIME_OUT] = impl.units.GetTimeUnits()

	return configurationUnits
}

func (impl *InfraConfigServiceImpl) Validate(profileToUpdate *ProfileBean, defaultProfile *ProfileBean) error {

	var err error = nil
	// validate configurations only contain default configurations types.(cpu_limit,cpu_request,mem_limit,mem_request,timeout)
	for _, propertyConfig := range profileToUpdate.Configurations {
		if !util.Contains(defaultProfile.Configurations, func(defaultConfig ConfigurationBean) bool {
			return propertyConfig.Key == defaultConfig.Key
		}) {
			errorMsg := fmt.Sprintf("invalid configuration property \"%s\"", propertyConfig.Key)
			if err == nil {
				err = errors.New(errorMsg)
			}
			err = errors.Wrap(err, errorMsg)
		}
	}

	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}

	err = impl.validateCpuMem(profileToUpdate)
	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validateCpuMem(profileBean *ProfileBean) error {

	// currently validating cpu and memory limits and reqs only
	var (
		cpuLimit *ConfigurationBean
		cpuReq   *ConfigurationBean
		memLimit *ConfigurationBean
		memReq   *ConfigurationBean
	)

	for i, _ := range profileBean.Configurations {
		// get cpu limit and req
		switch profileBean.Configurations[i].Key {
		case CPU_LIMIT:
			cpuLimit = &profileBean.Configurations[i]
		case CPU_REQUEST:
			cpuReq = &profileBean.Configurations[i]
		case MEMORY_LIMIT:
			memLimit = &profileBean.Configurations[i]
		case MEMORY_REQUEST:
			memReq = &profileBean.Configurations[i]
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

func (impl *InfraConfigServiceImpl) validateCPU(cpuLimit, cpuReq *ConfigurationBean) error {
	configurationUnits := impl.units
	cpuLimitUnitSuffix := units.CPUUnitStr(cpuLimit.Unit).GetCPUUnit()
	cpuReqUnitSuffix := units.CPUUnitStr(cpuReq.Unit).GetCPUUnit()
	var cpuLimitUnit units.Unit
	var cpuReqUnit units.Unit
	for cpuUnitSuffix, cpuUnit := range configurationUnits.GetCpuUnits() {
		if string(cpuLimitUnitSuffix.GetCPUUnitStr()) == cpuUnitSuffix {
			cpuLimitUnit = cpuUnit
		}

		if string(cpuReqUnitSuffix.GetCPUUnitStr()) == cpuUnitSuffix {
			cpuReqUnit = cpuUnit
		}

	}
	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRationCPU := cpuLimit.Value / cpuReq.Value
	convFactorCPU := cpuLimitUnit.ConversionFactor / cpuReqUnit.ConversionFactor

	if limitToReqRationCPU*convFactorCPU < 1 {
		return errors.New(CPULimReqErrorCompErr)
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validateMEM(memLimit, memReq *ConfigurationBean) error {
	configurationUnits := impl.units
	memLimitUnitSuffix := units.MemoryUnitStr(memLimit.Unit).GetMemoryUnit()
	memReqUnitSuffix := units.MemoryUnitStr(memReq.Unit).GetMemoryUnit()
	var memLimitUnit units.Unit
	var memReqUnit units.Unit

	for memUnitSuffix, memUnit := range configurationUnits.GetMemoryUnits() {
		if string(memLimitUnitSuffix.GetMemoryUnitStr()) == memUnitSuffix {
			memLimitUnit = memUnit
		}

		if string(memReqUnitSuffix.GetMemoryUnitStr()) == memUnitSuffix {
			memReqUnit = memUnit
		}
	}

	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRationMem := memLimit.Value / memReq.Value
	convFactorMem := memLimitUnit.ConversionFactor / memReqUnit.ConversionFactor

	if limitToReqRationMem*convFactorMem < 1 {
		return errors.New(MEMLimReqErrorCompErr)
	}
	return nil
}
