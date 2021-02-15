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
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/router/pubsub"
	pubsub2 "github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/terminal"
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
	ClusterAccountsRouter            ClusterAccountsRouter
	MigrateDbRouter                  MigrateDbRouter
	EnvironmentClusterMappingsRouter EnvironmentRouter
	AppListingRouter                 AppListingRouter
	ClusterRouter                    ClusterRouter
	ClusterHelmConfigRouter          ClusterHelmConfigRouter
	WebHookRouter                    WebhookRouter
	UserAuthRouter                   UserAuthRouter
	ApplicationRouter                ApplicationRouter
	CDRouter                         CDRouter
	ProjectManagementRouter          ProjectManagementRouter
	GitProviderRouter                GitProviderRouter
	DockerRegRouter                  DockerRegRouter
	NotificationRouter               NotificationRouter
	TeamRouter                       TeamRouter
	pubsubClient                     *pubsub2.PubSubClient
	UserRouter                       UserRouter
	gitWebhookHandler                pubsub.GitWebhookHandler
	workflowUpdateHandler            pubsub.WorkflowStatusUpdateHandler
	appUpdateHandler                 pubsub.ApplicationStatusUpdateHandler
	ciEventHandler                   pubsub.CiEventHandler
	cronBasedEventReceiver           pubsub.CronBasedEventReceiver
	ChartRefRouter                   ChartRefRouter
	ConfigMapRouter                  ConfigMapRouter
	AppStoreRouter                   AppStoreRouter
	ReleaseMetricsRouter             ReleaseMetricsRouter
	deploymentGroupRouter            DeploymentGroupRouter
	chartGroupRouter                 ChartGroupRouter
	batchOperationRouter             BatchOperationRouter
	testSuitRouter                   TestSuitRouter
	imageScanRouter                  ImageScanRouter
	policyRouter                     PolicyRouter
	gitOpsConfigRouter               GitOpsConfigRouter
	dashboardRouter                  DashboardRouter
}

