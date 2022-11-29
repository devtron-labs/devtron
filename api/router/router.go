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

package router

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/apiToken"
	"github.com/devtron-labs/devtron/api/appStore"
	appStoreDeployment "github.com/devtron-labs/devtron/api/appStore/deployment"
	"github.com/devtron-labs/devtron/api/chartRepo"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/dashboardEvent"
	"github.com/devtron-labs/devtron/api/deployment"
	"github.com/devtron-labs/devtron/api/externalLink"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/module"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/api/router/pubsub"
	"github.com/devtron-labs/devtron/api/server"
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/user"
	webhookHelm "github.com/devtron-labs/devtron/api/webhook/helm"
	"github.com/devtron-labs/devtron/client/cron"
	"github.com/devtron-labs/devtron/client/dashboard"
	pubsub2 "github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

type MuxRouter struct {
	logger                             *zap.SugaredLogger
	Router                             *mux.Router
	HelmRouter                         PipelineTriggerRouter
	PipelineConfigRouter               PipelineConfigRouter
	MigrateDbRouter                    MigrateDbRouter
	EnvironmentClusterMappingsRouter   cluster.EnvironmentRouter
	AppListingRouter                   AppListingRouter
	ClusterRouter                      cluster.ClusterRouter
	WebHookRouter                      WebhookRouter
	UserAuthRouter                     user.UserAuthRouter
	ApplicationRouter                  ApplicationRouter
	CDRouter                           CDRouter
	ProjectManagementRouter            ProjectManagementRouter
	GitProviderRouter                  GitProviderRouter
	GitHostRouter                      GitHostRouter
	DockerRegRouter                    DockerRegRouter
	NotificationRouter                 NotificationRouter
	TeamRouter                         team.TeamRouter
	pubsubClient                       *pubsub2.PubSubClient
	UserRouter                         user.UserRouter
	gitWebhookHandler                  pubsub.GitWebhookHandler
	workflowUpdateHandler              pubsub.WorkflowStatusUpdateHandler
	appUpdateHandler                   pubsub.ApplicationStatusUpdateHandler
	ciEventHandler                     pubsub.CiEventHandler
	ChartRefRouter                     ChartRefRouter
	ConfigMapRouter                    ConfigMapRouter
	AppStoreRouter                     appStore.AppStoreRouter
	ChartRepositoryRouter              chartRepo.ChartRepositoryRouter
	ReleaseMetricsRouter               ReleaseMetricsRouter
	deploymentGroupRouter              DeploymentGroupRouter
	chartGroupRouter                   ChartGroupRouter
	batchOperationRouter               BatchOperationRouter
	testSuitRouter                     TestSuitRouter
	imageScanRouter                    ImageScanRouter
	policyRouter                       PolicyRouter
	gitOpsConfigRouter                 GitOpsConfigRouter
	dashboardRouter                    dashboard.DashboardRouter
	attributesRouter                   AttributesRouter
	userAttributesRouter               UserAttributesRouter
	commonRouter                       CommonRouter
	grafanaRouter                      GrafanaRouter
	ssoLoginRouter                     sso.SsoLoginRouter
	telemetryRouter                    TelemetryRouter
	telemetryWatcher                   telemetry.TelemetryEventClient
	bulkUpdateRouter                   BulkUpdateRouter
	WebhookListenerRouter              WebhookListenerRouter
	appRouter                          AppRouter
	coreAppRouter                      CoreAppRouter
	helmAppRouter                      client.HelmAppRouter
	k8sApplicationRouter               k8s.K8sApplicationRouter
	pProfRouter                        PProfRouter
	deploymentConfigRouter             deployment.DeploymentConfigRouter
	dashboardTelemetryRouter           dashboardEvent.DashboardTelemetryRouter
	commonDeploymentRouter             appStoreDeployment.CommonDeploymentRouter
	globalPluginRouter                 GlobalPluginRouter
	externalLinkRouter                 externalLink.ExternalLinkRouter
	moduleRouter                       module.ModuleRouter
	serverRouter                       server.ServerRouter
	apiTokenRouter                     apiToken.ApiTokenRouter
	helmApplicationStatusUpdateHandler cron.CdApplicationStatusUpdateHandler
	k8sCapacityRouter                  k8s.K8sCapacityRouter
	webhookHelmRouter                  webhookHelm.WebhookHelmRouter
	globalCMCSRouter                   GlobalCMCSRouter
}

