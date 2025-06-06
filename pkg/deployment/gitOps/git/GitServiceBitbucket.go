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
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/retryFunc"
	"github.com/devtron-labs/go-bitbucket"
	"go.uber.org/zap"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	HTTP_URL_PROTOCOL              = "http://"
	HTTPS_URL_PROTOCOL             = "https://"
	BITBUCKET_CLONE_BASE_URL       = "https://bitbucket.org/"
	BITBUCKET_GITOPS_DIR           = "bitbucketGitOps"
	BITBUCKET_REPO_NOT_FOUND_ERROR = "404 Not Found"
	BITBUCKET_COMMIT_TIME_LAYOUT   = "2001-01-01T10:00:00+00:00"
)

type GitBitbucketClient struct {
	client       *bitbucket.Client
	logger       *zap.SugaredLogger
	gitOpsHelper *GitOpsHelper
}

func NewGitBitbucketClient(username, token, host string, logger *zap.SugaredLogger, gitOpsHelper *GitOpsHelper, tlsConfig *tls.Config) GitBitbucketClient {
	coreClient := bitbucket.NewBasicAuth(username, token)
	httpClient := util.GetHTTPClientWithTLSConfig(tlsConfig)
	coreClient.HttpClient = httpClient
	logger.Infow("bitbucket client created", "clientDetails", coreClient)
	return GitBitbucketClient{
		client:       coreClient,
		logger:       logger,
		gitOpsHelper: gitOpsHelper,
	}
}

func (impl GitBitbucketClient) DeleteRepository(config *bean2.GitOpsConfigDto) (err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("DeleteRepository", "GitBitbucketClient", start, err)
	}()
	repoOptions := &bitbucket.RepositoryOptions{
		Owner:     config.BitBucketWorkspaceId,
		RepoSlug:  config.GitRepoName,
		IsPrivate: "true",
		Project:   config.BitBucketProjectKey,
	}
	_, err = impl.client.Repositories.Repository.Delete(repoOptions)
	if err != nil {
		impl.logger.Errorw("error in deleting repo gitlab", "repoName", config.GitRepoName, "err", err)
	}
	return err
}

func (impl GitBitbucketClient) GetRepoUrl(config *bean2.GitOpsConfigDto) (repoUrl string, isRepoEmpty bool, err error) {
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("GetRepoUrl", "GitBitbucketClient", start, err)
	}()

	repoOptions := &bitbucket.RepositoryOptions{
		Owner:    config.BitBucketWorkspaceId,
		Project:  config.BitBucketProjectKey,
		RepoSlug: config.GitRepoName,
	}
	_, exists, err := impl.repoExists(repoOptions)
	if err != nil {
		return "", isRepoEmpty, err
	} else if !exists {
		return "", isRepoEmpty, fmt.Errorf("%s :repo not found", repoOptions.RepoSlug)
	} else {
		repoUrl = fmt.Sprintf(BITBUCKET_CLONE_BASE_URL+"%s/%s.git", repoOptions.Owner, repoOptions.RepoSlug)
		return repoUrl, isRepoEmpty, nil
	}
}

