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

package service

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"go.uber.org/zap"
)

type DeletePostProcessor interface {
	Process(app *app.App, installAppVersionRequest *appStoreBean.InstallAppVersionDTO)
}

type DeletePostProcessorImpl struct {
	logger *zap.SugaredLogger
}

func NewDeletePostProcessorImpl(
	logger *zap.SugaredLogger,
) *DeletePostProcessorImpl {
	return &DeletePostProcessorImpl{
		logger: logger,
	}
}

func (impl *DeletePostProcessorImpl) Process(app *app.App, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) {
}
