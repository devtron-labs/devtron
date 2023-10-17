package appStoreDeploymentGitopsTool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/constants"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	appStoreDeploymentFullMode "github.com/devtron-labs/devtron/pkg/appStore/deployment/fullMode"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/go-pg/pg"
	"github.com/golang/protobuf/ptypes/timestamp"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"k8s.io/utils/pointer"
	"net/http"
	"strings"
	"time"
)

// creating duplicates because cannot use

type AppStoreDeploymentArgoCdService interface {
	//InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) (*appStoreBean.InstallAppVersionDTO, error)
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

type AppStoreDeploymentArgoCdServiceImpl struct {
	Logger                               *zap.SugaredLogger
	appStoreDeploymentFullModeService    appStoreDeploymentFullMode.AppStoreDeploymentFullModeService
	acdClient                            application2.ServiceClient
	chartGroupDeploymentRepository       repository.ChartGroupDeploymentRepository
	installedAppRepository               repository.InstalledAppRepository
	installedAppRepositoryHistory        repository.InstalledAppVersionHistoryRepository
	chartTemplateService                 util.ChartTemplateService
	gitFactory                           *util.GitFactory
	argoUserService                      argo.ArgoUserService
	appStoreDeploymentCommonService      appStoreDeploymentCommon.AppStoreDeploymentCommonService
	helmAppService                       client.HelmAppService
	gitOpsConfigRepository               repository3.GitOpsConfigRepository
	appStatusService                     appStatus.AppStatusService
	pipelineStatusTimelineService        status.PipelineStatusTimelineService
	userService                          user.UserService
	pipelineStatusTimelineRepository     pipelineConfig.PipelineStatusTimelineRepository
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
}

func NewAppStoreDeploymentArgoCdServiceImpl(logger *zap.SugaredLogger, appStoreDeploymentFullModeService appStoreDeploymentFullMode.AppStoreDeploymentFullModeService,
	acdClient application2.ServiceClient, chartGroupDeploymentRepository repository.ChartGroupDeploymentRepository,
	installedAppRepository repository.InstalledAppRepository, installedAppRepositoryHistory repository.InstalledAppVersionHistoryRepository, chartTemplateService util.ChartTemplateService,
	gitFactory *util.GitFactory, argoUserService argo.ArgoUserService, appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	helmAppService client.HelmAppService, gitOpsConfigRepository repository3.GitOpsConfigRepository, appStatusService appStatus.AppStatusService,
	pipelineStatusTimelineService status.PipelineStatusTimelineService, userService user.UserService,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
) *AppStoreDeploymentArgoCdServiceImpl {
	return &AppStoreDeploymentArgoCdServiceImpl{
		Logger:                               logger,
		appStoreDeploymentFullModeService:    appStoreDeploymentFullModeService,
		acdClient:                            acdClient,
		chartGroupDeploymentRepository:       chartGroupDeploymentRepository,
		installedAppRepository:               installedAppRepository,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
		chartTemplateService:                 chartTemplateService,
		gitFactory:                           gitFactory,
		argoUserService:                      argoUserService,
		appStoreDeploymentCommonService:      appStoreDeploymentCommonService,
		helmAppService:                       helmAppService,
		gitOpsConfigRepository:               gitOpsConfigRepository,
		appStatusService:                     appStatusService,
		pipelineStatusTimelineService:        pipelineStatusTimelineService,
		userService:                          userService,
		pipelineStatusTimelineRepository:     pipelineStatusTimelineRepository,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
	}
}

// UpdateChartInfo this will update chart info in acd app, needed when repo for an app is changed
func (impl AppStoreDeploymentArgoCdServiceImpl) UpdateChartInfo(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ChartGitAttribute *util.ChartGitAttribute, installedAppVersionHistoryId int, ctx context.Context) error {
	installAppVersionRequest, err := impl.patchAcdApp(ctx, installAppVersionRequest, ChartGitAttribute)
	if err != nil {
		return err
	}
	return nil
}

func (impl AppStoreDeploymentArgoCdServiceImpl) SaveTimelineForACDHelmApps(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, status string, statusDetail string, tx *pg.Tx) error {

	if !util.IsAcdApp(installAppVersionRequest.DeploymentAppType) && !util.IsManifestDownload(installAppVersionRequest.DeploymentAppType) {
		return nil
	}
	timeline := &pipelineConfig.PipelineStatusTimeline{
		InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
		Status:                       status,
		StatusDetail:                 statusDetail,
		StatusTime:                   time.Now(),
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

func (impl AppStoreDeploymentArgoCdServiceImpl) InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, chartGitAttr *util.ChartGitAttribute, ctx context.Context, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {

	installAppVersionRequest, err := impl.appStoreDeploymentFullModeService.AppStoreDeployOperationACD(installAppVersionRequest, chartGitAttr, ctx)
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
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.Logger.Errorw("error in getting acd token", "err", err)
			return "", err
		}
		ctx = context.WithValue(ctx, "token", acdToken)
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
		//use this resp.Status to update app_status table
		err = impl.appStatusService.UpdateStatusWithAppIdEnvId(installedAppAndEnvDetails.AppId, installedAppAndEnvDetails.EnvironmentId, resp.Status)
		if err != nil {
			impl.Logger.Warnw("error in updating app status", "err", err, installedAppAndEnvDetails.AppId, "envId", installedAppAndEnvDetails.EnvironmentId)
		}
		return resp.Status, nil
	}
	return "", errors.New("invalid app name or env name")
}

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
	_, err = impl.installedAppRepositoryHistory.CreateInstalledAppVersionHistory(installedAppVersionHistory, tx)
	if err != nil {
		impl.Logger.Errorw("error while fetching from db", "error", err)
		return installedApp, false, err
	}
	installedApp.InstalledAppVersionHistoryId = installedAppVersionHistory.Id

	//creating deployment started status timeline when mono repo migration is not required
	timeline := &pipelineConfig.PipelineStatusTimeline{
		InstalledAppVersionHistoryId: installedApp.InstalledAppVersionHistoryId,
		Status:                       pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED,
		StatusDetail:                 "Deployment initiated successfully.",
		StatusTime:                   time.Now(),
		AuditLog: sql.AuditLog{
			CreatedBy: installedApp.UserId,
			CreatedOn: time.Now(),
			UpdatedBy: installedApp.UserId,
			UpdatedOn: time.Now(),
		},
	}
	isAppStore := true
	err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, tx, isAppStore)
	if err != nil {
		impl.Logger.Errorw("error in creating timeline status for deployment initiation for update of installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", installedApp.InstalledAppVersionHistoryId)
	}
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
	installedApp, err = impl.appStoreDeploymentFullModeService.UpdateValuesYaml(installedApp, tx)
	if err != nil {
		impl.Logger.Errorw("error", "err", err)
		return installedApp, false, nil
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
			return nil, fmt.Errorf("values are outdated. please fetch the latest version and try again")
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

func (impl *AppStoreDeploymentArgoCdServiceImpl) OnUpdateRepoInInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	//creating deployment started status timeline
	timeline := &pipelineConfig.PipelineStatusTimeline{
		InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
		Status:                       pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED,
		StatusDetail:                 "Deployment initiated successfully.",
		StatusTime:                   time.Now(),
		AuditLog: sql.AuditLog{
			CreatedBy: installAppVersionRequest.UserId,
			CreatedOn: time.Now(),
			UpdatedBy: installAppVersionRequest.UserId,
			UpdatedOn: time.Now(),
		},
	}
	err := impl.pipelineStatusTimelineService.SaveTimeline(timeline, tx, true)
	if err != nil {
		impl.Logger.Errorw("error in creating timeline status for deployment initiation for update of installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", installAppVersionRequest.InstalledAppVersionHistoryId)
	}
	//git operation pull push
	appStoreGitOpsResponse, err := impl.appStoreDeploymentCommonService.GenerateManifestAndPerformGitOperations(installAppVersionRequest)
	if err != nil {
		impl.Logger.Errorw("error in doing gitops operation", "err", err)
		_ = impl.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, fmt.Sprintf("Git commit failed - %v", err), tx)

	}

	_ = impl.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.", tx)

	//acd operation register, sync
	installAppVersionRequest, err = impl.patchAcdApp(ctx, installAppVersionRequest, appStoreGitOpsResponse.ChartGitAttribute)
	if err != nil {
		return installAppVersionRequest, err
	}

	return installAppVersionRequest, nil
}

