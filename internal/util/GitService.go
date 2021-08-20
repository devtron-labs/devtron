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
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

type GitClient interface {
	CreateRepository(name, description string) (url string, isNew bool, err error)
	CommitValues(config *ChartConfig) (commitHash string, err error)
	GetRepoUrl(projectName string) (repoUrl string, err error)
}

type GitFactory struct {
	Client           GitClient
	gitService       GitService
	GitWorkingDir    string
	logger           *zap.SugaredLogger
	gitOpsRepository repository.GitOpsConfigRepository
	gitCliUtil       *GitCliUtil
}

func (factory *GitFactory) Reload() error {
	logger.Infow("reloading gitops details")
	cfg, err := GetGitConfig(factory.gitOpsRepository)
	if err != nil {
		return err
	}
	gitService := NewGitServiceImpl(cfg, logger, factory.gitCliUtil)
	factory.gitService = gitService
	client, err := NewGitLabClient(cfg, logger, gitService)
	if err != nil {
		return err
	}
	factory.Client = client
	logger.Infow(" gitops details reload success")
	return nil
}

func NewGitFactory(logger *zap.SugaredLogger, gitOpsRepository repository.GitOpsConfigRepository, gitCliUtil *GitCliUtil) (*GitFactory, error) {
	cfg, err := GetGitConfig(gitOpsRepository)
	if err != nil {
		return nil, err
	}
	gitService := NewGitServiceImpl(cfg, logger, gitCliUtil)
	client, err := NewGitLabClient(cfg, logger, gitService)
	if err != nil {
		return nil, err
	}
	return &GitFactory{
		Client:           client,
		logger:           logger,
		gitService:       gitService,
		gitOpsRepository: gitOpsRepository,
		GitWorkingDir:    cfg.GitWorkingDir,
		gitCliUtil:       gitCliUtil,
	}, nil
}

type GitConfig struct {
	GitlabGroupId      string //local
	GitlabGroupPath    string //local
	GitToken           string //not null  // public
	GitUserName        string //not null  // public
	GitWorkingDir      string //working directory for git. might use pvc
	GithubOrganization string
	GitProvider        string // SUPPORTED VALUES  GITHUB, GITLAB
	GitHost            string
	AzureToken         string
	AzureProject       string
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
		GitlabGroupId:      gitOpsConfig.GitLabGroupId,
		GitToken:           gitOpsConfig.Token,
		GitUserName:        gitOpsConfig.Username,
		GitWorkingDir:      "/tmp/gitops/",
		GithubOrganization: gitOpsConfig.GitHubOrgId,
		GitProvider:        gitOpsConfig.Provider,
		GitHost:            gitOpsConfig.Host,
		AzureToken:         gitOpsConfig.Token,
		AzureProject:       gitOpsConfig.AzureProject,
	}
	return cfg, err
}

type GitLabClient struct {
	client     *gitlab.Client
	config     *GitConfig
	logger     *zap.SugaredLogger
	gitService GitService
}

func NewGitLabClient(config *GitConfig, logger *zap.SugaredLogger, gitService GitService) (GitClient, error) {
	if config.GitProvider == "GITLAB" {
		git := gitlab.NewClient(nil, config.GitToken)
		if len(config.GitHost) > 0 {
			_, err := url.ParseRequestURI(config.GitHost)
			if err != nil {
				return nil, err
			}
			err = git.SetBaseURL(config.GitHost)
			if err != nil {
				return nil, err
			}
		}

		gitlabGroupId := ""
		if len(config.GitlabGroupId) > 0 {
			if _, err := strconv.Atoi(config.GitlabGroupId); err == nil {
				gitlabGroupId = config.GitlabGroupId
			} else {
				groups, res, err := git.Groups.SearchGroup(config.GitlabGroupId)
				if err != nil {
					responseStatus := 0
					if res != nil {
						responseStatus = res.StatusCode

					}
					logger.Warnw("error connecting to gitlab", "status code", responseStatus, "err", err.Error())
				}
				logger.Debugw("gitlab groups found ", "group", groups)
				if len(groups) == 0 {
					logger.Warn("no matching namespace found for gitlab")
				}
				for _, group := range groups {
					if config.GitlabGroupId == group.Name {
						gitlabGroupId = strconv.Itoa(group.ID)
					}
				}
			}
		} else {
			return nil, fmt.Errorf("no gitlab group id found")
		}
		if len(gitlabGroupId) == 0 {
			return nil, fmt.Errorf("no gitlab group id found")
		}
		group, _, err := git.Groups.GetGroup(gitlabGroupId)
		if err != nil {
			return nil, err
		}
		if group != nil {
			config.GitlabGroupPath = group.FullPath
		}
		logger.Debugw("gitlab config", "config", config)
		return &GitLabClient{
			client:     git,
			config:     config,
			logger:     logger,
			gitService: gitService,
		}, nil
	} else if config.GitProvider == "GITHUB" {
		gitHubClient := NewGithubClient(config.GitToken, config.GithubOrganization, logger, gitService)
		return gitHubClient, nil
	} else if config.GitProvider == "AZURE_DEVOPS" {
		gitAzureClient := NewGitAzureClient(config.AzureToken, config.GitHost, config.AzureProject, logger, gitService)
		return gitAzureClient, nil
	} else {
		logger.Errorw("no gitops config provided, gitops will not work ")
		return nil, nil
	}
}

