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
	GetByName(name string) (*GlobalPolicy, error)
	GetAllByPolicyOfAndVersion(policyOf bean.GlobalPolicyType, policyVersion bean.GlobalPolicyVersion) ([]*GlobalPolicy, error)
	Create(model *GlobalPolicy) error
	Update(model *GlobalPolicy) error
	MarkDeletedById(id int, userId int32, tx *pg.Tx) error
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
	//set global policy dto
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

func (repo *GlobalPolicyRepositoryImpl) GetByName(name string) (*GlobalPolicy, error) {
	var model GlobalPolicy
	err := repo.dbConnection.Model(&model).Where("name = ?", name).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting policy by name", "err", err, "name", name)
		return nil, err
	}
	return &model, nil
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

func (repo *GlobalPolicyRepositoryImpl) Create(model *GlobalPolicy) error {
	err := repo.dbConnection.Insert(model)
	if err != nil {
		repo.logger.Errorw("error in creating global policy", "err", err, "model", model)
		return err
	}
	return nil
}

func (repo *GlobalPolicyRepositoryImpl) Update(model *GlobalPolicy) error {
	err := repo.dbConnection.Update(model)
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
