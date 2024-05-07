package audit

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	repositoryAdapter "github.com/devtron-labs/devtron/pkg/devtronResource/repository/adapter"
	"go.uber.org/zap"
)

type ObjectAuditService interface {
	SaveAudit(devtronResourceObject *repository.DevtronResourceObject, auditAction repository.AuditOperationType, auditPath []string)
}
type ObjectAuditServiceImpl struct {
	logger                               *zap.SugaredLogger
	devtronResourceObjectAuditRepository repository.DevtronResourceObjectAuditRepository
}

func NewObjectAuditServiceImpl(logger *zap.SugaredLogger,
	devtronResourceObjectAuditRepository repository.DevtronResourceObjectAuditRepository) *ObjectAuditServiceImpl {
	return &ObjectAuditServiceImpl{
		logger:                               logger,
		devtronResourceObjectAuditRepository: devtronResourceObjectAuditRepository,
	}
}

func (impl *ObjectAuditServiceImpl) SaveAudit(devtronResourceObject *repository.DevtronResourceObject, auditAction repository.AuditOperationType, auditPath []string) {
	auditModel := repositoryAdapter.GetResourceObjectAudit(devtronResourceObject, auditAction, auditPath)
	err := impl.devtronResourceObjectAuditRepository.Save(auditModel)
	if err != nil { //only logging not propagating to user
		impl.logger.Warnw("error in saving devtronResourceObject audit", "err", err, "auditModel", auditModel)
	}
}

func (impl *ObjectAuditServiceImpl) SaveAuditForTaskRun(devtronResourceObject *repository.DevtronResourceObject, auditAction repository.AuditOperationType, auditPath []string) {
	auditModel := repositoryAdapter.GetResourceObjectAudit(devtronResourceObject, auditAction, auditPath)
	err := impl.devtronResourceObjectAuditRepository.Save(auditModel)
	if err != nil { //only logging not propagating to user
		impl.logger.Warnw("error in saving devtronResourceObject audit", "err", err, "auditModel", auditModel)
	}
}
