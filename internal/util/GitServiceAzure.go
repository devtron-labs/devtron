package util

import (
	"context"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"go.uber.org/zap"
	"path/filepath"
	"time"
)

type GitAzureClient struct {
	client                 *git.Client
	logger                 *zap.SugaredLogger
	project                string
	gitService             GitService
	gitOpsConfigRepository repository.GitOpsConfigRepository
}

func (impl GitAzureClient) GetRepoUrl(config *bean2.GitOpsConfigDto) (repoUrl string, err error) {
	url, exists, err := impl.repoExists(config.GitRepoName, impl.project)
	if err != nil {
		return "", err
	} else if !exists {
		return "", fmt.Errorf("%s :repo not found", config.GitRepoName)
	} else {
		return url, nil
	}
}

func NewGitAzureClient(token string, host string, project string, logger *zap.SugaredLogger, gitService GitService,
	gitOpsConfigRepository repository.GitOpsConfigRepository) (GitAzureClient, error) {
	ctx := context.Background()
	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(host, token)
	// Create a client to interact with the Core area
	coreClient, err := git.NewClient(ctx, connection)
	if err != nil {
		logger.Errorw("error in creating azure gitops client, gitops related operation might fail", "err", err)
	}
	return GitAzureClient{
		client:                 &coreClient,
		project:                project,
		logger:                 logger,
		gitService:             gitService,
		gitOpsConfigRepository: gitOpsConfigRepository,
	}, err
}
func (impl GitAzureClient) DeleteRepository(config *bean2.GitOpsConfigDto) error {
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

func (impl GitAzureClient) EnsureRepoAvailableOnSsh(config *bean2.GitOpsConfigDto, repoUrl string) (string, error) {
	validated, err := impl.ensureProjectAvailabilityOnSsh(impl.project, repoUrl)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability in Azure", "project", config.GitRepoName, "err", err)
		return CloneSshStage, err
	}
	if !validated {
		return "unable to validate project", fmt.Errorf("%s in given time", config.GitRepoName)
	}
	return "", nil
}

func (impl GitAzureClient) EnsureRepoAvailableOnHttp(config *bean2.GitOpsConfigDto) (string, error) {
	validated, err := impl.ensureProjectAvailabilityOnHttp(config.GitRepoName)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability github", "project", config.GitRepoName, "err", err)
		return CloneHttpStage, err
	}
	if !validated {
		return "unable to validate project", fmt.Errorf(":%s in given time", config.GitRepoName)
	}
	return "", nil
}

func (impl GitAzureClient) CreateRepository(config *bean2.GitOpsConfigDto) (url string, isNew bool, detailedErrorGitOpsConfigActions DetailedErrorGitOpsConfigActions) {
	detailedErrorGitOpsConfigActions.StageErrorMap = make(map[string]error)
	ctx := context.Background()
	url, repoExists, err := impl.repoExists(config.GitRepoName, impl.project)
	if err != nil {
		impl.logger.Errorw("error in communication with azure", "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[GetRepoUrlStage] = err
		return "", false, detailedErrorGitOpsConfigActions
	}
	if repoExists {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, GetRepoUrlStage)
		return url, false, detailedErrorGitOpsConfigActions
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
		return "", true, detailedErrorGitOpsConfigActions
	}
	logger.Infow("repo created ", "r", operationReference.WebUrl)
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateRepoStage)
	validated, err := impl.ensureProjectAvailabilityOnHttp(config.GitRepoName)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability azure", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = err
		return *operationReference.WebUrl, true, detailedErrorGitOpsConfigActions
	}
	if !validated {
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneHttpStage] = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		return "", true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneHttpStage)

	_, err = impl.CreateReadme(config)
	if err != nil {
		impl.logger.Errorw("error in creating readme azure", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CreateReadmeStage] = err
		return *operationReference.WebUrl, true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CreateReadmeStage)

	validated, err = impl.ensureProjectAvailabilityOnSsh(impl.project, *operationReference.WebUrl)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability azure", "project", config.GitRepoName, "err", err)
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = err
		return *operationReference.WebUrl, true, detailedErrorGitOpsConfigActions
	}
	if !validated {
		detailedErrorGitOpsConfigActions.StageErrorMap[CloneSshStage] = fmt.Errorf("unable to validate project:%s in given time", config.GitRepoName)
		return "", true, detailedErrorGitOpsConfigActions
	}
	detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, CloneSshStage)
	return *operationReference.WebUrl, true, detailedErrorGitOpsConfigActions
}

