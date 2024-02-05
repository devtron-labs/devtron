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

package appStoreDeploymentCommon

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/adapter"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"go.uber.org/zap"
	"io/ioutil"
	"k8s.io/helm/pkg/chartutil"
	"path"
	"sigs.k8s.io/yaml"
)

type AppStoreDeploymentCommonService interface {
	// GetValuesString will return values string from the given valuesOverrideYaml
	GetValuesString(chartName, valuesOverrideYaml string) (string, error)
	// GetRequirementsString will return requirement dependencies for the given appStoreVersionId
	GetRequirementsString(appStoreVersionId int) (string, error)
	// CreateChartProxyAndGetPath parse chart in local directory and returns path of local dir and values.yaml
	CreateChartProxyAndGetPath(chartCreateRequest *util.ChartCreateRequest) (*util.ChartCreateResponse, error)
	// BuildChartWithValuesAndRequirementsConfig
	BuildChartWithValuesAndRequirementsConfig(appName, valuesString, requirementsString, chartName, chartVersion string) (chartBytesArr []byte, err error)
}

type AppStoreDeploymentCommonServiceImpl struct {
	logger                               *zap.SugaredLogger
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	chartTemplateService                 util.ChartTemplateService
}

func NewAppStoreDeploymentCommonServiceImpl(
	logger *zap.SugaredLogger,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	chartTemplateService util.ChartTemplateService) *AppStoreDeploymentCommonServiceImpl {
	return &AppStoreDeploymentCommonServiceImpl{
		logger:                               logger,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		chartTemplateService:                 chartTemplateService,
	}
}

func (impl AppStoreDeploymentCommonServiceImpl) GetValuesString(chartName, valuesOverrideYaml string) (string, error) {

	ValuesOverrideByte, err := yaml.YAMLToJSON([]byte(valuesOverrideYaml))
	if err != nil {
		impl.logger.Errorw("")
	}

	var dat map[string]interface{}
	err = json.Unmarshal(ValuesOverrideByte, &dat)
	if err != nil {
		impl.logger.Errorw("error in unmarshalling values override byte", "err", err)
		return "", err
	}

	valuesMap := make(map[string]map[string]interface{})
	valuesMap[chartName] = dat
	valuesByte, err := json.Marshal(valuesMap)
	if err != nil {
		impl.logger.Errorw("error in marshaling", "err", err)
		return "", err
	}
	return string(valuesByte), nil
}

func (impl AppStoreDeploymentCommonServiceImpl) GetRequirementsString(appStoreVersionId int) (string, error) {
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(appStoreVersionId)
	if err != nil {
		impl.logger.Errorw("fetching error", "err", err)
		return "", err
	}
	dependency := appStoreBean.Dependency{
		Name:    appStoreAppVersion.AppStore.Name,
		Version: appStoreAppVersion.Version,
	}
	if appStoreAppVersion.AppStore.ChartRepo != nil {
		dependency.Repository = appStoreAppVersion.AppStore.ChartRepo.Url
	}
	var dependencies []appStoreBean.Dependency
	dependencies = append(dependencies, dependency)
	requirementDependencies := &appStoreBean.Dependencies{
		Dependencies: dependencies,
	}
	requirementDependenciesByte, err := json.Marshal(requirementDependencies)
	if err != nil {
		return "", err
	}
	requirementDependenciesByte, err = yaml.JSONToYAML(requirementDependenciesByte)
	if err != nil {
		return "", err
	}
	return string(requirementDependenciesByte), nil
}

func (impl AppStoreDeploymentCommonServiceImpl) CreateChartProxyAndGetPath(chartCreateRequest *util.ChartCreateRequest) (*util.ChartCreateResponse, error) {
	ChartCreateResponse := &util.ChartCreateResponse{}
	valid, err := chartutil.IsChartDir(chartCreateRequest.ChartPath)
	if err != nil || !valid {
		impl.logger.Errorw("invalid base chart", "dir", chartCreateRequest.ChartPath, "err", err)
		return ChartCreateResponse, err
	}
	chartCreateResponse, err := impl.chartTemplateService.BuildChartProxyForHelmApps(chartCreateRequest)
	if err != nil {
		impl.logger.Errorw("Error in building chart proxy", "err", err)
		return chartCreateResponse, err
	}
	return chartCreateResponse, nil
}

func (impl AppStoreDeploymentCommonServiceImpl) BuildChartWithValuesAndRequirementsConfig(appName, valuesString, requirementsString, chartName, chartVersion string) (chartBytesArr []byte, err error) {

	chartBytesArr = make([]byte, 0)
	chartCreateRequest := adapter.ParseChartCreateRequest(appName, false)
	chartCreateResponse, err := impl.CreateChartProxyAndGetPath(chartCreateRequest)
	if err != nil {
		impl.logger.Errorw("error in building chart", "err", err)
	}

	valuesFilePath := path.Join(chartCreateResponse.BuiltChartPath, "values.yaml")
	err = ioutil.WriteFile(valuesFilePath, []byte(valuesString), 0600)
	if err != nil {
		return chartBytesArr, nil
	}

	requirementsFilePath := path.Join(chartCreateResponse.BuiltChartPath, "requirements.yaml")
	err = ioutil.WriteFile(requirementsFilePath, []byte(requirementsString), 0600)
	if err != nil {
		return chartBytesArr, nil
	}

	chartBytesArr, err = impl.chartTemplateService.LoadChartInBytes(chartCreateResponse.BuiltChartPath, true, chartName, chartVersion)
	if err != nil {
		impl.logger.Errorw("error in loading chart in bytes", "err", err)
		return chartBytesArr, nil
	}

	return chartBytesArr, err
}
