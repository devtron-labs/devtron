package infraConfig

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"time"
)

const InvalidProfileName = "profile name is invalid"
const PayloadValidationError = "payload validation failed"

type InfraConfigService interface {
	GetConfigurationUnits() map[ConfigKeyStr]map[string]units.Unit
	GetDefaultProfile() (*ProfileBean, error)
	UpdateProfile(userId int32, profileName string, profileBean *ProfileBean) error
	GetInfraConfigurationsByScope(scope Scope) (*InfraConfig, error)
}

type InfraConfigServiceImpl struct {
	logger           *zap.SugaredLogger
	infraProfileRepo InfraConfigRepository
	units            *units.Units
	infraConfig      *InfraConfig
	validator        *validator.Validate
}

func NewInfraConfigServiceImpl(logger *zap.SugaredLogger,
	infraProfileRepo InfraConfigRepository,
	units *units.Units,
	validator *validator.Validate) (*InfraConfigServiceImpl, error) {
	infraConfiguration := &InfraConfig{}
	err := env.Parse(infraConfiguration)
	if err != nil {
		return nil, err
	}
	infraProfileService := &InfraConfigServiceImpl{
		logger:           logger,
		infraProfileRepo: infraProfileRepo,
		units:            units,
		infraConfig:      infraConfiguration,
		validator:        validator,
	}
	err = infraProfileService.loadDefaultProfile()
	return infraProfileService, err
}
func (impl *InfraConfigServiceImpl) GetDefaultProfile() (*ProfileBean, error) {
	infraProfile, err := impl.infraProfileRepo.GetProfileByName(DEFAULT_PROFILE_NAME)
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

	configurationBeans := util.Transform(infraConfigurations, func(config *InfraProfileConfiguration) ConfigurationBean {
		configBean := config.ConvertToConfigurationBean()
		configBean.ProfileName = profileBean.Name
		return configBean
	})
	profileBean.Configurations = configurationBeans
	appCount, err := impl.infraProfileRepo.GetIdentifierCountForDefaultProfile()
	if err != nil {
		impl.logger.Errorw("error in fetching app count for default profile", "error", err)
		return nil, err
	}
	profileBean.AppCount = appCount
	return &profileBean, nil
}

func (impl *InfraConfigServiceImpl) UpdateProfile(userId int32, profileName string, profileBean *ProfileBean) error {
	if profileName == "" {
		return errors.New(InvalidProfileName)
	}

	// validation
	defaultProfile, err := impl.GetDefaultProfile()
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "profileName", profileName, "profileCreateRequest", profileBean, "error", err)
		return err
	}
	if err := impl.Validate(profileBean, defaultProfile); err != nil {
		impl.logger.Errorw("error occurred in validation the profile create request", "profileName", profileName, "profileCreateRequest", profileBean, "error", err)
		return err
	}
	// validations end

	infraProfile := profileBean.ConvertToInfraProfile()
	// user couldn't delete the profile, always set this to active
	infraProfile.Active = true
	infraConfigurations := util.Transform(profileBean.Configurations, func(config ConfigurationBean) *InfraProfileConfiguration {
		config.ProfileId = infraProfile.Id
		// user couldn't delete the configuration for default profile, always set this to active
		if infraProfile.Name == DEFAULT_PROFILE_NAME {
			config.Active = true
		}
		return config.ConvertToInfraProfileConfiguration()
	})

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update profile", "profileName", profileName, "profileCreateRequest", profileBean, "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)
	infraProfile.UpdatedOn = time.Now()
	infraProfile.UpdatedBy = userId
	err = impl.infraProfileRepo.UpdateProfile(tx, profileName, infraProfile)
	if err != nil {
		impl.logger.Errorw("error in updating profile", "error", "profileName", profileName, "profileCreateRequest", profileBean, err)
		return err
	}

	err = impl.infraProfileRepo.UpdateConfigurations(tx, infraConfigurations)
	if err != nil {
		impl.logger.Errorw("error in creating configurations", "error", "profileName", profileName, "profileCreateRequest", profileBean, err)
		return err
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update profile", "profileName", profileName, "profileCreateRequest", profileBean, "error", err)
	}
	return err
}

