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
	"github.com/devtron-labs/common-lib/utils/runTime"
	bean2 "github.com/devtron-labs/devtron/api/bean/gitOps"
	globalUtil "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/retryFunc"
	"github.com/google/go-github/github"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	http2 "net/http"
	"net/url"
	"path"
	"path/filepath"
	"time"
)

type GitHubClient struct {
	client       *github.Client
	logger       *zap.SugaredLogger
	org          string
	gitOpsHelper *GitOpsHelper
}

func NewGithubClient(host string, token string, org string, logger *zap.SugaredLogger,
	gitOpsHelper *GitOpsHelper, tlsConfig *tls.Config) (GitHubClient, error) {
	ctx := context.Background()

	httpTransport := &http2.Transport{Proxy: http2.ProxyFromEnvironment, TLSClientConfig: tlsConfig}
	httpClient := &http2.Client{Transport: httpTransport}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpClient)

	tc := oauth2.NewClient(ctx, ts)
	var client *github.Client
	var err error
	hostUrl, err := url.Parse(host)
	if err != nil {
		logger.Errorw("error in creating git client ", "host", hostUrl, "err", err)
		return GitHubClient{}, err
	}
	if hostUrl.Host == GITHUB_HOST {
		client = github.NewClient(tc)
	} else {
		logger.Infow("creating github EnterpriseClient with org", "host", host, "org", org)
		hostUrl.Path = path.Join(hostUrl.Path, GITHUB_API_V3)
		client, err = github.NewEnterpriseClient(hostUrl.String(), hostUrl.String(), tc)
	}

	return GitHubClient{
		client:       client,
		org:          org,
		logger:       logger,
		gitOpsHelper: gitOpsHelper,
	}, err
}

func (impl GitHubClient) DeleteRepository(config *bean2.GitOpsConfigDto) error {
	var err error
	start := time.Now()
	defer func() {
		globalUtil.TriggerGitOpsMetrics("DeleteRepository", "GitHubClient", start, err)
	}()

	_, err = impl.client.Repositories.Delete(context.Background(), config.GitHubOrgId, config.GitRepoName)
	if err != nil {
		impl.logger.Errorw("repo deletion failed for github", "repo", config.GitRepoName, "err", err)
		return err
	}
	return nil
}

func IsRepoNotFound(err error) bool {
	if err == nil {
		return false
	}
	var responseErr *github.ErrorResponse
	ok := errors.As(err, &responseErr)
	return ok && responseErr.Response.StatusCode == 404
}

