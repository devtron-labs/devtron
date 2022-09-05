//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/apiToken"
	appStoreDeployment "github.com/devtron-labs/devtron/api/appStore/deployment"
	appStoreDiscover "github.com/devtron-labs/devtron/api/appStore/discover"
	appStoreValues "github.com/devtron-labs/devtron/api/appStore/values"
	chartRepo "github.com/devtron-labs/devtron/api/chartRepo"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/dashboardEvent"
	"github.com/devtron-labs/devtron/api/externalLink"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/module"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/router"
	"github.com/devtron-labs/devtron/api/server"
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/user"
	webhookHelm "github.com/devtron-labs/devtron/api/webhook/helm"
	"github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreDeploymentTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool"
	appStoreDeploymentGitopsTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool/gitops"
	"github.com/devtron-labs/devtron/pkg/attributes"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		user.SelfRegistrationWireSet,

		sql.PgSqlWireSet,
		user.UserWireSet,
		sso.SsoConfigWireSet,
		AuthWireSet,
		externalLink.ExternalLinkWireSet,
		team.TeamsWireSet,
		cluster.ClusterWireSetEa,
		dashboard.DashboardWireSet,
		client.HelmAppWireSet,
		k8s.K8sApplicationWireSet,
		chartRepo.ChartRepositoryWireSet,
		appStoreDiscover.AppStoreDiscoverWireSet,
		appStoreValues.AppStoreValuesWireSet,
		appStoreDeployment.AppStoreDeploymentWireSet,
		server.ServerWireSet,
		module.ModuleWireSet,
		apiToken.ApiTokenWireSet,
		webhookHelm.WebhookHelmWireSet,

		NewApp,
		NewMuxRouter,
		util3.GetGlobalEnvVariables,
		util.NewHttpClient,
		util.NewSugardLogger,
		util.NewK8sUtil,
		util.IntValidator,
		util2.GetACDAuthConfig,
		telemetry.NewPosthogClient,
		delete2.NewDeleteServiceImpl,

		rbac.NewEnforcerUtilImpl,
		wire.Bind(new(rbac.EnforcerUtil), new(*rbac.EnforcerUtilImpl)),

		//acd session client bind with authenticator login
		wire.Bind(new(session.ServiceClient), new(*middleware.LoginService)),
		connector.NewPumpImpl,
		wire.Bind(new(connector.Pump), new(*connector.PumpImpl)),

		telemetry.NewTelemetryEventClientImpl,
		wire.Bind(new(telemetry.TelemetryEventClient), new(*telemetry.TelemetryEventClientImpl)),

		wire.Bind(new(delete2.DeleteService), new(*delete2.DeleteServiceImpl)),

		// needed for enforcer util
		pipelineConfig.NewPipelineRepositoryImpl,
		wire.Bind(new(pipelineConfig.PipelineRepository), new(*pipelineConfig.PipelineRepositoryImpl)),
		app2.NewAppRepositoryImpl,
		wire.Bind(new(app2.AppRepository), new(*app2.AppRepositoryImpl)),
		attributes.NewAttributesServiceImpl,
		wire.Bind(new(attributes.AttributesService), new(*attributes.AttributesServiceImpl)),
		repository.NewAttributesRepositoryImpl,
		wire.Bind(new(repository.AttributesRepository), new(*repository.AttributesRepositoryImpl)),
		pipelineConfig.NewCiPipelineRepositoryImpl,
		wire.Bind(new(pipelineConfig.CiPipelineRepository), new(*pipelineConfig.CiPipelineRepositoryImpl)),
		// // needed for enforcer util ends

		// binding gitops to helm (for hyperion)
		wire.Bind(new(appStoreDeploymentGitopsTool.AppStoreDeploymentArgoCdService), new(*appStoreDeploymentTool.AppStoreDeploymentHelmServiceImpl)),

		wire.Value(chartRepoRepository.RefChartDir("scripts/devtron-reference-helm-charts")),

		router.NewTelemetryRouterImpl,
		wire.Bind(new(router.TelemetryRouter), new(*router.TelemetryRouterImpl)),
		restHandler.NewTelemetryRestHandlerImpl,
		wire.Bind(new(restHandler.TelemetryRestHandler), new(*restHandler.TelemetryRestHandlerImpl)),

		//needed for sending events
		dashboardEvent.NewDashboardTelemetryRestHandlerImpl,
		wire.Bind(new(dashboardEvent.DashboardTelemetryRestHandler), new(*dashboardEvent.DashboardTelemetryRestHandlerImpl)),
		dashboardEvent.NewDashboardTelemetryRouterImpl,
		wire.Bind(new(dashboardEvent.DashboardTelemetryRouter),
			new(*dashboardEvent.DashboardTelemetryRouterImpl)),

		repository.NewGitOpsConfigRepositoryImpl,
		wire.Bind(new(repository.GitOpsConfigRepository), new(*repository.GitOpsConfigRepositoryImpl)),

		//binding argoUserService to helm via dummy implementation(HelmUserServiceImpl)
		argo.NewHelmUserServiceImpl,
		wire.Bind(new(argo.ArgoUserService), new(*argo.HelmUserServiceImpl)),

		router.NewUserAttributesRouterImpl,
		wire.Bind(new(router.UserAttributesRouter), new(*router.UserAttributesRouterImpl)),
		restHandler.NewUserAttributesRestHandlerImpl,
		wire.Bind(new(restHandler.UserAttributesRestHandler), new(*restHandler.UserAttributesRestHandlerImpl)),
		attributes.NewUserAttributesServiceImpl,
		wire.Bind(new(attributes.UserAttributesService), new(*attributes.UserAttributesServiceImpl)),
		repository.NewUserAttributesRepositoryImpl,
		wire.Bind(new(repository.UserAttributesRepository), new(*repository.UserAttributesRepositoryImpl)),
	)
	return &App{}, nil
}
