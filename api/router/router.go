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
	appStore "github.com/devtron-labs/devtron/api/appStore"
	chartRepo "github.com/devtron-labs/devtron/api/chartRepo"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/deployment"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/api/router/pubsub"
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/user"
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
	logger                           *zap.SugaredLogger
	Router                           *mux.Router
	HelmRouter                       HelmRouter
	PipelineConfigRouter             PipelineConfigRouter
	MigrateDbRouter                  MigrateDbRouter
	EnvironmentClusterMappingsRouter cluster.EnvironmentRouter
	AppListingRouter                 AppListingRouter
	ClusterRouter                    cluster.ClusterRouter
	WebHookRouter                    WebhookRouter
	UserAuthRouter                   user.UserAuthRouter
	ApplicationRouter                ApplicationRouter
	CDRouter                         CDRouter
	ProjectManagementRouter          ProjectManagementRouter
	GitProviderRouter                GitProviderRouter
	GitHostRouter                    GitHostRouter
	DockerRegRouter                  DockerRegRouter
	NotificationRouter               NotificationRouter
	TeamRouter                       team.TeamRouter
	pubsubClient                     *pubsub2.PubSubClient
	UserRouter                       user.UserRouter
	gitWebhookHandler                pubsub.GitWebhookHandler
	workflowUpdateHandler            pubsub.WorkflowStatusUpdateHandler
	appUpdateHandler                 pubsub.ApplicationStatusUpdateHandler
	ciEventHandler                   pubsub.CiEventHandler
	cronBasedEventReceiver           pubsub.CronBasedEventReceiver
	ChartRefRouter                   ChartRefRouter
	ConfigMapRouter                  ConfigMapRouter
	AppStoreRouter                   appStore.AppStoreRouter
	ChartRepositoryRouter            chartRepo.ChartRepositoryRouter
	ReleaseMetricsRouter             ReleaseMetricsRouter
	deploymentGroupRouter            DeploymentGroupRouter
	chartGroupRouter                 ChartGroupRouter
	batchOperationRouter             BatchOperationRouter
	testSuitRouter                   TestSuitRouter
	imageScanRouter                  ImageScanRouter
	policyRouter                     PolicyRouter
	gitOpsConfigRouter               GitOpsConfigRouter
	dashboardRouter                  dashboard.DashboardRouter
	attributesRouter                 AttributesRouter
	commonRouter                     CommonRouter
	grafanaRouter                    GrafanaRouter
	ssoLoginRouter                   sso.SsoLoginRouter
	telemetryRouter                  TelemetryRouter
	telemetryWatcher                 telemetry.TelemetryEventClient
	bulkUpdateRouter                 BulkUpdateRouter
	WebhookListenerRouter            WebhookListenerRouter
	appLabelsRouter                  AppLabelRouter
	coreAppRouter                    CoreAppRouter
	helmAppRouter                    client.HelmAppRouter
	k8sApplicationRouter             k8s.K8sApplicationRouter
	pProfRouter                      PProfRouter
	deploymentConfigRouter           deployment.DeploymentConfigRouter
}

