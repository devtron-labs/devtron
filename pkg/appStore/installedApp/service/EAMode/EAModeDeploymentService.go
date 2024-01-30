package EAMode

import (
	"context"
	"errors"
	bean2 "github.com/devtron-labs/devtron/api/helm-app/bean"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/bean"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"time"

	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"sigs.k8s.io/yaml"
)

type EAModeDeploymentService interface {
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error
	RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, bool, error)
	GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error)
	GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error)
	UpgradeDeployment(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *commonBean.ChartGitAttribute, installedAppVersionHistoryId int, ctx context.Context) error
}

type EAModeDeploymentServiceImpl struct {
	Logger                               *zap.SugaredLogger
	helmAppService                       client.HelmAppService
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	// TODO fix me next
	helmAppClient               gRPC.HelmAppClient // TODO refactoring: use HelmAppService instead
	installedAppRepository      repository.InstalledAppRepository
	OCIRegistryConfigRepository repository2.OCIRegistryConfigRepository
}

func NewEAModeDeploymentServiceImpl(logger *zap.SugaredLogger, helmAppService client.HelmAppService,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	helmAppClient gRPC.HelmAppClient,
	installedAppRepository repository.InstalledAppRepository,
	OCIRegistryConfigRepository repository2.OCIRegistryConfigRepository) *EAModeDeploymentServiceImpl {
	return &EAModeDeploymentServiceImpl{
		Logger:                               logger,
		helmAppService:                       helmAppService,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		helmAppClient:                        helmAppClient,
		installedAppRepository:               installedAppRepository,
		OCIRegistryConfigRepository:          OCIRegistryConfigRepository,
	}
}

func (impl *EAModeDeploymentServiceImpl) UpgradeDeployment(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *commonBean.ChartGitAttribute, installedAppVersionHistoryId int, ctx context.Context) error {
	err := impl.updateApplicationWithChartInfo(ctx, installAppVersionRequest.InstalledAppId, installAppVersionRequest.AppStoreVersion, installAppVersionRequest.ValuesOverrideYaml, installedAppVersionHistoryId)
	if err != nil {
		impl.Logger.Errorw("error in updating helm app", "err", err)
		return err
	}
	return nil
}

