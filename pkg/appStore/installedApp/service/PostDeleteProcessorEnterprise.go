package service

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"go.uber.org/zap"
)

type DeletePostProcessorEnterpriseImpl struct {
	devtronResourceService devtronResource.DevtronResourceService
	logger                 *zap.SugaredLogger
	*DeletePostProcessorImpl
}

func NewDeletePostProcessorEnterpriseImpl(
	devtronResourceService devtronResource.DevtronResourceService,
	logger *zap.SugaredLogger,
) *DeletePostProcessorEnterpriseImpl {
	return &DeletePostProcessorEnterpriseImpl{
		devtronResourceService:  devtronResourceService,
		logger:                  logger,
		DeletePostProcessorImpl: NewDeletePostProcessorImpl(logger),
	}
}

func (impl *DeletePostProcessorEnterpriseImpl) Process(app *app.App, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) {
	var err error
	go func() {
		errInResourceDelete := impl.devtronResourceService.DeleteObjectAndItsDependency(app.Id, bean.DevtronResourceApplication,
			bean.DevtronResourceHelmApplication, bean.DevtronResourceVersion1, installAppVersionRequest.UserId)
		if errInResourceDelete != nil {
			impl.logger.Errorw("error in deleting helm application resource and dependency data", "err", err, "appId", app.Id)
		}
	}()
}
