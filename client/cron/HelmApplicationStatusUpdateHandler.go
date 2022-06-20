package cron

import (
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type HelmApplicationStatusUpdateHandler interface {
	HelmApplicationStatusUpdate()
}

type HelmApplicationStatusUpdateHandlerImpl struct {
	logger              *zap.SugaredLogger
	cron                *cron.Cron
	appService          app.AppService
	workflowDagExecutor pipeline.WorkflowDagExecutor
	installedAppService service.InstalledAppService
	CdHandler           pipeline.CdHandler
}

const HelmAppStatusUpdateCronExpr string = "*/2 * * * *"

func NewHelmApplicationStatusUpdateHandlerImpl(logger *zap.SugaredLogger, appService app.AppService,
	workflowDagExecutor pipeline.WorkflowDagExecutor, installedAppService service.InstalledAppService,
	CdHandler pipeline.CdHandler) *HelmApplicationStatusUpdateHandlerImpl {
	cron := cron.New(
		cron.WithChain())
	cron.Start()
	impl := &HelmApplicationStatusUpdateHandlerImpl{
		logger:              logger,
		cron:                cron,
		appService:          appService,
		workflowDagExecutor: workflowDagExecutor,
		installedAppService: installedAppService,
		CdHandler:           CdHandler,
	}
	_, err := cron.AddFunc(HelmAppStatusUpdateCronExpr, impl.HelmApplicationStatusUpdate)
	if err != nil {
		logger.Errorw("error in starting helm application status update cron job", "err", err)
		return nil
	}
	return impl
}

func (impl *HelmApplicationStatusUpdateHandlerImpl) HelmApplicationStatusUpdate() {
	err := impl.CdHandler.CheckHelmAppStatusPeriodicallyAndUpdateInDb()
	if err != nil {
		impl.logger.Errorw("error helm app status update - cron job", "err", err)
		return
	}
	return
}
