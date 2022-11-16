package cron

import (
	"encoding/json"
	"fmt"
	client2 "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStore/deployment/service"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util"
	"github.com/go-pg/pg"
	"github.com/nats-io/nats.go"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"strconv"
)

type CdApplicationStatusUpdateHandler interface {
	HelmApplicationStatusUpdate()
	ArgoApplicationStatusUpdate()
	ArgoPipelineTimelineUpdate()
	Subscribe() error
	SyncPipelineStatusForResourceTreeCall(acdAppName string, appId, envId int) error
	ManualSyncPipelineStatus(appId, envId int, userId int32) error
}

type CdApplicationStatusUpdateHandlerImpl struct {
	logger                           *zap.SugaredLogger
	cron                             *cron.Cron
	appService                       app.AppService
	workflowDagExecutor              pipeline.WorkflowDagExecutor
	installedAppService              service.InstalledAppService
	CdHandler                        pipeline.CdHandler
	AppStatusConfig                  *app.AppStatusConfig
	pubsubClient                     *pubsub.PubSubClient
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository
	eventClient                      client2.EventClient
	appListingRepository             repository.AppListingRepository
}

func NewCdApplicationStatusUpdateHandlerImpl(logger *zap.SugaredLogger, appService app.AppService,
	workflowDagExecutor pipeline.WorkflowDagExecutor, installedAppService service.InstalledAppService,
	CdHandler pipeline.CdHandler, AppStatusConfig *app.AppStatusConfig, pubsubClient *pubsub.PubSubClient,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository,
	eventClient client2.EventClient, appListingRepository repository.AppListingRepository) *CdApplicationStatusUpdateHandlerImpl {
	cron := cron.New(
		cron.WithChain())
	cron.Start()
	impl := &CdApplicationStatusUpdateHandlerImpl{
		logger:                           logger,
		cron:                             cron,
		appService:                       appService,
		workflowDagExecutor:              workflowDagExecutor,
		installedAppService:              installedAppService,
		CdHandler:                        CdHandler,
		AppStatusConfig:                  AppStatusConfig,
		pubsubClient:                     pubsubClient,
		pipelineStatusTimelineRepository: pipelineStatusTimelineRepository,
		eventClient:                      eventClient,
		appListingRepository:             appListingRepository,
	}
	err := util.AddStream(pubsubClient.JetStrCtxt, util.ORCHESTRATOR_STREAM)
	if err != nil {
		return nil
	}
	err = impl.Subscribe()
	if err != nil {
		logger.Errorw("error on subscribe", "err", err)
		return nil
	}
	_, err = cron.AddFunc(AppStatusConfig.CdPipelineStatusCronTime, impl.HelmApplicationStatusUpdate)
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
	_, err := impl.pubsubClient.JetStrCtxt.QueueSubscribe(util.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, util.ARGO_PIPELINE_STATUS_UPDATE_GROUP, func(msg *nats.Msg) {
		impl.logger.Debug("received argo pipeline status update request")
		defer msg.Ack()
		statusUpdateEvent := pipeline.ArgoPipelineStatusSyncEvent{}
		err := json.Unmarshal([]byte(string(msg.Data)), &statusUpdateEvent)
		if err != nil {
			impl.logger.Errorw("unmarshal error on argo pipeline status update event", "err", err)
			return
		}
		impl.logger.Infow("ARGO_PIPELINE_STATUS_UPDATE_REQ", "stage", "subscribeDataUnmarshal", "data", statusUpdateEvent)
		err, _ = impl.CdHandler.UpdatePipelineTimelineAndStatusByLiveResourceTreeFetch(statusUpdateEvent.ArgoAppName, statusUpdateEvent.UserId)
		if err != nil {
			impl.logger.Errorw("error on argo pipeline status update", "err", err, "msg", string(msg.Data))
			return
		}
	}, nats.Durable(util.ARGO_PIPELINE_STATUS_UPDATE_DURABLE), nats.DeliverLast(), nats.ManualAck(), nats.BindStream(util.ORCHESTRATOR_STREAM))
	if err != nil {
		impl.logger.Error("error in subscribing to argo application status update topic", "err", err)
		return err
	}
	return nil
}

func (impl *CdApplicationStatusUpdateHandlerImpl) HelmApplicationStatusUpdate() {
	degradedTime, err := strconv.Atoi(impl.AppStatusConfig.PipelineDegradedTime)
	if err != nil {
		impl.logger.Errorw("error in converting string to int", "err", err)
		return
	}
	err = impl.CdHandler.CheckHelmAppStatusPeriodicallyAndUpdateInDb(degradedTime)
	if err != nil {
		impl.logger.Errorw("error helm app status update - cron job", "err", err)
		return
	}
	return
}

func (impl *CdApplicationStatusUpdateHandlerImpl) ArgoApplicationStatusUpdate() {
	degradedTime, err := strconv.Atoi(impl.AppStatusConfig.PipelineDegradedTime)
	if err != nil {
		impl.logger.Errorw("error in converting string to int", "err", err)
		return
	}

	err = impl.CdHandler.CheckArgoAppStatusPeriodicallyAndUpdateInDb(degradedTime)
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

func (impl *CdApplicationStatusUpdateHandlerImpl) SyncPipelineStatusForResourceTreeCall(acdAppName string, appId, envId int) error {
	deploymentStatus, err := impl.appListingRepository.FindLastDeployedStatusForAcdPipelineByAppIdAndEnvId(appId, envId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching deployment status", "err", err, "appId", appId, "envId", envId)
		return err
	}
	if !util.IsTerminalStatus(deploymentStatus.Status) {
		//create new nats event
		statusUpdateEvent := pipeline.ArgoPipelineStatusSyncEvent{
			ArgoAppName: acdAppName,
			UserId:      1,
		}
		//write event
		err := impl.eventClient.WriteNatsEvent(util.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, statusUpdateEvent)
		if err != nil {
			impl.logger.Errorw("error in writing nats event", "topic", util.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, "payload", statusUpdateEvent)
			return err
		}
	}
	return nil
}

func (impl *CdApplicationStatusUpdateHandlerImpl) ManualSyncPipelineStatus(appId, envId int, userId int32) error {
	deploymentStatus, err := impl.appListingRepository.FindLastDeployedStatusForAcdPipelineByAppIdAndEnvId(appId, envId)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in fetching deployment status", "err", err, "appId", appId, "envId", envId)
		return err
	}
	if !util.IsTerminalStatus(deploymentStatus.Status) {
		err, isTimelineUpdated := impl.CdHandler.UpdatePipelineTimelineAndStatusByLiveResourceTreeFetch(deploymentStatus.AppName, userId)
		if err != nil {
			impl.logger.Errorw("error on argo pipeline status update", "err", err)
			return nil
		}
		if !isTimelineUpdated {
			return fmt.Errorf("timeline unchanged")
		}
	}
	return nil
}
