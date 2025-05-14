/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package read

import (
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/adapter"
	bean2 "github.com/devtron-labs/devtron/pkg/build/git/gitProvider/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/repository"
	"go.uber.org/zap"
)

type GitProviderReadService interface {
	GetAll() ([]bean2.GitRegistry, error)
	FetchAllGitProviders() ([]bean2.GitRegistry, error)
	FetchOneGitProvider(id string) (*bean2.GitRegistry, error)
	FindByUrl(url string) (*bean2.GitRegistry, error)
}

type GitProviderReadServiceImpl struct {
	logger                *zap.SugaredLogger
	gitProviderRepository repository.GitProviderRepository
}

func NewGitProviderReadService(logger *zap.SugaredLogger,
	gitProviderRepository repository.GitProviderRepository) *GitProviderReadServiceImpl {
	return &GitProviderReadServiceImpl{
		logger:                logger,
		gitProviderRepository: gitProviderRepository,
	}
}

// get all active git providers
func (impl *GitProviderReadServiceImpl) GetAll() ([]bean2.GitRegistry, error) {
	impl.logger.Debug("get all provider request")
	providers, err := impl.gitProviderRepository.FindAllActiveForAutocomplete()
	if err != nil {
		impl.logger.Errorw("error in fetch all git providers", "err", err)
		return nil, err
	}
	var gitProviders []bean2.GitRegistry
	for _, provider := range providers {
		providerRes := adapter.ConvertGitRegistryDtoToBean(provider, false)
		gitProviders = append(gitProviders, providerRes)
	}
	return gitProviders, err
}

func (impl *GitProviderReadServiceImpl) FetchAllGitProviders() ([]bean2.GitRegistry, error) {
	impl.logger.Debug("fetch all git providers from db")
	providers, err := impl.gitProviderRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetch all git providers", "err", err)
		return nil, err
	}
	var gitProviders []bean2.GitRegistry
	for _, provider := range providers {
		providerRes := adapter.ConvertGitRegistryDtoToBean(provider, false)
		gitProviders = append(gitProviders, providerRes)
	}
	return gitProviders, err
}

func (impl *GitProviderReadServiceImpl) FetchOneGitProvider(providerId string) (*bean2.GitRegistry, error) {
	impl.logger.Debug("fetch git provider by ID from db")
	provider, err := impl.gitProviderRepository.FindOne(providerId)
	if err != nil {
		impl.logger.Errorw("error in fetch all git providers", "err", err)
		return nil, err
	}

	providerRes := adapter.ConvertGitRegistryDtoToBean(provider, true)
	return &providerRes, err
}
func (impl *GitProviderReadServiceImpl) FindByUrl(url string) (*bean2.GitRegistry, error) {
	provider, err := impl.gitProviderRepository.FindByUrl(url)
	if err != nil {
		impl.logger.Errorw("error in FindByUrl", "url", url, "err", err)
		return nil, err
	}
	gitRegistryBean := adapter.ConvertGitRegistryDtoToBean(provider, true)
	return &gitRegistryBean, nil
}
