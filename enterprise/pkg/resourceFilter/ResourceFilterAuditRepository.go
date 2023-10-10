package resourceFilter

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type FilterHistoryAudit struct {
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
	CreateResourceFilterEvaluation(tx *pg.Tx, filter *ResourceFilterEvaluation) (*ResourceFilterEvaluation, error)
	GetResourceFilterEvaluation(referenceType ReferenceType, referenceId int, subjectType SubjectType, subjectId int)
}

type FilterAuditRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}
