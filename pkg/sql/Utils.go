package sql

import "github.com/go-pg/pg"

type TransactionUtil interface {
	StartTx() (*pg.Tx, error)
	RollbackTx(tx *pg.Tx) error
	CommitTx(tx *pg.Tx) error
}

type TransactionUtilImpl struct {
	dbConnection *pg.DB
}

func NewTransactionUtilImpl(db *pg.DB) *TransactionUtilImpl {
	return &TransactionUtilImpl{
		dbConnection: db,
	}
}
func (impl *TransactionUtilImpl) RollbackTx(tx *pg.Tx) error {
	return tx.Rollback()
}
func (impl *TransactionUtilImpl) CommitTx(tx *pg.Tx) error {
	return tx.Commit()
}
func (impl *TransactionUtilImpl) StartTx() (*pg.Tx, error) {
	return impl.dbConnection.Begin()
}
