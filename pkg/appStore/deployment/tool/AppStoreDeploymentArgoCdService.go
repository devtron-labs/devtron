package appStoreDeploymentTool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"net/http"
	"strings"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	client "github.com/devtron-labs/devtron/api/helm-app"
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
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	appStoreDeploymentFullMode "github.com/devtron-labs/devtron/pkg/appStore/deployment/fullMode"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
)

// creating duplicates because cannot use

type AppStoreDeploymentArgoCdService interface {
	//InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error)
	DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error
	RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, deploymentVersion int32, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, bool, error)
	GetDeploymentHistory(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*client.HelmAppDeploymentHistory, error)
	GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error)
	GetGitOpsRepoName(appName string, environmentName string) (string, error)
	DeleteDeploymentApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error
	UpdateInstalledAppAndPipelineStatusForFailedDeploymentStatus(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error
	SaveTimelineForACDHelmApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, status string, statusDetail string, statusTime time.Time, tx *pg.Tx) error
	UpdateAndSyncACDApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *commonBean.ChartGitAttribute, isMonoRepoMigrationRequired bool, ctx context.Context, tx *pg.Tx) error
}

type AppStoreDeploymentArgoCdServiceImpl struct {
	Logger                               *zap.SugaredLogger
	appStoreDeploymentFullModeService    appStoreDeploymentFullMode.AppStoreDeploymentFullModeService
	acdClient                            application2.ServiceClient
	chartGroupDeploymentRepository       repository2.ChartGroupDeploymentRepository
	installedAppRepository               repository.InstalledAppRepository
	installedAppRepositoryHistory        repository.InstalledAppVersionHistoryRepository
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
}

func NewAppStoreDeploymentArgoCdServiceImpl(logger *zap.SugaredLogger, appStoreDeploymentFullModeService appStoreDeploymentFullMode.AppStoreDeploymentFullModeService,
	acdClient application2.ServiceClient, chartGroupDeploymentRepository repository2.ChartGroupDeploymentRepository,
	installedAppRepository repository.InstalledAppRepository, installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository,
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	helmAppService client.HelmAppService, appStatusService appStatus.AppStatusService,
	pipelineStatusTimelineService status.PipelineStatusTimelineService, userService user.UserService,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	argoClientWrapperService argocdServer.ArgoClientWrapperService, acdConfig *argocdServer.ACDConfig,
	gitOperationService git.GitOperationService) *AppStoreDeploymentArgoCdServiceImpl {
	return &AppStoreDeploymentArgoCdServiceImpl{
		Logger:                               logger,
		appStoreDeploymentFullModeService:    appStoreDeploymentFullModeService,
		acdClient:                            acdClient,
		chartGroupDeploymentRepository:       chartGroupDeploymentRepository,
		installedAppRepository:               installedAppRepository,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
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
	}
}

// UpdateAndSyncACDApps this will update chart info in acd app if required in case of mono repo migration and will refresh argo app
func (impl AppStoreDeploymentArgoCdServiceImpl) UpdateAndSyncACDApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *commonBean.ChartGitAttribute, isMonoRepoMigrationRequired bool, ctx context.Context, tx *pg.Tx) error {
	if isMonoRepoMigrationRequired {
		// update repo details on ArgoCD as repo is changed
		err := impl.UpdateChartInfo(installAppVersionRequest, ChartGitAttribute, 0, ctx)
		if err != nil {
			impl.Logger.Errorw("error in acd patch request", "err", err)
			return err
		}
	}
	acdAppName := installAppVersionRequest.ACDAppName
	argoApplication, err := impl.acdClient.Get(ctx, &application.ApplicationQuery{Name: &acdAppName})
	if err != nil {
		impl.Logger.Errorw("Service err:UpdateAndSyncACDApps - error in acd app by name", "acdAppName", acdAppName, "err", err)
		return err
	}

	err = impl.argoClientWrapperService.UpdateArgoCDSyncModeIfNeeded(ctx, argoApplication)
	if err != nil {
		impl.Logger.Errorw("error in updating argocd sync mode", "err", err)
		return err
	}
	syncTime := time.Now()
	err = impl.argoClientWrapperService.SyncArgoCDApplicationIfNeededAndRefresh(ctx, acdAppName)
	if err != nil {
		impl.Logger.Errorw("error in getting argocd application with normal refresh", "err", err, "argoAppName", installAppVersionRequest.ACDAppName)
		return err
	}
	if !impl.acdConfig.ArgoCDAutoSyncEnabled {
		err = impl.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_COMPLETED, "argocd sync completed", syncTime, tx)
		if err != nil {
			impl.Logger.Errorw("error in saving timeline for acd helm apps", "err", err)
			return err
		}
	}
	return nil
}

