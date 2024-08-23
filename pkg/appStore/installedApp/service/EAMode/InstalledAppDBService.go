/*
 * Copyright (c) 2020-2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package EAMode

import (
	helmBean "github.com/devtron-labs/devtron/api/helm-app/service/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	util4 "github.com/devtron-labs/devtron/pkg/appStore/util"
	bean3 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/deployment/common"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Pallinder/go-randomdata"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/adapter"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreRepo "github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/bean"
	util3 "github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type InstalledAppDBService interface {
	GetAll(filter *appStoreBean.AppStoreFilter) (appStoreBean.AppListDetail, error)
	CheckAppExists(appNames []*appStoreBean.AppNames) ([]*appStoreBean.AppNames, error)
	FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean2.AppDetailContainer, error)
	GetInstalledAppById(installedAppId int) (*appStoreRepo.InstalledApps, error)
	GetInstalledAppByClusterNamespaceAndName(clusterId int, namespace string, appName string) (*appStoreBean.InstallAppVersionDTO, error)
	GetInstalledAppByInstalledAppId(installedAppId int) (*appStoreBean.InstallAppVersionDTO, error)
	GetInstalledAppVersion(id int, userId int32) (*appStoreBean.InstallAppVersionDTO, error)
	CreateInstalledAppVersion(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreRepo.InstalledAppVersions, error)
	UpdateInstalledAppVersion(installedAppVersion *appStoreRepo.InstalledAppVersions, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreRepo.InstalledAppVersions, error)

	ChangeAppNameToDisplayNameForInstalledApp(installedApp *appStoreRepo.InstalledApps)
	GetReleaseInfo(appIdentifier *helmBean.AppIdentifier) (*appStoreBean.InstallAppVersionDTO, error)
	IsExternalAppLinkedToChartStore(appId int) (bool, []*appStoreRepo.InstalledApps, error)
	CreateNewAppEntryForAllInstalledApps(installedApps []*appStoreRepo.InstalledApps) error
}

type InstalledAppDBServiceImpl struct {
	Logger                        *zap.SugaredLogger
	InstalledAppRepository        appStoreRepo.InstalledAppRepository
	AppRepository                 app.AppRepository
	UserService                   user.UserService
	EnvironmentService            cluster.EnvironmentService
	InstalledAppRepositoryHistory appStoreRepo.InstalledAppVersionHistoryRepository
	deploymentConfigService       common.DeploymentConfigService
}

func NewInstalledAppDBServiceImpl(logger *zap.SugaredLogger,
	installedAppRepository appStoreRepo.InstalledAppRepository,
	appRepository app.AppRepository,
	userService user.UserService,
	environmentService cluster.EnvironmentService,
	installedAppRepositoryHistory appStoreRepo.InstalledAppVersionHistoryRepository,
	deploymentConfigService common.DeploymentConfigService) *InstalledAppDBServiceImpl {
	return &InstalledAppDBServiceImpl{
		Logger:                        logger,
		InstalledAppRepository:        installedAppRepository,
		AppRepository:                 appRepository,
		UserService:                   userService,
		EnvironmentService:            environmentService,
		InstalledAppRepositoryHistory: installedAppRepositoryHistory,
		deploymentConfigService:       deploymentConfigService,
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
		if util4.IsExternalChartStoreApp(appLocal.DisplayName) {
			//case of external app where display name is stored in app table
			helmAppResp.AppName = &appLocal.DisplayName
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

	deploymentConfig, err := impl.deploymentConfigService.GetConfigForHelmApps(installedAppVerison.InstalledApp.AppId, installedAppVerison.InstalledApp.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("error in getiting deployment config db object by appId and envId", "appId", installedAppVerison.InstalledApp.AppId, "envId", installedAppVerison.InstalledApp.EnvironmentId, "err", err)
		return bean2.AppDetailContainer{}, err
	}

	var chartName string
	if installedAppVerison.AppStoreApplicationVersion.AppStore.ChartRepoId != 0 {
		chartName = installedAppVerison.AppStoreApplicationVersion.AppStore.ChartRepo.Name
	} else {
		chartName = installedAppVerison.AppStoreApplicationVersion.AppStore.DockerArtifactStore.Id
	}

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
		ClusterName:                   installedAppVerison.InstalledApp.Environment.Cluster.ClusterName,
		DeploymentAppType:             deploymentConfig.DeploymentAppType,
		DeploymentAppDeleteRequest:    installedAppVerison.InstalledApp.DeploymentAppDeleteRequest,
		IsVirtualEnvironment:          installedAppVerison.InstalledApp.Environment.IsVirtualEnvironment,
		HelmReleaseInstallStatus:      helmReleaseInstallStatus,
		Status:                        status,
		DeploymentConfig:              deploymentConfig,
	}
	if util4.IsExternalChartStoreApp(installedAppVerison.InstalledApp.App.DisplayName) {
		deploymentContainer.AppName = installedAppVerison.InstalledApp.App.DisplayName
	}
	deploymentContainer.HelmPackageName = adapter.GetGeneratedHelmPackageName(deploymentContainer.AppName, deploymentContainer.EnvironmentName, installedAppVerison.InstalledApp.UpdatedOn)
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

func (impl *InstalledAppDBServiceImpl) GetInstalledAppById(installedAppId int) (*appStoreRepo.InstalledApps, error) {
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
		deploymentConfig, err := impl.deploymentConfigService.GetConfigForHelmApps(installedApp.AppId, installedApp.EnvironmentId)
		if err != nil {
			impl.Logger.Errorw("error in getiting deployment config db object by appId and envId", "appId", installedApp.AppId, "envId", installedApp.EnvironmentId, "err", err)
			return nil, err
		}
		return adapter.GenerateInstallAppVersionDTO(installedApp, installedAppVersion, deploymentConfig), nil
	}

	return nil, nil
}

func (impl *InstalledAppDBServiceImpl) GetInstalledAppByInstalledAppId(installedAppId int) (*appStoreBean.InstallAppVersionDTO, error) {
	installedAppVersion, err := impl.InstalledAppRepository.GetActiveInstalledAppVersionByInstalledAppId(installedAppId)
	if err != nil {
		return nil, err
	}
	installedApp := &installedAppVersion.InstalledApp
	deploymentConfig, err := impl.deploymentConfigService.GetConfigForHelmApps(installedApp.AppId, installedApp.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("error in getiting deployment config db object by appId and envId", "appId", installedApp.AppId, "envId", installedApp.EnvironmentId, "err", err)
		return nil, err
	}
	return adapter.GenerateInstallAppVersionDTO(installedApp, installedAppVersion, deploymentConfig), nil
}

func (impl *InstalledAppDBServiceImpl) GetInstalledAppVersion(id int, userId int32) (*appStoreBean.InstallAppVersionDTO, error) {
	model, err := impl.InstalledAppRepository.GetInstalledAppVersion(id)
	if err != nil {
		if util.IsErrNoRows(err) {
			return nil, &util.ApiError{HttpStatusCode: http.StatusBadRequest, Code: "400", UserMessage: "values are outdated. please fetch the latest version and try again", InternalMessage: err.Error()}
		}
		impl.Logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	deploymentConfig, err := impl.deploymentConfigService.GetConfigForHelmApps(model.InstalledApp.AppId, model.InstalledApp.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("error in getiting deployment config db object by appId and envId", "appId", model.InstalledApp.AppId, "envId", model.InstalledApp.EnvironmentId, "err", err)
		return nil, err
	}
	// update InstallAppVersion configurations
	installAppVersion := &appStoreBean.InstallAppVersionDTO{
		Id:                    model.Id,
		InstalledAppVersionId: model.Id,
		ValuesOverrideYaml:    model.ValuesYaml,
		ReferenceValueKind:    model.ReferenceValueKind,
		ReferenceValueId:      model.ReferenceValueId,
		InstalledAppId:        model.InstalledAppId,
		AppStoreVersion:       model.AppStoreApplicationVersionId, //check viki
		UserId:                userId,
	}

	// update App configurations
	adapter.UpdateAppDetails(installAppVersion, &model.InstalledApp.App)

	// update InstallApp configurations
	adapter.UpdateInstallAppDetails(installAppVersion, &model.InstalledApp, deploymentConfig)

	// update AppStoreApplication configurations
	adapter.UpdateAppStoreApplicationDetails(installAppVersion, &model.AppStoreApplicationVersion)

	environment, err := impl.EnvironmentService.GetExtendedEnvBeanById(installAppVersion.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("fetching environment error", "envId", installAppVersion.EnvironmentId, "err", err)
		return nil, err
	}
	// update environment details configurations
	adapter.UpdateAdditionalEnvDetails(installAppVersion, environment)
	return installAppVersion, err
}

func (impl *InstalledAppDBServiceImpl) CreateInstalledAppVersion(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreRepo.InstalledAppVersions, error) {
	installedAppVersion := adapter.NewInstalledAppVersionModel(installAppVersionRequest)
	_, err := impl.InstalledAppRepository.CreateInstalledAppVersion(installedAppVersion, tx)
	if err != nil {
		impl.Logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	return installedAppVersion, nil
}

func (impl *InstalledAppDBServiceImpl) UpdateInstalledAppVersion(installedAppVersion *appStoreRepo.InstalledAppVersions, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, tx *pg.Tx) (*appStoreRepo.InstalledAppVersions, error) {
	_, err := impl.InstalledAppRepository.UpdateInstalledAppVersion(installedAppVersion, tx)
	if err != nil {
		impl.Logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	return installedAppVersion, nil
}

func (impl *InstalledAppDBServiceImpl) ChangeAppNameToDisplayNameForInstalledApp(installedApp *appStoreRepo.InstalledApps) {
	installedApp.ChangeAppNameToDisplayName()
}

func (impl *InstalledAppDBServiceImpl) GetReleaseInfo(appIdentifier *helmBean.AppIdentifier) (*appStoreBean.InstallAppVersionDTO, error) {
	//for external-apps appName would be uniqueIdentifier
	appName := appIdentifier.GetUniqueAppNameIdentifier()
	installedAppVersionDto, err := impl.GetInstalledAppByClusterNamespaceAndName(appIdentifier.ClusterId, appIdentifier.Namespace, appName)
	if err != nil && !util.IsErrNoRows(err) {
		impl.Logger.Errorw("GetReleaseInfo, error in getting installed app by clusterId, namespace and appUniqueIdentifierName", "appIdentifier", appIdentifier, "appUniqueIdentifierName", appName, "error", err)
		return nil, err
	} else if util.IsErrNoRows(err) {
		// when app_name is not yet migrated to unique identifier
		installedAppVersionDto, err = impl.GetInstalledAppByClusterNamespaceAndName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
		if err != nil {
			impl.Logger.Errorw("GetReleaseInfo, error in getting installed app by clusterId, namespace and releaseName", "appIdentifier", appIdentifier, "error", err)
			return nil, err
		}
		//if dto found, check if release-info request is for the same namespace app as stored in installed_app because two ext-apps can have same
		//release name but in different namespaces, if they differ then release info request is for another ext-app with same name but in diff namespace
		if installedAppVersionDto != nil && installedAppVersionDto.Id > 0 && installedAppVersionDto.Namespace != appIdentifier.Namespace {
			installedAppVersionDto = nil
		}
	}
	return installedAppVersionDto, nil
}

// IsExternalAppLinkedToChartStore checks for an appId, if that app is linked to any chart-store app or not,
// if it's linked then it returns true along with all the installedApps linked to that appId
func (impl *InstalledAppDBServiceImpl) IsExternalAppLinkedToChartStore(appId int) (bool, []*appStoreRepo.InstalledApps, error) {
	installedApps, err := impl.InstalledAppRepository.FindInstalledAppsByAppId(appId)
	if err != nil && err != pg.ErrNoRows {
		impl.Logger.Errorw("IsExternalAppLinkedToChartStore, error in fetching installed apps by app id for external apps", "appId", appId, "err", err)
		return false, nil, err
	}
	if installedApps != nil && len(installedApps) > 0 {
		return true, installedApps, nil
	}
	return false, nil, nil
}

func (impl *InstalledAppDBServiceImpl) CreateNewAppEntryForAllInstalledApps(installedApps []*appStoreRepo.InstalledApps) error {
	// db operations
	dbConnection := impl.InstalledAppRepository.GetConnection()
	tx, err := dbConnection.Begin()
	if err != nil {
		return err
	}
	// Rollback tx on error.
	defer tx.Rollback()
	for _, installedApp := range installedApps {
		//check if for this unique identifier name an app already exists, if yes then continue
		appMetadata, err := impl.AppRepository.FindActiveByName(installedApp.GetUniqueAppNameIdentifier())
		if err != nil && !util.IsErrNoRows(err) {
			impl.Logger.Errorw("error in fetching app by unique app identifier", "appNameUniqueIdentifier", installedApp.GetUniqueAppNameIdentifier(), "err", err)
			return err
		}
		if appMetadata != nil && appMetadata.Id > 0 {
			//app already exists for this unique identifier hence not creating new app entry for this
			continue
		}

		appModel := &app.App{
			Active:          true,
			AppName:         installedApp.GetUniqueAppNameIdentifier(),
			TeamId:          installedApp.App.TeamId,
			AppType:         helper.ChartStoreApp,
			AppOfferingMode: installedApp.App.AppOfferingMode,
			DisplayName:     installedApp.App.AppName,
		}
		appModel.CreateAuditLog(bean3.SystemUserId)
		err = impl.AppRepository.SaveWithTxn(appModel, tx)
		if err != nil {
			impl.Logger.Errorw("error saving appModel", "err", err)
			return err
		}
		//updating the installedApp.AppId with new app entry
		installedApp.AppId = appModel.Id
		installedApp.UpdateAuditLog(bean3.SystemUserId)
		_, err = impl.InstalledAppRepository.UpdateInstalledApp(installedApp, tx)
		if err != nil {
			impl.Logger.Errorw("error saving updating installed app with new appId", "installedAppId", installedApp.Id, "err", err)
			return err
		}
	}

	tx.Commit()
	return nil
}
