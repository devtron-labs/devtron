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

package restHandler

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/jira"
	"github.com/devtron-labs/devtron/pkg/user"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type JiraRestHandler interface {
	SaveAccountConfig(w http.ResponseWriter, r *http.Request)
	UpdateIssueStatus(w http.ResponseWriter, r *http.Request)
}

type JiraRestHandlerImpl struct {
	jiraService     jira.ProjectManagementService
	logger          *zap.SugaredLogger
	userAuthService user.UserService
	validator       *validator.Validate
}

func NewJiraRestHandlerImpl(jiraService jira.ProjectManagementService, logger *zap.SugaredLogger, userAuthService user.UserService, validator *validator.Validate) *JiraRestHandlerImpl {
	return &JiraRestHandlerImpl{
		jiraService:     jiraService,
		logger:          logger,
		userAuthService: userAuthService,
		validator:       validator,
	}
}

func (impl JiraRestHandlerImpl) SaveAccountConfig(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var jiraConfigBean jira.ConfigBean
	err = decoder.Decode(&jiraConfigBean)
	if err != nil {
		impl.logger.Errorw("request err, SaveAccountConfig", "err", err, "payload", jiraConfigBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, SaveAccountConfig", "err", err, "payload", jiraConfigBean)
	err = impl.validator.Struct(jiraConfigBean)
	if err != nil {
		impl.logger.Errorw("validation err, SaveAccountConfig", "err", err, "payload", jiraConfigBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	account, err := impl.jiraService.SaveAccountDetails(&jiraConfigBean, userId)
	if err != nil {
		impl.logger.Errorw("service err, SaveAccountConfig", "err", err, "payload", jiraConfigBean)
		common.WriteJsonResp(w, err, "error in saving jira config", http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, account.Id, http.StatusOK)
}

func (impl JiraRestHandlerImpl) UpdateIssueStatus(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var updateBean jira.UpdateIssueBean
	err = json.NewDecoder(r.Body).Decode(&updateBean)
	if err != nil {
		impl.logger.Errorw("request err, UpdateIssueStatus", "err", err, "payload", updateBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, UpdateIssueStatus", "err", err, "payload", updateBean)
	err = impl.validator.Struct(updateBean)
	if err != nil {
		impl.logger.Errorw("validation err, UpdateIssueStatus", "err", err, "payload", updateBean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := impl.jiraService.UpdateJiraStatus(&updateBean, userId)
	if err != nil {
		impl.logger.Errorw("service err, UpdateIssueStatus", "err", err, "payload", updateBean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
