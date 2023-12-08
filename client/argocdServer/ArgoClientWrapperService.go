package argocdServer

import (
	"context"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
	"go.uber.org/zap"
)

type ArgoClientWrapperService interface {
	GetArgoAppWithNormalRefresh(context context.Context, argoAppName string) error
	SyncArgoCDApplicationWithRefresh(context context.Context, argoAppName string) error
}

type ArgoClientWrapperServiceImpl struct {
	logger    *zap.SugaredLogger
	acdClient application.ServiceClient
}

func NewArgoClientWrapperServiceImpl(logger *zap.SugaredLogger,
	acdClient application.ServiceClient,
) *ArgoClientWrapperServiceImpl {
	return &ArgoClientWrapperServiceImpl{
		logger:    logger,
		acdClient: acdClient,
	}
}

func (impl *ArgoClientWrapperServiceImpl) GetArgoAppWithNormalRefresh(context context.Context, argoAppName string) error {
	refreshType := bean.RefreshTypeNormal
	impl.logger.Debugw("trying to normal refresh application through get ", "argoAppName", argoAppName)
	_, err := impl.acdClient.Get(context, &application2.ApplicationQuery{Name: &argoAppName, Refresh: &refreshType})
	if err != nil {
		impl.logger.Errorw("cannot get application with refresh", "app", argoAppName)
		return err
	}
	impl.logger.Debugw("done getting the application with refresh with no error", "argoAppName", argoAppName)
	return nil
}

func (impl *ArgoClientWrapperServiceImpl) SyncArgoCDApplicationWithRefresh(context context.Context, argoAppName string) error {
	impl.logger.Info("argocd manual sync for app started", "argoAppName", argoAppName)
	revision := "master"
	_, syncErr := impl.acdClient.Sync(context, &application2.ApplicationSyncRequest{Name: &argoAppName, Revision: &revision})
	if syncErr != nil {
		impl.logger.Errorw("cannot get application with refresh", "app", argoAppName)
		return syncErr
	}
	refreshErr := impl.GetArgoAppWithNormalRefresh(context, argoAppName)
	if refreshErr != nil {
		impl.logger.Errorw("error in refreshing argo app", "err", refreshErr)
	}
	impl.logger.Debugw("argo app sync completed", "argoAppName", argoAppName)
	return nil
}
