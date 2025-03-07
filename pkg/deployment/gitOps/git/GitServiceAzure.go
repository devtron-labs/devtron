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

package git

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean/gitOps"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/gitUtil"
	"github.com/devtron-labs/devtron/util/retryFunc"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"go.uber.org/zap"
	http2 "net/http"
	"path/filepath"
	"time"
)

type GitAzureClient struct {
	client       *git.Client
	logger       *zap.SugaredLogger
	project      string
	gitOpsHelper *GitOpsHelper
}

func (impl GitAzureClient) GetRepoUrl(config *bean2.GitOpsConfigDto) (repoUrl string, isRepoEmpty bool, err error) {

	start := time.Now()
	defer func() {
		globalUtil.TriggerGitOpsMetrics("GetRepoUrl", "GitAzureClient", start, err)
	}()

	var (
		url    string
		exists bool
	)
	url, exists, isRepoEmpty, err = impl.repoExists(config.GitRepoName, impl.project)
	if err != nil {
		return "", isRepoEmpty, err
	} else if !exists {
		return "", isRepoEmpty, fmt.Errorf("%s :repo not found", config.GitRepoName)
	} else {
		return url, isRepoEmpty, nil
	}
}

func NewGitAzureClient(token string, host string, project string, logger *zap.SugaredLogger, gitOpsHelper *GitOpsHelper, tlsConfig *tls.Config) (GitAzureClient, error) {
	ctx := context.Background()
	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(host, token)
	connection.TlsConfig = tlsConfig
	// Create a client to interact with the Core area
	coreClient, err := git.NewClient(ctx, connection)
	if err != nil {
		logger.Errorw("error in creating azure gitops client, gitops related operation might fail", "err", err)
	}
	return GitAzureClient{
		client:       &coreClient,
		project:      project,
		logger:       logger,
		gitOpsHelper: gitOpsHelper,
	}, err
}

func (impl GitAzureClient) DeleteRepository(config *bean2.GitOpsConfigDto) (err error) {
	start := time.Now()
	defer func() {
		globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitAzureClient", start, err)
	}()

	clientAzure := *impl.client
	gitRepository, err := clientAzure.GetRepository(context.Background(), git.GetRepositoryArgs{
		RepositoryId: &config.GitRepoName,
		Project:      &config.AzureProjectName,
	})
	if err != nil || gitRepository == nil {
		impl.logger.Errorw("error in fetching repo azure", "project", config.GitRepoName, "err", err)
		return err
	}
	err = clientAzure.DeleteRepository(context.Background(), git.DeleteRepositoryArgs{RepositoryId: gitRepository.Id, Project: &impl.project})
	if err != nil {
		impl.logger.Errorw("error in deleting repo azure", "project", config.GitRepoName, "err", err)
	}
	return err
}

func (impl GitAzureClient) CreateRepository(ctx context.Context, config *bean2.GitOpsConfigDto) (url string, isNew bool, isEmpty bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions) {
	var (
		err        error
		repoExists bool
	)
	start := time.Now()
	defer func() {
		globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitAzureClient", start, err)
	}()

	detailedErrorGitOpsConfigActions.StageErrorMap = make(map[string]error)
	url, repoExists, isEmpty, err = impl.repoExists(config.GitRepoName, impl.project)
	if err != nil {
		impl.logger.Errorw("error in communication with azure", "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[GetRepoUrlStage] = err
		return "", false, isEmpty, detailedErrorGitOpsConfigActions
	}
	if repoExists {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, GetRepoUrlStage)
		return url, false, isEmpty, detailedErrorGitOpsConfigActions
	}
	gitRepositoryCreateOptions := git.GitRepositoryCreateOptions{
		Name: &config.GitRepoName,
	}
	clientAzure := *impl.client
	operationReference, err := clientAzure.CreateRepository(ctx, git.CreateRepositoryArgs{
		GitRepositoryToCreate: &gitRepositoryCreateOptions,
		Project:               &impl.project,
	})
	if err != nil {
		impl.logger.Errorw("error in creating repo azure", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateRepoStage] = err
		url, repoExists, isEmpty, err = impl.repoExists(config.GitRepoName, impl.project)
		if err != nil {
			impl.logger.Errorw("error in communication with azure", "err", err)
		}
		if err != nil || !repoExists {
			return "", true, isEmpty, detailedErrorGitOpsConfigActions
		}
	}
	impl.logger.Infow("repo created ", "r", operationReference.WebUrl)
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateRepoStage)
	validated, err := impl.ensureProjectAvailabilityOnHttp(config.GitRepoName)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability azure", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = err
		return *operationReference.WebUrl, true, isEmpty, detailedErrorGitOpsConfigActions
	}
	if !validated {
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		return "", true, isEmpty, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneHttpStage)

	_, err = impl.CreateReadme(ctx, config)
	if err != nil {
		impl.logger.Errorw("error in creating readme azure", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateReadmeStage] = err
		return *operationReference.WebUrl, true, isEmpty, detailedErrorGitOpsConfigActions
	}
	isEmpty = false //As we have created readme, repo is no longer empty
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateReadmeStage)
	validated, err = impl.ensureProjectAvailabilityOnSsh(impl.project, *operationReference.WebUrl, config.TargetRevision)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability azure", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = err
		return *operationReference.WebUrl, true, isEmpty, detailedErrorGitOpsConfigActions
	}
	if !validated {
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		return "", true, isEmpty, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneSshStage)
	return *operationReference.WebUrl, true, isEmpty, detailedErrorGitOpsConfigActions
}

