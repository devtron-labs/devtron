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
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type GitRegistryConfig interface {
	Create(request *GitRegistryRequest) (*GitRegistryRequest, error)
	GetAll() ([]GitRegistryRequest, error)
	FetchAllGitProviders() ([]GitRegistryRequest, error)
	FetchOneGitProvider(id string) (*GitRegistryRequest, error)
	Update(request *GitRegistryRequest) (*GitRegistryRequest, error)
}
type GitRegistryConfigImpl struct {
	logger          *zap.SugaredLogger
	gitProviderRepo repository.GitProviderRepository
	GitSensorClient gitSensor.GitSensorClient
}

type GitRegistryRequest struct {
	Id          int                 `json:"id,omitempty" validate:"number"`
	Name        string              `json:"name,omitempty" validate:"required"`
	Url         string              `json:"url,omitempty"`
	UserName    string              `json:"userName,omitempty"`
	Password    string              `json:"password,omitempty"`
	SshKey      string              `json:"sshKey,omitempty"`
	AccessToken string              `json:"accessToken,omitempty"`
	AuthMode    repository.AuthMode `json:"authMode,omitempty" validate:"required"`
	Active      bool                `json:"active"`
	UserId      int32               `json:"-"`
	GitHostId   int 				`json:"gitHostId"`
}

func NewGitRegistryConfigImpl(logger *zap.SugaredLogger, gitProviderRepo repository.GitProviderRepository, GitSensorClient gitSensor.GitSensorClient) *GitRegistryConfigImpl {
	return &GitRegistryConfigImpl{
		logger:          logger,
		gitProviderRepo: gitProviderRepo,
		GitSensorClient: GitSensorClient,
	}
}

func (impl GitRegistryConfigImpl) Create(request *GitRegistryRequest) (*GitRegistryRequest, error) {
	impl.logger.Debugw("get repo create request", "req", request)
	exist, err := impl.gitProviderRepo.ProviderExists(request.Url)
	if err != nil {
		impl.logger.Errorw("error in fetch ", "url", request.Url, "err", err)
		err = &util.ApiError{
			//Code:            constants.GitProviderCreateFailed,
			InternalMessage: "git provider creation failed, error in fetching by url",
			UserMessage:     "git provider creation failed, error in fetching by url",
		}
		return nil, err
	}
	if exist {
		impl.logger.Warnw("repo already exists", "url", request.Url)
		err = &util.ApiError{
			Code:            constants.GitProviderCreateFailedAlreadyExists,
			InternalMessage: "git provider already exists",
			UserMessage:     "git provider already exists",
		}
		return nil, errors.NewAlreadyExists(err, request.Url)
	}
	provider := &repository.GitProvider{
		Name:        request.Name,
		Url:         request.Url,
		Id:          request.Id,
		AuthMode:    request.AuthMode,
		Password:    request.Password,
		Active:      request.Active,
		AccessToken: request.AccessToken,
		SshKey:      request.SshKey,
		UserName:    request.UserName,
		AuditLog:    models.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},
		GitHostId: 	 request.GitHostId,
	}
	err = impl.gitProviderRepo.Save(provider)
	if err != nil {
		impl.logger.Errorw("error in saving git repo config", "data", provider, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderCreateFailedInDb,
			InternalMessage: "git provider failed to create in db",
			UserMessage:     "git provider failed to create in db",
		}
		return nil, err
	}
	err = impl.UpdateGitSensor(provider)
	if err != nil {
		impl.logger.Errorw("error in updating git repo config on sensor", "data", provider, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderUpdateFailedInSync,
			InternalMessage: err.Error(),
			UserMessage:     "git provider failed to update in sync",
		}
		return nil, err
	}
	request.Id = provider.Id
	return request, nil
}

//get all active git providers
func (impl GitRegistryConfigImpl) GetAll() ([]GitRegistryRequest, error) {
	impl.logger.Debug("get all provider request")
	providers, err := impl.gitProviderRepo.FindAllActiveForAutocomplete()
	if err != nil {
		impl.logger.Errorw("error in fetch all git providers", "err", err)
		return nil, err
	}
	var gitProviders []GitRegistryRequest
	for _, provider := range providers {
		providerRes := GitRegistryRequest{
			Id:   provider.Id,
			Name: provider.Name,
			Url:  provider.Url,
			GitHostId: provider.GitHostId,
		}
		gitProviders = append(gitProviders, providerRes)
	}
	return gitProviders, err
}

