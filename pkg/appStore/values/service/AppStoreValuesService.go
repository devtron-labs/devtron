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
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	appStoreValuesRepository "github.com/devtron-labs/devtron/pkg/appStore/values/repository"
	"go.uber.org/zap"
	"time"
)

type AppStoreValuesService interface {
	CreateAppStoreVersionValues(model *appStoreBean.AppStoreVersionValuesDTO) (*appStoreBean.AppStoreVersionValuesDTO, error)
	UpdateAppStoreVersionValues(model *appStoreBean.AppStoreVersionValuesDTO) (*appStoreBean.AppStoreVersionValuesDTO, error)
	FindValuesByIdAndKind(referenceId int, kind string) (*appStoreBean.AppStoreVersionValuesDTO, error)
	DeleteAppStoreVersionValues(appStoreValueId int) (bool, error)

	FindValuesByAppStoreId(appStoreId int, installedAppVersionId int) (*appStoreBean.AppSotoreVersionDTOWrapper, error)
	FindValuesByAppStoreIdAndReferenceType(appStoreVersionId int, referenceType string) ([]*appStoreBean.AppStoreVersionValuesDTO, error)
	GetSelectedChartMetaData(req *ChartMetaDataRequestWrapper) ([]*ChartMetaDataResponse, error)
}

type AppStoreValuesServiceImpl struct {
	logger                          *zap.SugaredLogger
	appStoreApplicationRepository   appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	installedAppRepository          repository.InstalledAppRepository
	appStoreVersionValuesRepository appStoreValuesRepository.AppStoreVersionValuesRepository
}

func NewAppStoreValuesServiceImpl(logger *zap.SugaredLogger,
	appStoreApplicationRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository, installedAppRepository repository.InstalledAppRepository,
	appStoreVersionValuesRepository appStoreValuesRepository.AppStoreVersionValuesRepository) *AppStoreValuesServiceImpl {
	return &AppStoreValuesServiceImpl{
		logger:                          logger,
		appStoreApplicationRepository:   appStoreApplicationRepository,
		installedAppRepository:          installedAppRepository,
		appStoreVersionValuesRepository: appStoreVersionValuesRepository,
	}
}

func (impl AppStoreValuesServiceImpl) CreateAppStoreVersionValues(request *appStoreBean.AppStoreVersionValuesDTO) (*appStoreBean.AppStoreVersionValuesDTO, error) {
	model := &appStoreValuesRepository.AppStoreVersionValues{
		Name:                         request.Name,
		ValuesYaml:                   request.Values,
		AppStoreApplicationVersionId: request.AppStoreVersionId,
		ReferenceType:                appStoreBean.REFERENCE_TYPE_TEMPLATE,
	}
	model.CreatedOn = time.Now()
	model.UpdatedOn = time.Now()
	model.CreatedBy = request.UserId
	model.UpdatedBy = request.UserId
	app, err := impl.appStoreVersionValuesRepository.CreateAppStoreVersionValues(model)
	if err != nil {
		impl.logger.Errorw("error while insert", "error", err)
		return nil, err
	}
	request.Id = app.Id
	return request, nil
}

func (impl AppStoreValuesServiceImpl) UpdateAppStoreVersionValues(request *appStoreBean.AppStoreVersionValuesDTO) (*appStoreBean.AppStoreVersionValuesDTO, error) {
	model, err := impl.appStoreVersionValuesRepository.FindById(request.Id)
	if err != nil && !util.IsErrNoRows(err) {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	} else if util.IsErrNoRows(err) {
		impl.logger.Errorw("invalid request for values update 404", "req", request, "error", err)
		return nil, err
	}

	model.Name = request.Name
	model.ValuesYaml = request.Values
	model.UpdatedBy = request.UserId
	model.UpdatedOn = time.Now()
	app, err := impl.appStoreVersionValuesRepository.UpdateAppStoreVersionValues(model)
	if err != nil {
		impl.logger.Errorw("error while updating", "error", err)
		return nil, err
	}
	request.Id = app.Id
	return request, nil
}

