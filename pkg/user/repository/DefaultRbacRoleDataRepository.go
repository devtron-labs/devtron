package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DefaultRbacRoleDataRepository interface {
	GetDefaultRoleDataForAllRoles() ([]*DefaultRbacRoleData, error)
	CreateNewDefaultRoleDataForRole() (*DefaultRbacRoleData, error)
	UpdateNewDefaultRoleDataForRole() (*DefaultRbacRoleData, error)
}

type DefaultRbacRoleDataRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewDefaultRbacRoleDataRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *DefaultRbacRoleDataRepositoryImpl {
	return &DefaultRbacRoleDataRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type DefaultRbacRoleData struct {
	TableName       struct{} `sql:"default_rbac_role_data" pg:",discard_unknown_columns"`
	Id              int      `sql:"id"`
	Entity          string   `sql:"entity"`
	AccessType      string   `sql:"access_type"`
	Role            string   `sql:"role"`
	RoleData        string   `sql:"role_data"`
	RoleDescription string   `sql:"role_description"`
	sql.AuditLog
}

func (repo *DefaultRbacRoleDataRepositoryImpl) GetDefaultRoleDataForAllRoles() ([]*DefaultRbacRoleData, error) {
	var models []*DefaultRbacRoleData
	err := repo.dbConnection.Model(&models).Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting default role data for all roles", "err", err)
		return nil, err
	}
	return models, nil
}

func (repo *DefaultRbacRoleDataRepositoryImpl) CreateNewDefaultRoleDataForRole() (*DefaultRbacRoleData, error) {
	var model DefaultRbacRoleData
	_, err := repo.dbConnection.Model(&model).Insert()
	if err != nil {
		repo.logger.Errorw("error in creating default role data for a role", "err", err)
		return nil, err
	}
	return &model, nil
}

func (repo *DefaultRbacRoleDataRepositoryImpl) UpdateNewDefaultRoleDataForRole() (*DefaultRbacRoleData, error) {
	var model DefaultRbacRoleData
	_, err := repo.dbConnection.Model(&model).Update()
	if err != nil {
		repo.logger.Errorw("error in updating default role data for a role", "err", err)
		return nil, err
	}
	return &model, nil
}