func (impl GitLabClient) CreateRepository(name, description string) (url string, isNew bool, err error) {
	impl.logger.Debugw("gitlab app create request ", "name", name, "description", description)
	repoUrl, err := impl.GetRepoUrl(name)
	if err != nil {
		impl.logger.Errorw("error in getting repo url ", "project", name, "err", err)
		return "", false, err
	}
	if len(repoUrl) > 0 {
		return repoUrl, false, nil
	} else {
		url, err = impl.createProject(name, description)
		if err != nil {
			return "", true, err
		}
	}
	repoUrl = url
	validated, err := impl.ensureProjectAvailability(name)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", name, "err", err)
		return "", true, err
	}
	if !validated {
		return "", true, fmt.Errorf("unable to validate project:%s  in given time", name)
	}
	_, err = impl.createReadme(impl.config.GitlabGroupPath, name)
	if err != nil {
		impl.logger.Errorw("error in creating readme ", "project", name, "err", err)
		return "", true, err
	}
	validated, err = impl.ensureProjectAvailabilityOnSsh(name, repoUrl)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", name, "err", err)
		return "", true, err
	}
	if !validated {
		return "", true, fmt.Errorf("unable to validate project:%s  in given time", name)
	}
	return url, true, nil
}

func (impl GitLabClient) DeleteProject(projectName string) (err error) {
	impl.logger.Infow("deleting project ", "name", projectName)
	_, err = impl.client.Projects.DeleteProject(fmt.Sprintf("%s/%s", impl.config.GitlabGroupPath, projectName))
	return err
}
func (impl GitLabClient) createProject(name, description string) (url string, err error) {
	var namespace = impl.config.GitlabGroupId
	namespaceId, err := strconv.Atoi(namespace)
	if err != nil {
		return "", err
	}

	// Create new project
	p := &gitlab.CreateProjectOptions{
		Name:                 gitlab.String(name),
		Description:          gitlab.String(description),
		MergeRequestsEnabled: gitlab.Bool(true),
		SnippetsEnabled:      gitlab.Bool(false),
		Visibility:           gitlab.Visibility(gitlab.PrivateVisibility),
		NamespaceID:          &namespaceId,
	}
	project, _, err := impl.client.Projects.CreateProject(p)
	if err != nil {
		impl.logger.Errorw("err in creating gitlab app", "req", p, "name", name, "err", err)
		return "", err
	}
	impl.logger.Infow("gitlab app created", "name", name, "url", project.HTTPURLToRepo)
	return project.HTTPURLToRepo, nil
}

