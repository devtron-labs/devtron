package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type RbacRoleDataRepository interface {
	GetRoleDataForAllRoles() ([]*RbacRoleData, error)
	CreateNewRoleDataForRoleWithTxn(model *RbacRoleData, tx *pg.Tx) (*RbacRoleData, error)
	UpdateRoleDataForRoleWithTxn(model *RbacRoleData, tx *pg.Tx) (*RbacRoleData, error)
}

type RbacRoleDataRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewRbacRoleDataRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *RbacRoleDataRepositoryImpl {
	return &RbacRoleDataRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type RbacRoleData struct {
	TableName       struct{} `sql:"rbac_role_data" pg:",discard_unknown_columns"`
	Id              int      `sql:"id"`
	Entity          string   `sql:"entity"`
	AccessType      string   `sql:"access_type"`
	Role            string   `sql:"role"`
	RoleData        string   `sql:"role_data"`
	RoleDisplayName string   `sql:"role_display_name"`
	RoleDescription string   `sql:"role_description"`
	IsPresetRole    bool     `sql:"is_preset_role,notnull"`
	Deleted         bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *RbacRoleDataRepositoryImpl) GetRoleDataForAllRoles() ([]*RbacRoleData, error) {
	var models []*RbacRoleData
	err := repo.dbConnection.Model(&models).Where("deleted = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting role data for all roles", "err", err)
		return nil, err
	}
	return models, nil
}

func (repo *RbacRoleDataRepositoryImpl) CreateNewRoleDataForRoleWithTxn(model *RbacRoleData, tx *pg.Tx) (*RbacRoleData, error) {
	_, err := tx.Model(model).Insert()
	if err != nil {
		repo.logger.Errorw("error in creating role data for a role", "err", err)
		return nil, err
	}
	return model, nil
}

func (repo *RbacRoleDataRepositoryImpl) UpdateRoleDataForRoleWithTxn(model *RbacRoleData, tx *pg.Tx) (*RbacRoleData, error) {
	_, err := tx.Model(model).UpdateNotNull()
	if err != nil {
		repo.logger.Errorw("error in updating role data for a role", "err", err)
		return nil, err
	}
	return model, nil
}
