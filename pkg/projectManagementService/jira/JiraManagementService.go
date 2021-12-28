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
	"strings"

	client "github.com/devtron-labs/devtron/client/jira"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"go.uber.org/zap"
)

type AccountService interface {
	UpdateJiraStatus(UpdateRequest *UpdateRequest, finalIssueStatus string) ([]string, error)
}

type UpdateRequest struct {
	UserId     int32
	UserName   string
	AuthToken  string
	AccountURL string
	CommitMap  map[string]string
}

type AccountServiceImpl struct {
	logger                *zap.SugaredLogger
	jiraAccountRepository repository.JiraAccountRepository
	jiraClient            client.JiraClient
}

func NewAccountServiceImpl(logger *zap.SugaredLogger, jiraAccountRepository repository.JiraAccountRepository, jiraClient client.JiraClient) *AccountServiceImpl {
	return &AccountServiceImpl{logger: logger, jiraAccountRepository: jiraAccountRepository, jiraClient: jiraClient}
}

func (impl *AccountServiceImpl) UpdateJiraStatus(updateRequest *UpdateRequest, finalIssueStatus string) ([]string, error) {
	if updateRequest == nil {
		return nil, errors.New("cannot UpdateJiraStatus, invalid updateRequest")
	}
	// TODO: create goroutines
	var errorCommits []string
	for issueId, commitId := range updateRequest.CommitMap {
		clientReq := client.CreateClientReq(updateRequest.UserName, updateRequest.AuthToken, updateRequest.AccountURL)
		transitionResponses, err := impl.jiraClient.FindIssueTransitions(clientReq, issueId)

		if err != nil {
			impl.logger.Errorw("could not find transitions for ", issueId, "err", err)
			errorCommits = impl.handleJiraUpdateError(issueId, err, errorCommits, commitId)
			continue
		}

		id := ""
		for _, jiraTransition := range transitionResponses {
			if strings.EqualFold(finalIssueStatus, jiraTransition.Name) {
				id = jiraTransition.Id
				break
			}
		}

		if id == "" {
			impl.logger.Errorw("no transition id not found for issue:", "issueId", issueId, "and status:", finalIssueStatus, "err", err)
			errorCommits = impl.handleJiraUpdateError(issueId, err, errorCommits, commitId)
			continue
		}

		resp, err := impl.jiraClient.UpdateJiraTransition(clientReq, issueId, id)
		if err != nil {
			impl.logger.Errorw("could not update transition for ", "issueId", issueId, "err", err)
			errorCommits = impl.handleJiraUpdateError(issueId, err, errorCommits, commitId)
			continue
		}
		if resp.StatusCode != 204 {
			impl.logger.Errorw("update transition jira response status ", "status", resp.Status)
			errorCommits = impl.handleJiraUpdateError(issueId, err, errorCommits, commitId)
		}
	}
	return errorCommits, nil
}

func (impl *AccountServiceImpl) handleJiraUpdateError(issueId string, err error, errorCommits []string, commitId string) []string {
	errorCommits = append(errorCommits, "could not update jira for commitId: "+commitId+" issueId: "+issueId)
	return errorCommits
}
