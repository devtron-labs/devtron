//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/authenticator/middleware"
	appStoreDiscover "github.com/devtron-labs/devtron/api/appStore/discover"
	chartRepo "github.com/devtron-labs/devtron/api/chartRepo"
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
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
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
		chartRepo.ChartRepositoryWireSet,
		appStoreDiscover.AppStoreDiscoverWireSet,

		NewApp,
		NewMuxRouter,

		util.NewHttpClient,
		util.NewSugardLogger,
		util.NewK8sUtil,
		util.IntValidator,
		util2.GetACDAuthConfig,
		telemetry.NewPosthogClient,
		telemetry.NewTelemetryEventClientImpl,
		delete2.NewDeleteServiceImpl,

		//acd session client bind with authenticator login
		wire.Bind(new(session.ServiceClient), new(*middleware.LoginService)),
		connector.NewPumpImpl,
		wire.Bind(new(connector.Pump), new(*connector.PumpImpl)),

		wire.Bind(new(delete2.DeleteService), new(*delete2.DeleteServiceImpl)),
	)
	return &App{}, nil
}
