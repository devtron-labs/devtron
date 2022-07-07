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
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDeploymentFullMode "github.com/devtron-labs/devtron/pkg/appStore/deployment/fullMode"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/values/service"
	repository5 "github.com/devtron-labs/devtron/pkg/cluster/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	repository4 "github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	util2 "github.com/devtron-labs/devtron/pkg/util"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/ktrysmt/go-bitbucket"

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
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	application2 "github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/repository"
	"github.com/devtron-labs/devtron/client/pubsub"
	repository3 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/bean"
	cluster2 "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/go-pg/pg"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

const (
	DEFAULT_ENVIRONMENT_OR_NAMESPACE_OR_PROJECT = "devtron"
	CLUSTER_COMPONENT_DIR_PATH                  = "/cluster/component"
)

type InstalledAppService interface {
	GetAll(filter *appStoreBean.AppStoreFilter) (openapi.AppList, error)
	DeployBulk(chartGroupInstallRequest *appStoreBean.ChartGroupInstallRequest) (*appStoreBean.ChartGroupInstallAppRes, error)
	performDeployStage(appId int, userId int32) (*appStoreBean.InstallAppVersionDTO, error)
	CheckAppExists(appNames []*appStoreBean.AppNames) ([]*appStoreBean.AppNames, error)
	DeployDefaultChartOnCluster(bean *cluster2.ClusterBean, userId int32) (bool, error)
	FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean2.AppDetailContainer, error)
	UpdateInstalledAppVersionStatus(application v1alpha1.Application) (bool, error)
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
	pubsubClient                         *pubsub.PubSubClient
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
	pubsubClient *pubsub.PubSubClient,
	tokenCache *util2.TokenCache,
	chartGroupDeploymentRepository repository2.ChartGroupDeploymentRepository,
	envService cluster2.EnvironmentService, argoK8sClient argocdServer.ArgoK8sClient,
	gitFactory *util.GitFactory, aCDAuthConfig *util2.ACDAuthConfig, gitOpsRepository repository3.GitOpsConfigRepository, userService user.UserService,
	appStoreDeploymentFullModeService appStoreDeploymentFullMode.AppStoreDeploymentFullModeService,
	appStoreDeploymentService AppStoreDeploymentService,
	installedAppRepositoryHistory repository2.InstalledAppVersionHistoryRepository) (*InstalledAppServiceImpl, error) {
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
	}
	err := util3.AddStream(impl.pubsubClient.JetStrCtxt, util3.ORCHESTRATOR_STREAM)
	if err != nil {
		return nil, err
	}
	err = impl.Subscribe()
	if err != nil {
		return nil, err
	}
	return impl, nil
}

func (impl InstalledAppServiceImpl) GetAll(filter *appStoreBean.AppStoreFilter) (openapi.AppList, error) {
	applicationType := "DEVTRON-CHART-STORE"
	var clusterIdsConverted []int32
	for _, clusterId := range filter.ClusterIds {
		clusterIdsConverted = append(clusterIdsConverted, int32(clusterId))
	}
	installedAppsResponse := openapi.AppList{
		ApplicationType: &applicationType,
		ClusterIds:      &clusterIdsConverted,
	}
	installedApps, err := impl.installedAppRepository.GetAllInstalledApps(filter)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return installedAppsResponse, err
	}
	var helmAppsResponse []openapi.HelmApp
	for _, a := range installedApps {
		appLocal := a // copied data from here because value is passed as reference
		if appLocal.TeamId == 0 {
			//skipping entries for empty projectId
			continue
		}
		appId := strconv.Itoa(appLocal.Id)
		projectId := int32(appLocal.TeamId)
		envId := int32(appLocal.EnvironmentId)
		clusterId := int32(appLocal.ClusterId)
		environmentDetails := openapi.AppEnvironmentDetail{
			EnvironmentName: &appLocal.EnvironmentName,
			EnvironmentId:   &envId,
			Namespace:       &appLocal.Namespace,
			ClusterName:     &appLocal.ClusterName,
			ClusterId:       &clusterId,
		}
		helmAppResp := openapi.HelmApp{
			AppName:           &appLocal.AppName,
			ChartName:         &appLocal.AppStoreApplicationName,
			AppId:             &appId,
			ProjectId:         &projectId,
			EnvironmentDetail: &environmentDetails,
			ChartAvatar:       &appLocal.Icon,
			LastDeployedAt:    &appLocal.UpdatedOn,
		}
		helmAppsResponse = append(helmAppsResponse, helmAppResp)
	}
	installedAppsResponse.HelmApps = &helmAppsResponse
	return installedAppsResponse, nil
}

