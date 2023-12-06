package lockConfiguration

import (
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type LockConfigurationService interface {
	GetLockConfiguration() (*bean.LockConfigResponse, error)
	SaveLockConfiguration(*bean.LockConfigRequest, int32) error
}

type LockConfigurationServiceImpl struct {
	logger                      *zap.SugaredLogger
	lockConfigurationRepository LockConfigurationRepository
}

func NewLockConfigurationServiceImpl(logger *zap.SugaredLogger,
	lockConfigurationRepository LockConfigurationRepository) *LockConfigurationServiceImpl {
	return &LockConfigurationServiceImpl{
		logger:                      logger,
		lockConfigurationRepository: lockConfigurationRepository,
	}
}

func (impl LockConfigurationServiceImpl) GetLockConfiguration() (*bean.LockConfigResponse, error) {
	impl.logger.Infow("Getting active lock configuration")

	lockConfigDto, err := impl.lockConfigurationRepository.GetActiveLockConfig()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	}
	if lockConfigDto == nil {
		return &bean.LockConfigResponse{}, nil
	}
	lockConfig := lockConfigDto.ConvertDBDtoToResponse()
	return lockConfig, nil
}

func (impl LockConfigurationServiceImpl) SaveLockConfiguration(lockConfig *bean.LockConfigRequest, createdBy int32) error {
	lockConfigDto, err := impl.lockConfigurationRepository.GetActiveLockConfig()
	if err != nil && err != pg.ErrNoRows {
		return err
	}
	dbConnection := impl.lockConfigurationRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()

	if lockConfigDto != nil {
		lockConfigDto.Active = false
		lockConfigDto.UpdatedOn = time.Now()
		lockConfigDto.UpdatedBy = createdBy
		err := impl.lockConfigurationRepository.Update(lockConfigDto, tx)
		if err != nil {
			return err
		}
	}

	newLockConfigDto := lockConfig.ConvertRequestToDBDto()
	newLockConfigDto.AuditLog = sql.NewDefaultAuditLog(createdBy)

	return impl.lockConfigurationRepository.Create(newLockConfigDto, tx)
}
