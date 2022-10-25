package util

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/go-pg/pg"
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
	client *github.Client

	//config *GitConfig
	logger                 *zap.SugaredLogger
	org                    string
	gitService             GitService
	gitOpsConfigRepository repository.GitOpsConfigRepository
}

func NewGithubClient(host string, token string, org string, logger *zap.SugaredLogger,
	gitService GitService, gitOpsConfigRepository repository.GitOpsConfigRepository) (GitHubClient, error) {
	ctx := context.Background()
	httpTransport := &http2.Transport{}
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
		client:                 client,
		org:                    org,
		logger:                 logger,
		gitService:             gitService,
		gitOpsConfigRepository: gitOpsConfigRepository,
	}, err
}
func (impl GitHubClient) DeleteRepository(name string) error {
	gitOpsConfig, err := impl.gitOpsConfigRepository.GetGitOpsConfigByProvider(GITHUB_PROVIDER)
	if err != nil {
		if err == pg.ErrNoRows {
			gitOpsConfig.GitHubOrgId = ""
		} else {
			impl.logger.Errorw("error in fetching gitOps github config", "err", err)
			return err
		}
	}
	_, err = impl.client.Repositories.Delete(context.Background(), gitOpsConfig.GitHubOrgId, name)
	if err != nil {
		impl.logger.Errorw("repo deletion failed for github", "repo", name, "err", err)
		return err
	}
	return nil
}
func (impl GitHubClient) CreateRepository(name, description, userName, userEmailId string) (url string, isNew bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions) {
	detailedErrorGitOpsConfigActions.StageErrorMap = make(map[string]error)
	ctx := context.Background()
	repoExists := true
	url, err := impl.GetRepoUrl(name)
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
		&github.Repository{Name: &name,
			Description: &description,
			Private:     &private,
			//			Visibility:  &visibility,
		})
	if err != nil {
		impl.logger.Errorw("error in creating github repo, ", "repo", name, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateRepoStage] = err
		return "", true, detailedErrorGitOpsConfigActions
	}
	logger.Infow("github repo created ", "r", r.CloneURL)
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateRepoStage)

	validated, err := impl.ensureProjectAvailabilityOnHttp(name)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability github", "project", name, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = err
		return *r.CloneURL, true, detailedErrorGitOpsConfigActions
	}
	if !validated {
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = fmt.Errorf("unable to validate project:%s in given time", name)
		return "", true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneHttpStage)

	_, err = impl.CreateReadme(name, userName, userEmailId)
	if err != nil {
		impl.logger.Errorw("error in creating readme github", "project", name, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateReadmeStage] = err
		return *r.CloneURL, true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateReadmeStage)

	validated, err = impl.ensureProjectAvailabilityOnSsh(name, *r.CloneURL)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability github", "project", name, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = err
		return *r.CloneURL, true, detailedErrorGitOpsConfigActions
	}
	if !validated {
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = fmt.Errorf("unable to validate project:%s in given time", name)
		return "", true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneSshStage)
	//_, err = impl.createReadme(name)
	return *r.CloneURL, true, detailedErrorGitOpsConfigActions
}

func (impl GitHubClient) CreateReadme(repoName, userName, userEmailId string) (string, error) {
	cfg := &ChartConfig{
		ChartName:      repoName,
		ChartLocation:  "",
		FileName:       "README.md",
		FileContent:    "@devtron",
		ReleaseMessage: "readme",
		ChartRepoName:  repoName,
		UserName:       userName,
		UserEmailId:    userEmailId,
	}
	hash, _, err := impl.CommitValues(cfg)
	if err != nil {
		impl.logger.Errorw("error in creating readme github", "repo", repoName, "err", err)
	}
	return hash, err
}

func (impl GitHubClient) CommitValues(config *ChartConfig) (commitHash string, commitTime time.Time, err error) {
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
	return *c.SHA, *c.Commit.Author.Date, nil
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
			impl.logger.Errorw("error in validating repo github", "project", projectName, "err", err)
			return false, err
		} else {
			impl.logger.Errorw("error in validating repo github", "project", projectName, "err", err)
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

func (impl GitHubClient) GetCommits(repoName, projectName string) ([]*GitCommitDto, error) {
	githubClient := impl.client
	gitCommits, _, err := githubClient.Repositories.ListCommits(context.Background(), impl.org, repoName, &github.CommitsListOptions{})
	if err != nil {
		impl.logger.Errorw("error in getting commits", "err", err, "repoName", repoName)
		return nil, err
	}
	var gitCommitsDto []*GitCommitDto
	for _, gitCommit := range gitCommits {
		gitCommitDto := &GitCommitDto{
			CommitHash: gitCommit.GetSHA(),
			AuthorName: *gitCommit.Author.Name,
			CommitTime: *gitCommit.Commit.Author.Date,
		}
		gitCommitsDto = append(gitCommitsDto, gitCommitDto)
	}
	return gitCommitsDto, nil
}
