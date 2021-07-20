package util

import (
	"context"
	"fmt"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"go.uber.org/zap"
	"path/filepath"
	"time"
)

type GitAzureClient struct {
	client     git.Client
	logger     *zap.SugaredLogger
	project    string
	gitService GitService
}

func (impl GitAzureClient) GetRepoUrl(repoName string) (repoUrl string, err error) {
	url, exists, err := impl.repoExists(repoName, impl.project)
	if err != nil {
		return "", err
	} else if !exists {
		return "", fmt.Errorf("%s :repo not found", repoName)
	} else {
		return url, nil
	}
}

func NewGitAzureClient(token string, host string, project string, logger *zap.SugaredLogger, gitService GitService) GitAzureClient {
	ctx := context.Background()
	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(host, token)
	// Create a client to interact with the Core area
	coreClient, err := git.NewClient(ctx, connection)
	if err != nil {
		logger.Errorw("error in creating azure  gitops client, gitops related operation might fail", "err", err)
	}
	return GitAzureClient{client: coreClient, project: project, logger: logger, gitService: gitService}
}

func (impl GitAzureClient) CreateRepository(name, description string) (url string, isNew bool, err error) {
	ctx := context.Background()
	url, repoExists, err := impl.repoExists(name, impl.project)
	if err != nil {
		impl.logger.Errorw("error in communication with azure", "err", err)
		return "", false, err
	}
	if repoExists {
		return url, false, nil
	}
	gitRepositoryCreateOptions := git.GitRepositoryCreateOptions{
		Name: &name,
	}
	operationReference, err := impl.client.CreateRepository(ctx, git.CreateRepositoryArgs{
		GitRepositoryToCreate: &gitRepositoryCreateOptions,
		Project:               &impl.project,
	})
	if err != nil {
		impl.logger.Errorw("error in creating repo, ", "repo", name, "err", err)
		return "", true, err
	}
	logger.Infow("repo created ", "r", operationReference.WebUrl)

	validated, err := impl.ensureProjectAvailabilityOnHttp(name)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", name, "err", err)
		return *operationReference.WebUrl, true, err
	}
	if !validated {
		return "", true, fmt.Errorf("unable to validate project:%s  in given time", name)
	}
	_, err = impl.createReadme(name)
	if err != nil {
		impl.logger.Errorw("error in creating readme", "err", err)
		return *operationReference.WebUrl, true, err
	}
	validated, err = impl.ensureProjectAvailabilityOnSsh(impl.project, name, *operationReference.WebUrl)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", name, "err", err)
		return *operationReference.WebUrl, true, err
	}
	if !validated {
		return "", true, fmt.Errorf("unable to validate project:%s  in given time", name)
	}
	return *operationReference.WebUrl, true, err
}

func (impl GitAzureClient) createReadme(repoName string) (string, error) {
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

func (impl GitAzureClient) CommitValues(config *ChartConfig) (commitHash string, err error) {
	branch := "master"
	branchfull := "refs/heads/master"
	path := filepath.Join(config.ChartLocation, config.FileName)
	ctx := context.Background()
	newFile := true
	oldObjId := "0000000000000000000000000000000000000000" //default commit hash
	// check if file exists and current hash
	// if file does not exists get hash from branch
	// if branch doesn't exists use default hash
	fc, err := impl.client.GetItem(ctx, git.GetItemArgs{
		RepositoryId: &config.ChartName,
		Path:         &path,
		Project:      &impl.project,
	})
	if err != nil {
		notFoundStatus := 404
		if e, ok := err.(azuredevops.WrappedError); ok && *e.StatusCode == notFoundStatus {
			branchStat, err := impl.client.GetBranch(ctx, git.GetBranchArgs{Project: &impl.project, Name: &branch, RepositoryId: &config.ChartName})
			if err != nil {
				if e, ok := err.(azuredevops.WrappedError); !ok || *e.StatusCode >= 500 {
					impl.logger.Errorw("error in fetching branch from azure devops", "err", err)
					return "", err
				}
			} else if branchStat != nil {
				oldObjId = *branchStat.Commit.CommitId
			}
		} else {
			impl.logger.Errorw("error in fetching file from azure devops", "err", err)
			return "", err
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
	})

	push, err := impl.client.CreatePush(ctx, git.CreatePushArgs{
		Push: &git.GitPush{
			Commits:    &commits,
			RefUpdates: &refUpdates,
		},
		RepositoryId: &config.ChartName,
		Project:      &impl.project,
	})

	if err != nil {
		impl.logger.Errorw("error in commit", "err", err)
		return "", err
	}
	//gitPush.Commits
	commitId := ""
	if len(*push.Commits) > 0 {
		commitId = *(*push.Commits)[0].CommitId
	}
	//	push.Commits[0].CommitId
	return commitId, nil
}

func (impl GitAzureClient) repoExists(repoName, projectName string) (repoUrl string, exists bool, err error) {
	ctx := context.Background()
	// Get first page of the list of team projects for your organization
	gitRepository, err := impl.client.GetRepository(ctx, git.GetRepositoryArgs{
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
			impl.logger.Errorw("error in validating repo", "err", err)
			return false, err
		} else {
			impl.logger.Errorw("repo not available on http", "repo")
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}

func (impl GitAzureClient) ensureProjectAvailabilityOnSsh(projectName string, repoName string, repoUrl string) (bool, error) {
	for count := 0; count < 8; count++ {
		_, err := impl.gitService.Clone(repoUrl, fmt.Sprintf("/ensure-clone/%s", projectName))
		if err == nil {
			impl.logger.Infow("ensureProjectAvailability clone passed", "try count", count, "repoUrl", repoUrl)
			return true, nil
		}
		impl.logger.Errorw("ensureProjectAvailability clone failed ssh ", "try count", count, "err", err)
		time.Sleep(10 * time.Second)
	}
	return false, nil
}
