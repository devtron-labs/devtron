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

package util

import (
	"fmt"
	"github.com/devtron-labs/devtron/util"
	"net/url"
	"path/filepath"
	"time"

	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
)

const (
	GIT_WORKING_DIR       = "/tmp/gitops/"
	GetRepoUrlStage       = "Get Repo RedirectionUrl"
	CreateRepoStage       = "Create Repo"
	CloneHttpStage        = "Clone Http"
	CreateReadmeStage     = "Create Readme"
	CloneSshStage         = "Clone Ssh"
	GITLAB_PROVIDER       = "GITLAB"
	GITHUB_PROVIDER       = "GITHUB"
	AZURE_DEVOPS_PROVIDER = "AZURE_DEVOPS"
	BITBUCKET_PROVIDER    = "BITBUCKET_CLOUD"
	GITHUB_API_V3         = "api/v3"
	GITHUB_HOST           = "github.com"
)

type GitClient interface {
	CreateRepository(config *bean2.GitOpsConfigDto) (url string, isNew bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions)
	CommitValues(config *ChartConfig, gitOpsConfig *bean2.GitOpsConfigDto) (commitHash string, commitTime time.Time, err error)
	GetRepoUrl(config *bean2.GitOpsConfigDto) (repoUrl string, err error)
	DeleteRepository(config *bean2.GitOpsConfigDto) error
	CreateReadme(config *bean2.GitOpsConfigDto) (string, error)
}

type GitFactory struct {
	Client           GitClient
	GitService       GitService
	GitWorkingDir    string
	logger           *zap.SugaredLogger
	gitOpsRepository repository.GitOpsConfigRepository
	gitManager       *GitManagerImpl
}

type DetailedErrorGitOpsConfigActions struct {
	SuccessfulStages []string         `json:"successfulStages"`
	StageErrorMap    map[string]error `json:"stageErrorMap"`
	ValidatedOn      time.Time        `json:"validatedOn"`
	DeleteRepoFailed bool             `json:"deleteRepoFailed"`
}

type GitCommitDto struct {
	CommitHash string    `json:"commitHash"`
	AuthorName string    `json:"authorName"`
	CommitTime time.Time `json:"commitTime"`
}

func (factory *GitFactory) Reload() error {
	var err error
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Reload", "GitService", start, err)
	}()
	logger.Infow("reloading gitops details")
	cfg, err := GetGitConfig(factory.gitOpsRepository)
	if err != nil {
		return err
	}
	gitService := NewGitServiceImpl(cfg, logger, factory.gitManager)
	factory.GitService = gitService
	client, err := NewGitOpsClient(cfg, logger, gitService, factory.gitOpsRepository)
	if err != nil {
		return err
	}
	factory.Client = client
	logger.Infow(" gitops details reload success")
	return nil
}

func (factory *GitFactory) GetGitLabGroupPath(gitOpsConfig *bean2.GitOpsConfigDto) (string, error) {
	start := time.Now()
	var gitLabClient *gitlab.Client
	var err error
	defer func() {
		util.TriggerGitOpsMetrics("GetGitLabGroupPath", "GitService", start, err)
	}()
	if len(gitOpsConfig.Host) > 0 {
		_, err = url.ParseRequestURI(gitOpsConfig.Host)
		if err != nil {
			return "", err
		}
		gitLabClient, err = gitlab.NewClient(gitOpsConfig.Token, gitlab.WithBaseURL(gitOpsConfig.Host))
		if err != nil {
			factory.logger.Errorw("error in getting new gitlab client", "err", err)
			return "", err
		}
	} else {
		gitLabClient, err = gitlab.NewClient(gitOpsConfig.Token)
		if err != nil {
			factory.logger.Errorw("error in getting new gitlab client", "err", err)
			return "", err
		}
	}
	group, _, err := gitLabClient.Groups.GetGroup(gitOpsConfig.GitLabGroupId, &gitlab.GetGroupOptions{})
	if err != nil {
		factory.logger.Errorw("error in fetching gitlab group name", "err", err, "gitLab groupID", gitOpsConfig.GitLabGroupId)
		return "", err
	}
	if group == nil {
		factory.logger.Errorw("no matching groups found for gitlab", "gitLab groupID", gitOpsConfig.GitLabGroupId, "err", err)
		return "", fmt.Errorf("no matching groups found for gitlab group ID : %s", gitOpsConfig.GitLabGroupId)
	}
	return group.FullPath, nil
}

