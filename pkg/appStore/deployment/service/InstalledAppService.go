/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package service

import (
	"bytes"
	"context"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"

	/* #nosec */
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	util4 "github.com/devtron-labs/common-lib-private/utils/k8s"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/chartGroup"
	repository6 "github.com/devtron-labs/devtron/pkg/appStore/chartGroup/repository"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	appStoreDeploymentFullMode "github.com/devtron-labs/devtron/pkg/appStore/deployment/fullMode"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/values/service"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/k8s"
	application3 "github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository4 "github.com/devtron-labs/devtron/pkg/team"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"

	"github.com/Pallinder/go-randomdata"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

// DB operation + chart group + nats msg consume(to be removed)
type InstalledAppService interface {
	GetAll(filter *appStoreBean.AppStoreFilter) (appStoreBean.AppListDetail, error)
	DeployBulk(chartGroupInstallRequest *chartGroup.ChartGroupInstallRequest) (*chartGroup.ChartGroupInstallAppRes, error)
	CheckAppExists(appNames []*appStoreBean.AppNames) ([]*appStoreBean.AppNames, error)
	DeployDefaultChartOnCluster(bean *cluster2.ClusterBean, userId int32) (bool, error)
	FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean2.AppDetailContainer, error)
	UpdateInstalledAppVersionStatus(application *v1alpha1.Application) (bool, error)
	MarkGitOpsInstalledAppsDeletedIfArgoAppIsDeleted(installedAppId int, envId int) error
	CheckAppExistsByInstalledAppId(installedAppId int) (*repository2.InstalledApps, error)
	GetChartBytesForLatestDeployment(installedAppId int, installedAppVersionId int) ([]byte, error)
	GetChartBytesForParticularDeployment(installedAppId int, installedAppVersionId int, installedAppVersionHistoryId int) ([]byte, error)

	FetchResourceTreeWithHibernateForACD(rctx context.Context, cn http.CloseNotifier, appDetail *bean2.AppDetailContainer) bean2.AppDetailContainer
	FetchResourceTree(rctx context.Context, cn http.CloseNotifier, appDetailsContainer *bean2.AppDetailsContainer, installedApp repository2.InstalledApps, helmReleaseInstallStatus string, status string) error

	//move to notes service
	FetchChartNotes(installedAppId int, envId int, token string, checkNotesAuth func(token string, appName string, envId int) bool) (string, error)
}

type InstalledAppServiceImpl struct {
	logger                               *zap.SugaredLogger
	installedAppRepository               repository2.InstalledAppRepository
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	environmentRepository                repository5.EnvironmentRepository
	teamRepository                       repository4.TeamRepository
	appRepository                        app.AppRepository
	acdClient                            application2.ServiceClient
	appStoreValuesService                service.AppStoreValuesService
	pubsubClient                         *pubsub.PubSubClientServiceImpl
	chartGroupDeploymentRepository       repository6.ChartGroupDeploymentRepository
	envService                           cluster2.EnvironmentService
	aCDAuthConfig                        *util2.ACDAuthConfig
	userService                          user.UserService
	appStoreDeploymentService            AppStoreDeploymentService
	appStoreDeploymentFullModeService    appStoreDeploymentFullMode.AppStoreDeploymentFullModeService
	installedAppRepositoryHistory        repository2.InstalledAppVersionHistoryRepository
	argoUserService                      argo.ArgoUserService
	helmAppClient                        client.HelmAppClient
	helmAppService                       client.HelmAppService
	appStatusService                     appStatus.AppStatusService
	K8sUtil                              *util4.K8sUtilExtended
	pipelineStatusTimelineService        status.PipelineStatusTimelineService
	appStoreDeploymentCommonService      appStoreDeploymentCommon.AppStoreDeploymentCommonService
	k8sCommonService                     k8s.K8sCommonService
	k8sApplicationService                application3.K8sApplicationService
	acdConfig                            *argocdServer.ACDConfig
	gitOpsConfigReadService              config.GitOpsConfigReadService
	gitOperationService                  git.GitOperationService
}

func NewInstalledAppServiceImpl(logger *zap.SugaredLogger,
	installedAppRepository repository2.InstalledAppRepository,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	environmentRepository repository5.EnvironmentRepository, teamRepository repository4.TeamRepository,
	appRepository app.AppRepository, acdClient application2.ServiceClient,
	appStoreValuesService service.AppStoreValuesService, pubsubClient *pubsub.PubSubClientServiceImpl,
	chartGroupDeploymentRepository repository6.ChartGroupDeploymentRepository,
	envService cluster2.EnvironmentService, aCDAuthConfig *util2.ACDAuthConfig, userService user.UserService,
	appStoreDeploymentFullModeService appStoreDeploymentFullMode.AppStoreDeploymentFullModeService,
	appStoreDeploymentService AppStoreDeploymentService,
	installedAppRepositoryHistory repository2.InstalledAppVersionHistoryRepository,
	argoUserService argo.ArgoUserService, helmAppClient client.HelmAppClient, helmAppService client.HelmAppService,
	appStatusService appStatus.AppStatusService, K8sUtil *util4.K8sUtilExtended,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	k8sCommonService k8s.K8sCommonService, k8sApplicationService application3.K8sApplicationService,
	acdConfig *argocdServer.ACDConfig, gitOpsConfigReadService config.GitOpsConfigReadService,
	gitOperationService git.GitOperationService) (*InstalledAppServiceImpl, error) {
	impl := &InstalledAppServiceImpl{
		logger:                               logger,
		installedAppRepository:               installedAppRepository,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		environmentRepository:                environmentRepository,
		teamRepository:                       teamRepository,
		appRepository:                        appRepository,
		acdClient:                            acdClient,
		appStoreValuesService:                appStoreValuesService,
		pubsubClient:                         pubsubClient,
		chartGroupDeploymentRepository:       chartGroupDeploymentRepository,
		envService:                           envService,
		aCDAuthConfig:                        aCDAuthConfig,
		userService:                          userService,
		appStoreDeploymentService:            appStoreDeploymentService,
		appStoreDeploymentFullModeService:    appStoreDeploymentFullModeService,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
		argoUserService:                      argoUserService,
		helmAppClient:                        helmAppClient,
		helmAppService:                       helmAppService,
		appStatusService:                     appStatusService,
		K8sUtil:                              K8sUtil,
		pipelineStatusTimelineService:        pipelineStatusTimelineService,
		appStoreDeploymentCommonService:      appStoreDeploymentCommonService,
		k8sCommonService:                     k8sCommonService,
		k8sApplicationService:                k8sApplicationService,
		acdConfig:                            acdConfig,
		gitOpsConfigReadService:              gitOpsConfigReadService,
		gitOperationService:                  gitOperationService,
	}
	err := impl.subscribe()
	if err != nil {
		return nil, err
	}
	return impl, nil
}