func (impl GitHubClient) CreateRepository(ctx context.Context, config *bean2.GitOpsConfigDto) (url string, isNew bool, isEmpty bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions) {

	var err error

	start := time.Now()

	detailedErrorGitOpsConfigActions.StageErrorMap = make(map[string]error)
	repoExists := true
	url, isEmpty, err = impl.getRepoUrl(ctx, config, IsRepoNotFound)
	if err != nil {
		if IsRepoNotFound(err) {
			repoExists = false
		} else {
			impl.logger.Errorw("error in creating github repo", "err", err)
			detailedErrorGitOpsConfigActions.StageErrorMap[GetRepoUrlStage] = err
			globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitHubClient", start, err)
			return "", false, isEmpty, detailedErrorGitOpsConfigActions
		}
	}
	if repoExists {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, GetRepoUrlStage)
		globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitHubClient", start, nil)
		return url, false, isEmpty, detailedErrorGitOpsConfigActions
	}
	private := true
	//	visibility := "private"
	r, _, err1 := impl.client.Repositories.Create(ctx, impl.org,
		&github.Repository{Name: &config.GitRepoName,
			Description: &config.Description,
			Private:     &private,
			//			Visibility:  &visibility,
		})
	if err1 != nil {
		impl.logger.Errorw("error in creating github repo, ", "repo", config.GitRepoName, "err", err1)
		url, isEmpty, err = impl.GetRepoUrl(config)
		if err != nil {
			impl.logger.Errorw("error in getting github repo", "repo", config.GitRepoName, "err", err)
			detailedErrorGitOpsConfigActions.StageErrorMap[CreateRepoStage] = err1
			globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitHubClient", start, err1)
			return "", true, isEmpty, detailedErrorGitOpsConfigActions
		}
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, GetRepoUrlStage)
		globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitHubClient", start, nil)
		return url, false, isEmpty, detailedErrorGitOpsConfigActions
	}
	impl.logger.Infow("github repo created ", "r", r.CloneURL)
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateRepoStage)

	validated, err := impl.ensureProjectAvailabilityOnHttp(config)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability github", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = err
		globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitHubClient", start, err)
		return *r.CloneURL, true, isEmpty, detailedErrorGitOpsConfigActions
	}
	if !validated {
		err = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = err
		globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitHubClient", start, err)
		return "", true, isEmpty, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneHttpStage)

	_, err = impl.CreateReadme(ctx, config)
	if err != nil {
		impl.logger.Errorw("error in creating readme github", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateReadmeStage] = err
		globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitHubClient", start, err)
		return *r.CloneURL, true, isEmpty, detailedErrorGitOpsConfigActions
	}
	isEmpty = false //As we have created readme, repo is no longer empty
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateReadmeStage)

	validated, err = impl.ensureProjectAvailabilityOnSsh(config.GitRepoName, *r.CloneURL, config.TargetRevision)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability github", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = err
		globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitHubClient", start, err)
		return *r.CloneURL, true, isEmpty, detailedErrorGitOpsConfigActions
	}
	if !validated {
		err = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = err
		globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitHubClient", start, err)
		return "", true, isEmpty, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneSshStage)
	//_, err = impl.createReadme(name)
	globalUtil.TriggerGitOpsMetrics("CreateRepository", "GitHubClient", start, nil)
	return *r.CloneURL, true, isEmpty, detailedErrorGitOpsConfigActions
}

func (impl GitHubClient) CreateFirstCommitOnHead(ctx context.Context, config *bean2.GitOpsConfigDto) (string, error) {
	return impl.CreateReadme(ctx, config)
}

