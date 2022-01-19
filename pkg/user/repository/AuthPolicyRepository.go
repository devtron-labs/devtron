package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type RoleType string

const (
	MANAGER_TYPE               RoleType = "manager"
	ADMIN_TYPE                 RoleType = "admin"
	TRIGGER_TYPE               RoleType = "trigger"
	VIEW_TYPE                  RoleType = "view"
	ENTITY_ALL_TYPE            RoleType = "entityAll"
	ENTITY_VIEW_TYPE           RoleType = "entityView"
	ENTITY_SPECIFIC_TYPE       RoleType = "entitySpecific"
	ENTITY_SPECIFIC_ADMIN_TYPE RoleType = "entitySpecificAdmin"
	ENTITY_SPECIFIC_VIEW_TYPE  RoleType = "entitySpecificView"
	ROLE_SPECIFIC_TYPE         RoleType = "roleSpecific"
)

type AuthPolicyRepository interface {
	CreatePolicy(policy *AuthPolicy) (*AuthPolicy, error)
	UpdatePolicy(policy *AuthPolicy) (*AuthPolicy, error)
	GetPolicyByRoleType(roleType RoleType) (policy string, err error)
}

type AuthPolicy struct {
	TableName struct{} `sql:"auth_policy" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	RoleType  string   `sql:"role_type,notnull"`
	Policy    string   `sql:"policy,notnull"`
	sql.AuditLog
}

type AuthPolicyRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewAuthPolicyRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *AuthPolicyRepositoryImpl {
	return &AuthPolicyRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl AuthPolicyRepositoryImpl) CreatePolicy(policy *AuthPolicy) (*AuthPolicy, error) {
	err := impl.dbConnection.Insert(policy)
	if err != nil {
		impl.logger.Error("error in creating auth policy", "err", err)
		return policy, err
	}
	return policy, nil
}

func (impl AuthPolicyRepositoryImpl) UpdatePolicy(policy *AuthPolicy) (*AuthPolicy, error) {
	err := impl.dbConnection.Update(policy)
	if err != nil {
		impl.logger.Error("error in updating auth policy", "err", err)
		return policy, err
	}
	return policy, nil
}

func (impl AuthPolicyRepositoryImpl) GetPolicyByRoleType(roleType RoleType) (policy string, err error) {
	var model AuthPolicy
	err = impl.dbConnection.Model(&model).Where("role = ?", roleType).Select()
	if err != nil {
		impl.logger.Error("error in getting policy by roleType", "err", err, "roleType", roleType)
		return "", err
	}
	return model.Policy, nil
}
