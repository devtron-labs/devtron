package appStoreDeploymentCommon

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/bean"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/google/go-github/github"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	dirCopy "github.com/otiai10/copy"
	"github.com/xanzy/go-gitlab"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

func (impl AppStoreDeploymentCommonServiceImpl) ParseGitRepoErrorResponse(err error) (bool, error) {
	//update values yaml in chart
	noTargetFound := false
	if err != nil {
		if errorResponse, ok := err.(*github.ErrorResponse); ok && errorResponse.Response.StatusCode == http.StatusNotFound {
			impl.logger.Errorw("no content found while updating git repo on github, do auto fix", "error", err)
			noTargetFound = true
		}
		if errorResponse, ok := err.(azuredevops.WrappedError); ok && *errorResponse.StatusCode == http.StatusNotFound {
			impl.logger.Errorw("no content found while updating git repo on azure, do auto fix", "error", err)
			noTargetFound = true
		}
		if errorResponse, ok := err.(*azuredevops.WrappedError); ok && *errorResponse.StatusCode == http.StatusNotFound {
			impl.logger.Errorw("no content found while updating git repo on azure, do auto fix", "error", err)
			noTargetFound = true
		}
		if errorResponse, ok := err.(*gitlab.ErrorResponse); ok && errorResponse.Response.StatusCode == http.StatusNotFound {
			impl.logger.Errorw("no content found while updating git repo gitlab, do auto fix", "error", err)
			noTargetFound = true
		}
		if err.Error() == util.BITBUCKET_REPO_NOT_FOUND_ERROR {
			impl.logger.Errorw("no content found while updating git repo bitbucket, do auto fix", "error", err)
			noTargetFound = true
		}
	}
	return noTargetFound, err
}

// CreateGitOpsRepoAndPushChart is a wrapper for creating GitOps repo and pushing chart to created repo
func (impl AppStoreDeploymentCommonServiceImpl) CreateGitOpsRepoAndPushChart(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, builtChartPath string, requirementsConfig *util.ChartConfig, valuesConfig *util.ChartConfig) (*commonBean.ChartGitAttribute, bool, string, error) {
	repoURL, isNew, err := impl.CreateGitOpsRepo(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("Error in creating gitops repo for ", "appName", installAppVersionRequest.AppName, "err", err)
		return nil, false, "", err
	}
	pushChartToGitRequest := ParseChartGitPushRequest(installAppVersionRequest, repoURL, builtChartPath)
	chartGitAttribute, commitHash, err := impl.pushChartToGitOpsRepo(pushChartToGitRequest, requirementsConfig, valuesConfig)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git", "err", err)
		return nil, false, "", err
	}
	return chartGitAttribute, isNew, commitHash, err
}

func (impl AppStoreDeploymentCommonServiceImpl) GitOpsOperations(manifestResponse *AppStoreManifestResponse, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*AppStoreGitOpsResponse, error) {
	appStoreGitOpsResponse := &AppStoreGitOpsResponse{}
	chartGitAttribute, isNew, githash, err := impl.CreateGitOpsRepoAndPushChart(installAppVersionRequest, manifestResponse.ChartResponse.BuiltChartPath, manifestResponse.RequirementsConfig, manifestResponse.ValuesConfig)
	if err != nil {
		impl.logger.Errorw("Error in pushing chart to git", "err", err)
		return appStoreGitOpsResponse, err
	}
	space := regexp.MustCompile(`\s+`)
	appStoreName := space.ReplaceAllString(installAppVersionRequest.AppName, "-")
	clonedDir := impl.gitFactory.GitWorkingDir + "" + appStoreName

	// Checking this is the first time chart has been pushed , if yes requirements.yaml has been already pushed with chart as there was sync-delay with github api.
	// step-2 commit dependencies and values in git
	if !isNew {
		_, _, err = impl.gitOpsRemoteOperationService.CommitValues(manifestResponse.RequirementsConfig)
		if err != nil {
			impl.logger.Errorw("error in committing dependency config to git", "err", err)
			return appStoreGitOpsResponse, err
		}
		err = impl.gitOpsRemoteOperationService.GitPull(clonedDir, chartGitAttribute.RepoUrl, appStoreName)
		if err != nil {
			impl.logger.Errorw("error in git pull", "err", err)
			return appStoreGitOpsResponse, err
		}

		githash, _, err = impl.gitOpsRemoteOperationService.CommitValues(manifestResponse.ValuesConfig)
		if err != nil {
			impl.logger.Errorw("error in committing values config to git", "err", err)
			return appStoreGitOpsResponse, err
		}
		err = impl.gitOpsRemoteOperationService.GitPull(clonedDir, chartGitAttribute.RepoUrl, appStoreName)
		if err != nil {
			impl.logger.Errorw("error in git pull", "err", err)
			return appStoreGitOpsResponse, err
		}
	}
	appStoreGitOpsResponse.ChartGitAttribute = chartGitAttribute
	appStoreGitOpsResponse.GitHash = githash
	return appStoreGitOpsResponse, nil
}

