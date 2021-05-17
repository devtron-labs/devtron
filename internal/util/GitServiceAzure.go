package util

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"go.uber.org/zap"
	"log"
	"time"
)

type GitAzureClient struct {
	client core.Client

	//config *GitConfig
	logger     *zap.SugaredLogger
	org        string
	gitService GitService
}

func NewGitAzureClient(token string, org string, logger *zap.SugaredLogger, gitService GitService) GitAzureClient {
	ctx := context.Background()
	// Create a connection to your organization
	connection := azuredevops.NewPatConnection("https://dev.azure.com/vikram10874", "lejajzxrkkujsveftgbceeik7bu3d2vsqvblg67vwfq3mymcwwwq")
	// Create a client to interact with the Core area
	coreClient, err := core.NewClient(ctx, connection)
	if err != nil {
		log.Fatal(err)
	}
	return GitAzureClient{client: coreClient, org: org, logger: logger, gitService: gitService}
}

func (impl GitAzureClient) CreateRepository(name, description string) (url string, isNew bool, err error) {
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
	teamProject := &core.TeamProject{
		Name:        &name,
		Description: &description,
		Visibility:  &core.ProjectVisibilityValues.Public,
	}
	operationReference, err := impl.client.QueueCreateProject(ctx, core.QueueCreateProjectArgs{
		ProjectToCreate: teamProject,
	})
	if err != nil {
		impl.logger.Errorw("error in creating repo, ", "repo", name, "err", err)
		return "", true, err
	}
	logger.Infow("repo created ", "r", operationReference.Url)

	validated, err := impl.ensureProjectAvailabilityOnHttp(name)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", name, "err", err)
		return *operationReference.Url, true, err
	}
	if !validated {
		return "", true, fmt.Errorf("unable to validate project:%s  in given time", name)
	}
	_, err = impl.createReadme(name)
	if err != nil {
		impl.logger.Errorw("error in creating readme", "err", err)
		return *operationReference.Url, true, err
	}
	validated, err = impl.ensureProjectAvailabilityOnSsh(name, *operationReference.Url)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", name, "err", err)
		return *operationReference.Url, true, err
	}
	if !validated {
		return "", true, fmt.Errorf("unable to validate project:%s  in given time", name)
	}
	return *operationReference.Url, true, err
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
	//branch := "master"
	//path := filepath.Join(config.ChartLocation, config.FileName)
	ctx := context.Background()
	repoExists := true
	url, err := impl.GetRepoUrl(config.FileName)
	if err != nil {
		responseErr, ok := err.(*github.ErrorResponse)
		if !ok || responseErr.Response.StatusCode != 404 {
			impl.logger.Errorw("error in creating repo", "err", err)
			return "", err
		} else {
			repoExists = false
		}
	}
	if repoExists {
		return url, nil
	}
	teamProject := &core.TeamProject{
		Name:       &config.FileName,
		Visibility: &core.ProjectVisibilityValues.Public,
	}
	operationReference, err := impl.client.UpdateProject(ctx, core.UpdateProjectArgs{
		ProjectUpdate: teamProject,
	})
	if err != nil {
		impl.logger.Errorw("error in creating repo, ", "repo", config.FileName, "err", err)
		return "", err
	}
	logger.Infow("repo created ", "r", operationReference.Url)

	validated, err := impl.ensureProjectAvailabilityOnHttp(config.FileName)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", config.FileName, "err", err)
		return *operationReference.Url, err
	}
	if !validated {
		return "", fmt.Errorf("unable to validate project:%s  in given time", config.FileName)
	}
	_, err = impl.createReadme(config.FileName)
	if err != nil {
		impl.logger.Errorw("error in creating readme", "err", err)
		return *operationReference.Url, err
	}
	validated, err = impl.ensureProjectAvailabilityOnSsh(config.FileName, *operationReference.Url)
	if err != nil {
		impl.logger.Errorw("error in ensuring project availability ", "project", config.FileName, "err", err)
		return *operationReference.Url, err
	}
	if !validated {
		return "", fmt.Errorf("unable to validate project:%s  in given time", config.FileName)
	}
	return *operationReference.Url, err
}

func (impl GitAzureClient) GetRepoUrl(projectName string) (repoUrl string, err error) {
	ctx := context.Background()
	// Get first page of the list of team projects for your organization
	// Get first page of the list of team projects for your organization
	teamProject, err := impl.client.GetProject(ctx, core.GetProjectArgs{
		ProjectId: &projectName,
	})
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	for teamProject == nil {
		return "", fmt.Errorf("no project found")
	}
	return *teamProject.Url, nil
}

func (impl GitAzureClient) ensureProjectAvailabilityOnHttp(projectName string) (bool, error) {
	count := 0
	for count < 3 {
		count = count + 1
		_, err := impl.GetRepoUrl(projectName)
		if err == nil {
			return true, nil
		}
		responseErr, ok := err.(*github.ErrorResponse)
		if !ok || responseErr.Response.StatusCode != 404 {
			impl.logger.Errorw("error in validating repo", "err", err)
			return false, err
		} else {
			impl.logger.Errorw("error in validating repo", "err", err)
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}

func (impl GitAzureClient) ensureProjectAvailabilityOnSsh(projectName string, repoUrl string) (bool, error) {
	count := 0
	for count < 3 {
		count = count + 1
		_, err := impl.gitService.Clone(repoUrl, fmt.Sprintf("/ensure-clone/%s", projectName))
		if err == nil {
			impl.logger.Infow("ensureProjectAvailability clone passed", "try count", count, "repoUrl", repoUrl)
			return true, nil
		}
		if err != nil {
			impl.logger.Errorw("ensureProjectAvailability clone failed", "try count", count, "err", err)
		}
		time.Sleep(10 * time.Second)
	}
	return false, nil
}
