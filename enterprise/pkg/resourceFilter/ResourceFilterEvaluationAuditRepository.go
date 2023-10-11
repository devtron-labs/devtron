package resourceFilter

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type SubjectType int

const Artifact SubjectType = 0
const CiMaterial SubjectType = 1

type ReferenceType int

const Pipeline ReferenceType = 0
const PipelineStage ReferenceType = 1
const CdWorkflowRunner ReferenceType = 2

type ResourceFilterEvaluationAudit struct {
	tableName            struct{}      `sql:"resource_filter_evaluation_audit" pg:",discard_unknown_columns"`
	Id                   int           `sql:"id"`
	ReferenceType        ReferenceType `sql:"reference_type"`
	ReferenceId          int           `sql:"reference_id"`
	FilterHistoryObjects string        `sql:"filter_history_objects"` //json of array of
	SubjectType          SubjectType   `sql:"subject_type"`
	SubjectIds           string        `sql:"subject_ids"` //comma seperated subject ids
	sql.AuditLog
}

type FilterEvaluationAuditRepository interface {
	//transaction util funcs
	sql.TransactionWrapper
	GetConnection() *pg.DB
	Create(filter *ResourceFilterEvaluationAudit) (*ResourceFilterEvaluationAudit, error)
	GetByRefAndSubject(referenceType ReferenceType, referenceId int, subjectType SubjectType, subjectIds []int) (*ResourceFilterEvaluationAudit, error)
	UpdateRefTypeAndRefId(id int, refType ReferenceType, refId int) error
}

type FilterEvaluationAuditRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewFilterEvaluationAuditRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *FilterEvaluationAuditRepositoryImpl {
	return &FilterEvaluationAuditRepositoryImpl{
		logger:              logger,
		dbConnection:        dbConnection,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}
}

func (repo *FilterEvaluationAuditRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo *FilterEvaluationAuditRepositoryImpl) Create(filter *ResourceFilterEvaluationAudit) (*ResourceFilterEvaluationAudit, error) {
	err := repo.dbConnection.Insert(filter)
	return filter, err
}

func (repo *FilterEvaluationAuditRepositoryImpl) GetByRefAndSubject(referenceType ReferenceType, referenceId int, subjectType SubjectType, subjectIds []int) (*ResourceFilterEvaluationAudit, error) {
	if len(subjectIds) == 0 {
		return nil, nil
	}
	res := &ResourceFilterEvaluationAudit{}
	err := repo.dbConnection.Model(res).
		Where("reference_type = ?", referenceType).
		Where("reference_id = ?", referenceId).
		Where("subject_type = ?", subjectType).
		Where("subject_ids = (?)", helper.GetCommaSepratedString(subjectIds)).
		Select()
	return res, err
}

func (repo *FilterEvaluationAuditRepositoryImpl) UpdateRefTypeAndRefId(id int, refType ReferenceType, refId int) error {
	var model ResourceFilterAudit
	_, err := repo.dbConnection.Model(&model).
		Set("reference_id = ?", refId).
		Set("reference_type = ?", refType).
		Where("id = ?", id).
		Update()
	return err
}
