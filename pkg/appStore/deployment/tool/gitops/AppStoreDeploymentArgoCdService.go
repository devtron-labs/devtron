package appStoreDeploymentGitopsTool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/constants"
	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentFullMode "github.com/devtron-labs/devtron/pkg/appStore/deployment/fullMode"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/ghodss/yaml"
	"github.com/go-pg/pg"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

type AppStoreDeploymentArgoCdService interface {
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
	GetAppStatus(installedAppAndEnvDetails repository.InstalledAppAndEnvDetails, w http.ResponseWriter, r *http.Request, token string) (string, error)
	DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error
	RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32) (*appStoreBean.InstallAppVersionDTO, bool, error)
	GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*client.HelmAppDeploymentHistory, error)
	GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error)
	GetGitOpsRepoName(appName string, environmentName string) (string, error)
	OnUpdateRepoInInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error)
	UpdateRequirementDependencies(environment *clusterRepository.Environment, installedAppVersion *repository.InstalledAppVersions, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error
	UpdateInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *clusterRepository.Environment, installedAppVersion *repository.InstalledAppVersions) (*appStoreBean.InstallAppVersionDTO, error)
}

type AppStoreDeploymentArgoCdServiceImpl struct {
	Logger                            *zap.SugaredLogger
	appStoreDeploymentFullModeService appStoreDeploymentFullMode.AppStoreDeploymentFullModeService
	acdClient                         application2.ServiceClient
	chartGroupDeploymentRepository    repository.ChartGroupDeploymentRepository
	installedAppRepository            repository.InstalledAppRepository
	installedAppRepositoryHistory     repository.InstalledAppVersionHistoryRepository
	chartTemplateService              util.ChartTemplateService
	gitOpsRepository                  repository2.GitOpsConfigRepository
	gitFactory                        *util.GitFactory
}

func NewAppStoreDeploymentArgoCdServiceImpl(logger *zap.SugaredLogger, appStoreDeploymentFullModeService appStoreDeploymentFullMode.AppStoreDeploymentFullModeService,
	acdClient application2.ServiceClient, chartGroupDeploymentRepository repository.ChartGroupDeploymentRepository,
	installedAppRepository repository.InstalledAppRepository, installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository, chartTemplateService util.ChartTemplateService,
	gitOpsRepository repository2.GitOpsConfigRepository, gitFactory *util.GitFactory) *AppStoreDeploymentArgoCdServiceImpl {
	return &AppStoreDeploymentArgoCdServiceImpl{
		Logger:                            logger,
		appStoreDeploymentFullModeService: appStoreDeploymentFullModeService,
		acdClient:                         acdClient,
		chartGroupDeploymentRepository:    chartGroupDeploymentRepository,
		installedAppRepository:            installedAppRepository,
		installedAppRepositoryHistory:     installedAppRepositoryHistory,
		chartTemplateService:              chartTemplateService,
		gitOpsRepository:                  gitOpsRepository,
		gitFactory:                        gitFactory,
	}
}

func (impl AppStoreDeploymentArgoCdServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {
	//step 2 git operation pull push
	installAppVersionRequest, chartGitAttr, err := impl.appStoreDeploymentFullModeService.AppStoreDeployOperationGIT(installAppVersionRequest)
	if err != nil {
		impl.Logger.Errorw(" error", "err", err)
		return installAppVersionRequest, err
	}
	//step 3 acd operation register, sync
	installAppVersionRequest, err = impl.appStoreDeploymentFullModeService.AppStoreDeployOperationACD(installAppVersionRequest, chartGitAttr, ctx)
	if err != nil {
		impl.Logger.Errorw(" error", "err", err)
		return installAppVersionRequest, err
	}
	return installAppVersionRequest, nil
}

// TODO: Test ACD to get status
func (impl AppStoreDeploymentArgoCdServiceImpl) GetAppStatus(installedAppAndEnvDetails repository.InstalledAppAndEnvDetails, w http.ResponseWriter, r *http.Request, token string) (string, error) {
	if len(installedAppAndEnvDetails.AppName) > 0 && len(installedAppAndEnvDetails.EnvironmentName) > 0 {
		acdAppName := installedAppAndEnvDetails.AppName + "-" + installedAppAndEnvDetails.EnvironmentName
		query := &application.ResourcesQuery{
			ApplicationName: &acdAppName,
		}
		ctx, cancel := context.WithCancel(r.Context())
		if cn, ok := w.(http.CloseNotifier); ok {
			go func(done <-chan struct{}, closed <-chan bool) {
				select {
				case <-done:
				case <-closed:
					cancel()
				}
			}(ctx.Done(), cn.CloseNotify())
		}
		ctx = context.WithValue(ctx, "token", token)
		defer cancel()
		impl.Logger.Debugf("Getting status for app %s in env %s", installedAppAndEnvDetails.AppName, installedAppAndEnvDetails.EnvironmentName)
		start := time.Now()
		resp, err := impl.acdClient.ResourceTree(ctx, query)
		elapsed := time.Since(start)
		impl.Logger.Debugf("Time elapsed %s in fetching application %s for environment %s", elapsed, installedAppAndEnvDetails.AppName, installedAppAndEnvDetails.EnvironmentName)
		if err != nil {
			impl.Logger.Errorw("error fetching resource tree", "error", err)
			err = &util.ApiError{
				Code:            constants.AppDetailResourceTreeNotFound,
				InternalMessage: "app detail fetched, failed to get resource tree from acd",
				UserMessage:     "app detail fetched, failed to get resource tree from acd",
			}
			return "", err

		}
		return resp.Status, nil
	}
	return "", errors.New("invalid app name or env name")
}

func (impl AppStoreDeploymentArgoCdServiceImpl) DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error {
	acdAppName := appName + "-" + environmentName
	err := impl.deleteACD(acdAppName, ctx)
	if err != nil {
		impl.Logger.Errorw("error in deleting ACD ", "name", acdAppName, "err", err)
		if installAppVersionRequest.ForceDelete {
			impl.Logger.Warnw("error while deletion of app in acd, continue to delete in db as this operation is force delete", "error", err)
		} else {
			//statusError, _ := err.(*errors2.StatusError)
			if strings.Contains(err.Error(), "code = NotFound") {
				err = &util.ApiError{
					UserMessage:     "Could not delete as application not found in argocd",
					InternalMessage: err.Error(),
				}
			} else {
				err = &util.ApiError{
					UserMessage:     "Could not delete application",
					InternalMessage: err.Error(),
				}
			}
			return err
		}
	}
	deployment, err := impl.chartGroupDeploymentRepository.FindByInstalledAppId(installedApps.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in fetching chartGroupMapping", "id", installedApps.Id, "err", err)
		return err
	} else if err == pg.ErrNoRows {
		impl.Logger.Infow("not a chart group deployment skipping chartGroupMapping delete", "id", installedApps.Id)
	} else {
		deployment.Deleted = true
		deployment.UpdatedOn = time.Now()
		deployment.UpdatedBy = installAppVersionRequest.UserId
		_, err := impl.chartGroupDeploymentRepository.Update(deployment, dbTransaction)
		if err != nil {
			impl.Logger.Errorw("error in mapping delete", "err", err)
			return err
		}
	}

	return nil
}

// returns - valuesYamlStr, success, error
func (impl AppStoreDeploymentArgoCdServiceImpl) RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, installedAppVersionHistoryId int32) (*appStoreBean.InstallAppVersionDTO, bool, error) {
	//request version id for
	versionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(int(installedAppVersionHistoryId))
	if err != nil {
		impl.Logger.Errorw("error", "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 404, UserMessage: fmt.Sprintf("No deployment history version found for id: %d", installedAppVersionHistoryId), InternalMessage: err.Error()}
		return installedApp, false, err
	}
	installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersionAny(versionHistory.InstalledAppVersionId)
	if err != nil {
		impl.Logger.Errorw("error", "err", err)
		err = &util.ApiError{Code: "404", HttpStatusCode: 404, UserMessage: fmt.Sprintf("No installed app version found for id: %d", versionHistory.InstalledAppVersionId), InternalMessage: err.Error()}
		return installedApp, false, err
	}
	activeInstalledAppVersion, err := impl.installedAppRepository.GetActiveInstalledAppVersionByInstalledAppId(installedApp.InstalledAppId)
	if err != nil {
		impl.Logger.Errorw("error", "err", err)
		return installedApp, false, err
	}

	//validate relations
	if installedApp.InstalledAppId != installedAppVersion.InstalledAppId {
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "bad request, requested version are not belongs to each other", InternalMessage: ""}
		return installedApp, false, err
	}

	installedApp.InstalledAppVersionId = installedAppVersion.Id
	installedApp.AppStoreVersion = installedAppVersion.AppStoreApplicationVersionId
	installedApp.ValuesOverrideYaml = versionHistory.ValuesYamlRaw
	installedApp.AppStoreId = installedAppVersion.AppStoreApplicationVersion.AppStoreId
	installedApp.AppStoreName = installedAppVersion.AppStoreApplicationVersion.AppStore.Name
	installedApp.GitOpsRepoName = installedAppVersion.InstalledApp.GitOpsRepoName
	installedApp.ACDAppName = fmt.Sprintf("%s-%s", installedApp.AppName, installedApp.EnvironmentName)
	//If current version upgrade/degrade to another, update requirement dependencies
	if versionHistory.InstalledAppVersionId != activeInstalledAppVersion.Id {
		err = impl.appStoreDeploymentFullModeService.UpdateRequirementYaml(installedApp, &installedAppVersion.AppStoreApplicationVersion)
		if err != nil {
			impl.Logger.Errorw("error", "err", err)
			return installedApp, false, nil
		}

		activeInstalledAppVersion.Active = false
		_, err = impl.installedAppRepository.UpdateInstalledAppVersion(activeInstalledAppVersion, nil)
		if err != nil {
			impl.Logger.Errorw("error", "err", err)
			return installedApp, false, nil
		}
	}
	//Update Values config
	installedApp, err = impl.appStoreDeploymentFullModeService.UpdateValuesYaml(installedApp)
	if err != nil {
		impl.Logger.Errorw("error", "err", err)
		return installedApp, false, nil
	}
	//ACD sync operation
	impl.appStoreDeploymentFullModeService.SyncACD(installedApp.ACDAppName, ctx)
	return installedApp, true, nil
}

