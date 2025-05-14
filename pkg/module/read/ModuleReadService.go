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
	"errors"
	"github.com/devtron-labs/devtron/pkg/module/read/adapter"
	"github.com/devtron-labs/devtron/pkg/module/read/bean"
	moduleErr "github.com/devtron-labs/devtron/pkg/module/read/error"
	moduleRepo "github.com/devtron-labs/devtron/pkg/module/repo"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

type ModuleReadService interface {
	GetModuleInfoByName(moduleName string) (*bean.ModuleInfoMin, error)
}

type ModuleReadServiceImpl struct {
	logger           *zap.SugaredLogger
	moduleRepository moduleRepo.ModuleRepository
}

func NewModuleReadServiceImpl(
	logger *zap.SugaredLogger,
	moduleRepository moduleRepo.ModuleRepository) *ModuleReadServiceImpl {
	return &ModuleReadServiceImpl{
		logger:           logger,
		moduleRepository: moduleRepository,
	}
}

func (impl ModuleReadServiceImpl) GetModuleInfoByName(moduleName string) (*bean.ModuleInfoMin, error) {
	module, err := impl.moduleRepository.FindOne(moduleName)
	if err != nil && !errors.Is(err, pg.ErrNoRows) {
		impl.logger.Errorw("error while fetching module info", "moduleName", moduleName, "error", err)
		return nil, err
	} else if errors.Is(err, pg.ErrNoRows) {
		impl.logger.Debugw("module not found", "moduleName", moduleName)
		return adapter.GetDefaultModuleInfo(moduleName), moduleErr.ModuleNotFoundError
	}
	return adapter.GetModuleInfoMin(module), nil
}
