package sql

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type SqliteMigrator interface {
	MigrateEntities(entities ...any)
}

type SqliteMigratorImpl struct {
	logger       *zap.SugaredLogger
	dbConnection *gorm.DB
}

func NewSqliteMigratorImpl(dbConnection *gorm.DB, logger *zap.SugaredLogger) SqliteMigrator {
	return &SqliteMigratorImpl{logger, dbConnection}
}

func (impl *SqliteMigratorImpl) MigrateEntities(entities ...interface{}) {
	err := impl.dbConnection.AutoMigrate(entities...)
	if err != nil {
		impl.logger.Fatalw("failed to migrate entities", "error", err)
	}
}
