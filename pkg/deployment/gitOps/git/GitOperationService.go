package git

import (
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/util"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
	dirCopy "github.com/otiai10/copy"
	"go.uber.org/zap"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"os"
	"path/filepath"
	"regexp"
	"sigs.k8s.io/yaml"
	"time"
)

type GitOperationService interface {
	CreateGitRepositoryForApp(gitOpsRepoName, baseTemplateName,
		version string, userId int32) (chartGitAttribute *commonBean.ChartGitAttribute, err error)
	PushChartToGitRepo(gitOpsRepoName, referenceTemplate, version,
		tempReferenceTemplateDir string, repoUrl string, userId int32) (err error)
	CreateReadmeInGitRepo(gitOpsRepoName string, userId int32) error
	CreateChartProxy(chartMetaData *chart.Metadata, refChartLocation string, envName string,
		chartProxyReq *bean.ChartProxyReqDto) (string, *commonBean.ChartGitAttribute, error)
	GitPull(clonedDir string, repoUrl string, appStoreName string) error

	CommitValues(chartGitAttr *ChartConfig) (commitHash string, commitTime time.Time, err error)
	CommitAndPushAllChanges(clonedDir, commitMsg, userName, userEmailId string) (commitHash string, err error)

	CreateRepository(dto *bean2.GitOpsConfigDto, userId int32) (string, bool, error)
	GetRepoUrlByRepoName(repoName string) (string, error)

	GetClonedDir(chartDir, repoUrl string) (string, error)
	CloneInDir(repoUrl, chartDir string) (string, error)
}

type GitOperationServiceImpl struct {
	logger                  *zap.SugaredLogger
	gitFactory              *GitFactory
	gitOpsConfigReadService config.GitOpsConfigReadService
	chartTemplateService    util.ChartTemplateService
}

func NewGitOperationServiceImpl(logger *zap.SugaredLogger, gitFactory *GitFactory,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	chartTemplateService util.ChartTemplateService) *GitOperationServiceImpl {
	return &GitOperationServiceImpl{
		logger:                  logger,
		gitFactory:              gitFactory,
		gitOpsConfigReadService: gitOpsConfigReadService,
		chartTemplateService:    chartTemplateService,
	}

}

func (impl *GitOperationServiceImpl) CreateGitRepositoryForApp(gitOpsRepoName, baseTemplateName,
	version string, userId int32) (chartGitAttribute *commonBean.ChartGitAttribute, err error) {
	//baseTemplateName  replace whitespace
	space := regexp.MustCompile(`\s+`)
	gitOpsRepoName = space.ReplaceAllString(gitOpsRepoName, "-")

	bitbucketMetadata, err := impl.gitOpsConfigReadService.GetBitbucketMetadata()
	if err != nil {
		impl.logger.Errorw("error in getting bitbucket metadata", "err", err)
		return nil, err
	}
	//getting user name & emailId for commit author data
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(userId)
	gitRepoRequest := &bean2.GitOpsConfigDto{
		GitRepoName:          gitOpsRepoName,
		Description:          fmt.Sprintf("helm chart for " + gitOpsRepoName),
		Username:             userName,
		UserEmailId:          userEmailId,
		BitBucketWorkspaceId: bitbucketMetadata.BitBucketWorkspaceId,
		BitBucketProjectKey:  bitbucketMetadata.BitBucketProjectKey,
	}
	repoUrl, _, detailedError := impl.gitFactory.Client.CreateRepository(gitRepoRequest)
	for _, err := range detailedError.StageErrorMap {
		if err != nil {
			impl.logger.Errorw("error in creating git project", "name", gitOpsRepoName, "err", err)
			return nil, err
		}
	}
	return &commonBean.ChartGitAttribute{RepoUrl: repoUrl, ChartLocation: filepath.Join(baseTemplateName, version)}, nil
}

