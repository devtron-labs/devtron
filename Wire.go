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
	pubsub1 "github.com/devtron-labs/common-lib/pubsub-lib"
	util4 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/apiToken"
	appStoreRestHandler "github.com/devtron-labs/devtron/api/appStore"
	chartProvider "github.com/devtron-labs/devtron/api/appStore/chartProvider"
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
	"github.com/devtron-labs/devtron/api/k8s"
	"github.com/devtron-labs/devtron/api/module"
	"github.com/devtron-labs/devtron/api/restHandler"
	pipeline2 "github.com/devtron-labs/devtron/api/restHandler/app"
	"github.com/devtron-labs/devtron/api/restHandler/scopedVariable"
	"github.com/devtron-labs/devtron/api/router"
	"github.com/devtron-labs/devtron/api/router/pubsub"
	"github.com/devtron-labs/devtron/api/server"
	"github.com/devtron-labs/devtron/api/sse"
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/terminal"
	"github.com/devtron-labs/devtron/api/user"
	util5 "github.com/devtron-labs/devtron/api/util"
	webhookHelm "github.com/devtron-labs/devtron/api/webhook/helm"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	cluster2 "github.com/devtron-labs/devtron/client/argocdServer/cluster"
	"github.com/devtron-labs/devtron/client/argocdServer/connection"
	repository2 "github.com/devtron-labs/devtron/client/argocdServer/repository"
	session2 "github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/client/cron"
	"github.com/devtron-labs/devtron/client/dashboard"
	eClient "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/client/grafana"
	jClient "github.com/devtron-labs/devtron/client/jira"
	"github.com/devtron-labs/devtron/client/lens"
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	appStatusRepo "github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	appWorkflow2 "github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	"github.com/devtron-labs/devtron/internal/sql/repository/bulkUpdate"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	repository8 "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	resourceGroup "github.com/devtron-labs/devtron/internal/sql/repository/resourceGroup"
	security2 "github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/internal/util/ArgoUtil"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/appClone"
	"github.com/devtron-labs/devtron/pkg/appClone/batch"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentFullMode "github.com/devtron-labs/devtron/pkg/appStore/deployment/fullMode"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	appStoreDeploymentGitopsTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool/gitops"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/bulkAction"
	"github.com/devtron-labs/devtron/pkg/chart"
	chartRepoRepository "github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/commonService"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	repository9 "github.com/devtron-labs/devtron/pkg/devtronResource/repository"
	"github.com/devtron-labs/devtron/pkg/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/generateManifest"
	"github.com/devtron-labs/devtron/pkg/git"
	"github.com/devtron-labs/devtron/pkg/gitops"
	jira2 "github.com/devtron-labs/devtron/pkg/jira"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	repository7 "github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs/repository"
	"github.com/devtron-labs/devtron/pkg/notifier"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/executors"
	history3 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	repository3 "github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	repository5 "github.com/devtron-labs/devtron/pkg/pipeline/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/devtron-labs/devtron/pkg/plugin"
	repository6 "github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/projectManagementService/jira"
	resourceGroup2 "github.com/devtron-labs/devtron/pkg/resourceGroup"
	"github.com/devtron-labs/devtron/pkg/resourceQualifiers"
	"github.com/devtron-labs/devtron/pkg/security"
	"github.com/devtron-labs/devtron/pkg/sql"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/parsers"
	repository10 "github.com/devtron-labs/devtron/pkg/variables/repository"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/rbac"
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
		util4.NewK8sUtil,
		user.UserWireSet,
		sso.SsoConfigWireSet,
		cluster.ClusterWireSet,
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
		// -------wireset end ----------
		//-------
		gitSensor.GetConfig,
		gitSensor.NewGitSensorClient,
		wire.Bind(new(gitSensor.Client), new(*gitSensor.ClientImpl)),
		//-------
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
		connection.SettingsManager,
		//auth.GetConfig,

		connection.GetConfig,
		wire.Bind(new(session2.ServiceClient), new(*middleware.LoginService)),

		sse.NewSSE,
		router.NewPipelineTriggerRouter,
		wire.Bind(new(router.PipelineTriggerRouter), new(*router.PipelineTriggerRouterImpl)),

		//---- pprof start ----
		restHandler.NewPProfRestHandler,
		wire.Bind(new(restHandler.PProfRestHandler), new(*restHandler.PProfRestHandlerImpl)),

		router.NewPProfRouter,
		wire.Bind(new(router.PProfRouter), new(*router.PProfRouterImpl)),
		//---- pprof end ----

		restHandler.NewPipelineRestHandler,
		wire.Bind(new(restHandler.PipelineTriggerRestHandler), new(*restHandler.PipelineTriggerRestHandlerImpl)),
		app.GetAppServiceConfig,
		app.NewAppService,
		wire.Bind(new(app.AppService), new(*app.AppServiceImpl)),

		bulkUpdate.NewBulkUpdateRepository,
		wire.Bind(new(bulkUpdate.BulkUpdateRepository), new(*bulkUpdate.BulkUpdateRepositoryImpl)),

		chartConfig.NewEnvConfigOverrideRepository,
		wire.Bind(new(chartConfig.EnvConfigOverrideRepository), new(*chartConfig.EnvConfigOverrideRepositoryImpl)),
		chartConfig.NewPipelineOverrideRepository,
		wire.Bind(new(chartConfig.PipelineOverrideRepository), new(*chartConfig.PipelineOverrideRepositoryImpl)),
		wire.Struct(new(util.MergeUtil), "*"),
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

		pipeline.GetDeploymentServiceTypeConfig,

		pipeline.NewPipelineBuilderImpl,
		wire.Bind(new(pipeline.PipelineBuilder), new(*pipeline.PipelineBuilderImpl)),
		pipeline.NewCiPipelineConfigServiceImpl,
		wire.Bind(new(pipeline.CiPipelineConfigService), new(*pipeline.CiPipelineConfigServiceImpl)),
		pipeline.NewCiMaterialConfigServiceImpl,
		wire.Bind(new(pipeline.CiMaterialConfigService), new(*pipeline.CiMaterialConfigServiceImpl)),

		pipeline.NewAppArtifactManagerImpl,
		wire.Bind(new(pipeline.AppArtifactManager), new(*pipeline.AppArtifactManagerImpl)),
		pipeline.NewDevtronAppCMCSServiceImpl,
		wire.Bind(new(pipeline.DevtronAppCMCSService), new(*pipeline.DevtronAppCMCSServiceImpl)),
		pipeline.NewDevtronAppStrategyServiceImpl,
		wire.Bind(new(pipeline.DevtronAppStrategyService), new(*pipeline.DevtronAppStrategyServiceImpl)),
		pipeline.NewAppDeploymentTypeChangeManagerImpl,
		wire.Bind(new(pipeline.AppDeploymentTypeChangeManager), new(*pipeline.AppDeploymentTypeChangeManagerImpl)),
		pipeline.NewCdPipelineConfigServiceImpl,
		wire.Bind(new(pipeline.CdPipelineConfigService), new(*pipeline.CdPipelineConfigServiceImpl)),
		pipeline.NewDevtronAppConfigServiceImpl,
		wire.Bind(new(pipeline.DevtronAppConfigService), new(*pipeline.DevtronAppConfigServiceImpl)),

		util5.NewLoggingMiddlewareImpl,
		wire.Bind(new(util5.LoggingMiddleware), new(*util5.LoggingMiddlewareImpl)),
		pipeline2.NewPipelineRestHandlerImpl,
		wire.Bind(new(pipeline2.PipelineConfigRestHandler), new(*pipeline2.PipelineConfigRestHandlerImpl)),
		router.NewPipelineRouterImpl,
		wire.Bind(new(router.PipelineConfigRouter), new(*router.PipelineConfigRouterImpl)),
		pipeline.NewCiCdPipelineOrchestrator,
		wire.Bind(new(pipeline.CiCdPipelineOrchestrator), new(*pipeline.CiCdPipelineOrchestratorImpl)),
		pipelineConfig.NewMaterialRepositoryImpl,
		wire.Bind(new(pipelineConfig.MaterialRepository), new(*pipelineConfig.MaterialRepositoryImpl)),
		router.NewMigrateDbRouterImpl,
		wire.Bind(new(router.MigrateDbRouter), new(*router.MigrateDbRouterImpl)),
		restHandler.NewMigrateDbRestHandlerImpl,
		wire.Bind(new(restHandler.MigrateDbRestHandler), new(*restHandler.MigrateDbRestHandlerImpl)),
		util.NewChartTemplateServiceImpl,
		wire.Bind(new(util.ChartTemplateService), new(*util.ChartTemplateServiceImpl)),
		util.NewChartDeploymentServiceImpl,
		wire.Bind(new(util.ChartDeploymentService), new(*util.ChartDeploymentServiceImpl)),

		//scoped variables start
		variables.NewScopedVariableServiceImpl,
		wire.Bind(new(variables.ScopedVariableService), new(*variables.ScopedVariableServiceImpl)),

		parsers.NewVariableTemplateParserImpl,
		wire.Bind(new(parsers.VariableTemplateParser), new(*parsers.VariableTemplateParserImpl)),
		repository10.NewVariableEntityMappingRepository,
		wire.Bind(new(repository10.VariableEntityMappingRepository), new(*repository10.VariableEntityMappingRepositoryImpl)),

		repository10.NewVariableSnapshotHistoryRepository,
		wire.Bind(new(repository10.VariableSnapshotHistoryRepository), new(*repository10.VariableSnapshotHistoryRepositoryImpl)),
		variables.NewVariableEntityMappingServiceImpl,
		wire.Bind(new(variables.VariableEntityMappingService), new(*variables.VariableEntityMappingServiceImpl)),
		variables.NewVariableSnapshotHistoryServiceImpl,
		wire.Bind(new(variables.VariableSnapshotHistoryService), new(*variables.VariableSnapshotHistoryServiceImpl)),

		variables.NewScopedVariableManagerImpl,
		wire.Bind(new(variables.ScopedVariableManager), new(*variables.ScopedVariableManagerImpl)),

		variables.NewScopedVariableCMCSManagerImpl,
		wire.Bind(new(variables.ScopedVariableCMCSManager), new(*variables.ScopedVariableCMCSManagerImpl)),

		//end

		chart.NewChartServiceImpl,
		wire.Bind(new(chart.ChartService), new(*chart.ChartServiceImpl)),
		bulkAction.NewBulkUpdateServiceImpl,
		wire.Bind(new(bulkAction.BulkUpdateService), new(*bulkAction.BulkUpdateServiceImpl)),

		repository.NewImageTagRepository,
		wire.Bind(new(repository.ImageTagRepository), new(*repository.ImageTagRepositoryImpl)),

		pipeline.NewCustomTagService,
		wire.Bind(new(pipeline.CustomTagService), new(*pipeline.CustomTagServiceImpl)),

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

		repository.NewDeploymentTemplateRepositoryImpl,
		wire.Bind(new(repository.DeploymentTemplateRepository), new(*repository.DeploymentTemplateRepositoryImpl)),
		generateManifest.NewDeploymentTemplateServiceImpl,
		wire.Bind(new(generateManifest.DeploymentTemplateService), new(*generateManifest.DeploymentTemplateServiceImpl)),

		router.NewJobRouterImpl,
		wire.Bind(new(router.JobRouter), new(*router.JobRouterImpl)),

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
		//otel.NewOtelTracingServiceImpl,
		//wire.Bind(new(otel.OtelTracingService), new(*otel.OtelTracingServiceImpl)),
		NewApp,
		//session.NewK8sClient,
		repository8.NewImageTaggingRepositoryImpl,
		wire.Bind(new(repository8.ImageTaggingRepository), new(*repository8.ImageTaggingRepositoryImpl)),
		pipeline.NewImageTaggingServiceImpl,
		wire.Bind(new(pipeline.ImageTaggingService), new(*pipeline.ImageTaggingServiceImpl)),
		argocdServer.NewVersionServiceImpl,
		wire.Bind(new(argocdServer.VersionService), new(*argocdServer.VersionServiceImpl)),

		router.NewGitProviderRouterImpl,
		wire.Bind(new(router.GitProviderRouter), new(*router.GitProviderRouterImpl)),
		restHandler.NewGitProviderRestHandlerImpl,
		wire.Bind(new(restHandler.GitProviderRestHandler), new(*restHandler.GitProviderRestHandlerImpl)),

		router.NewNotificationRouterImpl,
		wire.Bind(new(router.NotificationRouter), new(*router.NotificationRouterImpl)),
		restHandler.NewNotificationRestHandlerImpl,
		wire.Bind(new(restHandler.NotificationRestHandler), new(*restHandler.NotificationRestHandlerImpl)),

		notifier.NewSlackNotificationServiceImpl,
		wire.Bind(new(notifier.SlackNotificationService), new(*notifier.SlackNotificationServiceImpl)),
		repository.NewSlackNotificationRepositoryImpl,
		wire.Bind(new(repository.SlackNotificationRepository), new(*repository.SlackNotificationRepositoryImpl)),
		notifier.NewWebhookNotificationServiceImpl,
		wire.Bind(new(notifier.WebhookNotificationService), new(*notifier.WebhookNotificationServiceImpl)),
		repository.NewWebhookNotificationRepositoryImpl,
		wire.Bind(new(repository.WebhookNotificationRepository), new(*repository.WebhookNotificationRepositoryImpl)),

		notifier.NewNotificationConfigServiceImpl,
		wire.Bind(new(notifier.NotificationConfigService), new(*notifier.NotificationConfigServiceImpl)),
		app.NewAppListingViewBuilderImpl,
		wire.Bind(new(app.AppListingViewBuilder), new(*app.AppListingViewBuilderImpl)),
		repository.NewNotificationSettingsRepositoryImpl,
		wire.Bind(new(repository.NotificationSettingsRepository), new(*repository.NotificationSettingsRepositoryImpl)),
		util.IntValidator,
		types.GetCiCdConfig,

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

		pubsub1.NewPubSubClientServiceImpl,

		pubsub.NewGitWebhookHandler,
		wire.Bind(new(pubsub.GitWebhookHandler), new(*pubsub.GitWebhookHandlerImpl)),

		pubsub.NewWorkflowStatusUpdateHandlerImpl,
		wire.Bind(new(pubsub.WorkflowStatusUpdateHandler), new(*pubsub.WorkflowStatusUpdateHandlerImpl)),

		pubsub.NewApplicationStatusHandlerImpl,
		wire.Bind(new(pubsub.ApplicationStatusHandler), new(*pubsub.ApplicationStatusHandlerImpl)),

		pubsub.GetCiEventConfig,
		pubsub.NewCiEventHandlerImpl,
		wire.Bind(new(pubsub.CiEventHandler), new(*pubsub.CiEventHandlerImpl)),

		rbac.NewEnforcerUtilImpl,
		wire.Bind(new(rbac.EnforcerUtil), new(*rbac.EnforcerUtilImpl)),

		app.NewDeploymentEventHandlerImpl,
		wire.Bind(new(app.DeploymentEventHandler), new(*app.DeploymentEventHandlerImpl)),
		chartConfig.NewPipelineConfigRepository,
		wire.Bind(new(chartConfig.PipelineConfigRepository), new(*chartConfig.PipelineConfigRepositoryImpl)),

		repository10.NewScopedVariableRepository,
		wire.Bind(new(repository10.ScopedVariableRepository), new(*repository10.ScopedVariableRepositoryImpl)),

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
		appStoreRestHandler.NewAppStoreStatusTimelineRestHandlerImpl,
		wire.Bind(new(appStoreRestHandler.AppStoreStatusTimelineRestHandler), new(*appStoreRestHandler.AppStoreStatusTimelineRestHandlerImpl)),
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

		pipeline.NewCdHandlerImpl,
		wire.Bind(new(pipeline.CdHandler), new(*pipeline.CdHandlerImpl)),

		pipeline.NewBlobStorageConfigServiceImpl,
		wire.Bind(new(pipeline.BlobStorageConfigService), new(*pipeline.BlobStorageConfigServiceImpl)),

		pipeline.NewWorkflowDagExecutorImpl,
		wire.Bind(new(pipeline.WorkflowDagExecutor), new(*pipeline.WorkflowDagExecutorImpl)),
		appClone.NewAppCloneServiceImpl,
		wire.Bind(new(appClone.AppCloneService), new(*appClone.AppCloneServiceImpl)),

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
		security2.NewScanToolMetadataRepositoryImpl,
		wire.Bind(new(security2.ScanToolMetadataRepository), new(*security2.ScanToolMetadataRepositoryImpl)),
		router.NewPolicyRouterImpl,
		wire.Bind(new(router.PolicyRouter), new(*router.PolicyRouterImpl)),
		restHandler.NewPolicyRestHandlerImpl,
		wire.Bind(new(restHandler.PolicyRestHandler), new(*restHandler.PolicyRestHandlerImpl)),
		security.NewPolicyServiceImpl,
		wire.Bind(new(security.PolicyService), new(*security.PolicyServiceImpl)),
		security2.NewPolicyRepositoryImpl,
		wire.Bind(new(security2.CvePolicyRepository), new(*security2.CvePolicyRepositoryImpl)),
		security2.NewScanToolExecutionHistoryMappingRepositoryImpl,
		wire.Bind(new(security2.ScanToolExecutionHistoryMappingRepository), new(*security2.ScanToolExecutionHistoryMappingRepositoryImpl)),

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
		restHandler.NewCommonRestHandlerImpl,
		wire.Bind(new(restHandler.CommonRestHandler), new(*restHandler.CommonRestHandlerImpl)),

		router.NewScopedVariableRouterImpl,
		wire.Bind(new(router.ScopedVariableRouter), new(*router.ScopedVariableRouterImpl)),
		scopedVariable.NewScopedVariableRestHandlerImpl,
		wire.Bind(new(scopedVariable.ScopedVariableRestHandler), new(*scopedVariable.ScopedVariableRestHandlerImpl)),

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

		router.NewAppRouterImpl,
		wire.Bind(new(router.AppRouter), new(*router.AppRouterImpl)),
		restHandler.NewAppRestHandlerImpl,
		wire.Bind(new(restHandler.AppRestHandler), new(*restHandler.AppRestHandlerImpl)),

		app.NewAppCrudOperationServiceImpl,
		wire.Bind(new(app.AppCrudOperationService), new(*app.AppCrudOperationServiceImpl)),
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
		repository3.NewGitMaterialHistoryRepositoyImpl,
		wire.Bind(new(repository3.GitMaterialHistoryRepository), new(*repository3.GitMaterialHistoryRepositoryImpl)),

		history3.NewCiTemplateHistoryServiceImpl,
		wire.Bind(new(history3.CiTemplateHistoryService), new(*history3.CiTemplateHistoryServiceImpl)),

		repository3.NewCiTemplateHistoryRepositoryImpl,
		wire.Bind(new(repository3.CiTemplateHistoryRepository), new(*repository3.CiTemplateHistoryRepositoryImpl)),

		history3.NewCiPipelineHistoryServiceImpl,
		wire.Bind(new(history3.CiPipelineHistoryService), new(*history3.CiPipelineHistoryServiceImpl)),

		repository3.NewCiPipelineHistoryRepositoryImpl,
		wire.Bind(new(repository3.CiPipelineHistoryRepository), new(*repository3.CiPipelineHistoryRepositoryImpl)),

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
		history3.NewGitMaterialHistoryServiceImpl,
		wire.Bind(new(history3.GitMaterialHistoryService), new(*history3.GitMaterialHistoryServiceImpl)),

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

		connection.NewArgoCDConnectionManagerImpl,
		wire.Bind(new(connection.ArgoCDConnectionManager), new(*connection.ArgoCDConnectionManagerImpl)),
		argo.NewArgoUserServiceImpl,
		wire.Bind(new(argo.ArgoUserService), new(*argo.ArgoUserServiceImpl)),
		util2.GetDevtronSecretName,
		//	AuthWireSet,

		cron.NewCdApplicationStatusUpdateHandlerImpl,
		wire.Bind(new(cron.CdApplicationStatusUpdateHandler), new(*cron.CdApplicationStatusUpdateHandlerImpl)),

		//app_status
		appStatusRepo.NewAppStatusRepositoryImpl,
		wire.Bind(new(appStatusRepo.AppStatusRepository), new(*appStatusRepo.AppStatusRepositoryImpl)),
		appStatus.NewAppStatusServiceImpl,
		wire.Bind(new(appStatus.AppStatusService), new(*appStatus.AppStatusServiceImpl)),
		//app_status ends

		cron.GetCiWorkflowStatusUpdateConfig,
		cron.NewCiStatusUpdateCronImpl,
		wire.Bind(new(cron.CiStatusUpdateCron), new(*cron.CiStatusUpdateCronImpl)),

		cron.GetCiTriggerCronConfig,
		cron.NewCiTriggerCronImpl,
		wire.Bind(new(cron.CiTriggerCron), new(*cron.CiTriggerCronImpl)),

		restHandler.NewPipelineStatusTimelineRestHandlerImpl,
		wire.Bind(new(restHandler.PipelineStatusTimelineRestHandler), new(*restHandler.PipelineStatusTimelineRestHandlerImpl)),

		status.NewPipelineStatusTimelineServiceImpl,
		wire.Bind(new(status.PipelineStatusTimelineService), new(*status.PipelineStatusTimelineServiceImpl)),

		router.NewUserAttributesRouterImpl,
		wire.Bind(new(router.UserAttributesRouter), new(*router.UserAttributesRouterImpl)),
		restHandler.NewUserAttributesRestHandlerImpl,
		wire.Bind(new(restHandler.UserAttributesRestHandler), new(*restHandler.UserAttributesRestHandlerImpl)),
		attributes.NewUserAttributesServiceImpl,
		wire.Bind(new(attributes.UserAttributesService), new(*attributes.UserAttributesServiceImpl)),
		repository.NewUserAttributesRepositoryImpl,
		wire.Bind(new(repository.UserAttributesRepository), new(*repository.UserAttributesRepositoryImpl)),
		pipelineConfig.NewPipelineStatusTimelineRepositoryImpl,
		wire.Bind(new(pipelineConfig.PipelineStatusTimelineRepository), new(*pipelineConfig.PipelineStatusTimelineRepositoryImpl)),
		wire.Bind(new(pipeline.DeploymentConfigService), new(*pipeline.DeploymentConfigServiceImpl)),
		pipeline.NewDeploymentConfigServiceImpl,
		pipelineConfig.NewCiTemplateOverrideRepositoryImpl,
		wire.Bind(new(pipelineConfig.CiTemplateOverrideRepository), new(*pipelineConfig.CiTemplateOverrideRepositoryImpl)),
		pipelineConfig.NewCiBuildConfigRepositoryImpl,
		wire.Bind(new(pipelineConfig.CiBuildConfigRepository), new(*pipelineConfig.CiBuildConfigRepositoryImpl)),
		pipeline.NewCiBuildConfigServiceImpl,
		wire.Bind(new(pipeline.CiBuildConfigService), new(*pipeline.CiBuildConfigServiceImpl)),
		pipeline.NewCiTemplateServiceImpl,
		wire.Bind(new(pipeline.CiTemplateService), new(*pipeline.CiTemplateServiceImpl)),
		router.NewGlobalCMCSRouterImpl,
		wire.Bind(new(router.GlobalCMCSRouter), new(*router.GlobalCMCSRouterImpl)),
		restHandler.NewGlobalCMCSRestHandlerImpl,
		wire.Bind(new(restHandler.GlobalCMCSRestHandler), new(*restHandler.GlobalCMCSRestHandlerImpl)),
		pipeline.NewGlobalCMCSServiceImpl,
		wire.Bind(new(pipeline.GlobalCMCSService), new(*pipeline.GlobalCMCSServiceImpl)),
		repository.NewGlobalCMCSRepositoryImpl,
		wire.Bind(new(repository.GlobalCMCSRepository), new(*repository.GlobalCMCSRepositoryImpl)),

		//chartRepoRepository.NewGlobalStrategyMetadataRepositoryImpl,
		//wire.Bind(new(chartRepoRepository.GlobalStrategyMetadataRepository), new(*chartRepoRepository.GlobalStrategyMetadataRepositoryImpl)),
		chartRepoRepository.NewGlobalStrategyMetadataChartRefMappingRepositoryImpl,
		wire.Bind(new(chartRepoRepository.GlobalStrategyMetadataChartRefMappingRepository), new(*chartRepoRepository.GlobalStrategyMetadataChartRefMappingRepositoryImpl)),

		status.NewPipelineStatusTimelineResourcesServiceImpl,
		wire.Bind(new(status.PipelineStatusTimelineResourcesService), new(*status.PipelineStatusTimelineResourcesServiceImpl)),
		pipelineConfig.NewPipelineStatusTimelineResourcesRepositoryImpl,
		wire.Bind(new(pipelineConfig.PipelineStatusTimelineResourcesRepository), new(*pipelineConfig.PipelineStatusTimelineResourcesRepositoryImpl)),

		status.NewPipelineStatusSyncDetailServiceImpl,
		wire.Bind(new(status.PipelineStatusSyncDetailService), new(*status.PipelineStatusSyncDetailServiceImpl)),
		pipelineConfig.NewPipelineStatusSyncDetailRepositoryImpl,
		wire.Bind(new(pipelineConfig.PipelineStatusSyncDetailRepository), new(*pipelineConfig.PipelineStatusSyncDetailRepositoryImpl)),

		repository7.NewK8sResourceHistoryRepositoryImpl,
		wire.Bind(new(repository7.K8sResourceHistoryRepository), new(*repository7.K8sResourceHistoryRepositoryImpl)),

		kubernetesResourceAuditLogs.Newk8sResourceHistoryServiceImpl,
		wire.Bind(new(kubernetesResourceAuditLogs.K8sResourceHistoryService), new(*kubernetesResourceAuditLogs.K8sResourceHistoryServiceImpl)),

		router.NewResourceGroupingRouterImpl,
		wire.Bind(new(router.ResourceGroupingRouter), new(*router.ResourceGroupingRouterImpl)),
		restHandler.NewResourceGroupRestHandlerImpl,
		wire.Bind(new(restHandler.ResourceGroupRestHandler), new(*restHandler.ResourceGroupRestHandlerImpl)),
		resourceGroup2.NewResourceGroupServiceImpl,
		wire.Bind(new(resourceGroup2.ResourceGroupService), new(*resourceGroup2.ResourceGroupServiceImpl)),
		resourceGroup.NewResourceGroupRepositoryImpl,
		wire.Bind(new(resourceGroup.ResourceGroupRepository), new(*resourceGroup.ResourceGroupRepositoryImpl)),
		resourceGroup.NewResourceGroupMappingRepositoryImpl,
		wire.Bind(new(resourceGroup.ResourceGroupMappingRepository), new(*resourceGroup.ResourceGroupMappingRepositoryImpl)),
		executors.NewArgoWorkflowExecutorImpl,
		wire.Bind(new(executors.ArgoWorkflowExecutor), new(*executors.ArgoWorkflowExecutorImpl)),
		executors.NewSystemWorkflowExecutorImpl,
		wire.Bind(new(executors.SystemWorkflowExecutor), new(*executors.SystemWorkflowExecutorImpl)),
		repository5.NewManifestPushConfigRepository,
		wire.Bind(new(repository5.ManifestPushConfigRepository), new(*repository5.ManifestPushConfigRepositoryImpl)),
		app.NewGitOpsManifestPushServiceImpl,
		wire.Bind(new(app.GitOpsPushService), new(*app.GitOpsManifestPushServiceImpl)),

		// start: docker registry wire set injection
		router.NewDockerRegRouterImpl,
		wire.Bind(new(router.DockerRegRouter), new(*router.DockerRegRouterImpl)),
		restHandler.NewDockerRegRestHandlerExtendedImpl,
		wire.Bind(new(restHandler.DockerRegRestHandler), new(*restHandler.DockerRegRestHandlerExtendedImpl)),
		pipeline.NewDockerRegistryConfigImpl,
		wire.Bind(new(pipeline.DockerRegistryConfig), new(*pipeline.DockerRegistryConfigImpl)),
		dockerRegistry.NewDockerRegistryIpsConfigServiceImpl,
		wire.Bind(new(dockerRegistry.DockerRegistryIpsConfigService), new(*dockerRegistry.DockerRegistryIpsConfigServiceImpl)),
		dockerRegistryRepository.NewDockerArtifactStoreRepositoryImpl,
		wire.Bind(new(dockerRegistryRepository.DockerArtifactStoreRepository), new(*dockerRegistryRepository.DockerArtifactStoreRepositoryImpl)),
		dockerRegistryRepository.NewDockerRegistryIpsConfigRepositoryImpl,
		wire.Bind(new(dockerRegistryRepository.DockerRegistryIpsConfigRepository), new(*dockerRegistryRepository.DockerRegistryIpsConfigRepositoryImpl)),
		dockerRegistryRepository.NewOCIRegistryConfigRepositoryImpl,
		wire.Bind(new(dockerRegistryRepository.OCIRegistryConfigRepository), new(*dockerRegistryRepository.OCIRegistryConfigRepositoryImpl)),

		// end: docker registry wire set injection

		resourceQualifiers.NewQualifiersMappingRepositoryImpl,
		wire.Bind(new(resourceQualifiers.QualifiersMappingRepository), new(*resourceQualifiers.QualifiersMappingRepositoryImpl)),

		resourceQualifiers.NewQualifierMappingServiceImpl,
		wire.Bind(new(resourceQualifiers.QualifierMappingService), new(*resourceQualifiers.QualifierMappingServiceImpl)),

		repository9.NewDevtronResourceSearchableKeyRepositoryImpl,
		wire.Bind(new(repository9.DevtronResourceSearchableKeyRepository), new(*repository9.DevtronResourceSearchableKeyRepositoryImpl)),

		devtronResource.NewDevtronResourceSearchableKeyServiceImpl,
		wire.Bind(new(devtronResource.DevtronResourceSearchableKeyService), new(*devtronResource.DevtronResourceSearchableKeyServiceImpl)),

		argocdServer.NewArgoClientWrapperServiceImpl,
		wire.Bind(new(argocdServer.ArgoClientWrapperService), new(*argocdServer.ArgoClientWrapperServiceImpl)),
	)
	return &App{}, nil
}
