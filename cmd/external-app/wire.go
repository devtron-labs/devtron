//go:build wireinject
// +build wireinject

/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"github.com/devtron-labs/authenticator/middleware"
	cloudProviderIdentifier "github.com/devtron-labs/common-lib/cloud-provider-identifier"
	posthogTelemetry "github.com/devtron-labs/common-lib/telemetry"
	util4 "github.com/devtron-labs/common-lib/utils/k8s"
	"github.com/devtron-labs/devtron/api/apiToken"
	"github.com/devtron-labs/devtron/api/appStore"
	chartProvider "github.com/devtron-labs/devtron/api/appStore/chartProvider"
	appStoreDeployment "github.com/devtron-labs/devtron/api/appStore/deployment"
	appStoreDiscover "github.com/devtron-labs/devtron/api/appStore/discover"
	appStoreValues "github.com/devtron-labs/devtron/api/appStore/values"
	"github.com/devtron-labs/devtron/api/argoApplication"
	"github.com/devtron-labs/devtron/api/auth/sso"
	"github.com/devtron-labs/devtron/api/auth/user"
	chartRepo "github.com/devtron-labs/devtron/api/chartRepo"
	"github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/dashboardEvent"
	"github.com/devtron-labs/devtron/api/externalLink"
	"github.com/devtron-labs/devtron/api/fluxApplication"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/k8s"
	"github.com/devtron-labs/devtron/api/module"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/restHandler/app/appInfo"
	appList2 "github.com/devtron-labs/devtron/api/restHandler/app/appList"
	"github.com/devtron-labs/devtron/api/router"
	app3 "github.com/devtron-labs/devtron/api/router/app"
	appInfo2 "github.com/devtron-labs/devtron/api/router/app/appInfo"
	"github.com/devtron-labs/devtron/api/router/app/appList"
	"github.com/devtron-labs/devtron/api/server"
	"github.com/devtron-labs/devtron/api/team"
	"github.com/devtron-labs/devtron/api/terminal"
	"github.com/devtron-labs/devtron/api/userResource"
	webhookHelm "github.com/devtron-labs/devtron/api/webhook/helm"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/client/argocdServer/config"
	"github.com/devtron-labs/devtron/client/argocdServer/repoCredsK8sClient"
	"github.com/devtron-labs/devtron/client/argocdServer/session"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/client/grafana"
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	app2 "github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/app/dbMigration"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/chartGroup/repository"
	deployment2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode/deployment"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode/deployment"
	"github.com/devtron-labs/devtron/pkg/asyncProvider"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/build/git/gitMaterial"
	"github.com/devtron-labs/devtron/pkg/commonService"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps"
	"github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/read"
	"github.com/devtron-labs/devtron/pkg/deployment/providerConfig"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	repository2 "github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs/repository"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool"
	security2 "github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/ucid"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/commonEnforcementFunctionsUtil"
	"github.com/devtron-labs/devtron/util/cron"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/google/wire"
)

