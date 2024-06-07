/*
 * Copyright (c) 2024. Devtron Inc.
 */

package repository

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"time"
)

type GlobalPolicyRepository interface {
	GetDbTransaction() (*pg.Tx, error)
	CommitTransaction(tx *pg.Tx) error
	RollBackTransaction(tx *pg.Tx) error
	GetById(id int) (*GlobalPolicy, error)
	GetEnabledPoliciesByIds(ids []int) ([]*GlobalPolicy, error)
	GetByName(name string, policyType bean.GlobalPolicyType) (*GlobalPolicy, error)
	GetByNames(names []string, policyType bean.GlobalPolicyType) ([]*GlobalPolicy, error)
	GetAllByPolicyOfAndVersion(policyOf bean.GlobalPolicyType, policyVersion bean.GlobalPolicyVersion) ([]*GlobalPolicy, error)
	Create(model *GlobalPolicy, tx *pg.Tx) error
	Update(model *GlobalPolicy, tx *pg.Tx) error
	UpdatePolicyByName(name string, model *GlobalPolicy, tx *pg.Tx) (*GlobalPolicy, error)
	MarkDeletedById(id int, userId int32, tx *pg.Tx) error
	GetByIds(ids []int) ([]*GlobalPolicy, error)
	GetPolicyByType(policyType *bean.GlobalPolicyType) (*GlobalPolicy, error)
	DeletedById(tx *pg.Tx, id int, userId int32) error
	DeletedByName(tx *pg.Tx, name string, policyType bean.GlobalPolicyType, userId int32) error
	GetByNameSearchKey(policyType bean.GlobalPolicyType, nameSearchKey string, orderByName bool, sortOrderDesc bool) ([]*GlobalPolicy, error)
	GetIdByName(name string, policyType bean.GlobalPolicyType) (int, error)
	GetAllActiveByType(policyType bean.GlobalPolicyType) ([]*GlobalPolicy, error)
}

type GlobalPolicyRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewGlobalPolicyRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *GlobalPolicyRepositoryImpl {
	return &GlobalPolicyRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type GlobalPolicy struct {
	tableName   struct{} `sql:"global_policy" pg:",discard_unknown_columns"`
	Id          int      `sql:"id,pk"`
	Name        string   `sql:"name"`
	PolicyOf    string   `sql:"policy_of"`
	Version     string   `sql:"version"`
	Description string   `sql:"description"`
	PolicyJson  string   `sql:"policy_json"`
	Enabled     bool     `sql:"enabled,notnull"`
	Deleted     bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (globalPolicy *GlobalPolicy) GetGlobalPolicyDto() (*bean.GlobalPolicyDto, error) {
	policyDetailDto := &bean.GlobalPolicyDetailDto{}
	err := json.Unmarshal([]byte(globalPolicy.PolicyJson), policyDetailDto)
	if err != nil {
		return nil, err
	}
	// set global policy dto
	return &bean.GlobalPolicyDto{
		Id:                    globalPolicy.Id,
		Name:                  globalPolicy.Name,
		PolicyOf:              bean.GlobalPolicyType(globalPolicy.PolicyOf),
		PolicyVersion:         bean.GlobalPolicyVersion(globalPolicy.Version),
		Description:           globalPolicy.Description,
		Enabled:               globalPolicy.Enabled,
		GlobalPolicyDetailDto: policyDetailDto,
	}, nil
}
func (globalPolicy *GlobalPolicy) GetGlobalPolicyBaseModel() *bean.GlobalPolicyBaseModel {
	return &bean.GlobalPolicyBaseModel{
		Id:            globalPolicy.Id,
		Name:          globalPolicy.Name,
		Description:   globalPolicy.Description,
		Enabled:       globalPolicy.Enabled,
		PolicyOf:      bean.GlobalPolicyType(globalPolicy.PolicyOf),
		PolicyVersion: bean.GlobalPolicyVersion(globalPolicy.Version),
		JsonData:      globalPolicy.PolicyJson,
		UserId:        globalPolicy.CreatedBy,
	}
}
func (repo *GlobalPolicyRepositoryImpl) GetDbTransaction() (*pg.Tx, error) {
	return repo.dbConnection.Begin()
}

func (repo *GlobalPolicyRepositoryImpl) CommitTransaction(tx *pg.Tx) error {
	return tx.Commit()
}
func (repo *GlobalPolicyRepositoryImpl) RollBackTransaction(tx *pg.Tx) error {
	return tx.Rollback()
}

func (repo *GlobalPolicyRepositoryImpl) GetById(id int) (*GlobalPolicy, error) {
	var model GlobalPolicy
	err := repo.dbConnection.Model(&model).Where("id = ?", id).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting policy by id", "err", err, "id", id)
		return nil, err
	}
	return &model, nil
}

func (repo *GlobalPolicyRepositoryImpl) GetEnabledPoliciesByIds(ids []int) ([]*GlobalPolicy, error) {
	var models []*GlobalPolicy
	err := repo.dbConnection.Model(&models).Where("id in (?)", pg.In(ids)).
		Where("enabled = ?", true).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting policy by ids", "err", err, "ids", ids)
		return nil, err
	}
	return models, nil
}