func NewMuxRouter(logger *zap.SugaredLogger, HelmRouter PipelineTriggerRouter, PipelineConfigRouter PipelineConfigRouter,
	MigrateDbRouter MigrateDbRouter, AppListingRouter AppListingRouter,
	EnvironmentClusterMappingsRouter cluster.EnvironmentRouter, ClusterRouter cluster.ClusterRouter,
	WebHookRouter WebhookRouter, UserAuthRouter user.UserAuthRouter, ApplicationRouter ApplicationRouter,
	CDRouter CDRouter, ProjectManagementRouter ProjectManagementRouter,
	GitProviderRouter GitProviderRouter, GitHostRouter GitHostRouter,
	DockerRegRouter DockerRegRouter,
	NotificationRouter NotificationRouter,
	TeamRouter team.TeamRouter,
	gitWebhookHandler pubsub.GitWebhookHandler,
	workflowUpdateHandler pubsub.WorkflowStatusUpdateHandler,
	appUpdateHandler pubsub.ApplicationStatusUpdateHandler,
	ciEventHandler pubsub.CiEventHandler, pubsubClient *pubsub2.PubSubClient, UserRouter user.UserRouter,
	ChartRefRouter ChartRefRouter, ConfigMapRouter ConfigMapRouter, AppStoreRouter appStore.AppStoreRouter, chartRepositoryRouter chartRepo.ChartRepositoryRouter,
	ReleaseMetricsRouter ReleaseMetricsRouter, deploymentGroupRouter DeploymentGroupRouter, batchOperationRouter BatchOperationRouter,
	chartGroupRouter ChartGroupRouter, testSuitRouter TestSuitRouter, imageScanRouter ImageScanRouter,
	policyRouter PolicyRouter, gitOpsConfigRouter GitOpsConfigRouter, dashboardRouter dashboard.DashboardRouter, attributesRouter AttributesRouter, userAttributesRouter UserAttributesRouter,
	commonRouter CommonRouter, grafanaRouter GrafanaRouter, ssoLoginRouter sso.SsoLoginRouter, telemetryRouter TelemetryRouter, telemetryWatcher telemetry.TelemetryEventClient, bulkUpdateRouter BulkUpdateRouter, webhookListenerRouter WebhookListenerRouter, appRouter AppRouter,
	coreAppRouter CoreAppRouter, helmAppRouter client.HelmAppRouter, k8sApplicationRouter k8s.K8sApplicationRouter,
	pProfRouter PProfRouter, deploymentConfigRouter deployment.DeploymentConfigRouter, dashboardTelemetryRouter dashboardEvent.DashboardTelemetryRouter,
	commonDeploymentRouter appStoreDeployment.CommonDeploymentRouter, externalLinkRouter externalLink.ExternalLinkRouter,
	globalPluginRouter GlobalPluginRouter, moduleRouter module.ModuleRouter,
	serverRouter server.ServerRouter, apiTokenRouter apiToken.ApiTokenRouter,
	helmApplicationStatusUpdateHandler cron.CdApplicationStatusUpdateHandler, k8sCapacityRouter k8s.K8sCapacityRouter,
	webhookHelmRouter webhookHelm.WebhookHelmRouter, globalCMCSRouter GlobalCMCSRouter) *MuxRouter {
	r := &MuxRouter{
		Router:                             mux.NewRouter(),
		HelmRouter:                         HelmRouter,
		PipelineConfigRouter:               PipelineConfigRouter,
		MigrateDbRouter:                    MigrateDbRouter,
		EnvironmentClusterMappingsRouter:   EnvironmentClusterMappingsRouter,
		AppListingRouter:                   AppListingRouter,
		ClusterRouter:                      ClusterRouter,
		WebHookRouter:                      WebHookRouter,
		UserAuthRouter:                     UserAuthRouter,
		ApplicationRouter:                  ApplicationRouter,
		CDRouter:                           CDRouter,
		ProjectManagementRouter:            ProjectManagementRouter,
		DockerRegRouter:                    DockerRegRouter,
		GitProviderRouter:                  GitProviderRouter,
		GitHostRouter:                      GitHostRouter,
		NotificationRouter:                 NotificationRouter,
		TeamRouter:                         TeamRouter,
		logger:                             logger,
		gitWebhookHandler:                  gitWebhookHandler,
		workflowUpdateHandler:              workflowUpdateHandler,
		appUpdateHandler:                   appUpdateHandler,
		ciEventHandler:                     ciEventHandler,
		pubsubClient:                       pubsubClient,
		UserRouter:                         UserRouter,
		ChartRefRouter:                     ChartRefRouter,
		ConfigMapRouter:                    ConfigMapRouter,
		AppStoreRouter:                     AppStoreRouter,
		ChartRepositoryRouter:              chartRepositoryRouter,
		ReleaseMetricsRouter:               ReleaseMetricsRouter,
		deploymentGroupRouter:              deploymentGroupRouter,
		batchOperationRouter:               batchOperationRouter,
		chartGroupRouter:                   chartGroupRouter,
		testSuitRouter:                     testSuitRouter,
		imageScanRouter:                    imageScanRouter,
		policyRouter:                       policyRouter,
		gitOpsConfigRouter:                 gitOpsConfigRouter,
		attributesRouter:                   attributesRouter,
		userAttributesRouter:               userAttributesRouter,
		dashboardRouter:                    dashboardRouter,
		commonRouter:                       commonRouter,
		grafanaRouter:                      grafanaRouter,
		ssoLoginRouter:                     ssoLoginRouter,
		telemetryRouter:                    telemetryRouter,
		telemetryWatcher:                   telemetryWatcher,
		bulkUpdateRouter:                   bulkUpdateRouter,
		WebhookListenerRouter:              webhookListenerRouter,
		appRouter:                          appRouter,
		coreAppRouter:                      coreAppRouter,
		helmAppRouter:                      helmAppRouter,
		k8sApplicationRouter:               k8sApplicationRouter,
		pProfRouter:                        pProfRouter,
		deploymentConfigRouter:             deploymentConfigRouter,
		dashboardTelemetryRouter:           dashboardTelemetryRouter,
		commonDeploymentRouter:             commonDeploymentRouter,
		externalLinkRouter:                 externalLinkRouter,
		globalPluginRouter:                 globalPluginRouter,
		moduleRouter:                       moduleRouter,
		serverRouter:                       serverRouter,
		apiTokenRouter:                     apiTokenRouter,
		helmApplicationStatusUpdateHandler: helmApplicationStatusUpdateHandler,
		k8sCapacityRouter:                  k8sCapacityRouter,
		webhookHelmRouter:                  webhookHelmRouter,
		globalCMCSRouter:                   globalCMCSRouter,
	}
	return r
}