func NewMuxRouter(logger *zap.SugaredLogger, HelmRouter HelmRouter, PipelineConfigRouter PipelineConfigRouter,
	MigrateDbRouter MigrateDbRouter, ClusterAccountsRouter ClusterAccountsRouter, AppListingRouter AppListingRouter,
	EnvironmentClusterMappingsRouter EnvironmentRouter, ClusterRouter ClusterRouter, ClusterHelmConfigRouter ClusterHelmConfigRouter,
	WebHookRouter WebhookRouter, UserAuthRouter UserAuthRouter, ApplicationRouter ApplicationRouter,
	CDRouter CDRouter, ProjectManagementRouter ProjectManagementRouter,
	GitProviderRouter GitProviderRouter, DockerRegRouter DockerRegRouter,
	NotificationRouter NotificationRouter,
	TeamRouter TeamRouter,
	gitWebhookHandler pubsub.GitWebhookHandler,
	workflowUpdateHandler pubsub.WorkflowStatusUpdateHandler,
	appUpdateHandler pubsub.ApplicationStatusUpdateHandler,
	ciEventHandler pubsub.CiEventHandler, pubsubClient *pubsub2.PubSubClient, UserRouter UserRouter, cronBasedEventReceiver pubsub.CronBasedEventReceiver,
	ChartRefRouter ChartRefRouter, ConfigMapRouter ConfigMapRouter, AppStoreRouter AppStoreRouter,
	ReleaseMetricsRouter ReleaseMetricsRouter, deploymentGroupRouter DeploymentGroupRouter, batchOperationRouter BatchOperationRouter,
	chartGroupRouter ChartGroupRouter, testSuitRouter TestSuitRouter, imageScanRouter ImageScanRouter,
	policyRouter PolicyRouter, gitOpsConfigRouter GitOpsConfigRouter, dashboardRouter DashboardRouter) *MuxRouter {
	r := &MuxRouter{
		Router:                           mux.NewRouter(),
		HelmRouter:                       HelmRouter,
		PipelineConfigRouter:             PipelineConfigRouter,
		ClusterAccountsRouter:            ClusterAccountsRouter,
		MigrateDbRouter:                  MigrateDbRouter,
		EnvironmentClusterMappingsRouter: EnvironmentClusterMappingsRouter,
		AppListingRouter:                 AppListingRouter,
		ClusterRouter:                    ClusterRouter,
		ClusterHelmConfigRouter:          ClusterHelmConfigRouter,
		WebHookRouter:                    WebHookRouter,
		UserAuthRouter:                   UserAuthRouter,
		ApplicationRouter:                ApplicationRouter,
		CDRouter:                         CDRouter,
		ProjectManagementRouter:          ProjectManagementRouter,
		DockerRegRouter:                  DockerRegRouter,
		GitProviderRouter:                GitProviderRouter,
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
		ReleaseMetricsRouter:             ReleaseMetricsRouter,
		deploymentGroupRouter:            deploymentGroupRouter,
		batchOperationRouter:             batchOperationRouter,
		chartGroupRouter:                 chartGroupRouter,
		testSuitRouter:                   testSuitRouter,
		imageScanRouter:                  imageScanRouter,
		policyRouter:                     policyRouter,
		gitOpsConfigRouter:               gitOpsConfigRouter,
		dashboardRouter:                  dashboardRouter,
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
		response := restHandler.Response{}
		response.Code = 200
		response.Result = "OK"
		b, err := json.Marshal(response)
		if err != nil {
			b = []byte("OK")
			r.logger.Errorw("Unexpected error in apiError", "err", err)
		}
		_, _ = writer.Write(b)
	})

	pipelineConfigRouter := r.Router.PathPrefix("/orchestrator/app").Subrouter()
	r.PipelineConfigRouter.initPipelineConfigRouter(pipelineConfigRouter)
	r.AppListingRouter.initAppListingRouter(pipelineConfigRouter)
	r.HelmRouter.initHelmRouter(pipelineConfigRouter)

	migrateRouter := r.Router.PathPrefix("/orchestrator/migrate").Subrouter()
	r.MigrateDbRouter.InitMigrateDbRouter(migrateRouter)

	accountRouter := r.Router.PathPrefix("/orchestrator/account").Subrouter()
	r.ClusterAccountsRouter.InitClusterAccountsRouter(accountRouter)

	environmentClusterMappingsRouter := r.Router.PathPrefix("/orchestrator/env").Subrouter()
	r.EnvironmentClusterMappingsRouter.InitEnvironmentClusterMappingsRouter(environmentClusterMappingsRouter)

	clusterRouter := r.Router.PathPrefix("/orchestrator/cluster").Subrouter()
	r.ClusterRouter.InitClusterRouter(clusterRouter)

	clusterHelmConfigRouter := r.Router.PathPrefix("/orchestrator/helm").Subrouter()
	r.ClusterHelmConfigRouter.InitClusterHelmConfigRouter(clusterHelmConfigRouter)

	webHookRouter := r.Router.PathPrefix("/orchestrator/webhook").Subrouter()
	r.WebHookRouter.intWebhookRouter(webHookRouter)

	applicationRouter := r.Router.PathPrefix("/orchestrator/api/v1/applications").Subrouter()
	r.ApplicationRouter.initApplicationRouter(applicationRouter)

	rootRouter := r.Router.PathPrefix("/orchestrator").Subrouter()
	r.UserAuthRouter.initUserAuthRouter(rootRouter)

	projectManagementRouter := r.Router.PathPrefix("/orchestrator/project-management").Subrouter()
	r.ProjectManagementRouter.InitProjectManagementRouter(projectManagementRouter)

	gitRouter := r.Router.PathPrefix("/orchestrator/git").Subrouter()
	r.GitProviderRouter.InitGitProviderRouter(gitRouter)

	dockerRouter := r.Router.PathPrefix("/orchestrator/docker").Subrouter()
	r.DockerRegRouter.InitDockerRegRouter(dockerRouter)

	notificationRouter := r.Router.PathPrefix("/orchestrator/notification").Subrouter()
	r.NotificationRouter.InitNotificationRegRouter(notificationRouter)

	teamRouter := r.Router.PathPrefix("/orchestrator/team").Subrouter()
	r.TeamRouter.InitTeamRouter(teamRouter)

	userRouter := r.Router.PathPrefix("/orchestrator/user").Subrouter()
	r.UserRouter.initUserRouter(userRouter)

	chartRefRouter := r.Router.PathPrefix("/orchestrator/chartref").Subrouter()
	r.ChartRefRouter.initChartRefRouter(chartRefRouter)

	configMapRouter := r.Router.PathPrefix("/orchestrator/config").Subrouter()
	r.ConfigMapRouter.initConfigMapRouter(configMapRouter)

	appStoreRouter := r.Router.PathPrefix("/orchestrator/app-store").Subrouter()
	r.AppStoreRouter.initAppStoreRouter(appStoreRouter)
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
	
	dashboardRouter := r.Router.PathPrefix("/dashboard").Subrouter()
	r.dashboardRouter.initDashboardRouter(dashboardRouter)
}