func (impl *InfraConfigServiceImpl) loadDefaultProfile() error {

	profile, err := impl.infraProfileRepo.GetProfileByName(DEFAULT_PROFILE_NAME)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		return err
	}
	if profile.Id != 0 {
		// return here because entry already exists and we dont need to create it again
		return nil
	}

	infraConfiguration := impl.infraConfig
	cpuLimit, err := infraConfiguration.LoadCiLimitCpu()
	if err != nil {
		return err
	}
	memLimit, err := infraConfiguration.LoadCiLimitMem()
	if err != nil {
		return err
	}
	cpuReq, err := infraConfiguration.LoadCiReqCpu()
	if err != nil {
		return err
	}
	memReq, err := infraConfiguration.LoadCiReqMem()
	if err != nil {
		return err
	}
	timeout, err := infraConfiguration.LoadDefaultTimeout()
	if err != nil {
		return err
	}

	defaultConfigurations := []*InfraProfileConfiguration{cpuLimit, memLimit, cpuReq, memReq, timeout}
	defaultProfile := &InfraProfile{
		Name:        DEFAULT_PROFILE_NAME,
		Description: "",
		Active:      true,
		AuditLog:    sql.NewDefaultAuditLog(1),
	}
	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "error", err)
		return err
	}

	defer impl.infraProfileRepo.RollbackTx(tx)
	err = impl.infraProfileRepo.CreateProfile(tx, defaultProfile)
	if err != nil {
		impl.logger.Errorw("error in saving default profile", "error", err)
		return err
	}

	util.Transform(defaultConfigurations, func(config *InfraProfileConfiguration) *InfraProfileConfiguration {
		config.ProfileId = defaultProfile.Id
		config.Active = true
		config.AuditLog = sql.NewDefaultAuditLog(1)
		return config
	})
	err = impl.infraProfileRepo.CreateConfigurations(tx, defaultConfigurations)
	if err != nil {
		impl.logger.Errorw("error in saving default configurations", "error", err)
		return err
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to save default configurations", "error", err)
	}
	return err
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

func (impl *InfraConfigServiceImpl) Validate(profileBean *ProfileBean, defaultProfile *ProfileBean) error {
	err := impl.validator.Struct(profileBean)
	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}

	// validate configurations only contain default configurations types.(cpu_limit,cpu_request,mem_limit,mem_request,timeout)
	for _, propertyConfig := range profileBean.Configurations {
		if !util.Contains(defaultProfile.Configurations, func(defaultConfig ConfigurationBean) bool {
			return propertyConfig.Key == defaultConfig.Key
		}) {
			if err == nil {
				err = errors.New(fmt.Sprintf("invalid configuration property \"%s\"", propertyConfig.Key))
			}
			err = errors.Wrap(err, fmt.Sprintf("invalid configuration property \"%s\"", propertyConfig.Key))
		}
	}

	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}

	err = impl.validateCpuMem(profileBean, defaultProfile)
	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}
	return nil
}

