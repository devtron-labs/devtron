package operationAudit

import (
	"github.com/devtron-labs/devtron/api/bean"
	adapter2 "github.com/devtron-labs/devtron/pkg/operationAudit/adapter"
	bean2 "github.com/devtron-labs/devtron/pkg/operationAudit/bean"
	"github.com/devtron-labs/devtron/pkg/operationAudit/repository"
	"go.uber.org/zap"
)

type OperationAuditService interface {
	SaveAuditForUser(entityId int32, operationType bean2.OperationType,
		permissionsAuditDto *bean.PermissionsAuditDto, userIdForAuditLog int32) error
	SaveAuditForRoleGroup(entityId int32, operationType bean2.OperationType,
		permissionsAuditDto *bean.PermissionsAuditDto, userIdForAuditLog int32) error
}

type OperationAuditServiceImpl struct {
	logger                     *zap.SugaredLogger
	PermissionsAuditRepository repository.OperationAuditRepository
}

func NewOperationAuditServiceImpl(logger *zap.SugaredLogger, PermissionsAuditRepository repository.OperationAuditRepository) *OperationAuditServiceImpl {
	return &OperationAuditServiceImpl{
		logger:                     logger,
		PermissionsAuditRepository: PermissionsAuditRepository,
	}
}

func (impl *OperationAuditServiceImpl) saveAudit(entityId int32, entityType bean2.EntityType, operationType bean2.OperationType,
	permissionsAuditDto *bean.PermissionsAuditDto, userIdForAuditLog int32, schemaFor bean2.SchemaFor) error {
	model, err := adapter2.BuildPermissionAuditModel(entityId, entityType, operationType, permissionsAuditDto, userIdForAuditLog, schemaFor)
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

func (impl *OperationAuditServiceImpl) SaveAuditForUser(entityId int32, operationType bean2.OperationType,
	permissionsAuditDto *bean.PermissionsAuditDto, userIdForAuditLog int32) error {
	// by default v1 schema for now, when new schema are added this can be changed
	schemaFor := bean2.UserSchema
	err := impl.saveAudit(entityId, bean2.UserEntity, operationType, permissionsAuditDto, userIdForAuditLog, schemaFor)
	if err != nil {
		impl.logger.Errorw("error in SaveAuditForUser", "err", err)
		return err
	}
	return nil
}

func (impl *OperationAuditServiceImpl) SaveAuditForRoleGroup(entityId int32, operationType bean2.OperationType,
	permissionsAuditDto *bean.PermissionsAuditDto, userIdForAuditLog int32) error {
	// by default v1 schema for now, when new schema are added this can be changed
	schemaFor := bean2.RoleGroupSchema
	err := impl.saveAudit(entityId, bean2.RoleGroupEntity, operationType, permissionsAuditDto, userIdForAuditLog, schemaFor)
	if err != nil {
		impl.logger.Errorw("error in SaveAuditForRoleGroup", "err", err)
		return err
	}
	return nil
}
