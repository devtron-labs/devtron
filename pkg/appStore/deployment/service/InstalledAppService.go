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
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	client "github.com/devtron-labs/devtron/api/helm-app"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	bean3 "github.com/devtron-labs/devtron/api/restHandler/bean"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app/status"
	"github.com/devtron-labs/devtron/pkg/appStatus"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentCommon "github.com/devtron-labs/devtron/pkg/appStore/deployment/common"
	appStoreDeploymentFullMode "github.com/devtron-labs/devtron/pkg/appStore/deployment/fullMode"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDeploymentGitopsTool "github.com/devtron-labs/devtron/pkg/appStore/deployment/tool/gitops"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/values/service"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/k8s"
	application3 "github.com/devtron-labs/devtron/pkg/k8s/application"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository4 "github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	util4 "github.com/devtron-labs/devtron/util/k8s"
	"github.com/tidwall/gjson"
	"net/http"
	"regexp"

	/* #nosec */
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

const (
	DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT = "devtron"
	CLUSTER_COMPONENT_DIR_PATH                  = "/cluster/component"
	DEFAULT_CLUSTER_ID                          = 1
	DEFAULT_CLUSTER_NAMESPACE                   = "default"
)

type InstalledAppService interface {
	GetAll(filter *appStoreBean.AppStoreFilter) (AppListDetail, error)
	DeployBulk(chartGroupInstallRequest *appStoreBean.ChartGroupInstallRequest) (*appStoreBean.ChartGroupInstallAppRes, error)
	performDeployStage(appId int, installedAppVersionHistoryId int, userId int32) (*appStoreBean.InstallAppVersionDTO, error)
	CheckAppExists(appNames []*appStoreBean.AppNames) ([]*appStoreBean.AppNames, error)
	DeployDefaultChartOnCluster(bean *cluster2.ClusterBean, userId int32) (bool, error)
	FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean2.AppDetailContainer, error)
	UpdateInstalledAppVersionStatus(application *v1alpha1.Application) (bool, error)
	FetchResourceTree(rctx context.Context, cn http.CloseNotifier, resourceTreeAndNotesContainer *bean2.ResourceTreeAndNotesContainer, installedApp repository2.InstalledApps) error
	MarkGitOpsInstalledAppsDeletedIfArgoAppIsDeleted(installedAppId int, envId int) error
	CheckAppExistsByInstalledAppId(installedAppId int) (*repository2.InstalledApps, error)
	FindNotesForNonHelmApplication(installedAppId, envId int) (string, string, error)
	FetchChartNotes(installedAppId int, envId int, token string, checkNotesAuth func(token string, appName string, envId int) bool) (string, error)
	FetchResourceTreeWithHibernateForACD(rctx context.Context, cn http.CloseNotifier, appDetail *bean2.AppDetailContainer) bean2.AppDetailContainer
	fetchResourceTreeForACD(rctx context.Context, cn http.CloseNotifier, appId int, envId, clusterId int, deploymentAppName, namespace string) (map[string]interface{}, error)
	GetChartBytesForLatestDeployment(installedAppId int, installedAppVersionId int) ([]byte, error)
	GetChartBytesForParticularDeployment(installedAppId int, installedAppVersionId int, installedAppVersionHistoryId int) ([]byte, error)
}