func (impl InstalledAppServiceImpl) GetAll(filter *appStoreBean.AppStoreFilter) (appStoreBean.AppListDetail, error) {
	applicationType := "DEVTRON-CHART-STORE"
	var clusterIdsConverted []int32
	for _, clusterId := range filter.ClusterIds {
		clusterIdsConverted = append(clusterIdsConverted, int32(clusterId))
	}
	installedAppsResponse := appStoreBean.AppListDetail{
		ApplicationType: &applicationType,
		ClusterIds:      &clusterIdsConverted,
	}
	start := time.Now()
	installedApps, err := impl.installedAppRepository.GetAllInstalledApps(filter)
	middleware.AppListingDuration.WithLabelValues("getAllInstalledApps", "helm").Observe(time.Since(start).Seconds())
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return installedAppsResponse, err
	}
	var helmAppsResponse []appStoreBean.HelmAppDetails
	for _, a := range installedApps {
		appLocal := a // copied data from here because value is passed as reference
		if appLocal.TeamId == 0 && appLocal.AppOfferingMode != util3.SERVER_MODE_HYPERION {
			//skipping entries for empty projectId for non hyperion app (as app list should return the helm apps from installedApps)
			continue
		}
		appId := strconv.Itoa(appLocal.Id)
		projectId := int32(appLocal.TeamId)
		envId := int32(appLocal.EnvironmentId)
		clusterId := int32(appLocal.ClusterId)
		environmentDetails := appStoreBean.EnvironmentDetails{
			EnvironmentName:      &appLocal.EnvironmentName,
			EnvironmentId:        &envId,
			Namespace:            &appLocal.Namespace,
			ClusterName:          &appLocal.ClusterName,
			ClusterId:            &clusterId,
			IsVirtualEnvironment: &appLocal.IsVirtualEnvironment,
		}
		helmAppResp := appStoreBean.HelmAppDetails{
			AppName:           &appLocal.AppName,
			ChartName:         &appLocal.AppStoreApplicationName,
			AppId:             &appId,
			ProjectId:         &projectId,
			EnvironmentDetail: &environmentDetails,
			ChartAvatar:       &appLocal.Icon,
			LastDeployedAt:    &appLocal.UpdatedOn,
			AppStatus:         &appLocal.AppStatus,
		}
		helmAppsResponse = append(helmAppsResponse, helmAppResp)
	}
	installedAppsResponse.HelmApps = &helmAppsResponse
	return installedAppsResponse, nil
}

func (impl InstalledAppServiceImpl) CheckAppExists(appNames []*appStoreBean.AppNames) ([]*appStoreBean.AppNames, error) {
	if len(appNames) == 0 {
		return nil, nil
	}
	var names []string
	for _, appName := range appNames {
		names = append(names, appName.Name)
	}

	apps, err := impl.appRepository.CheckAppExists(names)
	if err != nil {
		return nil, err
	}
	existingApps := make(map[string]bool)
	for _, app := range apps {
		existingApps[app.AppName] = true
	}
	for _, appName := range appNames {
		if _, ok := existingApps[appName.Name]; ok {
			appName.Exists = true
			appName.SuggestedName = strings.ToLower(randomdata.SillyName())
		}
	}
	return appNames, nil
}

func (impl InstalledAppServiceImpl) DeployBulk(chartGroupInstallRequest *chartGroup.ChartGroupInstallRequest) (*chartGroup.ChartGroupInstallAppRes, error) {
	impl.logger.Debugw("bulk app install request", "req", chartGroupInstallRequest)
	//save in db
	// raise nats event

	var installAppVersionDTOList []*appStoreBean.InstallAppVersionDTO
	for _, chartGroupInstall := range chartGroupInstallRequest.ChartGroupInstallChartRequest {
		installAppVersionDTO, err := impl.requestBuilderForBulkDeployment(chartGroupInstall, chartGroupInstallRequest.ProjectId, chartGroupInstallRequest.UserId)
		if err != nil {
			impl.logger.Errorw("DeployBulk, error in request builder", "err", err)
			return nil, err
		}
		installAppVersionDTOList = append(installAppVersionDTOList, installAppVersionDTO)
	}
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	var installAppVersions []*appStoreBean.InstallAppVersionDTO
	// Rollback tx on error.
	defer tx.Rollback()
	for _, installAppVersionDTO := range installAppVersionDTOList {
		installAppVersionDTO, err = impl.appStoreDeploymentService.AppStoreDeployOperationDB(installAppVersionDTO, tx, false)
		if err != nil {
			impl.logger.Errorw("DeployBulk, error while app store deploy db operation", "err", err)
			return nil, err
		}
		installAppVersions = append(installAppVersions, installAppVersionDTO)
	}
	if chartGroupInstallRequest.ChartGroupId > 0 {
		groupINstallationId, err := impl.getInstallationId(installAppVersions)
		if err != nil {
			return nil, err
		}
		for _, installAppVersionDTO := range installAppVersions {
			chartGroupEntry := impl.createChartGroupEntryObject(installAppVersionDTO, chartGroupInstallRequest.ChartGroupId, groupINstallationId)
			err := impl.chartGroupDeploymentRepository.Save(tx, chartGroupEntry)
			if err != nil {
				impl.logger.Errorw("DeployBulk, error in creating ChartGroupEntryObject", "err", err)
				return nil, err
			}
		}
	}
	//commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("DeployBulk, error in tx commit", "err", err)
		return nil, err
	}
	//nats event
	impl.triggerDeploymentEvent(installAppVersions)
	return &chartGroup.ChartGroupInstallAppRes{}, nil
}

// generate unique installation ID using APPID
func (impl InstalledAppServiceImpl) getInstallationId(installAppVersions []*appStoreBean.InstallAppVersionDTO) (string, error) {
	var buffer bytes.Buffer
	for _, installAppVersionDTO := range installAppVersions {
		if installAppVersionDTO.AppId == 0 {
			return "", fmt.Errorf("app ID not present")
		}
		buffer.WriteString(
			strconv.Itoa(installAppVersionDTO.AppId))
	}
	/* #nosec */
	h := sha1.New()
	_, err := h.Write([]byte(buffer.String()))
	if err != nil {
		return "", err
	}
	bs := h.Sum(nil)
	return fmt.Sprintf("%x", bs), nil
}

