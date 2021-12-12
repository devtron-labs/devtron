//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/user"
	"github.com/devtron-labs/devtron/client/argocdServer/session"
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
		team.TeamsWireSet,
		cluster.ClusterWireSetEa,

		NewApp,
		NewMuxRouter,

		util.NewSugardLogger,
		util.NewK8sUtil,
		util.IntValidator,

		//acd session client bind with authenticator login
		wire.Bind(new(session.ServiceClient), new(*middleware.LoginService)),
	)
	return &App{}, nil
}
