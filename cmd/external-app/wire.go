//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		sql.PgSqlWireSet,
		//	team.TeamsWireSet,
		AuthWireSet,
		NewApp,
		util.NewSugardLogger,
	)
	return &App{}, nil
}
