package repository

import (
	"github.com/devtron-labs/devtron/pkg/globalPolicy/history/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type GlobalPolicyHistoryRepository interface {
	Save(model *GlobalPolicyHistory) error
}

type GlobalPolicyHistoryRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewGlobalPolicyHistoryRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *GlobalPolicyHistoryRepositoryImpl {
	return &GlobalPolicyHistoryRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type GlobalPolicyHistory struct {
	tableName       struct{}             `sql:"global_policy_history" pg:",discard_unknown_columns"`
	Id              int                  `sql:"id,pk"`
	GlobalPolicyId  int                  `sql:"global_policy_id"`
	HistoryOfAction bean.HistoryOfAction `sql:"history_of_action"`
	Enabled         bool                 `sql:"enabled,notnull"`
	Description     string               `sql:"description"`
	PolicyOf        string               `sql:"policy_of"`
	PolicyVersion   string               `sql:"policy_version"`
	PolicyData      string               `sql:"policy_data"`
	sql.AuditLog
}

func (repo *GlobalPolicyHistoryRepositoryImpl) Save(model *GlobalPolicyHistory) error {
	err := repo.dbConnection.Insert(model)
	if err != nil {
		repo.logger.Errorw("error in saving history entry for global policy", "err", err, "globalPolicyId", model.GlobalPolicyId)
		return err
	}
	return nil
}
