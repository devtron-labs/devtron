package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user/bean"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DefaultAuthPolicyRepository interface {
	CreatePolicy(policy *DefaultAuthPolicy) (*DefaultAuthPolicy, error)
	UpdatePolicyByRoleType(policy string, roleType bean.RoleType) (*DefaultAuthPolicy, error)
	GetPolicyByRoleTypeAndEntity(roleType bean.RoleType, accessType string, entity string) (policy string, err error)
}

type DefaultAuthPolicy struct {
	TableName  struct{} `sql:"default_auth_policy" pg:",discard_unknown_columns"`
	Id         int      `sql:"id,pk"`
	RoleType   string   `sql:"role_type,notnull"`
	Policy     string   `sql:"policy,notnull"`
	accessType string   `sql:"access_type"`
	entity     string   `sql:"entity,notnull"`
	sql.AuditLog
}

type DefaultAuthPolicyRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewDefaultAuthPolicyRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *DefaultAuthPolicyRepositoryImpl {
	return &DefaultAuthPolicyRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl DefaultAuthPolicyRepositoryImpl) CreatePolicy(policy *DefaultAuthPolicy) (*DefaultAuthPolicy, error) {
	err := impl.dbConnection.Insert(policy)
	if err != nil {
		impl.logger.Error("error in creating auth policy", "err", err)
		return policy, err
	}
	return policy, nil
}

func (impl DefaultAuthPolicyRepositoryImpl) UpdatePolicyByRoleType(policy string, roleType bean.RoleType) (*DefaultAuthPolicy, error) {
	var model DefaultAuthPolicy
	_, err := impl.dbConnection.Model(&model).Set("policy = ?", policy).
		Where("role_type = ?", roleType).Update()
	if err != nil {
		impl.logger.Error("error in updating auth policy", "err", err)
		return &model, err
	}
	return &model, nil
}

func (impl DefaultAuthPolicyRepositoryImpl) GetPolicyByRoleTypeAndEntity(roleType bean.RoleType, accessType string, entity string) (policy string, err error) {
	var model DefaultAuthPolicy
	query := "SELECT * FROM default_auth_policy WHERE role_type = ? "
	query += " and entity = '" + entity + "' "
	if accessType == "" {
		query += "and access_type IS NULL ;"
	} else {
		query += "and access_type ='" + accessType + "' ;"
	}

	_, err = impl.dbConnection.Query(&model, query, roleType)
	if err != nil {
		impl.logger.Error("error in getting policy by roleType", "err", err, "roleType", roleType, "entity", entity)
		return "", err
	}
	return model.Policy, nil

}
