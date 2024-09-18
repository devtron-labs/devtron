package sql

import (
	"github.com/devtron-labs/devtron/util/dir"
	"github.com/glebarez/sqlite"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"path"
)

type SqliteConnection struct {
	Migrator     SqliteMigrator
	DbConnection *gorm.DB
}

func NewSqliteConnection(logger *zap.SugaredLogger) *SqliteConnection {
	err, dbPath := createOrCheckSqliteDbPath(logger)
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		logger.Fatal("error occurred while opening db connection", "error", err)
	}
	migratorImpl := NewSqliteMigratorImpl(db, logger)
	return &SqliteConnection{migratorImpl, db}
}

func createOrCheckSqliteDbPath(logger *zap.SugaredLogger) (error, string) {
	err, devtronDirPath := dir.CheckOrCreateDevtronDir()
	if err != nil {
		logger.Errorw("error occurred while creating devtron dir ", "err", err)
		return err, ""
	}
	clusterDbPath := path.Join(devtronDirPath, "./client.db")
	return nil, clusterDbPath
}
