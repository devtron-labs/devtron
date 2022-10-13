package cron

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	client2 "github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/pubsub"
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
}

type CdApplicationStatusUpdateHandlerImpl struct {
	logger                           *zap.SugaredLogger
	cron                             *cron.Cron
	appService                       app.AppService
	workflowDagExecutor              pipeline.WorkflowDagExecutor
	installedAppService              service.InstalledAppService
	CdHandler                        pipeline.CdHandler
	AppStatusConfig                  *AppStatusConfig
	pubsubClient                     *pubsub.PubSubClient
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository
	eventClient                      client2.EventClient
}

type AppStatusConfig struct {
	CdPipelineStatusCronTime string `env:"CD_PIPELINE_STATUS_CRON_TIME" envDefault:"*/2 * * * *"`
	PipelineDegradedTime     string `env:"PIPELINE_DEGRADED_TIME" envDefault:"10"` //in minutes
}

func GetAppStatusConfig() (*AppStatusConfig, error) {
	cfg := &AppStatusConfig{}
	err := env.Parse(cfg)
	if err != nil {
		fmt.Println("failed to parse server app status config: " + err.Error())
		return nil, err
	}
	return cfg, nil
}

func NewCdApplicationStatusUpdateHandlerImpl(logger *zap.SugaredLogger, appService app.AppService,
	workflowDagExecutor pipeline.WorkflowDagExecutor, installedAppService service.InstalledAppService,
	CdHandler pipeline.CdHandler, AppStatusConfig *AppStatusConfig, pubsubClient *pubsub.PubSubClient,
	pipelineStatusTimelineRepository pipelineConfig.PipelineStatusTimelineRepository) *CdApplicationStatusUpdateHandlerImpl {
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
	}
	err := impl.Subscribe()
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
		impl.logger.Infow("ARGO_PIPELINE_STATUS_UPDATE_REQ", "stage", "raw", "data", msg.Data)
		statusUpdateEvent := pipeline.ArgoPipelineStatusEvent{}
		err := json.Unmarshal([]byte(string(msg.Data)), &statusUpdateEvent)
		if err != nil {
			impl.logger.Errorw("unmarshal error on argo pipeline status update event", "err", err)
			return
		}

		err = impl.CdHandler.UpdatePipelineTimelineAndStatusByLiveResourceTreeFetch(statusUpdateEvent.ArgoAppName, statusUpdateEvent.AppId, statusUpdateEvent.EnvId, statusUpdateEvent.IgnoreFailedWorkflowStatus)
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
	err := impl.CdHandler.CheckArgoPipelineTimelineStatusPeriodicallyAndUpdateInDb(30)
	if err != nil {
		impl.logger.Errorw("error argo app status update - cron job", "err", err)
		return
	}
	return
}

func (impl *CdApplicationStatusUpdateHandlerImpl) SyncPipelineStatusForResourceTreeCall(acdAppName string, appId, envId int) error {
	_, err := impl.pipelineStatusTimelineRepository.FetchTimelineUpdatedBeforeSecondsByAppIdAndEnvId(appId, envId, 10)
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("error in getting timeline", "err", err)
		return err
	}
	if err == pg.ErrNoRows {
		//create new nats event
		statusUpdateEvent := pipeline.ArgoPipelineStatusEvent{
			ArgoAppName:                acdAppName,
			AppId:                      appId,
			EnvId:                      envId,
			IgnoreFailedWorkflowStatus: false,
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
