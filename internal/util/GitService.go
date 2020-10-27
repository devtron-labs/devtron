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
	"github.com/caarlos0/env"
	"github.com/google/go-github/github"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"path/filepath"
	"time"
)

type GitClient interface {
	CreateRepository(name, description string) (url string, isNew bool, err error)
	CommitValues(config *ChartConfig) (commitHash string, err error)
	GetRepoUrl(projectName string) (repoUrl string, err error)
}

type GitConfig struct {
	GitlabNamespaceID   int    `env:"GITLAB_NAMESPACE_ID" `                            //not null //local
	GitlabNamespaceName string `env:"GITLAB_NAMESPACE_NAME" `                          //local
	GitToken            string `env:"GIT_TOKEN" `                                      //not null  // public
	GitUserName         string `env:"GIT_USERNAME" `                                   //not null  // public
	GitWorkingDir       string `env:"GIT_WORKING_DIRECTORY" envDefault:"/tmp/gitops/"` //working directory for git. might use pvc
	GithubOrganization  string `env:"GITHUB_ORGANIZATION"`
	GitProvider         string `env:"GIT_PROVIDER" envDefault:"GITHUB"` // SUPPORTED VALUES  GITHUB, GITLAB
}

func (cfg *GitConfig) GetUserName() (userName string) {
	return cfg.GitUserName
}
func (cfg *GitConfig) GetPassword() (password string) {
	return cfg.GitToken
}

func GetGitConfig() (*GitConfig, error) {
	cfg := &GitConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type GitLabClient struct {
	client *gitlab.Client
	config *GitConfig
	logger *zap.SugaredLogger
}

func NewGitLabClient(config *GitConfig, logger *zap.SugaredLogger) (GitClient, error) {
	if config.GitProvider == "GITLAB" {
		git := gitlab.NewClient(nil, config.GitToken)
		return &GitLabClient{
			client: git,
			config: config,
			logger: logger,
		}, nil
	} else if config.GitProvider == "GITHUB" {
		return NewGithubClient(config.GitToken, config.GithubOrganization, logger), nil
	} else {
		return nil, fmt.Errorf("unsupported git provider %s, supported values are  GITHUB, GITLAB", config.GitProvider)
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

	validated, err := impl.ensureProjectAvailability(name)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", name, "err", err)
		return "", true, err
	}
	if !validated {
		return "", true, fmt.Errorf("unable to validate project:%s  in given time", name)
	}
	_, err = impl.createReadme(impl.config.GitlabNamespaceName, name)
	if err != nil {
		impl.logger.Errorw("error in creating readme ", "project", name, "err", err)
		return "", true, err
	}
	return url, true, nil
}

func (impl GitLabClient) DeleteProject(projectName string) (err error) {
	impl.logger.Infow("deleting project ", "name", projectName)
	_, err = impl.client.Projects.DeleteProject(fmt.Sprintf("%s/%s", impl.config.GitlabNamespaceName, projectName))
	return err
}
func (impl GitLabClient) createProject(name, description string) (url string, err error) {
	var namespace = impl.config.GitlabNamespaceID
	// Create new project
	p := &gitlab.CreateProjectOptions{
		Name:                 gitlab.String(name),
		Description:          gitlab.String(description),
		MergeRequestsEnabled: gitlab.Bool(true),
		SnippetsEnabled:      gitlab.Bool(false),
		Visibility:           gitlab.Visibility(gitlab.PrivateVisibility),
		NamespaceID:          &namespace,
	}
	project, _, err := impl.client.Projects.CreateProject(p)
	if err != nil {
		impl.logger.Errorw("err in creating gitlab app", "name", name, "err", err)
		return "", err
	}
	impl.logger.Infow("gitlab app created", "name", name, "url", project.HTTPURLToRepo)
	return project.HTTPURLToRepo, nil
}

func (impl GitLabClient) ensureProjectAvailability(projectName string) (validated bool, err error) {
	pid := fmt.Sprintf("%s/%s", impl.config.GitlabNamespaceName, projectName)
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
func (impl GitLabClient) GetRepoUrl(projectName string) (repoUrl string, err error) {
	pid := fmt.Sprintf("%s/%s", impl.config.GitlabNamespaceName, projectName)
	prop, res, err := impl.client.Projects.GetProject(pid, &gitlab.GetProjectOptions{})
	if err != nil {
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
	_, _, err = impl.client.RepositoryFiles.GetFileMetaData(fmt.Sprintf("%s/%s", impl.config.GitlabNamespaceName, projectName), file, &gitlab.GetFileMetaDataOptions{Ref: &ref})
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
	c, _, err := impl.client.Commits.CreateCommit(fmt.Sprintf("%s/%s", impl.config.GitlabNamespaceName, config.ChartName), actions)
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
	Auth   transport.AuthMethod
	config *GitConfig
	logger *zap.SugaredLogger
}

func NewGitServiceImpl(config *GitConfig, logger *zap.SugaredLogger) *GitServiceImpl {
	auth := &http.BasicAuth{Password: config.GetPassword(), Username: config.GetUserName()}
	return &GitServiceImpl{
		Auth:   auth,
		logger: logger,
		config: config,
	}
}

func (impl GitServiceImpl) GetCloneDirectory(targetDir string) (clonedDir string) {
	clonedDir = filepath.Join(impl.config.GitWorkingDir, targetDir)
	return clonedDir
}

func (impl GitServiceImpl) Clone(url, targetDir string) (clonedDir string, err error) {
	impl.logger.Debugw("git checkout ", "url", url, "dir", targetDir)
	clonedDir = filepath.Join(impl.config.GitWorkingDir, targetDir)
	_, err = git.PlainClone(clonedDir, false, &git.CloneOptions{
		URL:  url,
		Auth: impl.Auth,
	})
	if err != nil {
		impl.logger.Errorw("error in git checkout ", "url", url, "targetDir", targetDir, "err", err)
		return "", err
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
	logger *zap.SugaredLogger
	org    string
}

func NewGithubClient(token string, org string, logger *zap.SugaredLogger) GitHubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)
	return GitHubClient{client: client, org: org, logger: logger}
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
	visibility := "private"
	r, _, err := impl.client.Repositories.Create(ctx, impl.org,
		&github.Repository{Name: &name,
			Description: &description,
			Private:     &private,
			Visibility:  &visibility,
		})
	if err != nil {
		impl.logger.Errorw("error in creating repo, ", "repo", name, "err", err)
	}
	logger.Infow("repo created ", "r", r.CloneURL)

	validated, err := impl.ensureProjectAvailability(name)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", name, "err", err)
		return "", true, err
	}
	if !validated {
		return "", true, fmt.Errorf("unable to validate project:%s  in given time", name)
	}
	_, err = impl.createReadme(name)
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
		impl.logger.Errorw("erorr in commit", "err", err)
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

func (impl GitHubClient) ensureProjectAvailability(projectName string) (validated bool, err error) {
	count := 0
	verified := false
	for count < 3 && !verified {
		count = count + 1
		_, err := impl.GetRepoUrl(projectName)
		if err==nil{
			return true, nil
		}
		responseErr, ok := err.(*github.ErrorResponse)
		if !ok || responseErr.Response.StatusCode != 404 {
			impl.logger.Errorw("error in validating repo", "err", err)
			return validated, err
		}else {
			impl.logger.Errorw("error in validating repo", "err", err)
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}