func (impl *GitOperationServiceImpl) PushChartToGitRepo(gitOpsRepoName, referenceTemplate, version,
	tempReferenceTemplateDir string, repoUrl string, userId int32) (err error) {
	chartDir := fmt.Sprintf("%s-%s", gitOpsRepoName, impl.chartTemplateService.GetDir())
	clonedDir, err := impl.GetClonedDir(chartDir, repoUrl)
	defer impl.chartTemplateService.CleanDir(clonedDir)
	if err != nil {
		impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
		return err
	}
	err = impl.GitPull(clonedDir, repoUrl, gitOpsRepoName)
	if err != nil {
		impl.logger.Errorw("error in pulling git repo", "url", repoUrl, "err", err)
		return err
	}
	dir := filepath.Join(clonedDir, referenceTemplate, version)
	performFirstCommitPush := true

	//if chart already exists don't overrides it by reference template
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			impl.logger.Errorw("error in making dir", "err", err)
			return err
		}
		err = dirCopy.Copy(tempReferenceTemplateDir, dir)
		if err != nil {
			impl.logger.Errorw("error copying dir", "err", err)
			return err
		}
	} else {
		// auto-healing : data corruption fix - sometimes reference chart contents are not pushed in git-ops repo.
		// copying content from reference template dir to cloned dir (if Chart.yaml file is not found)
		// if Chart.yaml file is not found, we are assuming here that reference chart contents are not pushed in git-ops repo
		if _, err := os.Stat(filepath.Join(dir, "Chart.yaml")); os.IsNotExist(err) {
			impl.logger.Infow("auto-healing: Chart.yaml not found in cloned repo from git-ops. copying content", "from", tempReferenceTemplateDir, "to", dir)
			err = dirCopy.Copy(tempReferenceTemplateDir, dir)
			if err != nil {
				impl.logger.Errorw("error copying content in auto-healing", "err", err)
				return err
			}
		} else {
			// chart exists on git, hence not performing first commit
			performFirstCommitPush = false
		}
	}

	// if push needed, then only push
	if performFirstCommitPush {
		userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(userId)
		commit, err := impl.gitFactory.GitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
		if err != nil {
			impl.logger.Errorw("error in pushing git", "err", err)
			impl.logger.Warn("re-trying, taking pull and then push again")
			err = impl.GitPull(clonedDir, repoUrl, gitOpsRepoName)
			if err != nil {
				return err
			}
			err = dirCopy.Copy(tempReferenceTemplateDir, dir)
			if err != nil {
				impl.logger.Errorw("error copying dir", "err", err)
				return err
			}
			commit, err = impl.gitFactory.GitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
			if err != nil {
				impl.logger.Errorw("error in pushing git", "err", err)
				return err
			}
		}
		impl.logger.Debugw("template committed", "url", repoUrl, "commit", commit)
	}

	return nil
}

func (impl *GitOperationServiceImpl) CreateReadmeInGitRepo(gitOpsRepoName string, userId int32) error {
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(userId)
	gitOpsConfig, err := impl.gitOpsConfigReadService.GetGitOpsConfigActive()
	if err != nil {
		impl.logger.Errorw("error in getting active gitOps config", "err", err)
		return err
	}
	//updating user email and name in request
	if gitOpsConfig != nil {
		gitOpsConfig.UserEmailId = userEmailId
		gitOpsConfig.Username = userName
		gitOpsConfig.GitRepoName = gitOpsRepoName
	}
	_, err = impl.gitFactory.Client.CreateReadme(gitOpsConfig)
	if err != nil {
		impl.logger.Errorw("error in creating readme", "err", err, "gitOpsRepoName", gitOpsRepoName, "userId", userId)
		return err
	}
	return nil
}

func (impl *GitOperationServiceImpl) CreateChartProxy(chartMetaData *chart.Metadata, refChartLocation string, envName string,
	chartProxyReq *bean.ChartProxyReqDto) (string, *commonBean.ChartGitAttribute, error) {
	chartMetaData.ApiVersion = "v2" // ensure always v2
	dir := impl.chartTemplateService.GetDir()
	chartDir := filepath.Join(util.CHART_WORKING_DIR_PATH, dir)
	impl.logger.Debugw("chart dir ", "chart", chartMetaData.Name, "dir", chartDir)
	err := os.MkdirAll(chartDir, os.ModePerm) //hack for concurrency handling
	if err != nil {
		impl.logger.Errorw("err in creating dir", "dir", chartDir, "err", err)
		return "", nil, err
	}
	defer impl.chartTemplateService.CleanDir(chartDir)
	err = dirCopy.Copy(refChartLocation, chartDir)

	if err != nil {
		impl.logger.Errorw("error in copying chart for app", "app", chartMetaData.Name, "error", err)
		return "", nil, err
	}
	archivePath, valuesYaml, err := impl.chartTemplateService.PackageChart(chartDir, chartMetaData)
	if err != nil {
		impl.logger.Errorw("error in creating archive", "err", err)
		return "", nil, err
	}

	chartGitAttr, err := impl.createAndPushToGitChartProxy(chartMetaData.Name, chartDir, envName, chartProxyReq)
	if err != nil {
		impl.logger.Errorw("error in pushing chart to git ", "path", archivePath, "err", err)
		return "", nil, err
	}
	if valuesYaml == "" {
		valuesYaml = "{}"
	} else {
		valuesYamlByte, err := yaml.YAMLToJSON([]byte(valuesYaml))
		if err != nil {
			return "", nil, err
		}
		valuesYaml = string(valuesYamlByte)
	}
	return valuesYaml, chartGitAttr, nil
}

