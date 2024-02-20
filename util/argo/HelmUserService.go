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

func (impl *HelmUserServiceImpl) BuildACDContext() (acdContext context.Context, err error) {
	return context.Background(), nil
}
