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

package pipeline

import (
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"time"
)

type GitHostConfig interface {
	GetAll() ([]GitHostRequest, error)
	GetById(id int) (*GitHostRequest, error)
	Create(request *GitHostRequest) (int, error)
}

type GitHostConfigImpl struct {
	logger           *zap.SugaredLogger
	gitHostRepo      repository.GitHostRepository
	attributeService attributes.AttributesService
}

func NewGitHostConfigImpl(gitHostRepo repository.GitHostRepository, logger *zap.SugaredLogger, attributeService attributes.AttributesService) *GitHostConfigImpl {
	return &GitHostConfigImpl{
		logger:           logger,
		gitHostRepo:      gitHostRepo,
		attributeService: attributeService,
	}
}

type GitHostRequest struct {
	Id              int    `json:"id,omitempty" validate:"number"`
	Name            string `json:"name,omitempty" validate:"required"`
	Active          bool   `json:"active"`
	WebhookUrl      string `json:"webhookUrl"`
	WebhookSecret   string `json:"webhookSecret"`
	EventTypeHeader string `json:"eventTypeHeader"`
	SecretHeader    string `json:"secretHeader"`
	SecretValidator string `json:"secretValidator"`
	UserId          int32  `json:"-"`
}

//get all git hosts
func (impl GitHostConfigImpl) GetAll() ([]GitHostRequest, error) {
	impl.logger.Debug("get all hosts request")
	hosts, err := impl.gitHostRepo.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetching all git hosts", "err", err)
		return nil, err
	}
	var gitHosts []GitHostRequest
	for _, host := range hosts {
		hostRes := GitHostRequest{
			Id:     host.Id,
			Name:   host.Name,
			Active: host.Active,
		}
		gitHosts = append(gitHosts, hostRes)
	}
	return gitHosts, err
}

//get git host by Id
func (impl GitHostConfigImpl) GetById(id int) (*GitHostRequest, error) {
	impl.logger.Debug("get hosts request for Id", id)
	host, err := impl.gitHostRepo.FindOneById(id)
	if err != nil {
		impl.logger.Errorw("error in fetching git host", "err", err)
		return nil, err
	}

	// get orchestrator host
	orchestratorHost, err := impl.attributeService.GetByKey("url")
	if err != nil {
		impl.logger.Errorw("error in fetching orchestrator host url from db", "err", err)
		return nil, err
	}

	gitHost := &GitHostRequest{
		Id:              host.Id,
		Name:            host.Name,
		Active:          host.Active,
		WebhookUrl:      orchestratorHost.Value + host.WebhookUrl,
		WebhookSecret:   host.WebhookSecret,
		EventTypeHeader: host.EventTypeHeader,
		SecretHeader:    host.SecretHeader,
		SecretValidator: host.SecretValidator,
	}

	return gitHost, err
}

// Create in DB
func (impl GitHostConfigImpl) Create(request *GitHostRequest) (int, error) {
	impl.logger.Debugw("get git host create request", "req", request)
	exist, err := impl.gitHostRepo.Exists(request.Name)
	if err != nil {
		impl.logger.Errorw("error in fetching git host ", "name", request.Name, "err", err)
		err = &util.ApiError{
			InternalMessage: "git host creation failed, error in fetching by name",
			UserMessage:     "git host creation failed, error in fetching by name",
		}
		return 0, err
	}
	if exist {
		impl.logger.Warnw("git host already exists", "name", request.Name)
		err = &util.ApiError{
			Code:            constants.GitHostCreateFailedAlreadyExists,
			InternalMessage: "git host already exists",
			UserMessage:     "git host already exists",
		}
		return 0, errors.NewAlreadyExists(err, request.Name)
	}
	gitHost := &repository.GitHost{
		Name:     request.Name,
		Active:   request.Active,
		AuditLog: models.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},
	}
	err = impl.gitHostRepo.Save(gitHost)
	if err != nil {
		impl.logger.Errorw("error in saving git host in db", "data", gitHost, "err", err)
		err = &util.ApiError{
			Code:            constants.GitHostCreateFailedInDb,
			InternalMessage: "git host failed to create in db",
			UserMessage:     "git host failed to create in db",
		}
		return 0, err
	}
	return gitHost.Id, nil
}
