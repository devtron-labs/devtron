package appStoreDeploymentTool

import (
	"context"
	"errors"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"net/http"
	"time"

	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

type AppStoreDeploymentHelmService interface {
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	GetAppStatus(installedAppAndEnvDetails repository.InstalledAppAndEnvDetails, w http.ResponseWriter, r *http.Request, token string) (string, error)
	DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error
	RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, bool, error)
	GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*client.HelmAppDeploymentHistory, error)
	GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error)
	GetGitOpsRepoName(appName string, environmentName string) (string, error)
	OnUpdateRepoInInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	UpdateRequirementDependencies(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error
	UpdateInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *clusterRepository.Environment, installedAppVersion *repository.InstalledAppVersions, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	DeleteDeploymentApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error
	UpdateInstalledAppAndPipelineStatusForFailedDeploymentStatus(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error
	SaveTimelineForACDHelmApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, status string, statusDetail string, tx *pg.Tx) error
	UpdateChartInfo(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *util.ChartGitAttribute, installedAppVersionHistoryId int, ctx context.Context) error
}

type AppStoreDeploymentHelmServiceImpl struct {
	Logger                               *zap.SugaredLogger
	helmAppService                       client.HelmAppService
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	environmentRepository                clusterRepository.EnvironmentRepository
	helmAppClient                        client.HelmAppClient
	installedAppRepository               repository.InstalledAppRepository
	appStoreDeploymentCommonService      appStoreDeploymentCommon.AppStoreDeploymentCommonService
	OCIRegistryConfigRepository          repository2.OCIRegistryConfigRepository
}

func NewAppStoreDeploymentHelmServiceImpl(logger *zap.SugaredLogger, helmAppService client.HelmAppService, appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	environmentRepository clusterRepository.EnvironmentRepository, helmAppClient client.HelmAppClient, installedAppRepository repository.InstalledAppRepository, appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService, OCIRegistryConfigRepository repository2.OCIRegistryConfigRepository) *AppStoreDeploymentHelmServiceImpl {
	return &AppStoreDeploymentHelmServiceImpl{
		Logger:                               logger,
		helmAppService:                       helmAppService,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		environmentRepository:                environmentRepository,
		helmAppClient:                        helmAppClient,
		installedAppRepository:               installedAppRepository,
		appStoreDeploymentCommonService:      appStoreDeploymentCommonService,
		OCIRegistryConfigRepository:          OCIRegistryConfigRepository,
	}
}

func (impl AppStoreDeploymentHelmServiceImpl) UpdateChartInfo(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *util.ChartGitAttribute, installedAppVersionHistoryId int, ctx context.Context) error {
	err := impl.updateApplicationWithChartInfo(ctx, installAppVersionRequest.InstalledAppId, installAppVersionRequest.AppStoreVersion, installAppVersionRequest.ValuesOverrideYaml, installedAppVersionHistoryId)
	if err != nil {
		impl.Logger.Errorw("error in updating helm app", "err", err)
		return err
	}
	return nil
}

func (impl AppStoreDeploymentHelmServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.Logger.Errorw("fetching error", "err", err)
		return installAppVersionRequest, err
	}
	var IsOCIRepo bool
	var registryCredential *client.RegistryCredential
	var chartRepository *client.ChartRepository
	dockerRegistryId := appStoreAppVersion.AppStore.DockerArtifactStoreId
	if dockerRegistryId != "" {
		ociRegistryConfigs, err := impl.OCIRegistryConfigRepository.FindByDockerRegistryId(dockerRegistryId)
		if err != nil {
			impl.Logger.Errorw("error in fetching oci registry config", "err", err)
			return nil, err
		}
		var ociRegistryConfig *repository2.OCIRegistryConfig
		for _, config := range ociRegistryConfigs {
			if config.RepositoryAction == repository2.STORAGE_ACTION_TYPE_PULL || config.RepositoryAction == repository2.STORAGE_ACTION_TYPE_PULL_AND_PUSH {
				ociRegistryConfig = config
				break
			}
		}
		IsOCIRepo = true
		registryCredential = &client.RegistryCredential{
			RegistryUrl:  appStoreAppVersion.AppStore.DockerArtifactStore.RegistryURL,
			Username:     appStoreAppVersion.AppStore.DockerArtifactStore.Username,
			Password:     appStoreAppVersion.AppStore.DockerArtifactStore.Password,
			AwsRegion:    appStoreAppVersion.AppStore.DockerArtifactStore.AWSRegion,
			AccessKey:    appStoreAppVersion.AppStore.DockerArtifactStore.AWSAccessKeyId,
			SecretKey:    appStoreAppVersion.AppStore.DockerArtifactStore.AWSSecretAccessKey,
			RegistryType: string(appStoreAppVersion.AppStore.DockerArtifactStore.RegistryType),
			RepoName:     appStoreAppVersion.AppStore.Name,
			IsPublic:     ociRegistryConfig.IsPublic,
		}
	} else {
		chartRepository = &client.ChartRepository{
			Name:     appStoreAppVersion.AppStore.ChartRepo.Name,
			Url:      appStoreAppVersion.AppStore.ChartRepo.Url,
			Username: appStoreAppVersion.AppStore.ChartRepo.UserName,
			Password: appStoreAppVersion.AppStore.ChartRepo.Password,
		}
	}
	installReleaseRequest := &client.InstallReleaseRequest{
		ChartName:       appStoreAppVersion.Name,
		ChartVersion:    appStoreAppVersion.Version,
		ValuesYaml:      installAppVersionRequest.ValuesOverrideYaml,
		ChartRepository: chartRepository,
		ReleaseIdentifier: &client.ReleaseIdentifier{
			ReleaseNamespace: installAppVersionRequest.Namespace,
			ReleaseName:      installAppVersionRequest.AppName,
		},
		IsOCIRepo:                  IsOCIRepo,
		RegistryCredential:         registryCredential,
		InstallAppVersionHistoryId: int32(installAppVersionRequest.InstalledAppVersionHistoryId),
	}

	_, err = impl.helmAppService.InstallRelease(ctx, installAppVersionRequest.ClusterId, installReleaseRequest)
	if err != nil {
		return installAppVersionRequest, err
	}
	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentHelmServiceImpl) GetAppStatus(installedAppAndEnvDetails repository.InstalledAppAndEnvDetails, w http.ResponseWriter, r *http.Request, token string) (string, error) {

	environment, err := impl.environmentRepository.FindById(installedAppAndEnvDetails.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("Error in getting environment", "err", err)
		return "", err
	}

	appIdentifier := &client.AppIdentifier{
		ClusterId:   environment.ClusterId,
		Namespace:   environment.Namespace,
		ReleaseName: installedAppAndEnvDetails.AppName,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	appDetail, err := impl.helmAppService.GetApplicationDetail(ctx, appIdentifier)
	if err != nil {
		// handling like argocd
		impl.Logger.Errorw("error fetching helm app resource tree", "error", err, "appIdentifier", appIdentifier)
		err = &util.ApiError{
			Code:            constants.AppDetailResourceTreeNotFound,
			InternalMessage: "Failed to get resource tree from helm",
			UserMessage:     "Failed to get resource tree from helm",
		}
		return "", err
	}

	return appDetail.ApplicationStatus, nil
}

func (impl AppStoreDeploymentHelmServiceImpl) DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error {
	if installAppVersionRequest.ForceDelete {
		return nil
	}
	appIdentifier := &client.AppIdentifier{
		ClusterId:   installAppVersionRequest.ClusterId,
		ReleaseName: installAppVersionRequest.AppName,
		Namespace:   installAppVersionRequest.Namespace,
	}

	isInstalled, err := impl.helmAppService.IsReleaseInstalled(ctx, appIdentifier)
	if err != nil {
		impl.Logger.Errorw("error in checking if helm release is installed or not", "error", err, "appIdentifier", appIdentifier)
		return err
	}

	if isInstalled {
		deleteResponse, err := impl.helmAppService.DeleteApplication(ctx, appIdentifier)
		if err != nil {
			impl.Logger.Errorw("error in deleting helm application", "error", err, "appIdentifier", appIdentifier)
			return err
		}
		if deleteResponse == nil || !deleteResponse.GetSuccess() {
			return errors.New("delete application response unsuccessful")
		}
	}

	return nil
}

// returns - valuesYamlStr, success, error
func (impl AppStoreDeploymentHelmServiceImpl) RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, bool, error) {

	// TODO : fetch values yaml from DB instead of fetching from helm cli
	// TODO Dependency : on updating helm APP, DB is not being updated. values yaml is sent directly to helm cli. After DB updatation development, we can fetch values yaml from DB, not from CLI.

	helmAppIdeltifier := &client.AppIdentifier{
		ClusterId:   installedApp.ClusterId,
		Namespace:   installedApp.Namespace,
		ReleaseName: installedApp.AppName,
	}

	helmAppDeploymentDetail, err := impl.helmAppService.GetDeploymentDetail(ctx, helmAppIdeltifier, deploymentVersion)
	if err != nil {
		impl.Logger.Errorw("Error in getting helm application deployment detail", "err", err)
		return installedApp, false, err
	}
	valuesYamlJson := helmAppDeploymentDetail.GetValuesYaml()
	valuesYamlByteArr, err := yaml.JSONToYAML([]byte(valuesYamlJson))
	if err != nil {
		impl.Logger.Errorw("Error in converting json to yaml", "err", err)
		return installedApp, false, err
	}

	// send to helm
	success, err := impl.helmAppService.RollbackRelease(ctx, helmAppIdeltifier, deploymentVersion)
	if err != nil {
		impl.Logger.Errorw("Error in helm rollback release", "err", err)
		return installedApp, false, err
	}
	installedApp.ValuesOverrideYaml = string(valuesYamlByteArr)
	return installedApp, success, nil
}

