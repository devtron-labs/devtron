package sql

import "github.com/google/wire"

var PgSqlWireSet = wire.NewSet(
	GetConfig,
	NewDbConnection,
	NewTransactionUtilImpl,
	wire.Bind(new(TransactionWrapper), new(*TransactionUtilImpl)),
)
