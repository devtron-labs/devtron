//+build wireinject

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
	"github.com/devtron-labs/devtron/api/router/pubsub"
	eClient "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/client/grafana"
	jClient "github.com/devtron-labs/devtron/client/jira"
	"github.com/devtron-labs/devtron/client/lens"
	pubsub2 "github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/casbin"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	appWorkflow2 "github.com/devtron-labs/devtron/internal/sql/repository/appWorkflow"
	appstore2 "github.com/devtron-labs/devtron/internal/sql/repository/appstore"
	"github.com/devtron-labs/devtron/internal/sql/repository/appstore/chartGroup"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	security2 "github.com/devtron-labs/devtron/internal/sql/repository/security"
	teamRepo "github.com/devtron-labs/devtron/internal/sql/repository/team"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/internal/util/ArgoUtil"
	"github.com/devtron-labs/devtron/pkg/appClone"
	"github.com/devtron-labs/devtron/pkg/appClone/batch"
	"github.com/devtron-labs/devtron/pkg/appWorkflow"
	"github.com/devtron-labs/devtron/pkg/appstore"
	"github.com/devtron-labs/devtron/pkg/commonService"
	"github.com/devtron-labs/devtron/pkg/deploymentGroup"
	"github.com/devtron-labs/devtron/pkg/dex"
	"github.com/devtron-labs/devtron/pkg/event"
	"github.com/devtron-labs/devtron/pkg/git"
	jira2 "github.com/devtron-labs/devtron/pkg/jira"
	"github.com/devtron-labs/devtron/pkg/notifier"
	"github.com/devtron-labs/devtron/pkg/projectManagementService/jira"
	"github.com/devtron-labs/devtron/pkg/security"
	"github.com/devtron-labs/devtron/pkg/sso"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/util/rbac"

	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/router"
	"github.com/devtron-labs/devtron/api/sse"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	cluster2 "github.com/devtron-labs/devtron/client/argocdServer/cluster"
	repository2 "github.com/devtron-labs/devtron/client/argocdServer/repository"
	session2 "github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/cluster"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	clusterAccounts2 "github.com/devtron-labs/devtron/pkg/cluster"

	"github.com/devtron-labs/devtron/pkg/pipeline"
	user2 "github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/session"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {

	wire.Build(
		//-----------------
		gitSensor.GetGitSensorConfig,
		gitSensor.NewGitSensorSession,
		wire.Bind(new(gitSensor.GitSensorClient), new(*gitSensor.GitSensorClientImpl)),
		//--------
		helper.NewAppListingRepositoryQueryBuilder,
		models.GetConfig,
		eClient.GetEventClientConfig,
		models.NewDbConnection,
		//app.GetACDAuthConfig,
		user2.GetACDAuthConfig,
		wire.Value(pipeline.RefChartDir("scripts/devtron-reference-helm-charts")),
		wire.Value(appstore.RefChartProxyDir("scripts/devtron-reference-helm-charts")),
		wire.Value(pipeline.DefaultChart("reference-app-rolling")),
		wire.Value(util.ChartWorkingDir("/tmp/charts/")),
		util.GetGitConfig,
		session.SettingsManager,
		session.CDSettingsManager,
		session.SessionManager,
		//auth.GetConfig,
		casbin.Create,
		rbac.NewEnforcerImpl,
		wire.Bind(new(rbac.Enforcer), new(*rbac.EnforcerImpl)),

		dex.GetConfig,
		argocdServer.GetConfig,
		session2.NewSessionServiceClient,
		wire.Bind(new(session2.ServiceClient), new(*session2.ServiceClientImpl)),

		sse.NewSSE,
		router.NewHelmRouter,
		wire.Bind(new(router.HelmRouter), new(*router.HelmRouterImpl)),

		restHandler.NewPipelineRestHandler,
		wire.Bind(new(restHandler.PipelineTriggerRestHandler), new(*restHandler.PipelineTriggerRestHandlerImpl)),

		app.NewAppService,
		wire.Bind(new(app.AppService), new(*app.AppServiceImpl)),

		chartConfig.NewChartRepository,
		wire.Bind(new(chartConfig.ChartRepository), new(*chartConfig.ChartRepositoryImpl)),
		chartConfig.NewEnvConfigOverrideRepository,
		wire.Bind(new(chartConfig.EnvConfigOverrideRepository), new(*chartConfig.EnvConfigOverrideRepositoryImpl)),
		chartConfig.NewPipelineOverrideRepository,
		wire.Bind(new(chartConfig.PipelineOverrideRepository), new(*chartConfig.PipelineOverrideRepositoryImpl)),
		util.MergeUtil{},
		util.NewSugardLogger,
		router.NewMuxRouter,

		pipelineConfig.NewAppRepositoryImpl,
		wire.Bind(new(pipelineConfig.AppRepository), new(*pipelineConfig.AppRepositoryImpl)),

		pipeline.NewPipelineBuilderImpl,
		wire.Bind(new(pipeline.PipelineBuilder), new(*pipeline.PipelineBuilderImpl)),
		restHandler.NewPipelineRestHandlerImpl,
		wire.Bind(new(restHandler.PipelineConfigRestHandler), new(*restHandler.PipelineConfigRestHandlerImpl)),
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
		pipeline.NewChartServiceImpl,
		wire.Bind(new(pipeline.ChartService), new(*pipeline.ChartServiceImpl)),
		chartConfig.NewChartRepoRepositoryImpl,
		wire.Bind(new(chartConfig.ChartRepoRepository), new(*chartConfig.ChartRepoRepositoryImpl)),
		chartConfig.NewChartRefRepositoryImpl,
		wire.Bind(new(chartConfig.ChartRefRepository), new(*chartConfig.ChartRefRepositoryImpl)),

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

		cluster.NewClusterRepositoryImpl,
		wire.Bind(new(cluster.ClusterRepository), new(*cluster.ClusterRepositoryImpl)),
		clusterAccounts2.NewClusterServiceImpl,
		wire.Bind(new(clusterAccounts2.ClusterService), new(*clusterAccounts2.ClusterServiceImpl)),
		restHandler.NewClusterRestHandlerImpl,
		wire.Bind(new(restHandler.ClusterRestHandler), new(*restHandler.ClusterRestHandlerImpl)),
		router.NewClusterRouterImpl,
		wire.Bind(new(router.ClusterRouter), new(*router.ClusterRouterImpl)),

		cluster.NewClusterAccountsRepositoryImpl,
		wire.Bind(new(cluster.ClusterAccountsRepository), new(*cluster.ClusterAccountsRepositoryImpl)),
		clusterAccounts2.NewClusterAccountsServiceImpl,
		wire.Bind(new(clusterAccounts2.ClusterAccountsService), new(*clusterAccounts2.ClusterAccountsServiceImpl)),
		restHandler.NewClusterAccountsRestHandlerImpl,
		wire.Bind(new(restHandler.ClusterAccountsRestHandler), new(*restHandler.ClusterAccountsRestHandlerImpl)),
		router.NewClusterAccountsRouterImpl,
		wire.Bind(new(router.ClusterAccountsRouter), new(*router.ClusterAccountsRouterImpl)),

		cluster.NewEnvironmentRepositoryImpl,
		wire.Bind(new(cluster.EnvironmentRepository), new(*cluster.EnvironmentRepositoryImpl)),
		clusterAccounts2.NewEnvironmentServiceImpl,
		wire.Bind(new(clusterAccounts2.EnvironmentService), new(*clusterAccounts2.EnvironmentServiceImpl)),
		restHandler.NewEnvironmentRestHandlerImpl,
		wire.Bind(new(restHandler.EnvironmentRestHandler), new(*restHandler.EnvironmentRestHandlerImpl)),
		router.NewEnvironmentRouterImpl,
		wire.Bind(new(router.EnvironmentRouter), new(*router.EnvironmentRouterImpl)),

		cluster.NewClusterHelmConfigRepositoryImpl,
		wire.Bind(new(cluster.ClusterHelmConfigRepository), new(*cluster.ClusterHelmConfigRepositoryImpl)),
		clusterAccounts2.NewClusterHelmConfigServiceImpl,
		wire.Bind(new(clusterAccounts2.ClusterHelmConfigService), new(*clusterAccounts2.ClusterHelmConfigServiceImpl)),
		restHandler.NewClusterHelmConfigRestHandlerImpl,
		wire.Bind(new(restHandler.ClusterHelmConfigRestHandler), new(*restHandler.ClusterHelmConfigRestHandlerImpl)),
		router.NewClusterHelmConfigRouterImpl,
		wire.Bind(new(router.ClusterHelmConfigRouter), new(*router.ClusterHelmConfigRouterImpl)),

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

		user2.NewTokenCache,

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
		util.NewGitLabClient,
		//wire.Bind(new(util.GitClient), new(*util.GitLabClient)),
		util.NewGitServiceImpl,
		wire.Bind(new(util.GitService), new(*util.GitServiceImpl)),

		application.NewApplicationClientImpl,
		wire.Bind(new(application.ServiceClient), new(*application.ServiceClientImpl)),
		cluster2.NewServiceClientImpl,
		wire.Bind(new(cluster2.ServiceClient), new(*cluster2.ServiceClientImpl)),
		connector.NewPumpImpl,
		repository2.NewServiceClientImpl,
		wire.Bind(new(repository2.ServiceClient), new(*repository2.ServiceClientImpl)),
		wire.Bind(new(connector.Pump), new(*connector.PumpImpl)),
		restHandler.NewApplicationRestHandlerImpl,
		wire.Bind(new(restHandler.ApplicationRestHandler), new(*restHandler.ApplicationRestHandlerImpl)),
		router.NewApplicationRouterImpl,
		wire.Bind(new(router.ApplicationRouter), new(*router.ApplicationRouterImpl)),
		//app.GetConfig,
		router.NewUserAuthRouterImpl,
		wire.Bind(new(router.UserAuthRouter), new(*router.UserAuthRouterImpl)),
		restHandler.NewUserAuthHandlerImpl,
		wire.Bind(new(restHandler.UserAuthHandler), new(*restHandler.UserAuthHandlerImpl)),
		user2.NewUserAuthServiceImpl,
		wire.Bind(new(user2.UserAuthService), new(*user2.UserAuthServiceImpl)),
		repository.NewUserAuthRepositoryImpl,
		wire.Bind(new(repository.UserAuthRepository), new(*repository.UserAuthRepositoryImpl)),

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

		router.NewTeamRouterImpl,
		wire.Bind(new(router.TeamRouter), new(*router.TeamRouterImpl)),
		restHandler.NewTeamRestHandlerImpl,
		wire.Bind(new(restHandler.TeamRestHandler), new(*restHandler.TeamRestHandlerImpl)),
		team.NewTeamServiceImpl,
		wire.Bind(new(team.TeamService), new(*team.TeamServiceImpl)),
		teamRepo.NewTeamRepositoryImpl,
		wire.Bind(new(teamRepo.TeamRepository), new(*teamRepo.TeamRepositoryImpl)),

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

		router.NewUserRouterImpl,
		wire.Bind(new(router.UserRouter), new(*router.UserRouterImpl)),
		restHandler.NewUserRestHandlerImpl,
		wire.Bind(new(restHandler.UserRestHandler), new(*restHandler.UserRestHandlerImpl)),
		user2.NewUserServiceImpl,
		wire.Bind(new(user2.UserService), new(*user2.UserServiceImpl)),
		repository.NewUserRepositoryImpl,
		wire.Bind(new(repository.UserRepository), new(*repository.UserRepositoryImpl)),

		rbac.NewEnforcerUtilImpl,
		wire.Bind(new(rbac.EnforcerUtil), new(*rbac.EnforcerUtilImpl)),

		repository.NewEventRepositoryImpl,
		wire.Bind(new(repository.EventRepository), new(*repository.EventRepositoryImpl)),

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

		notifier.NewNotificationConfigBuilderImpl,
		wire.Bind(new(notifier.NotificationConfigBuilder), new(*notifier.NotificationConfigBuilderImpl)),

		pubsub.NewCronBasedEventReceiverImpl,
		wire.Bind(new(pubsub.CronBasedEventReceiver), new(*pubsub.CronBasedEventReceiverImpl)),

		event.NewEventServiceImpl,
		wire.Bind(new(event.EventService), new(*event.EventServiceImpl)),

		restHandler.NewInstalledAppRestHandlerImpl,
		wire.Bind(new(restHandler.InstalledAppRestHandler), new(*restHandler.InstalledAppRestHandlerImpl)),
		appstore.NewInstalledAppServiceImpl,
		wire.Bind(new(appstore.InstalledAppService), new(*appstore.InstalledAppServiceImpl)),
		appstore2.NewInstalledAppRepositoryImpl,
		wire.Bind(new(appstore2.InstalledAppRepository), new(*appstore2.InstalledAppRepositoryImpl)),

		router.NewAppStoreRouterImpl,
		wire.Bind(new(router.AppStoreRouter), new(*router.AppStoreRouterImpl)),

		restHandler.NewAppStoreRestHandlerImpl,
		wire.Bind(new(restHandler.AppStoreRestHandler), new(*restHandler.AppStoreRestHandlerImpl)),

		appstore.NewAppStoreServiceImpl,
		wire.Bind(new(appstore.AppStoreService), new(*appstore.AppStoreServiceImpl)),

		appstore2.NewAppStoreRepositoryImpl,
		wire.Bind(new(appstore2.AppStoreRepository), new(*appstore2.AppStoreRepositoryImpl)),

		appstore2.NewAppStoreApplicationVersionRepositoryImpl,
		wire.Bind(new(appstore2.AppStoreApplicationVersionRepository), new(*appstore2.AppStoreApplicationVersionRepositoryImpl)),

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

		restHandler.NewAppStoreValuesRestHandlerImpl,
		wire.Bind(new(restHandler.AppStoreValuesRestHandler), new(*restHandler.AppStoreValuesRestHandlerImpl)),
		appstore.NewAppStoreValuesServiceImpl,
		wire.Bind(new(appstore.AppStoreValuesService), new(*appstore.AppStoreValuesServiceImpl)),
		appstore2.NewAppStoreVersionValuesRepositoryImpl,
		wire.Bind(new(appstore2.AppStoreVersionValuesRepository), new(*appstore2.AppStoreVersionValuesRepositoryImpl)),

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

		chartGroup.NewChartGroupReposotoryImpl,
		wire.Bind(new(chartGroup.ChartGroupReposotory), new(*chartGroup.ChartGroupReposotoryImpl)),
		chartGroup.NewChartGroupEntriesRepositoryImpl,
		wire.Bind(new(chartGroup.ChartGroupEntriesRepository), new(*chartGroup.ChartGroupEntriesRepositoryImpl)),
		appstore.NewChartGroupServiceImpl,
		wire.Bind(new(appstore.ChartGroupService), new(*appstore.ChartGroupServiceImpl)),
		restHandler.NewChartGroupRestHandlerImpl,
		wire.Bind(new(restHandler.ChartGroupRestHandler), new(*restHandler.ChartGroupRestHandlerImpl)),
		router.NewChartGroupRouterImpl,
		wire.Bind(new(router.ChartGroupRouter), new(*router.ChartGroupRouterImpl)),
		chartGroup.NewChartGroupDeploymentRepositoryImpl,
		wire.Bind(new(chartGroup.ChartGroupDeploymentRepository), new(*chartGroup.ChartGroupDeploymentRepositoryImpl)),

		user2.NewRoleGroupServiceImpl,
		wire.Bind(new(user2.RoleGroupService), new(*user2.RoleGroupServiceImpl)),
		repository.NewRoleGroupRepositoryImpl,
		wire.Bind(new(repository.RoleGroupRepository), new(*repository.RoleGroupRepositoryImpl)),

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
		appstore2.NewClusterInstalledAppsRepositoryImpl,
		wire.Bind(new(appstore2.ClusterInstalledAppsRepository), new(*appstore2.ClusterInstalledAppsRepositoryImpl)),
		terminal.NewTerminalSessionHandlerImpl,
		wire.Bind(new(terminal.TerminalSessionHandler), new(*terminal.TerminalSessionHandlerImpl)),
		argocdServer.NewArgoK8sClientImpl,
		wire.Bind(new(argocdServer.ArgoK8sClient), new(*argocdServer.ArgoK8sClientImpl)),

		sso.NewSSOLoginServiceImpl,
		wire.Bind(new(sso.SSOLoginService), new(*sso.SSOLoginServiceImpl)),
		repository.NewSSOLoginRepositoryImpl,
		wire.Bind(new(repository.SSOLoginRepository), new(*repository.SSOLoginRepositoryImpl)),

		router.NewGitOpsConfigRouterImpl,
		wire.Bind(new(router.GitOpsConfigRouter), new(*router.GitOpsConfigRouterImpl)),
		restHandler.NewGitOpsConfigRestHandlerImpl,
		wire.Bind(new(restHandler.GitOpsConfigRestHandler), new(*restHandler.GitOpsConfigRestHandlerImpl)),
		gitops.NewGitOpsConfigServiceImpl,
		wire.Bind(new(gitops.GitOpsConfigService), new(*gitops.GitOpsConfigServiceImpl)),
		repository.NewGitOpsConfigRepositoryImpl,
		wire.Bind(new(repository.GitOpsConfigRepository), new(*repository.GitOpsConfigRepositoryImpl)),
	)
	return &App{}, nil
}