type InstalledAppServiceImpl struct {
	logger                               *zap.SugaredLogger
	installedAppRepository               repository2.InstalledAppRepository
	chartTemplateService                 util.ChartTemplateService
	refChartDir                          appStoreBean.RefChartProxyDir
	repositoryService                    repository.ServiceClient
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	environmentRepository                repository5.EnvironmentRepository
	teamRepository                       repository4.TeamRepository
	appRepository                        app.AppRepository
	acdClient                            application2.ServiceClient
	appStoreValuesService                service.AppStoreValuesService
	pubsubClient                         *pubsub.PubSubClientServiceImpl
	tokenCache                           *util2.TokenCache
	chartGroupDeploymentRepository       repository2.ChartGroupDeploymentRepository
	envService                           cluster2.EnvironmentService
	ArgoK8sClient                        argocdServer.ArgoK8sClient
	gitFactory                           *util.GitFactory
	aCDAuthConfig                        *util2.ACDAuthConfig
	gitOpsRepository                     repository3.GitOpsConfigRepository
	userService                          user.UserService
	appStoreDeploymentService            AppStoreDeploymentService
	appStoreDeploymentFullModeService    appStoreDeploymentFullMode.AppStoreDeploymentFullModeService
	installedAppRepositoryHistory        repository2.InstalledAppVersionHistoryRepository
	argoUserService                      argo.ArgoUserService
	helmAppClient                        client.HelmAppClient
	helmAppService                       client.HelmAppService
	attributesRepository                 repository3.AttributesRepository
	appStatusService                     appStatus.AppStatusService
	K8sUtil                              *util4.K8sUtil
	pipelineStatusTimelineService        status.PipelineStatusTimelineService
	appStoreDeploymentCommonService      appStoreDeploymentCommon.AppStoreDeploymentCommonService
	k8sCommonService                     k8s.K8sCommonService
	k8sApplicationService                application3.K8sApplicationService
}

func NewInstalledAppServiceImpl(logger *zap.SugaredLogger,
	installedAppRepository repository2.InstalledAppRepository,
	chartTemplateService util.ChartTemplateService, refChartDir appStoreBean.RefChartProxyDir,
	repositoryService repository.ServiceClient,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	environmentRepository repository5.EnvironmentRepository, teamRepository repository4.TeamRepository,
	appRepository app.AppRepository,
	acdClient application2.ServiceClient,
	appStoreValuesService service.AppStoreValuesService,
	pubsubClient *pubsub.PubSubClientServiceImpl,
	tokenCache *util2.TokenCache,
	chartGroupDeploymentRepository repository2.ChartGroupDeploymentRepository,
	envService cluster2.EnvironmentService, argoK8sClient argocdServer.ArgoK8sClient,
	gitFactory *util.GitFactory, aCDAuthConfig *util2.ACDAuthConfig, gitOpsRepository repository3.GitOpsConfigRepository, userService user.UserService,
	appStoreDeploymentFullModeService appStoreDeploymentFullMode.AppStoreDeploymentFullModeService,
	appStoreDeploymentService AppStoreDeploymentService,
	installedAppRepositoryHistory repository2.InstalledAppVersionHistoryRepository,
	argoUserService argo.ArgoUserService, helmAppClient client.HelmAppClient, helmAppService client.HelmAppService,
	attributesRepository repository3.AttributesRepository,
	appStatusService appStatus.AppStatusService, K8sUtil *util4.K8sUtil,
	pipelineStatusTimelineService status.PipelineStatusTimelineService,
	appStoreDeploymentCommonService appStoreDeploymentCommon.AppStoreDeploymentCommonService,
	appStoreDeploymentArgoCdService appStoreDeploymentGitopsTool.AppStoreDeploymentArgoCdService, k8sCommonService k8s.K8sCommonService, k8sApplicationService application3.K8sApplicationService) (*InstalledAppServiceImpl, error) {
	impl := &InstalledAppServiceImpl{
		logger:                               logger,
		installedAppRepository:               installedAppRepository,
		chartTemplateService:                 chartTemplateService,
		refChartDir:                          refChartDir,
		repositoryService:                    repositoryService,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		environmentRepository:                environmentRepository,
		teamRepository:                       teamRepository,
		appRepository:                        appRepository,
		acdClient:                            acdClient,
		appStoreValuesService:                appStoreValuesService,
		pubsubClient:                         pubsubClient,
		tokenCache:                           tokenCache,
		chartGroupDeploymentRepository:       chartGroupDeploymentRepository,
		envService:                           envService,
		ArgoK8sClient:                        argoK8sClient,
		gitFactory:                           gitFactory,
		aCDAuthConfig:                        aCDAuthConfig,
		gitOpsRepository:                     gitOpsRepository,
		userService:                          userService,
		appStoreDeploymentService:            appStoreDeploymentService,
		appStoreDeploymentFullModeService:    appStoreDeploymentFullModeService,
		installedAppRepositoryHistory:        installedAppRepositoryHistory,
		argoUserService:                      argoUserService,
		helmAppClient:                        helmAppClient,
		helmAppService:                       helmAppService,
		attributesRepository:                 attributesRepository,
		appStatusService:                     appStatusService,
		K8sUtil:                              K8sUtil,
		pipelineStatusTimelineService:        pipelineStatusTimelineService,
		appStoreDeploymentCommonService:      appStoreDeploymentCommonService,
		k8sCommonService:                     k8sCommonService,
		k8sApplicationService:                k8sApplicationService,
	}
	err := impl.Subscribe()
	if err != nil {
		return nil, err
	}
	return impl, nil
}

