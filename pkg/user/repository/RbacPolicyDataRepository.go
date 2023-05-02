package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type RbacPolicyDataRepository interface {
	GetPolicyByRoleDetails(entity, accessType, role string) (*RbacPolicyData, error)
	GetPolicyDataForAllRoles() ([]*RbacPolicyData, error)
	CreateNewPolicyDataForRoleWithTxn(model *RbacPolicyData, tx *pg.Tx) (*RbacPolicyData, error)
	UpdatePolicyDataForRoleWithTxn(model *RbacPolicyData, tx *pg.Tx) (*RbacPolicyData, error)
}

type RbacPolicyDataRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewRbacPolicyDataRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *RbacPolicyDataRepositoryImpl {
	return &RbacPolicyDataRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type RbacPolicyData struct {
	TableName    struct{} `sql:"rbac_policy_data" pg:",discard_unknown_columns"`
	Id           int      `sql:"id"`
	Entity       string   `sql:"entity"`
	AccessType   string   `sql:"access_type"`
	Role         string   `sql:"role"`
	PolicyData   string   `sql:"policy_data"`
	IsPresetRole bool     `sql:"is_preset_role,notnull"`
	Deleted      bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *RbacPolicyDataRepositoryImpl) GetPolicyByRoleDetails(entity, accessType, role string) (*RbacPolicyData, error) {
	var model RbacPolicyData
	err := repo.dbConnection.Model(&model).Where("entity = ?", entity).
		Where("access_type = ?", accessType).Where("role = ?", role).
		Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting default policy by role detail", "err", err, "entity", entity, "accessType", accessType, "role", role)
		return nil, err
	}
	return &model, nil
}
func (repo *RbacPolicyDataRepositoryImpl) GetPolicyDataForAllRoles() ([]*RbacPolicyData, error) {
	var models []*RbacPolicyData
	err := repo.dbConnection.Model(&models).Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting policy data for all roles", "err", err)
		return nil, err
	}
	return models, nil
}

func (repo *RbacPolicyDataRepositoryImpl) CreateNewPolicyDataForRoleWithTxn(model *RbacPolicyData, tx *pg.Tx) (*RbacPolicyData, error) {
	_, err := tx.Model(model).Insert()
	if err != nil {
		repo.logger.Errorw("error in creating policy for a role", "err", err)
		return nil, err
	}
	return model, nil
}

func (repo *RbacPolicyDataRepositoryImpl) UpdatePolicyDataForRoleWithTxn(model *RbacPolicyData, tx *pg.Tx) (*RbacPolicyData, error) {
	_, err := tx.Model(model).WherePK().UpdateNotNull()
	if err != nil {
		repo.logger.Errorw("error in updating policy for a role", "err", err)
		return nil, err
	}
	return model, nil
}
