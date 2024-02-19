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

	// GetConfigurationUnits fetches all the units for the configurations.
	GetConfigurationUnits() map[ConfigKeyStr]map[string]units.Unit
	// GetProfileByName fetches the profile and its configurations matching the given profileName.
	GetProfileByName(name string) (*ProfileBean, error)
	// UpdateProfile updates the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	UpdateProfile(userId int32, profileName string, profileBean *ProfileBean) error

	GetInfraConfigurationsByScope(scope Scope) (*InfraConfig, error)
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
		if profileName == DEFAULT_PROFILE_NAME {
			config.Active = true
		}
		configuration := config.ConvertToInfraProfileConfigurationEntity()
		configuration.UpdatedOn = time.Now()
		configuration.UpdatedBy = userId
		return configuration
	})

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
	// todo: check the error logic here
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return err
	}
	defaultConfigurationsFromDBMap := make(map[ConfigKey]bool)
	for _, defaultConfigurationFromDB := range defaultConfigurationsFromDB {
		defaultConfigurationsFromDBMap[defaultConfigurationFromDB.Key] = true
	}

	creatableConfigurations := make([]*InfraProfileConfigurationEntity, 0, len(defaultConfigurationsFromEnv))
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

func (impl *InfraConfigServiceImpl) GetInfraConfigurationsByScope(scope Scope) (*InfraConfig, error) {
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
		timeUnit := units.TimeUnitStr(configurationBean.Unit)
		return int64(configurationBean.Value * impl.units.GetTimeUnits()[timeUnit].ConversionFactor)
	}
	if configurationBean.Unit == string(units.CORE) || configurationBean.Unit == string(units.BYTE) {
		return fmt.Sprintf("%v", configurationBean.Value)
	}
	return fmt.Sprintf("%v%v", configurationBean.Value, configurationBean.Unit)
}

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() map[ConfigKeyStr]map[string]units.Unit {
	configurationUnits := make(map[ConfigKeyStr]map[string]units.Unit)
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

	configurationUnits[CPU_REQUEST] = cpuUnits
	configurationUnits[CPU_LIMIT] = cpuUnits

	configurationUnits[MEMORY_REQUEST] = memUnits
	configurationUnits[MEMORY_LIMIT] = memUnits

	configurationUnits[TIME_OUT] = timeUnits

	return configurationUnits
}

func (impl *InfraConfigServiceImpl) Validate(profileToUpdate *ProfileBean, defaultProfile *ProfileBean) error {
	var err error = nil
	defaultConfigurationsKeyMap := GetDefaultConfigKeysMap()
	// validate configurations only contain default configurations types.(cpu_limit,cpu_request,mem_limit,mem_request,timeout)
	for _, propertyConfig := range profileToUpdate.Configurations {
		if _, ok := defaultConfigurationsKeyMap[propertyConfig.Key]; !ok {
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

	err = impl.validateCpuMem(profileToUpdate, defaultProfile)
	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validateCpuMem(profileBean *ProfileBean, defaultProfile *ProfileBean) error {

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
	cpuLimitUnitSuffix := units.CPUUnitStr(cpuLimit.Unit)
	cpuReqUnitSuffix := units.CPUUnitStr(cpuReq.Unit)
	cpuUnits := impl.units.GetCpuUnits()
	cpuLimitUnit, ok := cpuUnits[cpuLimitUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(InvalidUnit, cpuLimit.Unit, cpuLimit.Key))
	}
	cpuReqUnit, ok := cpuUnits[cpuReqUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(InvalidUnit, cpuReq.Unit, cpuReq.Key))
	}

	if !validLimReq(cpuLimit.Value, cpuLimitUnit.ConversionFactor, cpuReq.Value, cpuReqUnit.ConversionFactor) {
		return errors.New(CPULimReqErrorCompErr)
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validateMEM(memLimit, memReq *ConfigurationBean) error {
	memLimitUnitSuffix := units.MemoryUnitStr(memLimit.Unit)
	memReqUnitSuffix := units.MemoryUnitStr(memReq.Unit)
	memUnits := impl.units.GetMemoryUnits()
	memLimitUnit, ok := memUnits[memLimitUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(InvalidUnit, memLimit.Unit, memLimit.Key))
	}
	memReqUnit, ok := memUnits[memReqUnitSuffix]
	if !ok {
		return errors.New(fmt.Sprintf(InvalidUnit, memReq.Unit, memReq.Key))
	}

	if !validLimReq(memLimit.Value, memLimitUnit.ConversionFactor, memReq.Value, memReqUnit.ConversionFactor) {
		return errors.New(MEMLimReqErrorCompErr)
	}
	return nil
}

func validLimReq(lim, limFactor, req, reqFactor float64) bool {
	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRatio := lim / req
	convFactor := limFactor / reqFactor
	return limitToReqRatio*convFactor >= 1
}
