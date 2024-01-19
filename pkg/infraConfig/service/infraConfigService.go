package service

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/infraConfig"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository"
	"github.com/devtron-labs/devtron/pkg/infraConfig/units"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
	"time"
)

type InfraConfigService interface {
	GetDefaultProfile() (*infraConfig.ProfileBean, error)
	UpdateProfile(userId int32, profileName string, profileBean *infraConfig.ProfileBean) error
	GetConfigurationUnits() map[infraConfig.ConfigKeyStr][]units.Unit
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
	infraConfigurations, err := impl.infraProfileRepo.GetConfigurationsByProfileId(infraProfile.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching default configurations", "error", err)
		return nil, err
	}

	configurationBeans := Transform(infraConfigurations, func(config *infraConfig.InfraProfileConfiguration) infraConfig.ConfigurationBean {
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

func (impl *InfraConfigServiceImpl) UpdateProfile(userId int32, profileName string, profileBean *infraConfig.ProfileBean) error {
	infraProfile := profileBean.ConvertToInfraProfile()
	// user couldn't delete the profile, always set this to active
	infraProfile.Active = true
	infraConfigurations := Transform(profileBean.Configurations, func(config infraConfig.ConfigurationBean) *infraConfig.InfraProfileConfiguration {
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

	Transform(defaultConfigurations, func(config *infraConfig.InfraProfileConfiguration) *infraConfig.InfraProfileConfiguration {
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

// todo: delete thsi function
// Transform will iterate through elements of input slice and apply transform function on each object
// and returns the transformed slice
func Transform[T any, K any](input []T, transform func(inp T) K) []K {

	res := make([]K, len(input))
	for i, _ := range input {
		res[i] = transform(input[i])
	}
	return res

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
