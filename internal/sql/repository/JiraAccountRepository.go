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

package repository

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"time"
)

type JiraAccountDetails struct {
	tableName          struct{} `sql:"project_management_tool_config" pg:",discard_unknown_columns"`
	Id                 int32    `sql:"id,pk"`
	UserName           string   `sql:"user_name"`
	AccountURL         string   `sql:"account_url"`
	AuthToken          string   `sql:"auth_token"`
	CommitMessageRegex string   `sql:"commit_message_regex"`
	FinalIssueStatus   string   `sql:"final_issue_status"`
	PipelineStage      string   `sql:"pipeline_stage"`
	PipelineId         int32    `sql:"pipeline_id"`
	sql.AuditLog
}

type JiraAccountRepository interface {
	Save(accountDetails *JiraAccountDetails) error
	FindByPipelineIdAndStage(pipelineId int32, pipelineStage string) (*JiraAccountDetails, error)
	Update(accountDetails *JiraAccountDetails) error
}

type JiraAccountRepositoryImpl struct {
	dbConnection *pg.DB
}

func NewJiraAccountRepositoryImpl(dbConnection *pg.DB) *JiraAccountRepositoryImpl {
	return &JiraAccountRepositoryImpl{dbConnection: dbConnection}
}

func (impl *JiraAccountRepositoryImpl) FindByPipelineIdAndStage(pipelineId int32, pipelineStage string) (*JiraAccountDetails, error) {
	details := &JiraAccountDetails{}
	err := impl.dbConnection.Model(details).Where("pipeline_id = ?", pipelineId).Where("pipeline_stage = ?", pipelineStage).Select()
	return details, err
}

func (impl *JiraAccountRepositoryImpl) Save(jiraAccountDetails *JiraAccountDetails) error {
	model, err := impl.FindByPipelineIdAndStage(jiraAccountDetails.PipelineId, jiraAccountDetails.PipelineStage)
	if err == nil && model != nil {
		impl.buildAccountUpdateModel(jiraAccountDetails, model)
		return impl.Update(model)
	}
	return impl.dbConnection.Insert(jiraAccountDetails)
}

func (impl *JiraAccountRepositoryImpl) Update(jiraAccountDetails *JiraAccountDetails) error {
	return impl.dbConnection.Update(jiraAccountDetails)
}

func (impl *JiraAccountRepositoryImpl) buildAccountUpdateModel(jiraAccountDetails *JiraAccountDetails, model *JiraAccountDetails) {
	if jiraAccountDetails.AccountURL != "" {
		model.AccountURL = jiraAccountDetails.AccountURL
	}
	if jiraAccountDetails.AuthToken != "" {
		model.AuthToken = jiraAccountDetails.AuthToken
	}
	if jiraAccountDetails.UserName != "" {
		model.UserName = jiraAccountDetails.UserName
	}
	if jiraAccountDetails.CommitMessageRegex != "" {
		model.CommitMessageRegex = jiraAccountDetails.CommitMessageRegex
	}
	if jiraAccountDetails.PipelineId != 0 {
		model.PipelineId = jiraAccountDetails.PipelineId
	}
	if jiraAccountDetails.PipelineStage != "" {
		model.PipelineStage = jiraAccountDetails.PipelineStage
	}
	model.UpdatedOn = time.Now()
}