type EnvironmentDetails struct {
	EnvironmentName *string `json:"environmentName,omitempty"`
	// id in which app is deployed
	EnvironmentId *int32 `json:"environmentId,omitempty"`
	// namespace corresponding to the environemnt
	Namespace *string `json:"namespace,omitempty"`
	// if given environemnt is marked as production or not, nullable
	IsPrduction *bool `json:"isPrduction,omitempty"`
	// cluster corresponding to the environemt where application is deployed
	ClusterName *string `json:"clusterName,omitempty"`
	// clusterId corresponding to the environemt where application is deployed
	ClusterId *int32 `json:"clusterId,omitempty"`

	IsVirtualEnvironment *bool `json:"isVirtualEnvironment"`
}

type HelmAppDetails struct {
	// time when this application was last deployed/updated
	LastDeployedAt *time.Time `json:"lastDeployedAt,omitempty"`
	// name of the helm application/helm release name
	AppName *string `json:"appName,omitempty"`
	// unique identifier for app
	AppId *string `json:"appId,omitempty"`
	// name of the chart
	ChartName *string `json:"chartName,omitempty"`
	// url/location of the chart icon
	ChartAvatar *string `json:"chartAvatar,omitempty"`
	// unique identifier for the project, APP with no project will have id `0`
	ProjectId *int32 `json:"projectId,omitempty"`
	// chart version
	ChartVersion      *string             `json:"chartVersion,omitempty"`
	EnvironmentDetail *EnvironmentDetails `json:"environmentDetail,omitempty"`
	AppStatus         *string             `json:"appStatus,omitempty"`
}

type AppListDetail struct {
	// clusters to which result corresponds
	ClusterIds *[]int32 `json:"clusterIds,omitempty"`
	// application type inside the array
	ApplicationType *string `json:"applicationType,omitempty"`
	// if data fetch for that cluster produced error
	Errored *bool `json:"errored,omitempty"`
	// error msg if client failed to fetch
	ErrorMsg *string `json:"errorMsg,omitempty"`
	// all helm app list, EA+ devtronapp
	HelmApps *[]HelmAppDetails `json:"helmApps,omitempty"`
	// all helm app list, EA+ devtronapp
	DevtronApps *[]openapi.DevtronApp `json:"devtronApps,omitempty"`
}