func (impl AppStoreValuesServiceImpl) FindValuesByIdAndKind(referenceId int, kind string) (*appStoreBean.AppStoreVersionValuesDTO, error) {
	if kind == appStoreBean.REFERENCE_TYPE_TEMPLATE {
		appStoreVersionValues, err := impl.appStoreVersionValuesRepository.FindById(referenceId)
		if err != nil {
			impl.logger.Errorw("error while fetching from db", "error", err)
			return nil, err
		}
		filterItem, err := impl.adapter(appStoreVersionValues)
		if err != nil {
			impl.logger.Errorw("error while casting ", "error", err)
			return nil, err
		}
		return filterItem, err
	} else if kind == appStoreBean.REFERENCE_TYPE_DEFAULT {
		applicationVersion, err := impl.appStoreApplicationRepository.FindById(referenceId)
		if err != nil {
			impl.logger.Errorw("error while fetching AppStoreApplicationVersion from db", "error", err)
			return nil, err
		}
		valDto := &appStoreBean.AppStoreVersionValuesDTO{
			Name:              appStoreBean.REFERENCE_TYPE_DEFAULT,
			Id:                applicationVersion.Id,
			Values:            applicationVersion.RawValues,
			ChartVersion:      applicationVersion.Version,
			AppStoreVersionId: applicationVersion.Id,
		}
		return valDto, err
	} else if kind == appStoreBean.REFERENCE_TYPE_DEPLOYED {
		installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersion(referenceId)
		if err != nil {
			impl.logger.Errorw("error in fetching installed App", "id", referenceId, "err", err)
		}
		valDto := &appStoreBean.AppStoreVersionValuesDTO{
			Name:              appStoreBean.REFERENCE_TYPE_DEPLOYED,
			Id:                installedAppVersion.Id,
			Values:            installedAppVersion.ValuesYaml,
			ChartVersion:      installedAppVersion.AppStoreApplicationVersion.Version,
			AppStoreVersionId: installedAppVersion.AppStoreApplicationVersionId,
		}
		return valDto, err
	} else if kind == appStoreBean.REFERENCE_TYPE_EXISTING {
		installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersionAny(referenceId)
		if err != nil {
			impl.logger.Errorw("error in fetching installed App", "id", referenceId, "err", err)
		}
		valDto := &appStoreBean.AppStoreVersionValuesDTO{
			Name:              appStoreBean.REFERENCE_TYPE_EXISTING,
			Id:                installedAppVersion.Id,
			Values:            installedAppVersion.ValuesYaml,
			ChartVersion:      installedAppVersion.AppStoreApplicationVersion.Version,
			AppStoreVersionId: installedAppVersion.AppStoreApplicationVersionId,
		}
		return valDto, err
	} else {
		impl.logger.Errorw("unsupported kind", "kind", kind)
		return nil, fmt.Errorf("unsupported kind %s", kind)
	}

}

func (impl AppStoreValuesServiceImpl) DeleteAppStoreVersionValues(appStoreValueId int) (bool, error) {
	model, err := impl.appStoreVersionValuesRepository.FindById(appStoreValueId)
	if err != nil {
		impl.logger.Errorw("error while fetching app store version values app", "error", err)
		return false, err
	}
	model.Deleted = true
	_, err = impl.appStoreVersionValuesRepository.DeleteAppStoreVersionValues(model)
	if err != nil {
		impl.logger.Errorw("error while delete", "error", err)
		return false, err
	}
	return true, nil
}