func (impl AppStoreDeploymentCommonServiceImpl) GenerateManifestAndPerformGitOperations(installAppVersionRequest *appStoreBean.InstallAppVersionDTO) (*AppStoreGitOpsResponse, error) {
	appStoreGitOpsResponse := &AppStoreGitOpsResponse{}
	manifest, err := impl.GenerateManifest(installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("error in performing manifest and git operations", "err", err)
		return nil, err
	}
	gitOpsResponse, err := impl.GitOpsOperations(manifest, installAppVersionRequest)
	if err != nil {
		impl.logger.Errorw("error in performing gitops operation", "err", err)
		return nil, err
	}
	installAppVersionRequest.GitHash = gitOpsResponse.GitHash
	appStoreGitOpsResponse.ChartGitAttribute = gitOpsResponse.ChartGitAttribute
	return appStoreGitOpsResponse, nil
}

// pushChartToGitOpsRepo pushes built chart to GitOps repo
func (impl AppStoreDeploymentCommonServiceImpl) pushChartToGitOpsRepo(PushChartToGitRequest *appStoreBean.PushChartToGitRequestDTO, requirementsConfig *util.ChartConfig, valuesConfig *util.ChartConfig) (*commonBean.ChartGitAttribute, string, error) {
	space := regexp.MustCompile(`\s+`)
	appStoreName := space.ReplaceAllString(PushChartToGitRequest.ChartAppStoreName, "-")
	chartDir := fmt.Sprintf("%s-%s", PushChartToGitRequest.AppName, impl.chartTemplateService.GetDir())
	clonedDir := impl.gitFactory.GitService.GetCloneDirectory(chartDir)
	if _, err := os.Stat(clonedDir); os.IsNotExist(err) {
		clonedDir, err = impl.gitFactory.GitService.Clone(PushChartToGitRequest.RepoURL, chartDir)
		if err != nil {
			impl.logger.Errorw("error in cloning repo", "url", PushChartToGitRequest.RepoURL, "err", err)
			return nil, "", err
		}
	} else {
		err = impl.gitOpsRemoteOperationService.GitPull(clonedDir, PushChartToGitRequest.RepoURL, appStoreName)
		if err != nil {
			return nil, "", err
		}
	}
	acdAppName := fmt.Sprintf("%s-%s", PushChartToGitRequest.AppName, PushChartToGitRequest.EnvName)
	dir := filepath.Join(clonedDir, acdAppName)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		impl.logger.Errorw("error in making dir", "err", err)
		return nil, "", err
	}
	err = dirCopy.Copy(PushChartToGitRequest.TempChartRefDir, dir)
	if err != nil {
		impl.logger.Errorw("error copying dir", "err", err)
		return nil, "", err
	}
	err = impl.AddConfigFileToChart(requirementsConfig, dir, clonedDir)
	if err != nil {
		impl.logger.Errorw("error in adding requirements.yaml to chart", "err", err, "appName", PushChartToGitRequest.AppName)
		return nil, "", err
	}
	err = impl.AddConfigFileToChart(valuesConfig, dir, clonedDir)
	if err != nil {
		impl.logger.Errorw("error in adding values.yaml to chart", "err", err, "appName", PushChartToGitRequest.AppName)
		return nil, "", err
	}
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(PushChartToGitRequest.UserId)
	commit, err := impl.gitFactory.GitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
	if err != nil {
		impl.logger.Errorw("error in pushing git", "err", err)
		impl.logger.Warn("re-trying, taking pull and then push again")
		err = impl.gitOpsRemoteOperationService.GitPull(clonedDir, PushChartToGitRequest.RepoURL, acdAppName)
		if err != nil {
			impl.logger.Errorw("error in git pull", "err", err, "appName", acdAppName)
			return nil, "", err
		}
		err = dirCopy.Copy(PushChartToGitRequest.TempChartRefDir, dir)
		if err != nil {
			impl.logger.Errorw("error copying dir", "err", err)
			return nil, "", err
		}
		commit, err = impl.gitFactory.GitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
		if err != nil {
			impl.logger.Errorw("error in pushing git", "err", err)
			return nil, "", err
		}
	}
	impl.logger.Debugw("template committed", "url", PushChartToGitRequest.RepoURL, "commit", commit)
	defer impl.chartTemplateService.CleanDir(clonedDir)
	return &commonBean.ChartGitAttribute{RepoUrl: PushChartToGitRequest.RepoURL, ChartLocation: acdAppName}, commit, err
}
