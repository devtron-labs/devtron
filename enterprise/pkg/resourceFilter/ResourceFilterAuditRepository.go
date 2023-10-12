package resourceFilter

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type ActionType int

const Create ActionType = 0
const Update ActionType = 1
const Delete ActionType = 2

type ResourceFilterAudit struct {
	tableName    struct{}            `sql:"resource_filter_audit" pg:",discard_unknown_columns"`
	Id           int                 `sql:"id"`
	FilterId     int                 `sql:"filter_id"`
	Conditions   string              `sql:"conditions"` //json string
	TargetObject *FilterTargetObject `sql:"target_object"`
	Action       *ActionType         `sql:"action"`
	sql.AuditLog
}

func NewResourceFilterAudit(filterId int,
	conditions string,
	targetObject *FilterTargetObject,
	action *ActionType, userId int32) ResourceFilterAudit {
	return ResourceFilterAudit{
		FilterId:     filterId,
		Conditions:   conditions,
		TargetObject: targetObject,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: userId,
		},
		Action: action,
	}
}

type FilterAuditRepository interface {
	//transaction util funcs
	sql.TransactionWrapper
	GetConnection() *pg.DB
	CreateResourceFilterAudit(tx *pg.Tx, filter *ResourceFilterAudit) (*ResourceFilterAudit, error)
	GetLatestResourceFilterAuditByFilterIds(ids []int) ([]*ResourceFilterAudit, error)
}

type FilterAuditRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewFilterAuditRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *FilterAuditRepositoryImpl {
	return &FilterAuditRepositoryImpl{
		logger:              logger,
		dbConnection:        dbConnection,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}
}

func (repo *FilterAuditRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo *FilterAuditRepositoryImpl) CreateResourceFilterAudit(tx *pg.Tx, filter *ResourceFilterAudit) (*ResourceFilterAudit, error) {
	err := tx.Insert(filter)
	return filter, err
}

// GetLatestResourceFilterAuditByFilterIds will only fetch latest filter_audit id and filter_id for given filter id's
func (repo *FilterAuditRepositoryImpl) GetLatestResourceFilterAuditByFilterIds(filterIds []int) ([]*ResourceFilterAudit, error) {
	if len(filterIds) == 0 {
		return nil, nil
	}
	res := make([]*ResourceFilterAudit, 0)
	query := "SELECT max(id) " +
		"AS id,filter_id FROM " +
		"resource_filter_audit " +
		"WHERE filter_id IN (?) " +
		"GROUP BY filter_id"
	_, err := repo.dbConnection.Query(&res, query, pg.In(filterIds))
	return res, err
}