func (impl InstalledAppServiceImpl) createChartGroupEntryObject(installAppVersionDTO *appStoreBean.InstallAppVersionDTO, chartGroupId int, groupINstallationId string) *repository6.ChartGroupDeployment {
	return &repository6.ChartGroupDeployment{
		ChartGroupId:        chartGroupId,
		ChartGroupEntryId:   installAppVersionDTO.ChartGroupEntryId,
		InstalledAppId:      installAppVersionDTO.InstalledAppId,
		Deleted:             false,
		GroupInstallationId: groupINstallationId,
		AuditLog: sql.AuditLog{
			CreatedOn: time.Now(),
			CreatedBy: installAppVersionDTO.UserId,
			UpdatedOn: time.Now(),
			UpdatedBy: installAppVersionDTO.UserId,
		},
	}
}

func (impl InstalledAppServiceImpl) performDeployStageOnAcd(installedAppVersion *appStoreBean.InstallAppVersionDTO, ctx context.Context, userId int32) (*appStoreBean.InstallAppVersionDTO, error) {
	installedAppVersion.ACDAppName = fmt.Sprintf("%s-%s", installedAppVersion.AppName, installedAppVersion.Environment.Name)
	chartGitAttr := &commonBean.ChartGitAttribute{}
	if installedAppVersion.Status == appStoreBean.DEPLOY_INIT ||
		installedAppVersion.Status == appStoreBean.ENQUEUED ||
		installedAppVersion.Status == appStoreBean.QUE_ERROR ||
		installedAppVersion.Status == appStoreBean.GIT_ERROR {
		//step 2 git operation pull push
		//TODO: save git Timeline here
		appStoreGitOpsResponse, err := impl.appStoreDeploymentCommonService.GenerateManifestAndPerformGitOperations(installedAppVersion)
		if err != nil {
			impl.logger.Errorw(" error", "err", err)
			_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.GIT_ERROR)
			if err != nil {
				impl.logger.Errorw(" error", "err", err)
				return nil, err
			}
			timeline := &pipelineConfig.PipelineStatusTimeline{
				InstalledAppVersionHistoryId: installedAppVersion.InstalledAppVersionHistoryId,
				Status:                       pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED,
				StatusDetail:                 fmt.Sprintf("Git commit failed - %v", err),
				StatusTime:                   time.Now(),
				AuditLog: sql.AuditLog{
					CreatedBy: installedAppVersion.UserId,
					CreatedOn: time.Now(),
					UpdatedBy: installedAppVersion.UserId,
					UpdatedOn: time.Now(),
				},
			}
			_ = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, true)
			return nil, err
		}

		timeline := &pipelineConfig.PipelineStatusTimeline{
			InstalledAppVersionHistoryId: installedAppVersion.InstalledAppVersionHistoryId,
			Status:                       pipelineConfig.TIMELINE_STATUS_GIT_COMMIT,
			StatusDetail:                 "Git commit done successfully.",
			StatusTime:                   time.Now(),
			AuditLog: sql.AuditLog{
				CreatedBy: installedAppVersion.UserId,
				CreatedOn: time.Now(),
				UpdatedBy: installedAppVersion.UserId,
				UpdatedOn: time.Now(),
			},
		}
		_ = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, true)
		impl.logger.Infow("GIT SUCCESSFUL", "chartGitAttrDB", appStoreGitOpsResponse)
		_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.GIT_SUCCESS)
		if err != nil {
			impl.logger.Errorw(" error", "err", err)
			return nil, err
		}

		GitCommitSuccessTimeline := impl.pipelineStatusTimelineService.
			GetTimelineDbObjectByTimelineStatusAndTimelineDescription(0, installedAppVersion.InstalledAppVersionHistoryId, pipelineConfig.TIMELINE_STATUS_GIT_COMMIT, "Git commit done successfully.", installedAppVersion.UserId, time.Now())

		timelines := []*pipelineConfig.PipelineStatusTimeline{GitCommitSuccessTimeline}
		if !impl.acdConfig.ArgoCDAutoSyncEnabled {
			ArgocdSyncInitiatedTimeline := impl.pipelineStatusTimelineService.
				GetTimelineDbObjectByTimelineStatusAndTimelineDescription(0, installedAppVersion.InstalledAppVersionHistoryId, pipelineConfig.TIMELINE_STATUS_ARGOCD_SYNC_INITIATED, "ArgoCD sync initiated.", installedAppVersion.UserId, time.Now())

			timelines = append(timelines, ArgocdSyncInitiatedTimeline)
		}

		dbConnection := impl.installedAppRepository.GetConnection()
		tx, err := dbConnection.Begin()
		if err != nil {
			impl.logger.Errorw("error in getting db connection for saving timelines", "err", err)
			return nil, err
		}
		err = impl.pipelineStatusTimelineService.SaveTimelines(timelines, tx)
		if err != nil {
			impl.logger.Errorw("error in creating timeline status for deployment initiation for update of installedAppVersionHistoryId", "err", err, "installedAppVersionHistoryId", installedAppVersion.InstalledAppVersionHistoryId)
		}
		tx.Commit()
		// update build history for chart for argo_cd apps
		err = impl.appStoreDeploymentService.UpdateInstalledAppVersionHistoryWithGitHash(installedAppVersion, nil)
		if err != nil {
			impl.logger.Errorw("error on updating history for chart deployment", "error", err, "installedAppVersion", installedAppVersion)
			return nil, err
		}
		installedAppVersion.GitHash = appStoreGitOpsResponse.GitHash
		chartGitAttr.RepoUrl = appStoreGitOpsResponse.ChartGitAttribute.RepoUrl
		chartGitAttr.ChartLocation = appStoreGitOpsResponse.ChartGitAttribute.ChartLocation
	} else {
		impl.logger.Infow("DB and GIT operation already done for this app and env, proceed for further step", "installedAppId", installedAppVersion.InstalledAppId, "existing status", installedAppVersion.Status)
		environment, err := impl.environmentRepository.FindById(installedAppVersion.EnvironmentId)
		if err != nil {
			impl.logger.Errorw("fetching error", "err", err)
			return nil, err
		}

		repoUrl, err := impl.gitOperationService.GetRepoUrlByRepoName(installedAppVersion.GitOpsRepoName)
		if err != nil {
			//will allow to continue to persist status on next operation
			impl.logger.Errorw("error, GetRepoUrlByRepoName", "err", err)
		}

		chartGitAttr.RepoUrl = repoUrl
		chartGitAttr.ChartLocation = fmt.Sprintf("%s-%s", installedAppVersion.AppName, environment.Name)
		installedAppVersion.ACDAppName = fmt.Sprintf("%s-%s", installedAppVersion.AppName, environment.Name)
		installedAppVersion.Environment = environment
	}

	if installedAppVersion.Status == appStoreBean.DEPLOY_INIT ||
		installedAppVersion.Status == appStoreBean.ENQUEUED ||
		installedAppVersion.Status == appStoreBean.QUE_ERROR ||
		installedAppVersion.Status == appStoreBean.GIT_ERROR ||
		installedAppVersion.Status == appStoreBean.GIT_SUCCESS ||
		installedAppVersion.Status == appStoreBean.ACD_ERROR {
		//step 3 acd operation register, sync
		_, err := impl.appStoreDeploymentFullModeService.AppStoreDeployOperationACD(installedAppVersion, chartGitAttr, ctx, nil)
		if err != nil {
			impl.logger.Errorw("error", "chartGitAttr", chartGitAttr, "err", err)
			_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.ACD_ERROR)
			if err != nil {
				impl.logger.Errorw("error", "err", err)
				return nil, err
			}
			return nil, err
		}
		impl.logger.Infow("ACD SUCCESSFUL", "chartGitAttr", chartGitAttr)
		_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.ACD_SUCCESS)
		if err != nil {
			impl.logger.Errorw("error", "err", err)
			return nil, err
		}
	} else {
		impl.logger.Infow("DB and GIT and ACD operation already done for this app and env. process has been completed", "installedAppId", installedAppVersion.InstalledAppId, "existing status", installedAppVersion.Status)
	}
	return installedAppVersion, nil
}