func InitializeApp() (*App, error) {
	wire.Build(
		user.SelfRegistrationWireSet,
		sql.PgSqlWireSet,
		user.UserWireSet,
		sso.SsoConfigWireSet,
		AuthWireSet,
		util4.GetRuntimeConfig,
		util4.NewK8sUtil,
		externalLink.ExternalLinkWireSet,
		team.TeamsWireSet,
		cluster.ClusterWireSetEa,
		dashboard.DashboardWireSet,
		client.HelmAppWireSet,
		k8s.K8sApplicationWireSet,
		chartRepo.ChartRepositoryWireSet,
		appStoreDiscover.AppStoreDiscoverWireSet,
		chartProvider.AppStoreChartProviderWireSet,
		appStoreValues.WireSet,
		util3.GetEnvironmentVariables,
		appStoreDeployment.EAModeWireSet,
		server.ServerWireSet,
		module.ModuleWireSet,
		apiToken.ApiTokenWireSet,
		webhookHelm.WebhookHelmWireSet,
		terminal.TerminalWireSet,
		gitOps.GitOpsEAWireSet,
		providerConfig.DeploymentProviderConfigWireSet,
		argoApplication.ArgoApplicationWireSetEA,
		fluxApplication.FluxApplicationWireSet,
		userResource.UserResourceWireSetEA,
		NewApp,
		NewMuxRouter,
		util.NewHttpClient,
		util.NewSugardLogger,
		util.IntValidator,
		util2.GetACDAuthConfig,
		posthogTelemetry.NewPosthogClient,
		ucid.WireSet,
		delete2.NewDeleteServiceImpl,
		gitMaterial.GitMaterialWireSet,
		scanTool.ScanToolWireSet,

		sql.NewTransactionUtilImpl,

		// appStatus
		appStatus.NewAppStatusRepositoryImpl,
		wire.Bind(new(appStatus.AppStatusRepository), new(*appStatus.AppStatusRepositoryImpl)),
		// appStatus ends
		rbac.NewEnforcerUtilImpl,
		wire.Bind(new(rbac.EnforcerUtil), new(*rbac.EnforcerUtilImpl)),

		grafana.GetGrafanaClientConfig,
		grafana.NewGrafanaClientImpl,
		wire.Bind(new(grafana.GrafanaClient), new(*grafana.GrafanaClientImpl)),
		asyncProvider.WireSet,

		commonEnforcementFunctionsUtil.NewCommonEnforcementUtilImpl,
		wire.Bind(new(commonEnforcementFunctionsUtil.CommonEnforcementUtil), new(*commonEnforcementFunctionsUtil.CommonEnforcementUtilImpl)),

		appInfo2.NewAppInfoRouterImpl,
		wire.Bind(new(appInfo2.AppInfoRouter), new(*appInfo2.AppInfoRouterImpl)),
		appInfo.NewAppInfoRestHandlerImpl,
		wire.Bind(new(appInfo.AppInfoRestHandler), new(*appInfo.AppInfoRestHandlerImpl)),

		appList.NewAppFilteringRouterImpl,
		wire.Bind(new(appList.AppFilteringRouter), new(*appList.AppFilteringRouterImpl)),
		appList2.NewAppFilteringRestHandlerImpl,
		wire.Bind(new(appList2.AppFilteringRestHandler), new(*appList2.AppFilteringRestHandlerImpl)),

		app3.NewAppRouterEAModeImpl,
		wire.Bind(new(app3.AppRouterEAMode), new(*app3.AppRouterEAModeImpl)),

		app.NewAppCrudOperationServiceImpl,
		wire.Bind(new(app.AppCrudOperationService), new(*app.AppCrudOperationServiceImpl)),
		pipelineConfig.NewAppLabelRepositoryImpl,
		wire.Bind(new(pipelineConfig.AppLabelRepository), new(*pipelineConfig.AppLabelRepositoryImpl)),
		app.GetCrudOperationServiceConfig,
		// acd session client bind with authenticator login
		wire.Bind(new(session.ServiceClient), new(*middleware.LoginService)),
		connector.NewPumpImpl,
		wire.Bind(new(connector.Pump), new(*connector.PumpImpl)),
		cloudProviderIdentifier.NewProviderIdentifierServiceImpl,
		wire.Bind(new(cloudProviderIdentifier.ProviderIdentifierService), new(*cloudProviderIdentifier.ProviderIdentifierServiceImpl)),

		telemetry.NewTelemetryEventClientImpl,
		wire.Bind(new(telemetry.TelemetryEventClient), new(*telemetry.TelemetryEventClientImpl)),

		wire.Bind(new(delete2.DeleteService), new(*delete2.DeleteServiceImpl)),

		// needed for enforcer util
		pipelineConfig.NewPipelineRepositoryImpl,
		wire.Bind(new(pipelineConfig.PipelineRepository), new(*pipelineConfig.PipelineRepositoryImpl)),
		app2.NewAppRepositoryImpl,
		wire.Bind(new(app2.AppRepository), new(*app2.AppRepositoryImpl)),
		router.NewAttributesRouterImpl,
		wire.Bind(new(router.AttributesRouter), new(*router.AttributesRouterImpl)),
		restHandler.NewAttributesRestHandlerImpl,
		wire.Bind(new(restHandler.AttributesRestHandler), new(*restHandler.AttributesRestHandlerImpl)),
		attributes.NewAttributesServiceImpl,
		wire.Bind(new(attributes.AttributesService), new(*attributes.AttributesServiceImpl)),
		repository.NewAttributesRepositoryImpl,
		wire.Bind(new(repository.AttributesRepository), new(*repository.AttributesRepositoryImpl)),
		pipelineConfig.NewCiPipelineRepositoryImpl,
		wire.Bind(new(pipelineConfig.CiPipelineRepository), new(*pipelineConfig.CiPipelineRepositoryImpl)),
		// needed for enforcer util ends

		// binding gitops to helm (for hyperion)
		wire.Bind(new(deployment.FullModeDeploymentService), new(*deployment2.EAModeDeploymentServiceImpl)),

		wire.Bind(new(deployment.FullModeFluxDeploymentService), new(*deployment2.EAModeDeploymentServiceImpl)),

		router.NewTelemetryRouterImpl,
		wire.Bind(new(router.TelemetryRouter), new(*router.TelemetryRouterImpl)),
		restHandler.NewTelemetryRestHandlerImpl,
		wire.Bind(new(restHandler.TelemetryRestHandler), new(*restHandler.TelemetryRestHandlerImpl)),

		//needed for sending events
		dashboardEvent.NewDashboardTelemetryRestHandlerImpl,
		wire.Bind(new(dashboardEvent.DashboardTelemetryRestHandler), new(*dashboardEvent.DashboardTelemetryRestHandlerImpl)),
		dashboardEvent.NewDashboardTelemetryRouterImpl,
		wire.Bind(new(dashboardEvent.DashboardTelemetryRouter),
			new(*dashboardEvent.DashboardTelemetryRouterImpl)),

		//binding argoUserService to helm via dummy implementation(HelmUserServiceImpl)
		//argo.NewHelmUserServiceImpl,
		//wire.Bind(new(argo.ArgoUserService), new(*argo.HelmUserServiceImpl)),

		router.NewUserAttributesRouterImpl,
		wire.Bind(new(router.UserAttributesRouter), new(*router.UserAttributesRouterImpl)),
		restHandler.NewUserAttributesRestHandlerImpl,
		wire.Bind(new(restHandler.UserAttributesRestHandler), new(*restHandler.UserAttributesRestHandlerImpl)),
		attributes.NewUserAttributesServiceImpl,
		wire.Bind(new(attributes.UserAttributesService), new(*attributes.UserAttributesServiceImpl)),
		repository.NewUserAttributesRepositoryImpl,
		wire.Bind(new(repository.UserAttributesRepository), new(*repository.UserAttributesRepositoryImpl)),

		repository2.NewK8sResourceHistoryRepositoryImpl,
		wire.Bind(new(repository2.K8sResourceHistoryRepository), new(*repository2.K8sResourceHistoryRepositoryImpl)),

		kubernetesResourceAuditLogs.Newk8sResourceHistoryServiceImpl,
		wire.Bind(new(kubernetesResourceAuditLogs.K8sResourceHistoryService), new(*kubernetesResourceAuditLogs.K8sResourceHistoryServiceImpl)),

		security2.NewScanToolMetadataRepositoryImpl,
		wire.Bind(new(security2.ScanToolMetadataRepository), new(*security2.ScanToolMetadataRepositoryImpl)),

		// start: docker registry wire set injection
		router.NewDockerRegRouterImpl,
		wire.Bind(new(router.DockerRegRouter), new(*router.DockerRegRouterImpl)),
		restHandler.NewDockerRegRestHandlerImpl,
		wire.Bind(new(restHandler.DockerRegRestHandler), new(*restHandler.DockerRegRestHandlerImpl)),
		pipeline.NewDockerRegistryConfigImpl,
		wire.Bind(new(pipeline.DockerRegistryConfig), new(*pipeline.DockerRegistryConfigImpl)),
		dockerRegistryRepository.NewDockerArtifactStoreRepositoryImpl,
		wire.Bind(new(dockerRegistryRepository.DockerArtifactStoreRepository), new(*dockerRegistryRepository.DockerArtifactStoreRepositoryImpl)),
		dockerRegistryRepository.NewDockerRegistryIpsConfigRepositoryImpl,
		wire.Bind(new(dockerRegistryRepository.DockerRegistryIpsConfigRepository), new(*dockerRegistryRepository.DockerRegistryIpsConfigRepositoryImpl)),
		dockerRegistryRepository.NewOCIRegistryConfigRepositoryImpl,
		wire.Bind(new(dockerRegistryRepository.OCIRegistryConfigRepository), new(*dockerRegistryRepository.OCIRegistryConfigRepositoryImpl)),

		// chart group repository layer wire injection started
		repository4.NewChartGroupDeploymentRepositoryImpl,
		wire.Bind(new(repository4.ChartGroupDeploymentRepository), new(*repository4.ChartGroupDeploymentRepositoryImpl)),
		// chart group repository layer wire injection ended

		// end: docker registry wire set injection

		router.NewCommonRouterImpl,
		wire.Bind(new(router.CommonRouter), new(*router.CommonRouterImpl)),
		restHandler.NewCommonRestHandlerImpl,
		wire.Bind(new(restHandler.CommonRestHandler), new(*restHandler.CommonRestHandlerImpl)),

		commonService.NewCommonBaseServiceImpl,
		wire.Bind(new(commonService.CommonService), new(*commonService.CommonBaseServiceImpl)),

		cron.NewCronLoggerImpl,
		appStore.EAModeWireSet,

		common.WireSet,

		wire.Bind(new(util4.K8sService), new(*util4.K8sServiceImpl)),

		repoCredsK8sClient.NewRepositoryCredsK8sClientImpl,
		wire.Bind(new(repoCredsK8sClient.RepositoryCredsK8sClient), new(*repoCredsK8sClient.RepositoryCredsK8sClientImpl)),

		bean.GetConfig,

		config.NewArgoCDConfigGetter,
		wire.Bind(new(config.ArgoCDConfigGetter), new(*config.ArgoCDConfigGetterImpl)),

		argocdServer.NewArgoClientWrapperServiceEAImpl,
		wire.Bind(new(argocdServer.ArgoClientWrapperService), new(*argocdServer.ArgoClientWrapperServiceEAImpl)),

		dbMigration.NewDbMigrationServiceImpl,
		wire.Bind(new(dbMigration.DbMigration), new(*dbMigration.DbMigrationServiceImpl)),

		read.NewEnvConfigOverrideReadServiceImpl,
		wire.Bind(new(read.EnvConfigOverrideService), new(*read.EnvConfigOverrideReadServiceImpl)),

		chartConfig.NewEnvConfigOverrideRepository,
		wire.Bind(new(chartConfig.EnvConfigOverrideRepository), new(*chartConfig.EnvConfigOverrideRepositoryImpl)),
	)
	return &App{}, nil
}
