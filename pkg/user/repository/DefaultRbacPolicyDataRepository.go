package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DefaultRbacPolicyDataRepository interface {
	GetDefaultPolicyForAllRoles() ([]*DefaultRbacPolicyData, error)
	CreateNewDefaultPolicyForRole() (*DefaultRbacPolicyData, error)
	UpdateNewDefaultPolicyForRole() (*DefaultRbacPolicyData, error)
}

type DefaultRbacPolicyDataRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewDefaultRbacPolicyDataRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *DefaultRbacPolicyDataRepositoryImpl {
	return &DefaultRbacPolicyDataRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type DefaultRbacPolicyData struct {
	TableName  struct{} `sql:"default_rbac_policy_data" pg:",discard_unknown_columns"`
	Id         int      `sql:"id"`
	Entity     string   `sql:"entity"`
	AccessType string   `sql:"access_type"`
	Role       string   `sql:"role"`
	PolicyData string   `sql:"policy_data"`
	sql.AuditLog
}

func (repo *DefaultRbacPolicyDataRepositoryImpl) GetDefaultPolicyForAllRoles() ([]*DefaultRbacPolicyData, error) {
	var models []*DefaultRbacPolicyData
	err := repo.dbConnection.Model(&models).Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting default policy for all roles", "err", err)
		return nil, err
	}
	return models, nil
}

func (repo *DefaultRbacPolicyDataRepositoryImpl) CreateNewDefaultPolicyForRole() (*DefaultRbacPolicyData, error) {
	var model DefaultRbacPolicyData
	_, err := repo.dbConnection.Model(&model).Insert()
	if err != nil {
		repo.logger.Errorw("error in creating default policy for a role", "err", err)
		return nil, err
	}
	return &model, nil
}

func (repo *DefaultRbacPolicyDataRepositoryImpl) UpdateNewDefaultPolicyForRole() (*DefaultRbacPolicyData, error) {
	var model DefaultRbacPolicyData
	_, err := repo.dbConnection.Model(&model).Update()
	if err != nil {
		repo.logger.Errorw("error in updating default policy for a role", "err", err)
		return nil, err
	}
	return &model, nil
}
