package resourceFilter

import (
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

type ResourceFilterEvaluation struct {
	tableName            struct{}      `sql:"resource_filter_evaluation" pg:",discard_unknown_columns"`
	Id                   int           `sql:"id"`
	ReferenceType        ReferenceType `sql:"reference_type"`
	ReferenceId          int           `sql:"reference_id"`
	FilterHistoryObjects string        `sql:"filter_history_objects"` //json of array of
	subjectType          SubjectType   `sql:"subject_type"`
	subjectIds           string        `sql:"subject_ids"` //comma seperated subject ids
}

type FilterEvaluationRepository interface {
	//transaction util funcs
	sql.TransactionWrapper
	GetConnection() *pg.DB
	CreateResourceFilterEvaluation(tx *pg.Tx, filter *ResourceFilterEvaluation) (*ResourceFilterEvaluation, error)
	GetResourceFilterEvaluation(referenceType ReferenceType, referenceId int, subjectType SubjectType, subjectId int) (*ResourceFilterEvaluation, error)
	UpdateResourceFilterEvaluation(tx *pg.Tx, filter *ResourceFilterEvaluation) (*ResourceFilterEvaluation, error)
}

type FilterEvaluationRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
	*sql.TransactionUtilImpl
}

func NewFilterEvaluationRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *FilterEvaluationRepositoryImpl {
	return &FilterEvaluationRepositoryImpl{
		logger:              logger,
		dbConnection:        dbConnection,
		TransactionUtilImpl: sql.NewTransactionUtilImpl(dbConnection),
	}
}

func (repo *FilterEvaluationRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo *FilterEvaluationRepositoryImpl) CreateResourceFilterEvaluation(tx *pg.Tx, filter *ResourceFilterEvaluation) (*ResourceFilterEvaluation, error) {
	err := tx.Insert(filter)
	return filter, err
}

func (repo *FilterEvaluationRepositoryImpl) GetResourceFilterEvaluation(referenceType ReferenceType, referenceId int, subjectType SubjectType, subjectId int) (*ResourceFilterEvaluation, error) {
	res := &ResourceFilterEvaluation{}
	err := repo.dbConnection.Model(res).
		Where("reference_type = ?", referenceType).
		Where("reference_id = ?", referenceId).
		Where("subject_type = ?", subjectType).
		Where("subject_id = ?", subjectId).
		Select()
	return res, err
}

func (repo *FilterEvaluationRepositoryImpl) UpdateResourceFilterEvaluation(tx *pg.Tx, filter *ResourceFilterEvaluation) (*ResourceFilterEvaluation, error) {
	err := tx.Update(filter)
	return filter, err
}
