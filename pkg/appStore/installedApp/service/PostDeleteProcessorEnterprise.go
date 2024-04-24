package service

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/adapter"
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"github.com/devtron-labs/devtron/pkg/devtronResource/in"
	"go.uber.org/zap"
)

type DeletePostProcessorEnterpriseImpl struct {
	dtResourceInternalProcessingService in.InternalProcessingService
	logger                              *zap.SugaredLogger
	*DeletePostProcessorImpl
}

func NewDeletePostProcessorEnterpriseImpl(logger *zap.SugaredLogger,
	dtResourceInternalProcessingService in.InternalProcessingService,

) *DeletePostProcessorEnterpriseImpl {
	return &DeletePostProcessorEnterpriseImpl{
		logger:                              logger,
		dtResourceInternalProcessingService: dtResourceInternalProcessingService,
		DeletePostProcessorImpl:             NewDeletePostProcessorImpl(logger),
	}
}

func (impl *DeletePostProcessorEnterpriseImpl) Process(app *app.App, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) {
	var err error
	go func() {
		deleteReq := adapter.BuildDevtronResourceObjectDescriptorBean(app.Id, bean.DevtronResourceApplication,
			bean.DevtronResourceHelmApplication, bean.DevtronResourceVersion1, installAppVersionRequest.UserId)
		errInResourceDelete := impl.dtResourceInternalProcessingService.DeleteObjectAndItsDependency(deleteReq)
		if errInResourceDelete != nil {
			impl.logger.Errorw("error in deleting helm application resource and dependency data", "err", err, "appId", app.Id)
		}
	}()
}