func (impl AppStoreDeploymentArgoCdServiceImpl) deleteACD(acdAppName string, ctx context.Context) error {
	req := new(application.ApplicationDeleteRequest)
	req.Name = &acdAppName
	if ctx == nil {
		impl.Logger.Errorw("err in delete ACD for AppStore, ctx is NULL", "acdAppName", acdAppName)
		return fmt.Errorf("context is null")
	}
	if _, err := impl.acdClient.Delete(ctx, req); err != nil {
		impl.Logger.Errorw("err in delete ACD for AppStore", "acdAppName", acdAppName, "err", err)
		return err
	}
	return nil
}
func (impl AppStoreDeploymentArgoCdServiceImpl) getSourcesFromManifest(chartYaml string) ([]string, error) {
	var b map[string]interface{}
	var sources []string
	err := json.Unmarshal([]byte(chartYaml), &b)
	if err != nil {
		impl.Logger.Errorw("error while unmarshal chart yaml", "error", err)
		return sources, err
	}
	if b != nil && b["sources"] != nil {
		slice := b["sources"].([]interface{})
		for _, item := range slice {
			sources = append(sources, item.(string))
		}
	}
	return sources, nil
}
func (impl AppStoreDeploymentArgoCdServiceImpl) GetDeploymentHistory(ctx context.Context, installedAppDto *appStoreBean.InstallAppVersionDTO) (*client.HelmAppDeploymentHistory, error) {
	result := &client.HelmAppDeploymentHistory{}
	var history []*client.HelmAppDeploymentDetail
	//TODO - response setup

	installedAppVersions, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdMeta(installedAppDto.InstalledAppId)
	if err != nil {
		impl.Logger.Errorw("error while fetching installed version", "error", err)
		return result, err
	}
	for _, installedAppVersionModel := range installedAppVersions {

		sources, err := impl.getSourcesFromManifest(installedAppVersionModel.AppStoreApplicationVersion.ChartYaml)
		if err != nil {
			impl.Logger.Errorw("error while fetching sources", "error", err)
			//continues here, skip error in case found issue on fetching source
		}
		versionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistoryByVersionId(installedAppVersionModel.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.Logger.Errorw("error while fetching installed version history", "error", err)
			return result, err
		}
		for _, updateHistory := range versionHistory {
			history = append(history, &client.HelmAppDeploymentDetail{
				ChartMetadata: &client.ChartMetadata{
					ChartName:    installedAppVersionModel.AppStoreApplicationVersion.AppStore.Name,
					ChartVersion: installedAppVersionModel.AppStoreApplicationVersion.Version,
					Description:  installedAppVersionModel.AppStoreApplicationVersion.Description,
					Home:         installedAppVersionModel.AppStoreApplicationVersion.Home,
					Sources:      sources,
				},
				DockerImages: []string{installedAppVersionModel.AppStoreApplicationVersion.AppVersion},
				DeployedAt: &timestamp.Timestamp{
					Seconds: updateHistory.UpdatedOn.Unix(),
					Nanos:   int32(updateHistory.UpdatedOn.Nanosecond()),
				},
				Version: int32(updateHistory.Id),
			})
		}
	}

	if len(history) == 0 {
		history = make([]*client.HelmAppDeploymentDetail, 0)
	}
	result.DeploymentHistory = history
	return result, err
}

func (impl AppStoreDeploymentArgoCdServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error) {
	values := &openapi.HelmAppDeploymentManifestDetail{}
	versionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(int(version))
	if err != nil {
		impl.Logger.Errorw("error while fetching installed version history", "error", err)
		return nil, err
	}
	values.ValuesYaml = &versionHistory.ValuesYamlRaw
	return values, err
}

func (impl AppStoreDeploymentArgoCdServiceImpl) GetGitOpsRepoName(appName string, environmentName string) (string, error) {
	return impl.appStoreDeploymentFullModeService.GetGitOpsRepoName(appName, environmentName)
}

func (impl *AppStoreDeploymentArgoCdServiceImpl) OnUpdateRepoInInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error) {
	//git operation pull push
	installAppVersionRequest, chartGitAttr, err := impl.appStoreDeploymentFullModeService.AppStoreDeployOperationGIT(installAppVersionRequest)
	if err != nil {
		return installAppVersionRequest, err
	}

	//acd operation register, sync
	installAppVersionRequest, err = impl.patchAcdApp(ctx, installAppVersionRequest, chartGitAttr)
	if err != nil {
		return installAppVersionRequest, err
	}

	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentArgoCdServiceImpl) UpdateRequirementDependencies(environment *clusterRepository.Environment, installedAppVersion *repository.InstalledAppVersions, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error {
	argocdAppName := installAppVersionRequest.AppName + "-" + environment.Name
	dependency := appStoreBean.Dependency{
		Name:       appStoreAppVersion.AppStore.Name,
		Version:    appStoreAppVersion.Version,
		Repository: appStoreAppVersion.AppStore.ChartRepo.Url,
	}
	var dependencies []appStoreBean.Dependency
	dependencies = append(dependencies, dependency)
	requirementDependencies := &appStoreBean.Dependencies{
		Dependencies: dependencies,
	}
	requirementDependenciesByte, err := json.Marshal(requirementDependencies)
	if err != nil {
		return err
	}
	requirementDependenciesByte, err = yaml.JSONToYAML(requirementDependenciesByte)
	if err != nil {
		return err
	}
	//getting user name & emailId for commit author data
	userEmailId, userName := impl.chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit(installAppVersionRequest.UserId)
	requirmentYamlConfig := &util.ChartConfig{
		FileName:       appStoreBean.REQUIREMENTS_YAML_FILE,
		FileContent:    string(requirementDependenciesByte),
		ChartName:      installedAppVersion.AppStoreApplicationVersion.AppStore.Name,
		ChartLocation:  argocdAppName,
		ChartRepoName:  installAppVersionRequest.GitOpsRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", appStoreAppVersion.Id, environment.Id),
		UserEmailId:    userEmailId,
		UserName:       userName,
	}
	gitOpsConfigBitbucket, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(util.BITBUCKET_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
		} else {
			impl.Logger.Errorw("error in fetching gitOps bitbucket config", "err", err)
			return err
		}
	}
	_, err = impl.gitFactory.Client.CommitValues(requirmentYamlConfig, gitOpsConfigBitbucket.BitBucketWorkspaceId)
	if err != nil {
		impl.Logger.Errorw("error in git commit", "err", err)
		return err
	}
	return nil
}

