package deployment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/common"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	"net/http"
	"time"

	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/client/argocdServer"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/chartGroup/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/go-pg/pg"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
)

// FullModeDeploymentService TODO refactoring: Use extended binding over EAMode.EAModeDeploymentService
// Currently creating duplicate methods in EAMode.EAModeDeploymentService
type FullModeDeploymentService interface {
	// ArgoCd Services ---------------------------------

	// InstallApp will register git repo in Argo, create and sync the Argo App and finally update deployment status
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	// DeleteInstalledApp will delete entry from appStatus.AppStatusDto table and from repository.ChartGroupDeployment table (if exists)
	DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error
	// RollbackRelease will rollback to a previous deployment for the given installedAppVersionHistoryId; returns - valuesYamlStr, success, error
	RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, bool, error)
	// GetDeploymentHistory will return gRPC.HelmAppDeploymentHistory for the given installedAppDto.InstalledAppId
	GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error)
	// GetDeploymentHistoryInfo will return openapi.HelmAppDeploymentManifestDetail for the given appStoreBean.InstallAppVersionDTO
	GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error)

	InstalledAppArgoCdService
	DeploymentStatusService
	InstalledAppGitOpsService
	InstalledAppVirtualDeploymentService
}

type FullModeDeploymentServiceImpl struct {
	Logger                               *zap.SugaredLogger
	acdClient                            application2.ServiceClient
	argoK8sClient                        argocdServer.ArgoK8sClient
	aCDAuthConfig                        *util2.ACDAuthConfig
	chartGroupDeploymentRepository       repository2.ChartGroupDeploymentRepository
	installedAppRepository               repository.InstalledAppRepository
	installedAppRepositoryHistory        repository.InstalledAppVersionHistoryRepository
	argoUserService                      argo.ArgoUserService
	appStoreDeploymentCommonService      appStoreDeploymentCommon.AppStoreDeploymentCommonService
	helmAppService                       client.HelmAppService
	appStatusService                     appStatus.AppStatusService
	pipelineStatusTimelineService        status.PipelineStatusTimelineService
	pipelineStatusTimelineRepository     pipelineConfig.PipelineStatusTimelineRepository
	userService                          user.UserService
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	argoClientWrapperService             argocdServer.ArgoClientWrapperService
	acdConfig                            *argocdServer.ACDConfig
	gitOperationService                  git.GitOperationService
	gitOpsConfigReadService              config.GitOpsConfigReadService
	environmentRepository                repository5.EnvironmentRepository
}

func NewFullModeDeploymentServiceImpl(
	logger *zap.SugaredLogger,
	acdClient application2.ServiceClient,
	argoK8sClient argocdServer.ArgoK8sClient,
	aCDAuthConfig *util2.ACDAuthConfig,
	chartGroupDeploymentRepository repository2.ChartGroupDeploymentRepository,
	installedAppRepository repository.InstalledAppRepository,
	installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository,
	argoUserService argo.ArgoUserService,
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	helmAppService client.HelmAppService,
	appStatusService appStatus.AppStatusService,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	userService user.UserService,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	argoClientWrapperService argocdServer.ArgoClientWrapperService,
	acdConfig *argocdServer.ACDConfig,
	gitOperationService git.GitOperationService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	environmentRepository repository5.EnvironmentRepository) *FullModeDeploymentServiceImpl {
	return &FullModeDeploymentServiceImpl{
		Logger:                               logger,
		acdClient:                            acdClient,
		argoK8sClient:                        argoK8sClient,
		aCDAuthConfig:                        aCDAuthConfig,
		chartGroupDeploymentRepository:       chartGroupDeploymentRepository,
		installedAppRepository:               installedAppRepository,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
		argoUserService:                      argoUserService,
		appStoreDeploymentCommonService:      appStoreDeploymentCommonService,
		helmAppService:                       helmAppService,
		appStatusService:                     appStatusService,
		pipelineStatusTimelineService:        pipelineStatusTimelineService,
		pipelineStatusTimelineRepository:     pipelineStatusTimelineRepository,
		userService:                          userService,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		argoClientWrapperService:             argoClientWrapperService,
		acdConfig:                            acdConfig,
		gitOperationService:                  gitOperationService,
		gitOpsConfigReadService:              gitOpsConfigReadService,
		environmentRepository:                environmentRepository,
	}
}

