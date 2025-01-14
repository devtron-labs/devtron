package audit

import (
	audit2 "github.com/devtron-labs/devtron/pkg/infraConfig/adapter/audit"
	infraBean "github.com/devtron-labs/devtron/pkg/infraConfig/bean/v1"
	"github.com/devtron-labs/devtron/pkg/infraConfig/repository/audit"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type InfraConfigAuditService interface {
	SaveCiInfraConfigHistorySnapshot(tx *pg.Tx, workflowId int, triggeredBy int32, infraConfigs map[string]*infraBean.InfraConfig) error
	sql.TransactionWrapper
}

type InfraConfigAuditServiceImpl struct {
	logger                     *zap.SugaredLogger
	infraConfigAuditRepository audit.InfraConfigAuditRepository
	*sql.TransactionUtilImpl
}

func NewInfraConfigAuditServiceImpl(logger *zap.SugaredLogger,
	infraConfigAuditRepository audit.InfraConfigAuditRepository,
	transactionUtilImpl *sql.TransactionUtilImpl) *InfraConfigAuditServiceImpl {
	return &InfraConfigAuditServiceImpl{
		logger:                     logger,
		infraConfigAuditRepository: infraConfigAuditRepository,
		TransactionUtilImpl:        transactionUtilImpl,
	}
}

func (impl *InfraConfigAuditServiceImpl) SaveCiInfraConfigHistorySnapshot(tx *pg.Tx,
	workflowId int, triggeredBy int32, infraConfigs map[string]*infraBean.InfraConfig) error {
	infraConfigTriggerAudits := make([]*audit.InfraConfigTriggerHistory, 0)
	for platform, infraConfig := range infraConfigs {
		infraConfigTriggerHistories, err := audit2.GetInfraConfigTriggerAudit(infraConfig)
		if err != nil {
			impl.logger.Errorw("failed to get infra config trigger audit", "error", err, "infraConfig", infraConfig)
			return err
		}
		for _, infraConfigTriggerHistory := range infraConfigTriggerHistories {
			infraConfigTriggerHistory = infraConfigTriggerHistory.
				WithPlatform(platform).WithWorkflowId(workflowId).
				WithWorkflowType(audit.CIWorkflowType).WithAuditLog(triggeredBy)
		}
		infraConfigTriggerAudits = append(infraConfigTriggerAudits, infraConfigTriggerHistories...)
	}
	impl.logger.Debugw("saving infra config history snapshot", "workflowId", workflowId,
		"infraConfigs", infraConfigs, "infraConfigTriggerAudits", infraConfigTriggerAudits)
	err := impl.infraConfigAuditRepository.SaveInfraConfigHistorySnapshot(tx, infraConfigTriggerAudits)
	if err != nil {
		impl.logger.Errorw("failed to save infra config history snapshot", "error", err, "infraConfigTriggerAudits", infraConfigTriggerAudits)
		return err
	}
	return nil
}
