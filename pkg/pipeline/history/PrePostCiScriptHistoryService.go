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
	"time"
)

type PrePostCiScriptHistoryService interface {
	CreatePrePostCiScriptHistory(ciPipelineScript *pipelineConfig.CiPipelineScript, tx *pg.Tx, built bool, builtBy int32, builtOn time.Time) (historyModel *repository.PrePostCiScriptHistory, err error)
}

type PrePostCiScriptHistoryServiceImpl struct {
	logger                           *zap.SugaredLogger
	prePostCiScriptHistoryRepository repository.PrePostCiScriptHistoryRepository
}

func NewPrePostCiScriptHistoryServiceImpl(logger *zap.SugaredLogger, prePostCiScriptHistoryRepository repository.PrePostCiScriptHistoryRepository) *PrePostCiScriptHistoryServiceImpl {
	return &PrePostCiScriptHistoryServiceImpl{
		logger:                           logger,
		prePostCiScriptHistoryRepository: prePostCiScriptHistoryRepository,
	}
}

func (impl PrePostCiScriptHistoryServiceImpl) CreatePrePostCiScriptHistory(ciPipelineScript *pipelineConfig.CiPipelineScript, tx *pg.Tx, built bool, builtBy int32, builtOn time.Time) (historyModel *repository.PrePostCiScriptHistory, err error) {
	//creating new entry
	historyModel = &repository.PrePostCiScriptHistory{
		CiPipelineScriptsId: ciPipelineScript.Id,
		Script:              ciPipelineScript.Script,
		Stage:               ciPipelineScript.Stage,
		Name:                ciPipelineScript.Name,
		OutputLocation:      ciPipelineScript.OutputLocation,
		Built:               built,
		BuiltBy:             builtBy,
		BuiltOn:             builtOn,
		AuditLog: sql.AuditLog{
			CreatedOn: ciPipelineScript.CreatedOn,
			CreatedBy: ciPipelineScript.CreatedBy,
			UpdatedOn: ciPipelineScript.UpdatedOn,
			UpdatedBy: ciPipelineScript.UpdatedBy,
		},
	}
	if tx != nil {
		_, err = impl.prePostCiScriptHistoryRepository.CreateHistoryWithTxn(historyModel, tx)
	} else {
		_, err = impl.prePostCiScriptHistoryRepository.CreateHistory(historyModel)
	}
	if err != nil {
		impl.logger.Errorw("err in creating history entry for ci script", "err", err)
		return nil, err
	}
	return historyModel, err
}
