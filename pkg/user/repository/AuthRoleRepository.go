package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type AuthRoleRepository interface {
	CreateRole(role *AuthRole) (*AuthRole, error)
	UpdateRole(role *AuthRole) (*AuthRole, error)
	GetRoleByRoleType(roleType RoleType) (role string, err error)
}

type AuthRole struct {
	TableName struct{} `sql:"auth_role" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	RoleType  string   `sql:"role_type,notnull"`
	Role      string   `sql:"role,notnull"`
	sql.AuditLog
}

type AuthRoleRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewAuthRoleRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *AuthRoleRepositoryImpl {
	return &AuthRoleRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl AuthRoleRepositoryImpl) CreateRole(role *AuthRole) (*AuthRole, error) {
	err := impl.dbConnection.Insert(role)
	if err != nil {
		impl.logger.Error("error in creating auth role", "err", err)
		return role, err
	}
	return role, nil
}

func (impl AuthRoleRepositoryImpl) UpdateRole(role *AuthRole) (*AuthRole, error) {
	err := impl.dbConnection.Update(role)
	if err != nil {
		impl.logger.Error("error in updating auth role", "err", err)
		return role, err
	}
	return role, nil
}

func (impl AuthRoleRepositoryImpl) GetRoleByRoleType(roleType RoleType) (role string, err error) {
	var model AuthRole
	err = impl.dbConnection.Model(&model).Where("role = ?", roleType).Select()
	if err != nil {
		impl.logger.Error("error in getting role by roleType", "err", err, "roleType", roleType)
		return "", err
	}
	return model.Role, nil
}
