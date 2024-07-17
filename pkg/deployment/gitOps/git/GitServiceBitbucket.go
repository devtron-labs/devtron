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
	"errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean/gitOps"
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

func NewGitBitbucketClient(username, token, host string, logger *zap.SugaredLogger, gitOpsHelper *GitOpsHelper) GitBitbucketClient {
	coreClient := bitbucket.NewBasicAuth(username, token)
	logger.Infow("bitbucket client created", "clientDetails", coreClient)
	return GitBitbucketClient{
		client:       coreClient,
		logger:       logger,
		gitOpsHelper: gitOpsHelper,
	}
}

func (impl GitBitbucketClient) DeleteRepository(config *bean2.GitOpsConfigDto) error {
	repoOptions := &bitbucket.RepositoryOptions{
		Owner:     config.BitBucketWorkspaceId,
		RepoSlug:  config.GitRepoName,
		IsPrivate: "true",
		Project:   config.BitBucketProjectKey,
	}
	_, err := impl.client.Repositories.Repository.Delete(repoOptions)
	if err != nil {
		impl.logger.Errorw("error in deleting repo gitlab", "repoName", repoOptions.RepoSlug, "err", err)
	}
	return err
}

func (impl GitBitbucketClient) GetRepoUrl(config *bean2.GitOpsConfigDto) (repoUrl string, err error) {
	repoOptions := &bitbucket.RepositoryOptions{
		Owner:    config.BitBucketWorkspaceId,
		Project:  config.BitBucketProjectKey,
		RepoSlug: config.GitRepoName,
	}
	_, exists, err := impl.repoExists(repoOptions)
	if err != nil {
		return "", err
	} else if !exists {
		return "", fmt.Errorf("%s :repo not found", repoOptions.RepoSlug)
	} else {
		repoUrl = fmt.Sprintf(BITBUCKET_CLONE_BASE_URL+"%s/%s.git", repoOptions.Owner, repoOptions.RepoSlug)
		return repoUrl, nil
	}
}

func (impl GitBitbucketClient) CreateRepository(ctx context.Context, config *bean2.GitOpsConfigDto) (url string, isNew bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions) {
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
		return "", false, detailedErrorGitOpsConfigActions
	}
	if repoExists {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, GetRepoUrlStage)
		return repoUrl, false, detailedErrorGitOpsConfigActions
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
			return "", true, detailedErrorGitOpsConfigActions
		}
	}
	repoUrl = fmt.Sprintf(BITBUCKET_CLONE_BASE_URL+"%s/%s.git", repoOptions.Owner, repoOptions.RepoSlug)
	impl.logger.Infow("repo created ", "repoUrl", repoUrl)
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateRepoStage)

	validated, err := impl.ensureProjectAvailabilityOnHttp(repoOptions)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability bitbucket", "repoName", repoOptions.RepoSlug, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = err
		return "", true, detailedErrorGitOpsConfigActions
	}
	if !validated {
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		return "", true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneHttpStage)

	_, err = impl.CreateReadme(ctx, config)
	if err != nil {
		impl.logger.Errorw("error in creating readme bitbucket", "repoName", repoOptions.RepoSlug, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateReadmeStage] = err
		return "", true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateReadmeStage)

	validated, err = impl.ensureProjectAvailabilityOnSsh(repoOptions)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability bitbucket", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = err
		return "", true, detailedErrorGitOpsConfigActions
	}
	if !validated {
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		return "", true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneSshStage)
	return repoUrl, true, detailedErrorGitOpsConfigActions
}

func (impl GitBitbucketClient) repoExists(repoOptions *bitbucket.RepositoryOptions) (repoUrl string, exists bool, err error) {
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
	cfg := &ChartConfig{
		ChartName:      config.GitRepoName,
		ChartLocation:  "",
		FileName:       "README.md",
		FileContent:    "@devtron",
		ReleaseMessage: "pushing readme",
		ChartRepoName:  config.GitRepoName,
		UserName:       config.Username,
		UserEmailId:    config.UserEmailId,
	}
	cfg.SetBitBucketBaseDir(getDir())
	hash, _, err := impl.CommitValues(ctx, cfg, config)
	if err != nil {
		impl.logger.Errorw("error in creating readme bitbucket", "repo", config.GitRepoName, "err", err)
	}
	return hash, err
}

func (impl GitBitbucketClient) ensureProjectAvailabilityOnSsh(repoOptions *bitbucket.RepositoryOptions) (bool, error) {
	repoUrl := fmt.Sprintf(BITBUCKET_CLONE_BASE_URL+"%s/%s.git", repoOptions.Owner, repoOptions.RepoSlug)
	for count := 0; count < 5; count++ {
		_, err := impl.gitOpsHelper.Clone(repoUrl, fmt.Sprintf("/ensure-clone/%s", repoOptions.RepoSlug))
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

func (impl GitBitbucketClient) CommitValues(ctx context.Context, config *ChartConfig, gitOpsConfig *bean2.GitOpsConfigDto) (commitHash string, commitTime time.Time, err error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		impl.logger.Errorw("error in getting home dir", "err", err)
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
		impl.logger.Errorw("error in writing bitbucket commit file", "bitbucketCommitFilePath", bitbucketCommitFilePath, "err", err)
		return "", time.Time{}, err
	}
	fileName := filepath.Join(config.ChartLocation, config.FileName)

	//bitbucket needs author as - "Name <email-Id>"
	authorBitbucket := fmt.Sprintf("%s <%s>", config.UserName, config.UserEmailId)
	repoWriteOptions := &bitbucket.RepositoryBlobWriteOptions{
		Owner:    gitOpsConfig.BitBucketWorkspaceId,
		RepoSlug: config.ChartRepoName,
		FilePath: bitbucketCommitFilePath,
		FileName: fileName,
		Message:  config.ReleaseMessage,
		Branch:   "master",
		Author:   authorBitbucket,
	}
	repoWriteOptions.WithContext(ctx)
	err = impl.client.Repositories.Repository.WriteFileBlob(repoWriteOptions)
	if err != nil {
		impl.logger.Errorw("error in committing file to bitbucket", "repoWriteOptions", repoWriteOptions, "err", err)
		if e := (&bitbucket.UnexpectedResponseStatusError{}); errors.As(err, &e) && strings.Contains(e.Error(), "500 Internal Server Error") {
			return "", time.Time{}, retryFunc.NewRetryableError(err)
		}
		return "", time.Time{}, err
	}
	commitOptions := &bitbucket.CommitsOptions{
		RepoSlug:    config.ChartRepoName,
		Owner:       gitOpsConfig.BitBucketWorkspaceId,
		Branchortag: "master",
	}
	commits, err := impl.client.Repositories.Commits.GetCommits(commitOptions)
	if err != nil {
		impl.logger.Errorw("error in getting commits from bitbucket", "commitOptions", commitOptions, "err", err)
		return "", time.Time{}, err
	}

	//extracting the latest commit hash from the paginated api response of above method, reference of api & response - https://developer.atlassian.com/bitbucket/api/2/reference/resource/repositories/%7Bworkspace%7D/%7Brepo_slug%7D/commits
	commitHash = commits.(map[string]interface{})["values"].([]interface{})[0].(map[string]interface{})["hash"].(string)
	commitTimeString := commits.(map[string]interface{})["values"].([]interface{})[0].(map[string]interface{})["date"].(string)
	commitTime, err = time.Parse(time.RFC3339, commitTimeString)
	if err != nil {
		impl.logger.Errorw("error in getting commitTime", "err", err)
		return "", time.Time{}, err
	}
	return commitHash, commitTime, nil
}