// UpdateChartInfo this will update chart info in acd app, needed when repo for an app is changed
func (impl AppStoreDeploymentArgoCdServiceImpl) UpdateChartInfo(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *commonBean.ChartGitAttribute, installedAppVersionHistoryId int, ctx context.Context) error {
	installAppVersionRequest, err := impl.patchAcdApp(ctx, installAppVersionRequest, ChartGitAttribute)
	if err != nil {
		return err
	}
	return nil
}

func (impl AppStoreDeploymentArgoCdServiceImpl) SaveTimelineForACDHelmApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, status string, statusDetail string, statusTime time.Time, tx *pg.Tx) error {

	if !util.IsAcdApp(installAppVersionRequest.DeploymentAppType) && !util.IsManifestDownload(installAppVersionRequest.DeploymentAppType) {
		return nil
	}

	timeline := &pipelineConfig.PipelineStatusTimeline{
		InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
		Status:                       status,
		StatusDetail:                 statusDetail,
		StatusTime:                   statusTime,
		AuditLog: sql.AuditLog{
			CreatedBy: installAppVersionRequest.UserId,
			CreatedOn: time.Now(),
			UpdatedBy: installAppVersionRequest.UserId,
			UpdatedOn: time.Now(),
		},
	}
	timelineErr := impl.pipelineStatusTimelineService.SaveTimeline(timeline, tx, true)
	if timelineErr != nil {
		impl.Logger.Errorw("error in creating timeline status for git commit", "err", timelineErr, "timeline", timeline)
	}
	return timelineErr
}

func (impl AppStoreDeploymentArgoCdServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {

	installAppVersionRequest, err := impl.appStoreDeploymentFullModeService.AppStoreDeployOperationACD(installAppVersionRequest, chartGitAttr, ctx, tx)
	if err != nil {
		impl.Logger.Errorw(" error", "err", err)
		return installAppVersionRequest, err
	}
	return installAppVersionRequest, nil
}

//func (impl AppStoreDeploymentArgoCdServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error) {
//	//step 2 git operation pull push
//	installAppVersionRequest, chartGitAttr, err := impl.appStoreDeploymentFullModeService.AppStoreDeployOperationGIT(installAppVersionRequest)
//	if err != nil {
//		impl.Logger.Errorw(" error", "err", err)
//		return installAppVersionRequest, err
//	}
//	//step 3 acd operation register, sync
//	installAppVersionRequest, err = impl.appStoreDeploymentFullModeService.AppStoreDeployOperationACD(installAppVersionRequest, chartGitAttr, ctx)
//	if err != nil {
//		impl.Logger.Errorw(" error", "err", err)
//		return installAppVersionRequest, err
//	}
//	return installAppVersionRequest, nil
//}

func (impl AppStoreDeploymentArgoCdServiceImpl) DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *repository.InstalledApps, dbTransaction *pg.Tx) error {

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

// returns - valuesYamlStr, success, error
func (impl AppStoreDeploymentArgoCdServiceImpl) RollbackRelease(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, installedAppVersionHistoryId int32, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, bool, error) {
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
		err = impl.appStoreDeploymentFullModeService.UpdateRequirementYaml(installedApp, &installedAppVersion.AppStoreApplicationVersion)
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
	installedApp, err = impl.appStoreDeploymentFullModeService.UpdateValuesYaml(installedApp, tx)
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

func (impl AppStoreDeploymentArgoCdServiceImpl) deleteACD(acdAppName string, ctx context.Context, isNonCascade bool) error {
	req := new(application.ApplicationDeleteRequest)
	req.Name = &acdAppName
	cascadeDelete := !isNonCascade
	req.Cascade = &cascadeDelete
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
			history = append(history, &client.HelmAppDeploymentDetail{
				ChartMetadata: &client.ChartMetadata{
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
		history = make([]*client.HelmAppDeploymentDetail, 0)
	}
	result.DeploymentHistory = history
	return result, err
}

func (impl AppStoreDeploymentArgoCdServiceImpl) GetDeploymentHistoryInfo(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error) {
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

func (impl AppStoreDeploymentArgoCdServiceImpl) GetGitOpsRepoName(appName string, environmentName string) (string, error) {
	return impl.appStoreDeploymentFullModeService.GetGitOpsRepoName(appName, environmentName)
}

func (impl AppStoreDeploymentArgoCdServiceImpl) patchAcdApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *commonBean.ChartGitAttribute) (*appStoreBean.InstallAppVersionDTO, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()
	//registerInArgo
	err := impl.argoClientWrapperService.RegisterGitOpsRepoInArgo(ctx, chartGitAttr.RepoUrl)
	if err != nil {
		impl.Logger.Errorw("error in argo registry", "err", err)
		return nil, err
	}
	// update acd app
	patchReq := v1alpha1.Application{Spec: v1alpha1.ApplicationSpec{Source: &v1alpha1.ApplicationSource{Path: chartGitAttr.ChartLocation, RepoURL: chartGitAttr.RepoUrl, TargetRevision: "master"}}}
	reqbyte, err := json.Marshal(patchReq)
	if err != nil {
		impl.Logger.Errorw("error in creating patch", "err", err)
	}
	reqString := string(reqbyte)
	patchType := "merge"
	_, err = impl.acdClient.Patch(ctx, &application.ApplicationPatchRequest{Patch: &reqString, Name: &installAppVersionRequest.ACDAppName, PatchType: &patchType})
	if err != nil {
		impl.Logger.Errorw("error in creating argo app ", "name", installAppVersionRequest.ACDAppName, "patch", string(reqbyte), "err", err)
		return nil, err
	}
	//impl.appStoreDeploymentFullModeService.SyncACD(installAppVersionRequest.ACDAppName, ctx)
	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentArgoCdServiceImpl) updateValuesYaml(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {

	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.Logger.Errorw("fetching error", "err", err)
		return nil, err
	}
	valuesString, err := impl.appStoreDeploymentCommonService.GetValuesString(appStoreAppVersion.Name, installAppVersionRequest.ValuesOverrideYaml)
	if err != nil {
		impl.Logger.Errorw("error in building requirements config for helm app", "err", err)
		return nil, err
	}
	valuesGitConfig, err := impl.appStoreDeploymentCommonService.GetGitCommitConfig(installAppVersionRequest, valuesString, appStoreBean.VALUES_YAML_FILE)
	if err != nil {
		impl.Logger.Errorw("error in getting git config for helm app", "err", err)
		return nil, err
	}
	gitHash, _, err := impl.gitOperationService.CommitValues(valuesGitConfig)
	if err != nil {
		impl.Logger.Errorw("error in git commit", "err", err)
		_ = impl.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, fmt.Sprintf("Git commit failed - %v", err), time.Now(), tx)
		return nil, err
	}
	_ = impl.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.", time.Now(), tx)
	//update timeline status for git commit state
	installAppVersionRequest.GitHash = gitHash
	return installAppVersionRequest, nil
}

func (impl AppStoreDeploymentArgoCdServiceImpl) DeleteDeploymentApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) error {
	acdAppName := appName + "-" + environmentName
	var err error
	err = impl.deleteACD(acdAppName, ctx, installAppVersionRequest.NonCascadeDelete)
	if err != nil {
		impl.Logger.Errorw("error in deleting ACD ", "name", acdAppName, "err", err)
		if installAppVersionRequest.ForceDelete {
			impl.Logger.Warnw("error while deletion of app in acd, continue to delete in db as this operation is force delete", "error", err)
		} else {
			//statusError, _ := err.(*errors2.StatusError)
			if !installAppVersionRequest.NonCascadeDelete && strings.Contains(err.Error(), "code = NotFound") {
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
	return nil
}
