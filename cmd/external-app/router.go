package main

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/apiToken"
	appStoreDeployment "github.com/devtron-labs/devtron/api/appStore/deployment"
	appStoreDiscover "github.com/devtron-labs/devtron/api/appStore/discover"
	appStoreValues "github.com/devtron-labs/devtron/api/appStore/values"
	"github.com/devtron-labs/devtron/api/chartRepo"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/dashboardEvent"
	"github.com/devtron-labs/devtron/api/externalLink"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/module"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/api/router"
	"github.com/devtron-labs/devtron/api/server"
	"github.com/devtron-labs/devtron/api/sso"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/user"
	webhookHelm "github.com/devtron-labs/devtron/api/webhook/helm"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type MuxRouter struct {
	Router                   *mux.Router
	logger                   *zap.SugaredLogger
	ssoLoginRouter           sso.SsoLoginRouter
	teamRouter               team.TeamRouter
	UserAuthRouter           user.UserAuthRouter
	userRouter               user.UserRouter
	clusterRouter            cluster.ClusterRouter
	dashboardRouter          dashboard.DashboardRouter
	helmAppRouter            client.HelmAppRouter
	environmentRouter        cluster.EnvironmentRouter
	k8sApplicationRouter     k8s.K8sApplicationRouter
	chartRepositoryRouter    chartRepo.ChartRepositoryRouter
	appStoreDiscoverRouter   appStoreDiscover.AppStoreDiscoverRouter
	appStoreValuesRouter     appStoreValues.AppStoreValuesRouter
	appStoreDeploymentRouter appStoreDeployment.AppStoreDeploymentRouter
	dashboardTelemetryRouter dashboardEvent.DashboardTelemetryRouter
	commonDeploymentRouter   appStoreDeployment.CommonDeploymentRouter
	externalLinksRouter      externalLink.ExternalLinkRouter
	moduleRouter             module.ModuleRouter
	serverRouter             server.ServerRouter
	apiTokenRouter           apiToken.ApiTokenRouter
	k8sCapacityRouter        k8s.K8sCapacityRouter
	webhookHelmRouter        webhookHelm.WebhookHelmRouter
	userAttributesRouter     router.UserAttributesRouter
	telemetryRouter          router.TelemetryRouter
}

