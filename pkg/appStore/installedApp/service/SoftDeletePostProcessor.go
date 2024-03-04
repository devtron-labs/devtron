package service

import (
	"github.com/devtron-labs/devtron/internals/sql/repository/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
)

type DeletePostProcessor interface {
	Process(app *app.App, installAppVersionRequest *appStoreBean.InstallAppVersionDTO)
}

type DeletePostProcessorImpl struct {
}

func NewDeletePostProcessorImpl() *DeletePostProcessorImpl {
	return &DeletePostProcessorImpl{}
}

func (impl *DeletePostProcessorImpl) Process(app *app.App, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) {
}
