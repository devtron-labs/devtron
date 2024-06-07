/*
 * Copyright (c) 2024. Devtron Inc.
 */

package sql

import "github.com/google/wire"

var PgSqlWireSet = wire.NewSet(
	GetConfig,
	NewDbConnection,
	NewTransactionUtilImpl,
	wire.Bind(new(TransactionWrapper), new(*TransactionUtilImpl)),
)
