package history

import (
	"github.com/devtron-labs/devtron/pkg/globalPolicy/history/bean"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/history/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/globalPolicy/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"go.uber.org/zap"
)

type GlobalPolicyHistoryService interface {
	CreateHistoryEntry(globalPolicy *repository2.GlobalPolicy, action bean.HistoryOfAction) error
}

type GlobalPolicyHistoryServiceImpl struct {
	logger                        *zap.SugaredLogger
	globalPolicyHistoryRepository repository.GlobalPolicyHistoryRepository
}

func NewGlobalPolicyHistoryServiceImpl(logger *zap.SugaredLogger,
	globalPolicyHistoryRepository repository.GlobalPolicyHistoryRepository) *GlobalPolicyHistoryServiceImpl {
	return &GlobalPolicyHistoryServiceImpl{
		logger:                        logger,
		globalPolicyHistoryRepository: globalPolicyHistoryRepository,
	}
}

func (impl *GlobalPolicyHistoryServiceImpl) CreateHistoryEntry(globalPolicy *repository2.GlobalPolicy, action bean.HistoryOfAction) error {
	history := &repository.GlobalPolicyHistory{
		GlobalPolicyId:  globalPolicy.Id,
		Enabled:         globalPolicy.Enabled,
		Description:     globalPolicy.Description,
		PolicyOf:        globalPolicy.PolicyOf,
		PolicyVersion:   globalPolicy.Version,
		PolicyData:      globalPolicy.PolicyJson,
		HistoryOfAction: action,
		AuditLog: sql.AuditLog{
			CreatedOn: globalPolicy.CreatedOn,
			CreatedBy: globalPolicy.CreatedBy,
			UpdatedOn: globalPolicy.UpdatedOn,
			UpdatedBy: globalPolicy.UpdatedBy,
		},
	}

	err := impl.globalPolicyHistoryRepository.Save(history)
	if err != nil {
		impl.logger.Errorw("error in saving globalPolicy history", "err", err, "policy", globalPolicy)
		return err
	}
	return nil
}
