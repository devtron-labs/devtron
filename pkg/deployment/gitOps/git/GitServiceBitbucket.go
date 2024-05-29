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
	"encoding/base64"
	"fmt"
	"github.com/antihax/optional"
	"github.com/devtron-labs/bitbucketdc-gosdk/swagger"
	bean2 "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/ktrysmt/go-bitbucket"
	"go.uber.org/zap"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
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
	dcClient     *swagger.APIClient
	client       *bitbucket.Client
	logger       *zap.SugaredLogger
	gitOpsHelper *GitOpsHelper

	//for bit bucket dc
	DcHost    string
	ProjectId string
}

func (client GitBitbucketClient) GetRepoUrlDc(repoSlug string) string {
	return GetBitBucketDcRepoUrl(client.ProjectId, repoSlug, client.DcHost)
}

func NewGitBitbucketDcClient(username, token, host string, project string, logger *zap.SugaredLogger, gitOpsHelper *GitOpsHelper) GitBitbucketClient {

	usePat := UsePatAuth()

	apiPath, _ := url.JoinPath(host, "rest")

	headers := make(map[string]string)
	headers["referer"] = host

	if usePat {
		headers["Authorization"] = "Bearer " + token
	} else {
		basicToken := username + ":" + token
		enc := base64.StdEncoding.EncodeToString([]byte(basicToken))
		headers["Authorization"] = "Basic " + enc
	}

	dcClient := swagger.NewAPIClient(&swagger.Configuration{
		BasePath:      apiPath,
		Host:          "",
		Scheme:        "https",
		DefaultHeader: headers,
		UserAgent:     "",
		HTTPClient:    nil,
	})
	logger.Infow("bitbucket client created", "clientDetails", dcClient)
	return GitBitbucketClient{
		dcClient: dcClient,
		//client:       coreClient,
		logger:       logger,
		gitOpsHelper: gitOpsHelper,
		DcHost:       strings.TrimRight(host, "/"),
		ProjectId:    project,
	}
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
	var err error
	if impl.dcClient == nil {
		repoOptions := &bitbucket.RepositoryOptions{
			Owner:     config.BitBucketWorkspaceId,
			RepoSlug:  config.GitRepoName,
			IsPrivate: "true",
			Project:   config.BitBucketProjectKey,
		}
		_, err = impl.client.Repositories.Repository.Delete(repoOptions)
	} else {
		_, err = impl.dcClient.ProjectApi.DeleteRepository(context.Background(), config.BitBucketProjectKey, config.GitRepoName)
	}

	if err != nil {
		impl.logger.Errorw("error in deleting repo gitlab", "repoName", config.GitRepoName, "err", err)
	}
	return err
}

func (impl GitBitbucketClient) GetRepoUrl(config *bean2.GitOpsConfigDto) (repoUrl string, err error) {
	repoOptions := &bitbucket.RepositoryOptions{
		Owner:    config.BitBucketWorkspaceId,
		Project:  config.BitBucketProjectKey,
		RepoSlug: config.GitRepoName,
	}
	repoUrl, exists, err := impl.repoExists(repoOptions, impl.DcHost)
	if err != nil {
		return "", err
	} else if !exists {
		return "", fmt.Errorf("%s :repo not found", repoOptions.RepoSlug)
	}
	return repoUrl, nil

}

