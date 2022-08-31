//go:build wireinject
// +build wireinject

/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/devtron/api/apiToken"
	appStoreRestHandler "github.com/devtron-labs/devtron/api/appStore"
	appStoreDeployment "github.com/devtron-labs/devtron/api/appStore/deployment"
	appStoreDiscover "github.com/devtron-labs/devtron/api/appStore/discover"
	appStoreValues "github.com/devtron-labs/devtron/api/appStore/values"
	chartRepo "github.com/devtron-labs/devtron/api/chartRepo"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/dashboardEvent"
	"github.com/devtron-labs/devtron/api/deployment"
	"github.com/devtron-labs/devtron/api/externalLink"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/module"
	"github.com/devtron-labs/devtron/api/restHandler"
	pipeline2 "github.com/devtron-labs/devtron/api/restHandler/app"
	"github.com/devtron-labs/devtron/api/router"
	"github.com/devtron-labs/devtron/api/router/pubsub"
	"github.com/devtron-labs/devtron/api/server"
	"github.com/devtron-labs/devtron/api/sse"
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/user"
	webhookHelm "github.com/devtron-labs/devtron/api/webhook/helm"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	cluster2 "github.com/devtron-labs/devtron/client/argocdServer/cluster"
	repository2 "github.com/devtron-labs/devtron/client/argocdServer/repository"
	session2 "github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/client/cron"
	"github.com/devtron-labs/devtron/client/dashboard"
	eClient "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/client/grafana"
	jClient "github.com/devtron-labs/devtron/client/jira"
	"github.com/devtron-labs/devtron/client/lens"
	pubsub2 "github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	appWorkflow2 "github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/bulkUpdate"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	security2 "github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/internal/util/ArgoUtil"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appClone"
	"github.com/devtron-labs/devtron/pkg/appClone/batch"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentFullMode "github.com/devtron-labs/devtron/pkg/appStore/deployment/fullMode"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	appStoreDeploymentGitopsTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool/gitops"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/commonService"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/git"
	"github.com/devtron-labs/devtron/pkg/gitops"
	jira2 "github.com/devtron-labs/devtron/pkg/jira"
	"github.com/devtron-labs/devtron/pkg/notifier"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	history3 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	repository5 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/plugin"
	repository6 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/projectManagementService/jira"
	"github.com/devtron-labs/devtron/pkg/security"
	"github.com/devtron-labs/devtron/pkg/sql"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/devtron-labs/devtron/util/session"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {

	wire.Build(
		// ----- wireset start
		sql.PgSqlWireSet,
		user.SelfRegistrationWireSet,
		externalLink.ExternalLinkWireSet,
		team.TeamsWireSet,
		AuthWireSet,
		user.UserWireSet,
		sso.SsoConfigWireSet,
		cluster.ClusterWireSet,
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
		// -------wireset end ----------
		gitSensor.GetGitSensorConfig,
		gitSensor.NewGitSensorSession,
		wire.Bind(new(gitSensor.GitSensorClient), new(*gitSensor.GitSensorClientImpl)),
		//--------
		helper.NewAppListingRepositoryQueryBuilder,
		//sql.GetConfig,
		eClient.GetEventClientConfig,
		util2.GetGlobalEnvVariables,
		//sql.NewDbConnection,
		//app.GetACDAuthConfig,
		util3.GetACDAuthConfig,
		wire.Value(chartRepoRepository.RefChartDir("scripts/devtron-reference-helm-charts")),
		wire.Value(appStoreBean.RefChartProxyDir("scripts/devtron-reference-helm-charts")),
		wire.Value(chart.DefaultChart("reference-app-rolling")),
		wire.Value(util.ChartWorkingDir("/tmp/charts/")),
		session.SettingsManager,
		session.CDSettingsManager,
		//auth.GetConfig,

		argocdServer.GetConfig,
		wire.Bind(new(session2.ServiceClient), new(*middleware.LoginService)),

		sse.NewSSE,
		router.NewHelmRouter,
		wire.Bind(new(router.HelmRouter), new(*router.HelmRouterImpl)),

		//---- pprof start ----
		restHandler.NewPProfRestHandler,
		wire.Bind(new(restHandler.PProfRestHandler), new(*restHandler.PProfRestHandlerImpl)),

		router.NewPProfRouter,
		wire.Bind(new(router.PProfRouter), new(*router.PProfRouterImpl)),
		//---- pprof end ----

		restHandler.NewPipelineRestHandler,
		wire.Bind(new(restHandler.PipelineTriggerRestHandler), new(*restHandler.PipelineTriggerRestHandlerImpl)),

		app.NewAppService,
		wire.Bind(new(app.AppService), new(*app.AppServiceImpl)),

		bulkUpdate.NewBulkUpdateRepository,
		wire.Bind(new(bulkUpdate.BulkUpdateRepository), new(*bulkUpdate.BulkUpdateRepositoryImpl)),

		chartConfig.NewEnvConfigOverrideRepository,
		wire.Bind(new(chartConfig.EnvConfigOverrideRepository), new(*chartConfig.EnvConfigOverrideRepositoryImpl)),
		chartConfig.NewPipelineOverrideRepository,
		wire.Bind(new(chartConfig.PipelineOverrideRepository), new(*chartConfig.PipelineOverrideRepositoryImpl)),
		util.MergeUtil{},
		util.NewSugardLogger,

		deployment.NewDeploymentConfigRestHandlerImpl,
		wire.Bind(new(deployment.DeploymentConfigRestHandler), new(*deployment.DeploymentConfigRestHandlerImpl)),
		deployment.NewDeploymentRouterImpl,
		wire.Bind(new(deployment.DeploymentConfigRouter), new(*deployment.DeploymentConfigRouterImpl)),

		dashboardEvent.NewDashboardTelemetryRestHandlerImpl,
		wire.Bind(new(dashboardEvent.DashboardTelemetryRestHandler), new(*dashboardEvent.DashboardTelemetryRestHandlerImpl)),
		dashboardEvent.NewDashboardTelemetryRouterImpl,
		wire.Bind(new(dashboardEvent.DashboardTelemetryRouter),
			new(*dashboardEvent.DashboardTelemetryRouterImpl)),

		router.NewMuxRouter,

		app2.NewAppRepositoryImpl,
		wire.Bind(new(app2.AppRepository), new(*app2.AppRepositoryImpl)),

		pipeline.NewPipelineBuilderImpl,
		wire.Bind(new(pipeline.PipelineBuilder), new(*pipeline.PipelineBuilderImpl)),
		pipeline2.NewPipelineRestHandlerImpl,
		wire.Bind(new(pipeline2.PipelineConfigRestHandler), new(*pipeline2.PipelineConfigRestHandlerImpl)),
		router.NewPipelineRouterImpl,
		wire.Bind(new(router.PipelineConfigRouter), new(*router.PipelineConfigRouterImpl)),
		pipeline.NewDbPipelineOrchestrator,
		wire.Bind(new(pipeline.DbPipelineOrchestrator), new(*pipeline.DbPipelineOrchestratorImpl)),
		pipelineConfig.NewMaterialRepositoryImpl,
		wire.Bind(new(pipelineConfig.MaterialRepository), new(*pipelineConfig.MaterialRepositoryImpl)),

		router.NewMigrateDbRouterImpl,
		wire.Bind(new(router.MigrateDbRouter), new(*router.MigrateDbRouterImpl)),
		restHandler.NewMigrateDbRestHandlerImpl,
		wire.Bind(new(restHandler.MigrateDbRestHandler), new(*restHandler.MigrateDbRestHandlerImpl)),
		pipeline.NewDockerRegistryConfigImpl,
		wire.Bind(new(pipeline.DockerRegistryConfig), new(*pipeline.DockerRegistryConfigImpl)),
		repository.NewDockerArtifactStoreRepositoryImpl,
		wire.Bind(new(repository.DockerArtifactStoreRepository), new(*repository.DockerArtifactStoreRepositoryImpl)),
		util.NewChartTemplateServiceImpl,
		wire.Bind(new(util.ChartTemplateService), new(*util.ChartTemplateServiceImpl)),
		chart.NewChartServiceImpl,
		wire.Bind(new(chart.ChartService), new(*chart.ChartServiceImpl)),
		pipeline.NewBulkUpdateServiceImpl,
		wire.Bind(new(pipeline.BulkUpdateService), new(*pipeline.BulkUpdateServiceImpl)),

		repository.NewGitProviderRepositoryImpl,
		wire.Bind(new(repository.GitProviderRepository), new(*repository.GitProviderRepositoryImpl)),
		pipeline.NewGitRegistryConfigImpl,
		wire.Bind(new(pipeline.GitRegistryConfig), new(*pipeline.GitRegistryConfigImpl)),

		router.NewAppListingRouterImpl,
		wire.Bind(new(router.AppListingRouter), new(*router.AppListingRouterImpl)),
		restHandler.NewAppListingRestHandlerImpl,
		wire.Bind(new(restHandler.AppListingRestHandler), new(*restHandler.AppListingRestHandlerImpl)),
		app.NewAppListingServiceImpl,
		wire.Bind(new(app.AppListingService), new(*app.AppListingServiceImpl)),
		repository.NewAppListingRepositoryImpl,
		wire.Bind(new(repository.AppListingRepository), new(*repository.AppListingRepositoryImpl)),

		pipelineConfig.NewPipelineRepositoryImpl,
		wire.Bind(new(pipelineConfig.PipelineRepository), new(*pipelineConfig.PipelineRepositoryImpl)),
		pipeline.NewPropertiesConfigServiceImpl,
		wire.Bind(new(pipeline.PropertiesConfigService), new(*pipeline.PropertiesConfigServiceImpl)),

		router.NewProjectManagementRouterImpl,
		wire.Bind(new(router.ProjectManagementRouter), new(*router.ProjectManagementRouterImpl)),

		restHandler.NewJiraRestHandlerImpl,
		wire.Bind(new(restHandler.JiraRestHandler), new(*restHandler.JiraRestHandlerImpl)),

		jira2.NewProjectManagementServiceImpl,
		wire.Bind(new(jira2.ProjectManagementService), new(*jira2.ProjectManagementServiceImpl)),

		jira.NewAccountServiceImpl,
		wire.Bind(new(jira.AccountService), new(*jira.AccountServiceImpl)),

		util.NewHttpClient,

		jClient.NewJiraClientImpl,
		wire.Bind(new(jClient.JiraClient), new(*jClient.JiraClientImpl)),

		eClient.NewEventRESTClientImpl,
		wire.Bind(new(eClient.EventClient), new(*eClient.EventRESTClientImpl)),

		util3.NewTokenCache,

		eClient.NewEventSimpleFactoryImpl,
		wire.Bind(new(eClient.EventFactory), new(*eClient.EventSimpleFactoryImpl)),

		repository.NewJiraAccountRepositoryImpl,
		wire.Bind(new(repository.JiraAccountRepository), new(*repository.JiraAccountRepositoryImpl)),
		jira.NewAccountValidatorImpl,
		wire.Bind(new(jira.AccountValidator), new(*jira.AccountValidatorImpl)),

		repository.NewCiArtifactRepositoryImpl,
		wire.Bind(new(repository.CiArtifactRepository), new(*repository.CiArtifactRepositoryImpl)),
		pipeline.NewWebhookServiceImpl,
		wire.Bind(new(pipeline.WebhookService), new(*pipeline.WebhookServiceImpl)),

		router.NewWebhookRouterImpl,
		wire.Bind(new(router.WebhookRouter), new(*router.WebhookRouterImpl)),
		pipelineConfig.NewCiTemplateRepositoryImpl,
		wire.Bind(new(pipelineConfig.CiTemplateRepository), new(*pipelineConfig.CiTemplateRepositoryImpl)),
		pipelineConfig.NewCiPipelineRepositoryImpl,
		wire.Bind(new(pipelineConfig.CiPipelineRepository), new(*pipelineConfig.CiPipelineRepositoryImpl)),
		pipelineConfig.NewCiPipelineMaterialRepositoryImpl,
		wire.Bind(new(pipelineConfig.CiPipelineMaterialRepository), new(*pipelineConfig.CiPipelineMaterialRepositoryImpl)),
		util.NewGitFactory,

		application.NewApplicationClientImpl,
		wire.Bind(new(application.ServiceClient), new(*application.ServiceClientImpl)),
		cluster2.NewServiceClientImpl,
		wire.Bind(new(cluster2.ServiceClient), new(*cluster2.ServiceClientImpl)),
		connector.NewPumpImpl,
		repository2.NewServiceClientImpl,
		wire.Bind(new(repository2.ServiceClient), new(*repository2.ServiceClientImpl)),
		wire.Bind(new(connector.Pump), new(*connector.PumpImpl)),
		restHandler.NewArgoApplicationRestHandlerImpl,
		wire.Bind(new(restHandler.ArgoApplicationRestHandler), new(*restHandler.ArgoApplicationRestHandlerImpl)),
		router.NewApplicationRouterImpl,
		wire.Bind(new(router.ApplicationRouter), new(*router.ApplicationRouterImpl)),
		//app.GetConfig,

		router.NewCDRouterImpl,
		wire.Bind(new(router.CDRouter), new(*router.CDRouterImpl)),
		restHandler.NewCDRestHandlerImpl,
		wire.Bind(new(restHandler.CDRestHandler), new(*restHandler.CDRestHandlerImpl)),

		ArgoUtil.GetArgoConfig,
		ArgoUtil.NewArgoSession,
		ArgoUtil.NewResourceServiceImpl,
		wire.Bind(new(ArgoUtil.ResourceService), new(*ArgoUtil.ResourceServiceImpl)),
		//ArgoUtil.NewApplicationServiceImpl,
		//wire.Bind(new(ArgoUtil.ApplicationService), new(ArgoUtil.ApplicationServiceImpl)),
		//ArgoUtil.NewRepositoryService,
		//wire.Bind(new(ArgoUtil.RepositoryService), new(ArgoUtil.RepositoryServiceImpl)),

		pipelineConfig.NewDbMigrationConfigRepositoryImpl,
		wire.Bind(new(pipelineConfig.DbMigrationConfigRepository), new(*pipelineConfig.DbMigrationConfigRepositoryImpl)),
		pipeline.NewDbConfigService,
		wire.Bind(new(pipeline.DbConfigService), new(*pipeline.DbConfigServiceImpl)),

		repository.NewDbConfigRepositoryImpl,
		wire.Bind(new(repository.DbConfigRepository), new(*repository.DbConfigRepositoryImpl)),
		pipeline.NewDbMogrationService,
		wire.Bind(new(pipeline.DbMigrationService), new(*pipeline.DbMigrationServiceImpl)),
		//ArgoUtil.NewClusterServiceImpl,
		//wire.Bind(new(ArgoUtil.ClusterService), new(ArgoUtil.ClusterServiceImpl)),
		pipeline.GetEcrConfig,
		NewApp,
		//session.NewK8sClient,

		util.NewK8sUtil,
		argocdServer.NewVersionServiceImpl,
		wire.Bind(new(argocdServer.VersionService), new(*argocdServer.VersionServiceImpl)),

		router.NewGitProviderRouterImpl,
		wire.Bind(new(router.GitProviderRouter), new(*router.GitProviderRouterImpl)),
		restHandler.NewGitProviderRestHandlerImpl,
		wire.Bind(new(restHandler.GitProviderRestHandler), new(*restHandler.GitProviderRestHandlerImpl)),
		router.NewDockerRegRouterImpl,
		wire.Bind(new(router.DockerRegRouter), new(*router.DockerRegRouterImpl)),
		restHandler.NewDockerRegRestHandlerImpl,
		wire.Bind(new(restHandler.DockerRegRestHandler), new(*restHandler.DockerRegRestHandlerImpl)),

		router.NewNotificationRouterImpl,
		wire.Bind(new(router.NotificationRouter), new(*router.NotificationRouterImpl)),
		restHandler.NewNotificationRestHandlerImpl,
		wire.Bind(new(restHandler.NotificationRestHandler), new(*restHandler.NotificationRestHandlerImpl)),

		notifier.NewSlackNotificationServiceImpl,
		wire.Bind(new(notifier.SlackNotificationService), new(*notifier.SlackNotificationServiceImpl)),
		repository.NewSlackNotificationRepositoryImpl,
		wire.Bind(new(repository.SlackNotificationRepository), new(*repository.SlackNotificationRepositoryImpl)),

		notifier.NewNotificationConfigServiceImpl,
		wire.Bind(new(notifier.NotificationConfigService), new(*notifier.NotificationConfigServiceImpl)),
		app.NewAppListingViewBuilderImpl,
		wire.Bind(new(app.AppListingViewBuilder), new(*app.AppListingViewBuilderImpl)),
		repository.NewNotificationSettingsRepositoryImpl,
		wire.Bind(new(repository.NotificationSettingsRepository), new(*repository.NotificationSettingsRepositoryImpl)),
		util.IntValidator,
		pipeline.GetCiConfig,

		pipeline.NewWorkflowServiceImpl,
		wire.Bind(new(pipeline.WorkflowService), new(*pipeline.WorkflowServiceImpl)),

		pipeline.NewCiServiceImpl,
		wire.Bind(new(pipeline.CiService), new(*pipeline.CiServiceImpl)),

		pipelineConfig.NewCiWorkflowRepositoryImpl,
		wire.Bind(new(pipelineConfig.CiWorkflowRepository), new(*pipelineConfig.CiWorkflowRepositoryImpl)),

		restHandler.NewGitWebhookRestHandlerImpl,
		wire.Bind(new(restHandler.GitWebhookRestHandler), new(*restHandler.GitWebhookRestHandlerImpl)),

		git.NewGitWebhookServiceImpl,
		wire.Bind(new(git.GitWebhookService), new(*git.GitWebhookServiceImpl)),

		repository.NewGitWebhookRepositoryImpl,
		wire.Bind(new(repository.GitWebhookRepository), new(*repository.GitWebhookRepositoryImpl)),

		pipeline.NewCiHandlerImpl,
		wire.Bind(new(pipeline.CiHandler), new(*pipeline.CiHandlerImpl)),

		pipeline.NewCiLogServiceImpl,
		wire.Bind(new(pipeline.CiLogService), new(*pipeline.CiLogServiceImpl)),

		pubsub2.NewPubSubClient,

		pubsub.NewGitWebhookHandler,
		wire.Bind(new(pubsub.GitWebhookHandler), new(*pubsub.GitWebhookHandlerImpl)),

		pubsub.NewWorkflowStatusUpdateHandlerImpl,
		wire.Bind(new(pubsub.WorkflowStatusUpdateHandler), new(*pubsub.WorkflowStatusUpdateHandlerImpl)),

		pubsub.NewApplicationStatusUpdateHandlerImpl,
		wire.Bind(new(pubsub.ApplicationStatusUpdateHandler), new(*pubsub.ApplicationStatusUpdateHandlerImpl)),

		pubsub.NewCiEventHandlerImpl,
		wire.Bind(new(pubsub.CiEventHandler), new(*pubsub.CiEventHandlerImpl)),

		rbac.NewEnforcerUtilImpl,
		wire.Bind(new(rbac.EnforcerUtil), new(*rbac.EnforcerUtilImpl)),

		app.NewDeploymentFailureHandlerImpl,
		wire.Bind(new(app.DeploymentFailureHandler), new(*app.DeploymentFailureHandlerImpl)),
		chartConfig.NewPipelineConfigRepository,
		wire.Bind(new(chartConfig.PipelineConfigRepository), new(*chartConfig.PipelineConfigRepositoryImpl)),

		repository.NewLinkoutsRepositoryImpl,
		wire.Bind(new(repository.LinkoutsRepository), new(*repository.LinkoutsRepositoryImpl)),

		router.NewChartRefRouterImpl,
		wire.Bind(new(router.ChartRefRouter), new(*router.ChartRefRouterImpl)),
		restHandler.NewChartRefRestHandlerImpl,
		wire.Bind(new(restHandler.ChartRefRestHandler), new(*restHandler.ChartRefRestHandlerImpl)),

		router.NewConfigMapRouterImpl,
		wire.Bind(new(router.ConfigMapRouter), new(*router.ConfigMapRouterImpl)),
		restHandler.NewConfigMapRestHandlerImpl,
		wire.Bind(new(restHandler.ConfigMapRestHandler), new(*restHandler.ConfigMapRestHandlerImpl)),
		pipeline.NewConfigMapServiceImpl,
		wire.Bind(new(pipeline.ConfigMapService), new(*pipeline.ConfigMapServiceImpl)),
		chartConfig.NewConfigMapRepositoryImpl,
		wire.Bind(new(chartConfig.ConfigMapRepository), new(*chartConfig.ConfigMapRepositoryImpl)),

		notifier.NewSESNotificationServiceImpl,
		wire.Bind(new(notifier.SESNotificationService), new(*notifier.SESNotificationServiceImpl)),

		repository.NewSESNotificationRepositoryImpl,
		wire.Bind(new(repository.SESNotificationRepository), new(*repository.SESNotificationRepositoryImpl)),

		notifier.NewSMTPNotificationServiceImpl,
		wire.Bind(new(notifier.SMTPNotificationService), new(*notifier.SMTPNotificationServiceImpl)),

		repository.NewSMTPNotificationRepositoryImpl,
		wire.Bind(new(repository.SMTPNotificationRepository), new(*repository.SMTPNotificationRepositoryImpl)),

		notifier.NewNotificationConfigBuilderImpl,
		wire.Bind(new(notifier.NotificationConfigBuilder), new(*notifier.NotificationConfigBuilderImpl)),

		appStoreRestHandler.NewInstalledAppRestHandlerImpl,
		wire.Bind(new(appStoreRestHandler.InstalledAppRestHandler), new(*appStoreRestHandler.InstalledAppRestHandlerImpl)),
		service.NewInstalledAppServiceImpl,
		wire.Bind(new(service.InstalledAppService), new(*service.InstalledAppServiceImpl)),

		appStoreRestHandler.NewAppStoreRouterImpl,
		wire.Bind(new(appStoreRestHandler.AppStoreRouter), new(*appStoreRestHandler.AppStoreRouterImpl)),

		restHandler.NewAppWorkflowRestHandlerImpl,
		wire.Bind(new(restHandler.AppWorkflowRestHandler), new(*restHandler.AppWorkflowRestHandlerImpl)),

		appWorkflow.NewAppWorkflowServiceImpl,
		wire.Bind(new(appWorkflow.AppWorkflowService), new(*appWorkflow.AppWorkflowServiceImpl)),

		appWorkflow2.NewAppWorkflowRepositoryImpl,
		wire.Bind(new(appWorkflow2.AppWorkflowRepository), new(*appWorkflow2.AppWorkflowRepositoryImpl)),

		restHandler.NewExternalCiRestHandlerImpl,
		wire.Bind(new(restHandler.ExternalCiRestHandler), new(*restHandler.ExternalCiRestHandlerImpl)),
		repository.NewAppLevelMetricsRepositoryImpl,
		wire.Bind(new(repository.AppLevelMetricsRepository), new(*repository.AppLevelMetricsRepositoryImpl)),

		repository.NewEnvLevelAppMetricsRepositoryImpl,
		wire.Bind(new(repository.EnvLevelAppMetricsRepository), new(*repository.EnvLevelAppMetricsRepositoryImpl)),

		grafana.GetGrafanaClientConfig,
		grafana.NewGrafanaClientImpl,
		wire.Bind(new(grafana.GrafanaClient), new(*grafana.GrafanaClientImpl)),

		app.NewReleaseDataServiceImpl,
		wire.Bind(new(app.ReleaseDataService), new(*app.ReleaseDataServiceImpl)),
		restHandler.NewReleaseMetricsRestHandlerImpl,
		wire.Bind(new(restHandler.ReleaseMetricsRestHandler), new(*restHandler.ReleaseMetricsRestHandlerImpl)),
		router.NewReleaseMetricsRouterImpl,
		wire.Bind(new(router.ReleaseMetricsRouter), new(*router.ReleaseMetricsRouterImpl)),
		lens.GetLensConfig,
		lens.NewLensClientImpl,
		wire.Bind(new(lens.LensClient), new(*lens.LensClientImpl)),

		pipelineConfig.NewCdWorkflowRepositoryImpl,
		wire.Bind(new(pipelineConfig.CdWorkflowRepository), new(*pipelineConfig.CdWorkflowRepositoryImpl)),

		pipeline.NewCdWorkflowServiceImpl,
		wire.Bind(new(pipeline.CdWorkflowService), new(*pipeline.CdWorkflowServiceImpl)),

		pipeline.NewCdHandlerImpl,
		wire.Bind(new(pipeline.CdHandler), new(*pipeline.CdHandlerImpl)),

		pipeline.NewWorkflowDagExecutorImpl,
		wire.Bind(new(pipeline.WorkflowDagExecutor), new(*pipeline.WorkflowDagExecutorImpl)),
		appClone.NewAppCloneServiceImpl,
		wire.Bind(new(appClone.AppCloneService), new(*appClone.AppCloneServiceImpl)),
		pipeline.GetCdConfig,

		router.NewDeploymentGroupRouterImpl,
		wire.Bind(new(router.DeploymentGroupRouter), new(*router.DeploymentGroupRouterImpl)),
		restHandler.NewDeploymentGroupRestHandlerImpl,
		wire.Bind(new(restHandler.DeploymentGroupRestHandler), new(*restHandler.DeploymentGroupRestHandlerImpl)),
		deploymentGroup.NewDeploymentGroupServiceImpl,
		wire.Bind(new(deploymentGroup.DeploymentGroupService), new(*deploymentGroup.DeploymentGroupServiceImpl)),
		repository.NewDeploymentGroupRepositoryImpl,
		wire.Bind(new(repository.DeploymentGroupRepository), new(*repository.DeploymentGroupRepositoryImpl)),

		repository.NewDeploymentGroupAppRepositoryImpl,
		wire.Bind(new(repository.DeploymentGroupAppRepository), new(*repository.DeploymentGroupAppRepositoryImpl)),
		restHandler.NewPubSubClientRestHandlerImpl,
		wire.Bind(new(restHandler.PubSubClientRestHandler), new(*restHandler.PubSubClientRestHandlerImpl)),
		pubsub2.NewNatsPublishClientImpl,
		wire.Bind(new(pubsub2.NatsPublishClient), new(*pubsub2.NatsPublishClientImpl)),

		//Batch actions
		batch.NewWorkflowActionImpl,
		wire.Bind(new(batch.WorkflowAction), new(*batch.WorkflowActionImpl)),
		batch.NewDeploymentActionImpl,
		wire.Bind(new(batch.DeploymentAction), new(*batch.DeploymentActionImpl)),
		batch.NewBuildActionImpl,
		wire.Bind(new(batch.BuildAction), new(*batch.BuildActionImpl)),
		batch.NewDataHolderActionImpl,
		wire.Bind(new(batch.DataHolderAction), new(*batch.DataHolderActionImpl)),
		batch.NewDeploymentTemplateActionImpl,
		wire.Bind(new(batch.DeploymentTemplateAction), new(*batch.DeploymentTemplateActionImpl)),
		restHandler.NewBatchOperationRestHandlerImpl,
		wire.Bind(new(restHandler.BatchOperationRestHandler), new(*restHandler.BatchOperationRestHandlerImpl)),
		router.NewBatchOperationRouterImpl,
		wire.Bind(new(router.BatchOperationRouter), new(*router.BatchOperationRouterImpl)),

		repository4.NewChartGroupReposotoryImpl,
		wire.Bind(new(repository4.ChartGroupReposotory), new(*repository4.ChartGroupReposotoryImpl)),
		repository4.NewChartGroupEntriesRepositoryImpl,
		wire.Bind(new(repository4.ChartGroupEntriesRepository), new(*repository4.ChartGroupEntriesRepositoryImpl)),
		service.NewChartGroupServiceImpl,
		wire.Bind(new(service.ChartGroupService), new(*service.ChartGroupServiceImpl)),
		restHandler.NewChartGroupRestHandlerImpl,
		wire.Bind(new(restHandler.ChartGroupRestHandler), new(*restHandler.ChartGroupRestHandlerImpl)),
		router.NewChartGroupRouterImpl,
		wire.Bind(new(router.ChartGroupRouter), new(*router.ChartGroupRouterImpl)),
		repository4.NewChartGroupDeploymentRepositoryImpl,
		wire.Bind(new(repository4.ChartGroupDeploymentRepository), new(*repository4.ChartGroupDeploymentRepositoryImpl)),

		commonService.NewCommonServiceImpl,
		wire.Bind(new(commonService.CommonService), new(*commonService.CommonServiceImpl)),

		router.NewTestSuitRouterImpl,
		wire.Bind(new(router.TestSuitRouter), new(*router.TestSuitRouterImpl)),
		restHandler.NewTestSuitRestHandlerImpl,
		wire.Bind(new(restHandler.TestSuitRestHandler), new(*restHandler.TestSuitRestHandlerImpl)),

		router.NewImageScanRouterImpl,
		wire.Bind(new(router.ImageScanRouter), new(*router.ImageScanRouterImpl)),
		restHandler.NewImageScanRestHandlerImpl,
		wire.Bind(new(restHandler.ImageScanRestHandler), new(*restHandler.ImageScanRestHandlerImpl)),
		security.NewImageScanServiceImpl,
		wire.Bind(new(security.ImageScanService), new(*security.ImageScanServiceImpl)),
		security2.NewImageScanHistoryRepositoryImpl,
		wire.Bind(new(security2.ImageScanHistoryRepository), new(*security2.ImageScanHistoryRepositoryImpl)),
		security2.NewImageScanResultRepositoryImpl,
		wire.Bind(new(security2.ImageScanResultRepository), new(*security2.ImageScanResultRepositoryImpl)),
		security2.NewImageScanObjectMetaRepositoryImpl,
		wire.Bind(new(security2.ImageScanObjectMetaRepository), new(*security2.ImageScanObjectMetaRepositoryImpl)),
		security2.NewCveStoreRepositoryImpl,
		wire.Bind(new(security2.CveStoreRepository), new(*security2.CveStoreRepositoryImpl)),
		security2.NewImageScanDeployInfoRepositoryImpl,
		wire.Bind(new(security2.ImageScanDeployInfoRepository), new(*security2.ImageScanDeployInfoRepositoryImpl)),
		router.NewPolicyRouterImpl,
		wire.Bind(new(router.PolicyRouter), new(*router.PolicyRouterImpl)),
		restHandler.NewPolicyRestHandlerImpl,
		wire.Bind(new(restHandler.PolicyRestHandler), new(*restHandler.PolicyRestHandlerImpl)),
		security.NewPolicyServiceImpl,
		wire.Bind(new(security.PolicyService), new(*security.PolicyServiceImpl)),
		security2.NewPolicyRepositoryImpl,
		wire.Bind(new(security2.CvePolicyRepository), new(*security2.CvePolicyRepositoryImpl)),

		argocdServer.NewArgoK8sClientImpl,
		wire.Bind(new(argocdServer.ArgoK8sClient), new(*argocdServer.ArgoK8sClientImpl)),

		grafana.GetConfig,
		router.NewGrafanaRouterImpl,
		wire.Bind(new(router.GrafanaRouter), new(*router.GrafanaRouterImpl)),

		router.NewGitOpsConfigRouterImpl,
		wire.Bind(new(router.GitOpsConfigRouter), new(*router.GitOpsConfigRouterImpl)),
		restHandler.NewGitOpsConfigRestHandlerImpl,
		wire.Bind(new(restHandler.GitOpsConfigRestHandler), new(*restHandler.GitOpsConfigRestHandlerImpl)),
		gitops.NewGitOpsConfigServiceImpl,
		wire.Bind(new(gitops.GitOpsConfigService), new(*gitops.GitOpsConfigServiceImpl)),
		repository.NewGitOpsConfigRepositoryImpl,
		wire.Bind(new(repository.GitOpsConfigRepository), new(*repository.GitOpsConfigRepositoryImpl)),

		router.NewAttributesRouterImpl,
		wire.Bind(new(router.AttributesRouter), new(*router.AttributesRouterImpl)),
		restHandler.NewAttributesRestHandlerImpl,
		wire.Bind(new(restHandler.AttributesRestHandler), new(*restHandler.AttributesRestHandlerImpl)),
		attributes.NewAttributesServiceImpl,
		wire.Bind(new(attributes.AttributesService), new(*attributes.AttributesServiceImpl)),
		repository.NewAttributesRepositoryImpl,
		wire.Bind(new(repository.AttributesRepository), new(*repository.AttributesRepositoryImpl)),

		router.NewCommonRouterImpl,
		wire.Bind(new(router.CommonRouter), new(*router.CommonRouterImpl)),
		restHandler.NewCommonRestHanlderImpl,
		wire.Bind(new(restHandler.CommonRestHanlder), new(*restHandler.CommonRestHanlderImpl)),

		util.NewGitCliUtil,

		router.NewTelemetryRouterImpl,
		wire.Bind(new(router.TelemetryRouter), new(*router.TelemetryRouterImpl)),
		restHandler.NewTelemetryRestHandlerImpl,
		wire.Bind(new(restHandler.TelemetryRestHandler), new(*restHandler.TelemetryRestHandlerImpl)),
		telemetry.NewPosthogClient,

		telemetry.NewTelemetryEventClientImplExtended,
		wire.Bind(new(telemetry.TelemetryEventClient), new(*telemetry.TelemetryEventClientImplExtended)),

		router.NewBulkUpdateRouterImpl,
		wire.Bind(new(router.BulkUpdateRouter), new(*router.BulkUpdateRouterImpl)),
		restHandler.NewBulkUpdateRestHandlerImpl,
		wire.Bind(new(restHandler.BulkUpdateRestHandler), new(*restHandler.BulkUpdateRestHandlerImpl)),

		router.NewCoreAppRouterImpl,
		wire.Bind(new(router.CoreAppRouter), new(*router.CoreAppRouterImpl)),
		restHandler.NewCoreAppRestHandlerImpl,
		wire.Bind(new(restHandler.CoreAppRestHandler), new(*restHandler.CoreAppRestHandlerImpl)),

		// Webhook
		repository.NewGitHostRepositoryImpl,
		wire.Bind(new(repository.GitHostRepository), new(*repository.GitHostRepositoryImpl)),
		restHandler.NewGitHostRestHandlerImpl,
		wire.Bind(new(restHandler.GitHostRestHandler), new(*restHandler.GitHostRestHandlerImpl)),
		restHandler.NewWebhookEventHandlerImpl,
		wire.Bind(new(restHandler.WebhookEventHandler), new(*restHandler.WebhookEventHandlerImpl)),
		router.NewGitHostRouterImpl,
		wire.Bind(new(router.GitHostRouter), new(*router.GitHostRouterImpl)),
		router.NewWebhookListenerRouterImpl,
		wire.Bind(new(router.WebhookListenerRouter), new(*router.WebhookListenerRouterImpl)),
		git.NewWebhookSecretValidatorImpl,
		wire.Bind(new(git.WebhookSecretValidator), new(*git.WebhookSecretValidatorImpl)),
		pipeline.NewGitHostConfigImpl,
		wire.Bind(new(pipeline.GitHostConfig), new(*pipeline.GitHostConfigImpl)),
		repository.NewWebhookEventDataRepositoryImpl,
		wire.Bind(new(repository.WebhookEventDataRepository), new(*repository.WebhookEventDataRepositoryImpl)),
		pipeline.NewWebhookEventDataConfigImpl,
		wire.Bind(new(pipeline.WebhookEventDataConfig), new(*pipeline.WebhookEventDataConfigImpl)),
		restHandler.NewWebhookDataRestHandlerImpl,
		wire.Bind(new(restHandler.WebhookDataRestHandler), new(*restHandler.WebhookDataRestHandlerImpl)),

		router.NewAppLabelRouterImpl,
		wire.Bind(new(router.AppLabelRouter), new(*router.AppLabelRouterImpl)),
		restHandler.NewAppLabelRestHandlerImpl,
		wire.Bind(new(restHandler.AppLabelRestHandler), new(*restHandler.AppLabelRestHandlerImpl)),

		app.NewAppLabelServiceImpl,
		wire.Bind(new(app.AppLabelService), new(*app.AppLabelServiceImpl)),
		pipelineConfig.NewAppLabelRepositoryImpl,
		wire.Bind(new(pipelineConfig.AppLabelRepository), new(*pipelineConfig.AppLabelRepositoryImpl)),

		delete2.NewDeleteServiceExtendedImpl,
		wire.Bind(new(delete2.DeleteService), new(*delete2.DeleteServiceExtendedImpl)),
		delete2.NewDeleteServiceFullModeImpl,
		wire.Bind(new(delete2.DeleteServiceFullMode), new(*delete2.DeleteServiceFullModeImpl)),

		appStoreDeploymentFullMode.NewAppStoreDeploymentFullModeServiceImpl,
		wire.Bind(new(appStoreDeploymentFullMode.AppStoreDeploymentFullModeService), new(*appStoreDeploymentFullMode.AppStoreDeploymentFullModeServiceImpl)),
		appStoreDeploymentGitopsTool.NewAppStoreDeploymentArgoCdServiceImpl,
		wire.Bind(new(appStoreDeploymentGitopsTool.AppStoreDeploymentArgoCdService), new(*appStoreDeploymentGitopsTool.AppStoreDeploymentArgoCdServiceImpl)),
		//	util2.NewGoJsonSchemaCustomFormatChecker,

		//history starts
		restHandler.NewPipelineHistoryRestHandlerImpl,
		wire.Bind(new(restHandler.PipelineHistoryRestHandler), new(*restHandler.PipelineHistoryRestHandlerImpl)),

		repository3.NewConfigMapHistoryRepositoryImpl,
		wire.Bind(new(repository3.ConfigMapHistoryRepository), new(*repository3.ConfigMapHistoryRepositoryImpl)),
		repository3.NewDeploymentTemplateHistoryRepositoryImpl,
		wire.Bind(new(repository3.DeploymentTemplateHistoryRepository), new(*repository3.DeploymentTemplateHistoryRepositoryImpl)),
		repository3.NewPrePostCiScriptHistoryRepositoryImpl,
		wire.Bind(new(repository3.PrePostCiScriptHistoryRepository), new(*repository3.PrePostCiScriptHistoryRepositoryImpl)),
		repository3.NewPrePostCdScriptHistoryRepositoryImpl,
		wire.Bind(new(repository3.PrePostCdScriptHistoryRepository), new(*repository3.PrePostCdScriptHistoryRepositoryImpl)),
		repository3.NewPipelineStrategyHistoryRepositoryImpl,
		wire.Bind(new(repository3.PipelineStrategyHistoryRepository), new(*repository3.PipelineStrategyHistoryRepositoryImpl)),

		history3.NewPrePostCdScriptHistoryServiceImpl,
		wire.Bind(new(history3.PrePostCdScriptHistoryService), new(*history3.PrePostCdScriptHistoryServiceImpl)),
		history3.NewPrePostCiScriptHistoryServiceImpl,
		wire.Bind(new(history3.PrePostCiScriptHistoryService), new(*history3.PrePostCiScriptHistoryServiceImpl)),
		history3.NewDeploymentTemplateHistoryServiceImpl,
		wire.Bind(new(history3.DeploymentTemplateHistoryService), new(*history3.DeploymentTemplateHistoryServiceImpl)),
		history3.NewConfigMapHistoryServiceImpl,
		wire.Bind(new(history3.ConfigMapHistoryService), new(*history3.ConfigMapHistoryServiceImpl)),
		history3.NewPipelineStrategyHistoryServiceImpl,
		wire.Bind(new(history3.PipelineStrategyHistoryService), new(*history3.PipelineStrategyHistoryServiceImpl)),

		history3.NewDeployedConfigurationHistoryServiceImpl,
		wire.Bind(new(history3.DeployedConfigurationHistoryService), new(*history3.DeployedConfigurationHistoryServiceImpl)),
		//history ends

		//plugin starts
		repository6.NewGlobalPluginRepository,
		wire.Bind(new(repository6.GlobalPluginRepository), new(*repository6.GlobalPluginRepositoryImpl)),

		plugin.NewGlobalPluginService,
		wire.Bind(new(plugin.GlobalPluginService), new(*plugin.GlobalPluginServiceImpl)),

		restHandler.NewGlobalPluginRestHandler,
		wire.Bind(new(restHandler.GlobalPluginRestHandler), new(*restHandler.GlobalPluginRestHandlerImpl)),

		router.NewGlobalPluginRouter,
		wire.Bind(new(router.GlobalPluginRouter), new(*router.GlobalPluginRouterImpl)),

		repository5.NewPipelineStageRepository,
		wire.Bind(new(repository5.PipelineStageRepository), new(*repository5.PipelineStageRepositoryImpl)),

		pipeline.NewPipelineStageService,
		wire.Bind(new(pipeline.PipelineStageService), new(*pipeline.PipelineStageServiceImpl)),
		//plugin ends

		argo.NewArgoUserServiceImpl,
		wire.Bind(new(argo.ArgoUserService), new(*argo.ArgoUserServiceImpl)),
		argo.GetDevtronSecretName,
		//	AuthWireSet,
		cron.GetAppStatusConfig,
		cron.NewCdApplicationStatusUpdateHandlerImpl,
		wire.Bind(new(cron.CdApplicationStatusUpdateHandler), new(*cron.CdApplicationStatusUpdateHandlerImpl)),

		restHandler.NewPipelineStatusTimelineRestHandlerImpl,
		wire.Bind(new(restHandler.PipelineStatusTimelineRestHandler), new(*restHandler.PipelineStatusTimelineRestHandlerImpl)),

		app.NewPipelineStatusTimelineServiceImpl,
		wire.Bind(new(app.PipelineStatusTimelineService), new(*app.PipelineStatusTimelineServiceImpl)),

		pipelineConfig.NewPipelineStatusTimelineRepositoryImpl,
		wire.Bind(new(pipelineConfig.PipelineStatusTimelineRepository), new(*pipelineConfig.PipelineStatusTimelineRepositoryImpl)),
	)
	return &App{}, nil
}
