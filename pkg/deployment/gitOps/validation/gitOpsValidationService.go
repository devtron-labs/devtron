/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package validation

import (
	"fmt"
	apiBean "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/service/FullMode"
	chartService "github.com/devtron-labs/devtron/pkg/chart"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/config/bean"
	"github.com/devtron-labs/devtron/pkg/deployment/gitOps/git"
	gitOpsBean "github.com/devtron-labs/devtron/pkg/deployment/gitOps/validation/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/xanzy/go-gitlab"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

type GitOpsValidationService interface {
	// GitOpsValidateDryRun
	GitOpsValidateDryRun(config *apiBean.GitOpsConfigDto) apiBean.DetailedErrorGitOpsConfigResponse
	// ValidateCustomGitRepoURL performs the following validations:
	// "Get Repo URL", "Create Repo (if doesn't exist)", "Organisational URL Validation", "Unique GitOps Repo"
	// And returns: RepoUrl and isNew Repository url and error
	ValidateCustomGitRepoURL(request gitOpsBean.ValidateCustomGitRepoURLRequest) (string, bool, error)
}

type GitOpsValidationServiceImpl struct {
	logger                  *zap.SugaredLogger
	gitFactory              *git.GitFactory
	gitOpsConfigReadService config.GitOpsConfigReadService
	gitOperationService     git.GitOperationService
	chartTemplateService    util.ChartTemplateService
	chartService            chartService.ChartService
	installedAppService     FullMode.InstalledAppDBExtendedService
}

func NewGitOpsValidationServiceImpl(Logger *zap.SugaredLogger,
	gitFactory *git.GitFactory,
	gitOperationService git.GitOperationService,
	gitOpsConfigReadService config.GitOpsConfigReadService,
	chartTemplateService util.ChartTemplateService) *GitOpsValidationServiceImpl {
	return &GitOpsValidationServiceImpl{
		logger:                  Logger,
		gitFactory:              gitFactory,
		gitOpsConfigReadService: gitOpsConfigReadService,
		gitOperationService:     gitOperationService,
		chartTemplateService:    chartTemplateService,
	}
}

func (impl *GitOpsValidationServiceImpl) GitOpsValidateDryRun(config *apiBean.GitOpsConfigDto) apiBean.DetailedErrorGitOpsConfigResponse {
	if config.AllowCustomRepository {
		return apiBean.DetailedErrorGitOpsConfigResponse{
			ValidationSkipped: true,
		}
	}
	detailedErrorGitOpsConfigActions := git.DetailedErrorGitOpsConfigActions{}
	detailedErrorGitOpsConfigActions.StageErrorMap = make(map[string]error)

	if strings.ToUpper(config.Provider) == bean.BITBUCKET_PROVIDER {
		config.Host = git.BITBUCKET_CLONE_BASE_URL
		config.BitBucketProjectKey = strings.ToUpper(config.BitBucketProjectKey)
	}
	client, gitService, err := impl.gitFactory.NewClientForValidation(config)
	if err != nil {
		impl.logger.Errorw("error in creating new client for validation")
		detailedErrorGitOpsConfigActions.StageErrorMap[fmt.Sprintf("error in connecting with %s", strings.ToUpper(config.Provider))] = impl.extractErrorMessageByProvider(err, config.Provider)
		detailedErrorGitOpsConfigActions.ValidatedOn = time.Now()
		detailedErrorGitOpsConfigResponse := impl.convertDetailedErrorToResponse(detailedErrorGitOpsConfigActions)
		return detailedErrorGitOpsConfigResponse
	}
	appName := gitOpsBean.DryrunRepoName + util2.Generate(6)
	//getting user name & emailId for commit author data
	userEmailId, userName := impl.gitOpsConfigReadService.GetUserEmailIdAndNameForGitOpsCommit(config.UserId)
	config.UserEmailId = userEmailId
	config.GitRepoName = appName
	repoUrl, _, detailedErrorCreateRepo := client.CreateRepository(config)

	detailedErrorGitOpsConfigActions.StageErrorMap = detailedErrorCreateRepo.StageErrorMap
	detailedErrorGitOpsConfigActions.SuccessfulStages = detailedErrorCreateRepo.SuccessfulStages

	for stage, stageErr := range detailedErrorGitOpsConfigActions.StageErrorMap {
		if stage == gitOpsBean.CreateRepoStage || stage == gitOpsBean.GetRepoUrlStage {
			_, ok := detailedErrorGitOpsConfigActions.StageErrorMap[gitOpsBean.GetRepoUrlStage]
			if ok {
				detailedErrorGitOpsConfigActions.StageErrorMap[fmt.Sprintf("error in connecting with %s", strings.ToUpper(config.Provider))] = impl.extractErrorMessageByProvider(stageErr, config.Provider)
				delete(detailedErrorGitOpsConfigActions.StageErrorMap, gitOpsBean.GetRepoUrlStage)
			} else {
				detailedErrorGitOpsConfigActions.StageErrorMap[gitOpsBean.CreateRepoStage] = impl.extractErrorMessageByProvider(stageErr, config.Provider)
			}
			detailedErrorGitOpsConfigActions.ValidatedOn = time.Now()
			detailedErrorGitOpsConfigResponse := impl.convertDetailedErrorToResponse(detailedErrorGitOpsConfigActions)
			return detailedErrorGitOpsConfigResponse
		} else if stage == gitOpsBean.CloneHttp || stage == gitOpsBean.CreateReadmeStage {
			detailedErrorGitOpsConfigActions.StageErrorMap[stage] = impl.extractErrorMessageByProvider(stageErr, config.Provider)
		}
	}
	chartDir := fmt.Sprintf("%s-%s", appName, impl.chartTemplateService.GetDir())
	clonedDir := gitService.GetCloneDirectory(chartDir)
	if _, err := os.Stat(clonedDir); os.IsNotExist(err) {
		clonedDir, err = gitService.Clone(repoUrl, chartDir)
		if err != nil {
			impl.logger.Errorw("error in cloning repo", "url", repoUrl, "err", err)
			detailedErrorGitOpsConfigActions.StageErrorMap[gitOpsBean.CloneStage] = err
		} else {
			detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, gitOpsBean.CloneStage)
		}
	}

	commit, err := gitService.CommitAndPushAllChanges(clonedDir, "first commit", userName, userEmailId)
	if err != nil {
		impl.logger.Errorw("error in commit and pushing git", "err", err)
		if commit == "" {
			detailedErrorGitOpsConfigActions.StageErrorMap[gitOpsBean.CommitOnRestStage] = err
		} else {
			detailedErrorGitOpsConfigActions.StageErrorMap[gitOpsBean.PushStage] = err
		}
	} else {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, gitOpsBean.CommitOnRestStage, gitOpsBean.PushStage)
	}

	err = client.DeleteRepository(config)
	if err != nil {
		impl.logger.Errorw("error in deleting repo", "err", err)
		//here below the assignment of delete is removed for making this stage optional, and it's failure not preventing it from saving/updating gitOps config
		//detailedErrorGitOpsConfigActions.StageErrorMap[DeleteRepoStage] = impl.extractErrorMessageByProvider(err, config.Provider)
		detailedErrorGitOpsConfigActions.DeleteRepoFailed = true
	} else {
		detailedErrorGitOpsConfigActions.SuccessfulStages = append(detailedErrorGitOpsConfigActions.SuccessfulStages, gitOpsBean.DeleteRepoStage)
	}
	detailedErrorGitOpsConfigActions.ValidatedOn = time.Now()
	defer impl.chartTemplateService.CleanDir(clonedDir)
	detailedErrorGitOpsConfigResponse := impl.convertDetailedErrorToResponse(detailedErrorGitOpsConfigActions)
	return detailedErrorGitOpsConfigResponse
}