func (impl *AppStoreDeploymentHelmServiceImpl) GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*client.HelmAppDeploymentHistory, error) {
	helmAppIdeltifier := &client.AppIdentifier{
		ClusterId:   installedApp.ClusterId,
		Namespace:   installedApp.Namespace,
		ReleaseName: installedApp.AppName,
	}
	config, err := impl.helmAppService.GetClusterConf(helmAppIdeltifier.ClusterId)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "err", err)
		return nil, err
	}
	req := &client.AppDetailRequest{
		ClusterConfig: config,
		Namespace:     helmAppIdeltifier.Namespace,
		ReleaseName:   helmAppIdeltifier.ReleaseName,
	}
	history, err := impl.helmAppClient.GetDeploymentHistory(ctx, req)
	return history, err
}

func (impl *AppStoreDeploymentHelmServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error) {
	helmAppIdeltifier := &client.AppIdentifier{
		ClusterId:   installedApp.ClusterId,
		Namespace:   installedApp.Namespace,
		ReleaseName: installedApp.AppName,
	}
	config, err := impl.helmAppService.GetClusterConf(helmAppIdeltifier.ClusterId)
	if err != nil {
		impl.Logger.Errorw("error in fetching cluster detail", "clusterId", helmAppIdeltifier.ClusterId, "err", err)
		return nil, err
	}

	req := &client.DeploymentDetailRequest{
		ReleaseIdentifier: &client.ReleaseIdentifier{
			ClusterConfig:    config,
			ReleaseName:      helmAppIdeltifier.ReleaseName,
			ReleaseNamespace: helmAppIdeltifier.Namespace,
		},
		DeploymentVersion: version,
	}
	_, span := otel.Tracer("orchestrator").Start(ctx, "helmAppClient.GetDeploymentDetail")
	deploymentDetail, err := impl.helmAppClient.GetDeploymentDetail(ctx, req)
	span.End()
	if err != nil {
		impl.Logger.Errorw("error in getting deployment detail", "err", err)
		return nil, err
	}

	response := &openapi.HelmAppDeploymentManifestDetail{
		Manifest:   &deploymentDetail.Manifest,
		ValuesYaml: &deploymentDetail.ValuesYaml,
	}

	return response, nil
}

func (impl *AppStoreDeploymentHelmServiceImpl) GetGitOpsRepoName(appName string, environmentName string) (string, error) {
	return "", errors.New("method GetGitOpsRepoName not implemented")
}

func (impl *AppStoreDeploymentHelmServiceImpl) OnUpdateRepoInInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	//TODO: gitOps operations here based on flag
	if installAppVersionRequest.PerformGitOpsForHelmApp {
		_, err := impl.appStoreDeploymentCommonService.GenerateManifestAndPerformGitOperations(installAppVersionRequest)
		if err != nil {
			return installAppVersionRequest, err
		}
	}

	err := impl.updateApplicationWithChartInfo(ctx, installAppVersionRequest.InstalledAppId, installAppVersionRequest.AppStoreVersion, installAppVersionRequest.ValuesOverrideYaml, installAppVersionRequest.InstalledAppVersionHistoryId)
	if err != nil {
		return installAppVersionRequest, err
	}
	return installAppVersionRequest, err
}

