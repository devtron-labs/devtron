package service

import (
	"github.com/devtron-labs/devtron/internals/sql/repository/app"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
)

type SoftDeletePostProcessor interface {
	SoftDeletePostProcessor(app *app.App, installAppVersionRequest *appStoreBean.InstallAppVersionDTO)
}

type SoftDeletePostProcessorImpl struct {
}

func NewSoftDeletePostProcessorImpl() *SoftDeletePostProcessorImpl {
	return &SoftDeletePostProcessorImpl{}
}

func (impl *SoftDeletePostProcessorImpl) SoftDeletePostProcessor(app *app.App, installAppVersionRequest *appStoreBean.InstallAppVersionDTO) {
}