func (impl *GitOpsValidationServiceImpl) ValidateCustomGitRepoURL(request gitOpsBean.ValidateCustomGitRepoURLRequest) (string, bool, error) {
	gitOpsRepoName := ""
	if request.GitRepoURL == apiBean.GIT_REPO_DEFAULT || len(request.GitRepoURL) == 0 {
		gitOpsRepoName = impl.gitOpsConfigReadService.GetGitOpsRepoName(request.AppName)
	} else {
		gitOpsRepoName = impl.gitOpsConfigReadService.GetGitOpsRepoNameFromUrl(request.GitRepoURL)
	}

	// CreateGitRepositoryForApp will try to create repository if not present, and returns a sanitized repo url, use this repo url to maintain uniformity
	chartGitAttribute, err := impl.gitOperationService.CreateGitRepositoryForApp(gitOpsRepoName, request.UserId)
	if err != nil {
		impl.logger.Errorw("error in validating custom gitops repo", "err", err)
		return "", false, impl.extractErrorMessageByProvider(err, request.GitOpsProvider)
	}

	if request.GitRepoURL != apiBean.GIT_REPO_DEFAULT {
		// For custom git repo; we expect the chart is not present hence setting isNew flag to be true.
		chartGitAttribute.IsNewRepo = true

		// Validate: Organisational URL starts
		activeGitOpsConfig, err := impl.gitOpsConfigReadService.GetGitOpsConfigActive()
		if err != nil {
			impl.logger.Errorw("error in fetching active gitOps config", "err", err)
			return "", false, err
		}
		repoUrl := strings.ReplaceAll(git.SanitiseCustomGitRepoURL(*activeGitOpsConfig, request.GitRepoURL), ".git", "")
		if !strings.Contains(chartGitAttribute.RepoUrl, repoUrl) {
			impl.logger.Errorw("non-organisational custom gitops repo", "expected repo", chartGitAttribute.RepoUrl, "user given repo", repoUrl)
			nonOrgErr := impl.getValidationErrorForNonOrganisationalURL(*activeGitOpsConfig)
			if nonOrgErr != nil {
				impl.logger.Errorw("non-organisational custom gitops repo validation error", "err", err)
				return "", false, nonOrgErr
			}
		}
		// Validate: Organisational URL ends
	}

	// Validate: Unique GitOps repository URL starts
	isValid := impl.validateUniqueGitOpsRepo(chartGitAttribute.RepoUrl)
	if !isValid {
		impl.logger.Errorw("git repo url already exists", "repo url", chartGitAttribute.RepoUrl)
		return "", false, fmt.Errorf("invalid git repository! '%s' is already in use by another application! Use a different repository", chartGitAttribute.RepoUrl)
	}
	// Validate: Unique GitOps repository URL ends

	return chartGitAttribute.RepoUrl, chartGitAttribute.IsNewRepo, nil
}

