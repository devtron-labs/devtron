package resourceFilter

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ResourceFilterAudit struct {
	tableName    struct{}            `sql:"resource_filter_audit" pg:",discard_unknown_columns"`
	Id           int                 `sql:"id"`
	FilterId     int                 `sql:"filter_id"`
	Conditions   string              `sql:"conditions"` //json string
	TargetObject *FilterTargetObject `sql:"target_object"`
	sql.AuditLog
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

func (repo *FilterAuditRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo *FilterAuditRepositoryImpl) CreateResourceFilterAudit(tx *pg.Tx, filter *ResourceFilterAudit) (*ResourceFilterAudit, error) {
	err := tx.Insert(filter)
	return filter, err
}
func (repo *FilterAuditRepositoryImpl) GetLatestResourceFilterAuditIdsByFilterIds(filterIds []int) ([]int, error) {
	if len(filterIds) == 0 {
		return nil, nil
	}
	type filterAuditResp struct {
		Id int `sql:"Id"`
	}
	res := make([]filterAuditResp, 0)
	filterAuditIds := make([]int, 0)
	query := "SELECT max(id) " +
		"AS id FROM " +
		"resource_filter_audit " +
		"WHERE filter_id in (%s) " +
		"GROUP BY filter_id"
	_, err := repo.dbConnection.Query(&res, query)
	if err != nil {
		return filterAuditIds, err
	}
	for _, fResp := range res {
		filterAuditIds = append(filterAuditIds, fResp.Id)
	}
	return filterAuditIds, nil
}