func (impl GitBitbucketClient) CreateRepository(ctx context.Context, config *bean2.GitOpsConfigDto) (url string, isNew bool, isEmpty bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions) {
	var err error
	start := time.Now()

	detailedErrorGitOpsConfigActions.StageErrorMap = make(map[string]error)

	workSpaceId := config.BitBucketWorkspaceId
	projectKey := config.BitBucketProjectKey
	repoOptions := &bitbucket.RepositoryOptions{
		Owner:       workSpaceId,
		RepoSlug:    config.GitRepoName,
		Scm:         "git",
		IsPrivate:   "true",
		Description: config.Description,
		Project:     projectKey,
	}

	repoUrl, repoExists, err := impl.repoExists(repoOptions)
	if err != nil {
		impl.logger.Errorw("error in communication with bitbucket", "repoOptions", repoOptions, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[GetRepoUrlStage] = err
		util.TriggerGitOpsMetrics("CreateRepository", "GitBitbucketClient", start, err)
		return "", false, isEmpty, detailedErrorGitOpsConfigActions
	}
	if repoExists {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, GetRepoUrlStage)
		util.TriggerGitOpsMetrics("CreateRepository", "GitBitbucketClient", start, nil)
		return repoUrl, false, isEmpty, detailedErrorGitOpsConfigActions
	}
	_, err = impl.client.Repositories.Repository.Create(repoOptions)
	if err != nil {
		impl.logger.Errorw("error in creating repo bitbucket", "repoOptions", repoOptions, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateRepoStage] = err
		repoUrl, repoExists, err = impl.repoExists(repoOptions)
		if err != nil {
			impl.logger.Errorw("error in creating repo bitbucket", "repoOptions", repoOptions, "err", err)
		}
		if err != nil || !repoExists {
			util.TriggerGitOpsMetrics("CreateRepository", "GitBitbucketClient", start, err)
			return "", true, isEmpty, detailedErrorGitOpsConfigActions
		}
	}
	repoUrl = fmt.Sprintf(BITBUCKET_CLONE_BASE_URL+"%s/%s.git", repoOptions.Owner, repoOptions.RepoSlug)
	impl.logger.Infow("repo created ", "repoUrl", repoUrl)
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateRepoStage)

	validated, err := impl.ensureProjectAvailabilityOnHttp(repoOptions)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability bitbucket", "repoName", repoOptions.RepoSlug, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = err
		util.TriggerGitOpsMetrics("CreateRepository", "GitBitbucketClient", start, err)
		return "", true, isEmpty, detailedErrorGitOpsConfigActions
	}
	if !validated {
		err = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = err
		util.TriggerGitOpsMetrics("CreateRepository", "GitBitbucketClient", start, err)
		return "", true, isEmpty, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneHttpStage)

	_, err = impl.CreateReadme(ctx, config)
	if err != nil {
		impl.logger.Errorw("error in creating readme bitbucket", "repoName", repoOptions.RepoSlug, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateReadmeStage] = err
		util.TriggerGitOpsMetrics("CreateRepository", "GitBitbucketClient", start, err)
		return "", true, isEmpty, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateReadmeStage)

	validated, err = impl.ensureProjectAvailabilityOnSsh(repoOptions, config.TargetRevision)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability bitbucket", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = err
		util.TriggerGitOpsMetrics("CreateRepository", "GitBitbucketClient", start, err)
		return "", true, isEmpty, detailedErrorGitOpsConfigActions
	}
	if !validated {
		err = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = err
		util.TriggerGitOpsMetrics("CreateRepository", "GitBitbucketClient", start, err)
		return "", true, isEmpty, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneSshStage)
	util.TriggerGitOpsMetrics("CreateRepository", "GitBitbucketClient", start, nil)
	return repoUrl, true, isEmpty, detailedErrorGitOpsConfigActions
}

