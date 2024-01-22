package service

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/infraConfig"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

const CannotDeleteDefaultProfile = "cannot delete default profile"
const InvalidProfileName = "profile name is invalid"

type InfraConfigService interface {

	// GetDefaultProfile fetches the default profile and its configurations.
	GetDefaultProfile() (*infraConfig.ProfileBean, error)

	// UpdateProfile updates the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	UpdateProfile(userId int32, profileName string, profileBean *infraConfig.ProfileBean) error

	// GetConfigurationUnits fetches all the units for the configurations.
	GetConfigurationUnits() map[infraConfig.ConfigKeyStr][]units.Unit

	// GetProfileByName fetches the profile and its configurations matching the given profileName.
	GetProfileByName(profileName string) (*infraConfig.ProfileBean, error)

	// DeleteProfile deletes the profile and its configurations matching the given profileName.
	// If profileName is empty, it will return an error.
	DeleteProfile(profileName string) error

	// GetProfileList fetches all the profile and their configurations matching the given profileNameLike string.
	// If profileNameLike is empty, it will fetch all the active profiles.
	GetProfileList(profileNameLike string) ([]*infraConfig.ProfileBean, error)
}

type InfraConfigServiceImpl struct {
	logger           *zap.SugaredLogger
	infraProfileRepo repository.InfraConfigRepository
	units            *units.Units
	infraConfig      *infraConfig.InfraConfig
}

func NewInfraConfigServiceImpl(logger *zap.SugaredLogger, infraProfileRepo repository.InfraConfigRepository, units *units.Units) (*InfraConfigServiceImpl, error) {
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
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileId([]int{infraProfile.Id})
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return nil, err
	}

	configurationBeans := infraConfig.Transform(infraConfigurations, func(config *infraConfig.InfraProfileConfiguration) infraConfig.ConfigurationBean {
		configBean := config.ConvertToConfigurationBean()
		configBean.ProfileName = profileBean.Name
		return configBean
	})
	profileBean.Configurations = configurationBeans
	appCount, err := impl.infraProfileRepo.GetIdentifierCountForDefaultProfile(infraProfile.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching app count for default profile", "error", err)
		return nil, err
	}
	profileBean.AppCount = appCount
	return &profileBean, nil
}

func (impl *InfraConfigServiceImpl) GetProfileByName(profileName string) (*infraConfig.ProfileBean, error) {
	if profileName == "" {
		return nil, errors.New(InvalidProfileName)
	}

	infraProfile, err := impl.infraProfileRepo.GetProfileByName(profileName)
	if err != nil {
		impl.logger.Errorw("error in fetching default profile", "error", err)
		return nil, err
	}

	profileBean := infraProfile.ConvertToProfileBean()
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileId([]int{infraProfile.Id})
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return nil, err
	}

	configurationBeans := infraConfig.Transform(infraConfigurations, func(config *infraConfig.InfraProfileConfiguration) infraConfig.ConfigurationBean {
		configBean := config.ConvertToConfigurationBean()
		configBean.ProfileName = profileBean.Name
		return configBean
	})
	profileBean.Configurations = configurationBeans
	return &profileBean, nil
}

func (impl *InfraConfigServiceImpl) GetProfileList(profileNameLike string) ([]*infraConfig.ProfileBean, error) {

	profileAppCount, err := impl.infraProfileRepo.GetIdentifierCountForNonDefaultProfiles([]int{}, "APP")
	if err != nil {
		impl.logger.Errorw("error in fetching app count for non default profiles", "error", err)
		return nil, err
	}
	return nil, nil
}

func (impl *InfraConfigServiceImpl) UpdateProfile(userId int32, profileName string, profileBean *infraConfig.ProfileBean) error {
	if profileName == "" {
		return errors.New(InvalidProfileName)
	}

	infraProfile := profileBean.ConvertToInfraProfile()
	// user couldn't delete the profile, always set this to active
	infraProfile.Active = true
	infraConfigurations := infraConfig.Transform(profileBean.Configurations, func(config infraConfig.ConfigurationBean) *infraConfig.InfraProfileConfiguration {
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

func (impl *InfraConfigServiceImpl) DeleteProfile(profileName string) error {
	if profileName == "" {
		return errors.New(InvalidProfileName)
	}

	if profileName == repository.DEFAULT_PROFILE_NAME {
		return errors.New(CannotDeleteDefaultProfile)
	}

	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to delete profile", "profileName", profileName, "error", err)
		return err
	}
	defer impl.infraProfileRepo.RollbackTx(tx)

	// step1: delete profile
	err = impl.infraProfileRepo.DeleteProfile(tx, profileName)
	if err != nil {
		impl.logger.Errorw("error in deleting profile", "profileName", profileName, "error", err)
		return err
	}

	// step2: delete configurations
	err = impl.infraProfileRepo.DeleteConfigurations(tx, profileName)
	if err != nil {
		impl.logger.Errorw("error in deleting configurations", "profileName", profileName, "error", err)
	}

	// step3: delete profile identifier mappings
	// todo: delete from resource_identifier_mapping where resource_id is profileId and resource_type is infraProfile
	err = impl.infraProfileRepo.CommitTx(tx)
	if err != nil {
		impl.logger.Errorw("error in committing transaction to delete profile", "profileName", profileName, "error", err)
	}
	return err
}

func (impl *InfraConfigServiceImpl) loadDefaultProfile() error {
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
			CreatedBy: 1,
			CreatedOn: time.Now(),
		},
	}
	tx, err := impl.infraProfileRepo.StartTx()
	if err != nil {
		impl.logger.Errorw("error in starting transaction to save default configurations", "error", err)
		return err
	}

	defer impl.infraProfileRepo.RollbackTx(tx)
	err = impl.infraProfileRepo.CreateDefaultProfile(tx, defaultProfile)
	if err != nil {
		impl.logger.Errorw("error in saving default profile", "error", err)
		return err
	}

	infraConfig.Transform(defaultConfigurations, func(config *infraConfig.InfraProfileConfiguration) *infraConfig.InfraProfileConfiguration {
		config.ProfileId = defaultProfile.Id
		config.Active = true
		config.AuditLog = sql.AuditLog{
			CreatedBy: 1, // system user
			CreatedOn: time.Now(),
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

func (impl *InfraConfigServiceImpl) GetConfigurationUnits() map[infraConfig.ConfigKeyStr][]units.Unit {
	configurationUnits := make(map[infraConfig.ConfigKeyStr][]units.Unit)
	configurationUnits[infraConfig.CPU_REQUEST] = impl.units.GetCpuUnits()
	configurationUnits[infraConfig.CPU_LIMIT] = impl.units.GetCpuUnits()

	configurationUnits[infraConfig.MEMORY_REQUEST] = impl.units.GetMemoryUnits()
	configurationUnits[infraConfig.MEMORY_LIMIT] = impl.units.GetMemoryUnits()

	configurationUnits[infraConfig.TIME_OUT] = impl.units.GetTimeUnits()

	return configurationUnits
}
