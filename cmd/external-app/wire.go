//go:build wireinject
// +build wireinject

package main

import (
	"github.com/devtron-labs/authenticator/middleware"
	util4 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/apiToken"
	chartProvider "github.com/devtron-labs/devtron/api/appStore/chartProvider"
	appStoreDeployment "github.com/devtron-labs/devtron/api/appStore/deployment"
	appStoreDiscover "github.com/devtron-labs/devtron/api/appStore/discover"
	appStoreValues "github.com/devtron-labs/devtron/api/appStore/values"
	chartRepo "github.com/devtron-labs/devtron/api/chartRepo"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/dashboardEvent"
	"github.com/devtron-labs/devtron/api/externalLink"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/k8s"
	"github.com/devtron-labs/devtron/api/module"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/router"
	"github.com/devtron-labs/devtron/api/server"
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/terminal"
	"github.com/devtron-labs/devtron/api/user"
	webhookHelm "github.com/devtron-labs/devtron/api/webhook/helm"
	"github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	security2 "github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	repository3 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDeploymentTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool"
	appStoreDeploymentGitopsTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool/gitops"
	"github.com/devtron-labs/devtron/pkg/attributes"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	repository2 "github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/sql"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
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
		util4.NewK8sUtil,
		externalLink.ExternalLinkWireSet,
		team.TeamsWireSet,
		cluster.ClusterWireSetEa,
		dashboard.DashboardWireSet,
		client.HelmAppWireSet,
		k8s.K8sApplicationWireSet,
		chartRepo.ChartRepositoryWireSet,
		appStoreDiscover.AppStoreDiscoverWireSet,
		chartProvider.AppStoreChartProviderWireSet,
		appStoreValues.AppStoreValuesWireSet,
		appStoreDeployment.AppStoreDeploymentWireSet,
		server.ServerWireSet,
		module.ModuleWireSet,
		apiToken.ApiTokenWireSet,
		webhookHelm.WebhookHelmWireSet,
		terminal.TerminalWireSet,

		NewApp,
		NewMuxRouter,
		util3.GetGlobalEnvVariables,
		util.NewHttpClient,
		util.NewSugardLogger,
		util.IntValidator,
		util2.GetACDAuthConfig,
		telemetry.NewPosthogClient,
		delete2.NewDeleteServiceImpl,

		pipelineConfig.NewMaterialRepositoryImpl,
		wire.Bind(new(pipelineConfig.MaterialRepository), new(*pipelineConfig.MaterialRepositoryImpl)),
		//appStatus
		appStatus.NewAppStatusRepositoryImpl,
		wire.Bind(new(appStatus.AppStatusRepository), new(*appStatus.AppStatusRepositoryImpl)),
		//appStatus ends
		rbac.NewEnforcerUtilImpl,
		wire.Bind(new(rbac.EnforcerUtil), new(*rbac.EnforcerUtilImpl)),

		router.NewAppRouterImpl,
		wire.Bind(new(router.AppRouter), new(*router.AppRouterImpl)),
		restHandler.NewAppRestHandlerImpl,
		wire.Bind(new(restHandler.AppRestHandler), new(*restHandler.AppRestHandlerImpl)),

		app.NewAppCrudOperationServiceImpl,
		wire.Bind(new(app.AppCrudOperationService), new(*app.AppCrudOperationServiceImpl)),
		pipelineConfig.NewAppLabelRepositoryImpl,
		wire.Bind(new(pipelineConfig.AppLabelRepository), new(*pipelineConfig.AppLabelRepositoryImpl)),
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
		router.NewAttributesRouterImpl,
		wire.Bind(new(router.AttributesRouter), new(*router.AttributesRouterImpl)),
		restHandler.NewAttributesRestHandlerImpl,
		wire.Bind(new(restHandler.AttributesRestHandler), new(*restHandler.AttributesRestHandlerImpl)),
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
		util3.GetDevtronSecretName,

		repository2.NewK8sResourceHistoryRepositoryImpl,
		wire.Bind(new(repository2.K8sResourceHistoryRepository), new(*repository2.K8sResourceHistoryRepositoryImpl)),

		kubernetesResourceAuditLogs.Newk8sResourceHistoryServiceImpl,
		wire.Bind(new(kubernetesResourceAuditLogs.K8sResourceHistoryService), new(*kubernetesResourceAuditLogs.K8sResourceHistoryServiceImpl)),

		util.NewChartTemplateServiceImpl,
		wire.Bind(new(util.ChartTemplateService), new(*util.ChartTemplateServiceImpl)),
		wire.Value(util.ChartWorkingDir("/tmp/charts/")),
		wire.Value(appStoreBean.RefChartProxyDir("scripts/devtron-reference-helm-charts")),
		util.NewGitFactory,
		util.NewGitCliUtil,

		security2.NewScanToolMetadataRepositoryImpl,
		wire.Bind(new(security2.ScanToolMetadataRepository), new(*security2.ScanToolMetadataRepositoryImpl)),
		//pubsub_lib.NewPubSubClientServiceImpl,

		// start: docker registry wire set injection
		router.NewDockerRegRouterImpl,
		wire.Bind(new(router.DockerRegRouter), new(*router.DockerRegRouterImpl)),
		restHandler.NewDockerRegRestHandlerImpl,
		wire.Bind(new(restHandler.DockerRegRestHandler), new(*restHandler.DockerRegRestHandlerImpl)),
		pipeline.NewDockerRegistryConfigImpl,
		wire.Bind(new(pipeline.DockerRegistryConfig), new(*pipeline.DockerRegistryConfigImpl)),
		dockerRegistryRepository.NewDockerArtifactStoreRepositoryImpl,
		wire.Bind(new(dockerRegistryRepository.DockerArtifactStoreRepository), new(*dockerRegistryRepository.DockerArtifactStoreRepositoryImpl)),
		dockerRegistryRepository.NewDockerRegistryIpsConfigRepositoryImpl,
		wire.Bind(new(dockerRegistryRepository.DockerRegistryIpsConfigRepository), new(*dockerRegistryRepository.DockerRegistryIpsConfigRepositoryImpl)),
		dockerRegistryRepository.NewOCIRegistryConfigRepositoryImpl,
		wire.Bind(new(dockerRegistryRepository.OCIRegistryConfigRepository), new(*dockerRegistryRepository.OCIRegistryConfigRepositoryImpl)),

		// chart group repository layer wire injection started
		repository3.NewChartGroupDeploymentRepositoryImpl,
		wire.Bind(new(repository3.ChartGroupDeploymentRepository), new(*repository3.ChartGroupDeploymentRepositoryImpl)),
		// chart group repository layer wire injection ended

		// end: docker registry wire set injection
	)
	return &App{}, nil
}
