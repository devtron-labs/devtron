//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/connector"
	client2 "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/terminal"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/internal/util"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/pkg/user/noop"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		sql.NewNoopConnection,
		casbin.NewNoopEnforcer,
		wire.Bind(new(casbin.Enforcer), new(*casbin.NoopEnforcer)),
		cluster.ClusterWireSetK8sClient,
		dashboard.DashboardWireSet,
		client2.NewNoopServiceImpl,
		wire.Bind(new(client2.HelmAppService), new(*client2.HelmAppServiceImpl)),
		k8s.K8sApplicationWireSet,
		terminal.TerminalWireSetK8sClient,
		client.GetRuntimeConfig,

		noop.NewNoopUserService,
		wire.Bind(new(user.UserService), new(*noop.NoopUserService)),

		NewApp,
		NewMuxRouter,
		util.NewFileBaseSugaredLogger,
		util.NewK8sUtil,
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

		argo.NewHelmUserServiceImpl,
		wire.Bind(new(argo.ArgoUserService), new(*argo.HelmUserServiceImpl)),

		kubernetesResourceAuditLogs.NewNoopServiceImpl,
		wire.Bind(new(kubernetesResourceAuditLogs.K8sResourceHistoryService), new(*kubernetesResourceAuditLogs.K8sResourceHistoryServiceImpl)),
	)
	return &App{}, nil
}
