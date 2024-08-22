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

package appStoreDeploymentCommon

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chartutil"
	"net/url"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

type AppStoreDeploymentCommonService interface {
	// GetValuesString will return values string from the given valuesOverrideYaml
	GetValuesString(chartName, valuesOverrideYaml string) (string, error)
	// GetRequirementsString will return requirement dependencies for the given appStoreVersionId
	GetRequirementsString(appStoreApplicationVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) (string, error)
	// CreateChartProxyAndGetPath parse chart in local directory and returns path of local dir and values.yaml
	CreateChartProxyAndGetPath(chartCreateRequest *util.ChartCreateRequest) (*util.ChartCreateResponse, error)
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

func (impl AppStoreDeploymentCommonServiceImpl) GetRequirementsString(appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) (string, error) {

	dependency := appStoreBean.Dependency{
		Name:    appStoreAppVersion.AppStore.Name,
		Version: appStoreAppVersion.Version,
	}
	if appStoreAppVersion.AppStore.ChartRepo != nil {
		dependency.Repository = appStoreAppVersion.AppStore.ChartRepo.Url
	} else if appStoreAppVersion.AppStore.DockerArtifactStore != nil {
		dependency.Repository = appStoreAppVersion.AppStore.DockerArtifactStore.RegistryURL
		repositoryURL, repositoryName, err := sanitizeRepoNameAndURLForOCIRepo(dependency.Repository, dependency.Name)
		if err != nil {
			impl.logger.Errorw("error in getting sanitized repository name and url", "repositoryURL", repositoryURL, "repositoryName", repositoryName, "err", err)
			return "", err
		}
		dependency.Repository = repositoryURL
		dependency.Name = repositoryName
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

func sanitizeRepoNameAndURLForOCIRepo(repositoryURL, repositoryName string) (string, string, error) {

	if !strings.HasPrefix(repositoryURL, "oci://") {
		repositoryURL = strings.TrimSpace(repositoryURL)
		parsedUrl, err := url.Parse(repositoryURL)
		if err != nil {
			repositoryURL = fmt.Sprintf("//%s", repositoryURL)
			parsedUrl, err = url.Parse(repositoryURL)
			if err != nil {
				return "", "", err
			}
		}
		parsedHost := strings.TrimSpace(parsedUrl.Host)
		parsedUrlPath := strings.TrimSpace(parsedUrl.Path)
		repositoryURL = fmt.Sprintf("%s://%s", "oci", filepath.Join(parsedHost, parsedUrlPath))
	}

	idx := strings.LastIndex(repositoryName, "/")
	if idx != -1 {
		name := repositoryName[idx+1:]
		url := fmt.Sprintf("%s/%s", repositoryURL, repositoryName[0:idx])
		repositoryURL = url
		repositoryName = name
	}

	return repositoryURL, repositoryName, nil
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
