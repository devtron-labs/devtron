package user

import (
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"go.uber.org/zap"
)

type PermissionsAuditService interface {
}

type PermissionsAuditServiceImpl struct {
	logger                     *zap.SugaredLogger
	PermissionsAuditRepository repository.PermissionsAuditRepository
}

func NewPermissionsAuditServiceImpl(logger *zap.SugaredLogger, PermissionsAuditRepository repository.PermissionsAuditRepository) *PermissionsAuditServiceImpl {
	return &PermissionsAuditServiceImpl{
		logger:                     logger,
		PermissionsAuditRepository: PermissionsAuditRepository,
	}
}
func (impl *PermissionsAuditServiceImpl) SaveAudit(audit *repository.PermissionsAudit) error {
	err := impl.PermissionsAuditRepository.SaveAudit(audit)
	if err != nil {
		impl.logger.Errorw("error in saving audit", "audit", audit, "err", err)
		return err
	}
	return nil
}
