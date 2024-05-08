package repository

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DevtronResourceSearchableKeyRepository interface {
	GetAll() ([]*DevtronResourceSearchableKey, error)
}

type DevtronResourceSearchableKeyRepositoryImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *pg.DB
}

func NewDevtronResourceSearchableKeyRepositoryImpl(logger *zap.SugaredLogger,
	dbConnection *pg.DB) *DevtronResourceSearchableKeyRepositoryImpl {
	return &DevtronResourceSearchableKeyRepositoryImpl{
		logger:       logger,
		dbConnection: dbConnection,
	}
}

type DevtronResourceSearchableKey struct {
	tableName struct{}                              `sql:"devtron_resource_searchable_key" pg:",discard_unknown_columns"`
	Id        int                                   `sql:"id,pk"`
	Name      bean.DevtronResourceSearchableKeyName `sql:"name"`
	IsRemoved bool                                  `sql:"is_removed,notnull"`
	sql.AuditLog
}

func (repo *DevtronResourceSearchableKeyRepositoryImpl) GetAll() ([]*DevtronResourceSearchableKey, error) {
	var models []*DevtronResourceSearchableKey
	err := repo.dbConnection.Model(&models).Where("is_removed = ?", false).Select()
	if err != nil {
		repo.logger.Errorw("error in getting all devtron resources searchable key", "err", err)
		return nil, err
	}
	return models, nil
}
