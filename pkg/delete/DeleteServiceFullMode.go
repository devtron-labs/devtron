/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package delete

import (
	"fmt"
	dockerRegistryRepository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type DeleteServiceFullMode interface {
	DeleteGitProvider(deleteRequest *types.GitRegistry) error
	DeleteDockerRegistryConfig(deleteRequest *types.DockerArtifactStoreBean) error
	CanDeleteContainerRegistryConfig(storeId string) bool
}

type DeleteServiceFullModeImpl struct {
	logger                   *zap.SugaredLogger
	gitMaterialRepository    pipelineConfig.MaterialRepository
	gitRegistryConfig        pipeline.GitRegistryConfig
	ciTemplateRepository     pipelineConfig.CiTemplateRepository
	dockerRegistryConfig     pipeline.DockerRegistryConfig
	dockerRegistryRepository dockerRegistryRepository.DockerArtifactStoreRepository
}

func NewDeleteServiceFullModeImpl(logger *zap.SugaredLogger,
	gitMaterialRepository pipelineConfig.MaterialRepository,
	gitRegistryConfig pipeline.GitRegistryConfig,
	ciTemplateRepository pipelineConfig.CiTemplateRepository,
	dockerRegistryConfig pipeline.DockerRegistryConfig,
	dockerRegistryRepository dockerRegistryRepository.DockerArtifactStoreRepository,
) *DeleteServiceFullModeImpl {
	return &DeleteServiceFullModeImpl{
		logger:                   logger,
		gitMaterialRepository:    gitMaterialRepository,
		gitRegistryConfig:        gitRegistryConfig,
		ciTemplateRepository:     ciTemplateRepository,
		dockerRegistryConfig:     dockerRegistryConfig,
		dockerRegistryRepository: dockerRegistryRepository,
	}
}
func (impl DeleteServiceFullModeImpl) DeleteGitProvider(deleteRequest *types.GitRegistry) error {
	//finding if this git account is used in any git material, if yes then will not delete
	materials, err := impl.gitMaterialRepository.FindByGitProviderId(deleteRequest.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in deleting git provider", "gitProvider", deleteRequest.Name, "err", err)
		return err
	}
	if len(materials) > 0 {
		impl.logger.Errorw("err in deleting git provider, found git materials using provider", "gitProvider", deleteRequest.Name)
		return fmt.Errorf(" Please delete all related git materials before deleting this git account")
	}
	err = impl.gitRegistryConfig.Delete(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting git account", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}

func (impl DeleteServiceFullModeImpl) DeleteDockerRegistryConfig(deleteRequest *types.DockerArtifactStoreBean) error {
	//finding if docker reg is used in any app, if yes then will not delete
	ciTemplates, err := impl.ciTemplateRepository.FindByDockerRegistryId(deleteRequest.Id)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in fetching CI build configs attached with registry", "dockerRegistry", deleteRequest.Id, "err", err)
		return err
	}
	if len(ciTemplates) > 0 {
		impl.logger.Errorw("err in deleting docker registry, found docker build config using registry", "dockerRegistry", deleteRequest.Id, "err", err)
		return fmt.Errorf(" Please update all related docker config before deleting this registry")
	}

	//finding if docker reg chart is used in any deployment, if yes then will not delete
	store, err := impl.dockerRegistryRepository.FindOneWithDeploymentCount(deleteRequest.Id)
	if err != nil {
		impl.logger.Errorw("error in fetching registry chart deployment", "dockerRegistry", deleteRequest.Id, "err", err)
		return err
	}
	if store.DeploymentCount > 0 {
		impl.logger.Errorw("err in deleting docker registry, found chart deployments using registry", "dockerRegistry", deleteRequest.Id, "err", err)
		return fmt.Errorf(" Please update all related docker config before deleting this registry")
	}
	err = impl.dockerRegistryConfig.DeleteReg(deleteRequest)
	if err != nil {
		impl.logger.Errorw("error in deleting docker registry", "err", err, "deleteRequest", deleteRequest)
		return err
	}
	return nil
}

func (impl DeleteServiceFullModeImpl) CanDeleteContainerRegistryConfig(storeId string) bool {
	//finding if docker reg is used in any app, if yes then will not delete
	ciTemplates, err := impl.ciTemplateRepository.FindByDockerRegistryId(storeId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("err in deleting docker registry", "dockerRegistry", storeId, "err", err)
		return false
	}
	if len(ciTemplates) > 0 {
		return false
	}
	return true
}
