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

package app_store_discover

import (
	"github.com/devtron-labs/devtron/internal/util"
	app_store_bean "github.com/devtron-labs/devtron/pkg/app-store/bean"
	app_store_discover_repository "github.com/devtron-labs/devtron/pkg/app-store/discover/repository"
	"go.uber.org/zap"
)

type AppStoreService interface {
	FindAllApps(filter *app_store_bean.AppStoreFilter) ([]app_store_bean.AppStoreWithVersion, error)
	FindChartDetailsById(id int) (app_store_bean.AppStoreApplicationVersionResponse, error)
	FindChartVersionsByAppStoreId(appStoreId int) ([]app_store_bean.AppStoreVersionsResponse, error)
	GetReadMeByAppStoreApplicationVersionId(id int) (*app_store_bean.ReadmeRes, error)
	SearchAppStoreChartByName(chartName string) ([]*app_store_bean.ChartRepoSearch, error)
}

type AppStoreServiceImpl struct {
	logger                        *zap.SugaredLogger
	appStoreApplicationRepository app_store_discover_repository.AppStoreApplicationVersionRepository
}

func NewAppStoreServiceImpl(logger *zap.SugaredLogger,
	appStoreApplicationRepository app_store_discover_repository.AppStoreApplicationVersionRepository) *AppStoreServiceImpl {
	return &AppStoreServiceImpl{
		logger:                        logger,
		appStoreApplicationRepository: appStoreApplicationRepository,
	}
}

func (impl *AppStoreServiceImpl) FindAllApps(filter *app_store_bean.AppStoreFilter) ([]app_store_bean.AppStoreWithVersion, error) {
	appStoreApplications, err := impl.appStoreApplicationRepository.FindWithFilter(filter)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	return appStoreApplications, nil
}

func (impl *AppStoreServiceImpl) FindChartDetailsById(id int) (app_store_bean.AppStoreApplicationVersionResponse, error) {
	chartDetails, err := impl.appStoreApplicationRepository.FindById(id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return app_store_bean.AppStoreApplicationVersionResponse{}, err
	}
	appStoreApplicationVersion := app_store_bean.AppStoreApplicationVersionResponse{
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

func (impl *AppStoreServiceImpl) FindChartVersionsByAppStoreId(appStoreId int) ([]app_store_bean.AppStoreVersionsResponse, error) {
	appStoreVersions, err := impl.appStoreApplicationRepository.FindVersionsByAppStoreId(appStoreId)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Error(err)
		return nil, err
	}
	var appStoreVersionsResponse []app_store_bean.AppStoreVersionsResponse
	for _, a := range appStoreVersions {
		res := app_store_bean.AppStoreVersionsResponse{
			Id:      a.Id,
			Version: a.Version,
		}
		appStoreVersionsResponse = append(appStoreVersionsResponse, res)
	}
	return appStoreVersionsResponse, nil
}

func (impl *AppStoreServiceImpl) GetReadMeByAppStoreApplicationVersionId(id int) (*app_store_bean.ReadmeRes, error) {
	appVersion, err := impl.appStoreApplicationRepository.GetReadMeById(id)
	if err != nil {
		return nil, err
	}
	readme := &app_store_bean.ReadmeRes{
		AppStoreApplicationVersionId: appVersion.Id,
		Readme:                       appVersion.Readme,
	}
	return readme, nil
}

func (impl *AppStoreServiceImpl) SearchAppStoreChartByName(chartName string) ([]*app_store_bean.ChartRepoSearch, error) {
	appStoreApplications, err := impl.appStoreApplicationRepository.SearchAppStoreChartByName(chartName)
	if err != nil && !util.IsErrNoRows(err) {
		return nil, err
	}
	return appStoreApplications, nil
}
