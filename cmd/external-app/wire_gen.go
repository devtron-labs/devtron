// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/devtron-labs/authenticator/apiToken"
	"github.com/devtron-labs/authenticator/client"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/devtron-labs/common-lib/cloud-provider-identifier"
	"github.com/devtron-labs/common-lib/utils/grpc"
	"github.com/devtron-labs/common-lib/utils/k8s"
	apiToken2 "github.com/devtron-labs/devtron/api/apiToken"
	chartProvider2 "github.com/devtron-labs/devtron/api/appStore/chartProvider"
	"github.com/devtron-labs/devtron/api/appStore/deployment"
	"github.com/devtron-labs/devtron/api/appStore/discover"
	"github.com/devtron-labs/devtron/api/appStore/values"
	argoApplication2 "github.com/devtron-labs/devtron/api/argoApplication"
	sso2 "github.com/devtron-labs/devtron/api/auth/sso"
	user2 "github.com/devtron-labs/devtron/api/auth/user"
	chartRepo2 "github.com/devtron-labs/devtron/api/chartRepo"
	cluster2 "github.com/devtron-labs/devtron/api/cluster"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/dashboardEvent"
	externalLink2 "github.com/devtron-labs/devtron/api/externalLink"
	client2 "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	application2 "github.com/devtron-labs/devtron/api/k8s/application"
	capacity2 "github.com/devtron-labs/devtron/api/k8s/capacity"
	module2 "github.com/devtron-labs/devtron/api/module"
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/devtron-labs/devtron/api/restHandler/app/appInfo"
	"github.com/devtron-labs/devtron/api/restHandler/app/appList"
	"github.com/devtron-labs/devtron/api/router"
	app3 "github.com/devtron-labs/devtron/api/router/app"
	appInfo2 "github.com/devtron-labs/devtron/api/router/app/appInfo"
	appList2 "github.com/devtron-labs/devtron/api/router/app/appList"
	server2 "github.com/devtron-labs/devtron/api/server"
	team2 "github.com/devtron-labs/devtron/api/team"
	terminal2 "github.com/devtron-labs/devtron/api/terminal"
	webhookHelm2 "github.com/devtron-labs/devtron/api/webhook/helm"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/client/dashboard"
	"github.com/devtron-labs/devtron/client/telemetry"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/appStatus"
	"github.com/devtron-labs/devtron/internal/sql/repository/deploymentConfig"
	repository5 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/security"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/apiToken"
	app2 "github.com/devtron-labs/devtron/pkg/app"
	repository7 "github.com/devtron-labs/devtron/pkg/appStore/chartGroup/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/chartProvider"
	"github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	service3 "github.com/devtron-labs/devtron/pkg/appStore/discover/service"
	repository4 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	service2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/service"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"github.com/devtron-labs/devtron/pkg/appStore/values/repository"
	service4 "github.com/devtron-labs/devtron/pkg/appStore/values/service"
	"github.com/devtron-labs/devtron/pkg/argoApplication"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/auth/authentication"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/sso"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/auth/user/repository"
	"github.com/devtron-labs/devtron/pkg/chartRepo"
	"github.com/devtron-labs/devtron/pkg/chartRepo/repository"
	"github.com/devtron-labs/devtron/pkg/cluster"
	repository2 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/clusterTerminalAccess"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/providerConfig"
	"github.com/devtron-labs/devtron/pkg/externalLink"
	"github.com/devtron-labs/devtron/pkg/genericNotes"
	repository6 "github.com/devtron-labs/devtron/pkg/genericNotes/repository"
	k8s2 "github.com/devtron-labs/devtron/pkg/k8s"
	"github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/k8s/capacity"
	"github.com/devtron-labs/devtron/pkg/k8s/informer"
	"github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs"
	repository8 "github.com/devtron-labs/devtron/pkg/kubernetesResourceAuditLogs/repository"
	"github.com/devtron-labs/devtron/pkg/module"
	"github.com/devtron-labs/devtron/pkg/module/repo"
	"github.com/devtron-labs/devtron/pkg/module/store"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/server"
	"github.com/devtron-labs/devtron/pkg/server/config"
	"github.com/devtron-labs/devtron/pkg/server/store"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/terminal"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"github.com/devtron-labs/devtron/pkg/webhook/helm"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/devtron-labs/devtron/util/cron"
	"github.com/devtron-labs/devtron/util/rbac"
)

// Injectors from wire.go:

