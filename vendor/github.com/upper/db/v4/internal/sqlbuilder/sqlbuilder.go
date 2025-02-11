package sqlbuilder

import (
	"database/sql"
	"fmt"

	"github.com/upper/db/v4"
)

// Engine represents a SQL database engine.
type Engine interface {
	db.Session

	db.SQL
}

func lookupAdapter(adapterName string) (Adapter, error) {
	adapter := db.LookupAdapter(adapterName)
	if sqlAdapter, ok := adapter.(Adapter); ok {
		return sqlAdapter, nil
	}
	return nil, fmt.Errorf("%w %q", db.ErrMissingAdapter, adapterName)
}

func BindTx(adapterName string, tx *sql.Tx) (Tx, error) {
	adapter, err := lookupAdapter(adapterName)
	if err != nil {
		return nil, err
	}
	return adapter.NewTx(tx)
}

// Bind creates a binding between an adapter and a *sql.Tx or a *sql.DB.
func BindDB(adapterName string, sess *sql.DB) (db.Session, error) {
	adapter, err := lookupAdapter(adapterName)
	if err != nil {
		return nil, err
	}
	return adapter.New(sess)
}
