//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/authenticator/client"
	k8s2 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/api/k8s"
	"github.com/devtron-labs/devtron/api/terminal"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	noop2 "github.com/devtron-labs/devtron/pkg/auth/user/noop"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		sql.NewNoopConnection,
		telemetry.NewPosthogClient,
		casbin.NewNoopEnforcer,
		wire.Bind(new(casbin.Enforcer), new(*casbin.NoopEnforcer)),
		cluster.ClusterWireSetK8sClient,
		dashboard.DashboardWireSet,
		service.NewNoopServiceImpl,
		wire.Bind(new(service.HelmAppService), new(*service.HelmAppServiceImpl)),
		k8s.K8sApplicationWireSet,
		terminal.TerminalWireSetK8sClient,
		client.GetRuntimeConfig,

		noop2.NewNoopUserService,
		wire.Bind(new(user.UserService), new(*noop2.NoopUserService)),

		NewApp,
		NewMuxRouter,
		util.NewHttpClient,
		util.NewFileBaseSugaredLogger,
		k8s2.NewK8sUtil,
		util.IntValidator,
		util2.GetACDAuthConfig,
		wire.Bind(new(delete2.DeleteService), new(*delete2.DeleteServiceImpl)),
		delete2.NewNoopServiceImpl,

		rbac.NewNoopEnforcerUtilHelm,
		wire.Bind(new(rbac.EnforcerUtilHelm), new(*rbac.EnforcerUtilHelmImpl)),

		rbac.NewNoopEnforcerUtil,
		wire.Bind(new(rbac.EnforcerUtil), new(*rbac.EnforcerUtilImpl)),

		connector.NewPumpImpl,
		wire.Bind(new(connector.Pump), new(*connector.PumpImpl)),

		telemetry.NewK8sAppTelemetryEventClientImpl,
		wire.Bind(new(telemetry.TelemetryEventClient), new(*telemetry.TelemetryEventClientImpl)),

		argo.NewHelmUserServiceImpl,
		wire.Bind(new(argo.ArgoUserService), new(*argo.HelmUserServiceImpl)),

		kubernetesResourceAuditLogs.NewNoopServiceImpl,
		wire.Bind(new(kubernetesResourceAuditLogs.K8sResourceHistoryService), new(*kubernetesResourceAuditLogs.K8sResourceHistoryServiceImpl)),
	)
	return &App{}, nil
}