func InitializeApp() (*App, error) {
	sqlConfig, err := sql.GetConfig()
	if err != nil {
		return nil, err
	}
	sugaredLogger, err := util.NewSugardLogger()
	if err != nil {
		return nil, err
	}
	db, err := sql.NewDbConnection(sqlConfig, sugaredLogger)
	if err != nil {
		return nil, err
	}
	runtimeConfig, err := client.GetRuntimeConfig()
	if err != nil {
		return nil, err
	}
	k8sClient, err := client.NewK8sClient(runtimeConfig)
	if err != nil {
		return nil, err
	}
	dexConfig, err := client.BuildDexConfig(k8sClient)
	if err != nil {
		return nil, err
	}
	settings, err := client.GetSettings(dexConfig)
	if err != nil {
		return nil, err
	}
	apiTokenSecretStore := apiTokenAuth.InitApiTokenSecretStore()
	sessionManager := middleware.NewSessionManager(settings, dexConfig, apiTokenSecretStore)
	validate, err := util.IntValidator()
	if err != nil {
		return nil, err
	}
	syncedEnforcer, err := casbin.Create()
	if err != nil {
		return nil, err
	}
	casbinSyncedEnforcer, err := casbin.CreateV2()
	if err != nil {
		return nil, err
	}
	enforcerImpl, err := casbin.NewEnforcerImpl(syncedEnforcer, casbinSyncedEnforcer, sessionManager, sugaredLogger)
	if err != nil {
		return nil, err
	}
	defaultAuthPolicyRepositoryImpl := repository.NewDefaultAuthPolicyRepositoryImpl(db, sugaredLogger)
	defaultAuthRoleRepositoryImpl := repository.NewDefaultAuthRoleRepositoryImpl(db, sugaredLogger)
	userAuthRepositoryImpl := repository.NewUserAuthRepositoryImpl(db, sugaredLogger, defaultAuthPolicyRepositoryImpl, defaultAuthRoleRepositoryImpl)
	userRepositoryImpl := repository.NewUserRepositoryImpl(db, sugaredLogger)
	roleGroupRepositoryImpl := repository.NewRoleGroupRepositoryImpl(db, sugaredLogger)
	rbacPolicyDataRepositoryImpl := repository.NewRbacPolicyDataRepositoryImpl(sugaredLogger, db)
	rbacRoleDataRepositoryImpl := repository.NewRbacRoleDataRepositoryImpl(sugaredLogger, db)
	rbacDataCacheFactoryImpl := repository.NewRbacDataCacheFactoryImpl(sugaredLogger, rbacPolicyDataRepositoryImpl, rbacRoleDataRepositoryImpl)
	userCommonServiceImpl, err := user.NewUserCommonServiceImpl(userAuthRepositoryImpl, sugaredLogger, userRepositoryImpl, roleGroupRepositoryImpl, sessionManager, rbacDataCacheFactoryImpl)
	if err != nil {
		return nil, err
	}
	userAuditRepositoryImpl := repository.NewUserAuditRepositoryImpl(db)
	userAuditServiceImpl := user.NewUserAuditServiceImpl(sugaredLogger, userAuditRepositoryImpl)
	userServiceImpl := user.NewUserServiceImpl(userAuthRepositoryImpl, sugaredLogger, userRepositoryImpl, roleGroupRepositoryImpl, sessionManager, userCommonServiceImpl, userAuditServiceImpl)
	ssoLoginRepositoryImpl := sso.NewSSOLoginRepositoryImpl(db, sugaredLogger)
	k8sRuntimeConfig, err := k8s.GetRuntimeConfig()
	if err != nil {
		return nil, err
	}
	k8sServiceImpl := k8s.NewK8sUtil(sugaredLogger, k8sRuntimeConfig)
	environmentVariables, err := util2.GetEnvironmentVariables()
	if err != nil {
		return nil, err
	}
	selfRegistrationRolesRepositoryImpl := repository.NewSelfRegistrationRolesRepositoryImpl(db, sugaredLogger)
	userSelfRegistrationServiceImpl := user.NewUserSelfRegistrationServiceImpl(sugaredLogger, selfRegistrationRolesRepositoryImpl, userServiceImpl)
	userAuthOidcHelperImpl, err := authentication.NewUserAuthOidcHelperImpl(sugaredLogger, userSelfRegistrationServiceImpl, dexConfig, settings, sessionManager)
	if err != nil {
		return nil, err
	}
	ssoLoginServiceImpl := sso.NewSSOLoginServiceImpl(sugaredLogger, ssoLoginRepositoryImpl, k8sServiceImpl, environmentVariables, userAuthOidcHelperImpl)
	ssoLoginRestHandlerImpl := sso2.NewSsoLoginRestHandlerImpl(validate, sugaredLogger, enforcerImpl, userServiceImpl, ssoLoginServiceImpl)
	ssoLoginRouterImpl := sso2.NewSsoLoginRouterImpl(ssoLoginRestHandlerImpl)
	teamRepositoryImpl := team.NewTeamRepositoryImpl(db)
	loginService := middleware.NewUserLogin(sessionManager, k8sClient)
	userAuthServiceImpl := user.NewUserAuthServiceImpl(userAuthRepositoryImpl, sessionManager, loginService, sugaredLogger, userRepositoryImpl, roleGroupRepositoryImpl, userServiceImpl)
	teamServiceImpl := team.NewTeamServiceImpl(sugaredLogger, teamRepositoryImpl, userAuthServiceImpl)
	clusterRepositoryImpl := repository2.NewClusterRepositoryImpl(db, sugaredLogger)
	v := informer.NewGlobalMapClusterNamespace()
	k8sInformerFactoryImpl := informer.NewK8sInformerFactoryImpl(sugaredLogger, v, k8sServiceImpl)
	clusterServiceImpl := cluster.NewClusterServiceImpl(clusterRepositoryImpl, sugaredLogger, k8sServiceImpl, k8sInformerFactoryImpl, userAuthRepositoryImpl, userRepositoryImpl, roleGroupRepositoryImpl)
	appStatusRepositoryImpl := appStatus.NewAppStatusRepositoryImpl(db, sugaredLogger)
	environmentRepositoryImpl := repository2.NewEnvironmentRepositoryImpl(db, sugaredLogger, appStatusRepositoryImpl)
	attributesRepositoryImpl := repository3.NewAttributesRepositoryImpl(db)
	environmentServiceImpl := cluster.NewEnvironmentServiceImpl(environmentRepositoryImpl, clusterServiceImpl, sugaredLogger, k8sServiceImpl, k8sInformerFactoryImpl, userAuthServiceImpl, attributesRepositoryImpl)
	chartRepoRepositoryImpl := chartRepoRepository.NewChartRepoRepositoryImpl(db)
	acdAuthConfig, err := util3.GetACDAuthConfig()
	if err != nil {
		return nil, err
	}
	httpClient := util.NewHttpClient()
	serverEnvConfigServerEnvConfig, err := serverEnvConfig.ParseServerEnvConfig()
	if err != nil {
		return nil, err
	}
	chartRepositoryServiceImpl := chartRepo.NewChartRepositoryServiceImpl(sugaredLogger, chartRepoRepositoryImpl, k8sServiceImpl, clusterServiceImpl, acdAuthConfig, httpClient, serverEnvConfigServerEnvConfig)
	installedAppRepositoryImpl := repository4.NewInstalledAppRepositoryImpl(sugaredLogger, db)
	helmClientConfig, err := gRPC.GetConfig()
	if err != nil {
		return nil, err
	}
	configuration, err := grpc.GetConfiguration()
	if err != nil {
		return nil, err
	}
	helmAppClientImpl := gRPC.NewHelmAppClientImpl(sugaredLogger, helmClientConfig, configuration)
	pumpImpl := connector.NewPumpImpl(sugaredLogger)
	appRepositoryImpl := app.NewAppRepositoryImpl(db, sugaredLogger)
	enforcerUtilHelmImpl := rbac.NewEnforcerUtilHelmImpl(sugaredLogger, clusterRepositoryImpl, teamRepositoryImpl, appRepositoryImpl, installedAppRepositoryImpl)
	serverDataStoreServerDataStore := serverDataStore.InitServerDataStore()
	appStoreApplicationVersionRepositoryImpl := appStoreDiscoverRepository.NewAppStoreApplicationVersionRepositoryImpl(sugaredLogger, db)
	pipelineRepositoryImpl := pipelineConfig.NewPipelineRepositoryImpl(db, sugaredLogger)
	helmReleaseConfig, err := service.GetHelmReleaseConfig()
	if err != nil {
		return nil, err
	}
	helmAppServiceImpl := service.NewHelmAppServiceImpl(sugaredLogger, clusterServiceImpl, helmAppClientImpl, pumpImpl, enforcerUtilHelmImpl, serverDataStoreServerDataStore, serverEnvConfigServerEnvConfig, appStoreApplicationVersionRepositoryImpl, environmentServiceImpl, pipelineRepositoryImpl, installedAppRepositoryImpl, appRepositoryImpl, clusterRepositoryImpl, k8sServiceImpl, helmReleaseConfig)
	dockerArtifactStoreRepositoryImpl := repository5.NewDockerArtifactStoreRepositoryImpl(db)
	dockerRegistryIpsConfigRepositoryImpl := repository5.NewDockerRegistryIpsConfigRepositoryImpl(db)
	ociRegistryConfigRepositoryImpl := repository5.NewOCIRegistryConfigRepositoryImpl(db)
	dockerRegistryConfigImpl := pipeline.NewDockerRegistryConfigImpl(sugaredLogger, helmAppServiceImpl, dockerArtifactStoreRepositoryImpl, dockerRegistryIpsConfigRepositoryImpl, ociRegistryConfigRepositoryImpl)
	deleteServiceImpl := delete2.NewDeleteServiceImpl(sugaredLogger, teamServiceImpl, clusterServiceImpl, environmentServiceImpl, chartRepositoryServiceImpl, installedAppRepositoryImpl, dockerRegistryConfigImpl, dockerArtifactStoreRepositoryImpl)
	teamRestHandlerImpl := team2.NewTeamRestHandlerImpl(sugaredLogger, teamServiceImpl, userServiceImpl, enforcerImpl, validate, userAuthServiceImpl, deleteServiceImpl)
	teamRouterImpl := team2.NewTeamRouterImpl(teamRestHandlerImpl)
	userAuthHandlerImpl := user2.NewUserAuthHandlerImpl(userAuthServiceImpl, validate, sugaredLogger, enforcerImpl)
	userAuthRouterImpl := user2.NewUserAuthRouterImpl(sugaredLogger, userAuthHandlerImpl, userAuthOidcHelperImpl)
	roleGroupServiceImpl := user.NewRoleGroupServiceImpl(userAuthRepositoryImpl, sugaredLogger, userRepositoryImpl, roleGroupRepositoryImpl, userCommonServiceImpl)
	userRestHandlerImpl := user2.NewUserRestHandlerImpl(userServiceImpl, validate, sugaredLogger, enforcerImpl, roleGroupServiceImpl, userCommonServiceImpl)
	userRouterImpl := user2.NewUserRouterImpl(userRestHandlerImpl)
	transactionUtilImpl := sql.NewTransactionUtilImpl(db)
	genericNoteRepositoryImpl := repository6.NewGenericNoteRepositoryImpl(db, transactionUtilImpl)
	genericNoteHistoryRepositoryImpl := repository6.NewGenericNoteHistoryRepositoryImpl(db, transactionUtilImpl)
	genericNoteHistoryServiceImpl := genericNotes.NewGenericNoteHistoryServiceImpl(genericNoteHistoryRepositoryImpl, sugaredLogger)
	genericNoteServiceImpl := genericNotes.NewGenericNoteServiceImpl(genericNoteRepositoryImpl, genericNoteHistoryServiceImpl, userRepositoryImpl, sugaredLogger)
	clusterDescriptionRepositoryImpl := repository2.NewClusterDescriptionRepositoryImpl(db, sugaredLogger)
	clusterDescriptionServiceImpl := cluster.NewClusterDescriptionServiceImpl(clusterDescriptionRepositoryImpl, userRepositoryImpl, sugaredLogger)
	helmUserServiceImpl, err := argo.NewHelmUserServiceImpl(sugaredLogger)
	if err != nil {
		return nil, err
	}
	clusterRbacServiceImpl := cluster.NewClusterRbacServiceImpl(environmentServiceImpl, enforcerImpl, clusterServiceImpl, sugaredLogger, userServiceImpl)
	clusterRestHandlerImpl := cluster2.NewClusterRestHandlerImpl(clusterServiceImpl, genericNoteServiceImpl, clusterDescriptionServiceImpl, sugaredLogger, userServiceImpl, validate, enforcerImpl, deleteServiceImpl, helmUserServiceImpl, environmentServiceImpl, clusterRbacServiceImpl)
	clusterRouterImpl := cluster2.NewClusterRouterImpl(clusterRestHandlerImpl)
	dashboardConfig, err := dashboard.GetConfig()
	if err != nil {
		return nil, err
	}
	dashboardRouterImpl, err := dashboard.NewDashboardRouterImpl(sugaredLogger, dashboardConfig)
	if err != nil {
		return nil, err
	}
	installedAppVersionHistoryRepositoryImpl := repository4.NewInstalledAppVersionHistoryRepositoryImpl(sugaredLogger, db)
	repositoryImpl := deploymentConfig.NewRepositoryImpl(db)
	chartRepositoryImpl := chartRepoRepository.NewChartRepository(db, transactionUtilImpl)
	deploymentConfigServiceImpl := common.NewDeploymentConfigServiceImpl(repositoryImpl, sugaredLogger, chartRepositoryImpl, pipelineRepositoryImpl, appRepositoryImpl, installedAppRepositoryImpl, environmentVariables)
	installedAppDBServiceImpl := EAMode.NewInstalledAppDBServiceImpl(sugaredLogger, installedAppRepositoryImpl, appRepositoryImpl, userServiceImpl, environmentServiceImpl, installedAppVersionHistoryRepositoryImpl, deploymentConfigServiceImpl)
	gitOpsConfigRepositoryImpl := repository3.NewGitOpsConfigRepositoryImpl(sugaredLogger, db)
	gitOpsConfigReadServiceImpl := config.NewGitOpsConfigReadServiceImpl(sugaredLogger, gitOpsConfigRepositoryImpl, userServiceImpl, environmentVariables)
	attributesServiceImpl := attributes.NewAttributesServiceImpl(sugaredLogger, attributesRepositoryImpl)
	deploymentTypeOverrideServiceImpl := providerConfig.NewDeploymentTypeOverrideServiceImpl(sugaredLogger, environmentVariables, attributesServiceImpl)
	eaModeDeploymentServiceImpl := EAMode.NewEAModeDeploymentServiceImpl(sugaredLogger, helmAppServiceImpl, appStoreApplicationVersionRepositoryImpl, helmAppClientImpl, installedAppRepositoryImpl, ociRegistryConfigRepositoryImpl)
	appStoreValidatorImpl := service2.NewAppAppStoreValidatorImpl(sugaredLogger)
	appStoreDeploymentDBServiceImpl := service2.NewAppStoreDeploymentDBServiceImpl(sugaredLogger, installedAppRepositoryImpl, appStoreApplicationVersionRepositoryImpl, appRepositoryImpl, environmentServiceImpl, clusterServiceImpl, installedAppVersionHistoryRepositoryImpl, environmentVariables, gitOpsConfigReadServiceImpl, deploymentTypeOverrideServiceImpl, eaModeDeploymentServiceImpl, appStoreValidatorImpl, installedAppDBServiceImpl, deploymentConfigServiceImpl)
	chartGroupDeploymentRepositoryImpl := repository7.NewChartGroupDeploymentRepositoryImpl(db, sugaredLogger)
	acdConfig, err := argocdServer.GetACDDeploymentConfig()
	if err != nil {
		return nil, err
	}
	deletePostProcessorImpl := service2.NewDeletePostProcessorImpl(sugaredLogger)
	appStoreDeploymentServiceImpl := service2.NewAppStoreDeploymentServiceImpl(sugaredLogger, installedAppRepositoryImpl, installedAppDBServiceImpl, appStoreDeploymentDBServiceImpl, chartGroupDeploymentRepositoryImpl, appStoreApplicationVersionRepositoryImpl, appRepositoryImpl, eaModeDeploymentServiceImpl, eaModeDeploymentServiceImpl, environmentServiceImpl, helmAppServiceImpl, installedAppVersionHistoryRepositoryImpl, environmentVariables, acdConfig, gitOpsConfigReadServiceImpl, deletePostProcessorImpl, appStoreValidatorImpl, deploymentConfigServiceImpl)
	helmAppRestHandlerImpl := client2.NewHelmAppRestHandlerImpl(sugaredLogger, helmAppServiceImpl, enforcerImpl, clusterServiceImpl, enforcerUtilHelmImpl, appStoreDeploymentServiceImpl, installedAppDBServiceImpl, userServiceImpl, attributesServiceImpl, serverEnvConfigServerEnvConfig)
	helmAppRouterImpl := client2.NewHelmAppRouterImpl(helmAppRestHandlerImpl)
	argoApplicationServiceImpl := argoApplication.NewArgoApplicationServiceImpl(sugaredLogger, clusterRepositoryImpl, k8sServiceImpl, helmUserServiceImpl, helmAppServiceImpl)
	k8sCommonServiceImpl := k8s2.NewK8sCommonServiceImpl(sugaredLogger, k8sServiceImpl, clusterServiceImpl, argoApplicationServiceImpl)
	environmentRestHandlerImpl := cluster2.NewEnvironmentRestHandlerImpl(environmentServiceImpl, sugaredLogger, userServiceImpl, validate, enforcerImpl, deleteServiceImpl, k8sServiceImpl, k8sCommonServiceImpl)
	environmentRouterImpl := cluster2.NewEnvironmentRouterImpl(environmentRestHandlerImpl)
	k8sResourceHistoryRepositoryImpl := repository8.NewK8sResourceHistoryRepositoryImpl(db, sugaredLogger)
	k8sResourceHistoryServiceImpl := kubernetesResourceAuditLogs.Newk8sResourceHistoryServiceImpl(k8sResourceHistoryRepositoryImpl, sugaredLogger, appRepositoryImpl, environmentRepositoryImpl)
	ephemeralContainersRepositoryImpl := repository2.NewEphemeralContainersRepositoryImpl(db, transactionUtilImpl)
	ephemeralContainerServiceImpl := cluster.NewEphemeralContainerServiceImpl(ephemeralContainersRepositoryImpl, sugaredLogger)
	terminalSessionHandlerImpl := terminal.NewTerminalSessionHandlerImpl(environmentServiceImpl, clusterServiceImpl, sugaredLogger, k8sServiceImpl, ephemeralContainerServiceImpl, argoApplicationServiceImpl)
	k8sApplicationServiceImpl, err := application.NewK8sApplicationServiceImpl(sugaredLogger, clusterServiceImpl, pumpImpl, helmAppServiceImpl, k8sServiceImpl, acdAuthConfig, k8sResourceHistoryServiceImpl, k8sCommonServiceImpl, terminalSessionHandlerImpl, ephemeralContainerServiceImpl, ephemeralContainersRepositoryImpl, argoApplicationServiceImpl)
	if err != nil {
		return nil, err
	}
	ciPipelineRepositoryImpl := pipelineConfig.NewCiPipelineRepositoryImpl(db, sugaredLogger, transactionUtilImpl)
	enforcerUtilImpl := rbac.NewEnforcerUtilImpl(sugaredLogger, teamRepositoryImpl, appRepositoryImpl, environmentRepositoryImpl, pipelineRepositoryImpl, ciPipelineRepositoryImpl, clusterRepositoryImpl, enforcerImpl)
	k8sApplicationRestHandlerImpl := application2.NewK8sApplicationRestHandlerImpl(sugaredLogger, k8sApplicationServiceImpl, pumpImpl, terminalSessionHandlerImpl, enforcerImpl, enforcerUtilHelmImpl, enforcerUtilImpl, helmAppServiceImpl, userServiceImpl, k8sCommonServiceImpl, validate, environmentVariables)
	k8sApplicationRouterImpl := application2.NewK8sApplicationRouterImpl(k8sApplicationRestHandlerImpl)
	chartRepositoryRestHandlerImpl := chartRepo2.NewChartRepositoryRestHandlerImpl(sugaredLogger, userServiceImpl, chartRepositoryServiceImpl, enforcerImpl, validate, deleteServiceImpl, attributesServiceImpl)
	chartRepositoryRouterImpl := chartRepo2.NewChartRepositoryRouterImpl(chartRepositoryRestHandlerImpl)
	appStoreServiceImpl := service3.NewAppStoreServiceImpl(sugaredLogger, appStoreApplicationVersionRepositoryImpl)
	appStoreRestHandlerImpl := appStoreDiscover.NewAppStoreRestHandlerImpl(sugaredLogger, userServiceImpl, appStoreServiceImpl, enforcerImpl)
	appStoreDiscoverRouterImpl := appStoreDiscover.NewAppStoreDiscoverRouterImpl(appStoreRestHandlerImpl)
	appStoreVersionValuesRepositoryImpl := appStoreValuesRepository.NewAppStoreVersionValuesRepositoryImpl(sugaredLogger, db)
	appStoreValuesServiceImpl := service4.NewAppStoreValuesServiceImpl(sugaredLogger, appStoreApplicationVersionRepositoryImpl, installedAppRepositoryImpl, appStoreVersionValuesRepositoryImpl, userServiceImpl)
	appStoreValuesRestHandlerImpl := appStoreValues.NewAppStoreValuesRestHandlerImpl(sugaredLogger, userServiceImpl, appStoreValuesServiceImpl)
	appStoreValuesRouterImpl := appStoreValues.NewAppStoreValuesRouterImpl(appStoreValuesRestHandlerImpl)
	appStoreDeploymentRestHandlerImpl := appStoreDeployment.NewAppStoreDeploymentRestHandlerImpl(sugaredLogger, userServiceImpl, enforcerImpl, enforcerUtilImpl, enforcerUtilHelmImpl, appStoreDeploymentServiceImpl, appStoreDeploymentDBServiceImpl, validate, helmAppServiceImpl, helmUserServiceImpl, attributesServiceImpl, installedAppDBServiceImpl)
	appStoreDeploymentRouterImpl := appStoreDeployment.NewAppStoreDeploymentRouterImpl(appStoreDeploymentRestHandlerImpl)
	chartProviderServiceImpl := chartProvider.NewChartProviderServiceImpl(sugaredLogger, chartRepoRepositoryImpl, chartRepositoryServiceImpl, dockerArtifactStoreRepositoryImpl, ociRegistryConfigRepositoryImpl)
	chartProviderRestHandlerImpl := chartProvider2.NewChartProviderRestHandlerImpl(sugaredLogger, userServiceImpl, validate, chartProviderServiceImpl, enforcerImpl)
	chartProviderRouterImpl := chartProvider2.NewChartProviderRouterImpl(chartProviderRestHandlerImpl)
	dockerRegRestHandlerImpl := restHandler.NewDockerRegRestHandlerImpl(dockerRegistryConfigImpl, sugaredLogger, chartProviderServiceImpl, userServiceImpl, validate, enforcerImpl, teamServiceImpl, deleteServiceImpl)
	dockerRegRouterImpl := router.NewDockerRegRouterImpl(dockerRegRestHandlerImpl)
	posthogClient, err := telemetry.NewPosthogClient(sugaredLogger)
	if err != nil {
		return nil, err
	}
	moduleRepositoryImpl := moduleRepo.NewModuleRepositoryImpl(db)
	providerIdentifierServiceImpl := providerIdentifier.NewProviderIdentifierServiceImpl(sugaredLogger)
	cronLoggerImpl := cron.NewCronLoggerImpl(sugaredLogger)
	telemetryEventClientImpl, err := telemetry.NewTelemetryEventClientImpl(sugaredLogger, httpClient, clusterServiceImpl, k8sServiceImpl, acdAuthConfig, userServiceImpl, attributesRepositoryImpl, ssoLoginServiceImpl, posthogClient, moduleRepositoryImpl, serverDataStoreServerDataStore, userAuditServiceImpl, helmAppClientImpl, installedAppRepositoryImpl, providerIdentifierServiceImpl, cronLoggerImpl)
	if err != nil {
		return nil, err
	}
	dashboardTelemetryRestHandlerImpl := dashboardEvent.NewDashboardTelemetryRestHandlerImpl(sugaredLogger, telemetryEventClientImpl)
	dashboardTelemetryRouterImpl := dashboardEvent.NewDashboardTelemetryRouterImpl(dashboardTelemetryRestHandlerImpl)
	commonDeploymentRestHandlerImpl := appStoreDeployment.NewCommonDeploymentRestHandlerImpl(sugaredLogger, userServiceImpl, enforcerImpl, enforcerUtilImpl, enforcerUtilHelmImpl, appStoreDeploymentServiceImpl, installedAppDBServiceImpl, validate, helmAppServiceImpl, helmAppRestHandlerImpl, helmUserServiceImpl)
	commonDeploymentRouterImpl := appStoreDeployment.NewCommonDeploymentRouterImpl(commonDeploymentRestHandlerImpl)
	externalLinkMonitoringToolRepositoryImpl := externalLink.NewExternalLinkMonitoringToolRepositoryImpl(db)
	externalLinkIdentifierMappingRepositoryImpl := externalLink.NewExternalLinkIdentifierMappingRepositoryImpl(db)
	externalLinkRepositoryImpl := externalLink.NewExternalLinkRepositoryImpl(db)
	externalLinkServiceImpl := externalLink.NewExternalLinkServiceImpl(sugaredLogger, externalLinkMonitoringToolRepositoryImpl, externalLinkIdentifierMappingRepositoryImpl, externalLinkRepositoryImpl)
	externalLinkRestHandlerImpl := externalLink2.NewExternalLinkRestHandlerImpl(sugaredLogger, externalLinkServiceImpl, userServiceImpl, enforcerImpl, enforcerUtilImpl)
	externalLinkRouterImpl := externalLink2.NewExternalLinkRouterImpl(externalLinkRestHandlerImpl)
	moduleActionAuditLogRepositoryImpl := module.NewModuleActionAuditLogRepositoryImpl(db)
	serverCacheServiceImpl, err := server.NewServerCacheServiceImpl(sugaredLogger, serverEnvConfigServerEnvConfig, serverDataStoreServerDataStore, helmAppServiceImpl)
	if err != nil {
		return nil, err
	}
	moduleEnvConfig, err := module.ParseModuleEnvConfig()
	if err != nil {
		return nil, err
	}
	moduleCacheServiceImpl, err := module.NewModuleCacheServiceImpl(sugaredLogger, k8sServiceImpl, moduleEnvConfig, serverEnvConfigServerEnvConfig, serverDataStoreServerDataStore, moduleRepositoryImpl, teamServiceImpl)
	if err != nil {
		return nil, err
	}
	moduleServiceHelperImpl := module.NewModuleServiceHelperImpl(serverEnvConfigServerEnvConfig)
	moduleResourceStatusRepositoryImpl := moduleRepo.NewModuleResourceStatusRepositoryImpl(db)
	moduleDataStoreModuleDataStore := moduleDataStore.InitModuleDataStore()
	moduleCronServiceImpl, err := module.NewModuleCronServiceImpl(sugaredLogger, moduleEnvConfig, moduleRepositoryImpl, serverEnvConfigServerEnvConfig, helmAppServiceImpl, moduleServiceHelperImpl, moduleResourceStatusRepositoryImpl, moduleDataStoreModuleDataStore, cronLoggerImpl)
	if err != nil {
		return nil, err
	}
	scanToolMetadataRepositoryImpl := security.NewScanToolMetadataRepositoryImpl(db, sugaredLogger)
	moduleServiceImpl := module.NewModuleServiceImpl(sugaredLogger, serverEnvConfigServerEnvConfig, moduleRepositoryImpl, moduleActionAuditLogRepositoryImpl, helmAppServiceImpl, serverDataStoreServerDataStore, serverCacheServiceImpl, moduleCacheServiceImpl, moduleCronServiceImpl, moduleServiceHelperImpl, moduleResourceStatusRepositoryImpl, scanToolMetadataRepositoryImpl)
	moduleRestHandlerImpl := module2.NewModuleRestHandlerImpl(sugaredLogger, moduleServiceImpl, userServiceImpl, enforcerImpl, validate)
	moduleRouterImpl := module2.NewModuleRouterImpl(moduleRestHandlerImpl)
	serverActionAuditLogRepositoryImpl := server.NewServerActionAuditLogRepositoryImpl(db)
	serverServiceImpl := server.NewServerServiceImpl(sugaredLogger, serverActionAuditLogRepositoryImpl, serverDataStoreServerDataStore, serverEnvConfigServerEnvConfig, helmAppServiceImpl, moduleRepositoryImpl)
	serverRestHandlerImpl := server2.NewServerRestHandlerImpl(sugaredLogger, serverServiceImpl, userServiceImpl, enforcerImpl, validate)
	serverRouterImpl := server2.NewServerRouterImpl(serverRestHandlerImpl)
	apiTokenSecretServiceImpl, err := apiToken.NewApiTokenSecretServiceImpl(sugaredLogger, attributesServiceImpl, apiTokenSecretStore)
	if err != nil {
		return nil, err
	}
	apiTokenRepositoryImpl := apiToken.NewApiTokenRepositoryImpl(db)
	apiTokenServiceImpl := apiToken.NewApiTokenServiceImpl(sugaredLogger, apiTokenSecretServiceImpl, userServiceImpl, userAuditServiceImpl, apiTokenRepositoryImpl)
	apiTokenRestHandlerImpl := apiToken2.NewApiTokenRestHandlerImpl(sugaredLogger, apiTokenServiceImpl, userServiceImpl, enforcerImpl, validate)
	apiTokenRouterImpl := apiToken2.NewApiTokenRouterImpl(apiTokenRestHandlerImpl)
	k8sCapacityServiceImpl := capacity.NewK8sCapacityServiceImpl(sugaredLogger, k8sApplicationServiceImpl, k8sServiceImpl, k8sCommonServiceImpl)
	k8sCapacityRestHandlerImpl := capacity2.NewK8sCapacityRestHandlerImpl(sugaredLogger, k8sCapacityServiceImpl, userServiceImpl, enforcerImpl, clusterServiceImpl, environmentServiceImpl, clusterRbacServiceImpl)
	k8sCapacityRouterImpl := capacity2.NewK8sCapacityRouterImpl(k8sCapacityRestHandlerImpl)
	webhookHelmServiceImpl := webhookHelm.NewWebhookHelmServiceImpl(sugaredLogger, helmAppServiceImpl, clusterServiceImpl, chartRepositoryServiceImpl, attributesServiceImpl)
	webhookHelmRestHandlerImpl := webhookHelm2.NewWebhookHelmRestHandlerImpl(sugaredLogger, webhookHelmServiceImpl, userServiceImpl, enforcerImpl, validate)
	webhookHelmRouterImpl := webhookHelm2.NewWebhookHelmRouterImpl(webhookHelmRestHandlerImpl)
	userAttributesRepositoryImpl := repository3.NewUserAttributesRepositoryImpl(db)
	userAttributesServiceImpl := attributes.NewUserAttributesServiceImpl(sugaredLogger, userAttributesRepositoryImpl)
	userAttributesRestHandlerImpl := restHandler.NewUserAttributesRestHandlerImpl(sugaredLogger, enforcerImpl, userServiceImpl, userAttributesServiceImpl)
	userAttributesRouterImpl := router.NewUserAttributesRouterImpl(userAttributesRestHandlerImpl)
	telemetryRestHandlerImpl := restHandler.NewTelemetryRestHandlerImpl(sugaredLogger, telemetryEventClientImpl, enforcerImpl, userServiceImpl)
	telemetryRouterImpl := router.NewTelemetryRouterImpl(sugaredLogger, telemetryRestHandlerImpl)
	terminalAccessRepositoryImpl := repository3.NewTerminalAccessRepositoryImpl(db, sugaredLogger)
	userTerminalSessionConfig, err := clusterTerminalAccess.GetTerminalAccessConfig()
	if err != nil {
		return nil, err
	}
	userTerminalAccessServiceImpl, err := clusterTerminalAccess.NewUserTerminalAccessServiceImpl(sugaredLogger, terminalAccessRepositoryImpl, userTerminalSessionConfig, k8sCommonServiceImpl, terminalSessionHandlerImpl, k8sCapacityServiceImpl, k8sServiceImpl, cronLoggerImpl)
	if err != nil {
		return nil, err
	}
	userTerminalAccessRestHandlerImpl := terminal2.NewUserTerminalAccessRestHandlerImpl(sugaredLogger, userTerminalAccessServiceImpl, enforcerImpl, userServiceImpl, validate)
	userTerminalAccessRouterImpl := terminal2.NewUserTerminalAccessRouterImpl(userTerminalAccessRestHandlerImpl)
	attributesRestHandlerImpl := restHandler.NewAttributesRestHandlerImpl(sugaredLogger, enforcerImpl, userServiceImpl, attributesServiceImpl)
	attributesRouterImpl := router.NewAttributesRouterImpl(attributesRestHandlerImpl)
	appLabelRepositoryImpl := pipelineConfig.NewAppLabelRepositoryImpl(db)
	materialRepositoryImpl := pipelineConfig.NewMaterialRepositoryImpl(db)
	crudOperationServiceConfig, err := app2.GetCrudOperationServiceConfig()
	if err != nil {
		return nil, err
	}
	appCrudOperationServiceImpl := app2.NewAppCrudOperationServiceImpl(appLabelRepositoryImpl, sugaredLogger, appRepositoryImpl, userRepositoryImpl, installedAppRepositoryImpl, genericNoteServiceImpl, materialRepositoryImpl, installedAppDBServiceImpl, crudOperationServiceConfig)
	appInfoRestHandlerImpl := appInfo.NewAppInfoRestHandlerImpl(sugaredLogger, appCrudOperationServiceImpl, userServiceImpl, validate, enforcerUtilImpl, enforcerImpl, helmAppServiceImpl, enforcerUtilHelmImpl, genericNoteServiceImpl)
	appInfoRouterImpl := appInfo2.NewAppInfoRouterImpl(sugaredLogger, appInfoRestHandlerImpl)
	appFilteringRestHandlerImpl := appList.NewAppFilteringRestHandlerImpl(sugaredLogger, teamServiceImpl, enforcerImpl, userServiceImpl, clusterServiceImpl, environmentServiceImpl)
	appFilteringRouterImpl := appList2.NewAppFilteringRouterImpl(appFilteringRestHandlerImpl)
	appRouterEAModeImpl := app3.NewAppRouterEAModeImpl(appInfoRouterImpl, appFilteringRouterImpl)
	rbacRoleServiceImpl := user.NewRbacRoleServiceImpl(sugaredLogger, rbacRoleDataRepositoryImpl)
	rbacRoleRestHandlerImpl := user2.NewRbacRoleHandlerImpl(sugaredLogger, validate, rbacRoleServiceImpl, userServiceImpl, enforcerImpl, enforcerUtilImpl)
	rbacRoleRouterImpl := user2.NewRbacRoleRouterImpl(sugaredLogger, validate, rbacRoleRestHandlerImpl)
	argoApplicationRestHandlerImpl := argoApplication2.NewArgoApplicationRestHandlerImpl(argoApplicationServiceImpl, sugaredLogger, enforcerImpl)
	argoApplicationRouterImpl := argoApplication2.NewArgoApplicationRouterImpl(argoApplicationRestHandlerImpl)
	muxRouter := NewMuxRouter(sugaredLogger, ssoLoginRouterImpl, teamRouterImpl, userAuthRouterImpl, userRouterImpl, clusterRouterImpl, dashboardRouterImpl, helmAppRouterImpl, environmentRouterImpl, k8sApplicationRouterImpl, chartRepositoryRouterImpl, appStoreDiscoverRouterImpl, appStoreValuesRouterImpl, appStoreDeploymentRouterImpl, chartProviderRouterImpl, dockerRegRouterImpl, dashboardTelemetryRouterImpl, commonDeploymentRouterImpl, externalLinkRouterImpl, moduleRouterImpl, serverRouterImpl, apiTokenRouterImpl, k8sCapacityRouterImpl, webhookHelmRouterImpl, userAttributesRouterImpl, telemetryRouterImpl, userTerminalAccessRouterImpl, attributesRouterImpl, appRouterEAModeImpl, rbacRoleRouterImpl, argoApplicationRouterImpl)
	mainApp := NewApp(db, sessionManager, muxRouter, telemetryEventClientImpl, posthogClient, sugaredLogger)
	return mainApp, nil
}
