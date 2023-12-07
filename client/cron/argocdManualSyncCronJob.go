package cron

import (
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
)

type ArgocdAppSyncHandler interface {
	ArgocdAppManualSyncJob()
}

type ArgocdAppSyncCron struct {
	logger    *zap.SugaredLogger
	AppConfig *app.AppServiceConfig
	CdHandler pipeline.CdHandler
}

func NewArgocdAppSyncCron(logger *zap.SugaredLogger, AppConfig *app.AppServiceConfig,
	CdHandler pipeline.CdHandler) *ArgocdAppSyncCron {
	//cronLogger := &CronLoggerImpl{logger: logger}
	//cron := cron.New(
	//	cron.WithChain(cron.SkipIfStillRunning(cronLogger)))
	//cron.Start()
	impl := &ArgocdAppSyncCron{
		logger:    logger,
		AppConfig: AppConfig,
		CdHandler: CdHandler,
	}
	//_, err := cron.AddFunc("@every 1m", impl.ArgocdAppManualSyncJob)
	//if err != nil {
	//	logger.Errorw("error in starting argocd application manual sync cron job", "err", err)
	//	return nil
	//}
	return impl
}

func (impl *ArgocdAppSyncCron) ArgocdAppManualSyncJob() {
	err := impl.CdHandler.SyncArgoCdApps(impl.AppConfig.ArgocdManualSyncCronPipelineDeployedBefore)
	if err != nil {
		impl.logger.Errorw("error in syncing argocd apps", "err", err)
	}
}