func NewMuxRouter(
	logger *zap.SugaredLogger,
	ssoLoginRouter sso.SsoLoginRouter,
	teamRouter team.TeamRouter,
	UserAuthRouter user.UserAuthRouter,
	userRouter user.UserRouter,
	clusterRouter cluster.ClusterRouter,
	dashboardRouter dashboard.DashboardRouter,
	helmAppRouter client.HelmAppRouter,
	environmentRouter cluster.EnvironmentRouter,
	k8sApplicationRouter k8s.K8sApplicationRouter,
	chartRepositoryRouter chartRepo.ChartRepositoryRouter,
	appStoreDiscoverRouter appStoreDiscover.AppStoreDiscoverRouter,
	appStoreValuesRouter appStoreValues.AppStoreValuesRouter,
	appStoreDeploymentRouter appStoreDeployment.AppStoreDeploymentRouter,
	dashboardTelemetryRouter dashboardEvent.DashboardTelemetryRouter,
	commonDeploymentRouter appStoreDeployment.CommonDeploymentRouter,
	externalLinkRouter externalLink.ExternalLinkRouter,
	moduleRouter module.ModuleRouter,
	serverRouter server.ServerRouter, apiTokenRouter apiToken.ApiTokenRouter,
	k8sCapacityRouter k8s.K8sCapacityRouter,
	webhookHelmRouter webhookHelm.WebhookHelmRouter,
	userAttributesRouter router.UserAttributesRouter,
	telemetryRouter router.TelemetryRouter,
) *MuxRouter {
	r := &MuxRouter{
		Router:                   mux.NewRouter(),
		logger:                   logger,
		ssoLoginRouter:           ssoLoginRouter,
		teamRouter:               teamRouter,
		UserAuthRouter:           UserAuthRouter,
		userRouter:               userRouter,
		clusterRouter:            clusterRouter,
		dashboardRouter:          dashboardRouter,
		helmAppRouter:            helmAppRouter,
		environmentRouter:        environmentRouter,
		k8sApplicationRouter:     k8sApplicationRouter,
		chartRepositoryRouter:    chartRepositoryRouter,
		appStoreDiscoverRouter:   appStoreDiscoverRouter,
		appStoreValuesRouter:     appStoreValuesRouter,
		appStoreDeploymentRouter: appStoreDeploymentRouter,
		dashboardTelemetryRouter: dashboardTelemetryRouter,
		commonDeploymentRouter:   commonDeploymentRouter,
		externalLinksRouter:      externalLinkRouter,
		moduleRouter:             moduleRouter,
		serverRouter:             serverRouter,
		apiTokenRouter:           apiTokenRouter,
		k8sCapacityRouter:        k8sCapacityRouter,
		webhookHelmRouter:        webhookHelmRouter,
		userAttributesRouter:     userAttributesRouter,
		telemetryRouter:          telemetryRouter,
	}
	return r
}
func (r *MuxRouter) Init() {
	r.Router.StrictSlash(true)
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
	baseRouter := r.Router.PathPrefix("/orchestrator/").Subrouter()
	baseRouter.Path("/version").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
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

	ssoLoginRouter := baseRouter.PathPrefix("/sso").Subrouter()
	r.ssoLoginRouter.InitSsoLoginRouter(ssoLoginRouter)
	teamRouter := baseRouter.PathPrefix("/team").Subrouter()
	r.teamRouter.InitTeamRouter(teamRouter)
	rootRouter := baseRouter.PathPrefix("/").Subrouter()
	r.UserAuthRouter.InitUserAuthRouter(rootRouter)
	userRouter := baseRouter.PathPrefix("/user").Subrouter()
	r.userRouter.InitUserRouter(userRouter)

	clusterRouter := baseRouter.PathPrefix("/cluster").Subrouter()
	r.clusterRouter.InitClusterRouter(clusterRouter)

	environmentClusterMappingsRouter := r.Router.PathPrefix("/orchestrator/env").Subrouter()
	r.environmentRouter.InitEnvironmentClusterMappingsRouter(environmentClusterMappingsRouter)

	r.Router.Path("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/dashboard", 301)
	})
	dashboardRouter := r.Router.PathPrefix("/dashboard").Subrouter()
	r.dashboardRouter.InitDashboardRouter(dashboardRouter)

	applicationSubRouter := r.Router.PathPrefix("/orchestrator/application").Subrouter()
	r.helmAppRouter.InitAppListRouter(applicationSubRouter)
	r.commonDeploymentRouter.Init(applicationSubRouter)

	k8sApp := r.Router.PathPrefix("/orchestrator/k8s").Subrouter()
	r.k8sApplicationRouter.InitK8sApplicationRouter(k8sApp)

	k8sCapacityApp := r.Router.PathPrefix("/orchestrator/k8s/capacity").Subrouter()
	r.k8sCapacityRouter.InitK8sCapacityRouter(k8sCapacityApp)

	// chart-repo router starts
	chartRepoRouter := r.Router.PathPrefix("/orchestrator/chart-repo").Subrouter()
	r.chartRepositoryRouter.Init(chartRepoRouter)
	// chart-repo router ends

	// app-store discover router starts
	appStoreDiscoverSubRouter := r.Router.PathPrefix("/orchestrator/app-store/discover").Subrouter()
	r.appStoreDiscoverRouter.Init(appStoreDiscoverSubRouter)
	// app-store discover router ends

	//  app-store values starts
	appStoreValuesSubRouter := r.Router.PathPrefix("/orchestrator/app-store/values").Subrouter()
	r.appStoreValuesRouter.Init(appStoreValuesSubRouter)
	// app-store values router ends

	//  app-store deployment router starts
	appStoreDeploymentSubRouter := r.Router.PathPrefix("/orchestrator/app-store/deployment").Subrouter()
	r.appStoreDeploymentRouter.Init(appStoreDeploymentSubRouter)
	// app-store deployment router ends

	//  dashboard event router starts
	dashboardTelemetryRouter := r.Router.PathPrefix("/orchestrator/dashboard-event").Subrouter()
	r.dashboardTelemetryRouter.Init(dashboardTelemetryRouter)
	// dashboard event router ends

	externalLinkRouter := r.Router.PathPrefix("/orchestrator/external-links").Subrouter()
	r.externalLinksRouter.InitExternalLinkRouter(externalLinkRouter)

	// module router
	moduleRouter := r.Router.PathPrefix("/orchestrator/module").Subrouter()
	r.moduleRouter.Init(moduleRouter)

	// server router
	serverRouter := r.Router.PathPrefix("/orchestrator/server").Subrouter()
	r.serverRouter.Init(serverRouter)

	// api-token router
	apiTokenRouter := r.Router.PathPrefix("/orchestrator/api-token").Subrouter()
	r.apiTokenRouter.InitApiTokenRouter(apiTokenRouter)

	// webhook helm app router
	webhookHelmRouter := r.Router.PathPrefix("/orchestrator/webhook/helm").Subrouter()
	r.webhookHelmRouter.InitWebhookHelmRouter(webhookHelmRouter)

	userAttributeRouter := r.Router.PathPrefix("/orchestrator/attributes/user").Subrouter()
	r.userAttributesRouter.InitUserAttributesRouter(userAttributeRouter)

	telemetryRouter := r.Router.PathPrefix("/orchestrator/telemetry").Subrouter()
	r.telemetryRouter.InitTelemetryRouter(telemetryRouter)
}
