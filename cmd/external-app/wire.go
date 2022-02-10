//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/connector"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/user"
	"github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util/k8s"
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
		dashboard.DashboardWireSet,
		client.HelmAppWireSet,
		k8s.K8sApplicationWireSet,
		//telemetry.NewTelemetryEventClientWireSet,

		NewApp,
		NewMuxRouter,

		util.NewSugardLogger,
		util.NewK8sUtil,
		util.IntValidator,

		//acd session client bind with authenticator login
		wire.Bind(new(session.ServiceClient), new(*middleware.LoginService)),
		connector.NewPumpImpl,
		wire.Bind(new(connector.Pump), new(*connector.PumpImpl)),

		util.NewHttpClient,
		util2.GetACDAuthConfig,
		telemetry.NewPosthogClient,
		telemetry.NewTelemetryEventClientImpl,
		//wire.Bind(new(telemetry.TelemetryEventClient), new(*telemetry.TelemetryEventClientImpl)),
	)
	return &App{}, nil
}
