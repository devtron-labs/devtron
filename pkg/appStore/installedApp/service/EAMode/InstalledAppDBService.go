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

package EAMode

import (
	"fmt"
	"github.com/Pallinder/go-randomdata"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/adapter"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/bean"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type InstalledAppDBService interface {
	GetAll(filter *appStoreBean.AppStoreFilter) (appStoreBean.AppListDetail, error)
	CheckAppExists(appNames []*appStoreBean.AppNames) ([]*appStoreBean.AppNames, error)
	FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean2.AppDetailContainer, error)
	CheckAppExistsByInstalledAppId(installedAppId int) (*repository2.InstalledApps, error)
	GetInstalledAppByClusterNamespaceAndName(clusterId int, namespace string, appName string) (*appStoreBean.InstallAppVersionDTO, error)
	GetInstalledAppByInstalledAppId(installedAppId int) (*appStoreBean.InstallAppVersionDTO, error)
}

type InstalledAppDBServiceImpl struct {
	Logger                        *zap.SugaredLogger
	InstalledAppRepository        repository2.InstalledAppRepository
	AppRepository                 app.AppRepository
	UserService                   user.UserService
	InstalledAppRepositoryHistory repository2.InstalledAppVersionHistoryRepository
}

func NewInstalledAppDBServiceImpl(logger *zap.SugaredLogger,
	installedAppRepository repository2.InstalledAppRepository,
	appRepository app.AppRepository,
	userService user.UserService,
	installedAppRepositoryHistory repository2.InstalledAppVersionHistoryRepository) *InstalledAppDBServiceImpl {
	return &InstalledAppDBServiceImpl{
		Logger:                        logger,
		InstalledAppRepository:        installedAppRepository,
		AppRepository:                 appRepository,
		UserService:                   userService,
		InstalledAppRepositoryHistory: installedAppRepositoryHistory,
	}
}

func (impl *InstalledAppDBServiceImpl) GetAll(filter *appStoreBean.AppStoreFilter) (appStoreBean.AppListDetail, error) {
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
	installedApps, err := impl.InstalledAppRepository.GetAllInstalledApps(filter)
	middleware.AppListingDuration.WithLabelValues("getAllInstalledApps", "helm").Observe(time.Since(start).Seconds())
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Error(err)
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

func (impl *InstalledAppDBServiceImpl) CheckAppExists(appNames []*appStoreBean.AppNames) ([]*appStoreBean.AppNames, error) {
	if len(appNames) == 0 {
		return nil, nil
	}
	var names []string
	for _, appName := range appNames {
		names = append(names, appName.Name)
	}

	apps, err := impl.AppRepository.CheckAppExists(names)
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

func (impl *InstalledAppDBServiceImpl) FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean2.AppDetailContainer, error) {
	installedAppVerison, err := impl.InstalledAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		impl.Logger.Error(err)
		return bean2.AppDetailContainer{}, err
	}
	helmReleaseInstallStatus, status, err := impl.InstalledAppRepository.GetHelmReleaseStatusConfigByInstalledAppId(installedAppVerison.InstalledAppId)
	if err != nil {
		impl.Logger.Errorw("error in getting helm release status from db", "err", err)
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
	userInfo, err := impl.UserService.GetByIdIncludeDeleted(installedAppVerison.AuditLog.UpdatedBy)
	if err != nil {
		impl.Logger.Errorw("error fetching user info", "err", err)
		return bean2.AppDetailContainer{}, err
	}
	deploymentContainer.LastDeployedBy = userInfo.EmailId
	appDetail := bean2.AppDetailContainer{
		DeploymentDetailContainer: deploymentContainer,
	}
	return appDetail, nil
}

func (impl *InstalledAppDBServiceImpl) CheckAppExistsByInstalledAppId(installedAppId int) (*repository2.InstalledApps, error) {
	installedApp, err := impl.InstalledAppRepository.GetInstalledApp(installedAppId)
	if err != nil {
		return nil, err
	}
	return installedApp, err
}

func (impl *InstalledAppDBServiceImpl) GetInstalledAppByClusterNamespaceAndName(clusterId int, namespace string, appName string) (*appStoreBean.InstallAppVersionDTO, error) {
	installedApp, err := impl.InstalledAppRepository.GetInstalledApplicationByClusterIdAndNamespaceAndAppName(clusterId, namespace, appName)
	if err != nil {
		if err == pg.ErrNoRows {
			impl.Logger.Warnw("no installed apps found", "clusterId", clusterId)
			return nil, nil
		} else {
			impl.Logger.Errorw("error while fetching installed apps", "clusterId", clusterId, "error", err)
			return nil, err
		}
	}

	if installedApp.Id > 0 {
		installedAppVersion, err := impl.InstalledAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedApp.Id, installedApp.EnvironmentId)
		if err != nil {
			return nil, err
		}
		return adapter.GenerateInstallAppVersionDTO(installedApp, installedAppVersion), nil
	}

	return nil, nil
}

func (impl *InstalledAppDBServiceImpl) GetInstalledAppByInstalledAppId(installedAppId int) (*appStoreBean.InstallAppVersionDTO, error) {
	installedAppVersion, err := impl.InstalledAppRepository.GetActiveInstalledAppVersionByInstalledAppId(installedAppId)
	if err != nil {
		return nil, err
	}
	installedApp := &installedAppVersion.InstalledApp
	return adapter.GenerateInstallAppVersionDTO(installedApp, installedAppVersion), nil
}