func (impl GitBitbucketClient) CreateRepository(config *bean2.GitOpsConfigDto) (url string, isNew bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions) {
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
	repoUrl, repoExists, err := impl.repoExists(repoOptions, impl.DcHost)
	if err != nil {
		impl.logger.Errorw("error in communication with bitbucket", "repoOptions", repoOptions, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[GetRepoUrlStage] = err
		return "", false, detailedErrorGitOpsConfigActions
	}
	if repoExists {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, GetRepoUrlStage)
		return repoUrl, false, detailedErrorGitOpsConfigActions
	}

	if impl.dcClient != nil {
		// {"is_private":true,"name":"devtron-sample-repo-dryrun-avbkds","project":{"key":"TPROJ"},"scm":"git"}
		repository, h, err := impl.dcClient.ProjectApi.CreateRepository(context.Background(),
			repoOptions.Project, &swagger.ProjectApiCreateRepositoryOpts{Body: optional.NewInterface(
				swagger.RestRepository{
					Name:    repoOptions.RepoSlug,
					Public:  false,
					Project: &swagger.RestPullRequestFromRefRepositoryProject{Key: repoOptions.Project},
					Slug:    repoOptions.RepoSlug,
					ScmId:   "git",
				})})
		if err != nil {
			fmt.Println(repository, h, err)
			return "", false, DetailedErrorGitOpsConfigActions{}
		}
	} else {
		_, err = impl.client.Repositories.Repository.Create(repoOptions)
		repoUrl = fmt.Sprintf(BITBUCKET_CLONE_BASE_URL+"%s/%s.git", repoOptions.Owner, repoOptions.RepoSlug)
	}
	if err != nil {
		impl.logger.Errorw("error in creating repo bitbucket", "repoOptions", repoOptions, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateRepoStage] = err
		return "", true, detailedErrorGitOpsConfigActions
	}

	impl.logger.Infow("repo created ", "repoUrl", repoUrl)
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateRepoStage)

	validated, repoUrl, err := impl.ensureProjectAvailabilityOnHttp(repoOptions, config.Host)
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

	_, err = impl.CreateReadme(config)
	if err != nil {
		impl.logger.Errorw("error in creating readme bitbucket", "repoName", repoOptions.RepoSlug, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateReadmeStage] = err
		return "", true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateReadmeStage)

	validated, err = impl.ensureProjectAvailabilityOnSsh(repoOptions, repoUrl)
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

func (impl GitBitbucketClient) repoExists(repoOptions *bitbucket.RepositoryOptions, dcURL string) (repoUrl string, exists bool, err error) {
	var repo *bitbucket.Repository

	if impl.dcClient != nil {
		repository, h, err := impl.dcClient.ProjectApi.GetRepository(context.Background(), repoOptions.Project, repoOptions.RepoSlug)
		if h.Status == BITBUCKET_REPO_NOT_FOUND_ERROR {
			return "", false, nil
		}
		if err != nil {
			impl.logger.Infow("test", "r", repository, "h", h)
			return "", false, err
		}

		if err != nil {
			return "", false, err
		}
		// git clone https://bitbucket-cloud.devtron.ai/scm/tes/devtron-sample-repo-dryrun-qdyxo9.git

		repoUrl = GetBitBucketDcRepoUrl(repoOptions.Project, repoOptions.RepoSlug, dcURL)
	} else {
		repo, err = impl.client.Repositories.Repository.Get(repoOptions)

		if repo == nil && err.Error() == BITBUCKET_REPO_NOT_FOUND_ERROR {
			return "", false, nil
		}
		if err != nil {
			return "", false, err
		}
		repoUrl = fmt.Sprintf(BITBUCKET_CLONE_BASE_URL+"%s/%s.git", repoOptions.Owner, repoOptions.RepoSlug)
	}
	return repoUrl, true, nil
}

func GetBitBucketDcRepoUrl(project string, repoSlug string, dcURL string) string {
	return fmt.Sprintf(dcURL+"/scm/%s/%s.git", project, repoSlug)
}
func (impl GitBitbucketClient) ensureProjectAvailabilityOnHttp(repoOptions *bitbucket.RepositoryOptions, dcHostURL string) (bool, string, error) {
	for count := 0; count < 5; count++ {
		gitRepoUrl, exists, err := impl.repoExists(repoOptions, impl.DcHost)
		if err == nil && exists {
			impl.logger.Infow("repo validated successfully on https")
			return true, gitRepoUrl, nil
		} else if err != nil {
			impl.logger.Errorw("error in validating repo bitbucket", "repoDetails", repoOptions, "err", err)
			return false, gitRepoUrl, err
		} else {
			impl.logger.Errorw("repo not available on http", "repoDetails", repoOptions)
		}
		time.Sleep(10 * time.Second)
	}
	return false, "", nil
}

