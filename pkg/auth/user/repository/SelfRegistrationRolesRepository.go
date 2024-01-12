package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type SelfRegistrationRolesRepository interface {
	GetAll() ([]SelfRegistrationRoles, error)
}

type SelfRegistrationRoles struct {
	TableName struct{} `sql:"self_registration_roles" pg:",discard_unknown_columns"`
	Role      string   `sql:"role,notnull"`
	sql.AuditLog
}

type SelfRegistrationRolesRepositoryImpl struct {
	dbConnection *pg.DB
	logger       *zap.SugaredLogger
}

func NewSelfRegistrationRolesRepositoryImpl(dbConnection *pg.DB, logger *zap.SugaredLogger) *SelfRegistrationRolesRepositoryImpl {
	return &SelfRegistrationRolesRepositoryImpl{dbConnection: dbConnection, logger: logger}
}

func (impl *SelfRegistrationRolesRepositoryImpl) GetAll() ([]SelfRegistrationRoles, error) {
	var models []SelfRegistrationRoles
	err := impl.dbConnection.Model(&models).Select()
	if err != nil {
		impl.logger.Error(err)
		return models, err
	}
	return models, nil
}