func (impl InstalledAppServiceImpl) GetAll(filter *appStoreBean.AppStoreFilter) (AppListDetail, error) {
	applicationType := "DEVTRON-CHART-STORE"
	var clusterIdsConverted []int32
	for _, clusterId := range filter.ClusterIds {
		clusterIdsConverted = append(clusterIdsConverted, int32(clusterId))
	}
	installedAppsResponse := AppListDetail{
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
	var helmAppsResponse []HelmAppDetails
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
		environmentDetails := EnvironmentDetails{
			EnvironmentName:      &appLocal.EnvironmentName,
			EnvironmentId:        &envId,
			Namespace:            &appLocal.Namespace,
			ClusterName:          &appLocal.ClusterName,
			ClusterId:            &clusterId,
			IsVirtualEnvironment: &appLocal.IsVirtualEnvironment,
		}
		helmAppResp := HelmAppDetails{
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

// converts db object to bean
func (impl InstalledAppServiceImpl) chartAdaptor(chart *repository2.InstalledAppVersions) (*appStoreBean.InstallAppVersionDTO, error) {

	return &appStoreBean.InstallAppVersionDTO{
		InstalledAppId:     chart.InstalledAppId,
		Id:                 chart.Id,
		AppStoreVersion:    chart.AppStoreApplicationVersionId,
		ValuesOverrideYaml: chart.ValuesYaml,
	}, nil
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

func (impl InstalledAppServiceImpl) DeployBulk(chartGroupInstallRequest *appStoreBean.ChartGroupInstallRequest) (*appStoreBean.ChartGroupInstallAppRes, error) {
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
	return &appStoreBean.ChartGroupInstallAppRes{}, nil
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

func (impl InstalledAppServiceImpl) createChartGroupEntryObject(installAppVersionDTO *appStoreBean.InstallAppVersionDTO, chartGroupId int, groupINstallationId string) *repository2.ChartGroupDeployment {
	return &repository2.ChartGroupDeployment{
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
	chartGitAttr := &util.ChartGitAttribute{}
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
		gitOpsConfigBitbucket, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(util.BITBUCKET_PROVIDER)
		if err != nil {
			if err == pg.ErrNoRows {
				gitOpsConfigBitbucket.BitBucketWorkspaceId = ""
				gitOpsConfigBitbucket.BitBucketProjectKey = ""
			} else {
				return nil, err
			}
		}
		config := &bean2.GitOpsConfigDto{
			GitRepoName:          installedAppVersion.GitOpsRepoName,
			BitBucketWorkspaceId: gitOpsConfigBitbucket.BitBucketProjectKey,
			BitBucketProjectKey:  gitOpsConfigBitbucket.BitBucketProjectKey,
		}
		repoUrl, err := impl.gitFactory.Client.GetRepoUrl(config)
		if err != nil {
			//will allow to continue to persist status on next operation
			impl.logger.Errorw("fetching error", "err", err)
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
		_, err := impl.appStoreDeploymentFullModeService.AppStoreDeployOperationACD(installedAppVersion, chartGitAttr, ctx)
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

	if util.IsAcdApp(installedAppVersion.DeploymentAppType) {
		// update build history for chart for argo_cd apps
		err = impl.appStoreDeploymentService.UpdateInstalledAppVersionHistoryWithGitHash(installedAppVersion)
		if err != nil {
			impl.logger.Errorw("error on updating history for chart deployment", "error", err, "installedAppVersion", installedAppVersion)
			return nil, err
		}
	} else {
		// create build history for chart on default component deployed via helm
		err = impl.appStoreDeploymentService.UpdateInstallAppVersionHistory(installedAppVersion)
		if err != nil {
			impl.logger.Errorw("error on creating history for chart deployment", "error", err, "installedAppVersion", installedAppVersion)
			return nil, err
		}
	}

	return installedAppVersion, nil
}

func (impl InstalledAppServiceImpl) requestBuilderForBulkDeployment(installRequest *appStoreBean.ChartGroupInstallChartRequest, projectId int, userId int32) (*appStoreBean.InstallAppVersionDTO, error) {
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

func (impl *InstalledAppServiceImpl) Subscribe() error {
	callback := func(msg *pubsub.PubSubMsg) {
		impl.logger.Debug("cd stage event received")
		//defer msg.Ack()
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
	err := impl.pubsubClient.Subscribe(pubsub.BULK_APPSTORE_DEPLOY_TOPIC, callback)
	if err != nil {
		impl.logger.Error("err", err)
		return err
	}
	return nil
}

func (impl *InstalledAppServiceImpl) DeployDefaultChartOnCluster(bean *cluster2.ClusterBean, userId int32) (bool, error) {
	// STEP 1 - create environment with name "devton"
	impl.logger.Infow("STEP 1", "create environment for cluster component", bean)
	envName := fmt.Sprintf("%d-%s", bean.Id, DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT)
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
	t, err := impl.teamRepository.FindByTeamName(DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT)
	if err != nil && err != pg.ErrNoRows {
		return false, err
	}
	if err == pg.ErrNoRows {
		t := &repository4.Team{
			Name:     DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT,
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
	charts := &ChartComponents{}
	var chartComponents []*ChartComponent
	if _, err := os.Stat(CLUSTER_COMPONENT_DIR_PATH); os.IsNotExist(err) {
		impl.logger.Infow("default cluster component directory error", "cluster", bean.ClusterName, "err", err)
		return false, nil
	} else {
		fileInfo, err := ioutil.ReadDir(CLUSTER_COMPONENT_DIR_PATH)
		if err != nil {
			impl.logger.Errorw("DeployDefaultChartOnCluster, err while reading directory", "err", err)
			return false, err
		}
		for _, file := range fileInfo {
			impl.logger.Infow("file", "name", file.Name())
			if strings.Contains(file.Name(), ".yaml") {
				content, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", CLUSTER_COMPONENT_DIR_PATH, file.Name()))
				if err != nil {
					impl.logger.Errorw("DeployDefaultChartOnCluster, error on reading file", "err", err)
					return false, err
				}
				chartComponent := &ChartComponent{
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
			chartGroupInstallRequest := &appStoreBean.ChartGroupInstallRequest{}
			chartGroupInstallRequest.ProjectId = t.Id
			chartGroupInstallRequest.UserId = userId
			var chartGroupInstallChartRequests []*appStoreBean.ChartGroupInstallChartRequest
			for _, item := range charts.ChartComponent {
				appStore, err := impl.appStoreApplicationVersionRepository.FindByAppStoreName(item.Name)
				if err != nil {
					impl.logger.Errorw("DeployDefaultChartOnCluster, error in getting app store", "data", t, "err", err)
					return false, err
				}
				chartGroupInstallChartRequest := &appStoreBean.ChartGroupInstallChartRequest{
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

type ChartComponents struct {
	ChartComponent []*ChartComponent `json:"charts"`
}
type ChartComponent struct {
	Name   string `json:"name"`
	Values string `json:"values"`
}

func (impl InstalledAppServiceImpl) DeployDefaultComponent(chartGroupInstallRequest *appStoreBean.ChartGroupInstallRequest) (*appStoreBean.ChartGroupInstallAppRes, error) {
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

	return &appStoreBean.ChartGroupInstallAppRes{}, nil
}

func (impl *InstalledAppServiceImpl) FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean2.AppDetailContainer, error) {
	installedAppVerison, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		impl.logger.Error(err)
		return bean2.AppDetailContainer{}, err
	}
	updateTime := installedAppVerison.InstalledApp.UpdatedOn
	timeStampTag := updateTime.Format(bean.LayoutDDMMYY_HHMM12hr)

	deploymentContainer := bean2.DeploymentDetailContainer{
		InstalledAppId:                installedAppVerison.InstalledApp.Id,
		AppId:                         installedAppVerison.InstalledApp.App.Id,
		AppStoreInstalledAppVersionId: installedAppVerison.Id,
		EnvironmentId:                 installedAppVerison.InstalledApp.EnvironmentId,
		AppName:                       installedAppVerison.InstalledApp.App.AppName,
		AppStoreChartName:             installedAppVerison.AppStoreApplicationVersion.AppStore.ChartRepo.Name,
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
func (impl *InstalledAppServiceImpl) FetchChartNotes(installedAppId int, envId int, token string, checkNotesAuth func(token string, appName string, envId int) bool) (string, error) {
	//check notes.txt in db
	installedApp, err := impl.installedAppRepository.FetchNotes(installedAppId)
	installedAppVerison, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		impl.logger.Errorw("error fetching installed  app version in installed app service", "err", err)
		return "", err
	}
	chartVersion := installedAppVerison.AppStoreApplicationVersion.Version
	if err != nil {
		impl.logger.Errorw("error fetching chart  version in installed app service", "err", err)
		return "", err
	}
	re := regexp.MustCompile(`CHART VERSION: ([0-9]+\.[0-9]+\.[0-9]+)`)
	newStr := re.ReplaceAllString(installedApp.Notes, "CHART VERSION: "+chartVersion)
	installedApp.Notes = newStr
	appName := installedApp.App.AppName
	if err != nil {
		impl.logger.Errorw("error fetching notes from db", "err", err)
		return "", err
	}
	isValidAuth := checkNotesAuth(token, appName, envId)
	if !isValidAuth {
		impl.logger.Errorw("unauthorized user", "isValidAuth", isValidAuth)
		return "", fmt.Errorf("unauthorized user")
	}
	//if notes is not present in db then below call will happen
	if installedApp.Notes == "" {
		notes, _, err := impl.FindNotesForNonHelmApplication(installedAppId, envId)
		if err != nil {
			impl.logger.Errorw("error fetching notes", "err", err)
			return "", err
		}
		if notes == "" {
			impl.logger.Errorw("error fetching notes", "err", err)
		}
		return notes, err
	}

	return installedApp.Notes, nil
}
func (impl *InstalledAppServiceImpl) FindNotesForNonHelmApplication(installedAppId, envId int) (string, string, error) {
	installedAppVerison, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		impl.logger.Errorw("error fetching installed  app version in installed app service", "err", err)
		return "", "", err
	}
	var notes string
	appName := installedAppVerison.InstalledApp.App.AppName

	if util.IsAcdApp(installedAppVerison.InstalledApp.DeploymentAppType) || util.IsManifestDownload(installedAppVerison.InstalledApp.DeploymentAppType) {
		appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installedAppVerison.AppStoreApplicationVersion.Id)
		if err != nil {
			impl.logger.Errorw("error fetching app store app version in installed app service", "err", err)
			return notes, appName, err
		}
		k8sServerVersion, err := impl.K8sUtil.GetKubeVersion()
		if err != nil {
			impl.logger.Errorw("exception caught in getting k8sServerVersion", "err", err)
			return notes, appName, err
		}
		clusterId := int32(installedAppVerison.InstalledApp.Environment.ClusterId)
		namespace := installedAppVerison.InstalledApp.Environment.Namespace
		if installedAppVerison.InstalledApp.Environment.IsVirtualEnvironment {
			clusterId = int32(DEFAULT_CLUSTER_ID)
			namespace = DEFAULT_CLUSTER_NAMESPACE
		}
		installReleaseRequest := &client.InstallReleaseRequest{
			ChartName:    appStoreAppVersion.Name,
			ChartVersion: appStoreAppVersion.Version,
			ValuesYaml:   installedAppVerison.ValuesYaml,
			K8SVersion:   k8sServerVersion.String(),
			ChartRepository: &client.ChartRepository{
				Name:     appStoreAppVersion.AppStore.ChartRepo.Name,
				Url:      appStoreAppVersion.AppStore.ChartRepo.Url,
				Username: appStoreAppVersion.AppStore.ChartRepo.UserName,
				Password: appStoreAppVersion.AppStore.ChartRepo.Password,
			},
			ReleaseIdentifier: &client.ReleaseIdentifier{
				ReleaseNamespace: namespace,
				ReleaseName:      installedAppVerison.InstalledApp.App.AppName,
				ClusterConfig: &client.ClusterConfig{
					ClusterId: clusterId,
				},
			},
		}

		notes, err = impl.helmAppService.GetNotes(context.Background(), installReleaseRequest)
		if err != nil {
			impl.logger.Errorw("error in fetching notes", "err", err)
			return notes, appName, err
		}
		_, err = impl.appStoreDeploymentService.UpdateNotesForInstalledApp(installedAppId, notes)
		if err != nil {
			impl.logger.Errorw("error in updating notes in db ", "err", err)
			return notes, appName, err
		}
	}

	return notes, appName, nil
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
func (impl InstalledAppServiceImpl) FetchResourceTree(rctx context.Context, cn http.CloseNotifier, resourceTreeAndNotesContainer *bean2.ResourceTreeAndNotesContainer, installedApp repository2.InstalledApps) error {
	var err error
	var resourceTree map[string]interface{}
	deploymentAppName := fmt.Sprintf("%s-%s", installedApp.App.AppName, installedApp.Environment.Name)
	if util.IsAcdApp(installedApp.DeploymentAppType) {
		resourceTree, err = impl.fetchResourceTreeForACD(rctx, cn, installedApp.App.Id, installedApp.EnvironmentId, installedApp.Environment.ClusterId, deploymentAppName, installedApp.Environment.Namespace)
	} else if util.IsHelmApp(installedApp.DeploymentAppType) {
		config, err := impl.helmAppService.GetClusterConf(installedApp.Environment.ClusterId)
		if err != nil {
			impl.logger.Errorw("error in fetching cluster detail", "err", err)
		}
		req := &client.AppDetailRequest{
			ClusterConfig: config,
			Namespace:     installedApp.Environment.Namespace,
			ReleaseName:   installedApp.App.AppName,
		}
		detail, err := impl.helmAppClient.GetAppDetail(rctx, req)
		if err != nil {
			impl.logger.Errorw("error in fetching app detail", "err", err)
		}
		if detail != nil {
			resourceTree = util3.InterfaceToMapAdapter(detail.ResourceTreeResponse)
			resourceTree["status"] = detail.ApplicationStatus
			resourceTreeAndNotesContainer.Notes = detail.ChartMetadata.Notes
			impl.logger.Warnw("appName and envName not found - avoiding resource tree call", "app", installedApp.App.AppName, "env", installedApp.Environment.Name)
		}
	}
	version, err := impl.k8sCommonService.GetK8sServerVersion(installedApp.Environment.ClusterId)
	if err != nil {
		impl.logger.Errorw("error in fetching k8s version in resource tree call fetching", "clusterId", installedApp.Environment.ClusterId, "err", err)
	} else {
		resourceTree["serverVersion"] = version.String()
	}
	resourceTreeAndNotesContainer.ResourceTree = resourceTree
	return err
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
	return installedApp, err
}

func (impl InstalledAppServiceImpl) FetchResourceTreeWithHibernateForACD(rctx context.Context, cn http.CloseNotifier, appDetail *bean2.AppDetailContainer) bean2.AppDetailContainer {
	ctx, cancel := context.WithCancel(rctx)
	if cn != nil {
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
		impl.logger.Errorw("error in getting acd token", "err", err)
		return *appDetail
	}
	ctx = context.WithValue(ctx, "token", acdToken)
	defer cancel()
	deploymentAppName := fmt.Sprintf("%s-%s", appDetail.AppName, appDetail.EnvironmentName)
	resourceTree, err := impl.fetchResourceTreeForACD(rctx, cn, appDetail.InstalledAppId, appDetail.EnvironmentId, appDetail.ClusterId, deploymentAppName, appDetail.Namespace)
	appDetail.ResourceTree = resourceTree
	if err != nil {
		return *appDetail
	}
	if appDetail.ResourceTree["nodes"] == nil {
		return *appDetail
	}
	appDetail.ResourceTree = checkHibernate(impl, appDetail, ctx)
	return *appDetail
}
func checkHibernate(impl InstalledAppServiceImpl, resp *bean2.AppDetailContainer, ctx context.Context) map[string]interface{} {

	responseTree := resp.ResourceTree
	deploymentAppName := resp.AppName + "-" + resp.EnvironmentName

	for _, node := range responseTree["nodes"].(interface{}).([]interface{}) {
		currNode := node.(interface{}).(map[string]interface{})
		resName := util3.InterfaceToString(currNode["name"])
		resKind := util3.InterfaceToString(currNode["kind"])
		resGroup := util3.InterfaceToString(currNode["group"])
		resVersion := util3.InterfaceToString(currNode["version"])
		resNamespace := util3.InterfaceToString(currNode["namespace"])
		rQuery := &application.ApplicationResourceRequest{
			Name:         &deploymentAppName,
			ResourceName: &resName,
			Kind:         &resKind,
			Group:        &resGroup,
			Version:      &resVersion,
			Namespace:    &resNamespace,
		}
		ctx, _ := context.WithTimeout(ctx, 60*time.Second)
		if currNode["parentRefs"] == nil {
			t0 := time.Now()
			res, err := impl.acdClient.GetResource(ctx, rQuery)
			if err != nil {
				impl.logger.Errorw("error getting response from acdClient", "request", rQuery, "data", res, "timeTaken", time.Since(t0), "err", err)
				continue
			}
			if res.Manifest != nil {
				manifest, _ := gjson.Parse(*res.Manifest).Value().(map[string]interface{})
				replicas := util3.InterfaceToMapAdapter(manifest["spec"])["replicas"]
				if replicas != nil {
					currNode["canBeHibernated"] = true
				}
				annotations := util3.InterfaceToMapAdapter(manifest["metadata"])["annotations"]
				if annotations != nil {
					val := util3.InterfaceToMapAdapter(annotations)["hibernator.devtron.ai/replicas"]
					if val != nil {
						if util3.InterfaceToString(val) != "0" && util3.InterfaceToFloat(replicas) == 0 {
							currNode["isHibernated"] = true
						}
					}
				}

			}

		}
		node = currNode
	}
	return responseTree
}

func (impl InstalledAppServiceImpl) fetchResourceTreeForACD(rctx context.Context, cn http.CloseNotifier, appId int, envId, clusterId int, deploymentAppName, namespace string) (map[string]interface{}, error) {
	var resourceTree map[string]interface{}
	query := &application.ResourcesQuery{
		ApplicationName: &deploymentAppName,
	}
	ctx, cancel := context.WithCancel(rctx)
	if cn != nil {
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
		impl.logger.Errorw("error in getting acd token", "err", err)
		return resourceTree, err
	}
	ctx = context.WithValue(ctx, "token", acdToken)
	defer cancel()
	start := time.Now()
	resp, err := impl.acdClient.ResourceTree(ctx, query)
	elapsed := time.Since(start)
	impl.logger.Debugf("Time elapsed %s in fetching app-store installed application %s for environment %s", elapsed, deploymentAppName, envId)
	if err != nil {
		impl.logger.Errorw("service err, FetchAppDetailsForInstalledAppV2, fetching resource tree", "err", err, "installedAppId", appId, "envId", envId)
		err = &util.ApiError{
			Code:            constants.AppDetailResourceTreeNotFound,
			InternalMessage: "app detail fetched, failed to get resource tree from acd",
			UserMessage:     "app detail fetched, failed to get resource tree from acd",
		}
		return resourceTree, err
	}
	label := fmt.Sprintf("app.kubernetes.io/instance=%s", deploymentAppName)
	pods, err := impl.k8sApplicationService.GetPodListByLabel(clusterId, namespace, label)
	if err != nil {
		impl.logger.Errorw("error in getting pods by label", "err", err, "clusterId", clusterId, "namespace", namespace, "label", label)
		return resourceTree, err
	}
	ephemeralContainersMap := bean3.ExtractEphemeralContainers(pods)
	for _, metaData := range resp.PodMetadata {
		metaData.EphemeralContainers = ephemeralContainersMap[metaData.Name]
	}
	// TODO: using this resp.Status to update in app_status table
	resourceTree = util3.InterfaceToMapAdapter(resp)
	go func() {
		err = impl.appStatusService.UpdateStatusWithAppIdEnvId(appId, envId, resp.Status)
		if err != nil {
			impl.logger.Warnw("error in updating app status", "err", err, appId, "envId", envId)
		}
	}()
	impl.logger.Debugf("application %s in environment %s had status %+v\n", appId, envId, resp)
	return resourceTree, err
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
