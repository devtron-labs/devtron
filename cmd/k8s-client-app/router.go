package main

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/api/terminal"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/k8s"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type MuxRouter struct {
	Router *mux.Router
	logger *zap.SugaredLogger
	//ssoLoginRouter           sso.SsoLoginRouter
	//teamRouter               team.TeamRouter
	//UserAuthRouter           user.UserAuthRouter
	//userRouter               user.UserRouter
	clusterRouter   cluster.ClusterRouter
	dashboardRouter dashboard.DashboardRouter
	//helmAppRouter            client.HelmAppRouter
	//environmentRouter        cluster.EnvironmentRouter
	k8sApplicationRouter k8s.K8sApplicationRouter
	//chartRepositoryRouter    chartRepo.ChartRepositoryRouter
	//appStoreDiscoverRouter   appStoreDiscover.AppStoreDiscoverRouter
	//appStoreValuesRouter     appStoreValues.AppStoreValuesRouter
	//appStoreDeploymentRouter appStoreDeployment.AppStoreDeploymentRouter
	//dashboardTelemetryRouter dashboardEvent.DashboardTelemetryRouter
	//commonDeploymentRouter   appStoreDeployment.CommonDeploymentRouter
	//externalLinksRouter      externalLink.ExternalLinkRouter
	//moduleRouter             module.ModuleRouter
	//serverRouter             server.ServerRouter
	//apiTokenRouter           apiToken.ApiTokenRouter
	k8sCapacityRouter k8s.K8sCapacityRouter
	//webhookHelmRouter        webhookHelm.WebhookHelmRouter
	//userAttributesRouter     router.UserAttributesRouter
	//telemetryRouter          router.TelemetryRouter
	userTerminalAccessRouter terminal.UserTerminalAccessRouter
	//attributesRouter         router.AttributesRouter
	//appRouter                router.AppRouter
}

