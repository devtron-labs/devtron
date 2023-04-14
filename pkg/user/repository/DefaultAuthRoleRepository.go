package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DefaultAuthRoleRepository interface {
	CreateRole(role *DefaultAuthRole) (*DefaultAuthRole, error)
	UpdateRole(role *DefaultAuthRole) (*DefaultAuthRole, error)
	GetRoleByRoleTypeAndEntityType(roleType bean.RoleType, accessType string, entity string) (role string, err error)
}

type DefaultAuthRole struct {
	TableName  struct{} `sql:"default_auth_role" pg:",discard_unknown_columns"`
	Id         int      `sql:"id,pk"`
	RoleType   string   `sql:"role_type,notnull"`
	Role       string   `sql:"role,notnull"`
	accessType string   `sql:"access_type"`
	entity     string   `sql:"entity,notnull"`
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

func (impl DefaultAuthRoleRepositoryImpl) GetRoleByRoleTypeAndEntityType(roleType bean.RoleType, accessType string, entity string) (role string, err error) {
	var model DefaultAuthRole
	query := "SELECT * FROM default_auth_role WHERE role_type = ? "
	query += " and entity = '" + entity + "' "
	if accessType == "" {
		query += "and access_type IS NULL ;"
	} else {
		query += "and access_type ='" + accessType + "' ;"
	}

	_, err = impl.dbConnection.Query(&model, query, roleType)
	if err != nil {
		impl.logger.Error("error in getting role by roleType", "err", err, "roleType", roleType)
		return "", err
	}
	return model.Role, nil
}