func (factory *GitFactory) NewClientForValidation(gitOpsConfig *bean2.GitOpsConfigDto) (GitClient, *GitServiceImpl, error) {
	start := time.Now()
	var err error
	defer func() {
		util.TriggerGitOpsMetrics("NewClientForValidation", "GitService", start, err)
	}()
	cfg := &GitConfig{
		GitlabGroupId:        gitOpsConfig.GitLabGroupId,
		GitToken:             gitOpsConfig.Token,
		GitUserName:          gitOpsConfig.Username,
		GitWorkingDir:        GIT_WORKING_DIR,
		GithubOrganization:   gitOpsConfig.GitHubOrgId,
		GitProvider:          gitOpsConfig.Provider,
		GitHost:              gitOpsConfig.Host,
		AzureToken:           gitOpsConfig.Token,
		AzureProject:         gitOpsConfig.AzureProjectName,
		BitbucketWorkspaceId: gitOpsConfig.BitBucketWorkspaceId,
		BitbucketProjectKey:  gitOpsConfig.BitBucketProjectKey,
	}
	gitService := NewGitServiceImpl(cfg, logger, factory.gitManager)
	//factory.GitService = GitService
	client, err := NewGitOpsClient(cfg, logger, gitService, factory.gitOpsRepository)
	if err != nil {
		return client, gitService, err
	}

	//factory.Client = client
	logger.Infow("client changed successfully", "cfg", cfg)
	return client, gitService, nil
}

func NewGitFactory(logger *zap.SugaredLogger, gitOpsRepository repository.GitOpsConfigRepository, gitManager *GitManagerImpl) (*GitFactory, error) {
	cfg, err := GetGitConfig(gitOpsRepository)
	if err != nil {
		return nil, err
	}
	gitService := NewGitServiceImpl(cfg, logger, gitManager)
	client, err := NewGitOpsClient(cfg, logger, gitService, gitOpsRepository)
	if err != nil {
		logger.Errorw("error in creating gitOps client", "err", err, "gitProvider", cfg.GitProvider)
	}

	return &GitFactory{
		Client:           client,
		logger:           logger,
		GitService:       gitService,
		gitOpsRepository: gitOpsRepository,
		GitWorkingDir:    cfg.GitWorkingDir,
		gitManager:       gitManager,
	}, nil
}

type GitConfig struct {
	GitlabGroupId        string //local
	GitlabGroupPath      string //local
	GitToken             string //not null  // public
	GitUserName          string //not null  // public
	GitWorkingDir        string //working directory for git. might use pvc
	GithubOrganization   string
	GitProvider          string // SUPPORTED VALUES  GITHUB, GITLAB
	GitHost              string
	AzureToken           string
	AzureProject         string
	BitbucketWorkspaceId string
	BitbucketProjectKey  string
}

func GetGitConfig(gitOpsRepository repository.GitOpsConfigRepository) (*GitConfig, error) {
	gitOpsConfig, err := gitOpsRepository.GetGitOpsConfigActive()
	if err != nil && err != pg.ErrNoRows {
		return nil, err
	} else if err == pg.ErrNoRows {
		// adding this block for backward compatibility,TODO: remove in next  iteration
		// cfg := &GitConfig{}
		// err := env.Parse(cfg)
		// return cfg, err
		return &GitConfig{}, nil
	}

	if gitOpsConfig == nil || gitOpsConfig.Id == 0 {
		return nil, err
	}
	cfg := &GitConfig{
		GitlabGroupId:        gitOpsConfig.GitLabGroupId,
		GitToken:             gitOpsConfig.Token,
		GitUserName:          gitOpsConfig.Username,
		GitWorkingDir:        GIT_WORKING_DIR,
		GithubOrganization:   gitOpsConfig.GitHubOrgId,
		GitProvider:          gitOpsConfig.Provider,
		GitHost:              gitOpsConfig.Host,
		AzureToken:           gitOpsConfig.Token,
		AzureProject:         gitOpsConfig.AzureProject,
		BitbucketWorkspaceId: gitOpsConfig.BitBucketWorkspaceId,
		BitbucketProjectKey:  gitOpsConfig.BitBucketProjectKey,
	}
	return cfg, err
}

