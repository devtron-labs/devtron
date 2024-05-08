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

package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/util/JiraUtil"
	"go.uber.org/zap"
	"net/http"
)

const JiraGetIssueTransitionsApi = "/rest/api/latest/issue/%s/transitions/"
const JiraUpdateIssueTransitionsApi = "/rest/api/latest/issue/%s/transitions/"

type JiraClient interface {
	AuthenticateUserAccount(clientRequest JiraClientRequest) (*http.Response, error)
	FindIssueTransitions(clientRequest JiraClientRequest, issueId string) ([]JiraTransition, error)
	UpdateJiraTransition(clientRequest JiraClientRequest, issueId string, transitionId string) (*http.Response, error)
}

type JiraClientRequest struct {
	JiraAccountUrl   string
	EncodedAuthToken string
}

type JiraClientImpl struct {
	logger *zap.SugaredLogger
	client *http.Client
}

type JiraTransition struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type JiraTransitionUpdateRequest struct {
	JiraTransition JiraTransition `json:"transition"`
}

type TransitionResponse struct {
	Transitions []JiraTransition `json:"transitions"`
}

func NewJiraClientImpl(logger *zap.SugaredLogger, client *http.Client) *JiraClientImpl {
	return &JiraClientImpl{logger: logger, client: client}
}

func (jiraClientImpl *JiraClientImpl) AuthenticateUserAccount(clientRequest JiraClientRequest) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, clientRequest.JiraAccountUrl, nil)
	if err != nil {
		jiraClientImpl.logger.Errorw("could not create AuthenticateUserAccount request ", "err", err)
		return nil, err
	}
	resp, err := jiraClientImpl.sendRequest(req, clientRequest)
	return resp, err
}

func (jiraClientImpl *JiraClientImpl) FindIssueTransitions(clientRequest JiraClientRequest, issueId string) ([]JiraTransition, error) {
	relUrl := fmt.Sprintf(JiraGetIssueTransitionsApi, issueId)
	req, err := http.NewRequest(http.MethodGet, clientRequest.JiraAccountUrl+relUrl, nil)
	if err != nil {
		jiraClientImpl.logger.Errorw("could not create FindIssueTransitions request ", "err", err)
		return nil, err
	}
	resp, err := jiraClientImpl.sendRequest(req, clientRequest)
	if err != nil {
		jiraClientImpl.logger.Errorw("error while FindIssueTransitions request ", "err", err)
		return nil, err
	}
	decoder := json.NewDecoder(resp.Body)
	var transitionResponse TransitionResponse
	err = decoder.Decode(&transitionResponse)
	return transitionResponse.Transitions, err
}

func (jiraClientImpl *JiraClientImpl) UpdateJiraTransition(clientRequest JiraClientRequest, issueId string, transitionId string) (*http.Response, error) {
	relUrl := fmt.Sprintf(JiraUpdateIssueTransitionsApi, issueId)

	jiraTransitionUpdateRequest := JiraTransitionUpdateRequest{
		JiraTransition: JiraTransition{
			Id: transitionId,
		},
	}

	body, err := json.Marshal(jiraTransitionUpdateRequest)
	if err != nil {
		jiraClientImpl.logger.Errorw("error while marshaling UpdateJiraTransition request ", "err", err)
		return nil, err
	}
	var reqBody = []byte(body)
	req, err := http.NewRequest(http.MethodPost, clientRequest.JiraAccountUrl+relUrl, bytes.NewBuffer(reqBody))
	if err != nil {
		jiraClientImpl.logger.Errorw("error while UpdateJiraTransition request ", "err", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := jiraClientImpl.sendRequest(req, clientRequest)
	return resp, err
}

func (jiraClientImpl *JiraClientImpl) sendRequest(req *http.Request, clientRequest JiraClientRequest) (*http.Response, error) {
	req.Header.Set("Authorization", "Basic "+clientRequest.EncodedAuthToken)
	resp, err := jiraClientImpl.client.Do(req)
	return resp, err
}

func CreateClientReq(userName string, token string, jiraAccountUrl string) JiraClientRequest {
	authParamsEnc := jira.GetEncryptedAuthParams(userName, token)
	clientReq := JiraClientRequest{
		JiraAccountUrl:   jiraAccountUrl,
		EncodedAuthToken: authParamsEnc,
	}
	return clientReq
}