func (impl *GitOperationServiceImpl) createAndPushToGitChartProxy(appStoreName, tmpChartLocation string, envName string,
	chartProxyReq *bean.ChartProxyReqDto) (chartGitAttribute *commonBean.ChartGitAttribute, err error) {
	//baseTemplateName  replace whitespace
	space := regexp.MustCompile(`\s+`)
	appStoreName = space.ReplaceAllString(appStoreName, "-")

	if len(chartProxyReq.GitOpsRepoName) == 0 {
		//here git ops repo will be the app name, to breaking the mono repo structure
		gitOpsRepoName := impl.gitOpsConfigReadService.GetGitOpsRepoName(chartProxyReq.AppName)
		chartProxyReq.GitOpsRepoName = gitOpsRepoName
	}
	bitbucketMetadata, err := impl.gitOpsConfigReadService.GetBitbucketMetadata()
	if err != nil {
		impl.logger.Errorw("error in getting bitbucket metadata", "err", err)
		return nil, err
	}
	//getting user name & emailId for commit author data
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(chartProxyReq.UserId)
	gitRepoRequest := &bean2.GitOpsConfigDto{
		GitRepoName:          chartProxyReq.GitOpsRepoName,
		Description:          "helm chart for " + chartProxyReq.GitOpsRepoName,
		Username:             userName,
		UserEmailId:          userEmailId,
		BitBucketWorkspaceId: bitbucketMetadata.BitBucketWorkspaceId,
		BitBucketProjectKey:  bitbucketMetadata.BitBucketProjectKey,
	}
	repoUrl, _, detailedError := impl.gitFactory.Client.CreateRepository(gitRepoRequest)
	for _, err := range detailedError.StageErrorMap {
		if err != nil {
			impl.logger.Errorw("error in creating git project", "name", chartProxyReq.GitOpsRepoName, "err", err)
			return nil, err
		}
	}

	chartDir := fmt.Sprintf("%s-%s", chartProxyReq.AppName, impl.chartTemplateService.GetDir())
	clonedDir, err := impl.GetClonedDir(chartDir, repoUrl)
	if err != nil {
		impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
		return nil, err
	}

	err = impl.GitPull(clonedDir, repoUrl, appStoreName)
	if err != nil {
		return nil, err
	}

	acdAppName := fmt.Sprintf("%s-%s", chartProxyReq.AppName, envName)
	dir := filepath.Join(clonedDir, acdAppName)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		impl.logger.Errorw("error in making dir", "err", err)
		return nil, err
	}
	err = dirCopy.Copy(tmpChartLocation, dir)
	if err != nil {
		impl.logger.Errorw("error copying dir", "err", err)
		return nil, err
	}
	commit, err := impl.gitFactory.GitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
	if err != nil {
		impl.logger.Errorw("error in pushing git", "err", err)
		impl.logger.Warn("re-trying, taking pull and then push again")
		err = impl.GitPull(clonedDir, repoUrl, acdAppName)
		if err != nil {
			return nil, err
		}
		err = dirCopy.Copy(tmpChartLocation, dir)
		if err != nil {
			impl.logger.Errorw("error copying dir", "err", err)
			return nil, err
		}
		commit, err = impl.gitFactory.GitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
		if err != nil {
			impl.logger.Errorw("error in pushing git", "err", err)
			return nil, err
		}
	}
	impl.logger.Debugw("template committed", "url", repoUrl, "commit", commit)
	defer impl.chartTemplateService.CleanDir(clonedDir)
	return &commonBean.ChartGitAttribute{RepoUrl: repoUrl, ChartLocation: filepath.Join("", acdAppName)}, nil
}

