package config

import (
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config/bean"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"regexp"
	"strings"
)

type GitOpsConfigReadService interface {
	IsGitOpsConfigured() (bool, error)
	GetUserEmailIdAndNameForGitOpsCommit(userId int32) (string, string)
	GetGitOpsRepoName(appName string) string
	GetGitOpsRepoNameFromUrl(gitRepoUrl string) string
	GetBitbucketMetadata() (*bean.BitbucketProviderMetadata, error)
	GetGitOpsConfigActive() (*bean2.GitOpsConfigDto, error)
	GetConfiguredGitOpsCount() (int, error)
}

type GitOpsConfigReadServiceImpl struct {
	logger             *zap.SugaredLogger
	gitOpsRepository   repository.GitOpsConfigRepository
	userService        user.UserService
	globalEnvVariables *util.GlobalEnvVariables
}

func NewGitOpsConfigReadServiceImpl(logger *zap.SugaredLogger,
	gitOpsRepository repository.GitOpsConfigRepository,
	userService user.UserService,
	envVariables *util.EnvironmentVariables) *GitOpsConfigReadServiceImpl {
	return &GitOpsConfigReadServiceImpl{
		logger:             logger,
		gitOpsRepository:   gitOpsRepository,
		userService:        userService,
		globalEnvVariables: envVariables.GlobalEnvVariables,
	}
}

func (impl *GitOpsConfigReadServiceImpl) IsGitOpsConfigured() (bool, error) {
	isGitOpsConfigured := false
	gitOpsConfig, err := impl.gitOpsRepository.GetGitOpsConfigActive()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("GetGitOpsConfigActive, error while getting", "err", err)
		return false, err
	}
	if gitOpsConfig != nil && gitOpsConfig.Id > 0 {
		isGitOpsConfigured = true
	}
	return isGitOpsConfigured, nil
}

func (impl *GitOpsConfigReadServiceImpl) GetUserEmailIdAndNameForGitOpsCommit(userId int32) (string, string) {
	emailId := bean.GitOpsCommitDefaultEmailId
	name := bean.GitOpsCommitDefaultName
	//getting emailId associated with user
	userEmail, err := impl.userService.GetEmailById(userId)
	if err != nil {
		impl.logger.Errorw("error in getting user info by id", "err", err, "id", userId)
	}
	//TODO: export constant in user bean
	if userEmail != "admin" && userEmail != "system" && len(userEmail) > 0 {
		emailId = userEmail
	} else {
		emailIdGitOps, err := impl.gitOpsRepository.GetEmailIdFromActiveGitOpsConfig()
		if err != nil {
			impl.logger.Errorw("error in getting emailId from active gitOps config", "err", err)
		} else if len(emailIdGitOps) > 0 {
			emailId = emailIdGitOps
		}
	}
	//we are getting name from emailId(replacing special characters in <user-name part of email> with space)
	emailComponents := strings.Split(emailId, "@")
	regex, _ := regexp.Compile(`[^\w]`)
	if regex != nil {
		name = regex.ReplaceAllString(emailComponents[0], " ")
	}
	return emailId, name
}

func (impl *GitOpsConfigReadServiceImpl) GetGitOpsRepoName(appName string) string {
	var repoName string
	if len(impl.globalEnvVariables.GitOpsRepoPrefix) == 0 {
		repoName = appName
	} else {
		repoName = fmt.Sprintf("%s-%s", impl.globalEnvVariables.GitOpsRepoPrefix, appName)
	}
	return repoName
}

func (impl *GitOpsConfigReadServiceImpl) GetGitOpsRepoNameFromUrl(gitRepoUrl string) string {
	gitRepoUrl = gitRepoUrl[strings.LastIndex(gitRepoUrl, "/")+1:]
	gitRepoUrl = strings.ReplaceAll(gitRepoUrl, ".git", "")
	return gitRepoUrl
}

func (impl *GitOpsConfigReadServiceImpl) GetBitbucketMetadata() (*bean.BitbucketProviderMetadata, error) {
	metadata := &bean.BitbucketProviderMetadata{}
	gitOpsConfigBitbucket, err := impl.gitOpsRepository.GetGitOpsConfigByProvider(bean.BITBUCKET_PROVIDER)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching gitOps bitbucket config", "err", err)
		return nil, err
	}
	if gitOpsConfigBitbucket != nil {
		metadata.BitBucketWorkspaceId = gitOpsConfigBitbucket.BitBucketWorkspaceId
		metadata.BitBucketProjectKey = gitOpsConfigBitbucket.BitBucketProjectKey
	}
	return metadata, nil
}

func (impl *GitOpsConfigReadServiceImpl) GetGitOpsConfigActive() (*bean2.GitOpsConfigDto, error) {
	model, err := impl.gitOpsRepository.GetGitOpsConfigActive()
	if err != nil {
		impl.logger.Errorw("error, GetGitOpsConfigActive", "err", err)
		return nil, err
	}
	config := &bean2.GitOpsConfigDto{
		Id:                   model.Id,
		Provider:             model.Provider,
		GitHubOrgId:          model.GitHubOrgId,
		GitLabGroupId:        model.GitLabGroupId,
		Active:               model.Active,
		UserId:               model.CreatedBy,
		AzureProjectName:     model.AzureProject,
		BitBucketWorkspaceId: model.BitBucketWorkspaceId,
		BitBucketProjectKey:  model.BitBucketProjectKey,
	}
	return config, err
}

func (impl *GitOpsConfigReadServiceImpl) GetConfiguredGitOpsCount() (int, error) {
	count := 0
	models, err := impl.gitOpsRepository.GetAllGitOpsConfig()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error, GetGitOpsConfigActive", "err", err)
		return count, err
	}
	count = len(models)
	return count, nil
}