func (impl GitBitbucketClient) CreateReadme(config *bean2.GitOpsConfigDto) (string, error) {
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
	hash, _, err := impl.CommitValues(cfg, config)
	if err != nil {
		impl.logger.Errorw("error in creating readme bitbucket", "repo", config.GitRepoName, "err", err)
	}
	return hash, err
}

func (impl GitBitbucketClient) ensureProjectAvailabilityOnSsh(repoOptions *bitbucket.RepositoryOptions, repoUrl string) (bool, error) {
	//repoUrl := fmt.Sprintf(BITBUCKET_CLONE_BASE_URL+"%s/%s.git", repoOptions.Owner, repoOptions.RepoSlug)
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

func (impl GitBitbucketClient) CommitValues(config *ChartConfig, gitOpsConfig *bean2.GitOpsConfigDto) (commitHash string, commitTime time.Time, err error) {

	var cloneDir string
	var bitbucketCommitFilePath string
	fileName := filepath.Join(config.ChartLocation, config.FileName)
	//TODO need to refactor this
	if impl.dcClient != nil {
		//url := strings.Replace(config.RepoUrl, "https://", "https://"+impl.gitOpsHelper.Auth.Username+":"+impl.gitOpsHelper.Auth.Password+"@", -1)
		repoUrl := impl.GetRepoUrlDc(config.ChartRepoName)
		cloneDir, err = impl.gitOpsHelper.Clone(repoUrl, BITBUCKET_GITOPS_DIR)
		if err != nil {
			return "", time.Time{}, err
		}
		bitbucketCommitFilePath = path.Join(cloneDir, fileName)
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", time.Time{}, err
		}
		bitbucketGitOpsDirPath := path.Join(homeDir, BITBUCKET_GITOPS_DIR)
		if _, err = os.Stat(bitbucketGitOpsDirPath); !os.IsExist(err) {
			os.Mkdir(bitbucketGitOpsDirPath, 0777)
		}

		bitbucketCommitFilePath = path.Join(bitbucketGitOpsDirPath, config.FileName)

		if _, err = os.Stat(bitbucketCommitFilePath); os.IsExist(err) {
			os.Remove(bitbucketCommitFilePath)
		}
	}

	err = ioutil.WriteFile(bitbucketCommitFilePath, []byte(config.FileContent), 0666)
	if err != nil {
		return "", time.Time{}, err
	}

	//bitbucket needs author as - "Name <email-Id>"
	authorBitbucket := fmt.Sprintf("%s <%s>", config.UserName, config.UserEmailId)

	// commit readme file
	if impl.dcClient == nil {
		repoWriteOptions := &bitbucket.RepositoryBlobWriteOptions{
			Owner:    gitOpsConfig.BitBucketWorkspaceId,
			RepoSlug: config.ChartRepoName,
			FilePath: bitbucketCommitFilePath,
			FileName: fileName,
			Message:  config.ReleaseMessage,
			Branch:   "master",
			Author:   authorBitbucket,
		}
		err = impl.client.Repositories.Repository.WriteFileBlob(repoWriteOptions)
	} else {

		commitHash, err = impl.gitOpsHelper.CommitAndPushAllChanges(cloneDir, config.ReleaseMessage, authorBitbucket, config.UserEmailId)
	}

	_ = os.Remove(bitbucketCommitFilePath)
	if err != nil {
		return "", time.Time{}, err
	}

	//get latest commit hash and time
	if impl.dcClient == nil {
		commitOptions := &bitbucket.CommitsOptions{
			RepoSlug:    config.ChartRepoName,
			Owner:       gitOpsConfig.BitBucketWorkspaceId,
			Branchortag: "master",
		}
		commits, err := impl.client.Repositories.Commits.GetCommits(commitOptions)
		if err != nil {
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

	}

	return commitHash, commitTime, nil
}
