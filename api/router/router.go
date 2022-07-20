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
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type MuxRouter struct {
	logger                             *zap.SugaredLogger
	Router                             *mux.Router
	HelmRouter                         HelmRouter
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
	UserRouter                         user.UserRouter
	gitWebhookHandler                  pubsub.GitWebhookHandler
	workflowUpdateHandler              pubsub.WorkflowStatusUpdateHandler
	appUpdateHandler                   pubsub.ApplicationStatusUpdateHandler
	ciEventHandler                     pubsub.CiEventHandler
	cronBasedEventReceiver             pubsub.CronBasedEventReceiver
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
	commonRouter                       CommonRouter
	grafanaRouter                      GrafanaRouter
	ssoLoginRouter                     sso.SsoLoginRouter
	telemetryRouter                    TelemetryRouter
	telemetryWatcher                   telemetry.TelemetryEventClient
	bulkUpdateRouter                   BulkUpdateRouter
	WebhookListenerRouter              WebhookListenerRouter
	appLabelsRouter                    AppLabelRouter
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
	helmApplicationStatusUpdateHandler cron.HelmApplicationStatusUpdateHandler
	k8sCapacityRouter                  k8s.K8sCapacityRouter
	webhookHelmRouter                  webhookHelm.WebhookHelmRouter
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
	ciEventHandler pubsub.CiEventHandler, UserRouter user.UserRouter, cronBasedEventReceiver pubsub.CronBasedEventReceiver,
	ChartRefRouter ChartRefRouter, ConfigMapRouter ConfigMapRouter, AppStoreRouter appStore.AppStoreRouter, chartRepositoryRouter chartRepo.ChartRepositoryRouter,
	ReleaseMetricsRouter ReleaseMetricsRouter, deploymentGroupRouter DeploymentGroupRouter, batchOperationRouter BatchOperationRouter,
	chartGroupRouter ChartGroupRouter, testSuitRouter TestSuitRouter, imageScanRouter ImageScanRouter,
	policyRouter PolicyRouter, gitOpsConfigRouter GitOpsConfigRouter, dashboardRouter dashboard.DashboardRouter, attributesRouter AttributesRouter,
	commonRouter CommonRouter, grafanaRouter GrafanaRouter, ssoLoginRouter sso.SsoLoginRouter, telemetryRouter TelemetryRouter, telemetryWatcher telemetry.TelemetryEventClient, bulkUpdateRouter BulkUpdateRouter, webhookListenerRouter WebhookListenerRouter, appLabelsRouter AppLabelRouter,
	coreAppRouter CoreAppRouter, helmAppRouter client.HelmAppRouter, k8sApplicationRouter k8s.K8sApplicationRouter,
	pProfRouter PProfRouter, deploymentConfigRouter deployment.DeploymentConfigRouter, dashboardTelemetryRouter dashboardEvent.DashboardTelemetryRouter,
	commonDeploymentRouter appStoreDeployment.CommonDeploymentRouter, externalLinkRouter externalLink.ExternalLinkRouter,
	globalPluginRouter GlobalPluginRouter, moduleRouter module.ModuleRouter,
	serverRouter server.ServerRouter, apiTokenRouter apiToken.ApiTokenRouter,
	helmApplicationStatusUpdateHandler cron.HelmApplicationStatusUpdateHandler, k8sCapacityRouter k8s.K8sCapacityRouter, webhookHelmRouter webhookHelm.WebhookHelmRouter) *MuxRouter {
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
		UserRouter:                         UserRouter,
		cronBasedEventReceiver:             cronBasedEventReceiver,
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
		dashboardRouter:                    dashboardRouter,
		commonRouter:                       commonRouter,
		grafanaRouter:                      grafanaRouter,
		ssoLoginRouter:                     ssoLoginRouter,
		telemetryRouter:                    telemetryRouter,
		telemetryWatcher:                   telemetryWatcher,
		bulkUpdateRouter:                   bulkUpdateRouter,
		WebhookListenerRouter:              webhookListenerRouter,
		appLabelsRouter:                    appLabelsRouter,
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
	}
	return r
}

func (r MuxRouter) Init() {

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
}
