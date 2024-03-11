package git

import (
	"context"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean/gitOps"
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
	gitOpsHelper *GitOpsHelper) (GitHubClient, error) {
	ctx := context.Background()
	httpTransport := &http2.Transport{Proxy: http2.ProxyFromEnvironment}
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
	_, err := impl.client.Repositories.Delete(context.Background(), config.GitHubOrgId, config.GitRepoName)
	if err != nil {
		impl.logger.Errorw("repo deletion failed for github", "repo", config.GitRepoName, "err", err)
		return err
	}
	return nil
}

func (impl GitHubClient) CreateRepository(config *bean2.GitOpsConfigDto) (url string, isNew bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions) {
	detailedErrorGitOpsConfigActions.StageErrorMap = make(map[string]error)
	ctx := context.Background()
	repoExists := true
	url, err := impl.GetRepoUrl(config)
	if err != nil {
		responseErr, ok := err.(*github.ErrorResponse)
		if !ok || responseErr.Response.StatusCode != 404 {
			impl.logger.Errorw("error in creating github repo", "err", err)
			detailedErrorGitOpsConfigActions.StageErrorMap[GetRepoUrlStage] = err
			return "", false, detailedErrorGitOpsConfigActions
		} else {
			repoExists = false
		}
	}
	if repoExists {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, GetRepoUrlStage)
		return url, false, detailedErrorGitOpsConfigActions
	}
	private := true
	//	visibility := "private"
	r, _, err := impl.client.Repositories.Create(ctx, impl.org,
		&github.Repository{Name: &config.GitRepoName,
			Description: &config.Description,
			Private:     &private,
			//			Visibility:  &visibility,
		})
	if err != nil {
		impl.logger.Errorw("error in creating github repo, ", "repo", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateRepoStage] = err
		return "", true, detailedErrorGitOpsConfigActions
	}
	impl.logger.Infow("github repo created ", "r", r.CloneURL)
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateRepoStage)

	validated, err := impl.ensureProjectAvailabilityOnHttp(config)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability github", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = err
		return *r.CloneURL, true, detailedErrorGitOpsConfigActions
	}
	if !validated {
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		return "", true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneHttpStage)

	_, err = impl.CreateReadme(config)
	if err != nil {
		impl.logger.Errorw("error in creating readme github", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateReadmeStage] = err
		return *r.CloneURL, true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateReadmeStage)

	validated, err = impl.ensureProjectAvailabilityOnSsh(config.GitRepoName, *r.CloneURL)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability github", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = err
		return *r.CloneURL, true, detailedErrorGitOpsConfigActions
	}
	if !validated {
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		return "", true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneSshStage)
	//_, err = impl.createReadme(name)
	return *r.CloneURL, true, detailedErrorGitOpsConfigActions
}

func (impl GitHubClient) CreateReadme(config *bean2.GitOpsConfigDto) (string, error) {
	cfg := &ChartConfig{
		ChartName:      config.GitRepoName,
		ChartLocation:  "",
		FileName:       "README.md",
		FileContent:    "@devtron",
		ReleaseMessage: "readme",
		ChartRepoName:  config.GitRepoName,
		UserName:       config.Username,
		UserEmailId:    config.UserEmailId,
	}
	hash, _, err := impl.CommitValues(cfg, config)
	if err != nil {
		impl.logger.Errorw("error in creating readme github", "repo", config.GitRepoName, "err", err)
	}
	return hash, err
}

func (impl GitHubClient) CommitValues(config *ChartConfig, gitOpsConfig *bean2.GitOpsConfigDto) (commitHash string, commitTime time.Time, err error) {
	branch := "master"
	path := filepath.Join(config.ChartLocation, config.FileName)
	ctx := context.Background()
	newFile := false
	fc, _, _, err := impl.client.Repositories.GetContents(ctx, impl.org, config.ChartRepoName, path, &github.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		responseErr, ok := err.(*github.ErrorResponse)
		if !ok || responseErr.Response.StatusCode != 404 {
			impl.logger.Errorw("error in creating repo github", "err", err, "config", config)
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
	c, _, err := impl.client.Repositories.CreateFile(ctx, impl.org, config.ChartRepoName, path, options)
	if err != nil {
		impl.logger.Errorw("error in commit github", "err", err, "config", config)
		return "", time.Time{}, err
	}
	commitTime = time.Now() // default is current time, if found then will get updated accordingly
	if c != nil && c.Commit.Author != nil {
		commitTime = *c.Commit.Author.Date
	}
	return *c.SHA, commitTime, nil
}

func (impl GitHubClient) GetRepoUrl(config *bean2.GitOpsConfigDto) (repoUrl string, err error) {
	ctx := context.Background()
	repo, _, err := impl.client.Repositories.Get(ctx, impl.org, config.GitRepoName)
	if err != nil {
		return "", err
	}
	return *repo.CloneURL, nil
}

func (impl GitHubClient) ensureProjectAvailabilityOnHttp(config *bean2.GitOpsConfigDto) (bool, error) {
	count := 0
	for count < 3 {
		count = count + 1
		_, err := impl.GetRepoUrl(config)
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

func (impl GitHubClient) ensureProjectAvailabilityOnSsh(projectName string, repoUrl string) (bool, error) {
	count := 0
	for count < 3 {
		count = count + 1
		_, err := impl.gitOpsHelper.Clone(repoUrl, fmt.Sprintf("/ensure-clone/%s", projectName))
		if err == nil {
			impl.logger.Infow("github ensureProjectAvailability clone passed", "try count", count, "repoUrl", repoUrl)
			return true, nil
		}
		if err != nil {
			impl.logger.Errorw("github ensureProjectAvailability clone failed", "try count", count, "err", err)
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}
