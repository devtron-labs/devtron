/*
 * Copyright (c) 2024. Devtron Inc.
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

package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/adapter"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/bean"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	gitBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
	validationBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation/bean"
	chartRefBean "github.com/devtron-labs/devtron/pkg/deployment/manifest/deploymentTemplate/chartRef/bean"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/sliceUtil"
	"github.com/google/go-github/github"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/xanzy/go-gitlab"
	"helm.sh/helm/v3/pkg/chart"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
)

type InstalledAppGitOpsService interface {
	// GitOpsOperations performs git operations specific to helm app deployments
	// If appStoreBean.InstallAppVersionDTO has GitOpsRepoURL -> EMPTY string; then it will auto create repository and update into appStoreBean.InstallAppVersionDTO
	GitOpsOperations(manifestResponse *bean.AppStoreManifestResponse, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*bean.AppStoreGitOpsResponse, error)
	// GenerateManifest returns bean.AppStoreManifestResponse required in GitOps
	GenerateManifest(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreApplicationVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) (manifestResponse *bean.AppStoreManifestResponse, err error)
	// GenerateManifestAndPerformGitOperations is a wrapper function for both GenerateManifest and GitOpsOperations
	GenerateManifestAndPerformGitOperations(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreApplicationVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) (*bean.AppStoreGitOpsResponse, error)
	// UpdateAppGitOpsOperations internally uses
	// GitOpsOperations (If Repo is deleted OR Repo migration is required) OR
	// git.GitOperationService.CommitValues (If repo exists and Repo migration is not needed)
	// functions to perform GitOps during upgrade deployments (GitOps based Helm Apps)
	UpdateAppGitOpsOperations(manifest *bean.AppStoreManifestResponse, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, monoRepoMigrationRequired bool, commitRequirements bool) (*bean.AppStoreGitOpsResponse, error)
	ValidateCustomGitOpsConfig(request validationBean.ValidateGitOpsRepoRequest) (string, bool, error)
	CreateArgoRepoSecretIfNeeded(appStoreApplicationVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error
}

// GitOpsOperations handles all git operations for Helm App; and ensures that the return param bean.AppStoreGitOpsResponse is not nil
func (impl *FullModeDeploymentServiceImpl) GitOpsOperations(manifestResponse *bean.AppStoreManifestResponse, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*bean.AppStoreGitOpsResponse, error) {
	appStoreGitOpsResponse := &bean.AppStoreGitOpsResponse{}
	chartGitAttribute, githash, err := impl.createGitOpsRepoAndPushChart(installAppVersionRequest, manifestResponse.ChartResponse.BuiltChartPath, manifestResponse.ValuesConfig)
	if err != nil {
		impl.Logger.Errorw("Error in pushing chart to git", "err", err)
		return appStoreGitOpsResponse, err
	}
	appStoreGitOpsResponse.ChartGitAttribute = chartGitAttribute
	appStoreGitOpsResponse.GitHash = githash
	return appStoreGitOpsResponse, nil
}

func (impl *FullModeDeploymentServiceImpl) GenerateManifest(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreApplicationVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) (manifestResponse *bean.AppStoreManifestResponse, err error) {
	manifestResponse = &bean.AppStoreManifestResponse{}
	values, dependencies, err := impl.getValuesAndRequirement(installAppVersionRequest, appStoreApplicationVersion)
	if err != nil {
		impl.Logger.Errorw("error in fetching values and requirements.yaml config while generating manifest", "err", err)
		return manifestResponse, err
	}
	chartCreateRequest := adapter.ParseChartCreateRequest(installAppVersionRequest.AppName, dependencies, true)
	chartCreateResponse, err := impl.createChartProxyAndGetPath(chartCreateRequest)
	if err != nil {
		impl.Logger.Errorw("Error in building chart while generating manifest", "err", err)
		return manifestResponse, err
	}
	var chartMetaData *chart.Metadata
	if chartCreateResponse != nil {
		if chartCreateResponse.ChartMetaData != nil {
			chartCreateResponse.ChartMetaData.Dependencies = dependencies
		}
		chartMetaData = chartCreateResponse.ChartMetaData
	}
	// valuesConfig and chartMetaDataConfig's ChartConfig object contains ChartRepoName which is extracted from gitOpsRepoUrl
	// that resides in the db and not from the current orchestrator cm prefix and appName.
	valuesConfig, chartMetaDataConfig, err := impl.getValuesAndChartMetaDataForGitConfig(installAppVersionRequest, values, chartMetaData)
	if err != nil {
		impl.Logger.Errorw("error in getting values and requirements config for git commit", "err", err)
		return manifestResponse, err
	}
	manifestResponse.ChartResponse = chartCreateResponse
	manifestResponse.ValuesConfig = valuesConfig
	manifestResponse.ChartMetaDataConfig = chartMetaDataConfig
	return manifestResponse, nil
}

func (impl *FullModeDeploymentServiceImpl) GenerateManifestAndPerformGitOperations(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreApplicationVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) (*bean.AppStoreGitOpsResponse, error) {
	appStoreGitOpsResponse := &bean.AppStoreGitOpsResponse{}
	manifest, err := impl.GenerateManifest(installAppVersionRequest, appStoreApplicationVersion)
	if err != nil {
		impl.Logger.Errorw("error in performing manifest and git operations", "err", err)
		return nil, err
	}
	gitOpsResponse, err := impl.GitOpsOperations(manifest, installAppVersionRequest)
	if err != nil {
		impl.Logger.Errorw("error in performing gitops operation", "err", err)
		return nil, err
	}
	installAppVersionRequest.GitHash = gitOpsResponse.GitHash
	appStoreGitOpsResponse.ChartGitAttribute = gitOpsResponse.ChartGitAttribute
	return appStoreGitOpsResponse, nil
}

func (impl *FullModeDeploymentServiceImpl) UpdateAppGitOpsOperations(manifest *bean.AppStoreManifestResponse,
	installAppVersionRequest *appStoreBean.InstallAppVersionDTO, monoRepoMigrationRequired bool, updateDependencies bool) (*bean.AppStoreGitOpsResponse, error) {
	var requirementsCommitErr, valuesCommitErr error
	var gitHash string
	if monoRepoMigrationRequired {
		// overriding GitOpsRepoURL to migrate to new repo
		installAppVersionRequest.GitOpsRepoURL = ""
		return impl.GitOpsOperations(manifest, installAppVersionRequest)
	}

	gitOpsResponse := &bean.AppStoreGitOpsResponse{}
	ctx := context.Background()
	if updateDependencies {
		// update dependency if chart or chart version is changed
		_, _, requirementsCommitErr = impl.gitOperationService.CommitValues(ctx, manifest.ChartMetaDataConfig)
		gitHash, _, valuesCommitErr = impl.gitOperationService.CommitValues(ctx, manifest.ValuesConfig)
	} else {
		cloneChartToGitRequest := adapter.ParseChartGitPushRequest(installAppVersionRequest, "")
		migrateDependencies, err := impl.shouldMigrateProxyChartDependencies(cloneChartToGitRequest)
		if err != nil {
			impl.Logger.Errorw("error in checking if proxy chart dependencies should be migrated", "err", err)
			return nil, err
		}
		if migrateDependencies {
			_, _, requirementsCommitErr = impl.gitOperationService.CommitValues(ctx, manifest.ChartMetaDataConfig)
		}
		// only values are changed in update, so commit values config
		gitHash, _, valuesCommitErr = impl.gitOperationService.CommitValues(ctx, manifest.ValuesConfig)
	}

	if valuesCommitErr != nil || requirementsCommitErr != nil {
		noTargetFoundForValues, _ := impl.parseGitRepoErrorResponse(valuesCommitErr)
		noTargetFoundForRequirements, _ := impl.parseGitRepoErrorResponse(requirementsCommitErr)
		if noTargetFoundForRequirements || noTargetFoundForValues {
			//create repo again and try again  -  auto fix
			_, _, err := impl.createGitOpsRepo(impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(installAppVersionRequest.GitOpsRepoURL),
				installAppVersionRequest.GetTargetRevision(), installAppVersionRequest.UserId)
			if err != nil {
				impl.Logger.Errorw("error in creating GitOps repo for valuesCommitErr or requirementsCommitErr", "gitRepoUrl", installAppVersionRequest.GitOpsRepoURL)
				return nil, err
			}
			return impl.GitOpsOperations(manifest, installAppVersionRequest)
		}
		impl.Logger.Errorw("error in performing GitOps for upgrade deployment", "ValuesCommitErr", valuesCommitErr, "RequirementsCommitErr", requirementsCommitErr)
		return nil, fmt.Errorf("error in committing values and requirements to git repository")
	}
	gitOpsResponse.GitHash = gitHash
	gitOpsResponse.ChartGitAttribute = &commonBean.ChartGitAttribute{
		RepoUrl:        installAppVersionRequest.GitOpsRepoURL,
		TargetRevision: installAppVersionRequest.GetTargetRevision(),
		ChartLocation:  installAppVersionRequest.ACDAppName,
	}
	return gitOpsResponse, nil
}

// parseGitRepoErrorResponse will return noTargetFound (if git API returns 404 status)
func (impl *FullModeDeploymentServiceImpl) parseGitRepoErrorResponse(err error) (bool, error) {
	//update values yaml in chart
	noTargetFound := false
	if err != nil {
		if errorResponse, ok := err.(*github.ErrorResponse); ok && errorResponse.Response.StatusCode == http.StatusNotFound {
			impl.Logger.Errorw("no content found while updating git repo on github, do auto fix", "error", err)
			noTargetFound = true
		}
		if errorResponse, ok := err.(azuredevops.WrappedError); ok && *errorResponse.StatusCode == http.StatusNotFound {
			impl.Logger.Errorw("no content found while updating git repo on azure, do auto fix", "error", err)
			noTargetFound = true
		}
		if errorResponse, ok := err.(*azuredevops.WrappedError); ok && *errorResponse.StatusCode == http.StatusNotFound {
			impl.Logger.Errorw("no content found while updating git repo on azure, do auto fix", "error", err)
			noTargetFound = true
		}
		if errorResponse, ok := err.(*gitlab.ErrorResponse); ok && errorResponse.Response.StatusCode == http.StatusNotFound {
			impl.Logger.Errorw("no content found while updating git repo gitlab, do auto fix", "error", err)
			noTargetFound = true
		}
		if strings.Contains(err.Error(), git.BitbucketRepoNotFoundError.Error()) {
			impl.Logger.Errorw("no content found while updating git repo bitbucket, do auto fix", "error", err)
			noTargetFound = true
		}
	}
	return noTargetFound, err
}

// createGitOpsRepoAndPushChart is a wrapper for creating GitOps repo and pushing chart to created repo
func (impl *FullModeDeploymentServiceImpl) createGitOpsRepoAndPushChart(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, builtChartPath string, valuesConfig *git.ChartConfig) (*commonBean.ChartGitAttribute, string, error) {
	// in case of monorepo migration installAppVersionRequest.GitOpsRepoURL is ""
	if len(installAppVersionRequest.GitOpsRepoURL) == 0 {
		gitOpsConfigStatus, err := impl.gitOpsConfigReadService.GetGitOpsConfigActive()
		if err != nil {
			return nil, "", err
		}
		installedApp, err := impl.installedAppRepository.GetInstalledApp(installAppVersionRequest.InstalledAppId)
		if err != nil {
			impl.Logger.Errorw("service err, installedApp", "installedAppId", installAppVersionRequest.InstalledAppId, "err", err)
			return nil, "", err
		}
		if gitOpsConfigStatus.AllowCustomRepository && installedApp.IsCustomRepository {
			errMSg := fmt.Sprintf("Invalid request! Git repository URL is not found for installed app '%s'", installAppVersionRequest.AppName)
			return nil, "", util.NewApiError(http.StatusBadRequest, errMSg, errMSg)
		}
		gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoName(installAppVersionRequest.AppName)
		gitOpsRepoURL, isNew, err := impl.createGitOpsRepo(gitOpsRepoName, installAppVersionRequest.GetTargetRevision(), installAppVersionRequest.UserId)
		if err != nil {
			impl.Logger.Errorw("Error in creating gitops repo for ", "appName", installAppVersionRequest.AppName, "err", err)
			return nil, "", err
		}
		installAppVersionRequest.GitOpsRepoURL = gitOpsRepoURL
		installAppVersionRequest.IsCustomRepository = false
		installAppVersionRequest.IsNewGitOpsRepo = isNew

	}
	pushChartToGitRequest := adapter.ParseChartGitPushRequest(installAppVersionRequest, builtChartPath)
	chartGitAttribute, commitHash, err := impl.gitOperationService.PushChartToGitOpsRepoForHelmApp(context.Background(), pushChartToGitRequest, valuesConfig)
	if err != nil {
		impl.Logger.Errorw("error in pushing chart to git", "err", err)
		return nil, "", err
	}
	return chartGitAttribute, commitHash, err
}

// createGitOpsRepo creates a gitOps repo with readme
func (impl *FullModeDeploymentServiceImpl) createGitOpsRepo(gitOpsRepoName string, targetRevision string, userId int32) (string, bool, error) {
	bitbucketMetadata, err := impl.gitOpsConfigReadService.GetBitbucketMetadata()
	if err != nil {
		impl.Logger.Errorw("error in getting bitbucket metadata", "err", err)
		return "", false, err
	}
	//getting user name & emailId for commit author data
	gitRepoRequest := &bean2.GitOpsConfigDto{
		GitRepoName:          gitOpsRepoName,
		TargetRevision:       targetRevision,
		Description:          "helm chart for " + gitOpsRepoName,
		BitBucketWorkspaceId: bitbucketMetadata.BitBucketWorkspaceId,
		BitBucketProjectKey:  bitbucketMetadata.BitBucketProjectKey,
	}
	repoUrl, isNew, _, err := impl.gitOperationService.CreateRepository(context.Background(), gitRepoRequest, userId)
	if err != nil {
		impl.Logger.Errorw("error in creating git project", "name", gitOpsRepoName, "err", err)
		return "", false, err
	}
	return repoUrl, isNew, err
}

// createChartProxyAndGetPath parse chart in local directory and returns path of local dir and values.yaml
func (impl *FullModeDeploymentServiceImpl) createChartProxyAndGetPath(chartCreateRequest *util.ChartCreateRequest) (*util.ChartCreateResponse, error) {
	chartCreateResponse, err := impl.appStoreDeploymentCommonService.CreateChartProxyAndGetPath(chartCreateRequest)
	if err != nil {
		impl.Logger.Errorw("Error in building chart proxy", "err", err)
		return chartCreateResponse, err
	}
	return chartCreateResponse, nil
}

// getGitCommitConfig will return util.ChartConfig (git commit config) for GitOps
func (impl *FullModeDeploymentServiceImpl) getGitCommitConfig(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, fileString string, filename string) (*git.ChartConfig, error) {
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.Logger.Errorw("fetching error", "err", err)
		return nil, err
	}
	environment, err := impl.environmentRepository.FindById(installAppVersionRequest.EnvironmentId)
	if err != nil {
		impl.Logger.Errorw("fetching error", "err", err)
		return nil, err
	}

	argoCdAppName := globalUtil.BuildDeployedAppName(installAppVersionRequest.AppName, environment.Name)
	if util.IsAcdApp(installAppVersionRequest.DeploymentAppType) &&
		len(installAppVersionRequest.GitOpsRepoURL) == 0 &&
		installAppVersionRequest.InstalledAppId != 0 {
		InstalledApp, err := impl.installedAppRepository.GetInstalledApp(installAppVersionRequest.InstalledAppId)
		if err != nil && !util.IsErrNoRows(err) {
			impl.Logger.Errorw("service err, installedApp", "err", err)
			return nil, err
		}
		if util.IsErrNoRows(err) {
			apiErr := &util.ApiError{HttpStatusCode: http.StatusNotFound, Code: strconv.Itoa(http.StatusNotFound), InternalMessage: "Invalid request! No InstalledApp found.", UserMessage: "Invalid request! No InstalledApp found."}
			return nil, apiErr
		}
		deploymentConfig, err := impl.deploymentConfigService.GetConfigForHelmApps(InstalledApp.AppId, InstalledApp.EnvironmentId)
		if err != nil {
			impl.Logger.Errorw("error in getiting deployment config db object by appId and envId", "appId", InstalledApp.AppId, "envId", InstalledApp.EnvironmentId, "err", err)
			return nil, err
		}
		if util.IsErrNoRows(err) {
			apiErr := &util.ApiError{HttpStatusCode: http.StatusNotFound, Code: strconv.Itoa(http.StatusNotFound), InternalMessage: "Invalid request! No InstalledApp found.", UserMessage: "Invalid request! No InstalledApp found."}
			return nil, apiErr
		}
		//installAppVersionRequest.GitOpsRepoURL = InstalledApp.GitOpsRepoUrl
		installAppVersionRequest.GitOpsRepoURL = deploymentConfig.GetRepoURL()
		installAppVersionRequest.TargetRevision = deploymentConfig.GetTargetRevision()
	}
	gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(installAppVersionRequest.GitOpsRepoURL)
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(installAppVersionRequest.UserId)
	YamlConfig := &git.ChartConfig{
		FileName:       filename,
		FileContent:    fileString,
		ChartName:      installAppVersionRequest.AppName,
		ChartLocation:  argoCdAppName,
		ChartRepoName:  gitOpsRepoName,
		TargetRevision: installAppVersionRequest.GetTargetRevision(),
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", appStoreAppVersion.Id, environment.Id),
		UserEmailId:    userEmailId,
		UserName:       userName,
	}
	bitBucketBaseDir := fmt.Sprintf("%d-%s", appStoreAppVersion.Id, impl.chartTemplateService.GetDir())
	YamlConfig.SetBitBucketBaseDir(bitBucketBaseDir)
	return YamlConfig, nil
}

// getValuesAndRequirement will return chart values(*git.ChartConfig) and requirements(*git.ChartConfig) for git commit
func (impl *FullModeDeploymentServiceImpl) getValuesAndRequirement(installAppVersionRequest *appStoreBean.InstallAppVersionDTO,
	appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) (values map[string]map[string]any, dependencies []*chart.Dependency, err error) {
	if appStoreAppVersion == nil {
		appStoreAppVersion, err = impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
		if err != nil {
			impl.Logger.Errorw("fetching error", "err", err)
			return values, dependencies, err
		}
	}
	values, err = impl.appStoreDeploymentCommonService.GetProxyChartValues(appStoreAppVersion, installAppVersionRequest.ValuesOverrideYaml)
	if err != nil {
		impl.Logger.Errorw("error in getting values fot installedAppVersionRequest", "err", err)
		return values, dependencies, err
	}
	dependencies, err = impl.appStoreDeploymentCommonService.GetProxyChartRequirements(appStoreAppVersion)
	if err != nil {
		impl.Logger.Errorw("error in getting dependencies array fot installedAppVersionRequest", "err", err)
		return values, dependencies, err
	}
	return values, dependencies, err
}

// getValuesAndChartMetaDataForGitConfig will return chart values(*git.ChartConfig) and requirements(*git.ChartConfig) for git commit
func (impl *FullModeDeploymentServiceImpl) getValuesAndChartMetaDataForGitConfig(installAppVersionRequest *appStoreBean.InstallAppVersionDTO,
	values map[string]map[string]any, chartMetaData *chart.Metadata) (*git.ChartConfig, *git.ChartConfig, error) {
	if chartMetaData == nil || values == nil {
		err := fmt.Errorf("chartMetaData or values cannot be nil")
		impl.Logger.Errorw("error in getting chartMetaData or values for git commit", "err", err)
		return nil, nil, err
	}
	valuesContent, err := json.Marshal(values)
	if err != nil {
		impl.Logger.Errorw("error in marshalling values content", "err", err)
		return nil, nil, err
	}
	valuesConfig, err := impl.getGitCommitConfig(installAppVersionRequest, string(valuesContent), appStoreBean.VALUES_YAML_FILE)
	if err != nil {
		impl.Logger.Errorw("error in creating values config for git", "err", err)
		return nil, nil, err
	}
	chartMetaDataContent, err := json.Marshal(chartMetaData)
	if err != nil {
		impl.Logger.Errorw("error in marshalling chartMetaData content", "err", err)
		return nil, nil, err
	}
	chartMetaDataConfig, err := impl.getGitCommitConfig(installAppVersionRequest, string(chartMetaDataContent), chartRefBean.CHART_YAML_FILE)
	if err != nil {
		impl.Logger.Errorw("error in creating dependency config for git", "err", err)
		return nil, nil, err
	}
	return valuesConfig, chartMetaDataConfig, nil
}

func (impl *FullModeDeploymentServiceImpl) ValidateCustomGitOpsConfig(request validationBean.ValidateGitOpsRepoRequest) (string, bool, error) {
	request.TargetRevision = globalUtil.GetDefaultTargetRevision()
	return impl.gitOpsValidationService.ValidateCustomGitOpsConfig(request)
}

func (impl *FullModeDeploymentServiceImpl) CreateArgoRepoSecretIfNeeded(appStoreApplicationVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error {

	var err error

	if len(appStoreApplicationVersion.AppStore.DockerArtifactStoreId) == 0 {
		//only create repository secret for oci registry chart
		return nil
	}

	appStore := appStoreApplicationVersion.AppStore
	dockerArtifactStore := appStoreApplicationVersion.AppStore.DockerArtifactStore

	err = impl.argoClientWrapperService.AddOrUpdateOCIRegistry(
		dockerArtifactStore.Username,
		dockerArtifactStore.Password,
		dockerArtifactStore.OCIRegistryConfig[0].Id,
		dockerArtifactStore.RegistryURL,
		appStore.Name,
		dockerArtifactStore.OCIRegistryConfig[0].IsPublic,
	)
	if err != nil {
		impl.Logger.Errorw("error in creating repository secret", "dockerArtifactStoreId", dockerArtifactStore.Id, "err", err)
		return err
	}
	return nil
}

func (impl *FullModeDeploymentServiceImpl) shouldMigrateProxyChartDependencies(pushChartToGitRequest *gitBean.PushChartToGitRequestDTO) (bool, error) {
	clonedDir, err := impl.gitOperationService.CloneChartForHelmApp(pushChartToGitRequest.AppName, pushChartToGitRequest.RepoURL, pushChartToGitRequest.TargetRevision)
	if err != nil {
		impl.Logger.Errorw("error in cloning chart for helm app", "appName", pushChartToGitRequest.AppName, "repoUrl", pushChartToGitRequest.RepoURL, "err", err)
		return false, err
	}
	defer impl.chartTemplateService.CleanDir(clonedDir)
	gitOpsChartLocation := fmt.Sprintf("%s-%s", pushChartToGitRequest.AppName, pushChartToGitRequest.EnvName)
	dir := filepath.Join(clonedDir, gitOpsChartLocation)
	requirementsYamlPath := filepath.Join(dir, appStoreBean.REQUIREMENTS_YAML_FILE)
	if _, err := os.Stat(requirementsYamlPath); os.IsNotExist(err) {
		impl.Logger.Debugw("requirements.yaml not found in cloned repo from git-ops, no migrations required", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName)
		return false, nil
	} else if err != nil {
		impl.Logger.Errorw("error in checking requirements.yaml file", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "err", err)
		return false, err
	}
	chartYamlPath := filepath.Join(dir, chartRefBean.CHART_YAML_FILE)
	if _, err := os.Stat(chartYamlPath); os.IsNotExist(err) {
		impl.Logger.Debugw("chart.yaml not found in cloned repo from git-ops, no migrations required", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName)
		return false, nil
	} else if err != nil {
		impl.Logger.Errorw("error in checking chart.yaml file", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "err", err)
		return false, err
	}
	// load requirements.yaml file to check if it has dependencies
	requirementsYamlContent, err := os.ReadFile(requirementsYamlPath)
	if err != nil {
		impl.Logger.Errorw("error in reading requirements.yaml file", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "err", err)
		return false, err
	}
	dependencies := appStoreBean.Dependencies{}
	requirementsJsonContent, err := yaml.YAMLToJSON(requirementsYamlContent)
	if err != nil {
		impl.Logger.Errorw("error in converting requirements.yaml to json", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "err", err)
		return false, err
	}
	err = json.Unmarshal(requirementsJsonContent, &dependencies)
	if err != nil {
		impl.Logger.Errorw("error in unmarshalling requirements.yaml file", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "err", err)
		return false, err
	}
	if len(dependencies.Dependencies) == 0 {
		impl.Logger.Debugw("no dependencies found in requirements.yaml file", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName)
		return false, nil
	}
	impl.Logger.Debugw("dependencies found in requirements.yaml file", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "dependencies", dependencies.Dependencies)
	// check if chart.yaml file has dependencies
	chartYamlContent, err := os.ReadFile(chartYamlPath)
	if err != nil {
		impl.Logger.Errorw("error in reading chart.yaml file", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "err", err)
		return false, err
	}
	chartMetadata := &chart.Metadata{}
	chartJsonContent, err := yaml.YAMLToJSON(chartYamlContent)
	if err != nil {
		impl.Logger.Errorw("error in converting chart.yaml to json", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "err", err)
		return false, err
	}
	err = json.Unmarshal(chartJsonContent, chartMetadata)
	if err != nil {
		impl.Logger.Errorw("error in unmarshalling chart.yaml file", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "err", err)
		return false, err
	}
	if len(chartMetadata.Dependencies) == 0 {
		impl.Logger.Debugw("no dependencies found in chart.yaml file, need to migrate proxy chart dependencies", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName)
		return true, nil
	}
	impl.Logger.Debugw("dependencies found in chart.yaml file, validating against requirements.yaml", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "chartDependencies", chartMetadata.Dependencies)
	// validate if chart.yaml dependencies are present in requirements.yaml
	latestDependencies := sliceUtil.NewMapFromFuncExec(chartMetadata.Dependencies, func(dependency *chart.Dependency) string {
		return getUniqueKeyFromDependency(dependency)
	})
	previousDependencies := sliceUtil.NewMapFromFuncExec(dependencies.Dependencies, func(dependency *chart.Dependency) string {
		return getUniqueKeyFromDependency(dependency)
	})
	for key := range latestDependencies {
		if _, ok := previousDependencies[key]; !ok {
			impl.Logger.Debugw("dependency found in chart.yaml but not in requirements.yaml, need to migrate proxy chart dependencies", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName, "dependency", key)
			return true, nil
		}
	}
	impl.Logger.Debugw("all dependencies found in chart.yaml and requirements.yaml, no migration required", "appName", pushChartToGitRequest.AppName, "envName", pushChartToGitRequest.EnvName)
	return false, nil
}

func getUniqueKeyFromDependency(dependency *chart.Dependency) string {
	// return unique key for dependency
	return fmt.Sprintf("%s-%s-%s",
		strings.ToLower(strings.TrimSpace(dependency.Name)),
		strings.ToLower(strings.TrimSpace(dependency.Version)),
		strings.ToLower(strings.TrimSpace(dependency.Repository)))
}
