package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DefaultAuthRoleRepository interface {
	CreateRole(role *DefaultAuthRole) (*DefaultAuthRole, error)
	UpdateRole(role *DefaultAuthRole) (*DefaultAuthRole, error)
	GetRoleByRoleType(roleType RoleType) (role string, err error)
}

type DefaultAuthRole struct {
	TableName struct{} `sql:"default_auth_role" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	RoleType  string   `sql:"role_type,notnull"`
	Role      string   `sql:"role,notnull"`
	sql.AuditLog
}

type DefaultAuthRoleRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewDefaultAuthRoleRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *DefaultAuthRoleRepositoryImpl {
	return &DefaultAuthRoleRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl DefaultAuthRoleRepositoryImpl) CreateRole(role *DefaultAuthRole) (*DefaultAuthRole, error) {
	err := impl.dbConnection.Insert(role)
	if err != nil {
		impl.logger.Error("error in creating auth role", "err", err)
		return role, err
	}
	return role, nil
}

func (impl DefaultAuthRoleRepositoryImpl) UpdateRole(role *DefaultAuthRole) (*DefaultAuthRole, error) {
	err := impl.dbConnection.Update(role)
	if err != nil {
		impl.logger.Error("error in updating auth role", "err", err)
		return role, err
	}
	return role, nil
}

func (impl DefaultAuthRoleRepositoryImpl) GetRoleByRoleType(roleType RoleType) (role string, err error) {
	var model DefaultAuthRole
	err = impl.dbConnection.Model(&model).Where("role_type = ?", roleType).Select()
	if err != nil {
		impl.logger.Error("error in getting role by roleType", "err", err, "roleType", roleType)
		return "", err
	}
	return model.Role, nil
}