func (impl InstalledAppServiceImpl) performDeployStage(installedAppVersionId int, installedAppVersionHistoryId int, userId int32) (*appStoreBean.InstallAppVersionDTO, error) {
	ctx := context.Background()
	installedAppVersion, err := impl.appStoreDeploymentService.GetInstalledAppVersion(installedAppVersionId, userId)
	if err != nil {
		return nil, err
	}
	installedAppVersion.InstalledAppVersionHistoryId = installedAppVersionHistoryId
	if util.IsAcdApp(installedAppVersion.DeploymentAppType) {
		//this method should only call in case of argo-integration installed and git-ops has configured
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.logger.Errorw("error in getting acd token", "err", err)
			return nil, err
		}
		ctx = context.WithValue(ctx, "token", acdToken)
		timeline := &pipelineConfig.PipelineStatusTimeline{
			InstalledAppVersionHistoryId: installedAppVersion.InstalledAppVersionHistoryId,
			Status:                       pipelineConfig.TIMELINE_STATUS_DEPLOYMENT_INITIATED,
			StatusDetail:                 "Deployment initiated successfully.",
			StatusTime:                   time.Now(),
			AuditLog: sql.AuditLog{
				CreatedBy: installedAppVersion.UserId,
				CreatedOn: time.Now(),
				UpdatedBy: installedAppVersion.UserId,
				UpdatedOn: time.Now(),
			},
		}
		err = impl.pipelineStatusTimelineService.SaveTimeline(timeline, nil, true)
		if err != nil {
			impl.logger.Errorw("error in creating timeline status for deployment initiation for this app store application", "err", err, "timeline", timeline)
		}
		_, err = impl.performDeployStageOnAcd(installedAppVersion, ctx, userId)
		if err != nil {
			impl.logger.Errorw("error", "err", err)
			return nil, err
		}
	} else if util.IsHelmApp(installedAppVersion.DeploymentAppType) {

		_, err = impl.appStoreDeploymentService.InstallAppByHelm(installedAppVersion, ctx)
		if err != nil {
			impl.logger.Errorw("error", "err", err)
			_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.HELM_ERROR)
			if err != nil {
				impl.logger.Errorw("error", "err", err)
				return nil, err
			}
			return nil, err
		}
	}

	//step 4 db operation status triggered
	_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.DEPLOY_SUCCESS)
	if err != nil {
		impl.logger.Errorw("error", "err", err)
		return nil, err
	}

	return installedAppVersion, nil
}

func (impl InstalledAppServiceImpl) requestBuilderForBulkDeployment(installRequest *chartGroup.ChartGroupInstallChartRequest, projectId int, userId int32) (*appStoreBean.InstallAppVersionDTO, error) {
	valYaml := installRequest.ValuesOverrideYaml
	if valYaml == "" {
		valVersion, err := impl.appStoreValuesService.FindValuesByIdAndKind(installRequest.ReferenceValueId, installRequest.ReferenceValueKind)
		if err != nil {
			return nil, err
		}
		valYaml = valVersion.Values
	}
	req := &appStoreBean.InstallAppVersionDTO{
		AppName:                 installRequest.AppName,
		TeamId:                  projectId,
		EnvironmentId:           installRequest.EnvironmentId,
		AppStoreVersion:         installRequest.AppStoreVersion,
		ValuesOverrideYaml:      valYaml,
		UserId:                  userId,
		ReferenceValueId:        installRequest.ReferenceValueId,
		ReferenceValueKind:      installRequest.ReferenceValueKind,
		ChartGroupEntryId:       installRequest.ChartGroupEntryId,
		DefaultClusterComponent: installRequest.DefaultClusterComponent,
	}
	return req, nil
}

//------------ nats config

func (impl *InstalledAppServiceImpl) triggerDeploymentEvent(installAppVersions []*appStoreBean.InstallAppVersionDTO) {

	for _, versions := range installAppVersions {
		var status appStoreBean.AppstoreDeploymentStatus
		payload := &appStoreBean.DeployPayload{InstalledAppVersionId: versions.InstalledAppVersionId, InstalledAppVersionHistoryId: versions.InstalledAppVersionHistoryId}
		data, err := json.Marshal(payload)
		if err != nil {
			status = appStoreBean.QUE_ERROR
		} else {
			err = impl.pubsubClient.Publish(pubsub.BULK_APPSTORE_DEPLOY_TOPIC, string(data))
			if err != nil {
				impl.logger.Errorw("err while publishing msg for app-store bulk deploy", "msg", data, "err", err)
				status = appStoreBean.QUE_ERROR
			} else {
				status = appStoreBean.ENQUEUED
			}

		}
		if versions.Status == appStoreBean.DEPLOY_INIT || versions.Status == appStoreBean.QUE_ERROR || versions.Status == appStoreBean.ENQUEUED {
			impl.logger.Debugw("status for bulk app-store deploy", "status", status)
			_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(payload.InstalledAppVersionId, status)
			if err != nil {
				impl.logger.Errorw("error while bulk app-store deploy status update", "err", err)
			}
		}
	}
}