func (impl *GitOperationServiceImpl) GitPull(clonedDir string, repoUrl string, appStoreName string) error {
	err := impl.gitFactory.GitService.Pull(clonedDir) //TODO check for local repo exists before clone
	if err != nil {
		impl.logger.Errorw("error in pulling git", "clonedDir", clonedDir, "err", err)
		_, err := impl.gitFactory.GitService.Clone(repoUrl, appStoreName)
		if err != nil {
			impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
			return err
		}
		return nil
	}
	return nil
}

func (impl *GitOperationServiceImpl) CommitValues(chartGitAttr *ChartConfig) (commitHash string, commitTime time.Time, err error) {
	bitbucketMetadata, err := impl.gitOpsConfigReadService.GetBitbucketMetadata()
	if err != nil {
		impl.logger.Errorw("error in getting bitbucket metadata", "err", err)
		return commitHash, commitTime, err
	}
	gitOpsConfig := &bean2.GitOpsConfigDto{BitBucketWorkspaceId: bitbucketMetadata.BitBucketWorkspaceId}
	commitHash, commitTime, err = impl.gitFactory.Client.CommitValues(chartGitAttr, gitOpsConfig)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return commitHash, commitTime, err
	}
	return commitHash, commitTime, nil
}

func (impl *GitOperationServiceImpl) CommitAndPushAllChanges(clonedDir, commitMsg, userName, userEmailId string) (commitHash string, err error) {
	commitHash, err = impl.gitFactory.GitService.CommitAndPushAllChanges(clonedDir, commitMsg, userName, userEmailId)
	if err != nil {
		impl.logger.Errorw("error in pushing git", "err", err)
		return commitHash, err
	}
	return commitHash, nil
}

func (impl *GitOperationServiceImpl) CreateRepository(dto *bean2.GitOpsConfigDto, userId int32) (string, bool, error) {
	//getting user name & emailId for commit author data
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(userId)
	if dto != nil {
		dto.UserEmailId = userEmailId
		dto.Username = userName
	}
	repoUrl, isNew, detailedError := impl.gitFactory.Client.CreateRepository(dto)
	for _, err := range detailedError.StageErrorMap {
		if err != nil {
			impl.logger.Errorw("error in creating git project", "err", err, "req", dto)
			return "", false, err
		}
	}
	return repoUrl, isNew, nil
}

func (impl *GitOperationServiceImpl) GetRepoUrlByRepoName(repoName string) (string, error) {
	repoUrl := ""
	bitbucketMetadata, err := impl.gitOpsConfigReadService.GetBitbucketMetadata()
	if err != nil {
		impl.logger.Errorw("error in getting bitbucket metadata", "err", err)
		return repoUrl, err
	}
	dto := &bean2.GitOpsConfigDto{
		GitRepoName:          repoName,
		BitBucketWorkspaceId: bitbucketMetadata.BitBucketWorkspaceId,
		BitBucketProjectKey:  bitbucketMetadata.BitBucketProjectKey,
	}
	repoUrl, err = impl.gitFactory.Client.GetRepoUrl(dto)
	if err != nil {
		//will allow to continue to persist status on next operation
		impl.logger.Errorw("error in getting repo url", "err", err, "repoName", repoName)
	}
	return repoUrl, nil
}

func (impl *GitOperationServiceImpl) GetClonedDir(chartDir, repoUrl string) (string, error) {
	clonedDir := impl.gitFactory.GitService.GetCloneDirectory(chartDir)
	if _, err := os.Stat(clonedDir); os.IsNotExist(err) {
		return impl.CloneInDir(repoUrl, chartDir)
	} else if err != nil {
		impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
		return "", err
	}
	return clonedDir, nil
}

func (impl *GitOperationServiceImpl) CloneInDir(repoUrl, chartDir string) (string, error) {
	clonedDir, err := impl.gitFactory.GitService.Clone(repoUrl, chartDir)
	if err != nil {
		impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
		return "", err
	}
	return clonedDir, nil
}