func (impl GitLabClient) ensureProjectAvailability(projectName string) (bool, error) {
	pid := fmt.Sprintf("%s/%s", impl.config.GitlabGroupPath, projectName)
	count := 0
	verified := false
	for count < 3 && !verified {
		count = count + 1
		_, res, err := impl.client.Projects.GetProject(pid, &gitlab.GetProjectOptions{})
		if err != nil {
			return verified, err
		}
		if res.StatusCode >= 200 && res.StatusCode <= 299 {
			verified = true
			return verified, nil
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}

func (impl GitLabClient) ensureProjectAvailabilityOnSsh(projectName string, repoUrl string) (bool, error) {
	count := 0
	for count < 3 {
		count = count + 1
		_, err := impl.gitService.Clone(repoUrl, fmt.Sprintf("/ensure-clone/%s", projectName))
		if err == nil {
			impl.logger.Infow("ensureProjectAvailability clone passed", "try count", count, "repoUrl", repoUrl)
			return true, nil
		}
		if err != nil {
			impl.logger.Errorw("ensureProjectAvailability clone failed", "try count", count, "err", err)
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}

func (impl GitLabClient) GetRepoUrl(projectName string) (repoUrl string, err error) {
	pid := fmt.Sprintf("%s/%s", impl.config.GitlabGroupPath, projectName)
	prop, res, err := impl.client.Projects.GetProject(pid, &gitlab.GetProjectOptions{})
	if err != nil {
		impl.logger.Debugw("get project err", "pod", pid, "err", err)
		if res != nil && res.StatusCode == 404 {
			return "", nil
		}
		return "", err
	}
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return prop.HTTPURLToRepo, nil
	}
	return "", nil
}

func (impl GitLabClient) createReadme(namespace, projectName string) (res interface{}, err error) {
	actions := &gitlab.CreateCommitOptions{
		Branch:        gitlab.String("master"),
		CommitMessage: gitlab.String("test commit"),
		Actions:       []*gitlab.CommitAction{{Action: gitlab.FileCreate, FilePath: "README.md", Content: "devtron licence"}},
	}
	c, _, err := impl.client.Commits.CreateCommit(fmt.Sprintf("%s/%s", namespace, projectName), actions)
	return c, err
}
func (impl GitLabClient) checkIfFileExists(projectName, ref, file string) (exists bool, err error) {
	_, _, err = impl.client.RepositoryFiles.GetFileMetaData(fmt.Sprintf("%s/%s", impl.config.GitlabGroupPath, projectName), file, &gitlab.GetFileMetaDataOptions{Ref: &ref})
	return err == nil, err
}

func (impl GitLabClient) CommitValues(config *ChartConfig) (commitHash string, err error) {
	branch := "master"
	path := filepath.Join(config.ChartLocation, config.FileName)
	exists, err := impl.checkIfFileExists(config.ChartName, branch, path)
	var fileAction gitlab.FileAction
	if exists {
		fileAction = gitlab.FileUpdate
	} else {
		fileAction = gitlab.FileCreate
	}
	actions := &gitlab.CreateCommitOptions{
		Branch:        &branch,
		CommitMessage: gitlab.String(config.ReleaseMessage),
		Actions:       []*gitlab.CommitAction{{Action: fileAction, FilePath: path, Content: config.FileContent}},
	}
	c, _, err := impl.client.Commits.CreateCommit(fmt.Sprintf("%s/%s", impl.config.GitlabGroupPath, config.ChartName), actions)
	if err != nil {
		return "", err
	}
	return c.ID, err
}

type ChartConfig struct {
	ChartName      string
	ChartLocation  string
	FileName       string //filename
	FileContent    string
	ReleaseMessage string
}

//-------------------- go-git integration -------------------
type GitService interface {
	Clone(url, targetDir string) (clonedDir string, err error)
	CommitAndPushAllChanges(repoRoot, commitMsg string) (commitHash string, err error)
	ForceResetHead(repoRoot string) (err error)
	CommitValues(config *ChartConfig) (commitHash string, err error)

	GetCloneDirectory(targetDir string) (clonedDir string)
	Pull(repoRoot string) (err error)
}
type GitServiceImpl struct {
	Auth       *http.BasicAuth
	config     *GitConfig
	logger     *zap.SugaredLogger
	gitCliUtil *GitCliUtil
}

func NewGitServiceImpl(config *GitConfig, logger *zap.SugaredLogger, GitCliUtil *GitCliUtil) *GitServiceImpl {
	auth := &http.BasicAuth{Password: config.GitToken, Username: config.GitUserName}
	return &GitServiceImpl{
		Auth:       auth,
		logger:     logger,
		config:     config,
		gitCliUtil: GitCliUtil,
	}
}

func (impl GitServiceImpl) GetCloneDirectory(targetDir string) (clonedDir string) {
	clonedDir = filepath.Join(impl.config.GitWorkingDir, targetDir)
	return clonedDir
}

func (impl GitServiceImpl) Clone(url, targetDir string) (clonedDir string, err error) {
	impl.logger.Debugw("git checkout ", "url", url, "dir", targetDir)
	clonedDir = filepath.Join(impl.config.GitWorkingDir, targetDir)
	_, errorMsg, err := impl.gitCliUtil.Clone(clonedDir, url, impl.Auth.Username, impl.Auth.Password)
	if err != nil {
		impl.logger.Errorw("error in git checkout", "url", url, "targetDir", targetDir, "err", err)
		return "", err
	}
	if errorMsg != "" {
		return "", fmt.Errorf(errorMsg)
	}
	return clonedDir, nil
}

func (impl GitServiceImpl) CommitAndPushAllChanges(repoRoot, commitMsg string) (commitHash string, err error) {
	repo, workTree, err := impl.getRepoAndWorktree(repoRoot)
	if err != nil {
		return "", err
	}
	err = workTree.AddGlob("")
	if err != nil {
		return "", err
	}
	//--  commit
	commit, err := workTree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Devtron Boat",
			Email: "manifest-boat@github.com/devtron-labs",
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", err
	}
	impl.logger.Debugw("git hash", "repo", repoRoot, "hash", commit.String())
	//-----------push
	err = repo.Push(&git.PushOptions{
		Auth: impl.Auth,
	})

	return commit.String(), err
}

func (impl GitServiceImpl) getRepoAndWorktree(repoRoot string) (*git.Repository, *git.Worktree, error) {
	r, err := git.PlainOpen(repoRoot)
	if err != nil {
		return nil, nil, err
	}
	w, err := r.Worktree()
	return r, w, err
}

func (impl GitServiceImpl) ForceResetHead(repoRoot string) (err error) {
	_, workTree, err := impl.getRepoAndWorktree(repoRoot)
	if err != nil {
		return err
	}
	err = workTree.Reset(&git.ResetOptions{Mode: git.HardReset})
	if err != nil {
		return err
	}
	err = workTree.Pull(&git.PullOptions{
		Auth:         impl.Auth,
		Force:        true,
		SingleBranch: true,
	})
	return err
}

func (impl GitServiceImpl) CommitValues(config *ChartConfig) (commitHash string, err error) {
	//TODO acquire lock
	gitDir := filepath.Join(impl.config.GitWorkingDir, config.ChartName)
	if err != nil {
		return "", err
	}
	err = ioutil.WriteFile(filepath.Join(gitDir, config.ChartLocation, config.FileName), []byte(config.FileContent), 0600)
	if err != nil {
		return "", err
	}
	hash, err := impl.CommitAndPushAllChanges(gitDir, config.ReleaseMessage)
	return hash, err
}

func (impl GitServiceImpl) Pull(repoRoot string) (err error) {
	_, workTree, err := impl.getRepoAndWorktree(repoRoot)
	if err != nil {
		return err
	}
	//-----------pull
	err = workTree.PullContext(context.Background(), &git.PullOptions{
		Auth: impl.Auth,
	})
	if err != nil && err.Error() == "already up-to-date" {
		err = nil
		return nil
	}
	return err
}

//github

type GitHubClient struct {
	client *github.Client

	//config *GitConfig
	logger     *zap.SugaredLogger
	org        string
	gitService GitService
}

func NewGithubClient(token string, org string, logger *zap.SugaredLogger, gitService GitService) GitHubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return GitHubClient{client: client, org: org, logger: logger, gitService: gitService}
}

func (impl GitHubClient) CreateRepository(name, description string) (url string, isNew bool, err error) {
	ctx := context.Background()
	repoExists := true
	url, err = impl.GetRepoUrl(name)
	if err != nil {
		responseErr, ok := err.(*github.ErrorResponse)
		if !ok || responseErr.Response.StatusCode != 404 {
			impl.logger.Errorw("error in creating repo", "err", err)
			return "", false, err
		} else {
			repoExists = false
		}
	}
	if repoExists {
		return url, false, nil
	}
	private := true
//	visibility := "private"
	r, _, err := impl.client.Repositories.Create(ctx, impl.org,
		&github.Repository{Name: &name,
			Description: &description,
			Private:     &private,
//			Visibility:  &visibility,
		})
	if err != nil {
		impl.logger.Errorw("error in creating repo, ", "repo", name, "err", err)
		return "", true, err
	}
	logger.Infow("repo created ", "r", r.CloneURL)

	validated, err := impl.ensureProjectAvailabilityOnHttp(name)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", name, "err", err)
		return *r.CloneURL, true, err
	}
	if !validated {
		return "", true, fmt.Errorf("unable to validate project:%s  in given time", name)
	}
	_, err = impl.createReadme(name)
	if err != nil {
		impl.logger.Errorw("error in creating readme", "err", err)
		return *r.CloneURL, true, err
	}
	validated, err = impl.ensureProjectAvailabilityOnSsh(name, *r.CloneURL)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", name, "err", err)
		return *r.CloneURL, true, err
	}
	if !validated {
		return "", true, fmt.Errorf("unable to validate project:%s  in given time", name)
	}
	//_, err = impl.createReadme(name)
	return *r.CloneURL, true, err
}

