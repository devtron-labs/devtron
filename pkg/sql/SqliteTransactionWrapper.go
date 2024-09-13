package sql

import (
	"github.com/go-pg/pg"
	"gorm.io/gorm"
)

type SqliteTransactionUtilImpl struct {
	dbConnection *gorm.DB
}

func NewSqliteTransactionUtilImpl(dbConnection *gorm.DB) *SqliteTransactionUtilImpl {
	return &SqliteTransactionUtilImpl{dbConnection}
}

func (impl SqliteTransactionUtilImpl) StartTx() (*pg.Tx, error) {
	impl.dbConnection.Begin()
}

func (impl SqliteTransactionUtilImpl) RollbackTx(tx *pg.Tx) error {
	//TODO implement me
	panic("implement me")
}

func (impl SqliteTransactionUtilImpl) CommitTx(tx *pg.Tx) error {
	//TODO implement me
	panic("implement me")
}
