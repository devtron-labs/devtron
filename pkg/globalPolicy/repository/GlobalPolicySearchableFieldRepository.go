package repository

import (
	"fmt"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"go.uber.org/zap"
	"time"
)

type GlobalPolicySearchableFieldRepository interface {
	CreateInBatchWithTxn(models []*GlobalPolicySearchableField, tx *pg.Tx) error
	DeleteByPolicyId(policyId int, tx *pg.Tx) error
	GetSearchableFields(searchableKeyIdValueMapWhereOrGroup, searchableKeyIdValueMapWhereAndGroup map[int][]string) ([]*GlobalPolicySearchableField, error)
	GetSearchableFieldByIds(policyId []int) ([]*GlobalPolicySearchableField, error)
	GetSortedPoliciesByPolicyKey(policyId []int, sortKey util.SearchableField, sortOrderDesc bool) ([]*GlobalPolicySearchableField, error)
}

type GlobalPolicySearchableFieldRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewGlobalPolicySearchableFieldRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *GlobalPolicySearchableFieldRepositoryImpl {
	return &GlobalPolicySearchableFieldRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type GlobalPolicySearchableField struct {
	tableName       struct{}                   `sql:"global_policy_searchable_field" pg:",discard_unknown_columns"`
	Id              int                        `sql:"id,pk"`
	FieldName       string                     `sql:"field_name"`
	GlobalPolicyId  int                        `sql:"global_policy_id"`
	SearchableKeyId int                        `sql:"searchable_key_id"`
	Value           string                     `sql:"value"`
	ValueInt        int                        `sql:"value_int"`
	ValueTimeStamp  time.Time                  `sql:"value_time_stamp"`
	IsRegex         bool                       `sql:"is_regex,notnull"`
	PolicyComponent bean.GlobalPolicyComponent `sql:"policy_component"`
	sql.AuditLog
}

func (repo *GlobalPolicySearchableFieldRepositoryImpl) CreateInBatchWithTxn(models []*GlobalPolicySearchableField, tx *pg.Tx) error {
	err := tx.Insert(&models)
	if err != nil {
		repo.logger.Errorw("error in creating global policy searchable fields", "err", err, "models", models)
		return err
	}
	return nil
}

func (repo *GlobalPolicySearchableFieldRepositoryImpl) DeleteByPolicyId(policyId int, tx *pg.Tx) error {
	var model GlobalPolicySearchableField
	_, err := tx.Model(&model).Where("global_policy_id = ?", policyId).Delete()
	if err != nil {
		repo.logger.Errorw("error in deleting global policy searchable fields by policyId", "err", err, "policyId", policyId)
		return err
	}
	return nil
}

func (repo *GlobalPolicySearchableFieldRepositoryImpl) GetSearchableFields(searchableKeyIdValueMapWhereOrGroup, searchableKeyIdValueMapWhereAndGroup map[int][]string) ([]*GlobalPolicySearchableField, error) {
	var models []*GlobalPolicySearchableField
	q := repo.dbConnection.Model(&models)
	q.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
		for searchableKeyId, searchableKeyValues := range searchableKeyIdValueMapWhereOrGroup {
			q.WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
				q = q.Where("searchable_key_id = ?", searchableKeyId).
					Where("value in (?)", pg.In(searchableKeyValues))
				return q, nil
			})
		}
		return q, nil
	})
	// for searchableKeyId, searchableKeyValues := range searchableKeyIdValueMapWhereAndGroup {
	//	q.WhereOrGroup(func(q *orm.Query) (*orm.Query, error) {
	//		q = q.Where("searchable_key_id = ?", searchableKeyId).
	//			Where("value in (?)", pg.In(searchableKeyValues))
	//		return q, nil
	//	})
	// }
	// adding is_regex fields always
	q.WhereOr("is_regex = ?", true)
	err := q.Select()
	if err != nil {
		repo.logger.Errorw("error, GetSearchableFields", "err", err, "searchableKeyIdValueMapWhereOr", searchableKeyIdValueMapWhereOrGroup, "searchableKeyIdValueMapWhereAnd", searchableKeyIdValueMapWhereAndGroup)
		return nil, err
	}

	var filterredPolicyIds []int
	for _, searchableField := range models {
		filterredPolicyIds = append(filterredPolicyIds, searchableField.GlobalPolicyId)
	}
	if len(filterredPolicyIds) == 0 {
		return models, nil
	}
	var finalResult []*GlobalPolicySearchableField

	q1 := repo.dbConnection.Model(&finalResult)
	for searchableKeyId, searchableKeyValues := range searchableKeyIdValueMapWhereAndGroup {
		q1.WhereGroup(func(q1 *orm.Query) (*orm.Query, error) {
			q1 = q1.Where("searchable_key_id = ?", searchableKeyId).
				Where("value in (?)", pg.In(searchableKeyValues))
			return q1, nil
		})
	}
	q1.Where("global_policy_id in (?)", pg.In(filterredPolicyIds))
	err = q1.Select()
	if err != nil {
		repo.logger.Errorw("error, while fetching actionable search fields", "err", err, "searchableKeyIdValueMapWhereOr", searchableKeyIdValueMapWhereOrGroup, "searchableKeyIdValueMapWhereAnd", searchableKeyIdValueMapWhereAndGroup)
		return nil, err
	}

	return finalResult, nil
}
func (repo *GlobalPolicySearchableFieldRepositoryImpl) GetSearchableFieldByIds(policyIds []int) ([]*GlobalPolicySearchableField, error) {
	var models []*GlobalPolicySearchableField
	err := repo.dbConnection.Model(&models).
		Where("global_policy_id IN (?)", pg.In(policyIds)).
		Select()
	if err != nil {
		repo.logger.Errorw("error in fetching GlobalPolicySearchableField", "policyIds", policyIds, "err", err)
		return nil, err
	}
	return models, err
}

func (repo *GlobalPolicySearchableFieldRepositoryImpl) GetSortedPoliciesByPolicyKey(policyId []int, sortKey util.SearchableField, sortOrderDesc bool) ([]*GlobalPolicySearchableField, error) {

	var models []*GlobalPolicySearchableField

	var orderExp string
	switch sortKey.FieldType {
	case util.NumericType:
		orderExp = "value"
	case util.StringType:
		orderExp = "value_int"
	case util.DateTimeType:
		orderExp = "value_time_stamp"
	}
	if sortOrderDesc {
		orderExp = fmt.Sprintf("%s DESC", orderExp)
	}
	err := repo.dbConnection.Model(&models).
		Where("global_policy_id IN (?)", pg.In(policyId)).
		Where("field_name = ? ", sortKey.FieldName).
		OrderExpr(orderExp).
		Select()
	if err != nil {
		repo.logger.Errorw("error in fetching GlobalPolicySearchableField", "policyIds", policyId, "err", err)
		return nil, err
	}
	return models, nil
}
