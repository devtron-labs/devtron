package appStoreDeploymentTool

import (
	"context"
	appStoreBean "github.com/devtron-labs/devtron/pkg/appStore/bean"
	appStoreRepository "github.com/devtron-labs/devtron/pkg/appStore/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"net/http"
)

type AppStoreDeploymentToolService interface {
	InstallApp(installAppVersionRequest *appStoreBean.InstallAppVersionDTO, ctx context.Context) error
	GetAppStatus(installedAppAndEnvDetails appStoreRepository.InstalledAppAndEnvDetails, w http.ResponseWriter, r *http.Request, token string) (string, error)
	DeleteInstalledApp(ctx context.Context, appName string, environmentName string, installAppVersionRequest *appStoreBean.InstallAppVersionDTO, installedApps *appStoreRepository.InstalledApps, dbTransaction *pg.Tx) error
}

type AppStoreDeploymentToolServiceImpl struct {
	Logger *zap.SugaredLogger
}
