package service

import (
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/pkg/infraProfiles"
	"github.com/devtron-labs/devtron/pkg/infraProfiles/repository"
	"github.com/devtron-labs/devtron/pkg/infraProfiles/units"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
	"time"
)

type InfraProfileService interface {
	GetDefaultProfile() (*infraProfiles.Profile, error)
}

type InfraProfileServiceImpl struct {
	logger           *zap.SugaredLogger
	infraProfileRepo repository.InfraProfileRepository
	units            *units.Units
	infraConfig      *infraProfiles.InfraConfig
}

func NewInfraProfileServiceImpl(logger *zap.SugaredLogger, infraProfileRepo repository.InfraProfileRepository, units *units.Units, config *types.CiCdConfig) (*InfraProfileServiceImpl, error) {
	infraConfig := &infraProfiles.InfraConfig{}
	err := env.Parse(infraConfig)
	if err != nil {
		return nil, err
	}
	infraProfileService := &InfraProfileServiceImpl{
		logger:           logger,
		infraProfileRepo: infraProfileRepo,
		units:            units,
		infraConfig:      infraConfig,
	}
	err = infraProfileService.loadDefaultProfile()
	return infraProfileService, err
}

func (impl InfraProfileServiceImpl) CreateDefaultProfile(profile *infraProfiles.Profile) error {
	return nil
}

func (impl InfraProfileServiceImpl) GetDefaultProfile() (*infraProfiles.Profile, error) {
	return nil, nil
}

func (impl InfraProfileServiceImpl) loadDefaultProfile() error {
	infraConfig := impl.infraConfig
	cpuLimit, err := infraConfig.GetCiLimitCpu()
	if err != nil {
		return err
	}
	memLimit, err := infraConfig.GetCiLimitMem()
	if err != nil {
		return err
	}
	cpuReq, err := infraConfig.GetCiReqCpu()
	if err != nil {
		return err
	}
	memReq, err := infraConfig.GetCiReqMem()
	if err != nil {
		return err
	}
	timeout, err := infraConfig.GetDefaultTimeout()
	if err != nil {
		return err
	}

	defaultConfigurations := []*infraProfiles.InfraProfileConfiguration{cpuLimit, memLimit, cpuReq, memReq, timeout}
	defaultProfile := &infraProfiles.InfraProfile{
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

	Transform(defaultConfigurations, func(config *infraProfiles.InfraProfileConfiguration) *infraProfiles.InfraProfileConfiguration {
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

// Transform will iterate through elements of input slice and apply transform function on each object
// and returns the transformed slice
func Transform[T any, K any](input []T, transform func(inp T) K) []K {

	res := make([]K, len(input))
	for i, _ := range input {
		res[i] = transform(input[i])
	}
	return res

}
