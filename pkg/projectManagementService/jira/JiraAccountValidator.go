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
	"github.com/devtron-labs/devtron/client/jira"
	"errors"
	"go.uber.org/zap"
)

type AccountValidator interface {
	ValidateUserAccount(jiraAccountUrl string, userName string, token string) (bool, error)
}

type AccountValidatorImpl struct {
	logger     *zap.SugaredLogger
	jiraClient client.JiraClient
}

func NewAccountValidatorImpl(logger *zap.SugaredLogger, jiraClient client.JiraClient) *AccountValidatorImpl {
	return &AccountValidatorImpl{logger: logger, jiraClient: jiraClient}
}

func (impl *AccountValidatorImpl) ValidateUserAccount(jiraAccountUrl string, userName string, token string) (bool, error) {
	if jiraAccountUrl == "" || userName == "" || token == "" {
		impl.logger.Errorw("cannot validate user account for invalid params", "jiraAccountUrl", jiraAccountUrl, "userName", userName, "token")
		return false, errors.New("cannot validate user account for invalid params")
	}

	clientReq := client.CreateClientReq(userName, token, jiraAccountUrl+"/test")
	resp, err := impl.jiraClient.AuthenticateUserAccount(clientReq)

	if err != nil {
		return false, err
	}
	if resp.StatusCode == 401 {
		impl.logger.Errorw("User jira could not be authenticated", "code", resp.StatusCode)
		return false, nil
	}
	return true, nil
}