func (impl GitAzureClient) CreateReadme(ctx context.Context, config *bean2.GitOpsConfigDto) (string, error) {

	var err error
	start := time.Now()
	defer func() {
		globalUtil.TriggerGitOpsMetrics("CreateReadme", "GitAzureClient", start, err)
	}()

	cfg := &ChartConfig{
		ChartName:      config.GitRepoName,
		ChartLocation:  "",
		FileName:       "README.md",
		FileContent:    "@devtron",
		ReleaseMessage: "readme",
		ChartRepoName:  config.GitRepoName,
		TargetRevision: config.TargetRevision,
		UserName:       config.Username,
		UserEmailId:    config.UserEmailId,
	}
	hash, _, err := impl.CommitValues(ctx, cfg, config)
	if err != nil {
		impl.logger.Errorw("error in creating readme azure", "repo", config.GitRepoName, "err", err)
	}
	return hash, err
}

func (impl GitAzureClient) CommitValues(ctx context.Context, config *ChartConfig, gitOpsConfig *bean2.GitOpsConfigDto) (commitHash string, commitTime time.Time, err error) {
	branch := config.TargetRevision
	if len(branch) == 0 {
		branch = globalUtil.GetDefaultTargetRevision()
	}
	branchRefHead := gitUtil.GetRefBranchHead(branch)
	path := filepath.Join(config.ChartLocation, config.FileName)
	newFile := true
	oldObjId := "0000000000000000000000000000000000000000" //default commit hash
	// check if file exists and current hash
	// if file does not exist get hash from branch
	// if branch doesn't exist use default hash
	clientAzure := *impl.client
	fc, err := clientAzure.GetItem(ctx, git.GetItemArgs{
		RepositoryId: &config.ChartRepoName,
		Path:         &path,
		Project:      &impl.project,
	})
	if err != nil {
		notFoundStatus := 404
		if e, ok := err.(azuredevops.WrappedError); ok && *e.StatusCode == notFoundStatus {
			clientAzure := *impl.client
			branchStat, err := clientAzure.GetBranch(ctx, git.GetBranchArgs{Project: &impl.project, Name: &branch, RepositoryId: &config.ChartRepoName})
			if err != nil {
				if e, ok := err.(azuredevops.WrappedError); !ok || *e.StatusCode >= 500 {
					impl.logger.Errorw("error in fetching branch from azure devops", "err", err)
					return "", time.Time{}, err
				}
			} else if branchStat != nil {
				oldObjId = *branchStat.Commit.CommitId
			}
		} else {
			impl.logger.Errorw("error in fetching file from azure devops", "err", err)
			return "", time.Time{}, err
		}
	} else {
		oldObjId = *fc.CommitId
		newFile = false
	}

	var refUpdates []git.GitRefUpdate
	refUpdates = append(refUpdates, git.GitRefUpdate{
		Name:        &branchRefHead,
		OldObjectId: &oldObjId,
	})
	var changeType git.VersionControlChangeType
	if newFile {
		changeType = git.VersionControlChangeTypeValues.Add
	} else {
		changeType = git.VersionControlChangeTypeValues.Edit
	}
	gitChange := git.GitChange{ChangeType: &changeType,
		Item: &git.GitItemDescriptor{Path: &path},
		NewContent: &git.ItemContent{
			Content:     &config.FileContent,
			ContentType: &git.ItemContentTypeValues.RawText,
		}}
	var contents []interface{}
	contents = append(contents, gitChange)

	var commits []git.GitCommitRef
	commits = append(commits, git.GitCommitRef{
		Changes: &contents,
		Comment: &config.ReleaseMessage,
		Author: &git.GitUserDate{
			Date: &azuredevops.Time{
				Time: time.Now(),
			},
			Email: &config.UserEmailId,
			Name:  &config.UserName,
		},
		Committer: &git.GitUserDate{
			Date: &azuredevops.Time{
				Time: time.Now(),
			},
			Email: &config.UserEmailId,
			Name:  &config.UserName,
		},
	})
	push, err := clientAzure.CreatePush(ctx, git.CreatePushArgs{
		Push: &git.GitPush{
			Commits:    &commits,
			RefUpdates: &refUpdates,
		},
		RepositoryId: &config.ChartRepoName,
		Project:      &impl.project,
	})
	if e := (azuredevops.WrappedError{}); errors.As(err, &e) && e.StatusCode != nil && *e.StatusCode == http2.StatusConflict {
		impl.logger.Warn("conflict found in commit azure", "err", err, "config", config)
		return "", time.Time{}, retryFunc.NewRetryableError(err)
	} else if err != nil {
		impl.logger.Errorw("error in commit azure", "err", err)
		return "", time.Time{}, err
	}
	//gitPush.Commits
	commitId := ""
	commitAuthorTime := time.Now() //default is current time, if found then will get updated accordingly
	if len(*push.Commits) > 0 {
		commitId = *(*push.Commits)[0].CommitId
		commitAuthorTime = (*push.Commits)[0].Author.Date.Time
	}
	//	push.Commits[0].CommitId
	return commitId, commitAuthorTime, nil
}