func NewMuxRouter(logger *zap.SugaredLogger, HelmRouter HelmRouter, PipelineConfigRouter PipelineConfigRouter,
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
	ciEventHandler pubsub.CiEventHandler, pubsubClient *pubsub2.PubSubClient, UserRouter user.UserRouter, cronBasedEventReceiver pubsub.CronBasedEventReceiver,
	ChartRefRouter ChartRefRouter, ConfigMapRouter ConfigMapRouter, AppStoreRouter appStore.AppStoreRouter, chartRepositoryRouter chartRepo.ChartRepositoryRouter,
	ReleaseMetricsRouter ReleaseMetricsRouter, deploymentGroupRouter DeploymentGroupRouter, batchOperationRouter BatchOperationRouter,
	chartGroupRouter ChartGroupRouter, testSuitRouter TestSuitRouter, imageScanRouter ImageScanRouter,
	policyRouter PolicyRouter, gitOpsConfigRouter GitOpsConfigRouter, dashboardRouter dashboard.DashboardRouter, attributesRouter AttributesRouter,
	commonRouter CommonRouter, grafanaRouter GrafanaRouter, ssoLoginRouter sso.SsoLoginRouter, telemetryRouter TelemetryRouter, telemetryWatcher telemetry.TelemetryEventClient, bulkUpdateRouter BulkUpdateRouter, webhookListenerRouter WebhookListenerRouter, appLabelsRouter AppLabelRouter,
	coreAppRouter CoreAppRouter, helmAppRouter client.HelmAppRouter, k8sApplicationRouter k8s.K8sApplicationRouter,
	pProfRouter PProfRouter, deploymentConfigRouter deployment.DeploymentConfigRouter) *MuxRouter {
	r := &MuxRouter{
		Router:                           mux.NewRouter(),
		HelmRouter:                       HelmRouter,
		PipelineConfigRouter:             PipelineConfigRouter,
		MigrateDbRouter:                  MigrateDbRouter,
		EnvironmentClusterMappingsRouter: EnvironmentClusterMappingsRouter,
		AppListingRouter:                 AppListingRouter,
		ClusterRouter:                    ClusterRouter,
		WebHookRouter:                    WebHookRouter,
		UserAuthRouter:                   UserAuthRouter,
		ApplicationRouter:                ApplicationRouter,
		CDRouter:                         CDRouter,
		ProjectManagementRouter:          ProjectManagementRouter,
		DockerRegRouter:                  DockerRegRouter,
		GitProviderRouter:                GitProviderRouter,
		GitHostRouter:                    GitHostRouter,
		NotificationRouter:               NotificationRouter,
		TeamRouter:                       TeamRouter,
		logger:                           logger,
		gitWebhookHandler:                gitWebhookHandler,
		workflowUpdateHandler:            workflowUpdateHandler,
		appUpdateHandler:                 appUpdateHandler,
		ciEventHandler:                   ciEventHandler,
		pubsubClient:                     pubsubClient,
		UserRouter:                       UserRouter,
		cronBasedEventReceiver:           cronBasedEventReceiver,
		ChartRefRouter:                   ChartRefRouter,
		ConfigMapRouter:                  ConfigMapRouter,
		AppStoreRouter:                   AppStoreRouter,
		ChartRepositoryRouter:            chartRepositoryRouter,
		ReleaseMetricsRouter:             ReleaseMetricsRouter,
		deploymentGroupRouter:            deploymentGroupRouter,
		batchOperationRouter:             batchOperationRouter,
		chartGroupRouter:                 chartGroupRouter,
		testSuitRouter:                   testSuitRouter,
		imageScanRouter:                  imageScanRouter,
		policyRouter:                     policyRouter,
		gitOpsConfigRouter:               gitOpsConfigRouter,
		attributesRouter:                 attributesRouter,
		dashboardRouter:                  dashboardRouter,
		commonRouter:                     commonRouter,
		grafanaRouter:                    grafanaRouter,
		ssoLoginRouter:                   ssoLoginRouter,
		telemetryRouter:                  telemetryRouter,
		telemetryWatcher:                 telemetryWatcher,
		bulkUpdateRouter:                 bulkUpdateRouter,
		WebhookListenerRouter:            webhookListenerRouter,
		appLabelsRouter:                  appLabelsRouter,
		coreAppRouter:                    coreAppRouter,
		helmAppRouter:                    helmAppRouter,
		k8sApplicationRouter:             k8sApplicationRouter,
		pProfRouter:                      pProfRouter,
		deploymentConfigRouter:           deploymentConfigRouter,
	}
	return r
}

