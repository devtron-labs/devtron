package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DefaultRbacRoleDataRepository interface {
	GetAllDefaultRbacRole() ([]*DefaultRbacRoleDto, error)
}

type DefaultRbacRoleDto struct {
	TableName       struct{} `sql:"default_rbac_role_data" pg:",discard_unknown_columns"`
	Id              int      `sql:"id"`
	Role            string   `sql:"role"`
	DefaultRoleData string   `sql:"default_role_data"`
	Enabled         bool     `json:"enabled"`
	sql.AuditLog
}

type DefaultRbacRoleDataRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewDefaultRbacRoleDataRepositoryImpl(logger *zap.SugaredLogger, dbConnection *pg.DB) *DefaultRbacRoleDataRepositoryImpl {
	return &DefaultRbacRoleDataRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

func (repo *DefaultRbacRoleDataRepositoryImpl) GetAllDefaultRbacRole() ([]*DefaultRbacRoleDto, error) {
	var defaultRoles []*DefaultRbacRoleDto
	err := repo.dbConnection.Model(&defaultRoles).Select()
	if err == pg.ErrNoRows {
		repo.logger.Debugw("default rbac role not configured")
		return defaultRoles, nil
	}
	if err != nil {
		repo.logger.Errorw("error occurred while fetching default rbac roles", "err", err)
	}
	return defaultRoles, err
}
