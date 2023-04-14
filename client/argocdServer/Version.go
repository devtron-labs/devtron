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

package argocdServer

import (
	"context"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/version"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
)

type VersionService interface {
	CheckVersion() (err error)
	GetVersion() (apiVersion string, err error)
}

type VersionServiceImpl struct {
	logger                  *zap.SugaredLogger
	argoCDConnectionManager ArgoCDConnectionManager
}

func NewVersionServiceImpl(logger *zap.SugaredLogger, argoCDConnectionManager ArgoCDConnectionManager) *VersionServiceImpl {
	return &VersionServiceImpl{logger: logger, argoCDConnectionManager: argoCDConnectionManager}
}

func (service VersionServiceImpl) CheckVersion() (err error) {
	conn := service.argoCDConnectionManager.GetConnection("")
	version, err := version.NewVersionServiceClient(conn).Version(context.Background(), &empty.Empty{})
	if err != nil {
		return err
	}
	service.logger.Infow("connected argocd", "serverVersion", version.Version)
	return nil
}

// GetVersion deprecated
func (service VersionServiceImpl) GetVersion() (apiVersion string, err error) {
	conn := service.argoCDConnectionManager.GetConnection("")
	version, err := version.NewVersionServiceClient(conn).Version(context.Background(), &empty.Empty{})
	if err != nil {
		return "", err
	}
	service.logger.Infow("connected argocd", "serverVersion", version.Version)
	return version.Version, nil
}
