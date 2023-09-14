package cron

import (
	"encoding/json"
	"fmt"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/api/bean"
	client2 "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/internal/middleware"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	repository2 "github.com/devtron-labs/devtron/pkg/appStore/deployment/repository"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type CdApplicationStatusUpdateHandler interface {
	HelmApplicationStatusUpdate()
	ArgoApplicationStatusUpdate()
	ArgoPipelineTimelineUpdate()
	Subscribe() error
	SyncPipelineStatusForResourceTreeCall(pipeline *pipelineConfig.Pipeline) error
	SyncPipelineStatusForAppStoreForResourceTreeCall(installedAppVersion *repository2.InstalledAppVersions) error
	ManualSyncPipelineStatus(appId, envId int, userId int32) error
}

type CdApplicationStatusUpdateHandlerImpl struct {
	logger                               *zap.SugaredLogger
	cron                                 *cron.Cron
	appService                           app.AppService
	workflowDagExecutor                  pipeline.WorkflowDagExecutor
	installedAppService                  service.InstalledAppService
	CdHandler                            pipeline.CdHandler
	AppStatusConfig                      *app.AppServiceConfig
	pubsubClient                         *pubsub.PubSubClientServiceImpl
	pipelineStatusTimelineRepository     pipelineConfig.PipelineStatusTimelineRepository
	eventClient                          client2.EventClient
	appListingRepository                 repository.AppListingRepository
	cdWorkflowRepository                 pipelineConfig.CdWorkflowRepository
	pipelineRepository                   pipelineConfig.PipelineRepository
	installedAppVersionHistoryRepository repository2.InstalledAppVersionHistoryRepository
	installedAppVersionRepository        repository2.InstalledAppRepository
}

type CronLoggerImpl struct {
	logger *zap.SugaredLogger
}

func (impl *CronLoggerImpl) Info(msg string, keysAndValues ...interface{}) {
	impl.logger.Infow(msg, keysAndValues...)
}

func (impl *CronLoggerImpl) Error(err error, msg string, keysAndValues ...interface{}) {
	keysAndValues = append([]interface{}{"err", err}, keysAndValues...)
	impl.logger.Errorw(msg, keysAndValues...)
}

func NewCdApplicationStatusUpdateHandlerImpl(logger *zap.SugaredLogger, appService app.AppService,
	workflowDagExecutor pipeline.WorkflowDagExecutor, installedAppService service.InstalledAppService,
	CdHandler pipeline.CdHandler, AppStatusConfig *app.AppServiceConfig, pubsubClient *pubsub.PubSubClientServiceImpl,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	eventClient client2.EventClient, appListingRepository repository.AppListingRepository,
	cdWorkflowRepository pipelineConfig.CdWorkflowRepository,
	pipelineRepository pipelineConfig.PipelineRepository, installedAppVersionHistoryRepository repository2.InstalledAppVersionHistoryRepository,
	installedAppVersionRepository repository2.InstalledAppRepository) *CdApplicationStatusUpdateHandlerImpl {
	cronLogger := &CronLoggerImpl{logger: logger}
	cron := cron.New(
		cron.WithChain(cron.SkipIfStillRunning(cronLogger)))
	cron.Start()
	impl := &CdApplicationStatusUpdateHandlerImpl{
		logger:                               logger,
		cron:                                 cron,
		appService:                           appService,
		workflowDagExecutor:                  workflowDagExecutor,
		installedAppService:                  installedAppService,
		CdHandler:                            CdHandler,
		AppStatusConfig:                      AppStatusConfig,
		pubsubClient:                         pubsubClient,
		pipelineStatusTimelineRepository:     pipelineStatusTimelineRepository,
		eventClient:                          eventClient,
		appListingRepository:                 appListingRepository,
		cdWorkflowRepository:                 cdWorkflowRepository,
		pipelineRepository:                   pipelineRepository,
		installedAppVersionHistoryRepository: installedAppVersionHistoryRepository,
		installedAppVersionRepository:        installedAppVersionRepository,
	}

	err := impl.Subscribe()
	if err != nil {
		logger.Errorw("error on subscribe", "err", err)
		return nil
	}
	_, err = cron.AddFunc(AppStatusConfig.CdHelmPipelineStatusCronTime, impl.HelmApplicationStatusUpdate)
	if err != nil {
		logger.Errorw("error in starting helm application status update cron job", "err", err)
		return nil
	}
	_, err = cron.AddFunc(AppStatusConfig.CdPipelineStatusCronTime, impl.ArgoApplicationStatusUpdate)
	if err != nil {
		logger.Errorw("error in starting argo application status update cron job", "err", err)
		return nil
	}
	_, err = cron.AddFunc("@every 1m", impl.ArgoPipelineTimelineUpdate)
	if err != nil {
		logger.Errorw("error in starting argo application status update cron job", "err", err)
		return nil
	}
	return impl
}

func (impl *CdApplicationStatusUpdateHandlerImpl) Subscribe() error {
	callback := func(msg *pubsub.PubSubMsg) {
		statusUpdateEvent := pipeline.ArgoPipelineStatusSyncEvent{}
		var err error
		var cdPipeline *pipelineConfig.Pipeline
		var installedApp repository2.InstalledApps

		err = json.Unmarshal([]byte(string(msg.Data)), &statusUpdateEvent)
		if err != nil {
			impl.logger.Errorw("unmarshal error on argo pipeline status update event", "err", err)
			return
		}
		impl.logger.Debugw("ARGO_PIPELINE_STATUS_UPDATE_REQ", "stage", "subscribeDataUnmarshal", "data", statusUpdateEvent)

		if statusUpdateEvent.IsAppStoreApplication {
			installedApp, err = impl.installedAppVersionRepository.GetInstalledAppByInstalledAppVersionId(statusUpdateEvent.PipelineId)
			if err != nil {
				impl.logger.Errorw("error in getting installedAppVersion by id", "err", err, "id", statusUpdateEvent.PipelineId)
				return
			}
		} else {
			cdPipeline, err = impl.pipelineRepository.FindById(statusUpdateEvent.PipelineId)
			if err != nil {
				impl.logger.Errorw("error in getting cdPipeline by id", "err", err, "id", statusUpdateEvent.PipelineId)
				return
			}
		}

		err, _ = impl.CdHandler.UpdatePipelineTimelineAndStatusByLiveApplicationFetch(cdPipeline, installedApp, statusUpdateEvent.UserId)
		if err != nil {
			impl.logger.Errorw("error on argo pipeline status update", "err", err, "msg", string(msg.Data))
			return
		}
	}
	err := impl.pubsubClient.Subscribe(pubsub.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, callback)
	if err != nil {
		impl.logger.Errorw("error in subscribing to argo application status update topic", "err", err)
		return err
	}
	return nil
}

func (impl *CdApplicationStatusUpdateHandlerImpl) HelmApplicationStatusUpdate() {
	cronProcessStartTime := time.Now()
	defer func() {
		middleware.DeploymentStatusCronDuration.WithLabelValues(pipeline.DEVTRON_APP_HELM_PIPELINE_STATUS_UPDATE_CRON).Observe(time.Since(cronProcessStartTime).Seconds())
	}()
	HelmPipelineStatusCheckEligibleTime, err := strconv.Atoi(impl.AppStatusConfig.HelmPipelineStatusCheckEligibleTime)
	if err != nil {
		impl.logger.Errorw("error in converting string to int", "err", err)
		return
	}
	getPipelineDeployedWithinHours := impl.AppStatusConfig.GetPipelineDeployedWithinHours
	err = impl.CdHandler.CheckHelmAppStatusPeriodicallyAndUpdateInDb(HelmPipelineStatusCheckEligibleTime, getPipelineDeployedWithinHours)
	if err != nil {
		impl.logger.Errorw("error helm app status update - cron job", "err", err)
		return
	}
	return
}

func (impl *CdApplicationStatusUpdateHandlerImpl) ArgoApplicationStatusUpdate() {
	cronProcessStartTime := time.Now()
	defer func() {
		middleware.DeploymentStatusCronDuration.WithLabelValues(pipeline.DEVTRON_APP_ARGO_PIPELINE_STATUS_UPDATE_CRON).Observe(time.Since(cronProcessStartTime).Seconds())
	}()
	//TODO: remove below cron with division of cron for argo pipelines of devtron-apps and helm-apps
	defer func() {
		middleware.DeploymentStatusCronDuration.WithLabelValues(pipeline.HELM_APP_ARGO_PIPELINE_STATUS_UPDATE_CRON).Observe(time.Since(cronProcessStartTime).Seconds())
	}()
	getPipelineDeployedBeforeMinutes, err := strconv.Atoi(impl.AppStatusConfig.PipelineDegradedTime)
	if err != nil {
		impl.logger.Errorw("error in converting string to int", "err", err)
		return
	}
	getPipelineDeployedWithinHours := impl.AppStatusConfig.GetPipelineDeployedWithinHours
	err = impl.CdHandler.CheckArgoAppStatusPeriodicallyAndUpdateInDb(getPipelineDeployedBeforeMinutes, getPipelineDeployedWithinHours)
	if err != nil {
		impl.logger.Errorw("error argo app status update - cron job", "err", err)
		return
	}
	return
}

func (impl *CdApplicationStatusUpdateHandlerImpl) ArgoPipelineTimelineUpdate() {
	degradedTime, err := strconv.Atoi(impl.AppStatusConfig.PipelineDegradedTime)
	if err != nil {
		impl.logger.Errorw("error in converting string to int", "err", err)
		return
	}
	err = impl.CdHandler.CheckArgoPipelineTimelineStatusPeriodicallyAndUpdateInDb(30, degradedTime)
	if err != nil {
		impl.logger.Errorw("error argo app status update - cron job", "err", err)
		return
	}
	return
}

func (impl *CdApplicationStatusUpdateHandlerImpl) SyncPipelineStatusForResourceTreeCall(pipeline *pipelineConfig.Pipeline) error {
	cdWfr, err := impl.cdWorkflowRepository.FindLastStatusByPipelineIdAndRunnerType(pipeline.Id, bean.CD_WORKFLOW_TYPE_DEPLOY)
	if err != nil {
		impl.logger.Errorw("error in getting latest cdWfr by cdPipelineId", "err", err, "pipelineId", pipeline.Id)
		return nil
	}
	if !util.IsTerminalStatus(cdWfr.Status) {
		impl.CdHandler.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(pipeline.Id, 1, false)
	}
	return nil
}

func (impl *CdApplicationStatusUpdateHandlerImpl) SyncPipelineStatusForAppStoreForResourceTreeCall(installedAppVersion *repository2.InstalledAppVersions) error {
	//find installed app version history using parameter obj
	installedAppVersionHistory, err := impl.installedAppVersionHistoryRepository.GetLatestInstalledAppVersionHistory(installedAppVersion.Id)
	if err != nil {
		impl.logger.Errorw("error in getting latest installedAppVersionHistory by installedAppVersionId", "err", err, "installedAppVersionId", installedAppVersion.Id)
		return nil
	}
	if !util.IsTerminalStatus(installedAppVersionHistory.Status) {
		impl.CdHandler.CheckAndSendArgoPipelineStatusSyncEventIfNeeded(installedAppVersion.Id, 1, true)
	}
	return nil
}

func (impl *CdApplicationStatusUpdateHandlerImpl) ManualSyncPipelineStatus(appId, envId int, userId int32) error {
	var cdPipeline *pipelineConfig.Pipeline
	var installedApp repository2.InstalledApps
	var err error

	if envId == 0 {
		installedApp, err = impl.installedAppVersionRepository.GetInstalledAppByAppId(appId)
		if err != nil {
			impl.logger.Errorw("error in getting installed app by appId", "err", err, "appid", appId)
			return nil
		}

	} else {
		cdPipelines, err := impl.pipelineRepository.FindActiveByAppIdAndEnvironmentId(appId, envId)
		if err != nil {
			impl.logger.Errorw("error in getting cdPipeline by appId and envId", "err", err, "appid", appId, "envId", envId)
			return nil
		}
		if len(cdPipelines) != 1 {
			return fmt.Errorf("invalid number of cd pipelines found")
		}
		cdPipeline = cdPipelines[0]
	}

	err, isTimelineUpdated := impl.CdHandler.UpdatePipelineTimelineAndStatusByLiveApplicationFetch(cdPipeline, installedApp, userId)
	if err != nil {
		impl.logger.Errorw("error on argo pipeline status update", "err", err)
		return nil
	}
	if !isTimelineUpdated {
		return fmt.Errorf("timeline unchanged")
	}

	return nil
}
