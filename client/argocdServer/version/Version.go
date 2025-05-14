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

package version

import (
	"context"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/version"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type VersionService interface {
	CheckVersion(conn *grpc.ClientConn) (err error)
	GetVersion(conn *grpc.ClientConn) (apiVersion string, err error)
}

type VersionServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewVersionServiceImpl(logger *zap.SugaredLogger) *VersionServiceImpl {
	return &VersionServiceImpl{logger: logger}
}

func (service VersionServiceImpl) CheckVersion(conn *grpc.ClientConn) (err error) {
	version, err := version.NewVersionServiceClient(conn).Version(context.Background(), &empty.Empty{})
	if err != nil {
		return err
	}
	service.logger.Infow("connected argocd", "serverVersion", version.Version)
	return nil
}

// GetVersion deprecated
func (service VersionServiceImpl) GetVersion(conn *grpc.ClientConn) (apiVersion string, err error) {
	version, err := version.NewVersionServiceClient(conn).Version(context.Background(), &empty.Empty{})
	if err != nil {
		return "", err
	}
	service.logger.Infow("connected argocd", "serverVersion", version.Version)
	return version.Version, nil
}
