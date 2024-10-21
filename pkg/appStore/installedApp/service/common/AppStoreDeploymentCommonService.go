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
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/helm-app/service"
	openapi2 "github.com/devtron-labs/devtron/api/openapi/openapiClient"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/bean/workflow/cdWorkflow"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/adapter"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/EAMode"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/go-pg/pg"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/chartutil"
	"k8s.io/utils/pointer"
	"net/http"
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
	GetDeploymentHistoryFromDB(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error)
	GetDeploymentHistoryInfoFromDB(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error)
}

type AppStoreDeploymentCommonServiceImpl struct {
	logger                               *zap.SugaredLogger
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository
	chartTemplateService                 util.ChartTemplateService
	helmAppService                       service.HelmAppService
	userService                          user.UserService
	installedAppDBService                EAMode.InstalledAppDBService
}

func NewAppStoreDeploymentCommonServiceImpl(
	logger *zap.SugaredLogger,
	appStoreApplicationVersionRepository appStoreDiscoverRepository.AppStoreApplicationVersionRepository,
	chartTemplateService util.ChartTemplateService,
	userService user.UserService,
	helmAppService service.HelmAppService,
	installedAppDBService EAMode.InstalledAppDBService,
) *AppStoreDeploymentCommonServiceImpl {
	return &AppStoreDeploymentCommonServiceImpl{
		logger:                               logger,
		appStoreApplicationVersionRepository: appStoreApplicationVersionRepository,
		chartTemplateService:                 chartTemplateService,
		userService:                          userService,
		helmAppService:                       helmAppService,
		installedAppDBService:                installedAppDBService,
	}
}
func (impl *AppStoreDeploymentCommonServiceImpl) GetDeploymentHistoryFromDB(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO) (*gRPC.HelmAppDeploymentHistory, error) {
	result := &gRPC.HelmAppDeploymentHistory{}
	var history []*gRPC.HelmAppDeploymentDetail
	//TODO - response setup
	installedAppVersions, err := impl.installedAppDBService.GetInstalledAppVersionByInstalledAppIdMeta(installedApp.InstalledAppId)
	if err != nil {
		if err == pg.ErrNoRows {
			return nil, &util.ApiError{HttpStatusCode: http.StatusBadRequest, Code: "400", UserMessage: "values are outdated. please fetch the latest version and try again", InternalMessage: err.Error()}
		}
		impl.logger.Errorw("error while fetching installed version", "error", err)
		return result, err
	}
	for _, installedAppVersionModel := range installedAppVersions {
		sources, jsonErr := impl.getSourcesFromManifest(installedAppVersionModel.AppStoreApplicationVersion.ChartYaml)
		if jsonErr != nil {
			impl.logger.Errorw("error while fetching sources", "error", jsonErr)
			//continues here, skip error in case found issue on fetching source
		}
		versionHistory, err := impl.installedAppDBService.GetInstalledAppVersionHistoryByVersionId(installedAppVersionModel.Id)
		if err != nil && err != pg.ErrNoRows {
			impl.logger.Errorw("error while fetching installed version history", "error", err)
			return result, err
		}
		for _, updateHistory := range versionHistory {
			if len(updateHistory.Message) == 0 && updateHistory.Status == cdWorkflow.WorkflowFailed {
				// if message is empty and status is failed, then update message from helm release status config. for migration purpose
				// updateHistory.HelmReleaseStatusConfig stores the failed description, if async operation failed
				updateHistory.Message = impl.migrateDeploymentHistoryMessage(ctx, updateHistory)
			}
			emailId, err := impl.userService.GetEmailById(updateHistory.CreatedBy)
			if err != nil {
				impl.logger.Errorw("error while fetching user Details", "error", err)
				return result, err
			}
			history = append(history, adapter.BuildDeploymentHistory(installedAppVersionModel, sources, updateHistory, emailId))
		}
	}
	if len(history) == 0 {
		history = make([]*gRPC.HelmAppDeploymentDetail, 0)
	}
	result.DeploymentHistory = history
	return result, err
}
func (impl *AppStoreDeploymentCommonServiceImpl) GetDeploymentHistoryInfoFromDB(ctx context.Context, installedApp *appStoreBean.InstallAppVersionDTO, version int32) (*openapi.HelmAppDeploymentManifestDetail, error) {
	values := &openapi.HelmAppDeploymentManifestDetail{}
	_, span := otel.Tracer("orchestrator").Start(ctx, "installedAppRepositoryHistory.GetInstalledAppVersionHistory")
	versionHistory, err := impl.installedAppDBService.GetInstalledAppVersionHistory(int(version))
	span.End()
	if err != nil {
		impl.logger.Errorw("error while fetching installed version history", "error", err)
		return nil, err
	}
	values.ValuesYaml = &versionHistory.ValuesYamlRaw

	envId := int32(installedApp.EnvironmentId)
	clusterId := int32(installedApp.ClusterId)
	appStoreApplicationVersionId, err := impl.installedAppDBService.GetAppStoreApplicationVersionIdByInstalledAppVersionHistoryId(int(version))
	appStoreVersionId := pointer.Int32(int32(appStoreApplicationVersionId))

	// as virtual environment doesn't exist on actual cluster, we will use default cluster for running helm template command
	if installedApp.IsVirtualEnvironment {
		clusterId = appStoreBean.DEFAULT_CLUSTER_ID
		installedApp.Namespace = appStoreBean.DEFAULT_NAMESPACE
	}

	manifestRequest := openapi2.TemplateChartRequest{
		EnvironmentId:                &envId,
		ClusterId:                    &clusterId,
		Namespace:                    &installedApp.Namespace,
		ReleaseName:                  &installedApp.AppName,
		AppStoreApplicationVersionId: appStoreVersionId,
		ValuesYaml:                   values.ValuesYaml,
	}

	_, span = otel.Tracer("orchestrator").Start(ctx, "helmAppService.TemplateChart")
	templateChart, manifestErr := impl.helmAppService.TemplateChart(ctx, &manifestRequest)
	span.End()
	manifest := templateChart.GetManifest()

	if manifestErr != nil {
		impl.logger.Errorw("error in genetating manifest for argocd app", "err", manifestErr)
	} else {
		values.Manifest = &manifest
	}
	return values, err
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

func (impl *AppStoreDeploymentCommonServiceImpl) migrateDeploymentHistoryMessage(ctx context.Context, updateHistory *repository.InstalledAppVersionHistory) (helmInstallStatusMsg string) {
	_, span := otel.Tracer("orchestrator").Start(ctx, "FullModeDeploymentServiceImp.migrateDeploymentHistoryMessage")
	defer span.End()
	helmInstallStatusMsg = updateHistory.Message
	helmInstallStatus := &appStoreBean.HelmReleaseStatusConfig{}
	jsonErr := json.Unmarshal([]byte(updateHistory.HelmReleaseStatusConfig), helmInstallStatus)
	if jsonErr != nil {
		impl.logger.Errorw("error while unmarshal helm release status config", "helmReleaseStatusConfig", updateHistory.HelmReleaseStatusConfig, "error", jsonErr)
		return helmInstallStatusMsg
	} else if helmInstallStatus.ErrorInInstallation {
		helmInstallStatusMsg = fmt.Sprintf("Deployment failed: %v", helmInstallStatus.Message)
		dbErr := impl.installedAppDBService.UpdateDeploymentHistoryMessage(updateHistory.Id, helmInstallStatusMsg)
		if dbErr != nil {
			impl.logger.Errorw("error while updating deployment history helmInstallStatusMsg", "error", dbErr)
		}
		return helmInstallStatusMsg
	}
	return helmInstallStatusMsg
}

func (impl *AppStoreDeploymentCommonServiceImpl) getSourcesFromManifest(chartYaml string) ([]string, error) {
	var b map[string]interface{}
	var sources []string
	err := json.Unmarshal([]byte(chartYaml), &b)
	if err != nil {
		impl.logger.Errorw("error while unmarshal chart yaml", "error", err)
		return sources, err
	}
	if b != nil && b["sources"] != nil {
		slice := b["sources"].([]interface{})
		for _, item := range slice {
			sources = append(sources, item.(string))
		}
	}
	return sources, nil
}
