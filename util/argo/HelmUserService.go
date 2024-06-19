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

package argo

import (
	"context"
	"errors"
	"go.uber.org/zap"
)

// TODO : remove this service completely
type HelmUserServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewHelmUserServiceImpl(Logger *zap.SugaredLogger) (*HelmUserServiceImpl, error) {
	helmUserServiceImpl := &HelmUserServiceImpl{
		logger: Logger,
	}
	return helmUserServiceImpl, nil
}

func (impl *HelmUserServiceImpl) GetLatestDevtronArgoCdUserToken() (string, error) {
	return "", errors.New("method GetLatestDevtronArgoCdUserToken not implemented")
}

func (impl *HelmUserServiceImpl) ValidateGitOpsAndGetOrUpdateArgoCdUserDetail() string {
	return ""
}

func (impl *HelmUserServiceImpl) GetOrUpdateArgoCdUserDetail() string {
	return ""
}

func (impl *HelmUserServiceImpl) GetACDContext(context.Context) (acdContext context.Context, err error) {
	return context.Background(), nil
}

func (impl *HelmUserServiceImpl) SetAcdTokenInContext(ctx context.Context) (context.Context, error) {
	return context.Background(), nil
}
