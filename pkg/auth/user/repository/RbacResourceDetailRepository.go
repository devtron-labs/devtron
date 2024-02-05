package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

//----------------------------------------RBAC Policy Resource Detail Repository

type RbacPolicyResourceDetailRepository interface {
	GetConnection() *pg.DB
	GetAllPolicyResourceDetail() ([]*RbacPolicyResourceDetail, error)
	GetPolicyResourceDetailByEntityAccessType(entityAccessType string) ([]*RbacPolicyResourceDetail, error)
	SaveNewPolicyResourceDetail(model *RbacPolicyResourceDetail) (*RbacPolicyResourceDetail, error)
	UpdatePolicyResourceDetail(model *RbacPolicyResourceDetail) (*RbacPolicyResourceDetail, error)
}

type RbacPolicyResourceDetailRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewRbacPolicyResourceDetailRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *RbacPolicyResourceDetailRepositoryImpl {
	return &RbacPolicyResourceDetailRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type RbacPolicyResourceDetail struct {
	TableName                 struct{} `sql:"rbac_policy_resource_detail" pg:",discard_unknown_columns"`
	Id                        int      `sql:"id"`
	Resource                  string   `sql:"resource"`
	PolicyResourceValue       string   `sql:"policy_resource_value"`
	AllowedActions            []string `sql:"allowed_actions" pg:",array"`
	ResourceObject            string   `sql:"resource_object"`
	EligibleEntityAccessTypes []string `sql:"eligible_entity_access_types"  pg:",array"` //array of "entity/accessType" or "entity" which have this resource
	Deleted                   bool     `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *RbacPolicyResourceDetailRepositoryImpl) GetConnection() *pg.DB {
	return repo.dbConnection
}

func (repo *RbacPolicyResourceDetailRepositoryImpl) GetAllPolicyResourceDetail() ([]*RbacPolicyResourceDetail, error) {
	var models []*RbacPolicyResourceDetail
	err := repo.dbConnection.Model(&models).
		Select()
	if err != nil {
		repo.logger.Error("error, GetAllPolicyResourceDetail", "err", err)
		return nil, err
	}
	return models, nil
}

func (repo *RbacPolicyResourceDetailRepositoryImpl) GetPolicyResourceDetailByEntityAccessType(entityAccessType string) ([]*RbacPolicyResourceDetail, error) {
	var models []*RbacPolicyResourceDetail
	err := repo.dbConnection.Model(&models).
		Where("? = ANY(eligible_entity_access_types)", entityAccessType).
		Select()
	if err != nil {
		repo.logger.Error("error, GetPolicyResourceDetailByEntityAccessType", "err", err, "entityAccessType", entityAccessType)
		return nil, err
	}
	return models, nil
}

func (repo *RbacPolicyResourceDetailRepositoryImpl) SaveNewPolicyResourceDetail(model *RbacPolicyResourceDetail) (*RbacPolicyResourceDetail, error) {
	_, err := repo.dbConnection.Model(model).Insert()
	if err != nil {
		repo.logger.Error("error, SaveNewPolicyResourceDetail", "err", err, "model", model)
		return nil, err
	}
	return model, nil
}

func (repo *RbacPolicyResourceDetailRepositoryImpl) UpdatePolicyResourceDetail(model *RbacPolicyResourceDetail) (*RbacPolicyResourceDetail, error) {
	_, err := repo.dbConnection.Model(model).Update()
	if err != nil {
		repo.logger.Error("error, UpdatePolicyResourceDetail", "err", err, "model", model)
		return nil, err
	}
	return model, nil
}

//----------------------------------------RBAC Role Resource Detail Repository

type RbacRoleResourceDetailRepository interface {
	GetAllRoleResourceDetail() ([]*RbacRoleResourceDetail, error)
	GetRoleResourceDetailByEntityAccessType(entityAccessType string) ([]*RbacRoleResourceDetail, error)
	SaveNewRoleResourceDetail(model *RbacRoleResourceDetail) (*RbacRoleResourceDetail, error)
	UpdateRoleResourceDetail(model *RbacRoleResourceDetail) (*RbacRoleResourceDetail, error)
}

type RbacRoleResourceDetailRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewRbacRoleResourceDetailRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *RbacRoleResourceDetailRepositoryImpl {
	return &RbacRoleResourceDetailRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type RbacRoleResourceDetail struct {
	TableName                 struct{}      `sql:"rbac_role_resource_detail" pg:",discard_unknown_columns"`
	Id                        int           `sql:"id"`
	Resource                  string        `sql:"resource"`
	RoleResourceKey           string        `sql:"role_resource_key"`
	RoleResourceUpdateKey     PValUpdateKey `sql:"role_resource_update_key"`
	EligibleEntityAccessTypes []string      `sql:"eligible_entity_access_types" pg:",array"` //array of "entity/accessType" or "entity" which have this resource
	Deleted                   bool          `sql:"deleted,notnull"`
	sql.AuditLog
}

func (repo *RbacRoleResourceDetailRepositoryImpl) GetAllRoleResourceDetail() ([]*RbacRoleResourceDetail, error) {
	var models []*RbacRoleResourceDetail
	err := repo.dbConnection.Model(&models).
		Select()
	if err != nil {
		repo.logger.Error("error, GetAllRoleResourceDetail", "err", err)
		return nil, err
	}
	return models, nil
}

func (repo *RbacRoleResourceDetailRepositoryImpl) GetRoleResourceDetailByEntityAccessType(entityAccessType string) ([]*RbacRoleResourceDetail, error) {
	var models []*RbacRoleResourceDetail
	err := repo.dbConnection.Model(&models).
		Where("? = ANY(eligible_entity_access_types)", entityAccessType).
		Select()
	if err != nil {
		repo.logger.Error("error, GetRoleResourceDetailByEntityAccessType", "err", err, "entityAccessType", entityAccessType)
		return nil, err
	}
	return models, nil
}

func (repo *RbacRoleResourceDetailRepositoryImpl) SaveNewRoleResourceDetail(model *RbacRoleResourceDetail) (*RbacRoleResourceDetail, error) {
	_, err := repo.dbConnection.Model(model).Insert()
	if err != nil {
		repo.logger.Error("error, SaveNewRoleResourceDetail", "err", err, "model", model)
		return nil, err
	}
	return model, nil
}

func (repo *RbacRoleResourceDetailRepositoryImpl) UpdateRoleResourceDetail(model *RbacRoleResourceDetail) (*RbacRoleResourceDetail, error) {
	_, err := repo.dbConnection.Model(model).Update()
	if err != nil {
		repo.logger.Error("error, UpdateRoleResourceDetail", "err", err, "model", model)
		return nil, err
	}
	return model, nil
}