//converts db object to bean
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
		installAppVersionDTO, err = impl.appStoreDeploymentService.AppStoreDeployOperationDB(installAppVersionDTO, tx)
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

//generate unique installation ID using APPID
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

func (impl InstalledAppServiceImpl) performDeployStage(installedAppVersionId int, userId int32) (*appStoreBean.InstallAppVersionDTO, error) {
	ctx, err := impl.tokenCache.BuildACDSynchContext()
	if err != nil {
		return nil, err
	}

	isGitOpsConfigured := false
	gitOpsConfig, err := impl.gitOpsRepository.GetGitOpsConfigActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GetGitOpsConfigActive, error while getting", "err", err)
		return nil, err
	}
	if gitOpsConfig != nil && gitOpsConfig.Id > 0 {
		isGitOpsConfigured = true
	}

	installedAppVersion, err := impl.appStoreDeploymentService.GetInstalledAppVersion(installedAppVersionId, userId)
	if err != nil {
		return nil, err
	}

	if !isGitOpsConfigured {
		// if git-ops not configured, bulk chart deployment not supported
		_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.TRIGGER_ERROR)
		if err != nil {
			impl.logger.Errorw("error", "err", err)
			return nil, err
		}
		err = &util.ApiError{Code: "400", HttpStatusCode: 200, UserMessage: "unable to found git-ops configuration in cluster, please configure"}
		return nil, err
	}

	chartGitAttr := &util.ChartGitAttribute{}
	if installedAppVersion.Status == appStoreBean.DEPLOY_INIT ||
		installedAppVersion.Status == appStoreBean.ENQUEUED ||
		installedAppVersion.Status == appStoreBean.QUE_ERROR ||
		installedAppVersion.Status == appStoreBean.GIT_ERROR {
		//step 2 git operation pull push
		installedAppVersion, chartGitAttrDB, err := impl.appStoreDeploymentFullModeService.AppStoreDeployOperationGIT(installedAppVersion)
		if err != nil {
			impl.logger.Errorw(" error", "err", err)
			_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.GIT_ERROR)
			if err != nil {
				impl.logger.Errorw(" error", "err", err)
				return nil, err
			}
			return nil, err
		}
		impl.logger.Infow("GIT SUCCESSFUL", "chartGitAttrDB", chartGitAttrDB)
		_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.GIT_SUCCESS)
		if err != nil {
			impl.logger.Errorw(" error", "err", err)
			return nil, err
		}
		chartGitAttr.RepoUrl = chartGitAttrDB.RepoUrl
		chartGitAttr.ChartLocation = chartGitAttrDB.ChartLocation
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
		bitbucketRepoOptions := &bitbucket.RepositoryOptions{
			Owner:    gitOpsConfigBitbucket.BitBucketWorkspaceId,
			Project:  gitOpsConfigBitbucket.BitBucketProjectKey,
			RepoSlug: installedAppVersion.AppStoreName,
		}
		repoUrl, err := impl.gitFactory.Client.GetRepoUrl(installedAppVersion.AppStoreName, bitbucketRepoOptions)
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
		_, err = impl.appStoreDeploymentFullModeService.AppStoreDeployOperationACD(installedAppVersion, chartGitAttr, ctx)
		if err != nil {
			impl.logger.Errorw(" error", "chartGitAttr", chartGitAttr, "err", err)
			_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.ACD_ERROR)
			if err != nil {
				impl.logger.Errorw(" error", "err", err)
				return nil, err
			}
			return nil, err
		}
		impl.logger.Infow("ACD SUCCESSFUL", "chartGitAttr", chartGitAttr)
		_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.ACD_SUCCESS)
		if err != nil {
			impl.logger.Errorw(" error", "err", err)
			return nil, err
		}
	} else {
		impl.logger.Infow("DB and GIT and ACD operation already done for this app and env. process has been completed", "installedAppId", installedAppVersion.InstalledAppId, "existing status", installedAppVersion.Status)
	}
	//step 4 db operation status triggered
	_, err = impl.appStoreDeploymentService.AppStoreDeployOperationStatusUpdate(installedAppVersion.InstalledAppId, appStoreBean.DEPLOY_SUCCESS)
	if err != nil {
		impl.logger.Errorw(" error", "err", err)
		return nil, err
	}

	// create build history for chart on default component
	err = impl.appStoreDeploymentService.UpdateInstallAppVersionHistory(installedAppVersion)
	if err != nil {
		impl.logger.Errorw("error on creating history for chart deployment", "error", err)
		return nil, err
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
		payload := &appStoreBean.DeployPayload{InstalledAppVersionId: versions.InstalledAppVersionId}
		data, err := json.Marshal(payload)
		if err != nil {
			status = appStoreBean.QUE_ERROR
		} else {
			err := util3.AddStream(impl.pubsubClient.JetStrCtxt, util3.ORCHESTRATOR_STREAM)

			if err != nil {
				impl.logger.Errorw("Error while adding stream.", "error", err)
			}
			//Generate random string for passing as Header Id in message
			randString := "MsgHeaderId-" + util3.Generate(10)
			_, err = impl.pubsubClient.JetStrCtxt.Publish(util3.BULK_APPSTORE_DEPLOY_TOPIC, data, nats.MsgId(randString))
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
	_, err := impl.pubsubClient.JetStrCtxt.QueueSubscribe(util3.BULK_APPSTORE_DEPLOY_TOPIC, util3.BULK_APPSTORE_DEPLOY_GROUP, func(msg *nats.Msg) {
		impl.logger.Debug("cd stage event received")
		defer msg.Ack()
		deployPayload := &appStoreBean.DeployPayload{}
		err := json.Unmarshal([]byte(string(msg.Data)), &deployPayload)
		if err != nil {
			impl.logger.Error("Error while unmarshalling deployPayload json object", "error", err)
			return
		}
		impl.logger.Debugw("deployPayload:", "deployPayload", deployPayload)
		//using userId 1 - for system user
		_, err = impl.performDeployStage(deployPayload.InstalledAppVersionId, 1)
		if err != nil {
			impl.logger.Errorw("error in performing deploy stage", "deployPayload", deployPayload, "err", err)
		}
	}, nats.Durable(util3.BULK_APPSTORE_DEPLOY_DURABLE), nats.DeliverLast(), nats.ManualAck(), nats.BindStream(util3.ORCHESTRATOR_STREAM))
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
		installAppVersionDTO, err = impl.appStoreDeploymentService.AppStoreDeployOperationDB(installAppVersionDTO, tx)
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
		_, err := impl.performDeployStage(versions.InstalledAppVersionId, chartGroupInstallRequest.UserId)
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

func (impl InstalledAppServiceImpl) UpdateInstalledAppVersionStatus(application v1alpha1.Application) (bool, error) {
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
		versionHistory.Status = application.Status.Health.Status
		versionHistory.UpdatedOn = time.Now()
		versionHistory.UpdatedBy = 1
		impl.installedAppRepositoryHistory.UpdateInstalledAppVersionHistory(versionHistory, tx)
	}
	err = tx.Commit()
	if err != nil {
		impl.logger.Errorw("error while committing transaction to db", "error", err)
		return isHealthy, err
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