func (impl *FullModeDeploymentServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	//STEP 4: registerInArgo
	err := impl.argoClientWrapperService.RegisterGitOpsRepoInArgo(ctx, chartGitAttr.RepoUrl)
	if err != nil {
		impl.Logger.Errorw("error in argo registry", "err", err)
		return nil, err
	}
	//STEP 5: createInArgo
	err = impl.createInArgo(chartGitAttr, *installAppVersionRequest.Environment, installAppVersionRequest.ACDAppName)
	if err != nil {
		impl.Logger.Errorw("error in create in argo", "err", err)
		return nil, err
	}
	//STEP 6: Force Sync ACD - works like trigger deployment
	//impl.SyncACD(installAppVersionRequest.ACDAppName, ctx)

	//STEP 7: normal refresh ACD - update for step 6 to avoid delay
	syncTime := time.Now()
	err = impl.argoClientWrapperService.SyncArgoCDApplicationIfNeededAndRefresh(ctx, installAppVersionRequest.ACDAppName)
	if err != nil {
		impl.Logger.Errorw("error in getting the argo application with normal refresh", "err", err)
		return nil, err
	}
	if !impl.acdConfig.ArgoCDAutoSyncEnabled {
		timeline := &pipelineConfig.PipelineStatusTimeline{
			InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
			Status:                       pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED,
			StatusDetail:                 "argocd sync completed.",
			StatusTime:                   syncTime,
			AuditLog: sql.AuditLog{
				CreatedBy: installAppVersionRequest.UserId,
				CreatedOn: time.Now(),
				UpdatedBy: installAppVersionRequest.UserId,
				UpdatedOn: time.Now(),
			},
		}
		err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, tx, true)
		if err != nil {
			impl.Logger.Errorw("error in creating timeline for argocd sync", "err", err, "timeline", timeline)
		}
	}

	return installAppVersionRequest, nil
}

