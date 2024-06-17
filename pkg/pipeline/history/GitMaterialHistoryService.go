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

package history

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type GitMaterialHistoryService interface {
	CreateMaterialHistory(tx *pg.Tx, inputMaterial *pipelineConfig.GitMaterial) error
	CreateDeleteMaterialHistory(materials []*pipelineConfig.GitMaterial) error
	MarkMaterialDeletedAndCreateHistory(tx *pg.Tx, material *pipelineConfig.GitMaterial) error
}

type GitMaterialHistoryServiceImpl struct {
	gitMaterialHistoryRepository repository.GitMaterialHistoryRepository
	logger                       *zap.SugaredLogger
}

func NewGitMaterialHistoryServiceImpl(gitMaterialHistoryRepository repository.GitMaterialHistoryRepository,
	logger *zap.SugaredLogger) *GitMaterialHistoryServiceImpl {

	return &GitMaterialHistoryServiceImpl{
		gitMaterialHistoryRepository: gitMaterialHistoryRepository,
		logger:                       logger,
	}
}

func (impl GitMaterialHistoryServiceImpl) CreateMaterialHistory(tx *pg.Tx, inputMaterial *pipelineConfig.GitMaterial) error {

	material := &repository.GitMaterialHistory{
		GitMaterialId:   inputMaterial.Id,
		Url:             inputMaterial.Url,
		AppId:           inputMaterial.AppId,
		Name:            inputMaterial.Name,
		GitProviderId:   inputMaterial.GitProviderId,
		Active:          inputMaterial.Active,
		CheckoutPath:    inputMaterial.CheckoutPath,
		FetchSubmodules: inputMaterial.FetchSubmodules,
		FilterPattern:   inputMaterial.FilterPattern,
		AuditLog:        sql.AuditLog{UpdatedBy: inputMaterial.UpdatedBy, CreatedBy: inputMaterial.CreatedBy, UpdatedOn: inputMaterial.UpdatedOn, CreatedOn: inputMaterial.CreatedOn},
	}
	err := impl.gitMaterialHistoryRepository.SaveGitMaterialHistory(tx, material)
	if err != nil {
		impl.logger.Errorw("error in saving create/update history for git repository")
	}

	return nil

}

func (impl GitMaterialHistoryServiceImpl) CreateDeleteMaterialHistory(materials []*pipelineConfig.GitMaterial) error {

	materialsHistory := []*repository.GitMaterialHistory{}

	for _, material := range materials {

		materialHistory := repository.GitMaterialHistory{
			GitMaterialId:   material.Id,
			AppId:           material.AppId,
			GitProviderId:   material.GitProviderId,
			Active:          material.Active,
			Url:             material.Url,
			Name:            material.Name,
			CheckoutPath:    material.CheckoutPath,
			FetchSubmodules: material.FetchSubmodules,
			FilterPattern:   material.FilterPattern,
			AuditLog: sql.AuditLog{
				CreatedOn: material.CreatedOn,
				CreatedBy: material.CreatedBy,
				UpdatedOn: material.UpdatedOn,
				UpdatedBy: material.UpdatedBy,
			},
		}
		materialsHistory = append(materialsHistory, &materialHistory)
	}

	err := impl.gitMaterialHistoryRepository.SaveDeleteMaterialHistory(materialsHistory)

	if err != nil {
		impl.logger.Errorw("Error in saving delete history for git material Repository")
	}

	return nil

}

func (impl GitMaterialHistoryServiceImpl) MarkMaterialDeletedAndCreateHistory(tx *pg.Tx, material *pipelineConfig.GitMaterial) error {

	material.Active = false

	err := impl.CreateMaterialHistory(tx, material)

	if err != nil {
		impl.logger.Errorw("error in saving delete history for git material repository")
	}

	return nil

}
