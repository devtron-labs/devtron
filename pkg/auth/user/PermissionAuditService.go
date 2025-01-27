package user

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user/adapter"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"go.uber.org/zap"
)

type PermissionsAuditService interface {
	SaveAudit(entityId int32, entityType repository.EntityType,
		operationType repository.OperationType, permissionsAuditDto *bean.PermissionsAuditDto, userIdForAuditLog int32) error
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

func (impl *PermissionsAuditServiceImpl) SaveAudit(entityId int32, entityType repository.EntityType,
	operationType repository.OperationType, permissionsAuditDto *bean.PermissionsAuditDto, userIdForAuditLog int32) error {
	model, err := adapter.BuildPermissionAuditModel(entityId, entityType, operationType, permissionsAuditDto, userIdForAuditLog)
	if err != nil {
		impl.logger.Errorw("error in BuildPermissionAuditModel", "entityId", entityId, "operationType", operationType, "permissionsAuditDto", permissionsAuditDto, "err", err)
		return err
	}
	err = impl.PermissionsAuditRepository.SaveAudit(model)
	if err != nil {
		impl.logger.Errorw("error in saving audit", "entityId", entityId, "operationType", operationType, "permissionsAuditDto", permissionsAuditDto, "err", err)
		return err
	}
	return nil
}
