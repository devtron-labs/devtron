//go:build wireinject
// +build wireinject

package main

import (
	cloudProviderIdentifier "github.com/devtron-labs/common-lib/cloud-provider-identifier"
	k8s2 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/argoApplication"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/fluxApplication"
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
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/cron"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		sql.NewNoopTransactionUtilImpl,
		sql.NewSqliteConnection,
		telemetry.NewPosthogClient,
		casbin.NewNoopEnforcer,
		wire.Bind(new(casbin.Enforcer), new(*casbin.NoopEnforcer)),
		cron.NewCronLoggerImpl,
		util3.GetEnvironmentVariables,
		cluster.ClusterWireSetK8sClient,
		dashboard.DashboardWireSet,
		service.NewNoopServiceImpl,
		wire.Bind(new(service.HelmAppService), new(*service.HelmAppServiceImpl)),
		k8s.K8sApplicationWireSetForK8sApp,
		terminal.TerminalWireSetK8sClient,
		k8s2.GetRuntimeConfig,

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

		cloudProviderIdentifier.NewProviderIdentifierServiceImpl,
		wire.Bind(new(cloudProviderIdentifier.ProviderIdentifierService), new(*cloudProviderIdentifier.ProviderIdentifierServiceImpl)),

		telemetry.NewK8sAppTelemetryEventClientImpl,
		wire.Bind(new(telemetry.TelemetryEventClient), new(*telemetry.TelemetryEventClientImpl)),

		argo.NewHelmUserServiceImpl,
		wire.Bind(new(argo.ArgoUserService), new(*argo.HelmUserServiceImpl)),

		kubernetesResourceAuditLogs.NewNoopServiceImpl,
		wire.Bind(new(kubernetesResourceAuditLogs.K8sResourceHistoryService), new(*kubernetesResourceAuditLogs.K8sResourceHistoryServiceImpl)),

		argoApplication.ArgoApplicationWireSetForK8sApp,
		fluxApplication.FluxApplicationWireSetForK8sApp,
	)
	return &App{}, nil
}