func (impl *AppStoreDeploymentArgoCdServiceImpl) UpdateRequirementDependencies(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error {
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
}

func (impl AppStoreDeploymentArgoCdServiceImpl) UpdateInstalledApp(ctx context.Context, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, environment *clusterRepository.Environment, installedAppVersion *repository.InstalledAppVersions, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {
	//creating deployment started status timeline when mono repo migration is not required
	timeline := &pipelineConfig.PipelineStatusTimeline{
		InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
		Status:                       pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED,
		StatusDetail:                 "Deployment initiated successfully.",
		StatusTime:                   time.Now(),
		AuditLog: sql.AuditLog{
			CreatedBy: installAppVersionRequest.UserId,
			CreatedOn: time.Now(),
			UpdatedBy: installAppVersionRequest.UserId,
			UpdatedOn: time.Now(),
		},
	}
	err := impl.pipelineStatusTimelineService.SaveTimeline(timeline, tx, true)
	if err != nil {
		impl.Logger.Errorw("error in creating timeline status for deployment initiation for update of installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", installAppVersionRequest.InstalledAppVersionHistoryId)
	}
	//update values yaml in chart
	installAppVersionRequest, err = impl.updateValuesYaml(environment, installedAppVersion, installAppVersionRequest, tx)
	if err != nil {
		impl.Logger.Errorw("error while commit values to git", "error", err)
		noTargetFound, _ := impl.appStoreDeploymentCommonService.ParseGitRepoErrorResponse(err)
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
	installAppVersionRequest.Environment = environment

	//ACD sync operation
	//impl.appStoreDeploymentFullModeService.SyncACD(installAppVersionRequest.ACDAppName, ctx)

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
	patchReq := v1alpha1.Application{Spec: v1alpha1.ApplicationSpec{Source: v1alpha1.ApplicationSource{Path: chartGitAttr.ChartLocation, RepoURL: chartGitAttr.RepoUrl, TargetRevision: "master"}}}
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

func (impl AppStoreDeploymentArgoCdServiceImpl) updateValuesYaml(environment *clusterRepository.Environment, installedAppVersion *repository.InstalledAppVersions,
	installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreBean.InstallAppVersionDTO, error) {

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
	gitHash, err := impl.appStoreDeploymentCommonService.CommitConfigToGit(valuesGitConfig)
	if err != nil {
		impl.Logger.Errorw("error in git commit", "err", err)
		_ = impl.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED, fmt.Sprintf("Git commit failed - %v", err), tx)
		return nil, err
	}
	_ = impl.SaveTimelineForACDHelmApps(installAppVersionRequest, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.", tx)
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

func (impl AppStoreDeploymentArgoCdServiceImpl) UpdateInstalledAppAndPipelineStatusForFailedDeploymentStatus(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, triggeredAt time.Time, err error) error {
	if err != nil {
		terminalStatusExists, timelineErr := impl.pipelineStatusTimelineRepository.CheckIfTerminalStatusTimelinePresentByInstalledAppVersionHistoryId(installAppVersionRequest.InstalledAppVersionHistoryId)
		if timelineErr != nil {
			impl.Logger.Errorw("error in checking if terminal status timeline exists by installedAppVersionHistoryId", "err", timelineErr, "installedAppVersionHistoryId", installAppVersionRequest.InstalledAppVersionHistoryId)
			return timelineErr
		}
		if !terminalStatusExists {
			impl.Logger.Infow("marking pipeline deployment failed", "err", err)
			timeline := &pipelineConfig.PipelineStatusTimeline{
				InstalledAppVersionHistoryId: installAppVersionRequest.InstalledAppVersionHistoryId,
				Status:                       pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_FAILED,
				StatusDetail:                 fmt.Sprintf("Deployment failed: %v", err),
				StatusTime:                   time.Now(),
				AuditLog: sql.AuditLog{
					CreatedBy: 1,
					CreatedOn: time.Now(),
					UpdatedBy: 1,
					UpdatedOn: time.Now(),
				},
			}
			isAppStore := true
			timelineErr = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, isAppStore)
			if timelineErr != nil {
				impl.Logger.Errorw("error in creating timeline status for deployment fail", "err", timelineErr, "timeline", timeline)
			}
		}
		impl.Logger.Errorw("error in triggering installed application deployment, setting status as fail ", "versionHistoryId", installAppVersionRequest.InstalledAppVersionHistoryId, "err", err)

		installedAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(installAppVersionRequest.InstalledAppVersionHistoryId)
		if err != nil {
			impl.Logger.Errorw("error in getting installedAppVersionHistory by installedAppVersionHistoryId", "installedAppVersionHistoryId", installAppVersionRequest.InstalledAppVersionHistoryId, "err", err)
			return err
		}
		installedAppVersionHistory.Status = pipelineConfig.WorkflowFailed
		installedAppVersionHistory.FinishedOn = triggeredAt
		installedAppVersionHistory.UpdatedOn = time.Now()
		installedAppVersionHistory.UpdatedBy = installAppVersionRequest.UserId
		_, err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(installedAppVersionHistory, nil)
		if err != nil {
			impl.Logger.Errorw("error updating installed app version history status", "err", err, "installedAppVersionHistory", installedAppVersionHistory)
			return err
		}

	} else {
		//update [n,n-1] statuses as failed if not terminal
		terminalStatus := []string{string(health.HealthStatusHealthy), pipelineConfig.WorkflowAborted, pipelineConfig.WorkflowFailed, pipelineConfig.WorkflowSucceeded}
		previousNonTerminalHistory, err := impl.installedAppRepositoryHistory.FindPreviousInstalledAppVersionHistoryByStatus(installAppVersionRequest.Id, installAppVersionRequest.InstalledAppVersionHistoryId, terminalStatus)
		if err != nil {
			impl.Logger.Errorw("error fetching previous installed app version history, updating installed app version history status,", "err", err, "installAppVersionRequest", installAppVersionRequest)
			return err
		} else if len(previousNonTerminalHistory) == 0 {
			impl.Logger.Errorw("no previous history found in updating installedAppVersionHistory status,", "err", err, "installAppVersionRequest", installAppVersionRequest)
			return nil
		}
		dbConnection := impl.installedAppRepositoryHistory.GetConnection()
		tx, err := dbConnection.Begin()
		if err != nil {
			impl.Logger.Errorw("error on update status, txn begin failed", "err", err)
			return err
		}
		// Rollback tx on error.
		defer tx.Rollback()
		var timelines []*pipelineConfig.PipelineStatusTimeline
		for _, previousHistory := range previousNonTerminalHistory {
			if previousHistory.Status == string(health.HealthStatusHealthy) ||
				previousHistory.Status == pipelineConfig.WorkflowSucceeded ||
				previousHistory.Status == pipelineConfig.WorkflowAborted ||
				previousHistory.Status == pipelineConfig.WorkflowFailed {
				//terminal status return
				impl.Logger.Infow("skip updating installedAppVersionHistory status as previous history status is", "status", previousHistory.Status)
				continue
			}
			impl.Logger.Infow("updating installedAppVersionHistory status as previous runner status is", "status", previousHistory.Status)
			previousHistory.FinishedOn = triggeredAt
			previousHistory.Status = pipelineConfig.WorkflowFailed
			previousHistory.UpdatedOn = time.Now()
			previousHistory.UpdatedBy = installAppVersionRequest.UserId
			timeline := &pipelineConfig.PipelineStatusTimeline{
				InstalledAppVersionHistoryId: previousHistory.Id,
				Status:                       pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_SUPERSEDED,
				StatusDetail:                 "This deployment is superseded.",
				StatusTime:                   time.Now(),
				AuditLog: sql.AuditLog{
					CreatedBy: 1,
					CreatedOn: time.Now(),
					UpdatedBy: 1,
					UpdatedOn: time.Now(),
				},
			}
			timelines = append(timelines, timeline)
		}

		err = impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistoryWithTxn(previousNonTerminalHistory, tx)
		if err != nil {
			impl.Logger.Errorw("error updating cd wf runner status", "err", err, "previousNonTerminalHistory", previousNonTerminalHistory)
			return err
		}
		err = impl.pipelineStatusTimelineRepository.SaveTimelinesWithTxn(timelines, tx)
		if err != nil {
			impl.Logger.Errorw("error updating pipeline status timelines", "err", err, "timelines", timelines)
			return err
		}
		err = tx.Commit()
		if err != nil {
			impl.Logger.Errorw("error in db transaction commit", "err", err)
			return err
		}
	}
	return nil
}