func (router *MuxRouter) Init() {

	router.Router.PathPrefix("/orchestrator/api/vi/pod/exec/ws").Handler(terminal.CreateAttachHandler("/orchestrator/api/vi/pod/exec/ws"))

	router.Router.StrictSlash(true)
	router.Router.Handle("/metrics", promhttp.Handler())
	//prometheus.MustRegister(pipeline.CiTriggerCounter)
	//prometheus.MustRegister(app.CdTriggerCounter)
	router.Router.Path("/health").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := common.Response{}
		response.Code = 200
		response.Result = "OK"
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			router.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})

	router.Router.Path("/orchestrator/version").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := common.Response{}
		response.Code = 200
		response.Result = util.GetDevtronVersion()
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			router.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})
	coreAppRouter := router.Router.PathPrefix("/orchestrator/core").Subrouter()
	router.coreAppRouter.initCoreAppRouter(coreAppRouter)

	pipelineConfigRouter := router.Router.PathPrefix("/orchestrator/app").Subrouter()
	router.PipelineConfigRouter.initPipelineConfigRouter(pipelineConfigRouter)
	router.AppListingRouter.initAppListingRouter(pipelineConfigRouter)
	router.HelmRouter.initPipelineTriggerRouter(pipelineConfigRouter)
	router.appRouter.initAppRouter(pipelineConfigRouter)

	migrateRouter := router.Router.PathPrefix("/orchestrator/migrate").Subrouter()
	router.MigrateDbRouter.InitMigrateDbRouter(migrateRouter)

	environmentClusterMappingsRouter := router.Router.PathPrefix("/orchestrator/env").Subrouter()
	router.EnvironmentClusterMappingsRouter.InitEnvironmentClusterMappingsRouter(environmentClusterMappingsRouter)

	clusterRouter := router.Router.PathPrefix("/orchestrator/cluster").Subrouter()
	router.ClusterRouter.InitClusterRouter(clusterRouter)

	webHookRouter := router.Router.PathPrefix("/orchestrator/webhook").Subrouter()
	router.WebHookRouter.intWebhookRouter(webHookRouter)

	applicationRouter := router.Router.PathPrefix("/orchestrator/api/v1/applications").Subrouter()
	router.ApplicationRouter.initApplicationRouter(applicationRouter)

	rootRouter := router.Router.PathPrefix("/orchestrator").Subrouter()
	router.UserAuthRouter.InitUserAuthRouter(rootRouter)

	projectManagementRouter := router.Router.PathPrefix("/orchestrator/project-management").Subrouter()
	router.ProjectManagementRouter.InitProjectManagementRouter(projectManagementRouter)

	gitRouter := router.Router.PathPrefix("/orchestrator/git").Subrouter()
	router.GitProviderRouter.InitGitProviderRouter(gitRouter)
	router.GitHostRouter.InitGitHostRouter(gitRouter)

	dockerRouter := router.Router.PathPrefix("/orchestrator/docker").Subrouter()
	router.DockerRegRouter.InitDockerRegRouter(dockerRouter)

	notificationRouter := router.Router.PathPrefix("/orchestrator/notification").Subrouter()
	router.NotificationRouter.InitNotificationRegRouter(notificationRouter)

	teamRouter := router.Router.PathPrefix("/orchestrator/team").Subrouter()
	router.TeamRouter.InitTeamRouter(teamRouter)

	userRouter := router.Router.PathPrefix("/orchestrator/user").Subrouter()
	router.UserRouter.InitUserRouter(userRouter)

	chartRefRouter := router.Router.PathPrefix("/orchestrator/chartref").Subrouter()
	router.ChartRefRouter.initChartRefRouter(chartRefRouter)

	configMapRouter := router.Router.PathPrefix("/orchestrator/config").Subrouter()
	router.ConfigMapRouter.initConfigMapRouter(configMapRouter)

	appStoreRouter := router.Router.PathPrefix("/orchestrator/app-store").Subrouter()
	router.AppStoreRouter.Init(appStoreRouter)

	chartRepoRouter := router.Router.PathPrefix("/orchestrator/chart-repo").Subrouter()
	router.ChartRepositoryRouter.Init(chartRepoRouter)

	deploymentMetricsRouter := router.Router.PathPrefix("/orchestrator/deployment-metrics").Subrouter()
	router.ReleaseMetricsRouter.initReleaseMetricsRouter(deploymentMetricsRouter)

	deploymentGroupRouter := router.Router.PathPrefix("/orchestrator/deployment-group").Subrouter()
	router.deploymentGroupRouter.initDeploymentGroupRouter(deploymentGroupRouter)

	router.batchOperationRouter.initBatchOperationRouter(rootRouter)

	chartGroupRouter := router.Router.PathPrefix("/orchestrator/chart-group").Subrouter()
	router.chartGroupRouter.initChartGroupRouter(chartGroupRouter)

	testSuitRouter := router.Router.PathPrefix("/orchestrator/test-report").Subrouter()
	router.testSuitRouter.InitTestSuitRouter(testSuitRouter)

	imageScanRouter := router.Router.PathPrefix("/orchestrator/security/scan").Subrouter()
	router.imageScanRouter.InitImageScanRouter(imageScanRouter)

	policyRouter := router.Router.PathPrefix("/orchestrator/security/policy").Subrouter()
	router.policyRouter.InitPolicyRouter(policyRouter)

	gitOpsRouter := router.Router.PathPrefix("/orchestrator/gitops").Subrouter()
	router.gitOpsConfigRouter.InitGitOpsConfigRouter(gitOpsRouter)

	attributeRouter := router.Router.PathPrefix("/orchestrator/attributes").Subrouter()
	router.attributesRouter.initAttributesRouter(attributeRouter)

	userAttributeRouter := router.Router.PathPrefix("/orchestrator/attributes/user").Subrouter()
	router.userAttributesRouter.InitUserAttributesRouter(userAttributeRouter)

	dashboardRouter := router.Router.PathPrefix("/dashboard").Subrouter()
	router.dashboardRouter.InitDashboardRouter(dashboardRouter)

	grafanaRouter := router.Router.PathPrefix("/grafana").Subrouter()
	router.grafanaRouter.initGrafanaRouter(grafanaRouter)

	router.Router.Path("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/dashboard", 301)
	})

	commonRouter := router.Router.PathPrefix("/orchestrator/global").Subrouter()
	router.commonRouter.InitCommonRouter(commonRouter)

	ssoLoginRouter := router.Router.PathPrefix("/orchestrator/sso").Subrouter()
	router.ssoLoginRouter.InitSsoLoginRouter(ssoLoginRouter)

	telemetryRouter := router.Router.PathPrefix("/orchestrator/telemetry").Subrouter()
	router.telemetryRouter.InitTelemetryRouter(telemetryRouter)

	bulkUpdateRouter := router.Router.PathPrefix("/orchestrator/batch").Subrouter()
	router.bulkUpdateRouter.initBulkUpdateRouter(bulkUpdateRouter)

	webhookListenerRouter := router.Router.PathPrefix("/orchestrator/webhook/git").Subrouter()
	router.WebhookListenerRouter.InitWebhookListenerRouter(webhookListenerRouter)

	k8sApp := router.Router.PathPrefix("/orchestrator/k8s").Subrouter()
	router.k8sApplicationRouter.InitK8sApplicationRouter(k8sApp)

	pProfListenerRouter := router.Router.PathPrefix("/orchestrator/debug/pprof").Subrouter()
	router.pProfRouter.initPProfRouter(pProfListenerRouter)

	globalPluginRouter := router.Router.PathPrefix("/orchestrator/plugin").Subrouter()
	router.globalPluginRouter.initGlobalPluginRouter(globalPluginRouter)

	//  deployment router starts
	deploymentConfigSubRouter := router.Router.PathPrefix("/orchestrator/deployment/template").Subrouter()
	router.deploymentConfigRouter.Init(deploymentConfigSubRouter)
	// deployment router ends

	//  dashboard event router starts
	dashboardTelemetryRouter := router.Router.PathPrefix("/orchestrator/dashboard-event").Subrouter()
	router.dashboardTelemetryRouter.Init(dashboardTelemetryRouter)
	// dashboard event router ends

	//GitOps,Acd + HelmCLi both apps deployment related api's
	applicationSubRouter := router.Router.PathPrefix("/orchestrator/application").Subrouter()
	router.commonDeploymentRouter.Init(applicationSubRouter)
	//this router must placed after commonDeploymentRouter
	router.helmAppRouter.InitAppListRouter(applicationSubRouter)

	externalLinkRouter := router.Router.PathPrefix("/orchestrator/external-links").Subrouter()
	router.externalLinkRouter.InitExternalLinkRouter(externalLinkRouter)

	// module router
	moduleRouter := router.Router.PathPrefix("/orchestrator/module").Subrouter()
	router.moduleRouter.Init(moduleRouter)

	// server router
	serverRouter := router.Router.PathPrefix("/orchestrator/server").Subrouter()
	router.serverRouter.Init(serverRouter)

	// api-token router
	apiTokenRouter := router.Router.PathPrefix("/orchestrator/api-token").Subrouter()
	router.apiTokenRouter.InitApiTokenRouter(apiTokenRouter)

	k8sCapacityApp := router.Router.PathPrefix("/orchestrator/k8s/capacity").Subrouter()
	router.k8sCapacityRouter.InitK8sCapacityRouter(k8sCapacityApp)

	// webhook helm app router
	webhookHelmRouter := router.Router.PathPrefix("/orchestrator/webhook/helm").Subrouter()
	router.webhookHelmRouter.InitWebhookHelmRouter(webhookHelmRouter)

	globalCMCSRouter := router.Router.PathPrefix("/orchestrator/global/cm-cs").Subrouter()
	router.globalCMCSRouter.initGlobalCMCSRouter(globalCMCSRouter)
}