func (repo *GlobalPolicyRepositoryImpl) GetByName(name string, policyType bean.GlobalPolicyType) (*GlobalPolicy, error) {
	var model GlobalPolicy
	err := repo.dbConnection.Model(&model).
		Where("name = ?", name).
		Where("deleted = ?", false).
		Where("policy_of = ?", policyType).
		Select()
	if err != nil {
		return nil, err
	}
	return &model, err
}

func (repo *GlobalPolicyRepositoryImpl) GetByNames(names []string, policyType bean.GlobalPolicyType) ([]*GlobalPolicy, error) {
	policies := make([]*GlobalPolicy, 0)
	if len(names) == 0 {
		return policies, nil
	}
	err := repo.dbConnection.Model(&policies).
		Where("name IN (?)", pg.In(names)).
		Where("deleted = ?", false).
		Where("policy_of = ?", policyType).
		Select()
	return policies, err
}

func (repo *GlobalPolicyRepositoryImpl) GetAllByPolicyOfAndVersion(policyOf bean.GlobalPolicyType, policyVersion bean.GlobalPolicyVersion) ([]*GlobalPolicy, error) {
	var models []*GlobalPolicy
	err := repo.dbConnection.Model(&models).Where("policy_of = ?", policyOf).
		Where("version = ?", policyVersion).Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting all policies by policyOf and version", "err", err, "policyOf", policyOf, "version", policyVersion)
		return nil, err
	}
	return models, nil
}

func (repo *GlobalPolicyRepositoryImpl) Create(model *GlobalPolicy, tx *pg.Tx) error {
	var err error
	if tx == nil {
		err = repo.dbConnection.Insert(model)
	} else {
		err = tx.Insert(model)
	}
	if err != nil {
		repo.logger.Errorw("error in creating global policy", "err", err, "model", model)
		return err
	}
	return nil
}

func (repo *GlobalPolicyRepositoryImpl) Update(model *GlobalPolicy, tx *pg.Tx) error {
	var err error
	if tx == nil {
		err = repo.dbConnection.Update(model)
	} else {
		err = tx.Update(model)
	}
	if err != nil {
		repo.logger.Errorw("error in updating global policy", "err", err, "model", model)
		return err
	}
	return nil
}

func (repo *GlobalPolicyRepositoryImpl) MarkDeletedById(id int, userId int32, tx *pg.Tx) error {
	var model GlobalPolicy
	_, err := tx.Model(&model).Set("enabled = ?", false).
		Set("deleted = ?", true).Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", userId).Where("id = ?", id).Update()
	if err != nil {
		repo.logger.Errorw("error in marking global policy deleted", "err", err, "id", id)
		return err
	}
	return nil
}

