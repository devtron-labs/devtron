package operationAudit

import (
	"github.com/devtron-labs/devtron/pkg/operationAudit/repository"
	"github.com/google/wire"
)

var AuditWireSet = wire.NewSet(
	NewOperationAuditServiceImpl,
	wire.Bind(new(OperationAuditService), new(*OperationAuditServiceImpl)),

	repository.NewOperationAuditRepositoryImpl,
	wire.Bind(new(repository.OperationAuditRepository), new(*repository.OperationAuditRepositoryImpl)),
)