func (impl AppStoreDeploymentArgoCdServiceImpl) UpdateInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *clusterRepository.Environment, installedAppVersion *repository.InstalledAppVersions) (*appStoreBean.InstallAppVersionDTO, error) {
	//update values yaml in chart
	installAppVersionRequest, err := impl.updateValuesYaml(environment, installedAppVersion, installAppVersionRequest)
	if err != nil {
		impl.Logger.Errorw("error while commit values to git", "error", err)
		return nil, err
	}
	installAppVersionRequest.Environment = environment

	//ACD sync operation
	impl.appStoreDeploymentFullModeService.SyncACD(installAppVersionRequest.ACDAppName, ctx)

	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentArgoCdServiceImpl) patchAcdApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute) (*appStoreBean.InstallAppVersionDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	//registerInArgo
	err := impl.appStoreDeploymentFullModeService.RegisterInArgo(chartGitAttr, ctx)
	if err != nil {
		impl.Logger.Errorw("error in argo registry", "err", err)
		return nil, err
	}
	// update acd app
	patchReq := v1alpha1.Application{Spec: v1alpha1.ApplicationSpec{Source: v1alpha1.ApplicationSource{Path: chartGitAttr.ChartLocation, RepoURL: chartGitAttr.RepoUrl}}}
	reqbyte, err := json.Marshal(patchReq)
	if err != nil {
		impl.Logger.Errorw("error in creating patch", "err", err)
	}
	_, err = impl.acdClient.Patch(ctx, &application.ApplicationPatchRequest{Patch: string(reqbyte), Name: &installAppVersionRequest.ACDAppName, PatchType: "merge"})
	if err != nil {
		impl.Logger.Errorw("error in creating argo app ", "name", installAppVersionRequest.ACDAppName, "patch", string(reqbyte), "err", err)
		return nil, err
	}
	impl.appStoreDeploymentFullModeService.SyncACD(installAppVersionRequest.ACDAppName, ctx)
	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentArgoCdServiceImpl) updateValuesYaml(environment *clusterRepository.Environment, installedAppVersion *repository.InstalledAppVersions,
	installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error) {

	argocdAppName := installAppVersionRequest.AppName + "-" + environment.Name
	valuesOverrideByte, err := yaml.YAMLToJSON([]byte(installAppVersionRequest.ValuesOverrideYaml))
	if err != nil {
		impl.Logger.Errorw("error in json patch", "err", err)
		return installAppVersionRequest, err
	}
	var dat map[string]interface{}
	err = json.Unmarshal(valuesOverrideByte, &dat)
	if err != nil {
		impl.Logger.Errorw("error in unmarshal", "err", err)
		return installAppVersionRequest, err
	}
	valuesMap := make(map[string]map[string]interface{})
	valuesMap[installedAppVersion.AppStoreApplicationVersion.AppStore.Name] = dat
	valuesByte, err := json.Marshal(valuesMap)
	if err != nil {
		impl.Logger.Errorw("error in marshaling", "err", err)
		return installAppVersionRequest, err
	}
	//getting user name & emailId for commit author data
	userEmailId, userName := impl.chartTemplateService.GetUserEmailIdAndNameForGitOpsCommit(installAppVersionRequest.UserId)
	valuesConfig := &util.ChartConfig{
		FileName:       appStoreBean.VALUES_YAML_FILE,
		FileContent:    string(valuesByte),
		ChartName:      installedAppVersion.AppStoreApplicationVersion.AppStore.Name,
		ChartLocation:  argocdAppName,
		ChartRepoName:  installAppVersionRequest.GitOpsRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", installedAppVersion.AppStoreApplicationVersion.Id, environment.Id),
		UserEmailId:    userEmailId,
		UserName:       userName,
	}
	gitOpsConfigBitbucket, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(util.BITBUCKET_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
		} else {
			impl.Logger.Errorw("error in fetching gitOps bitbucket config", "err", err)
			return installAppVersionRequest, err
		}
	}
	commitHash, err := impl.gitFactory.Client.CommitValues(valuesConfig, gitOpsConfigBitbucket.BitBucketWorkspaceId)
	if err != nil {
		impl.Logger.Errorw("error in git commit", "err", err)
		return installAppVersionRequest, err
	}
	installAppVersionRequest.GitHash = commitHash
	return installAppVersionRequest, nil
}