func (impl *InstalledAppServiceImpl) subscribe() error {
	callback := func(msg *model.PubSubMsg) {
		deployPayload := &appStoreBean.DeployPayload{}
		err := json.Unmarshal([]byte(string(msg.Data)), &deployPayload)
		if err != nil {
			impl.logger.Error("Error while unmarshalling deployPayload json object", "error", err)
			return
		}
		impl.logger.Debugw("deployPayload:", "deployPayload", deployPayload)
		//using userId 1 - for system user
		_, err = impl.performDeployStage(deployPayload.InstalledAppVersionId, deployPayload.InstalledAppVersionHistoryId, 1)
		if err != nil {
			impl.logger.Errorw("error in performing deploy stage", "deployPayload", deployPayload, "err", err)
		}
	}

	// add required logging here
	var loggerFunc pubsub.LoggerFunc = func(msg model.PubSubMsg) (string, []interface{}) {
		deployPayload := &appStoreBean.DeployPayload{}
		err := json.Unmarshal([]byte(string(msg.Data)), &deployPayload)
		if err != nil {
			return "error while unmarshalling deployPayload json object", []interface{}{"error", err}
		}
		return "got message for deploy app-store apps in bulk", []interface{}{"installedAppVersionId", deployPayload.InstalledAppVersionId, "installedAppVersionHistoryId", deployPayload.InstalledAppVersionHistoryId}
	}

	err := impl.pubsubClient.Subscribe(pubsub.BULK_APPSTORE_DEPLOY_TOPIC, callback, loggerFunc)
	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}

func (impl *InstalledAppServiceImpl) DeployDefaultChartOnCluster(bean *cluster2.ClusterBean, userId int32) (bool, error) {
	// STEP 1 - create environment with name "devton"
	impl.logger.Infow("STEP 1", "create environment for cluster component", bean)
	envName := fmt.Sprintf("%d-%s", bean.Id, appStoreBean.DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT)
	env, err := impl.envService.FindOne(envName)
	if err != nil && err != pg.ErrNoRows {
		return false, err
	}
	if err == pg.ErrNoRows {
		env = &cluster2.EnvironmentBean{
			Environment: envName,
			ClusterId:   bean.Id,
			Namespace:   envName,
			Default:     false,
			Active:      true,
		}
		_, err := impl.envService.Create(env, userId)
		if err != nil {
			impl.logger.Errorw("DeployDefaultChartOnCluster, error in creating environment", "data", env, "err", err)
			return false, err
		}
	}

	// STEP 2 - create project with name "devtron"
	impl.logger.Info("STEP 2", "create project for cluster components")
	t, err := impl.teamRepository.FindByTeamName(appStoreBean.DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT)
	if err != nil && err != pg.ErrNoRows {
		return false, err
	}
	if err == pg.ErrNoRows {
		t := &repository4.Team{
			Name:     appStoreBean.DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT,
			Active:   true,
			AuditLog: sql.AuditLog{CreatedBy: userId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: userId},
		}
		err = impl.teamRepository.Save(t)
		if err != nil {
			impl.logger.Errorw("DeployDefaultChartOnCluster, error in creating team", "data", t, "err", err)
			return false, err
		}
	}

	// STEP 3- read the input data from env variables
	impl.logger.Info("STEP 3", "read the input data from env variables")
	charts := &appStoreBean.ChartComponents{}
	var chartComponents []*appStoreBean.ChartComponent
	if _, err := os.Stat(appStoreBean.CLUSTER_COMPONENT_DIR_PATH); os.IsNotExist(err) {
		impl.logger.Infow("default cluster component directory error", "cluster", bean.ClusterName, "err", err)
		return false, nil
	} else {
		fileInfo, err := ioutil.ReadDir(appStoreBean.CLUSTER_COMPONENT_DIR_PATH)
		if err != nil {
			impl.logger.Errorw("DeployDefaultChartOnCluster, err while reading directory", "err", err)
			return false, err
		}
		for _, file := range fileInfo {
			impl.logger.Infow("file", "name", file.Name())
			if strings.Contains(file.Name(), ".yaml") {
				content, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", appStoreBean.CLUSTER_COMPONENT_DIR_PATH, file.Name()))
				if err != nil {
					impl.logger.Errorw("DeployDefaultChartOnCluster, error on reading file", "err", err)
					return false, err
				}
				chartComponent := &appStoreBean.ChartComponent{
					Name:   strings.ReplaceAll(file.Name(), ".yaml", ""),
					Values: string(content),
				}
				chartComponents = append(chartComponents, chartComponent)
			}
		}

		if len(chartComponents) > 0 {
			charts.ChartComponent = chartComponents
			impl.logger.Info("STEP 4 - prepare a bulk request")
			// STEP 4 - prepare a bulk request (unique names need to apply for deploying chart)
			// STEP 4.1 - fetch chart for required name(actual chart name (app-store)) with default values
			// STEP 4.2 - update all the required charts, override values.yaml with env variables.
			chartGroupInstallRequest := &chartGroup.ChartGroupInstallRequest{}
			chartGroupInstallRequest.ProjectId = t.Id
			chartGroupInstallRequest.UserId = userId
			var chartGroupInstallChartRequests []*chartGroup.ChartGroupInstallChartRequest
			for _, item := range charts.ChartComponent {
				appStore, err := impl.appStoreApplicationVersionRepository.FindByAppStoreName(item.Name)
				if err != nil {
					impl.logger.Errorw("DeployDefaultChartOnCluster, error in getting app store", "data", t, "err", err)
					return false, err
				}
				chartGroupInstallChartRequest := &chartGroup.ChartGroupInstallChartRequest{
					AppName:                 fmt.Sprintf("%d-%d-%s", bean.Id, env.Id, item.Name),
					EnvironmentId:           env.Id,
					ValuesOverrideYaml:      item.Values,
					AppStoreVersion:         appStore.AppStoreApplicationVersionId,
					ReferenceValueId:        appStore.AppStoreApplicationVersionId,
					ReferenceValueKind:      appStoreBean.REFERENCE_TYPE_DEFAULT,
					DefaultClusterComponent: true,
				}
				chartGroupInstallChartRequests = append(chartGroupInstallChartRequests, chartGroupInstallChartRequest)
			}
			chartGroupInstallRequest.ChartGroupInstallChartRequest = chartGroupInstallChartRequests

			impl.logger.Info("STEP 5 - deploy bulk initiated")
			// STEP 5 - deploy
			_, err = impl.DeployDefaultComponent(chartGroupInstallRequest)
			if err != nil {
				impl.logger.Errorw("DeployDefaultChartOnCluster, error on bulk deploy", "err", err)
				return false, err
			}
		}
	}
	return true, nil
}