func (impl AppStoreValuesServiceImpl) FindValuesByAppStoreId(appStoreId int, installedAppVersionId int) (*appStoreBean.AppSotoreVersionDTOWrapper, error) {
	appStoreVersionValues, err := impl.appStoreVersionValuesRepository.FindValuesByAppStoreId(appStoreId)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	var appStoreVersionValuesDTO []*appStoreBean.AppStoreVersionValuesDTO
	for _, item := range appStoreVersionValues {
		filterItem, err := impl.adapter(item)
		if err != nil {
			impl.logger.Errorw("error while casting ", "error", err)
			return nil, err
		}
		appStoreVersionValuesDTO = append(appStoreVersionValuesDTO, filterItem)
	}
	templateVal := &appStoreBean.AppStoreVersionValuesCategoryWiseDTO{
		Values: appStoreVersionValuesDTO,
		Kind:   appStoreBean.REFERENCE_TYPE_TEMPLATE,
	}
	// default val
	appVersions, err := impl.appStoreApplicationRepository.FindChartVersionByAppStoreId(appStoreId)
	if err != nil {
		impl.logger.Errorw("error while  getting default version", "error", err)
		return nil, err
	}
	defaultVal := &appStoreBean.AppStoreVersionValuesCategoryWiseDTO{
		Kind: appStoreBean.REFERENCE_TYPE_DEFAULT,
	}
	for _, appVersion := range appVersions {
		defaultValTemplate := &appStoreBean.AppStoreVersionValuesDTO{
			Id:           appVersion.Id,
			Name:         "Default",
			ChartVersion: appVersion.Version,
		}
		defaultVal.Values = append(defaultVal.Values, defaultValTemplate)
	}

	// installed app
	installedAppVersions, err := impl.installedAppRepository.GetInstalledAppVersionByAppStoreId(appStoreId)
	if err != nil {
		impl.logger.Errorw("error in fetching installed app", "appStoreVersionId", appStoreId, "err", err)
		return nil, err
	}
	installedVal := &appStoreBean.AppStoreVersionValuesCategoryWiseDTO{
		Values: []*appStoreBean.AppStoreVersionValuesDTO{},
		Kind:   appStoreBean.REFERENCE_TYPE_DEPLOYED,
	}
	for _, installedAppVersion := range installedAppVersions {
		appStoreVersion := &appStoreBean.AppStoreVersionValuesDTO{
			Id:                installedAppVersion.Id,
			AppStoreVersionId: installedAppVersion.AppStoreApplicationVersionId,
			Name:              installedAppVersion.InstalledApp.App.AppName,
			ChartVersion:      installedAppVersion.AppStoreApplicationVersion.Version,
			EnvironmentName:   installedAppVersion.InstalledApp.Environment.Name,
		}
		installedVal.Values = append(installedVal.Values, appStoreVersion)
	}

	existingVal := &appStoreBean.AppStoreVersionValuesCategoryWiseDTO{
		Values: []*appStoreBean.AppStoreVersionValuesDTO{},
		Kind:   appStoreBean.REFERENCE_TYPE_EXISTING,
	}
	if installedAppVersionId > 0 {
		installedAppVersion, err := impl.installedAppRepository.GetInstalledAppVersion(installedAppVersionId)
		if err != nil {
			impl.logger.Errorw("error in fetching installed app", "appStoreVersionId", appStoreId, "err", err)
			return nil, err
		}
		appStoreVersion := &appStoreBean.AppStoreVersionValuesDTO{
			Id:                installedAppVersion.Id,
			AppStoreVersionId: installedAppVersion.AppStoreApplicationVersionId,
			Name:              installedAppVersion.InstalledApp.App.AppName,
			ChartVersion:      installedAppVersion.AppStoreApplicationVersion.Version,
			EnvironmentName:   installedAppVersion.InstalledApp.Environment.Name,
		}
		existingVal.Values = append(existingVal.Values, appStoreVersion)
	}

	///-------- installed app end
	res := &appStoreBean.AppSotoreVersionDTOWrapper{Values: []*appStoreBean.AppStoreVersionValuesCategoryWiseDTO{defaultVal, templateVal, installedVal, existingVal}} //order is important.
	return res, err
}

