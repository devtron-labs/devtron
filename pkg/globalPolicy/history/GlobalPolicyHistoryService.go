/*
 * Copyright (c) 2024. Devtron Inc.
 */

package history

import (
	"github.com/devtron-labs/devtron/pkg/globalPolicy/history/bean"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/history/repository"
	repository2 "github.com/devtron-labs/devtron/pkg/globalPolicy/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type GlobalPolicyHistoryService interface {
	CreateHistoryEntry(tx *pg.Tx, globalPolicy *repository2.GlobalPolicy, action bean.HistoryOfAction) error
	GetByIds(ids []int) ([]*repository.GlobalPolicyHistory, error)
	GetIdsByPolicyIds(policyIds []int) ([]int, error)
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

func (impl *GlobalPolicyHistoryServiceImpl) CreateHistoryEntry(tx *pg.Tx, globalPolicy *repository2.GlobalPolicy, action bean.HistoryOfAction) error {
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

	err := impl.globalPolicyHistoryRepository.Save(tx, history)
	if err != nil {
		impl.logger.Errorw("error in saving globalPolicy history", "err", err, "policy", globalPolicy)
		return err
	}
	return nil
}

func (impl *GlobalPolicyHistoryServiceImpl) GetByIds(ids []int) ([]*repository.GlobalPolicyHistory, error) {
	return impl.globalPolicyHistoryRepository.GetByIds(ids)
}

func (impl *GlobalPolicyHistoryServiceImpl) GetIdsByPolicyIds(policyIds []int) ([]int, error) {
	return impl.globalPolicyHistoryRepository.GetIdsByPolicyIds(policyIds)
}
