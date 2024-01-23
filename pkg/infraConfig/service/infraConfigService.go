package service

import (
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/infraConfig"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
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
	GetConfigurationUnits() map[infraConfig.ConfigKeyStr]map[string]units.Unit
	GetDefaultProfile() (*infraConfig.ProfileBean, error)
	UpdateProfile(userId int32, profileName string, profileBean *infraConfig.ProfileBean) error
}

type InfraConfigServiceImpl struct {
	logger           *zap.SugaredLogger
	infraProfileRepo repository.InfraConfigRepository
	units            *units.Units
	infraConfig      *infraConfig.InfraConfig
	validator        *validator.Validate
}

func NewInfraConfigServiceImpl(logger *zap.SugaredLogger,
	infraProfileRepo repository.InfraConfigRepository,
	units *units.Units,
	validator *validator.Validate) (*InfraConfigServiceImpl, error) {
	infraConfiguration := &infraConfig.InfraConfig{}
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
func (impl *InfraConfigServiceImpl) GetDefaultProfile() (*infraConfig.ProfileBean, error) {
	infraProfile, err := impl.infraProfileRepo.GetProfileByName(repository.DEFAULT_PROFILE_NAME)
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

	configurationBeans := util.Transform(infraConfigurations, func(config *infraConfig.InfraProfileConfiguration) infraConfig.ConfigurationBean {
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

func (impl *InfraConfigServiceImpl) UpdateProfile(userId int32, profileName string, profileBean *infraConfig.ProfileBean) error {
	if profileName == "" {
		return errors.New(InvalidProfileName)
	}

	// validation
	defaultProfile, err := impl.GetDefaultProfile()
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "profileCreateRequest", profileBean, "error", err)
		return err
	}
	if err := impl.Validate(profileBean, defaultProfile); err != nil {
		impl.logger.Errorw("error occurred in validation the profile create request", "profileCreateRequest", profileBean, "error", err)
		return err
	}
	// validations end

	infraProfile := profileBean.ConvertToInfraProfile()
	// user couldn't delete the profile, always set this to active
	infraProfile.Active = true
	infraConfigurations := util.Transform(profileBean.Configurations, func(config infraConfig.ConfigurationBean) *infraConfig.InfraProfileConfiguration {
		config.ProfileId = infraProfile.Id
		// user couldn't delete the configuration for default profile, always set this to active
		if infraProfile.Name == repository.DEFAULT_PROFILE_NAME {
			config.Active = true
		}
		return config.ConvertToInfraProfileConfiguration()
	})

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to update profile", "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)
	infraProfile.UpdatedOn = time.Now()
	infraProfile.UpdatedBy = userId
	err = impl.infraProfileRepo.UpdateProfile(tx, profileName, infraProfile)
	if err != nil {
		impl.logger.Errorw("error in updating profile", "error", err)
		return err
	}

	err = impl.infraProfileRepo.UpdateConfigurations(tx, infraConfigurations)
	if err != nil {
		impl.logger.Errorw("error in creating configurations", "error", err)
		return err
	}
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to update profile", "error", err)
	}
	return err
}

func (impl *InfraConfigServiceImpl) loadDefaultProfile() error {

	profile, err := impl.infraProfileRepo.GetProfileByName(repository.DEFAULT_PROFILE_NAME)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		return err
	}
	if profile.Id != 0 {
		// return here because entry already exists and we dont need to create it again
		return nil
	}

	infraConfiguration := impl.infraConfig
	cpuLimit, err := infraConfiguration.GetCiLimitCpu()
	if err != nil {
		return err
	}
	memLimit, err := infraConfiguration.GetCiLimitMem()
	if err != nil {
		return err
	}
	cpuReq, err := infraConfiguration.GetCiReqCpu()
	if err != nil {
		return err
	}
	memReq, err := infraConfiguration.GetCiReqMem()
	if err != nil {
		return err
	}
	timeout, err := infraConfiguration.GetDefaultTimeout()
	if err != nil {
		return err
	}

	defaultConfigurations := []*infraConfig.InfraProfileConfiguration{cpuLimit, memLimit, cpuReq, memReq, timeout}
	defaultProfile := &infraConfig.InfraProfile{
		Name:        repository.DEFAULT_PROFILE_NAME,
		Description: "",
		Active:      true,
		AuditLog: sql.AuditLog{
			CreatedBy: 1, // system user
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
			UpdatedBy: 1, // system user
		},
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

	util.Transform(defaultConfigurations, func(config *infraConfig.InfraProfileConfiguration) *infraConfig.InfraProfileConfiguration {
		config.ProfileId = defaultProfile.Id
		config.Active = true
		config.AuditLog = sql.AuditLog{
			CreatedBy: 1, // system user
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
			UpdatedBy: 1, // system user
		}
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

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() map[infraConfig.ConfigKeyStr]map[string]units.Unit {
	configurationUnits := make(map[infraConfig.ConfigKeyStr]map[string]units.Unit)
	configurationUnits[infraConfig.CPU_REQUEST] = impl.units.GetCpuUnits()
	configurationUnits[infraConfig.CPU_LIMIT] = impl.units.GetCpuUnits()

	configurationUnits[infraConfig.MEMORY_REQUEST] = impl.units.GetMemoryUnits()
	configurationUnits[infraConfig.MEMORY_LIMIT] = impl.units.GetMemoryUnits()

	configurationUnits[infraConfig.TIME_OUT] = impl.units.GetTimeUnits()

	return configurationUnits
}

func (impl *InfraConfigServiceImpl) Validate(profileBean *infraConfig.ProfileBean, defaultProfile *infraConfig.ProfileBean) error {
	err := impl.validator.Struct(profileBean)
	if err != nil {
		err = errors.Wrap(err, PayloadValidationError)
		return err
	}

	// validate configurations only contain default configurations types.(cpu_limit,cpu_request,mem_limit,mem_request,timeout)
	for _, propertyConfig := range profileBean.Configurations {
		if !util.Contains(defaultProfile.Configurations, func(defaultConfig infraConfig.ConfigurationBean) bool {
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

func (impl *InfraConfigServiceImpl) validateCpuMem(profileBean *infraConfig.ProfileBean, defaultProfile *infraConfig.ProfileBean) error {

	configurationUnits := impl.units
	// currently validating cpu and memory limits and reqs only
	var (
		cpuLimit *infraConfig.ConfigurationBean
		cpuReq   *infraConfig.ConfigurationBean
		memLimit *infraConfig.ConfigurationBean
		memReq   *infraConfig.ConfigurationBean
	)

	for _, propertyConfig := range profileBean.Configurations {
		// get cpu limit and req
		switch propertyConfig.Key {
		case infraConfig.CPU_LIMIT:
			cpuLimit = &propertyConfig
		case infraConfig.CPU_REQUEST:
			cpuReq = &propertyConfig
		case infraConfig.MEMORY_LIMIT:
			memLimit = &propertyConfig
		case infraConfig.MEMORY_REQUEST:
			memReq = &propertyConfig
		}
	}

	for _, defaultPropertyConfig := range defaultProfile.Configurations {
		// get cpu limit and req
		switch defaultPropertyConfig.Key {
		case infraConfig.CPU_LIMIT:
			if cpuLimit == nil {
				cpuLimit = &defaultPropertyConfig
			}
		case infraConfig.CPU_REQUEST:
			if cpuReq == nil {
				cpuReq = &defaultPropertyConfig
			}
		case infraConfig.MEMORY_LIMIT:
			if memLimit == nil {
				memLimit = &defaultPropertyConfig
			}
		case infraConfig.MEMORY_REQUEST:
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
