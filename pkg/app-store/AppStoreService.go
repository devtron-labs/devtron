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

package app_store

import (
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	app_store "github.com/devtron-labs/devtron/pkg/app-store/repository"
	bean2 "github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
)

type AppStoreService interface {
	FindAllApps(filter *app_store.AppStoreFilter) ([]app_store.AppStoreWithVersion, error)
	FindChartDetailsById(id int) (AppStoreApplicationVersionResponse, error)
	FindChartVersionsByAppStoreId(appStoreId int) ([]AppStoreVersionsResponse, error)
	FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean.AppDetailContainer, error)
	GetReadMeByAppStoreApplicationVersionId(id int) (*ReadmeRes, error)
	SearchAppStoreChartByName(chartName string) ([]*app_store.ChartRepoSearch, error)
}

type AppStoreServiceImpl struct {
	logger                        *zap.SugaredLogger
	appStoreApplicationRepository app_store.AppStoreApplicationVersionRepository
	installedAppRepository        app_store.InstalledAppRepository
	userService                   user.UserService
}

func NewAppStoreServiceImpl(logger *zap.SugaredLogger,
	appStoreApplicationRepository app_store.AppStoreApplicationVersionRepository, installedAppRepository app_store.InstalledAppRepository,
	userService user.UserService) *AppStoreServiceImpl {
	return &AppStoreServiceImpl{
		logger:                        logger,
		appStoreApplicationRepository: appStoreApplicationRepository,
		installedAppRepository:        installedAppRepository,
		userService:                   userService,
	}
}

func (impl *AppStoreServiceImpl) FindAllApps(filter *app_store.AppStoreFilter) ([]app_store.AppStoreWithVersion, error) {
	appStoreApplications, err := impl.appStoreApplicationRepository.FindWithFilter(filter)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	return appStoreApplications, nil
}

func (impl *AppStoreServiceImpl) FindChartDetailsById(id int) (AppStoreApplicationVersionResponse, error) {
	chartDetails, err := impl.appStoreApplicationRepository.FindById(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return AppStoreApplicationVersionResponse{}, err
	}
	appStoreApplicationVersion := AppStoreApplicationVersionResponse{
		Id:                      chartDetails.Id,
		Version:                 chartDetails.Version,
		AppVersion:              chartDetails.AppVersion,
		Created:                 chartDetails.Created,
		Deprecated:              chartDetails.Deprecated,
		Description:             chartDetails.Description,
		Digest:                  chartDetails.Digest,
		Icon:                    chartDetails.Icon,
		Name:                    chartDetails.Name,
		ChartName:               chartDetails.AppStore.ChartRepo.Name,
		AppStoreApplicationName: chartDetails.AppStore.Name,
		Home:                    chartDetails.Home,
		Source:                  chartDetails.Source,
		ValuesYaml:              chartDetails.ValuesYaml,
		ChartYaml:               chartDetails.ChartYaml,
		AppStoreId:              chartDetails.AppStoreId,
		Latest:                  chartDetails.Latest,
		CreatedOn:               chartDetails.CreatedOn,
		UpdatedOn:               chartDetails.UpdatedOn,
		RawValues:               chartDetails.RawValues,
		Readme:                  chartDetails.Readme,
		IsChartRepoActive:       chartDetails.AppStore.ChartRepo.Active,
	}
	return appStoreApplicationVersion, nil
}

func (impl *AppStoreServiceImpl) FindChartVersionsByAppStoreId(appStoreId int) ([]AppStoreVersionsResponse, error) {
	appStoreVersions, err := impl.appStoreApplicationRepository.FindVersionsByAppStoreId(appStoreId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return nil, err
	}
	var appStoreVersionsResponse []AppStoreVersionsResponse
	for _, a := range appStoreVersions {
		res := AppStoreVersionsResponse{
			Id:      a.Id,
			Version: a.Version,
		}
		appStoreVersionsResponse = append(appStoreVersionsResponse, res)
	}
	return appStoreVersionsResponse, nil
}

func (impl *AppStoreServiceImpl) FindAppDetailsForAppstoreApplication(installedAppId, envId int) (bean.AppDetailContainer, error) {
	installedAppVerison, err := impl.installedAppRepository.GetInstalledAppVersionByInstalledAppIdAndEnvId(installedAppId, envId)
	if err != nil {
		impl.logger.Error(err)
		return bean.AppDetailContainer{}, err
	}
	deploymentContainer := bean.DeploymentDetailContainer{
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
		LastDeployedTime:              installedAppVerison.UpdatedOn.Format(bean2.LayoutRFC3339),
		Namespace:                     installedAppVerison.InstalledApp.Environment.Namespace,
		Deprecated:                    installedAppVerison.AppStoreApplicationVersion.Deprecated,
	}
	userInfo, err := impl.userService.GetByIdIncludeDeleted(installedAppVerison.AuditLog.UpdatedBy)
	if err != nil {
		impl.logger.Errorw("error fetching user info", "err", err)
		return bean.AppDetailContainer{}, err
	}
	deploymentContainer.LastDeployedBy = userInfo.EmailId
	appDetail := bean.AppDetailContainer{
		DeploymentDetailContainer: deploymentContainer,
	}
	return appDetail, nil
}

func (impl *AppStoreServiceImpl) GetReadMeByAppStoreApplicationVersionId(id int) (*ReadmeRes, error) {
	appVersion, err := impl.appStoreApplicationRepository.GetReadMeById(id)
	if err != nil {
		return nil, err
	}
	readme := &ReadmeRes{
		AppStoreApplicationVersionId: appVersion.Id,
		Readme:                       appVersion.Readme,
	}
	return readme, nil
}

func (impl *AppStoreServiceImpl) SearchAppStoreChartByName(chartName string) ([]*app_store.ChartRepoSearch, error) {
	appStoreApplications, err := impl.appStoreApplicationRepository.SearchAppStoreChartByName(chartName)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	return appStoreApplications, nil
}