func (impl GitRegistryConfigImpl) FetchAllGitProviders() ([]GitRegistryRequest, error) {
	impl.logger.Debug("fetch all git providers from db")
	providers, err := impl.gitProviderRepo.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetch all git providers", "err", err)
		return nil, err
	}
	var gitProviders []GitRegistryRequest
	for _, provider := range providers {
		providerRes := GitRegistryRequest{
			Id:          provider.Id,
			Name:        provider.Name,
			Url:         provider.Url,
			UserName:    provider.UserName,
			Password:    provider.Password,
			AuthMode:    provider.AuthMode,
			AccessToken: provider.AccessToken,
			SshKey:      provider.SshKey,
			Active:      provider.Active,
			UserId:      provider.CreatedBy,
			GitHostId: 	 provider.GitHostId,
		}
		gitProviders = append(gitProviders, providerRes)
	}
	return gitProviders, err
}

func (impl GitRegistryConfigImpl) FetchOneGitProvider(providerId string) (*GitRegistryRequest, error) {
	impl.logger.Debug("fetch git provider by ID from db")
	provider, err := impl.gitProviderRepo.FindOne(providerId)
	if err != nil {
		impl.logger.Errorw("error in fetch all git providers", "err", err)
		return nil, err
	}

	providerRes := &GitRegistryRequest{
		Id:          provider.Id,
		Name:        provider.Name,
		Url:         provider.Url,
		UserName:    provider.UserName,
		Password:    provider.Password,
		AuthMode:    provider.AuthMode,
		AccessToken: provider.AccessToken,
		SshKey:      provider.SshKey,
		Active:      provider.Active,
		UserId:      provider.CreatedBy,
		GitHostId:   provider.GitHostId,
	}

	return providerRes, err
}

func (impl GitRegistryConfigImpl) Update(request *GitRegistryRequest) (*GitRegistryRequest, error) {
	impl.logger.Debugw("get repo create request", "req", request)

	/*
		exist, err := impl.gitProviderRepo.ProviderExists(request.Url)
		if err != nil {
			impl.logger.Errorw("error in fetch ", "url", request.Url, "err", err)
			return nil, err
		}
		if exist {
			impl.logger.Infow("repo already exists", "url", request.Url)
			return nil, errors.NewAlreadyExists(err, request.Url)
		}
	*/

	providerId := strconv.Itoa(request.Id)
	existingProvider, err0 := impl.gitProviderRepo.FindOne(providerId)
	if err0 != nil {
		impl.logger.Errorw("No matching entry found for update.", "err", err0)
		err0 = &util.ApiError{
			Code:            constants.GitProviderUpdateProviderNotExists,
			InternalMessage: "git provider update failed, provider does not exist",
			UserMessage:     "git provider update failed, provider does not exist",
		}
		return nil, err0
	}
	provider := &repository.GitProvider{
		Name:        request.Name,
		Url:         request.Url,
		Id:          request.Id,
		AuthMode:    request.AuthMode,
		Password:    request.Password,
		Active:      request.Active,
		AccessToken: request.AccessToken,
		SshKey:      request.SshKey,
		UserName:    request.UserName,
		GitHostId:   request.GitHostId,
		AuditLog:    models.AuditLog{CreatedBy: existingProvider.CreatedBy, CreatedOn: existingProvider.CreatedOn, UpdatedOn: time.Now(), UpdatedBy: request.UserId},
	}
	err := impl.gitProviderRepo.Update(provider)
	if err != nil {
		impl.logger.Errorw("error in updating git repo config", "data", provider, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderUpdateFailedInDb,
			InternalMessage: "git provider failed to update in db",
			UserMessage:     "git provider failed to update in db",
		}
		return nil, err
	}
	request.Id = provider.Id
	err = impl.UpdateGitSensor(provider)
	if err != nil {
		impl.logger.Errorw("error in updating git repo config on sensor", "data", provider, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderUpdateFailedInSync,
			InternalMessage: err.Error(),
			UserMessage:     "git provider failed to update in sync",
		}
		return nil, err
	}
	return request, nil
}

func (impl GitRegistryConfigImpl) UpdateGitSensor(provider *repository.GitProvider) error {
	sensorGitProvider := &gitSensor.GitProvider{
		Id:          provider.Id,
		Url:         provider.Url,
		Name:        provider.Name,
		UserName:    provider.UserName,
		AccessToken: provider.AccessToken,
		Password:    provider.Password,
		Active:      provider.Active,
		SshKey:      provider.SshKey,
		AuthMode:    provider.AuthMode,
	}
	_, err := impl.GitSensorClient.SaveGitProvider(sensorGitProvider)
	return err
}
