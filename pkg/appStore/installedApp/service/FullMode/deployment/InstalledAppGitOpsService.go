package deployment

import (
	"errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreDiscoverRepository "github.com/devtron-labs/devtron/pkg/appStore/discover/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/adapter"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/bean"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	"github.com/google/go-github/github"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/xanzy/go-gitlab"
	"net/http"
	"regexp"
)

type InstalledAppGitOpsService interface {
	// GitOpsOperations performs git operations specific to helm app deployments
	GitOpsOperations(manifestResponse *bean.AppStoreManifestResponse, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*bean.AppStoreGitOpsResponse, error)
	// GenerateManifest returns bean.AppStoreManifestResponse required in GitOps
	GenerateManifest(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (manifestResponse *bean.AppStoreManifestResponse, err error)
	// GenerateManifestAndPerformGitOperations is a wrapper function for both GenerateManifest and GitOpsOperations
	GenerateManifestAndPerformGitOperations(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*bean.AppStoreGitOpsResponse, error)
	// UpdateAppGitOpsOperations internally uses
	// GitOpsOperations (If Repo is deleted OR Repo migration is required) OR
	// git.GitOperationService.CommitValues (If repo exists and Repo migration is not needed)
	// functions to perform GitOps during upgrade deployments (GitOps based Helm Apps)
	UpdateAppGitOpsOperations(manifest *bean.AppStoreManifestResponse, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, monoRepoMigrationRequired *bool, commitRequirements bool) (*bean.AppStoreGitOpsResponse, error)
}

// GitOpsOperations handles all git operations for Helm App; and ensures that the return param bean.AppStoreGitOpsResponse is not nil
func (impl *FullModeDeploymentServiceImpl) GitOpsOperations(manifestResponse *bean.AppStoreManifestResponse, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*bean.AppStoreGitOpsResponse, error) {
	appStoreGitOpsResponse := &bean.AppStoreGitOpsResponse{}
	chartGitAttribute, isNew, githash, err := impl.createGitOpsRepoAndPushChart(installAppVersionRequest, manifestResponse.ChartResponse.BuiltChartPath, manifestResponse.RequirementsConfig, manifestResponse.ValuesConfig)
	if err != nil {
		impl.Logger.Errorw("Error in pushing chart to git", "err", err)
		return appStoreGitOpsResponse, err
	}
	space := regexp.MustCompile(`\s+`)
	appStoreName := space.ReplaceAllString(installAppVersionRequest.AppName, "-")

	// Checking this is the first time chart has been pushed , if yes requirements.yaml has been already pushed with chart as there was sync-delay with github api.
	// step-2 commit dependencies and values in git
	if !isNew {
		githash, err = impl.gitOperationService.CommitRequirementsAndValues(appStoreName, chartGitAttribute.RepoUrl, manifestResponse.RequirementsConfig, manifestResponse.ValuesConfig)
		if err != nil {
			impl.Logger.Errorw("error in committing config to git", "err", err)
			return appStoreGitOpsResponse, err
		}
	}
	appStoreGitOpsResponse.ChartGitAttribute = chartGitAttribute
	appStoreGitOpsResponse.GitHash = githash
	return appStoreGitOpsResponse, nil
}

func (impl *FullModeDeploymentServiceImpl) GenerateManifest(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (manifestResponse *bean.AppStoreManifestResponse, err error) {

	manifestResponse = &bean.AppStoreManifestResponse{}

	ChartCreateResponse, err := impl.createChartProxyAndGetPath(installAppVersionRequest)
	if err != nil {
		impl.Logger.Errorw("Error in building chart while generating manifest", "err", err)
		return manifestResponse, err
	}
	valuesConfig, dependencyConfig, err := impl.getValuesAndRequirementForGitConfig(installAppVersionRequest)
	if err != nil {
		impl.Logger.Errorw("error in fetching values and requirements.yaml config while generating manifest", "err", err)
		return manifestResponse, err
	}

	manifestResponse.ChartResponse = ChartCreateResponse
	manifestResponse.ValuesConfig = valuesConfig
	manifestResponse.RequirementsConfig = dependencyConfig

	return manifestResponse, nil
}

func (impl *FullModeDeploymentServiceImpl) GenerateManifestAndPerformGitOperations(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*bean.AppStoreGitOpsResponse, error) {
	appStoreGitOpsResponse := &bean.AppStoreGitOpsResponse{}
	manifest, err := impl.GenerateManifest(installAppVersionRequest)
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

func (impl *FullModeDeploymentServiceImpl) UpdateAppGitOpsOperations(manifest *bean.AppStoreManifestResponse, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, monoRepoMigrationRequired *bool, commitRequirements bool) (*bean.AppStoreGitOpsResponse, error) {
	var requirementsCommitErr, valuesCommitErr error
	var gitHash string
	if *monoRepoMigrationRequired {
		return impl.GitOpsOperations(manifest, installAppVersionRequest)
	}

	gitOpsResponse := &bean.AppStoreGitOpsResponse{}
	if commitRequirements {
		// update dependency if chart or chart version is changed
		_, _, requirementsCommitErr = impl.gitOperationService.CommitValues(manifest.RequirementsConfig)
		gitHash, _, valuesCommitErr = impl.gitOperationService.CommitValues(manifest.ValuesConfig)
	} else {
		// only values are changed in update, so commit values config
		gitHash, _, valuesCommitErr = impl.gitOperationService.CommitValues(manifest.ValuesConfig)
	}

	if valuesCommitErr != nil || requirementsCommitErr != nil {
		noTargetFoundForValues, _ := impl.parseGitRepoErrorResponse(valuesCommitErr)
		noTargetFoundForRequirements, _ := impl.parseGitRepoErrorResponse(requirementsCommitErr)
		if noTargetFoundForRequirements || noTargetFoundForValues {
			//create repo again and try again  -  auto fix
			*monoRepoMigrationRequired = true // since repo is created again, will use this flag to check if ACD patch operation required
			return impl.GitOpsOperations(manifest, installAppVersionRequest)
		}
		impl.Logger.Errorw("error in performing GitOps for upgrade deployment", "ValuesCommitErr", valuesCommitErr, "RequirementsCommitErr", requirementsCommitErr)
		return nil, fmt.Errorf("error in committing values and requirements to git repository")
	}
	gitOpsResponse.GitHash = gitHash
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
		if err.Error() == git.BITBUCKET_REPO_NOT_FOUND_ERROR {
			impl.Logger.Errorw("no content found while updating git repo bitbucket, do auto fix", "error", err)
			noTargetFound = true
		}
	}
	return noTargetFound, err
}

// createGitOpsRepoAndPushChart is a wrapper for creating GitOps repo and pushing chart to created repo
func (impl *FullModeDeploymentServiceImpl) createGitOpsRepoAndPushChart(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, builtChartPath string, requirementsConfig *git.ChartConfig, valuesConfig *git.ChartConfig) (*commonBean.ChartGitAttribute, bool, string, error) {
	repoURL, isNew, err := impl.createGitOpsRepo(installAppVersionRequest)
	if err != nil {
		impl.Logger.Errorw("Error in creating gitops repo for ", "appName", installAppVersionRequest.AppName, "err", err)
		return nil, false, "", err
	}
	pushChartToGitRequest := adapter.ParseChartGitPushRequest(installAppVersionRequest, repoURL, builtChartPath)
	chartGitAttribute, commitHash, err := impl.gitOperationService.PushChartToGitOpsRepoForHelmApp(pushChartToGitRequest, requirementsConfig, valuesConfig)
	if err != nil {
		impl.Logger.Errorw("error in pushing chart to git", "err", err)
		return nil, false, "", err
	}
	return chartGitAttribute, isNew, commitHash, err
}

// createGitOpsRepo creates a gitOps repo with readme
func (impl *FullModeDeploymentServiceImpl) createGitOpsRepo(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (string, bool, error) {
	if len(installAppVersionRequest.GitOpsRepoName) == 0 {
		//here gitops repo will be the app name, to breaking the mono repo structure
		gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoName(installAppVersionRequest.AppName)
		installAppVersionRequest.GitOpsRepoName = gitOpsRepoName
	}
	bitbucketMetadata, err := impl.gitOpsConfigReadService.GetBitbucketMetadata()
	if err != nil {
		impl.Logger.Errorw("error in getting bitbucket metadata", "err", err)
		return "", false, err
	}
	//getting user name & emailId for commit author data
	gitRepoRequest := &bean2.GitOpsConfigDto{
		GitRepoName:          installAppVersionRequest.GitOpsRepoName,
		Description:          "helm chart for " + installAppVersionRequest.GitOpsRepoName,
		BitBucketWorkspaceId: bitbucketMetadata.BitBucketWorkspaceId,
		BitBucketProjectKey:  bitbucketMetadata.BitBucketProjectKey,
	}
	repoUrl, isNew, err := impl.gitOperationService.CreateRepository(gitRepoRequest, installAppVersionRequest.UserId)
	if err != nil {
		impl.Logger.Errorw("error in creating git project", "name", installAppVersionRequest.GitOpsRepoName, "err", err)
		return "", false, err
	}
	return repoUrl, isNew, err
}

func (impl *FullModeDeploymentServiceImpl) updateValuesYamlInGit(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*appStoreBean.InstallAppVersionDTO, error) {
	valuesString, err := impl.appStoreDeploymentCommonService.GetValuesString(installAppVersionRequest.AppStoreName, installAppVersionRequest.ValuesOverrideYaml)
	if err != nil {
		impl.Logger.Errorw("error in getting values string", "err", err)
		return nil, err
	}

	valuesGitConfig, err := impl.getGitCommitConfig(installAppVersionRequest, valuesString, appStoreBean.VALUES_YAML_FILE)
	if err != nil {
		impl.Logger.Errorw("error in getting git commit config", "err", err)
	}

	commitHash, _, err := impl.gitOperationService.CommitValues(valuesGitConfig)
	if err != nil {
		impl.Logger.Errorw("error in git commit", "err", err)
		return installAppVersionRequest, errors.New(pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED)
	}
	//update timeline status for git commit state
	installAppVersionRequest.GitHash = commitHash
	return installAppVersionRequest, nil
}

func (impl *FullModeDeploymentServiceImpl) updateRequirementYamlInGit(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, appStoreAppVersion *appStoreDiscoverRepository.AppStoreApplicationVersion) error {
	requirementsString, err := impl.appStoreDeploymentCommonService.GetRequirementsString(appStoreAppVersion.Id)
	if err != nil {
		impl.Logger.Errorw("error in getting requirements string", "err", err)
		return err
	}

	requirementsGitConfig, err := impl.getGitCommitConfig(installAppVersionRequest, requirementsString, appStoreBean.REQUIREMENTS_YAML_FILE)
	if err != nil {
		impl.Logger.Errorw("error in getting git commit config", "err", err)
		return err
	}

	_, _, err = impl.gitOperationService.CommitValues(requirementsGitConfig)
	if err != nil {
		impl.Logger.Errorw("error in values commit", "err", err)
		return errors.New(pipelineConfig.TIMELINE_STATUS_GIT_COMMIT_FAILED)
	}

	return nil
}

// createChartProxyAndGetPath parse chart in local directory and returns path of local dir and values.yaml
func (impl *FullModeDeploymentServiceImpl) createChartProxyAndGetPath(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*util.ChartCreateResponse, error) {
	chartCreateRequest := adapter.ParseChartCreateRequest(installAppVersionRequest.AppName)
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

	argocdAppName := installAppVersionRequest.AppName + "-" + environment.Name
	gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoName(installAppVersionRequest.AppName)
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(installAppVersionRequest.UserId)
	YamlConfig := &git.ChartConfig{
		FileName:       filename,
		FileContent:    fileString,
		ChartName:      installAppVersionRequest.AppName,
		ChartLocation:  argocdAppName,
		ChartRepoName:  gitOpsRepoName,
		ReleaseMessage: fmt.Sprintf("release-%d-env-%d ", appStoreAppVersion.Id, environment.Id),
		UserEmailId:    userEmailId,
		UserName:       userName,
	}
	return YamlConfig, nil
}

// getValuesAndRequirementForGitConfig will return chart values(*util.ChartConfig) and requirements(*util.ChartConfig) for git commit
func (impl *FullModeDeploymentServiceImpl) getValuesAndRequirementForGitConfig(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*git.ChartConfig, *git.ChartConfig, error) {
	appStoreAppVersion, err := impl.appStoreApplicationVersionRepository.FindById(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.Logger.Errorw("fetching error", "err", err)
		return nil, nil, err
	}
	values, err := impl.appStoreDeploymentCommonService.GetValuesString(appStoreAppVersion.AppStore.Name, installAppVersionRequest.ValuesOverrideYaml)
	if err != nil {
		impl.Logger.Errorw("error in getting values fot installedAppVersionRequest", "err", err)
		return nil, nil, err
	}
	dependency, err := impl.appStoreDeploymentCommonService.GetRequirementsString(installAppVersionRequest.AppStoreVersion)
	if err != nil {
		impl.Logger.Errorw("error in getting dependency array fot installedAppVersionRequest", "err", err)
		return nil, nil, err
	}
	valuesConfig, err := impl.getGitCommitConfig(installAppVersionRequest, values, appStoreBean.VALUES_YAML_FILE)
	if err != nil {
		impl.Logger.Errorw("error in creating values config for git", "err", err)
		return nil, nil, err
	}
	RequirementConfig, err := impl.getGitCommitConfig(installAppVersionRequest, dependency, appStoreBean.REQUIREMENTS_YAML_FILE)
	if err != nil {
		impl.Logger.Errorw("error in creating dependency config for git", "err", err)
		return nil, nil, err
	}
	return valuesConfig, RequirementConfig, nil
}