func (impl InstalledAppServiceImpl) DeployDefaultComponent(chartGroupInstallRequest *chartGroup.ChartGroupInstallRequest) (*chartGroup.ChartGroupInstallAppRes, error) {
	impl.logger.Debugw("bulk app install request", "req", chartGroupInstallRequest)
	//save in db
	// raise nats event

	var installAppVersionDTOList []*appStoreBean.InstallAppVersionDTO
	for _, installRequest := range chartGroupInstallRequest.ChartGroupInstallChartRequest {
		installAppVersionDTO, err := impl.requestBuilderForBulkDeployment(installRequest, chartGroupInstallRequest.ProjectId, chartGroupInstallRequest.UserId)
		if err != nil {
			impl.logger.Errorw("DeployBulk, error in request builder", "err", err)
			return nil, err
		}
		installAppVersionDTOList = append(installAppVersionDTOList, installAppVersionDTO)
	}
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return nil, err
	}
	var installAppVersions []*appStoreBean.InstallAppVersionDTO
	// Rollback tx on error.
	defer tx.Rollback()
	for _, installAppVersionDTO := range installAppVersionDTOList {
		installAppVersionDTO, err = impl.appStoreDeploymentService.AppStoreDeployOperationDB(installAppVersionDTO, tx, false)
		if err != nil {
			impl.logger.Errorw("DeployBulk, error while app store deploy db operation", "err", err)
			return nil, err
		}
		installAppVersions = append(installAppVersions, installAppVersionDTO)
	}
	if chartGroupInstallRequest.ChartGroupId > 0 {
		groupINstallationId, err := impl.getInstallationId(installAppVersions)
		if err != nil {
			return nil, err
		}
		for _, installAppVersionDTO := range installAppVersions {
			chartGroupEntry := impl.createChartGroupEntryObject(installAppVersionDTO, chartGroupInstallRequest.ChartGroupId, groupINstallationId)
			err := impl.chartGroupDeploymentRepository.Save(tx, chartGroupEntry)
			if err != nil {
				impl.logger.Errorw("DeployBulk, error in creating ChartGroupEntryObject", "err", err)
				return nil, err
			}
		}
	}
	//commit transaction
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("DeployBulk, error in tx commit", "err", err)
		return nil, err
	}
	//nats event

	for _, versions := range installAppVersions {
		_, err := impl.performDeployStage(versions.InstalledAppVersionId, versions.InstalledAppVersionHistoryId, chartGroupInstallRequest.UserId)
		if err != nil {
			impl.logger.Errorw("error in performing deploy stage", "deployPayload", versions, "err", err)
			_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(versions.InstalledAppVersionId, appStoreBean.QUE_ERROR)
			if err != nil {
				impl.logger.Errorw("error while bulk app-store deploy status update", "err", err)
			}
		}
	}

	return &chartGroup.ChartGroupInstallAppRes{}, nil
}

func (impl *InstalledAppServiceImpl) FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean2.AppDetailContainer, error) {
	installedAppVerison, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		impl.logger.Error(err)
		return bean2.AppDetailContainer{}, err
	}
	helmReleaseInstallStatus, status, err := impl.installedAppRepository.GetHelmReleaseStatusConfigByInstalledAppId(installedAppVerison.InstalledAppId)
	if err != nil {
		impl.logger.Errorw("error in getting helm release status from db", "err", err)
		return bean2.AppDetailContainer{}, err
	}
	var chartName string
	if installedAppVerison.AppStoreApplicationVersion.AppStore.ChartRepoId != 0 {
		chartName = installedAppVerison.AppStoreApplicationVersion.AppStore.ChartRepo.Name
	} else {
		chartName = installedAppVerison.AppStoreApplicationVersion.AppStore.DockerArtifactStore.Id
	}
	updateTime := installedAppVerison.InstalledApp.UpdatedOn
	timeStampTag := updateTime.Format(bean.LayoutDDMMYY_HHMM12hr)

	deploymentContainer := bean2.DeploymentDetailContainer{
		InstalledAppId:                installedAppVerison.InstalledApp.Id,
		AppId:                         installedAppVerison.InstalledApp.App.Id,
		AppStoreInstalledAppVersionId: installedAppVerison.Id,
		EnvironmentId:                 installedAppVerison.InstalledApp.EnvironmentId,
		AppName:                       installedAppVerison.InstalledApp.App.AppName,
		AppStoreChartName:             chartName,
		AppStoreChartId:               installedAppVerison.AppStoreApplicationVersion.AppStore.Id,
		AppStoreAppName:               installedAppVerison.AppStoreApplicationVersion.Name,
		AppStoreAppVersion:            installedAppVerison.AppStoreApplicationVersion.Version,
		EnvironmentName:               installedAppVerison.InstalledApp.Environment.Name,
		LastDeployedTime:              installedAppVerison.UpdatedOn.Format(bean.LayoutRFC3339),
		Namespace:                     installedAppVerison.InstalledApp.Environment.Namespace,
		Deprecated:                    installedAppVerison.AppStoreApplicationVersion.Deprecated,
		ClusterId:                     installedAppVerison.InstalledApp.Environment.ClusterId,
		DeploymentAppType:             installedAppVerison.InstalledApp.DeploymentAppType,
		DeploymentAppDeleteRequest:    installedAppVerison.InstalledApp.DeploymentAppDeleteRequest,
		IsVirtualEnvironment:          installedAppVerison.InstalledApp.Environment.IsVirtualEnvironment,
		HelmPackageName:               fmt.Sprintf("%s-%s-%s (GMT)", installedAppVerison.InstalledApp.App.AppName, installedAppVerison.InstalledApp.Environment.Name, timeStampTag),
		HelmReleaseInstallStatus:      helmReleaseInstallStatus,
		Status:                        status,
	}
	userInfo, err := impl.userService.GetByIdIncludeDeleted(installedAppVerison.AuditLog.UpdatedBy)
	if err != nil {
		impl.logger.Errorw("error fetching user info", "err", err)
		return bean2.AppDetailContainer{}, err
	}
	deploymentContainer.LastDeployedBy = userInfo.EmailId
	appDetail := bean2.AppDetailContainer{
		DeploymentDetailContainer: deploymentContainer,
	}
	return appDetail, nil
}