func (impl GitAzureClient) CreateReadme(config *bean2.GitOpsConfigDto) (string, error) {
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
		impl.logger.Errorw("error in creating readme azure", "repo", config.GitRepoName, "err", err)
	}
	return hash, err
}

func (impl GitAzureClient) CommitValues(config *ChartConfig, gitOpsConfig *bean2.GitOpsConfigDto) (commitHash string, commitTime time.Time, err error) {
	branch := "master"
	branchfull := "refs/heads/master"
	path := filepath.Join(config.ChartLocation, config.FileName)
	ctx := context.Background()
	newFile := true
	oldObjId := "0000000000000000000000000000000000000000" //default commit hash
	// check if file exists and current hash
	// if file does not exists get hash from branch
	// if branch doesn't exists use default hash
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
		Name:        &branchfull,
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

	if err != nil {
		impl.logger.Errorw("error in commit azure", "err", err)
		return "", time.Time{}, err
	}
	//gitPush.Commits
	commitId := ""
	commitAuthorTime := time.Time{}
	if len(*push.Commits) > 0 {
		commitId = *(*push.Commits)[0].CommitId
		commitAuthorTime = (*push.Commits)[0].Author.Date.Time
	}
	//	push.Commits[0].CommitId
	return commitId, commitAuthorTime, nil
}

func (impl GitAzureClient) repoExists(repoName, projectName string) (repoUrl string, exists bool, err error) {
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
			return "", false, nil
		} else {
			return "", false, err
		}

	}
	for gitRepository == nil {
		return "", false, nil
	}
	return *gitRepository.WebUrl, true, nil
}

func (impl GitAzureClient) ensureProjectAvailabilityOnHttp(repoName string) (bool, error) {
	for count := 0; count < 5; count++ {
		_, exists, err := impl.repoExists(repoName, impl.project)
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

func (impl GitAzureClient) ensureProjectAvailabilityOnSsh(projectName string, repoUrl string) (bool, error) {
	for count := 0; count < 8; count++ {
		_, err := impl.gitService.Clone(repoUrl, fmt.Sprintf("/ensure-clone/%s", projectName))
		if err == nil {
			impl.logger.Infow("ensureProjectAvailability clone passed azure", "try count", count, "repoUrl", repoUrl)
			return true, nil
		}
		impl.logger.Errorw("ensureProjectAvailability clone failed ssh azure", "try count", count, "err", err)
		time.Sleep(10 * time.Second)
	}
	return false, nil
}

func (impl GitAzureClient) GetCommits(repoName, projectName string) ([]*GitCommitDto, error) {
	azureClient := *impl.client
	getCommitsArgs := git.GetCommitsArgs{
		RepositoryId: &repoName,
		Project:      &projectName,
	}
	gitCommits, err := azureClient.GetCommits(context.Background(), getCommitsArgs)
	if err != nil {
		impl.logger.Errorw("error in getting commits", "err", err, "repoName", repoName, "projectName", projectName)
		return nil, err
	}
	var gitCommitsDto []*GitCommitDto
	for _, gitCommit := range *gitCommits {
		gitCommitDto := &GitCommitDto{
			CommitHash: *gitCommit.CommitId,
			AuthorName: *gitCommit.Author.Name,
			CommitTime: gitCommit.Author.Date.Time,
		}
		gitCommitsDto = append(gitCommitsDto, gitCommitDto)
	}
	return gitCommitsDto, nil
}

func (impl GitAzureClient) GetCommitsCount(repoName, projectName string) (int, error) {
	azureClient := *impl.client
	getCommitsArgs := git.GetCommitsArgs{
		RepositoryId:   &repoName,
		Project:        &projectName,
		SearchCriteria: &git.GitQueryCommitsCriteria{},
	}
	gitCommits, err := azureClient.GetCommits(context.Background(), getCommitsArgs)
	if err != nil {
		impl.logger.Errorw("error in getting commits", "err", err, "repoName", repoName, "projectName", projectName)
		return 0, err
	}
	return len(*gitCommits), nil
}
