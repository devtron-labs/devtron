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