func (impl *AppStoreDeploymentHelmServiceImpl) UpdateRequirementDependencies(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error {
	RequirementsString, err := impl.appStoreDeploymentCommonService.GetRequirementsString(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.Logger.Errorw("error in building requirements config for helm app", "err", err)
		return err
	}
	requirementsGitConfig, err := impl.appStoreDeploymentCommonService.GetGitCommitConfig(installAppVersionRequest, RequirementsString, appStoreBean.REQUIREMENTS_YAML_FILE)
	if err != nil {
		impl.Logger.Errorw("error in getting git config for helm app", "err", err)
		return err
	}
	_, err = impl.appStoreDeploymentCommonService.CommitConfigToGit(requirementsGitConfig)
	if err != nil {
		impl.Logger.Errorw("error in committing config to git for helm app", "err", err)
		return err
	}
	return nil
	//return errors.New("method UpdateRequirementDependencies not implemented")
}

func (impl *AppStoreDeploymentHelmServiceImpl) UpdateValuesDependencies(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.Logger.Errorw("fetching error", "err", err)
		return err
	}
	valuesString, err := impl.appStoreDeploymentCommonService.GetValuesString(appStoreAppVersion.Name, installAppVersionRequest.ValuesOverrideYaml)
	if err != nil {
		impl.Logger.Errorw("error in building requirements config for helm app", "err", err)
		return err
	}
	valuesGitConfig, err := impl.appStoreDeploymentCommonService.GetGitCommitConfig(installAppVersionRequest, valuesString, appStoreBean.VALUES_YAML_FILE)
	if err != nil {
		impl.Logger.Errorw("error in getting git config for helm app", "err", err)
		return err
	}
	_, err = impl.appStoreDeploymentCommonService.CommitConfigToGit(valuesGitConfig)
	if err != nil {
		impl.Logger.Errorw("error in committing config to git for helm app", "err", err)
		return err
	}
	return nil
}