func (repo *GlobalPolicyRepositoryImpl) GetByIds(ids []int) ([]*GlobalPolicy, error) {
	var models []*GlobalPolicy
	if len(ids) == 0 {
		return models, nil
	}
	err := repo.dbConnection.Model(&models).Where("id in (?)", pg.In(ids)).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting policy by ids", "err", err, "ids", ids)
		return nil, err
	}
	return models, nil
}
func (repo *GlobalPolicyRepositoryImpl) GetPolicyByType(policyType *bean.GlobalPolicyType) (*GlobalPolicy, error) {
	var model GlobalPolicy
	err := repo.dbConnection.Model(&model).Where("policy_of = ?", policyType).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting policy by name", "err", err, "policyType", policyType)
		return nil, err
	}
	return &model, nil
}

func (repo *GlobalPolicyRepositoryImpl) DeletedById(tx *pg.Tx, id int, userId int32) error {
	var model GlobalPolicy
	_, err := tx.Model(&model).
		Set("deleted = ?", true).Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", userId).Where("id = ?", id).Update()
	if err != nil {
		repo.logger.Errorw("error in marking global policy deleted", "err", err, "id", id)
		return err
	}
	return nil
}

func (repo *GlobalPolicyRepositoryImpl) DeletedByName(tx *pg.Tx, name string, policyType bean.GlobalPolicyType, userId int32) error {
	var model GlobalPolicy
	_, err := tx.Model(&model).
		Set("deleted = ?", true).Set("updated_on = ?", time.Now()).
		Set("updated_by = ?", userId).Where("name = ?", name).
		Where("policy_of = ?", policyType).
		Update()
	if err != nil {
		repo.logger.Errorw("error in marking global policy deleted", "err", err, "name", name)
		return err
	}
	return nil
}

func (repo *GlobalPolicyRepositoryImpl) GetByNameSearchKey(policyType bean.GlobalPolicyType, nameSearchKey string, orderByName bool, sortOrderDesc bool) ([]*GlobalPolicy, error) {
	var model []*GlobalPolicy
	query := repo.dbConnection.Model(&model).Where("name like ? ", nameSearchKey).
		Where("deleted = ?", false).Where("policy_of = ? ", policyType)

	if orderByName {
		orderExp := "name"
		if sortOrderDesc {
			orderExp = "name DESC"
		}
		query = query.OrderExpr(orderExp)
	}
	err := query.Select()
	if err != nil {
		repo.logger.Errorw("error in getting policy by name", "err", err, "nameSearchKey", nameSearchKey)
		return nil, err
	}
	return model, nil
}

func (repo *GlobalPolicyRepositoryImpl) GetIdByName(name string, policyType bean.GlobalPolicyType) (int, error) {
	id := 0
	err := repo.dbConnection.Model((*GlobalPolicy)(nil)).Column("id").
		Where("deleted = ?", false).
		Where("policy_of = ?", policyType).
		Where("name = ?", name).
		Select(&id)
	return id, err
}

func (repo *GlobalPolicyRepositoryImpl) GetAllActiveByType(policyType bean.GlobalPolicyType) ([]*GlobalPolicy, error) {
	var model []*GlobalPolicy
	err := repo.dbConnection.Model(&model).
		Where("deleted = ?", false).
		Where("policy_of = ?", policyType).
		Select()
	if err == pg.ErrNoRows {
		return model, nil
	}
	return model, err
}

func (repo *GlobalPolicyRepositoryImpl) UpdatePolicyByName(name string, model *GlobalPolicy, tx *pg.Tx) (*GlobalPolicy, error) {

	_, err := tx.Model(model).
		Set("description=?", model.Description).
		Set("name=?", model.Name).
		Set("policy_json=?", model.PolicyJson).
		Set("enabled=?", model.Enabled).
		Where("policy_of=?", model.PolicyOf).
		Where("deleted=?", false).
		Where("name=?", name).
		Update()
	if err != nil {
		repo.logger.Errorw("error in updating global policy", "err", err, "model", model)
	}
	return model, err
}