func (impl GitHubClient) createReadme(repoName string) (string, error) {
	cfg := &ChartConfig{
		ChartName:      repoName,
		ChartLocation:  "",
		FileName:       "README.md",
		FileContent:    "@devtron",
		ReleaseMessage: "readme",
	}
	hash, err := impl.CommitValues(cfg)
	if err != nil {
		impl.logger.Errorw("error in creating readme", "repo", repoName, "err", err)
	}
	return hash, err
}

func (impl GitHubClient) CommitValues(config *ChartConfig) (commitHash string, err error) {
	branch := "master"
	path := filepath.Join(config.ChartLocation, config.FileName)
	ctx := context.Background()
	newFile := false
	fc, _, _, err := impl.client.Repositories.GetContents(ctx, impl.org, config.ChartName, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		responseErr, ok := err.(*github.ErrorResponse)
		if !ok || responseErr.Response.StatusCode != 404 {
			impl.logger.Errorw("error in creating repo", "err", err)
			return "", err
		} else {
			newFile = true
		}
	}
	currentSHA := ""
	if !newFile {
		currentSHA = *fc.SHA
	}
	options := &github.RepositoryContentFileOptions{
		Message: &config.ReleaseMessage,
		Content: []byte(config.FileContent),
		SHA:     &currentSHA,
		Branch:  &branch,
	}
	c, _, err := impl.client.Repositories.CreateFile(ctx, impl.org, config.ChartName, path, options)
	if err != nil {
		impl.logger.Errorw("error in commit", "err", err)
		return "", err
	}
	return *c.SHA, nil
}