func (impl GitHubClient) CreateReadme(ctx context.Context, config *bean2.GitOpsConfigDto) (string, error) {
	var err error
	start := time.Now()
	defer func() {
		globalUtil.TriggerGitOpsMetrics("CreateReadme", "GitHubClient", start, err)
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
	hash, _, err := impl.CommitValues(ctx, cfg, config, true)
	if err != nil {
		impl.logger.Errorw("error in creating readme github", "repo", config.GitRepoName, "err", err)
	}
	return hash, err
}

func (impl GitHubClient) CommitValues(ctx context.Context, config *ChartConfig, gitOpsConfig *bean2.GitOpsConfigDto, publishStatusConflictErrorMetrics bool) (commitHash string, commitTime time.Time, err error) {

	start := time.Now()

	branch := config.TargetRevision
	if len(branch) == 0 {
		branch = globalUtil.GetDefaultTargetRevision()
	}
	path := filepath.Join(config.ChartLocation, config.FileName)
	newFile := false
	fc, _, _, err := impl.client.Repositories.GetContents(ctx, impl.org, config.ChartRepoName, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		responseErr, ok := err.(*github.ErrorResponse)
		if !ok || responseErr.Response.StatusCode != 404 {
			impl.logger.Errorw("error in creating repo github", "err", err, "config", config)
			globalUtil.TriggerGitOpsMetrics("CommitValues", "GitHubClient", start, err)
			return "", time.Time{}, err
		} else {
			newFile = true
		}
	}
	currentSHA := ""
	if !newFile {
		currentSHA = *fc.SHA
	}
	timeNow := time.Now()
	options := &github.RepositoryContentFileOptions{
		Message: &config.ReleaseMessage,
		Content: []byte(config.FileContent),
		SHA:     &currentSHA,
		Branch:  &branch,
		Author: &github.CommitAuthor{
			Date:  &timeNow,
			Email: &config.UserEmailId,
			Name:  &config.UserName,
		},
		Committer: &github.CommitAuthor{
			Date:  &timeNow,
			Email: &config.UserEmailId,
			Name:  &config.UserName,
		},
	}
	c, httpRes, err := impl.client.Repositories.CreateFile(ctx, impl.org, config.ChartRepoName, path, options)
	if err != nil && httpRes != nil && httpRes.StatusCode == http2.StatusConflict {
		impl.logger.Warn("conflict found in commit github", "err", err, "config", config)
		if publishStatusConflictErrorMetrics {
			globalUtil.TriggerGitOpsMetrics("CommitValues", "GitHubClient", start, err)
		}
		return "", time.Time{}, retryFunc.NewRetryableError(err)
	} else if err != nil {
		impl.logger.Errorw("error in commit github", "err", err, "config", config)
		globalUtil.TriggerGitOpsMetrics("CommitValues", "GitHubClient", start, err)
		return "", time.Time{}, err
	}
	commitTime = time.Now() // default is current time, if found then will get updated accordingly
	if c != nil && c.Commit.Author != nil {
		commitTime = *c.Commit.Author.Date
	}
	globalUtil.TriggerGitOpsMetrics("CommitValues", "GitHubClient", start, nil)
	return *c.SHA, commitTime, nil
}

func (impl GitHubClient) GetRepoUrl(config *bean2.GitOpsConfigDto) (repoUrl string, isRepoEmpty bool, err error) {
	ctx := context.Background()
	return impl.getRepoUrl(ctx, config, globalUtil.AllPublishableError())
}

func (impl GitHubClient) getRepoUrl(ctx context.Context, config *bean2.GitOpsConfigDto, isNonPublishableError globalUtil.EvalIsNonPublishableErr) (repoUrl string, isRepoEmpty bool, err error) {
	start := time.Now()
	defer func() {
		if isNonPublishableError(err) {
			impl.logger.Debugw("found non publishable error. skipping metrics publish!", "caller method", runTime.GetCallerFunctionName(), "err", err)
			return
		}
		globalUtil.TriggerGitOpsMetrics("GetRepoUrl", "GitHubClient", start, err)
	}()

	repo, _, err := impl.client.Repositories.Get(ctx, impl.org, config.GitRepoName)
	if err != nil {
		impl.logger.Errorw("error in getting repo url by repo name", "org", impl.org, "gitRepoName", config.GitRepoName, "err", err)
		return "", false, err
	}
	return repo.GetCloneURL(), repo.GetSize() == 0, nil
}

func (impl GitHubClient) ensureProjectAvailabilityOnHttp(config *bean2.GitOpsConfigDto) (bool, error) {
	var err error
	start := time.Now()
	defer func() {
		globalUtil.TriggerGitOpsMetrics("ensureProjectAvailabilityOnHttp", "GitHubClient", start, err)
	}()

	count := 0
	for count < 3 {
		count = count + 1
		_, _, err := impl.GetRepoUrl(config)
		if err == nil {
			return true, nil
		}
		responseErr, ok := err.(*github.ErrorResponse)
		if !ok || responseErr.Response.StatusCode != 404 {
			impl.logger.Errorw("error in validating repo github", "project", config.GitRepoName, "err", err)
			return false, err
		} else {
			impl.logger.Errorw("error in validating repo github", "project", config.GitRepoName, "err", err)
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}

func (impl GitHubClient) ensureProjectAvailabilityOnSsh(projectName string, repoUrl, targetRevision string) (bool, error) {
	var err error
	start := time.Now()
	defer func() {
		globalUtil.TriggerGitOpsMetrics("ensureProjectAvailabilityOnSsh", "GitHubClient", start, err)
	}()

	count := 0
	for count < 3 {
		count = count + 1
		_, err := impl.gitOpsHelper.Clone(repoUrl, fmt.Sprintf("/ensure-clone/%s", projectName), targetRevision)
		if err == nil {
			impl.logger.Infow("github ensureProjectAvailability clone passed", "try count", count, "repoUrl", repoUrl)
			return true, nil
		} else {
			impl.logger.Errorw("github ensureProjectAvailability clone failed", "try count", count, "err", err)
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}