func (impl InstalledAppServiceImpl) GetInstalledAppVersionHistory(installedAppId int) (*appStoreBean.InstallAppVersionHistoryDto, error) {
	result := &appStoreBean.InstallAppVersionHistoryDto{}
	var history []*appStoreBean.IAVHistory
	//TODO - response setup

	installedAppVersions, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdMeta(installedAppId)
	if err != nil {
		impl.logger.Errorw("error while fetching installed version", "error", err)
		return result, err
	}
	for _, installedAppVersionModel := range installedAppVersions {
		versionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistoryByVersionId(installedAppVersionModel.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error while fetching installed version history", "error", err)
			return result, err
		}
		for _, updateHistory := range versionHistory {
			history = append(history, &appStoreBean.IAVHistory{
				ChartMetaData: appStoreBean.IAVHistoryChartMetaData{
					ChartName:    installedAppVersionModel.AppStoreApplicationVersion.AppStore.Name,
					ChartVersion: installedAppVersionModel.AppStoreApplicationVersion.Version,
					Description:  installedAppVersionModel.AppStoreApplicationVersion.Description,
					Home:         installedAppVersionModel.AppStoreApplicationVersion.Home,
					Sources:      []string{installedAppVersionModel.AppStoreApplicationVersion.Source},
				},
				DockerImages: []string{installedAppVersionModel.AppStoreApplicationVersion.AppVersion},
				DeployedAt: appStoreBean.IAVHistoryDeployedAt{
					Nanos:   updateHistory.CreatedOn.Nanosecond(),
					Seconds: updateHistory.CreatedOn.Unix(),
				},
				Version:               updateHistory.Id,
				InstalledAppVersionId: installedAppVersionModel.Id,
			})
		}
	}

	if len(history) == 0 {
		history = make([]*appStoreBean.IAVHistory, 0)
	}
	result.IAVHistory = history
	installedApp, err := impl.installedAppRepository.GetInstalledApp(installedAppId)
	if err != nil {
		impl.logger.Errorw("error while fetching installed version", "error", err)
		return result, err
	}
	result.InstalledAppInfo = &appStoreBean.InstalledAppDto{
		AppId:           installedApp.AppId,
		EnvironmentName: installedApp.Environment.Name,
		AppOfferingMode: installedApp.App.AppOfferingMode,
		InstalledAppId:  installedApp.Id,
		ClusterId:       installedApp.Environment.ClusterId,
		EnvironmentId:   installedApp.EnvironmentId,
	}
	return result, err
}

func (impl InstalledAppServiceImpl) UpdateInstalledAppVersionStatus(application *v1alpha1.Application) (bool, error) {
	isHealthy := false
	dbConnection := impl.installedAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return isHealthy, err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	gitHash := ""
	if application.Operation != nil && application.Operation.Sync != nil {
		gitHash = application.Operation.Sync.Revision
	} else if application.Status.OperationState != nil && application.Status.OperationState.Operation.Sync != nil {
		gitHash = application.Status.OperationState.Operation.Sync.Revision
	}
	versionHistory, err := impl.installedAppRepositoryHistory.GetLatestInstalledAppVersionHistoryByGitHash(gitHash)
	if err != nil {
		impl.logger.Errorw("error while fetching installed version history", "error", err)
		return isHealthy, err
	}
	if versionHistory.Status != (application2.Healthy) {
		versionHistory.Status = string(application.Status.Health.Status)
		versionHistory.UpdatedOn = time.Now()
		versionHistory.UpdatedBy = 1
		impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(versionHistory, tx)
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while committing transaction to db", "error", err)
		return isHealthy, err
	}

	appId, envId, err := impl.installedAppRepositoryHistory.GetAppIdAndEnvIdWithInstalledAppVersionId(versionHistory.InstalledAppVersionId)
	if err == nil {
		err = impl.appStatusService.UpdateStatusWithAppIdEnvId(appId, envId, string(application.Status.Health.Status))
		if err != nil {
			impl.logger.Errorw("error while updating app status in app_status table", "error", err, "appId", appId, "envId", envId)
		}
	}
	return true, nil
}

func (impl InstalledAppServiceImpl) GetInstalledAppVersionHistoryValues(installedAppVersionHistoryId int) (*appStoreBean.IAVHistoryValues, error) {
	values := &appStoreBean.IAVHistoryValues{}
	versionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(installedAppVersionHistoryId)
	if err != nil {
		impl.logger.Errorw("error while fetching installed version history", "error", err)
		return nil, err
	}
	values.ValuesYaml = versionHistory.ValuesYamlRaw
	return values, err
}

func (impl InstalledAppServiceImpl) getReleaseStatusFromHelmReleaseInstallStatus(helmReleaseInstallStatus string, status string) *client.ReleaseStatus {
	//release status is sent in resource tree call and is shown on UI as helm config apply status
	releaseStatus := &client.ReleaseStatus{}
	if len(helmReleaseInstallStatus) > 0 {
		helmInstallStatus := &appStoreBean.HelmReleaseStatusConfig{}
		err := json.Unmarshal([]byte(helmReleaseInstallStatus), helmInstallStatus)
		if err != nil {
			impl.logger.Errorw("error in unmarshalling helm release install status")
			return releaseStatus
		}
		if status == appStoreBean.HELM_RELEASE_STATUS_FAILED {
			releaseStatus.Status = status
			releaseStatus.Description = helmInstallStatus.Message
			releaseStatus.Message = "Release install/upgrade failed"
		} else if status == appStoreBean.HELM_RELEASE_STATUS_PROGRESSING {
			releaseStatus.Status = status
			releaseStatus.Description = helmInstallStatus.Message
			releaseStatus.Message = helmInstallStatus.Message
		} else {
			// there can be a case when helm release is created but we are not able to fetch it
			releaseStatus.Status = appStoreBean.HELM_RELEASE_STATUS_UNKNOWN
			releaseStatus.Description = "Unable to fetch release for app"
			releaseStatus.Message = "Unable to fetch release for app"
		}
	} else {
		releaseStatus.Status = appStoreBean.HELM_RELEASE_STATUS_UNKNOWN
		releaseStatus.Description = "Release not found"
		releaseStatus.Message = "Release not found "
	}
	return releaseStatus
}

