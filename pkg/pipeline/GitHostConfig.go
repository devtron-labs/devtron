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
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"go.uber.org/zap"
)

type GitHostConfig interface {
	GetAll() ([]GitHostRequest, error)
	GetById(id int) (*GitHostRequest, error)
}

type GitHostConfigImpl struct {
	logger          *zap.SugaredLogger
	gitHostRepo repository.GitHostRepository
}

func NewGitHostConfigImpl(gitHostRepo repository.GitHostRepository, logger *zap.SugaredLogger) *GitHostConfigImpl {
	return &GitHostConfigImpl{
		logger: logger,
		gitHostRepo: gitHostRepo,
	}
}

type GitHostRequest struct {
	Id          	int                 `json:"id,omitempty" validate:"number"`
	Name        	string              `json:"name,omitempty" validate:"required"`
	Active      	bool                `json:"active"`
	WebhookUrl  	string 				`json:"webhookUrl"`
	WebhookSecret 	string 				`json:"webhookSecret"`
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
			Id:   host.Id,
			Name: host.Name,
			Active:  host.Active,
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

	gitHost := &GitHostRequest{
		Id:   host.Id,
		Name: host.Name,
		Active:  host.Active,
		WebhookUrl : host.WebhookUrl,
		WebhookSecret  : host.WebhookSecret,
	}

	return gitHost, err
}