func NewGitOpsClient(config *GitConfig, logger *zap.SugaredLogger, gitService GitService, gitOpsConfigRepository repository.GitOpsConfigRepository) (GitClient, error) {
	if config.GitProvider == GITLAB_PROVIDER {
		gitLabClient, err := NewGitLabClient(config, logger, gitService)
		return gitLabClient, err
	} else if config.GitProvider == GITHUB_PROVIDER {
		gitHubClient, err := NewGithubClient(config.GitHost, config.GitToken, config.GithubOrganization, logger, gitService, gitOpsConfigRepository)
		return gitHubClient, err
	} else if config.GitProvider == AZURE_DEVOPS_PROVIDER {
		gitAzureClient, err := NewGitAzureClient(config.AzureToken, config.GitHost, config.AzureProject, logger, gitService, gitOpsConfigRepository)
		return gitAzureClient, err
	} else if config.GitProvider == BITBUCKET_PROVIDER {
		gitBitbucketClient := NewGitBitbucketClient(config.GitUserName, config.GitToken, config.GitHost, logger, gitService, gitOpsConfigRepository)
		return gitBitbucketClient, nil
	} else {
		logger.Errorw("no gitops config provided, gitops will not work ")
		return nil, nil
	}
}

type ChartConfig struct {
	ChartName      string
	ChartLocation  string
	FileName       string //filename
	FileContent    string
	ReleaseMessage string
	ChartRepoName  string
	UserName       string
	UserEmailId    string
}

// -------------------- go-git and cli integration -------------------
type GitService interface {
	Clone(url, targetDir string) (clonedDir string, err error)
	CommitAndPushAllChanges(repoRoot, commitMsg, name, emailId string) (commitHash string, err error)
	GetCloneDirectory(targetDir string) (clonedDir string)
	Pull(repoRoot string) (err error)
}

// BasicAuth represent a HTTP basic auth
type BasicAuth struct {
	username, password string
}

type GitServiceImpl struct {
	Auth       *BasicAuth
	config     *GitConfig
	logger     *zap.SugaredLogger
	gitManager *GitManagerImpl
}

func NewGitServiceImpl(config *GitConfig, logger *zap.SugaredLogger, gitManager *GitManagerImpl) *GitServiceImpl {
	auth := &BasicAuth{password: config.GitToken, username: config.GitUserName}

	return &GitServiceImpl{
		Auth:       auth,
		logger:     logger,
		config:     config,
		gitManager: gitManager,
	}
}

func (impl GitServiceImpl) GetCloneDirectory(targetDir string) (clonedDir string) {

	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("GetCloneDirectory", "GitService", start, nil)
	}()
	clonedDir = filepath.Join(impl.config.GitWorkingDir, targetDir)
	return clonedDir
}

func (impl GitServiceImpl) Clone(url, targetDir string) (clonedDir string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Clone", "GitService", start, err)
	}()
	impl.logger.Debugw("git checkout ", "url", url, "dir", targetDir)
	clonedDir = filepath.Join(impl.config.GitWorkingDir, targetDir)
	errorMsg, err := impl.gitManager.Clone(clonedDir, url, impl.Auth)
	if err != nil {
		impl.logger.Errorw("error in git checkout", "url", url, "targetDir", targetDir, "err", err)
		return "", err
	}
	if errorMsg != "" {
		return "", fmt.Errorf(errorMsg)
	}
	return clonedDir, nil
}

func (impl GitServiceImpl) Pull(repoRoot string) (err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("Pull", "GitService", start, err)
	}()
	return impl.gitManager.Pull(repoRoot, impl.Auth)
}

func (impl GitServiceImpl) CommitAndPushAllChanges(repoRoot, commitMsg, name, emailId string) (commitHash string, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("CommitAndPushAllChanges", "GitService", start, err)
	}()
	return impl.gitManager.CommitAndPush(repoRoot, commitMsg, name, emailId, impl.Auth)
}