func (impl GitBitbucketClient) repoExists(repoOptions *bitbucket.RepositoryOptions) (repoUrl string, exists bool, err error) {

	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("repoExists", "GitBitbucketClient", start, err)
	}()

	repo, err := impl.client.Repositories.Repository.Get(repoOptions)
	if repo == nil && err.Error() == BITBUCKET_REPO_NOT_FOUND_ERROR {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	repoUrl = fmt.Sprintf(BITBUCKET_CLONE_BASE_URL+"%s/%s.git", repoOptions.Owner, repoOptions.RepoSlug)
	return repoUrl, true, nil
}
func (impl GitBitbucketClient) ensureProjectAvailabilityOnHttp(repoOptions *bitbucket.RepositoryOptions) (bool, error) {
	for count := 0; count < 5; count++ {
		_, exists, err := impl.repoExists(repoOptions)
		if err == nil && exists {
			impl.logger.Infow("repo validated successfully on https")
			return true, nil
		} else if err != nil {
			impl.logger.Errorw("error in validating repo bitbucket", "repoDetails", repoOptions, "err", err)
			return false, err
		} else {
			impl.logger.Errorw("repo not available on http", "repoDetails", repoOptions)
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}

func getDir() string {
	/* #nosec */
	r1 := rand.New(rand.NewSource(time.Now().UnixNano())).Int63()
	return strconv.FormatInt(r1, 10)
}

func (impl GitBitbucketClient) CreateReadme(ctx context.Context, config *bean2.GitOpsConfigDto) (string, error) {
	var err error
	start := time.Now()
	defer func() {
		util.TriggerGitOpsMetrics("CreateReadme", "GitBitbucketClient", start, err)
	}()

	cfg := &ChartConfig{
		ChartName:      config.GitRepoName,
		ChartLocation:  "",
		FileName:       "README.md",
		FileContent:    "@devtron",
		ReleaseMessage: "pushing readme",
		ChartRepoName:  config.GitRepoName,
		TargetRevision: config.TargetRevision,
		UserName:       config.Username,
		UserEmailId:    config.UserEmailId,
	}
	cfg.SetBitBucketBaseDir(getDir())
	hash, _, err := impl.CommitValues(ctx, cfg, config, true)
	if err != nil {
		impl.logger.Errorw("error in creating readme bitbucket", "repo", config.GitRepoName, "err", err)
	}
	return hash, err
}

func (impl GitBitbucketClient) ensureProjectAvailabilityOnSsh(repoOptions *bitbucket.RepositoryOptions, targetRevision string) (bool, error) {
	repoUrl := fmt.Sprintf(BITBUCKET_CLONE_BASE_URL+"%s/%s.git", repoOptions.Owner, repoOptions.RepoSlug)
	for count := 0; count < 5; count++ {
		_, err := impl.gitOpsHelper.Clone(repoUrl, fmt.Sprintf("/ensure-clone/%s", repoOptions.RepoSlug), targetRevision)
		if err == nil {
			impl.logger.Infow("ensureProjectAvailability clone passed Bitbucket", "try count", count, "repoUrl", repoUrl)
			return true, nil
		}
		impl.logger.Errorw("ensureProjectAvailability clone failed ssh Bitbucket", "try count", count, "err", err)
		time.Sleep(10 * time.Second)
	}
	return false, nil
}

func (impl GitBitbucketClient) cleanUp(cloneDir string) {
	err := os.RemoveAll(cloneDir)
	if err != nil {
		impl.logger.Errorw("error cleaning work path for git-ops", "err", err, "cloneDir", cloneDir)
	}
}

func (impl GitBitbucketClient) CommitValues(ctx context.Context, config *ChartConfig, gitOpsConfig *bean2.GitOpsConfigDto, publishStatusConflictError bool) (commitHash string, commitTime time.Time, err error) {

	start := time.Now()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		impl.logger.Errorw("error in getting home dir", "err", err)
		util.TriggerGitOpsMetrics("CommitValues", "GitBitbucketClient", start, err)
		return "", time.Time{}, err
	}
	bitbucketGitOpsDirPath := path.Join(homeDir, BITBUCKET_GITOPS_DIR, config.GetBitBucketBaseDir())
	osErr := os.MkdirAll(bitbucketGitOpsDirPath, os.ModePerm)
	if osErr != nil {
		impl.logger.Errorw("error in creating bitbucket commit base dir", "bitbucketGitOpsDirPath", bitbucketGitOpsDirPath, "err", osErr)
	}
	defer impl.cleanUp(bitbucketGitOpsDirPath)

	bitbucketCommitFilePath := path.Join(bitbucketGitOpsDirPath, config.FileName)
	impl.logger.Debugw("bitbucket commit FilePath", "bitbucketCommitFilePath", bitbucketCommitFilePath)

	err = ioutil.WriteFile(bitbucketCommitFilePath, []byte(config.FileContent), 0666)
	if err != nil {
		util.TriggerGitOpsMetrics("CommitValues", "GitBitbucketClient", start, err)
		impl.logger.Errorw("error in writing bitbucket commit file", "bitbucketCommitFilePath", bitbucketCommitFilePath, "err", err)
		return "", time.Time{}, err
	}
	fileName := filepath.Join(config.ChartLocation, config.FileName)

	branch := config.TargetRevision
	if len(branch) == 0 {
		branch = util.GetDefaultTargetRevision()
	}
	//bitbucket needs author as - "Name <email-Id>"
	authorBitbucket := fmt.Sprintf("%s <%s>", config.UserName, config.UserEmailId)
	repoWriteOptions := &bitbucket.RepositoryBlobWriteOptions{
		Owner:    gitOpsConfig.BitBucketWorkspaceId,
		RepoSlug: config.ChartRepoName,
		FilePath: bitbucketCommitFilePath,
		FileName: fileName,
		Message:  config.ReleaseMessage,
		Branch:   branch,
		Author:   authorBitbucket,
	}
	repoWriteOptions.WithContext(ctx)
	err = impl.client.Repositories.Repository.WriteFileBlob(repoWriteOptions)
	if err != nil {
		impl.logger.Errorw("error in committing file to bitbucket", "repoWriteOptions", repoWriteOptions, "err", err)
		if e := (&bitbucket.UnexpectedResponseStatusError{}); errors.As(err, &e) && strings.Contains(e.Error(), "500 Internal Server Error") {
			if publishStatusConflictError {
				util.TriggerGitOpsMetrics("CommitValues", "GitBitbucketClient", start, err)
			}
			return "", time.Time{}, retryFunc.NewRetryableError(err)
		}
		util.TriggerGitOpsMetrics("CommitValues", "GitBitbucketClient", start, err)
		return "", time.Time{}, err
	}
	commitOptions := &bitbucket.CommitsOptions{
		RepoSlug:    config.ChartRepoName,
		Owner:       gitOpsConfig.BitBucketWorkspaceId,
		Branchortag: config.TargetRevision,
	}
	commits, err := impl.client.Repositories.Commits.GetCommits(commitOptions)
	if err != nil {
		util.TriggerGitOpsMetrics("CommitValues", "GitBitbucketClient", start, err)
		impl.logger.Errorw("error in getting commits from bitbucket", "commitOptions", commitOptions, "err", err)
		return "", time.Time{}, err
	}

	//extracting the latest commit hash from the paginated api response of above method, reference of api & response - https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Bworkspace%7D/%7Brepo_slug%7D/commits
	commitsMap, ok := commits.(map[string]interface{})
	if !ok {
		impl.logger.Errorw("unexpected response format from bitbucket", "commits", commits)
		return "", time.Time{}, fmt.Errorf("unexpected response format from bitbucket")
	}

	values, ok := commitsMap["values"]
	if !ok || values == nil {
		impl.logger.Errorw("no values found in bitbucket response", "commits", commits, "commitsMap", commitsMap)
		return "", time.Time{}, fmt.Errorf("no commits found in bitbucket response")
	}

	valuesArray, ok := values.([]interface{})
	if !ok || len(valuesArray) == 0 {
		impl.logger.Errorw("empty values array in bitbucket response", "commits", commits, "values", values)
		return "", time.Time{}, fmt.Errorf("empty commits array in bitbucket response")
	}

	firstCommit, ok := valuesArray[0].(map[string]interface{})
	if !ok {
		impl.logger.Errorw("invalid commit format in bitbucket response", "commits", commits, "firstCommit", valuesArray[0])
		return "", time.Time{}, fmt.Errorf("invalid commit format in bitbucket response")
	}

	commitHash, ok = firstCommit["hash"].(string)
	if !ok || commitHash == "" {
		impl.logger.Errorw("no hash found in commit", "commits", commits, "firstCommit", firstCommit)
		return "", time.Time{}, fmt.Errorf("no hash found in commit")
	}

	dateStr, ok := firstCommit["date"].(string)
	if !ok || dateStr == "" {
		impl.logger.Errorw("no date found in commit", "firstCommit", firstCommit)
		return "", time.Time{}, fmt.Errorf("no date found in commit response")
	}

	commitTime, err = time.Parse(time.RFC3339, dateStr)
	if err != nil {
		util.TriggerGitOpsMetrics("CommitValues", "GitBitbucketClient", start, err)
		impl.logger.Errorw("error in getting commitTime", "err", err)
		return "", time.Time{}, err
	}
	util.TriggerGitOpsMetrics("CommitValues", "GitBitbucketClient", start, nil)
	return commitHash, commitTime, nil
}
