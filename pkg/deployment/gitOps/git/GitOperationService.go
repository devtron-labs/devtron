package git

import (
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/util"
	util2 "github.com/devtron-labs/devtron/pkg/appStore/util"
	commonBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/common/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git/bean"
	dirCopy "github.com/otiai10/copy"
	"go.uber.org/zap"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type GitOperationService interface {
	CreateGitRepositoryForApp(gitOpsRepoName string, userId int32) (chartGitAttribute *commonBean.ChartGitAttribute, err error)
	CreateReadmeInGitRepo(gitOpsRepoName string, userId int32) error
	GitPull(clonedDir string, repoUrl string, appStoreName string) error

	CommitValues(chartGitAttr *ChartConfig) (commitHash string, commitTime time.Time, err error)
	CommitAndPushAllChanges(clonedDir, commitMsg, userName, userEmailId string) (commitHash string, err error)
	PushChartToGitRepo(gitOpsRepoName, referenceTemplate, version,
		tempReferenceTemplateDir string, repoUrl string, userId int32) (err error)
	PushChartToGitOpsRepoForHelmApp(PushChartToGitRequest *bean.PushChartToGitRequestDTO,
		requirementsConfig *ChartConfig, valuesConfig *ChartConfig) (*commonBean.ChartGitAttribute, string, error)

	CreateRepository(dto *apiBean.GitOpsConfigDto, userId int32) (string, bool, error)
	GetRepoUrlByRepoName(repoName string) (string, error)

	GetClonedDir(chartDir, repoUrl string) (string, error)
	CloneInDir(repoUrl, chartDir string) (string, error)
	ReloadGitOpsProvider() error
	UpdateGitHostUrlByProvider(request *apiBean.GitOpsConfigDto) error
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

func (impl *GitOperationServiceImpl) CreateGitRepositoryForApp(gitOpsRepoName string, userId int32) (chartGitAttribute *commonBean.ChartGitAttribute, err error) {
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
	gitRepoRequest := &apiBean.GitOpsConfigDto{
		GitRepoName:          gitOpsRepoName,
		Description:          fmt.Sprintf("helm chart for " + gitOpsRepoName),
		Username:             userName,
		UserEmailId:          userEmailId,
		BitBucketWorkspaceId: bitbucketMetadata.BitBucketWorkspaceId,
		BitBucketProjectKey:  bitbucketMetadata.BitBucketProjectKey,
	}
	repoUrl, isNew, detailedError := impl.gitFactory.Client.CreateRepository(gitRepoRequest)
	for _, err := range detailedError.StageErrorMap {
		if err != nil {
			impl.logger.Errorw("error in creating git project", "name", gitOpsRepoName, "err", err)
			return nil, err
		}
	}
	return &commonBean.ChartGitAttribute{RepoUrl: repoUrl, IsNewRepo: isNew}, nil
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
		commit, err := impl.gitFactory.GitOpsHelper.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
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
			commit, err = impl.gitFactory.GitOpsHelper.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
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

func (impl *GitOperationServiceImpl) GitPull(clonedDir string, repoUrl string, appStoreName string) error {
	//TODO refactoring: remove invalid param appStoreName
	//TODO check for local repo exists before clone
	//TODO verify remote has repoUrl; or delete and clone
	err := impl.gitFactory.GitOpsHelper.Pull(clonedDir)
	if err != nil {
		impl.logger.Errorw("error in pulling git", "clonedDir", clonedDir, "err", err)
		_, err := impl.gitFactory.GitOpsHelper.Clone(repoUrl, appStoreName)
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
	gitOpsConfig := &apiBean.GitOpsConfigDto{BitBucketWorkspaceId: bitbucketMetadata.BitBucketWorkspaceId}
	commitHash, commitTime, err = impl.gitFactory.Client.CommitValues(chartGitAttr, gitOpsConfig)
	if err != nil {
		impl.logger.Errorw("error in git commit", "err", err)
		return commitHash, commitTime, err
	}
	return commitHash, commitTime, nil
}

func (impl *GitOperationServiceImpl) CommitAndPushAllChanges(clonedDir, commitMsg, userName, userEmailId string) (commitHash string, err error) {
	commitHash, err = impl.gitFactory.GitOpsHelper.CommitAndPushAllChanges(clonedDir, commitMsg, userName, userEmailId)
	if err != nil {
		impl.logger.Errorw("error in pushing git", "err", err)
		return commitHash, err
	}
	return commitHash, nil
}

func (impl *GitOperationServiceImpl) CreateRepository(dto *apiBean.GitOpsConfigDto, userId int32) (string, bool, error) {
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
	dto := &apiBean.GitOpsConfigDto{
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

// TODO refactoring: Make a common method for both PushChartToGitRepo and PushChartToGitOpsRepoForHelmApp
// PushChartToGitOpsRepoForHelmApp pushes built chart to GitOps repo (Specific implementation for Helm Apps)
func (impl *GitOperationServiceImpl) PushChartToGitOpsRepoForHelmApp(PushChartToGitRequest *bean.PushChartToGitRequestDTO, requirementsConfig *ChartConfig, valuesConfig *ChartConfig) (*commonBean.ChartGitAttribute, string, error) {
	space := regexp.MustCompile(`\s+`)
	appStoreName := space.ReplaceAllString(PushChartToGitRequest.ChartAppStoreName, "-")
	chartDir := fmt.Sprintf("%s-%s", PushChartToGitRequest.AppName, impl.chartTemplateService.GetDir())
	clonedDir := impl.gitFactory.GitOpsHelper.GetCloneDirectory(chartDir)
	if _, err := os.Stat(clonedDir); os.IsNotExist(err) {
		clonedDir, err = impl.gitFactory.GitOpsHelper.Clone(PushChartToGitRequest.RepoURL, chartDir)
		if err != nil {
			impl.logger.Errorw("error in cloning repo", "url", PushChartToGitRequest.RepoURL, "err", err)
			return nil, "", err
		}
	} else {
		err = impl.GitPull(clonedDir, PushChartToGitRequest.RepoURL, appStoreName)
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
	err = impl.addConfigFileToChart(requirementsConfig, dir, clonedDir)
	if err != nil {
		impl.logger.Errorw("error in adding requirements.yaml to chart", "err", err, "appName", PushChartToGitRequest.AppName)
		return nil, "", err
	}
	err = impl.addConfigFileToChart(valuesConfig, dir, clonedDir)
	if err != nil {
		impl.logger.Errorw("error in adding values.yaml to chart", "err", err, "appName", PushChartToGitRequest.AppName)
		return nil, "", err
	}
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(PushChartToGitRequest.UserId)
	commit, err := impl.gitFactory.GitOpsHelper.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
	if err != nil {
		impl.logger.Errorw("error in pushing git", "err", err)
		impl.logger.Warn("re-trying, taking pull and then push again")
		err = impl.GitPull(clonedDir, PushChartToGitRequest.RepoURL, acdAppName)
		if err != nil {
			impl.logger.Errorw("error in git pull", "err", err, "appName", acdAppName)
			return nil, "", err
		}
		err = dirCopy.Copy(PushChartToGitRequest.TempChartRefDir, dir)
		if err != nil {
			impl.logger.Errorw("error copying dir", "err", err)
			return nil, "", err
		}
		commit, err = impl.gitFactory.GitOpsHelper.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
		if err != nil {
			impl.logger.Errorw("error in pushing git", "err", err)
			return nil, "", err
		}
	}
	impl.logger.Debugw("template committed", "url", PushChartToGitRequest.RepoURL, "commit", commit)
	defer impl.chartTemplateService.CleanDir(clonedDir)
	return &commonBean.ChartGitAttribute{RepoUrl: PushChartToGitRequest.RepoURL, ChartLocation: acdAppName}, commit, err
}

func (impl *GitOperationServiceImpl) GetClonedDir(chartDir, repoUrl string) (string, error) {
	clonedDir := impl.gitFactory.GitOpsHelper.GetCloneDirectory(chartDir)
	if _, err := os.Stat(clonedDir); os.IsNotExist(err) {
		return impl.CloneInDir(repoUrl, chartDir)
	} else if err != nil {
		impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
		return "", err
	}
	return clonedDir, nil
}

func (impl *GitOperationServiceImpl) CloneInDir(repoUrl, chartDir string) (string, error) {
	clonedDir, err := impl.gitFactory.GitOpsHelper.Clone(repoUrl, chartDir)
	if err != nil {
		impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
		return "", err
	}
	return clonedDir, nil
}
func (impl *GitOperationServiceImpl) ReloadGitOpsProvider() error {
	return impl.gitFactory.Reload(impl.gitOpsConfigReadService)
}

func (impl *GitOperationServiceImpl) UpdateGitHostUrlByProvider(request *apiBean.GitOpsConfigDto) error {
	switch strings.ToUpper(request.Provider) {
	case GITHUB_PROVIDER:
		orgUrl, err := buildGithubOrgUrl(request.Host, request.GitHubOrgId)
		if err != nil {
			return err
		}
		request.Host = orgUrl

	case GITLAB_PROVIDER:
		groupName, err := impl.gitFactory.GetGitLabGroupPath(request)
		if err != nil {
			return err
		}
		slashSuffixPresent := strings.HasSuffix(request.Host, "/")
		if slashSuffixPresent {
			request.Host += groupName
		} else {
			request.Host = fmt.Sprintf(request.Host+"/%s", groupName)
		}
	case BITBUCKET_PROVIDER:
		request.Host = BITBUCKET_CLONE_BASE_URL + request.BitBucketWorkspaceId
	}
	return nil
}

func buildGithubOrgUrl(host, orgId string) (orgUrl string, err error) {
	if !strings.HasPrefix(host, HTTP_URL_PROTOCOL) && !strings.HasPrefix(host, HTTPS_URL_PROTOCOL) {
		return orgUrl, fmt.Errorf("invalid host url '%s'", host)
	}
	hostUrl, err := url.Parse(host)
	if err != nil {
		return "", err
	}
	hostUrl.Path = path.Join(hostUrl.Path, orgId)
	return hostUrl.String(), nil
}

// addConfigFileToChart will override requirements.yaml or values.yaml file in chart
func (impl *GitOperationServiceImpl) addConfigFileToChart(config *ChartConfig, destinationDir string, clonedDir string) error {
	filePath := filepath.Join(clonedDir, config.FileName)
	filePath, err := util2.CreateFileAtFilePathAndWrite(filePath, config.FileContent)
	if err != nil {
		impl.logger.Errorw("error in creating yaml file", "err", err)
		return err
	}
	destinationFilePath := filepath.Join(destinationDir, config.FileName)
	err = util2.MoveFileToDestination(filePath, destinationFilePath)
	if err != nil {
		impl.logger.Errorw("error in moving file from source to destination", "err", err)
		return err
	}
	return nil
}