func (impl *InfraConfigServiceImpl) validateCpuMem(profileBean *ProfileBean, defaultProfile *ProfileBean) error {

	configurationUnits := impl.units
	// currently validating cpu and memory limits and reqs only
	var (
		cpuLimit *ConfigurationBean
		cpuReq   *ConfigurationBean
		memLimit *ConfigurationBean
		memReq   *ConfigurationBean
	)

	for _, propertyConfig := range profileBean.Configurations {
		// get cpu limit and req
		switch propertyConfig.Key {
		case CPU_LIMIT:
			cpuLimit = &propertyConfig
		case CPU_REQUEST:
			cpuReq = &propertyConfig
		case MEMORY_LIMIT:
			memLimit = &propertyConfig
		case MEMORY_REQUEST:
			memReq = &propertyConfig
		}
	}

	for _, defaultPropertyConfig := range defaultProfile.Configurations {
		// get cpu limit and req
		switch defaultPropertyConfig.Key {
		case CPU_LIMIT:
			if cpuLimit == nil {
				cpuLimit = &defaultPropertyConfig
			}
		case CPU_REQUEST:
			if cpuReq == nil {
				cpuReq = &defaultPropertyConfig
			}
		case MEMORY_LIMIT:
			if memLimit == nil {
				memLimit = &defaultPropertyConfig
			}
		case MEMORY_REQUEST:
			if memReq == nil {
				memReq = &defaultPropertyConfig
			}
		}
	}

	// validate cpu
	cpuLimitUnitSuffix := units.GetCPUUnit(units.CPUUnitStr(cpuLimit.Unit))
	cpuReqUnitSuffix := units.GetCPUUnit(units.CPUUnitStr(cpuReq.Unit))
	var cpuLimitUnit *units.Unit
	var cpuReqUnit *units.Unit
	for cpuUnitSuffix, cpuUnit := range configurationUnits.GetCpuUnits() {
		if string(units.GetCPUUnitStr(cpuLimitUnitSuffix)) == cpuUnitSuffix {
			cpuLimitUnit = &cpuUnit
		}

		if string(units.GetCPUUnitStr(cpuReqUnitSuffix)) == cpuUnitSuffix {
			cpuReqUnit = &cpuUnit
		}

	}

	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRationCPU := cpuLimit.Value / cpuReq.Value
	convFactorCPU := cpuLimitUnit.ConversionFactor / cpuReqUnit.ConversionFactor

	if limitToReqRationCPU*convFactorCPU < 1 {
		return errors.New("cpu limit should not be less than cpu request")
	}

	// validate mem

	memLimitUnitSuffix := units.GetMemoryUnit(units.MemoryUnitStr(memLimit.Unit))
	memReqUnitSuffix := units.GetMemoryUnit(units.MemoryUnitStr(memReq.Unit))
	var memLimitUnit *units.Unit
	var memReqUnit *units.Unit

	for memUnitSuffix, memUnit := range configurationUnits.GetMemoryUnits() {
		if string(units.GetMemoryUnitStr(memLimitUnitSuffix)) == memUnitSuffix {
			memLimitUnit = &memUnit
		}

		if string(units.GetMemoryUnitStr(memReqUnitSuffix)) == memUnitSuffix {
			memReqUnit = &memUnit
		}
	}

	// this condition should be true for valid case => (lim/req)*(lf/rf) >= 1
	limitToReqRationMem := memLimit.Value / memReq.Value
	convFactorMem := memLimitUnit.ConversionFactor / memReqUnit.ConversionFactor

	if limitToReqRationMem*convFactorMem < 1 {
		return errors.New("memory limit should not be less than memory request")
	}

	return nil
}

func (impl *InfraConfigServiceImpl) GetInfraConfigurationsByScope(scope Scope) (*InfraConfig, error) {
	infraConfiguration := &InfraConfig{}
	updateInfraConfig := func(config ConfigurationBean) {
		switch config.Key {
		case CPU_LIMIT:
			infraConfiguration.setCiLimitCpu(impl.GetResolvedValue(config).(string))
		case CPU_REQUEST:
			infraConfiguration.setCiReqCpu(impl.GetResolvedValue(config).(string))
		case MEMORY_LIMIT:
			infraConfiguration.setCiLimitMem(impl.GetResolvedValue(config).(string))
		case MEMORY_REQUEST:
			infraConfiguration.setCiReqMem(impl.GetResolvedValue(config).(string))
		case TIME_OUT:
			infraConfiguration.setCiDefaultTimeout(impl.GetResolvedValue(config).(int64))
		}
	}
	defaultConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileName(DEFAULT_PROFILE_NAME)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "scope", scope, "error", err)
		return nil, err
	}

	for _, defaultConfiguration := range defaultConfigurations {
		defaultConfigurationBean := defaultConfiguration.ConvertToConfigurationBean()
		updateInfraConfig(defaultConfigurationBean)
	}
	return infraConfiguration, nil
}

func (impl *InfraConfigServiceImpl) GetResolvedValue(configurationBean ConfigurationBean) interface{} {
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
