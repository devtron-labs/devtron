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

type DefaultAuthPolicyRepository interface {
	CreatePolicy(policy *DefaultAuthPolicy) (*DefaultAuthPolicy, error)
	UpdatePolicyByRoleType(policy string, roleType RoleType) (*DefaultAuthPolicy, error)
	GetPolicyByRoleType(roleType RoleType) (policy string, err error)
}

type DefaultAuthPolicy struct {
	TableName struct{} `sql:"default_auth_policy" pg:",discard_unknown_columns"`
	Id        int      `sql:"id,pk"`
	RoleType  string   `sql:"role_type,notnull"`
	Policy    string   `sql:"policy,notnull"`
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

func (impl DefaultAuthPolicyRepositoryImpl) UpdatePolicyByRoleType(policy string, roleType RoleType) (*DefaultAuthPolicy, error) {
	var model DefaultAuthPolicy
	_, err := impl.dbConnection.Model(&model).Set("policy = ?", policy).
		Where("role_type = ?", roleType).Update()
	if err != nil {
		impl.logger.Error("error in updating auth policy", "err", err)
		return &model, err
	}
	return &model, nil
}

func (impl DefaultAuthPolicyRepositoryImpl) GetPolicyByRoleType(roleType RoleType) (policy string, err error) {
	var model DefaultAuthPolicy
	err = impl.dbConnection.Model(&model).Where("role_type = ?", roleType).Select()
	if err != nil {
		impl.logger.Error("error in getting policy by roleType", "err", err, "roleType", roleType)
		return "", err
	}
	return model.Policy, nil
}