func (impl InstalledAppServiceImpl) MarkGitOpsInstalledAppsDeletedIfArgoAppIsDeleted(installedAppId int, envId int) error {
	apiError := &util.ApiError{}
	installedApp, err := impl.installedAppRepository.GetGitOpsInstalledAppsWhereArgoAppDeletedIsTrue(installedAppId, envId)
	if err != nil {
		impl.logger.Errorw("error in fetching partially deleted argoCd apps from installed app repo", "err", err)
		apiError.HttpStatusCode = http.StatusInternalServerError
		apiError.InternalMessage = "error in fetching partially deleted argoCd apps from installed app repo"
		return apiError
	}
	acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
	if err != nil {
		impl.logger.Errorw("error in getting acd token", "err", err)
		apiError.HttpStatusCode = http.StatusInternalServerError
		apiError.InternalMessage = "error in getting acd token"
		return apiError
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "token", acdToken)

	acdAppName := fmt.Sprintf("%s-%s", installedApp.App.AppName, installedApp.Environment.Name)
	_, err = impl.acdClient.Get(ctx, &application.ApplicationQuery{Name: &acdAppName})

	if err == nil {
		apiError.HttpStatusCode = http.StatusInternalServerError
		apiError.InternalMessage = "App Exist in argo, error in fetching resource tree"
		return apiError
	}

	impl.logger.Warnw("app not found in argo, deleting from db ", "err", err)
	//make call to delete it from pipeline DB
	deleteRequest := &appStoreBean.InstallAppVersionDTO{}
	deleteRequest.ForceDelete = false
	deleteRequest.NonCascadeDelete = false
	deleteRequest.AcdPartialDelete = false
	deleteRequest.InstalledAppId = installedApp.Id
	deleteRequest.AppId = installedApp.AppId
	deleteRequest.AppName = installedApp.App.AppName
	deleteRequest.Namespace = installedApp.Environment.Namespace
	deleteRequest.ClusterId = installedApp.Environment.ClusterId
	deleteRequest.EnvironmentId = installedApp.EnvironmentId
	deleteRequest.AppOfferingMode = installedApp.App.AppOfferingMode
	deleteRequest.UserId = 1
	_, err = impl.appStoreDeploymentService.DeleteInstalledApp(context.Background(), deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting installed app", "err", err)
		apiError.HttpStatusCode = http.StatusNotFound
		apiError.InternalMessage = "error in deleting installed app"
		return apiError
	}
	apiError.HttpStatusCode = http.StatusNotFound
	return apiError
}

func (impl InstalledAppServiceImpl) CheckAppExistsByInstalledAppId(installedAppId int) (*repository2.InstalledApps, error) {
	installedApp, err := impl.installedAppRepository.GetInstalledApp(installedAppId)
	if err != nil {
		return nil, err
	}
	return installedApp, err
}

func (impl InstalledAppServiceImpl) GetChartBytesForLatestDeployment(installedAppId int, installedAppVersionId int) ([]byte, error) {

	chartBytes := make([]byte, 0)

	installedApp, err := impl.installedAppRepository.GetInstalledApp(installedAppId)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app", "err", err, "installed_app_id", installedAppId)
		return chartBytes, err
	}
	installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersion(installedAppVersionId)
	if err != nil {
		impl.logger.Errorw("Service err, BuildChartWithValuesAndRequirementsConfig", err, "installed_app_version_id", installedAppVersionId)
		return chartBytes, err
	}

	valuesString, err := impl.appStoreDeploymentCommonService.GetValuesString(installedAppVersion.AppStoreApplicationVersion.AppStore.Name, installedAppVersion.ValuesYaml)
	if err != nil {
		return chartBytes, err
	}
	requirementsString, err := impl.appStoreDeploymentCommonService.GetRequirementsString(installedAppVersion.AppStoreApplicationVersionId)
	if err != nil {
		return chartBytes, err
	}

	updateTime := installedApp.UpdatedOn
	timeStampTag := updateTime.Format(bean.LayoutDDMMYY_HHMM12hr)
	chartName := fmt.Sprintf("%s-%s-%s (GMT)", installedApp.App.AppName, installedApp.Environment.Name, timeStampTag)
	chartBytes, err = impl.appStoreDeploymentCommonService.BuildChartWithValuesAndRequirementsConfig(installedApp.App.AppName, valuesString, requirementsString, chartName, fmt.Sprint(installedApp.Id))

	if err != nil {
		return chartBytes, err
	}
	return chartBytes, nil
}

func (impl InstalledAppServiceImpl) GetChartBytesForParticularDeployment(installedAppId int, installedAppVersionId int, installedAppVersionHistoryId int) ([]byte, error) {

	chartBytes := make([]byte, 0)

	installedApp, err := impl.installedAppRepository.GetInstalledApp(installedAppId)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app", "err", err, "installed_app_id", installedAppId)
		return chartBytes, err
	}
	installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersionAny(installedAppVersionId)
	if err != nil {
		impl.logger.Errorw("Service err, BuildChartWithValuesAndRequirementsConfig", err, "installed_app_version_id", installedAppVersionId)
		return chartBytes, err
	}
	installedAppVersionHistory, err := impl.installedAppRepositoryHistory.GetInstalledAppVersionHistory(installedAppVersionHistoryId)

	valuesString, err := impl.appStoreDeploymentCommonService.GetValuesString(installedAppVersion.AppStoreApplicationVersion.AppStore.Name, installedAppVersionHistory.ValuesYamlRaw)
	if err != nil {
		return chartBytes, err
	}
	requirementsString, err := impl.appStoreDeploymentCommonService.GetRequirementsString(installedAppVersion.AppStoreApplicationVersionId)
	if err != nil {
		return chartBytes, err
	}

	updateTime := installedApp.UpdatedOn
	timeStampTag := updateTime.Format(bean.LayoutDDMMYY_HHMM12hr)
	chartName := fmt.Sprintf("%s-%s-%s (GMT)", installedApp.App.AppName, installedApp.Environment.Name, timeStampTag)

	chartBytes, err = impl.appStoreDeploymentCommonService.BuildChartWithValuesAndRequirementsConfig(installedApp.App.AppName, valuesString, requirementsString, chartName, fmt.Sprint(installedApp.Id))
	if err != nil {
		return chartBytes, err
	}
	return chartBytes, nil
}