func (impl GitHubClient) GetRepoUrl(projectName string) (repoUrl string, err error) {
	ctx := context.Background()
	repo, _, err := impl.client.Repositories.Get(ctx, impl.org, projectName)
	if err != nil {
		return "", err
	}
	return *repo.CloneURL, nil
}

func (impl GitHubClient) ensureProjectAvailabilityOnHttp(projectName string) (bool, error) {
	count := 0
	for count < 3 {
		count = count + 1
		_, err := impl.GetRepoUrl(projectName)
		if err == nil {
			return true, nil
		}
		responseErr, ok := err.(*github.ErrorResponse)
		if !ok || responseErr.Response.StatusCode != 404 {
			impl.logger.Errorw("error in validating repo", "err", err)
			return false, err
		} else {
			impl.logger.Errorw("error in validating repo", "err", err)
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}

func (impl GitHubClient) ensureProjectAvailabilityOnSsh(projectName string, repoUrl string) (bool, error) {
	count := 0
	for count < 3 {
		count = count + 1
		_, err := impl.gitService.Clone(repoUrl, fmt.Sprintf("/ensure-clone/%s", projectName))
		if err == nil {
			impl.logger.Infow("ensureProjectAvailability clone passed", "try count", count, "repoUrl", repoUrl)
			return true, nil
		}
		if err != nil {
			impl.logger.Errorw("ensureProjectAvailability clone failed", "try count", count, "err", err)
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}
