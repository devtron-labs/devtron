//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/user"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		sql.PgSqlWireSet,
		user.UserWireSet,
		sso.SsoConfigWireSet,
		AuthWireSet,
		//team.TeamsWireSet,

		//	team.TeamsWireSet,
		NewApp,
		NewMuxRouter,


		util.NewSugardLogger,
		util.NewK8sUtil,
		util.IntValidator,
	)
	return &App{}, nil
}
