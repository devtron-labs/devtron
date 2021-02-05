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

package gitops

import (
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
	"time"
)

type GitOpsConfigService interface {
	CreateGitOpsConfig(config *GitOpsConfigDto) (*GitOpsConfigDto, error)
	UpdateGitOpsConfig(config *GitOpsConfigDto) error
	GetGitOpsConfigById(id int) (*GitOpsConfigDto, error)
	GetAllGitOpsConfig() ([]*GitOpsConfigDto, error)
}

type GitOpsConfigDto struct {
	Id            int    `json:"id,omitempty"`
	Provider      string `json:"provider"`
	Username      string `json:"username"`
	Token         string `json:"token"`
	GitLabGroupId string `json:"gitLabGroupId"`
	GitHubOrgId   string `json:"gitHubOrgId"`
	Host          string `json:"host"`
	Active        bool   `json:"active"`
	UserId        int32  `json:"-"`
}

type GitOpsConfigServiceImpl struct {
	logger           *zap.SugaredLogger
	gitOpsRepository repository.GitOpsConfigRepository
}

func NewGitOpsConfigServiceImpl(Logger *zap.SugaredLogger, ciHandler pipeline.CiHandler, gitOpsRepository repository.GitOpsConfigRepository) *GitOpsConfigServiceImpl {
	return &GitOpsConfigServiceImpl{
		logger:           Logger,
		gitOpsRepository: gitOpsRepository,
	}
}
func (impl *GitOpsConfigServiceImpl) CreateGitOpsConfig(request *GitOpsConfigDto) (*GitOpsConfigDto, error) {
	impl.logger.Debugw("gitops create request", "req", request)
	model := &repository.GitOpsConfig{
		Provider:      request.Provider,
		Username:      request.Username,
		Token:         request.Token,
		GitHubOrgId:   request.GitHubOrgId,
		GitLabGroupId: request.GitLabGroupId,
		Host:          request.Host,
		Active:        request.Active,
		AuditLog:      models.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},
	}
	model, err := impl.gitOpsRepository.CreateGitOpsConfig(model)
	if err != nil {
		impl.logger.Errorw("error in saving gitops config", "data", model, "err", err)
		err = &util.ApiError{
			InternalMessage: "gitops config failed to create in db",
			UserMessage:     "gitops config failed to create in db",
		}
		return nil, err
	}
	request.Id = model.Id
	return request, nil
}
func (impl *GitOpsConfigServiceImpl) UpdateGitOpsConfig(request *GitOpsConfigDto) error {
	impl.logger.Debugw("gitops config update request", "req", request)
	model, err := impl.gitOpsRepository.GetGitOpsConfigById(request.Id)
	if err != nil {
		impl.logger.Errorw("No matching entry found for update.", "id", request.Id)
		err = &util.ApiError{
			InternalMessage: "gitops config update failed, does not exist",
			UserMessage:     "gitops config update failed, does not exist",
		}
		return err
	}
	model.Provider = request.Provider
	model.Username = request.Username
	model.Token = request.Token
	model.GitLabGroupId = request.GitLabGroupId
	model.GitHubOrgId = request.GitHubOrgId
	model.Host = request.Host
	model.Active = request.Active
	err = impl.gitOpsRepository.UpdateGitOpsConfig(model)
	if err != nil {
		impl.logger.Errorw("error in updating team", "data", model, "err", err)
		err = &util.ApiError{
			InternalMessage: "gitops config failed to update in db",
			UserMessage:     "gitops config failed to update in db",
		}
		return err
	}
	request.Id = model.Id
	return nil
}

func (impl *GitOpsConfigServiceImpl) GetGitOpsConfigById(id int) (*GitOpsConfigDto, error) {
	model, err := impl.gitOpsRepository.GetGitOpsConfigById(id)
	if err != nil {
		impl.logger.Errorw("GetGitOpsConfigById, error while get by id", "err", err, "id", id)
		return nil, err
	}
	config := &GitOpsConfigDto{
		Id:            model.Id,
		Provider:      model.Provider,
		GitHubOrgId:   model.GitHubOrgId,
		GitLabGroupId: model.GitLabGroupId,
		Active:        model.Active,
		UserId:        model.CreatedBy,
	}

	return config, err
}

func (impl *GitOpsConfigServiceImpl) GetAllGitOpsConfig() ([]*GitOpsConfigDto, error) {
	models, err := impl.gitOpsRepository.GetAllGitOpsConfig()
	if err != nil {
		impl.logger.Errorw("GetAllGitOpsConfig, error while fetch all", "err", err)
		return nil, err
	}
	var configs []*GitOpsConfigDto
	for _, model := range models {
		config := &GitOpsConfigDto{
			Id:            model.Id,
			Provider:      model.Provider,
			GitHubOrgId:   model.GitHubOrgId,
			GitLabGroupId: model.GitLabGroupId,
			Active:        model.Active,
			UserId:        model.CreatedBy,
		}
		configs = append(configs, config)
	}
	return configs, err
}