func (impl GitAzureClient) repoExists(repoName, projectName string) (repoUrl string, exists, isRepoEmpty bool, err error) {

	start := time.Now()
	defer func() {
		globalUtil.TriggerGitOpsMetrics("repoExists", "GitAzureClient", start, err)
	}()

	ctx := context.Background()
	// Get first page of the list of team projects for your organization
	clientAzure := *impl.client
	gitRepository, err := clientAzure.GetRepository(ctx, git.GetRepositoryArgs{
		RepositoryId: &repoName,
		Project:      &projectName,
	})
	notFoundStatus := 404
	if err != nil {
		if e, ok := err.(azuredevops.WrappedError); ok && *e.StatusCode == notFoundStatus {
			return "", false, isRepoEmpty, nil
		} else {
			return "", false, isRepoEmpty, err
		}

	}
	for gitRepository == nil {
		return "", false, isRepoEmpty, nil
	}

	return *gitRepository.WebUrl, true, *gitRepository.Size == 0, nil
}

func (impl GitAzureClient) ensureProjectAvailabilityOnHttp(repoName string) (bool, error) {
	var err error
	start := time.Now()
	defer func() {
		globalUtil.TriggerGitOpsMetrics("ensureProjectAvailabilityOnHttp", "GitAzureClient", start, err)
	}()

	for count := 0; count < 5; count++ {
		_, exists, _, err := impl.repoExists(repoName, impl.project)
		if err == nil && exists {
			impl.logger.Infow("repo validated successfully on https")
			return true, nil
		} else if err != nil {
			impl.logger.Errorw("error in validating repo azure", "repo", repoName, "err", err)
			return false, err
		} else {
			impl.logger.Errorw("repo not available on http", "repo")
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}

func (impl GitAzureClient) ensureProjectAvailabilityOnSsh(projectName string, repoUrl string, targetRevision string) (bool, error) {
	for count := 0; count < 8; count++ {
		_, err := impl.gitOpsHelper.Clone(repoUrl, fmt.Sprintf("/ensure-clone/%s", projectName), targetRevision)
		if err == nil {
			impl.logger.Infow("ensureProjectAvailability clone passed azure", "try count", count, "repoUrl", repoUrl)
			return true, nil
		}
		impl.logger.Errorw("ensureProjectAvailability clone failed ssh azure", "try count", count, "err", err)
		time.Sleep(10 * time.Second)
	}
	return false, nil
}