func (impl AppStoreValuesServiceImpl) FindValuesByAppStoreIdAndReferenceType(appStoreId int, referenceType string) ([]*appStoreBean.AppStoreVersionValuesDTO, error) {
	appStoreVersionValues, err := impl.appStoreVersionValuesRepository.FindValuesByAppStoreIdAndReferenceType(appStoreId, referenceType)
	if err != nil {
		impl.logger.Errorw("error while fetching from db", "error", err)
		return nil, err
	}
	var appStoreVersionValuesDTO []*appStoreBean.AppStoreVersionValuesDTO
	for _, item := range appStoreVersionValues {
		filterItem, err := impl.adapter(item)
		if err != nil {
			impl.logger.Errorw("error while casting ", "error", err)
			return nil, err
		}
		appStoreVersionValuesDTO = append(appStoreVersionValuesDTO, filterItem)
	}
	return appStoreVersionValuesDTO, err
}

//converts db object to bean
func (impl AppStoreValuesServiceImpl) adapter(values *appStoreValuesRepository.AppStoreVersionValues) (*appStoreBean.AppStoreVersionValuesDTO, error) {

	version := ""
	if values.AppStoreApplicationVersion != nil {
		version = values.AppStoreApplicationVersion.Version
	}
	return &appStoreBean.AppStoreVersionValuesDTO{
		Name:              values.Name,
		Id:                values.Id,
		Values:            values.ValuesYaml,
		ChartVersion:      version,
		AppStoreVersionId: values.AppStoreApplicationVersionId,
	}, nil
}

/*func (impl AppStoreValuesServiceImpl) adaptorForValuesCategoryWise(values *appStore.AppStoreVersionValues) (val *AppStoreVersionValuesCategoryWiseDTO) {
	version := ""
	if values.AppStoreApplicationVersion != nil {
		version = values.AppStoreApplicationVersion.Version
	}

	valDto:= &AppStoreVersionValuesDTO{
		Name:              values.Name,
		Id:                values.Id,
		Values:            values.ValuesYaml,
		ChartVersion:      version,
		AppStoreVersionId: values.AppStoreApplicationVersionId,
	}
	val = &AppStoreVersionValuesCategoryWiseDTO{
		Values:valDto
	}
	return val
}
*/
type ChartMetaDataRequest struct {
	Kind  string `json:"kind"`
	Value int    `json:"value"`
}
type ChartMetaDataRequestWrapper struct {
	Values []*ChartMetaDataRequest `json:"values"`
}

type ChartMetaDataResponse struct {
	//version, name, rep, char val name,
	ChartName                    string `json:"chartName"`
	ChartRepoName                string `json:"chartRepoName"`
	AppStoreApplicationVersionId int    `json:"appStoreApplicationVersionId"`
	Icon                         string `json:"icon"`
	Kind                         string `json:"kind"`
}

func (impl AppStoreValuesServiceImpl) GetSelectedChartMetaData(req *ChartMetaDataRequestWrapper) ([]*ChartMetaDataResponse, error) {
	var defaultValuesId []int
	var templateValuesId []int
	var deployedValuesId []int
	for _, v := range req.Values {
		switch v.Kind {
		case appStoreBean.REFERENCE_TYPE_DEFAULT:
			defaultValuesId = append(defaultValuesId, v.Value)
		case appStoreBean.REFERENCE_TYPE_TEMPLATE:
			templateValuesId = append(templateValuesId, v.Value)
		case appStoreBean.REFERENCE_TYPE_DEPLOYED:
			deployedValuesId = append(deployedValuesId, v.Value)
		default:
			impl.logger.Warnw("unsupported kind", "kind", v.Kind)
		}
	}
	appVersions, err := impl.appStoreApplicationRepository.FindByIds(defaultValuesId)
	if err != nil {
		return nil, err
	}
	var res []*ChartMetaDataResponse
	for _, appversion := range appVersions {
		chartMeta := &ChartMetaDataResponse{
			ChartName:                    appversion.AppStore.Name,
			ChartRepoName:                appversion.AppStore.ChartRepo.Name,
			AppStoreApplicationVersionId: appversion.Id,
			Icon:                         appversion.Icon,
			Kind:                         appStoreBean.REFERENCE_TYPE_DEFAULT,
		}
		res = append(res, chartMeta)
	}
	return res, err
}
