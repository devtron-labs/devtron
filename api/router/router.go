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
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/router/pubsub"
	pubsub2 "github.com/devtron-labs/devtron/client/pubsub"
	"encoding/json"
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
	policyRouter PolicyRouter) *MuxRouter {
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
	}
	return r
}

func (r MuxRouter) Init() {
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

	pipelineConfigRouter := r.Router.PathPrefix("/app").Subrouter()
	r.PipelineConfigRouter.initPipelineConfigRouter(pipelineConfigRouter)
	r.AppListingRouter.initAppListingRouter(pipelineConfigRouter)
	r.HelmRouter.initHelmRouter(pipelineConfigRouter)

	migrateRouter := r.Router.PathPrefix("/migrate").Subrouter()
	r.MigrateDbRouter.InitMigrateDbRouter(migrateRouter)

	accountRouter := r.Router.PathPrefix("/account").Subrouter()
	r.ClusterAccountsRouter.InitClusterAccountsRouter(accountRouter)

	environmentClusterMappingsRouter := r.Router.PathPrefix("/env").Subrouter()
	r.EnvironmentClusterMappingsRouter.InitEnvironmentClusterMappingsRouter(environmentClusterMappingsRouter)

	clusterRouter := r.Router.PathPrefix("/cluster").Subrouter()
	r.ClusterRouter.InitClusterRouter(clusterRouter)

	clusterHelmConfigRouter := r.Router.PathPrefix("/helm").Subrouter()
	r.ClusterHelmConfigRouter.InitClusterHelmConfigRouter(clusterHelmConfigRouter)

	webHookRouter := r.Router.PathPrefix("/webhook").Subrouter()
	r.WebHookRouter.intWebhookRouter(webHookRouter)

	applicationRouter := r.Router.PathPrefix("/api/v1/applications").Subrouter()
	r.ApplicationRouter.initApplicationRouter(applicationRouter)

	rootRouter := r.Router.PathPrefix("").Subrouter()
	r.UserAuthRouter.initUserAuthRouter(rootRouter)

	projectManagementRouter := r.Router.PathPrefix("/project-management").Subrouter()
	r.ProjectManagementRouter.InitProjectManagementRouter(projectManagementRouter)

	gitRouter := r.Router.PathPrefix("/git").Subrouter()
	r.GitProviderRouter.InitGitProviderRouter(gitRouter)

	dockerRouter := r.Router.PathPrefix("/docker").Subrouter()
	r.DockerRegRouter.InitDockerRegRouter(dockerRouter)

	notificationRouter := r.Router.PathPrefix("/notification").Subrouter()
	r.NotificationRouter.InitNotificationRegRouter(notificationRouter)

	teamRouter := r.Router.PathPrefix("/team").Subrouter()
	r.TeamRouter.InitTeamRouter(teamRouter)

	userRouter := r.Router.PathPrefix("/user").Subrouter()
	r.UserRouter.initUserRouter(userRouter)

	chartRefRouter := r.Router.PathPrefix("/chartref").Subrouter()
	r.ChartRefRouter.initChartRefRouter(chartRefRouter)

	configMapRouter := r.Router.PathPrefix("/config").Subrouter()
	r.ConfigMapRouter.initConfigMapRouter(configMapRouter)

	appStoreRouter := r.Router.PathPrefix("/app-store").Subrouter()
	r.AppStoreRouter.initAppStoreRouter(appStoreRouter)
	deploymentMetricsRouter := r.Router.PathPrefix("/deployment-metrics").Subrouter()
	r.ReleaseMetricsRouter.initReleaseMetricsRouter(deploymentMetricsRouter)

	deploymentGroupRouter := r.Router.PathPrefix("/deployment-group").Subrouter()
	r.deploymentGroupRouter.initDeploymentGroupRouter(deploymentGroupRouter)

	r.batchOperationRouter.initBatchOperationRouter(rootRouter)

	chartGroupRouter := r.Router.PathPrefix("/chart-group").Subrouter()
	r.chartGroupRouter.initChartGroupRouter(chartGroupRouter)

	testSuitRouter := r.Router.PathPrefix("/test").Subrouter()
	r.testSuitRouter.InitTestSuitRouter(testSuitRouter)

	imageScanRouter := r.Router.PathPrefix("/security/scan").Subrouter()
	r.imageScanRouter.InitImageScanRouter(imageScanRouter)
	policyRouter := r.Router.PathPrefix("/security/policy").Subrouter()
	r.policyRouter.InitPolicyRouter(policyRouter)
}
