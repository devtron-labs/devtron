package service

import (
	"context"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
)

func (impl *HelmAppServiceImpl) GetAppStatusV2(ctx context.Context, req *gRPC.AppDetailRequest, clusterId int) (*gRPC.AppStatus, error) {
	return nil, nil
}

func (impl *HelmAppServiceImpl) updateAppDetailRequestWithCacheConfig(req *gRPC.AppDetailRequest) {
	return
}

func (impl *HelmAppServiceImpl) updateExternalResTreeRequestWithCacheConfig(clusterId int, req *gRPC.ExternalResourceTreeRequest) {
	return
}