func (impl *AppStoreDeploymentHelmServiceImpl) UpdateInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *clusterRepository.Environment, installedAppVersion *repository.InstalledAppVersions, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {

	noTargetFound := false

	if installAppVersionRequest.PerformGitOps {
		err := impl.UpdateValuesDependencies(installAppVersionRequest)
		if err != nil {
			impl.Logger.Errorw("error while commit values to git", "error", err)
			noTargetFound, _ = impl.appStoreDeploymentCommonService.ParseGitRepoErrorResponse(err)
			if noTargetFound {
				//if by mistake no content found while updating git repo, do auto fix
				installAppVersionRequest, err = impl.OnUpdateRepoInInstalledApp(ctx, installAppVersionRequest, tx)
				if err != nil {
					impl.Logger.Errorw("error while update repo on helm update", "error", err)
					return nil, err
				}
			} else {
				return nil, err
			}
		}
	}
	if !noTargetFound {
		// update chart application already called, hence skipping
		err := impl.updateApplicationWithChartInfo(ctx, installAppVersionRequest.InstalledAppId, installAppVersionRequest.AppStoreVersion, installAppVersionRequest.ValuesOverrideYaml, 0)
		if err != nil {
			return nil, err
		}
	}
	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentHelmServiceImpl) updateApplicationWithChartInfo(ctx context.Context, installedAppId int, appStoreApplicationVersionId int, valuesOverrideYaml string, installAppVersionHistoryId int) error {

	installedApp, err := impl.installedAppRepository.GetInstalledApp(installedAppId)
	if err != nil {
		impl.Logger.Errorw("error in getting in installedApp", "installedAppId", installedAppId, "err", err)
		return err
	}
	appStoreApplicationVersion, err := impl.appStoreApplicationVersionRepository.FindById(appStoreApplicationVersionId)
	if err != nil {
		impl.Logger.Errorw("error in getting in appStoreApplicationVersion", "appStoreApplicationVersionId", appStoreApplicationVersionId, "err", err)
		return err
	}
	var IsOCIRepo bool
	var registryCredential *client.RegistryCredential
	var chartRepository *client.ChartRepository
	dockerRegistryId := appStoreApplicationVersion.AppStore.DockerArtifactStoreId
	if dockerRegistryId != "" {
		ociRegistryConfigs, err := impl.OCIRegistryConfigRepository.FindByDockerRegistryId(dockerRegistryId)
		if err != nil {
			impl.Logger.Errorw("error in fetching oci registry config", "err", err)
			return err
		}
		var ociRegistryConfig *repository2.OCIRegistryConfig
		for _, config := range ociRegistryConfigs {
			if config.RepositoryAction == repository2.STORAGE_ACTION_TYPE_PULL || config.RepositoryAction == repository2.STORAGE_ACTION_TYPE_PULL_AND_PUSH {
				ociRegistryConfig = config
				break
			}
		}
		IsOCIRepo = true
		registryCredential = &client.RegistryCredential{
			RegistryUrl:  appStoreApplicationVersion.AppStore.DockerArtifactStore.RegistryURL,
			Username:     appStoreApplicationVersion.AppStore.DockerArtifactStore.Username,
			Password:     appStoreApplicationVersion.AppStore.DockerArtifactStore.Password,
			AwsRegion:    appStoreApplicationVersion.AppStore.DockerArtifactStore.AWSRegion,
			AccessKey:    appStoreApplicationVersion.AppStore.DockerArtifactStore.AWSAccessKeyId,
			SecretKey:    appStoreApplicationVersion.AppStore.DockerArtifactStore.AWSSecretAccessKey,
			RegistryType: string(appStoreApplicationVersion.AppStore.DockerArtifactStore.RegistryType),
			RepoName:     appStoreApplicationVersion.AppStore.Name,
			IsPublic:     ociRegistryConfig.IsPublic,
		}
	} else {
		chartRepository = &client.ChartRepository{
			Name:     appStoreApplicationVersion.AppStore.ChartRepo.Name,
			Url:      appStoreApplicationVersion.AppStore.ChartRepo.Url,
			Username: appStoreApplicationVersion.AppStore.ChartRepo.UserName,
			Password: appStoreApplicationVersion.AppStore.ChartRepo.Password,
		}
	}

	updateReleaseRequest := &client.UpdateApplicationWithChartInfoRequestDto{
		InstallReleaseRequest: &client.InstallReleaseRequest{
			ValuesYaml: valuesOverrideYaml,
			ReleaseIdentifier: &client.ReleaseIdentifier{
				ReleaseNamespace: installedApp.Environment.Namespace,
				ReleaseName:      installedApp.App.AppName,
			},
			ChartName:                  appStoreApplicationVersion.Name,
			ChartVersion:               appStoreApplicationVersion.Version,
			ChartRepository:            chartRepository,
			RegistryCredential:         registryCredential,
			IsOCIRepo:                  IsOCIRepo,
			InstallAppVersionHistoryId: int32(installAppVersionHistoryId),
		},
		SourceAppType: client.SOURCE_HELM_APP,
	}
	res, err := impl.helmAppService.UpdateApplicationWithChartInfo(ctx, installedApp.Environment.ClusterId, updateReleaseRequest)
	if err != nil {
		impl.Logger.Errorw("error in updating helm application", "err", err)
		return err
	}
	if !res.GetSuccess() {
		return errors.New("helm application update unsuccessful")
	}
	return nil
}
func (impl *AppStoreDeploymentHelmServiceImpl) DeleteDeploymentApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	return nil
}

func (impl *AppStoreDeploymentHelmServiceImpl) SaveTimelineForACDHelmApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, status string, statusDetail string, tx *pg.Tx) error {
	return nil
}

func (impl *AppStoreDeploymentHelmServiceImpl) UpdateInstalledAppAndPipelineStatusForFailedDeploymentStatus(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error {
	return nil
}

// TODO: Need to refactor this,refer below reason
// This is being done as in ea mode wire argocd service is being binded to helmServiceImpl due to which we are restricted to implement this here.
// RefreshAndUpdateACDApp this will update chart info in acd app if required in case of mono repo migration and will refresh argo app
func (impl *AppStoreDeploymentHelmServiceImpl) UpdateAndSyncACDApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *util.ChartGitAttribute, isMonoRepoMigrationRequired bool, ctx context.Context, tx *pg.Tx) error {
	return errors.New("this is not implemented")
}
