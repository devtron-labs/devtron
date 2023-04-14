package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type RbacPolicyDataRepository interface {
	GetPolicyDataForAllRoles() ([]*RbacPolicyData, error)
	CreateNewPolicyDataForRole(model *RbacPolicyData) (*RbacPolicyData, error)
	UpdatePolicyDataForRole(model *RbacPolicyData) (*RbacPolicyData, error)
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

func (repo *RbacPolicyDataRepositoryImpl) GetPolicyDataForAllRoles() ([]*RbacPolicyData, error) {
	var models []*RbacPolicyData
	err := repo.dbConnection.Model(&models).Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting policy data for all roles", "err", err)
		return nil, err
	}
	return models, nil
}

func (repo *RbacPolicyDataRepositoryImpl) CreateNewPolicyDataForRole(model *RbacPolicyData) (*RbacPolicyData, error) {
	_, err := repo.dbConnection.Model(&model).Insert()
	if err != nil {
		repo.logger.Errorw("error in creating policy for a role", "err", err)
		return nil, err
	}
	return model, nil
}

func (repo *RbacPolicyDataRepositoryImpl) UpdatePolicyDataForRole(model *RbacPolicyData) (*RbacPolicyData, error) {
	_, err := repo.dbConnection.Model(&model).Update()
	if err != nil {
		repo.logger.Errorw("error in updating policy for a role", "err", err)
		return nil, err
	}
	return model, nil
}
