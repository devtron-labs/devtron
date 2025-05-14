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
	"github.com/devtron-labs/devtron/pkg/build/git/gitMaterial/repository"
	"go.uber.org/zap"
)

type GitMaterialReadService interface {
	FindByAppId(appId int) ([]*repository.GitMaterial, error)
	FindByAppIds(appIds []int) ([]*repository.GitMaterial, error)
	FindById(Id int) (*repository.GitMaterial, error)
	FindByAppIdAndGitMaterialId(appId, id int) (*repository.GitMaterial, error)
	FindByAppIdAndCheckoutPath(appId int, checkoutPath string) (*repository.GitMaterial, error)
	FindByGitProviderId(gitProviderId int) (materials []*repository.GitMaterial, err error)
	FindNumberOfAppsWithGitRepo(appIds []int) (int, error)
}
type GitMaterialReadServiceImpl struct {
	logger             *zap.SugaredLogger
	materialRepository repository.MaterialRepository
}

func NewGitMaterialReadServiceImpl(logger *zap.SugaredLogger,
	materialRepository repository.MaterialRepository) *GitMaterialReadServiceImpl {
	return &GitMaterialReadServiceImpl{
		logger:             logger,
		materialRepository: materialRepository,
	}

}

func (impl *GitMaterialReadServiceImpl) FindByAppId(appId int) ([]*repository.GitMaterial, error) {
	return impl.materialRepository.FindByAppId(appId)
}

func (impl *GitMaterialReadServiceImpl) FindByAppIds(appIds []int) ([]*repository.GitMaterial, error) {
	return impl.materialRepository.FindByAppIds(appIds)
}

func (impl *GitMaterialReadServiceImpl) FindById(id int) (*repository.GitMaterial, error) {
	return impl.materialRepository.FindById(id)
}

func (impl *GitMaterialReadServiceImpl) FindByAppIdAndGitMaterialId(appId, id int) (*repository.GitMaterial, error) {
	return impl.materialRepository.FindByAppIdAndGitMaterialId(appId, id)
}

func (impl *GitMaterialReadServiceImpl) FindByAppIdAndCheckoutPath(appId int, checkoutPath string) (*repository.GitMaterial, error) {
	return impl.materialRepository.FindByAppIdAndCheckoutPath(appId, checkoutPath)
}

func (impl *GitMaterialReadServiceImpl) FindByGitProviderId(gitProviderId int) (materials []*repository.GitMaterial, err error) {
	return impl.materialRepository.FindByGitProviderId(gitProviderId)
}

func (impl *GitMaterialReadServiceImpl) FindNumberOfAppsWithGitRepo(appIds []int) (int, error) {
	return impl.materialRepository.FindNumberOfAppsWithGitRepo(appIds)
}