func NewMuxRouter(
	logger *zap.SugaredLogger,
	//ssoLoginRouter sso.SsoLoginRouter,
	//teamRouter team.TeamRouter,
	//UserAuthRouter user.UserAuthRouter,
	//userRouter user.UserRouter,
	clusterRouter cluster.ClusterRouter,
	dashboardRouter dashboard.DashboardRouter,
	//helmAppRouter client.HelmAppRouter,
	//environmentRouter cluster.EnvironmentRouter,
	k8sApplicationRouter k8s.K8sApplicationRouter,
	//chartRepositoryRouter chartRepo.ChartRepositoryRouter,
	//appStoreDiscoverRouter appStoreDiscover.AppStoreDiscoverRouter,
	//appStoreValuesRouter appStoreValues.AppStoreValuesRouter,
	//appStoreDeploymentRouter appStoreDeployment.AppStoreDeploymentRouter,
	//dashboardTelemetryRouter dashboardEvent.DashboardTelemetryRouter,
	//commonDeploymentRouter appStoreDeployment.CommonDeploymentRouter,
	//externalLinkRouter externalLink.ExternalLinkRouter,
	//moduleRouter module.ModuleRouter,
	//serverRouter server.ServerRouter, apiTokenRouter apiToken.ApiTokenRouter,
	k8sCapacityRouter k8s.K8sCapacityRouter,
	// webhookHelmRouter webhookHelm.WebhookHelmRouter,
	// userAttributesRouter router.UserAttributesRouter,
	// telemetryRouter router.TelemetryRouter,
	userTerminalAccessRouter terminal.UserTerminalAccessRouter,
	// attributesRouter router.AttributesRouter,
	// appRouter router.AppRouter,
) *MuxRouter {
	r := &MuxRouter{
		Router: mux.NewRouter(),
		logger: logger,
		//ssoLoginRouter:           ssoLoginRouter,
		//teamRouter:               teamRouter,
		//UserAuthRouter:           UserAuthRouter,
		//userRouter:               userRouter,
		clusterRouter:   clusterRouter,
		dashboardRouter: dashboardRouter,
		//helmAppRouter:            helmAppRouter,
		//environmentRouter:        environmentRouter,
		k8sApplicationRouter: k8sApplicationRouter,
		//chartRepositoryRouter:    chartRepositoryRouter,
		//appStoreDiscoverRouter:   appStoreDiscoverRouter,
		//appStoreValuesRouter:     appStoreValuesRouter,
		//appStoreDeploymentRouter: appStoreDeploymentRouter,
		//dashboardTelemetryRouter: dashboardTelemetryRouter,
		//commonDeploymentRouter:   commonDeploymentRouter,
		//externalLinksRouter:      externalLinkRouter,
		//moduleRouter:             moduleRouter,
		//serverRouter:             serverRouter,
		//apiTokenRouter:           apiTokenRouter,
		k8sCapacityRouter: k8sCapacityRouter,
		//webhookHelmRouter:        webhookHelmRouter,
		//userAttributesRouter:     userAttributesRouter,
		//telemetryRouter:          telemetryRouter,
		userTerminalAccessRouter: userTerminalAccessRouter,
		//attributesRouter:         attributesRouter,
		//appRouter:                appRouter,
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

	//ssoLoginRouter := baseRouter.PathPrefix("/sso").Subrouter()
	//r.ssoLoginRouter.InitSsoLoginRouter(ssoLoginRouter)
	//teamRouter := baseRouter.PathPrefix("/team").Subrouter()
	//r.teamRouter.InitTeamRouter(teamRouter)
	//rootRouter := baseRouter.PathPrefix("/").Subrouter()
	//r.UserAuthRouter.InitUserAuthRouter(rootRouter)
	//userRouter := baseRouter.PathPrefix("/user").Subrouter()
	//r.userRouter.InitUserRouter(userRouter)

	clusterRouter := baseRouter.PathPrefix("/cluster").Subrouter()
	r.clusterRouter.InitClusterRouter(clusterRouter)

	//environmentClusterMappingsRouter := r.Router.PathPrefix("/orchestrator/env").Subrouter()
	//r.environmentRouter.InitEnvironmentClusterMappingsRouter(environmentClusterMappingsRouter)

	r.Router.Path("/").HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Redirect(writer, request, "/dashboard", 301)
	})

	//dashboardRouter := r.Router.PathPrefix("/dashboard").Subrouter()
	//r.dashboardRouter.InitDashboardRouter(dashboardRouter)

	//HelmApplicationSubRouter := r.Router.PathPrefix("/orchestrator/application").Subrouter()
	//r.helmAppRouter.InitAppListRouter(HelmApplicationSubRouter)
	//r.commonDeploymentRouter.Init(HelmApplicationSubRouter)

	//ApplicationSubRouter := r.Router.PathPrefix("/orchestrator/app").Subrouter()
	//r.appRouter.InitAppRouter(ApplicationSubRouter)

	k8sApp := r.Router.PathPrefix("/orchestrator/k8s").Subrouter()
	r.k8sApplicationRouter.InitK8sApplicationRouter(k8sApp)

	k8sCapacityApp := r.Router.PathPrefix("/orchestrator/k8s/capacity").Subrouter()
	r.k8sCapacityRouter.InitK8sCapacityRouter(k8sCapacityApp)

	// chart-repo router starts
	//chartRepoRouter := r.Router.PathPrefix("/orchestrator/chart-repo").Subrouter()
	//r.chartRepositoryRouter.Init(chartRepoRouter)
	//// chart-repo router ends
	//
	//// app-store discover router starts
	//appStoreDiscoverSubRouter := r.Router.PathPrefix("/orchestrator/app-store/discover").Subrouter()
	//r.appStoreDiscoverRouter.Init(appStoreDiscoverSubRouter)
	//// app-store discover router ends
	//
	////  app-store values starts
	//appStoreValuesSubRouter := r.Router.PathPrefix("/orchestrator/app-store/values").Subrouter()
	//r.appStoreValuesRouter.Init(appStoreValuesSubRouter)
	//// app-store values router ends
	//
	////  app-store deployment router starts
	//appStoreDeploymentSubRouter := r.Router.PathPrefix("/orchestrator/app-store/deployment").Subrouter()
	//r.appStoreDeploymentRouter.Init(appStoreDeploymentSubRouter)
	// app-store deployment router ends

	//  dashboard event router starts
	//dashboardTelemetryRouter := r.Router.PathPrefix("/orchestrator/dashboard-event").Subrouter()
	//r.dashboardTelemetryRouter.Init(dashboardTelemetryRouter)
	//// dashboard event router ends
	//
	//externalLinkRouter := r.Router.PathPrefix("/orchestrator/external-links").Subrouter()
	//r.externalLinksRouter.InitExternalLinkRouter(externalLinkRouter)
	//
	//// module router
	//moduleRouter := r.Router.PathPrefix("/orchestrator/module").Subrouter()
	//r.moduleRouter.Init(moduleRouter)
	//
	//// server router
	//serverRouter := r.Router.PathPrefix("/orchestrator/server").Subrouter()
	//r.serverRouter.Init(serverRouter)
	//
	//// api-token router
	//apiTokenRouter := r.Router.PathPrefix("/orchestrator/api-token").Subrouter()
	//r.apiTokenRouter.InitApiTokenRouter(apiTokenRouter)
	//
	//// webhook helm app router
	//webhookHelmRouter := r.Router.PathPrefix("/orchestrator/webhook/helm").Subrouter()
	//r.webhookHelmRouter.InitWebhookHelmRouter(webhookHelmRouter)
	//
	//userAttributeRouter := r.Router.PathPrefix("/orchestrator/attributes/user").Subrouter()
	//r.userAttributesRouter.InitUserAttributesRouter(userAttributeRouter)
	//
	//telemetryRouter := r.Router.PathPrefix("/orchestrator/telemetry").Subrouter()
	//r.telemetryRouter.InitTelemetryRouter(telemetryRouter)

	userTerminalAccessRouter := r.Router.PathPrefix("/orchestrator/user/terminal").Subrouter()
	r.userTerminalAccessRouter.InitTerminalAccessRouter(userTerminalAccessRouter)

	r.Router.PathPrefix("/orchestrator/attributes").Queries("key", "{key}").
		HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			vars := mux.Vars(request)
			key := vars["key"]
			if key == "DEFAULT_TERMINAL_IMAGE_LIST" {
				defaultAttrDto := &attributes.AttributesDto{
					Active: true,
					Key:    "DEFAULT_TERMINAL_IMAGE_LIST",
					Value:  "[{\"groupId\":\"latest\",\"groupRegex\":\"v1\\\\.2[4-8]\\\\..+\",\"imageList\":[{\"image\":\"quay.io/devtron/ubuntu-k8s-utils:latest\",\"name\":\"Ubuntu: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS\"},{\"image\":\"quay.io/devtron/alpine-k8s-utils:latest\",\"name\":\"Alpine: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS\"},{\"image\":\"quay.io/devtron/centos-k8s-utils:latest\",\"name\":\"CentOS: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS\"},{\"image\":\"quay.io/devtron/alpine-netshoot:latest\",\"name\":\"Alpine: Netshoot\",\"description\":\"Contains Docker + Kubernetes network troubleshooting utilities.\"}]},{\"groupId\":\"v1.22\",\"groupRegex\":\"v1\\\\.(21|22|23)\\\\..+\",\"imageList\":[{\"image\":\"quay.io/devtron/ubuntu-k8s-utils:1.22\",\"name\":\"Ubuntu: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS\"},{\"image\":\"quay.io/devtron/alpine-k8s-utils:1.22\",\"name\":\"Alpine: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS\"},{\"image\":\"quay.io/devtron/centos-k8s-utils:1.22\",\"name\":\"CentOS: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS\"},{\"image\":\"quay.io/devtron/alpine-netshoot:latest\",\"name\":\"Alpine: Netshoot\",\"description\":\"Contains Docker + Kubernetes network troubleshooting utilities.\"}]},{\"groupId\":\"v1.19\",\"groupRegex\":\"v1\\\\.(18|19|20)\\\\..+\",\"imageList\":[{\"image\":\"quay.io/devtron/ubuntu-k8s-utils:1.19\",\"name\":\"Ubuntu: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS\"},{\"image\":\"quay.io/devtron/alpine-k8s-utils:1.19\",\"name\":\"Alpine: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS\"},{\"image\":\"quay.io/devtron/centos-k8s-utils:1.19\",\"name\":\"CentOS: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS\"},{\"image\":\"quay.io/devtron/alpine-netshoot:latest\",\"name\":\"Alpine: Netshoot\",\"description\":\"Contains Docker + Kubernetes network troubleshooting utilities.\"}]},{\"groupId\":\"v1.16\",\"groupRegex\":\"v1\\\\.(15|16|17)\\\\..+\",\"imageList\":[{\"image\":\"quay.io/devtron/ubuntu-k8s-utils:1.16\",\"name\":\"Ubuntu: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on ubuntu OS\"},{\"image\":\"quay.io/devtron/alpine-k8s-utils:1.16\",\"name\":\"Alpine: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on alpine OS\"},{\"image\":\"quay.io/devtron/centos-k8s-utils:1.16\",\"name\":\"CentOS: Kubernetes utilites\",\"description\":\"Contains kubectl, helm, curl, git, busybox, wget, jq, nslookup, telnet on Cent OS\"},{\"image\":\"quay.io/devtron/alpine-netshoot:latest\",\"name\":\"Alpine: Netshoot\",\"description\":\"Contains Docker + Kubernetes network troubleshooting utilities.\"}]}]",
				}
				common.WriteJsonResp(writer, nil, defaultAttrDto, http.StatusOK)
			}
		}).Methods("GET")
	//attributeRouter := r.Router.PathPrefix("/orchestrator/attributes").Subrouter()
	//r.attributesRouter.InitAttributesRouter(attributeRouter)
}