func (r MuxRouter) Init() {

	r.Router.PathPrefix("/orchestrator/api/vi/pod/exec/ws").Handler(terminal.CreateAttachHandler("/orchestrator/api/vi/pod/exec/ws"))

	r.Router.StrictSlash(true)
	r.Router.Handle("/metrics", promhttp.Handler())
	//prometheus.MustRegister(pipeline.CiTriggerCounter)
	//prometheus.MustRegister(app.CdTriggerCounter)
	r.Router.Path("/health").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := common.Response{}
		response.Code = 200
		response.Result = "OK"
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			r.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})

	r.Router.Path("/orchestrator/version").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)
		response := common.Response{}
		response.Code = 200
		response.Result = util.GetDevtronVersion()
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			r.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})
	coreAppRouter := r.Router.PathPrefix("/orchestrator/core").Subrouter()
	r.coreAppRouter.initCoreAppRouter(coreAppRouter)

	pipelineConfigRouter := r.Router.PathPrefix("/orchestrator/app").Subrouter()
	r.PipelineConfigRouter.initPipelineConfigRouter(pipelineConfigRouter)
	r.AppListingRouter.initAppListingRouter(pipelineConfigRouter)
	r.HelmRouter.initHelmRouter(pipelineConfigRouter)
	r.appLabelsRouter.initLabelRouter(pipelineConfigRouter)

	migrateRouter := r.Router.PathPrefix("/orchestrator/migrate").Subrouter()
	r.MigrateDbRouter.InitMigrateDbRouter(migrateRouter)

	environmentClusterMappingsRouter := r.Router.PathPrefix("/orchestrator/env").Subrouter()
	r.EnvironmentClusterMappingsRouter.InitEnvironmentClusterMappingsRouter(environmentClusterMappingsRouter)

	clusterRouter := r.Router.PathPrefix("/orchestrator/cluster").Subrouter()
	r.ClusterRouter.InitClusterRouter(clusterRouter)

	webHookRouter := r.Router.PathPrefix("/orchestrator/webhook").Subrouter()
	r.WebHookRouter.intWebhookRouter(webHookRouter)

	applicationRouter := r.Router.PathPrefix("/orchestrator/api/v1/applications").Subrouter()
	r.ApplicationRouter.initApplicationRouter(applicationRouter)

	rootRouter := r.Router.PathPrefix("/orchestrator").Subrouter()
	r.UserAuthRouter.InitUserAuthRouter(rootRouter)

	projectManagementRouter := r.Router.PathPrefix("/orchestrator/project-management").Subrouter()
	r.ProjectManagementRouter.InitProjectManagementRouter(projectManagementRouter)

	gitRouter := r.Router.PathPrefix("/orchestrator/git").Subrouter()
	r.GitProviderRouter.InitGitProviderRouter(gitRouter)
	r.GitHostRouter.InitGitHostRouter(gitRouter)

	dockerRouter := r.Router.PathPrefix("/orchestrator/docker").Subrouter()
	r.DockerRegRouter.InitDockerRegRouter(dockerRouter)

	notificationRouter := r.Router.PathPrefix("/orchestrator/notification").Subrouter()
	r.NotificationRouter.InitNotificationRegRouter(notificationRouter)

	teamRouter := r.Router.PathPrefix("/orchestrator/team").Subrouter()
	r.TeamRouter.InitTeamRouter(teamRouter)

	userRouter := r.Router.PathPrefix("/orchestrator/user").Subrouter()
	r.UserRouter.InitUserRouter(userRouter)

	chartRefRouter := r.Router.PathPrefix("/orchestrator/chartref").Subrouter()
	r.ChartRefRouter.initChartRefRouter(chartRefRouter)

	configMapRouter := r.Router.PathPrefix("/orchestrator/config").Subrouter()
	r.ConfigMapRouter.initConfigMapRouter(configMapRouter)

	appStoreRouter := r.Router.PathPrefix("/orchestrator/app-store").Subrouter()
	r.AppStoreRouter.Init(appStoreRouter)

	chartRepoRouter := r.Router.PathPrefix("/orchestrator/chart-repo").Subrouter()
	r.ChartRepositoryRouter.Init(chartRepoRouter)

	deploymentMetricsRouter := r.Router.PathPrefix("/orchestrator/deployment-metrics").Subrouter()
	r.ReleaseMetricsRouter.initReleaseMetricsRouter(deploymentMetricsRouter)

	deploymentGroupRouter := r.Router.PathPrefix("/orchestrator/deployment-group").Subrouter()
	r.deploymentGroupRouter.initDeploymentGroupRouter(deploymentGroupRouter)

	r.batchOperationRouter.initBatchOperationRouter(rootRouter)

	chartGroupRouter := r.Router.PathPrefix("/orchestrator/chart-group").Subrouter()
	r.chartGroupRouter.initChartGroupRouter(chartGroupRouter)

	testSuitRouter := r.Router.PathPrefix("/orchestrator/test-report").Subrouter()
	r.testSuitRouter.InitTestSuitRouter(testSuitRouter)

	imageScanRouter := r.Router.PathPrefix("/orchestrator/security/scan").Subrouter()
	r.imageScanRouter.InitImageScanRouter(imageScanRouter)

	policyRouter := r.Router.PathPrefix("/orchestrator/security/policy").Subrouter()
	r.policyRouter.InitPolicyRouter(policyRouter)

	gitOpsRouter := r.Router.PathPrefix("/orchestrator/gitops").Subrouter()
	r.gitOpsConfigRouter.InitGitOpsConfigRouter(gitOpsRouter)

	attributeRouter := r.Router.PathPrefix("/orchestrator/attributes").Subrouter()
	r.attributesRouter.initAttributesRouter(attributeRouter)

	dashboardRouter := r.Router.PathPrefix("/dashboard").Subrouter()
	r.dashboardRouter.InitDashboardRouter(dashboardRouter)

	grafanaRouter := r.Router.PathPrefix("/grafana").Subrouter()
	r.grafanaRouter.initGrafanaRouter(grafanaRouter)

	r.Router.Path("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/dashboard", 301)
	})

	commonRouter := r.Router.PathPrefix("/orchestrator/global").Subrouter()
	r.commonRouter.InitCommonRouter(commonRouter)

	ssoLoginRouter := r.Router.PathPrefix("/orchestrator/sso").Subrouter()
	r.ssoLoginRouter.InitSsoLoginRouter(ssoLoginRouter)

	telemetryRouter := r.Router.PathPrefix("/orchestrator/telemetry").Subrouter()
	r.telemetryRouter.initTelemetryRouter(telemetryRouter)

	bulkUpdateRouter := r.Router.PathPrefix("/orchestrator/batch").Subrouter()
	r.bulkUpdateRouter.initBulkUpdateRouter(bulkUpdateRouter)

	webhookListenerRouter := r.Router.PathPrefix("/orchestrator/webhook/git").Subrouter()
	r.WebhookListenerRouter.InitWebhookListenerRouter(webhookListenerRouter)

	helmApp := r.Router.PathPrefix("/orchestrator/application").Subrouter()
	r.helmAppRouter.InitAppListRouter(helmApp)

	k8sApp := r.Router.PathPrefix("/orchestrator/k8s").Subrouter()
	r.k8sApplicationRouter.InitK8sApplicationRouter(k8sApp)

	pProfListenerRouter := r.Router.PathPrefix("/orchestrator/debug/pprof").Subrouter()
	r.pProfRouter.initPProfRouter(pProfListenerRouter)

	//  deployment router starts
	deploymentConfigSubRouter := r.Router.PathPrefix("/orchestrator/deployment/template").Subrouter()
	r.deploymentConfigRouter.Init(deploymentConfigSubRouter)
	// deployment router ends
}
