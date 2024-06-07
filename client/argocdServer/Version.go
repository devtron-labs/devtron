/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package argocdServer

import (
	"context"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/version"
	"github.com/devtron-labs/devtron/client/argocdServer/connection"
	"github.com/golang/protobuf/ptypes/empty"
	"go.uber.org/zap"
)

type VersionService interface {
	CheckVersion() (err error)
	GetVersion() (apiVersion string, err error)
}

type VersionServiceImpl struct {
	logger                  *zap.SugaredLogger
	argoCDConnectionManager connection.ArgoCDConnectionManager
}

func NewVersionServiceImpl(logger *zap.SugaredLogger, argoCDConnectionManager connection.ArgoCDConnectionManager) *VersionServiceImpl {
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