func (impl *GitOpsValidationServiceImpl) extractErrorMessageByProvider(err error, provider string) error {
	switch provider {
	case git.GITLAB_PROVIDER:
		errorResponse, ok := err.(*gitlab.ErrorResponse)
		if ok {
			errorMessage := fmt.Errorf("gitlab client error: %s", errorResponse.Message)
			return errorMessage
		}
		return fmt.Errorf("gitlab client error: %s", err.Error())
	case git.AZURE_DEVOPS_PROVIDER:
		if errorResponse, ok := err.(azuredevops.WrappedError); ok {
			errorMessage := fmt.Errorf("azure devops client error: %s", *errorResponse.Message)
			return errorMessage
		} else if errorResponse, ok := err.(*azuredevops.WrappedError); ok {
			errorMessage := fmt.Errorf("azure devops client error: %s", *errorResponse.Message)
			return errorMessage
		}
		return fmt.Errorf("azure devops client error: %s", err.Error())
	case git.BITBUCKET_PROVIDER:
		return fmt.Errorf("bitbucket client error: %s", err.Error())
	case git.GITHUB_PROVIDER:
		return fmt.Errorf("github client error: %s", err.Error())
	}
	return err
}

func (impl *GitOpsValidationServiceImpl) convertDetailedErrorToResponse(detailedErrorGitOpsConfigActions git.DetailedErrorGitOpsConfigActions) (detailedErrorResponse apiBean.DetailedErrorGitOpsConfigResponse) {
	detailedErrorResponse.StageErrorMap = make(map[string]string)
	detailedErrorResponse.SuccessfulStages = detailedErrorGitOpsConfigActions.SuccessfulStages
	for stage, err := range detailedErrorGitOpsConfigActions.StageErrorMap {
		detailedErrorResponse.StageErrorMap[stage] = err.Error()
	}
	detailedErrorResponse.DeleteRepoFailed = detailedErrorGitOpsConfigActions.DeleteRepoFailed
	detailedErrorResponse.ValidatedOn = detailedErrorGitOpsConfigActions.ValidatedOn
	return detailedErrorResponse
}

func (impl *GitOpsValidationServiceImpl) getValidationErrorForNonOrganisationalURL(activeGitOpsConfig apiBean.GitOpsConfigDto) error {
	var errorMessageKey, errorMessage string
	switch strings.ToUpper(activeGitOpsConfig.Provider) {
	case git.GITHUB_PROVIDER:
		errorMessageKey = "The repository must belong to GitHub organization"
		errorMessage = fmt.Sprintf("%s as configured in global configurations > GitOps", activeGitOpsConfig.GitHubOrgId)

	case git.GITLAB_PROVIDER:
		errorMessageKey = "The repository must belong to gitLab Group ID"
		errorMessage = fmt.Sprintf("%s as configured in global configurations > GitOps", activeGitOpsConfig.GitHubOrgId)

	case git.BITBUCKET_PROVIDER:
		errorMessageKey = "The repository must belong to BitBucket Workspace"
		errorMessage = fmt.Sprintf("%s as configured in global configurations > GitOps", activeGitOpsConfig.BitBucketWorkspaceId)

	case git.AZURE_DEVOPS_PROVIDER:
		errorMessageKey = "The repository must belong to Azure DevOps Project"
		errorMessage = fmt.Sprintf("%s as configured in global configurations > GitOps", activeGitOpsConfig.AzureProjectName)
	}
	return fmt.Errorf("%s: %s", errorMessageKey, errorMessage)
}

func (impl *GitOpsValidationServiceImpl) validateUniqueGitOpsRepo(repoUrl string) (isValid bool) {
	isDevtronAppRegistered, err := impl.chartService.IsGitOpsRepoAlreadyRegistered(repoUrl)
	if err != nil || isDevtronAppRegistered {
		return isValid
	}
	isHelmAppRegistered, err := impl.installedAppService.IsGitOpsRepoAlreadyRegistered(repoUrl)
	if err != nil || isHelmAppRegistered {
		return isValid
	}
	isValid = true
	return isValid
}
