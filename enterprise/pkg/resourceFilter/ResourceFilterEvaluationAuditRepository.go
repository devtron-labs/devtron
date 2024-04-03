package resourceFilter

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"sort"
	"time"
)

type ResourceFilterType int

const FILTER_CONDITION ResourceFilterType = 1
const ARTIFACT_PROMOTION_POLICY ResourceFilterType = 2
const DEPLOYMENT_WINDOW ResourceFilterType = 3

type SubjectType int

const Artifact SubjectType = 0
const CiMaterial SubjectType = 1

type ReferenceType int

const Pipeline ReferenceType = 0
const PipelineStage ReferenceType = 1
const CdWorkflowRunner ReferenceType = 2
const PrePipelineStageYaml ReferenceType = 3
const PostPipelineStageYaml ReferenceType = 4
const Deploy ReferenceType = 5
const PreDeploy ReferenceType = 6
const PostDeploy ReferenceType = 7

type ResourceFilterEvaluationAudit struct {
	tableName            struct{}           `sql:"resource_filter_evaluation_audit" pg:",discard_unknown_columns"`
	Id                   int                `sql:"id"`
	ReferenceType        *ReferenceType     `sql:"reference_type"`
	ReferenceId          int                `sql:"reference_id"`
	FilterHistoryObjects string             `sql:"filter_history_objects"` // json of array of
	SubjectType          *SubjectType       `sql:"subject_type"`
	SubjectId            int                `sql:"subject_id"` // comma seperated subject ids
	FilterType           ResourceFilterType `sql:"filter_type"`
	// add metadata column in future to store multi-git case for SubjectType CiPipelineMaterials
	sql.AuditLog
}

func NewResourceFilterEvaluationAudit(referenceType *ReferenceType, referenceId int, filterHistoryObjects string, subjectType *SubjectType, subjectId int, auditLog sql.AuditLog, filterType ResourceFilterType) ResourceFilterEvaluationAudit {
	return ResourceFilterEvaluationAudit{
		SubjectType:          subjectType,
		SubjectId:            subjectId,
		ReferenceType:        referenceType,
		ReferenceId:          referenceId,
		AuditLog:             auditLog,
		FilterHistoryObjects: filterHistoryObjects,
		FilterType:           filterType,
	}
}

type FilterEvaluationAuditRepository interface {
	// transaction util funcs
	sql.TransactionWrapper
	GetConnection() *pg.DB
	Create(tx *pg.Tx, filter *ResourceFilterEvaluationAudit) (*ResourceFilterEvaluationAudit, error)
	GetByRefAndMultiSubject(referenceType ReferenceType, referenceId int, subjectType SubjectType, subjectIds []int) ([]*ResourceFilterEvaluationAudit, error)
	GetLatestByRefAndMultiSubjectAndFilterType(referenceType ReferenceType, referenceId int, subjectType SubjectType, subjectIds []int, filterType ResourceFilterType) ([]*ResourceFilterEvaluationAudit, error)
	GetByMultiRefAndMultiSubject(referenceType ReferenceType, referenceIds []int, subjectType SubjectType, subjectIds []int) ([]*ResourceFilterEvaluationAudit, error)
	UpdateRefTypeAndRefId(id int, refType ReferenceType, refId int) error
	GetByIds(ids []int) ([]*ResourceFilterEvaluationAudit, error)
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

func (repo *FilterEvaluationAuditRepositoryImpl) Create(tx *pg.Tx, filter *ResourceFilterEvaluationAudit) (*ResourceFilterEvaluationAudit, error) {
	if tx != nil {
		err := tx.Insert(filter)
		return filter, err
	}
	err := repo.dbConnection.Insert(filter)
	return filter, err
}

func (repo *FilterEvaluationAuditRepositoryImpl) GetByMultiRefAndMultiSubject(referenceType ReferenceType, referenceIds []int, subjectType SubjectType, subjectIds []int) ([]*ResourceFilterEvaluationAudit, error) {
	res := make([]*ResourceFilterEvaluationAudit, 0)
	err := repo.dbConnection.Model(&res).
		Where("reference_type = ?", referenceType).
		Where("reference_id IN (?)", pg.In(referenceIds)).
		Where("subject_type = ?", subjectType).
		Where("subject_id IN (?) ", pg.In(subjectIds)).
		Where("resource_type = ?", FILTER_CONDITION).
		Select()
	if err == pg.ErrNoRows {
		return res, nil
	}
	return res, err
}

func (repo *FilterEvaluationAuditRepositoryImpl) GetLatestByRefAndMultiSubjectAndFilterType(referenceType ReferenceType, referenceId int, subjectType SubjectType, subjectIds []int, filterType ResourceFilterType) ([]*ResourceFilterEvaluationAudit, error) {

	var res []*ResourceFilterEvaluationAudit
	err := repo.dbConnection.Model(&res).
		Where("reference_type = ?", referenceType).
		Where("reference_id = ?", referenceId).
		Where("subject_type = ?", subjectType).
		Where("subject_id IN (?) ", pg.In(subjectIds)).
		Where("filter_type = ?", filterType).
		Select()
	if err == pg.ErrNoRows {
		return res, nil
	}

	subjectIdMap := make(map[int][]*ResourceFilterEvaluationAudit)
	for _, result := range res {
		subjectIdMap[result.SubjectId] = append(subjectIdMap[result.SubjectId], result)
	}

	finalListWithLatest := make([]*ResourceFilterEvaluationAudit, 0)
	for _, subjects := range subjectIdMap {
		sort.Slice(subjects, func(i, j int) bool {
			return subjects[i].Id > subjects[j].Id
		})
		finalListWithLatest = append(finalListWithLatest, subjects[0])
	}

	return finalListWithLatest, err
}

func (repo *FilterEvaluationAuditRepositoryImpl) GetByRefAndMultiSubject(referenceType ReferenceType, referenceId int, subjectType SubjectType, subjectIds []int) ([]*ResourceFilterEvaluationAudit, error) {
	res := make([]*ResourceFilterEvaluationAudit, 0)
	err := repo.dbConnection.Model(&res).
		Where("reference_type = ?", referenceType).
		Where("reference_id = ?", referenceId).
		Where("subject_type = ?", subjectType).
		Where("subject_id IN (?) ", pg.In(subjectIds)).
		Where("filter_type = ?", FILTER_CONDITION).
		Select()
	if err == pg.ErrNoRows {
		return res, nil
	}
	return res, err
}

func (repo *FilterEvaluationAuditRepositoryImpl) UpdateRefTypeAndRefId(id int, refType ReferenceType, refId int) error {
	var model ResourceFilterEvaluationAudit
	_, err := repo.dbConnection.Model(&model).
		Set("reference_id = ?", refId).
		Set("reference_type = ?", refType).
		Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", 1).
		Where("id = ?", id).
		Where("resource_type = ?", FILTER_CONDITION).
		Update()
	return err
}

func (repo *FilterEvaluationAuditRepositoryImpl) GetByIds(ids []int) ([]*ResourceFilterEvaluationAudit, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	models := make([]*ResourceFilterEvaluationAudit, 0)
	err := repo.dbConnection.Model(&models).
		Where("id IN (?)", pg.In(ids)).
		Select()

	return models, err
}