func (impl *FullModeDeploymentServiceImpl) DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error {

	err := impl.appStatusService.DeleteWithAppIdEnvId(dbTransaction, installedApps.AppId, installedApps.EnvironmentId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("error in deleting app_status", "appId", installedApps.AppId, "envId", installedApps.EnvironmentId, "err", err)
		return err
	} else if err == pg.ErrNoRows {
		impl.Logger.Warnw("App status not present, skipping app status delete ")
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

func (impl *FullModeDeploymentServiceImpl) RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, installedAppVersionHistoryId int32, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, bool, error) {
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

	//create an entry in version history table
	installedAppVersionHistory := &repository.InstalledAppVersionHistory{}
	installedAppVersionHistory.InstalledAppVersionId = installedApp.InstalledAppVersionId
	installedAppVersionHistory.ValuesYamlRaw = installedApp.ValuesOverrideYaml
	installedAppVersionHistory.CreatedBy = installedApp.UserId
	installedAppVersionHistory.CreatedOn = time.Now()
	installedAppVersionHistory.UpdatedBy = installedApp.UserId
	installedAppVersionHistory.UpdatedOn = time.Now()
	installedAppVersionHistory.StartedOn = time.Now()
	installedAppVersionHistory.Status = pipelineConfig.WorkflowInProgress
	installedAppVersionHistory, err = impl.installedAppRepositoryHistory.CreateInstalledAppVersionHistory(installedAppVersionHistory, tx)
	if err != nil {
		impl.Logger.Errorw("error while fetching from db", "error", err)
		return installedApp, false, err
	}
	installedApp.InstalledAppVersionHistoryId = installedAppVersionHistory.Id

	//creating deployment started status timeline when mono repo migration is not required
	deploymentInitiatedTimeline := impl.pipelineStatusTimelineService.
		GetTimelineDbObjectByTimelineStatusAndTimelineDescription(0, installedApp.InstalledAppVersionHistoryId, pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED, "Deployment initiated successfully.", installedApp.UserId, time.Now())

	isAppStore := true
	err = impl.pipelineStatusTimelineService.SaveTimeline(deploymentInitiatedTimeline, tx, isAppStore)
	if err != nil {
		impl.Logger.Errorw("error in creating timeline status for deployment initiation for update of installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", installedApp.InstalledAppVersionHistoryId)
	}
	//If current version upgrade/degrade to another, update requirement dependencies
	if versionHistory.InstalledAppVersionId != activeInstalledAppVersion.Id {
		err = impl.updateRequirementYamlInGit(installedApp, &installedAppVersion.AppStoreApplicationVersion)
		if err != nil {
			if errors.Is(err, errors.New(pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED)) {
				impl.Logger.Errorw("error", "err", err)
				GitCommitFailTimeline := impl.pipelineStatusTimelineService.
					GetTimelineDbObjectByTimelineStatusAndTimelineDescription(0, installedApp.InstalledAppVersionHistoryId, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, "Git commit failed.", installedApp.UserId, time.Now())
				_ = impl.pipelineStatusTimelineService.SaveTimeline(GitCommitFailTimeline, tx, isAppStore)
			}
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
	installedApp, err = impl.updateValuesYamlInGit(installedApp)
	if err != nil {
		impl.Logger.Errorw("error", "err", err)
		if errors.Is(err, errors.New(pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED)) {
			GitCommitFailTimeline := impl.pipelineStatusTimelineService.
				GetTimelineDbObjectByTimelineStatusAndTimelineDescription(0, installedApp.InstalledAppVersionHistoryId, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, "Git commit failed.", installedApp.UserId, time.Now())
			_ = impl.pipelineStatusTimelineService.SaveTimeline(GitCommitFailTimeline, tx, isAppStore)
		}
		return installedApp, false, nil
	}
	installedAppVersionHistory.GitHash = installedApp.GitHash
	_, err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(installedAppVersionHistory, tx)
	if err != nil {
		impl.Logger.Errorw("error in updating installed app version history repository", "err", err)
		return installedApp, false, err
	}

	isManualSync := !impl.acdConfig.ArgoCDAutoSyncEnabled

	GitCommitSuccessTimeline := impl.pipelineStatusTimelineService.
		GetTimelineDbObjectByTimelineStatusAndTimelineDescription(0, installedApp.InstalledAppVersionHistoryId, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.", installedApp.UserId, time.Now())
	timelines := []*pipelineConfig.PipelineStatusTimeline{GitCommitSuccessTimeline}
	if isManualSync {
		// add ARGOCD_SYNC_INITIATED timeline if manual sync
		ArgocdSyncInitiatedTimeline := impl.pipelineStatusTimelineService.
			GetTimelineDbObjectByTimelineStatusAndTimelineDescription(0, installedApp.InstalledAppVersionHistoryId, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_INITIATED, "ArgoCD sync initiated.", installedApp.UserId, time.Now())
		timelines = append(timelines, ArgocdSyncInitiatedTimeline)
	}
	err = impl.pipelineStatusTimelineService.SaveTimelines(timelines, tx)
	if err != nil {
		impl.Logger.Errorw("error in creating timeline status for deployment initiation for update of installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", installedApp.InstalledAppVersionHistoryId)
	}

	err = impl.argoClientWrapperService.SyncArgoCDApplicationIfNeededAndRefresh(ctx, installedApp.ACDAppName)
	if err != nil {
		impl.Logger.Errorw("error in getting the argo application with normal refresh", "err", err)
		return installedApp, true, nil
	}
	syncTime := time.Now()
	if isManualSync {
		ArgocdSyncCompletedTimeline := impl.pipelineStatusTimelineService.
			GetTimelineDbObjectByTimelineStatusAndTimelineDescription(0, installedApp.InstalledAppVersionHistoryId, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED, "ArgoCD sync completed.", installedApp.UserId, syncTime)
		err = impl.pipelineStatusTimelineService.SaveTimeline(ArgocdSyncCompletedTimeline, tx, isAppStore)
		if err != nil {
			impl.Logger.Errorw("error in creating timeline status for deployment initiation for update of installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", installedApp.InstalledAppVersionHistoryId)
		}
	}
	//ACD sync operation
	//impl.appStoreDeploymentFullModeService.SyncACD(installedApp.ACDAppName, ctx)
	return installedApp, true, nil
}

func (impl *FullModeDeploymentServiceImpl) GetDeploymentHistory(ctx context.Context, installedAppDto *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error) {
	result := &gRPC.HelmAppDeploymentHistory{}
	var history []*gRPC.HelmAppDeploymentDetail
	//TODO - response setup

	installedAppVersions, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdMeta(installedAppDto.InstalledAppId)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, &util.ApiError{HttpStatusCode: http.StatusBadRequest, Code: "400", UserMessage: "values are outdated. please fetch the latest version and try again", InternalMessage: err.Error()}
		}
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
			emailId := "anonymous"
			user, err := impl.userService.GetByIdIncludeDeleted(updateHistory.CreatedBy)
			if err != nil && !util.IsErrNoRows(err) {
				impl.Logger.Errorw("error while fetching user Details", "error", err)
				return result, err
			}
			if user != nil {
				emailId = user.EmailId
			}
			history = append(history, &gRPC.HelmAppDeploymentDetail{
				ChartMetadata: &gRPC.ChartMetadata{
					ChartName:    installedAppVersionModel.AppStoreApplicationVersion.AppStore.Name,
					ChartVersion: installedAppVersionModel.AppStoreApplicationVersion.Version,
					Description:  installedAppVersionModel.AppStoreApplicationVersion.Description,
					Home:         installedAppVersionModel.AppStoreApplicationVersion.Home,
					Sources:      sources,
				},
				DeployedBy:   emailId,
				DockerImages: []string{installedAppVersionModel.AppStoreApplicationVersion.AppVersion},
				DeployedAt: &timestamp.Timestamp{
					Seconds: updateHistory.CreatedOn.Unix(),
					Nanos:   int32(updateHistory.CreatedOn.Nanosecond()),
				},
				Version: int32(updateHistory.Id),
				Status:  updateHistory.Status,
			})
		}
	}

	if len(history) == 0 {
		history = make([]*gRPC.HelmAppDeploymentDetail, 0)
	}
	result.DeploymentHistory = history
	return result, err
}

// TODO refactoring: use InstalledAppVersionHistoryId from appStoreBean.InstallAppVersionDTO instead of version int32
func (impl *FullModeDeploymentServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error) {
	values := &openapi.HelmAppDeploymentManifestDetail{}
	_, span := otel.Tracer("orchestrator").Start(ctx, "installedAppRepositoryHistory.GetInstalledAppVersionHistory")
	versionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(int(version))
	span.End()
	if err != nil {
		impl.Logger.Errorw("error while fetching installed version history", "error", err)
		return nil, err
	}
	values.ValuesYaml = &versionHistory.ValuesYamlRaw

	envId := int32(installedApp.EnvironmentId)
	clusterId := int32(installedApp.ClusterId)
	appStoreApplicationVersionId, err := impl.installedAppRepositoryHistory.GetAppStoreApplicationVersionIdByInstalledAppVersionHistoryId(int(version))
	appStoreVersionId := pointer.Int32(int32(appStoreApplicationVersionId))

	// as virtual environment doesn't exist on actual cluster, we will use default cluster for running helm template command
	if installedApp.IsVirtualEnvironment {
		clusterId = appStoreBean.DEFAULT_CLUSTER_ID
		installedApp.Namespace = appStoreBean.DEFAULT_NAMESPACE
	}

	manifestRequest := openapi2.TemplateChartRequest{
		EnvironmentId:                &envId,
		ClusterId:                    &clusterId,
		Namespace:                    &installedApp.Namespace,
		ReleaseName:                  &installedApp.AppName,
		AppStoreApplicationVersionId: appStoreVersionId,
		ValuesYaml:                   values.ValuesYaml,
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "helmAppService.TemplateChart")
	templateChart, manifestErr := impl.helmAppService.TemplateChart(ctx, &manifestRequest)
	span.End()
	manifest := templateChart.GetManifest()

	if manifestErr != nil {
		impl.Logger.Errorw("error in genetating manifest for argocd app", "err", manifestErr)
	} else {
		values.Manifest = &manifest
	}

	return values, err
}

func (impl *FullModeDeploymentServiceImpl) getSourcesFromManifest(chartYaml string) ([]string, error) {
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
