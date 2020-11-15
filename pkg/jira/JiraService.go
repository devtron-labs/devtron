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

package jira

import (
	"errors"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	jiraUtil "github.com/devtron-labs/devtron/internal/util/JiraUtil"
	"github.com/devtron-labs/devtron/pkg/projectManagementService/jira"
	"go.uber.org/zap"
	"time"
)

type ProjectManagementService interface {
	UpdateJiraStatus(updateIssueBean *UpdateIssueBean, userId int32) (map[string][]string, error)
	SaveAccountDetails(jiraConfig *ConfigBean, userId int32) (*repository.JiraAccountDetails, error)
}

type ProjectManagementServiceImpl struct {
	logger                *zap.SugaredLogger
	accountValidator      jira.AccountValidator
	jiraAccountService    jira.AccountService
	jiraAccountRepository repository.JiraAccountRepository
}

func NewProjectManagementServiceImpl(logger *zap.SugaredLogger, jiraAccountService jira.AccountService,
	jiraAccountRepository repository.JiraAccountRepository, accountValidator jira.AccountValidator) *ProjectManagementServiceImpl {
	return &ProjectManagementServiceImpl{
		logger:                logger,
		jiraAccountService:    jiraAccountService,
		jiraAccountRepository: jiraAccountRepository,
		accountValidator:      accountValidator,
	}
}

type Commit struct {
	CommitId      string `json:"commitId"`
	CommitMessage string `json:"commitMessage"`
}

type UpdateIssueBean struct {
	PipelineId    int32    `json:"pipelineId" validate:"number,required"`
	PipelineStage string   `json:"pipelineStage" validate:"required"`
	Commits       []Commit `json:"commits" validate:"required" `
}

type ConfigBean struct {
	UserId                         int32  `json:"userId" validate:"number,required"`
	PipelineId                     int32  `json:"pipelineId" validate:"number,required"`
	PipelineStage                  string `json:"pipelineStage" validate:"required"`
	FinalIssueStatus               string `json:"finalIssueStatus" validate:"required"`
	ProjectManagementToolAuthToken string `json:"projectManagementToolAuthToken" validate:"required"`
	CommitIdRegex                  string `json:"commitIdRegex" validate:"required"`
	ToolUserName                   string `json:"toolUserName" validate:"required"`
	CompanyToolUrl                 string `json:"companyToolUrl" validate:"required"`
}

type AccountDetails struct {
	UserId                         int32
	ProjectManagementToolAuthToken string
	CommitIdRegex                  string
	UserName                       string
	URL                            string
	FinalIssueStatus               string
}

func (impl *ProjectManagementServiceImpl) SaveAccountDetails(jiraConfig *ConfigBean, userId int32) (*repository.JiraAccountDetails, error) {
	if jiraConfig == nil {
		impl.logger.Errorw("error in saving account details", "jiraConfig", jiraConfig)
		return nil, errors.New("failed SaveJiraAccountDetails, invalid accountDetails")
	}

	isAuthenticated, err := impl.accountValidator.ValidateUserAccount(jiraConfig.CompanyToolUrl, jiraConfig.ToolUserName, jiraConfig.ProjectManagementToolAuthToken)
	if err != nil {
		impl.logger.Errorw("some error in saving account details", "err", err)
		return nil, err
	}
	if !isAuthenticated {
		impl.logger.Errorw("cannot SaveJiraAccountDetails, jira authentication failed", "isAuthenticated", isAuthenticated)
		return nil, errors.New("jira authentication failed")
	}

	account := &repository.JiraAccountDetails{
		UserName:           jiraConfig.ToolUserName,
		AccountURL:         jiraConfig.CompanyToolUrl,
		AuthToken:          jiraConfig.ProjectManagementToolAuthToken,
		CommitMessageRegex: jiraConfig.CommitIdRegex,
		FinalIssueStatus:   jiraConfig.FinalIssueStatus,
		PipelineStage:      jiraConfig.PipelineStage,
		PipelineId:         jiraConfig.PipelineId,
		AuditLog: models.AuditLog{
			CreatedBy: jiraConfig.UserId,
			CreatedOn: time.Now(),
			UpdatedOn: time.Now(),
			UpdatedBy: jiraConfig.UserId,
		},
	}

	err = impl.jiraAccountRepository.Save(account)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (impl *ProjectManagementServiceImpl) UpdateJiraStatus(UpdateIssueBean *UpdateIssueBean, userId int32) (map[string][]string, error) {
	commits := UpdateIssueBean.Commits
	if len(commits) == 0 {
		impl.logger.Errorw("no commits provided", "commits len", len(commits))
		return nil, errors.New("no commits provided")
	}
	accountDetails, err := impl.jiraAccountRepository.FindByPipelineIdAndStage(UpdateIssueBean.PipelineId, UpdateIssueBean.PipelineStage)
	if err != nil {
		impl.logger.Errorw("failed to get user account details", "err", err)
		return nil, err
	}

	regex := accountDetails.CommitMessageRegex
	commitMap, invalidRegexCommits := impl.buildCommitMap(commits, regex)

	finalIssueStatus := accountDetails.FinalIssueStatus

	updateRequest := &jira.UpdateRequest{
		UserId:     userId,
		UserName:   accountDetails.UserName,
		AuthToken:  accountDetails.AuthToken,
		AccountURL: accountDetails.AccountURL,
		CommitMap:  commitMap,
	}
	resp, err := impl.jiraAccountService.UpdateJiraStatus(updateRequest, finalIssueStatus)
	if err != nil {
		return nil, err
	}
	invalidCommits := map[string][]string{"invalidCommits": resp}
	if invalidRegexCommits != nil {
		invalidCommits["invalidRegexCommits"] = invalidRegexCommits
	}
	return invalidCommits, nil
}

func (impl *ProjectManagementServiceImpl) buildCommitMap(commits []Commit, regex string) (map[string]string, []string) {
	var invalidCommits []string
	commitsMap := make(map[string]string)
	for _, commit := range commits {
		issueIds, err := jiraUtil.ExtractRegex(regex, commit.CommitMessage)
		if err != nil {
			impl.logger.Errorw("failed to extract issueIds for commit: ", commit.CommitId, "err", err)
			invalidCommits = append(invalidCommits, "failed to extract regex for commitId: "+commit.CommitId+" and commit msg: "+commit.CommitMessage)
			continue
		}
		for _, issueId := range issueIds {
			if commitsMap[issueId] == "" {
				commitsMap[issueId] = commit.CommitId
			}
		}
	}
	return commitsMap, invalidCommits
}