func (impl *EAModeDeploymentServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	installAppVersionRequest.DeploymentAppType = util.PIPELINE_DEPLOYMENT_TYPE_HELM
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.Logger.Errorw("fetching error", "err", err)
		return installAppVersionRequest, err
	}
	var IsOCIRepo bool
	var registryCredential *gRPC.RegistryCredential
	var chartRepository *gRPC.ChartRepository
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
		registryCredential = &gRPC.RegistryCredential{
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
		chartRepository = &gRPC.ChartRepository{
			Name:     appStoreAppVersion.AppStore.ChartRepo.Name,
			Url:      appStoreAppVersion.AppStore.ChartRepo.Url,
			Username: appStoreAppVersion.AppStore.ChartRepo.UserName,
			Password: appStoreAppVersion.AppStore.ChartRepo.Password,
		}
	}
	installReleaseRequest := &gRPC.InstallReleaseRequest{
		ChartName:       appStoreAppVersion.Name,
		ChartVersion:    appStoreAppVersion.Version,
		ValuesYaml:      installAppVersionRequest.ValuesOverrideYaml,
		ChartRepository: chartRepository,
		ReleaseIdentifier: &gRPC.ReleaseIdentifier{
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

func (impl *EAModeDeploymentServiceImpl) DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error {
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
func (impl *EAModeDeploymentServiceImpl) RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, bool, error) {

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

func (impl *EAModeDeploymentServiceImpl) GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error) {
	helmAppIdentifier := &client.AppIdentifier{
		ClusterId:   installedApp.ClusterId,
		Namespace:   installedApp.Namespace,
		ReleaseName: installedApp.AppName,
	}
	history, err := impl.helmAppService.GetDeploymentHistory(ctx, helmAppIdentifier)
	return history, err
}

func (impl *EAModeDeploymentServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error) {
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

	req := &gRPC.DeploymentDetailRequest{
		ReleaseIdentifier: &gRPC.ReleaseIdentifier{
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

func (impl *EAModeDeploymentServiceImpl) updateApplicationWithChartInfo(ctx context.Context, installedAppId int, appStoreApplicationVersionId int, valuesOverrideYaml string, installAppVersionHistoryId int) error {
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
	var registryCredential *gRPC.RegistryCredential
	var chartRepository *gRPC.ChartRepository
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
		registryCredential = &gRPC.RegistryCredential{
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
		chartRepository = &gRPC.ChartRepository{
			Name:     appStoreApplicationVersion.AppStore.ChartRepo.Name,
			Url:      appStoreApplicationVersion.AppStore.ChartRepo.Url,
			Username: appStoreApplicationVersion.AppStore.ChartRepo.UserName,
			Password: appStoreApplicationVersion.AppStore.ChartRepo.Password,
		}
	}

	updateReleaseRequest := &bean2.UpdateApplicationWithChartInfoRequestDto{
		InstallReleaseRequest: &gRPC.InstallReleaseRequest{
			ValuesYaml: valuesOverrideYaml,
			ReleaseIdentifier: &gRPC.ReleaseIdentifier{
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
		SourceAppType: bean2.SOURCE_HELM_APP,
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

func (impl *EAModeDeploymentServiceImpl) GetAcdAppGitOpsRepoName(appName string, environmentName string) (string, error) {
	return "", errors.New("method GetGitOpsRepoName not implemented")
}

func (impl *EAModeDeploymentServiceImpl) DeleteACDAppObject(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	return errors.New("this is not implemented")
}

func (impl *EAModeDeploymentServiceImpl) SaveTimelineForHelmApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, status string, statusDetail string, statusTime time.Time, tx *pg.Tx) error {
	return errors.New("this is not implemented")
}

func (impl *EAModeDeploymentServiceImpl) UpdateInstalledAppAndPipelineStatusForFailedDeploymentStatus(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error {
	return errors.New("this is not implemented")
}

// TODO: Need to refactor this,refer below reason
// This is being done as in ea mode wire argocd service is being binded to helmServiceImpl due to which we are restricted to implement this here.
// RefreshAndUpdateACDApp this will update chart info in acd app if required in case of mono repo migration and will refresh argo app
func (impl *EAModeDeploymentServiceImpl) UpdateAndSyncACDApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *commonBean.ChartGitAttribute, isMonoRepoMigrationRequired bool, ctx context.Context, tx *pg.Tx) error {
	return errors.New("this is not implemented")
}

func (impl *EAModeDeploymentServiceImpl) UpdateValuesDependencies(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	return errors.New("this is not implemented")
}

func (impl *EAModeDeploymentServiceImpl) ParseGitRepoErrorResponse(err error) (bool, error) {
	return false, errors.New("this is not implemented")
}

func (impl *EAModeDeploymentServiceImpl) GitOpsOperations(manifestResponse *bean.AppStoreManifestResponse, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*bean.AppStoreGitOpsResponse, error) {
	return nil, errors.New("this is not implemented")
}

func (impl *EAModeDeploymentServiceImpl) GenerateManifestAndPerformGitOperations(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*bean.AppStoreGitOpsResponse, error) {
	return nil, errors.New("this is not implemented")
}

func (impl *EAModeDeploymentServiceImpl) GenerateManifest(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (manifestResponse *bean.AppStoreManifestResponse, err error) {
	return nil, errors.New("this is not implemented")
}

func (impl *EAModeDeploymentServiceImpl) CheckIfArgoAppExists(acdAppName string) (isFound bool, err error) {
	return isFound, errors.New("this is not implemented")
}

func (impl *EAModeDeploymentServiceImpl) CommitValues(chartGitAttr *git.ChartConfig) (commitHash string, commitTime time.Time, err error) {
	return commitHash, commitTime, errors.New("this is not implemented")
}
