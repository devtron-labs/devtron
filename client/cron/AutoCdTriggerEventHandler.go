package cron

import (
	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"time"
)

type AutoCdTriggerEventHandler interface {
	SubscribeAutoCdTriggerEventHandler() error
}

type AutoCdTriggerEventHandlerImpl struct {
	logger               *zap.SugaredLogger
	pubsubClient         *pubsub.PubSubClient
	workflowDagExecutor  pipeline.WorkflowDagExecutor
	pipelineRepository   pipelineConfig.PipelineRepository
	ciArtifactRepository repository.CiArtifactRepository
	config               *AutoCdTriggerHandlerConfig
}

type AutoCdTriggerHandlerConfig struct {
	AutoCdTriggerConsumerAckWaitInSecs int `env:"AUTO_CD_TRIGGER_ACK_WAIT_IN_SECS" envDefault:"150"`
}

func NewAutoCdTriggerEventHandlerImpl(logger *zap.SugaredLogger, pubsubClient *pubsub.PubSubClient, workflowDagExecutor pipeline.WorkflowDagExecutor,
	pipelineRepository pipelineConfig.PipelineRepository, ciArtifactRepository repository.CiArtifactRepository) (*AutoCdTriggerEventHandlerImpl, error) {
	config := &AutoCdTriggerHandlerConfig{}
	err := env.Parse(config)
	if err != nil {
		logger.Errorw("error occurred while parsing config", "err", err)
		return nil, err
	}
	impl := &AutoCdTriggerEventHandlerImpl{
		logger:               logger,
		pubsubClient:         pubsubClient,
		workflowDagExecutor:  workflowDagExecutor,
		pipelineRepository:   pipelineRepository,
		ciArtifactRepository: ciArtifactRepository,
		config:               config,
	}
	err = util.AddStream(impl.pubsubClient.JetStrCtxt, util.ORCHESTRATOR_STREAM)
	if err != nil {
		logger.Errorw("error while adding stream", "streamName", util.ORCHESTRATOR_STREAM, "err", err)
		return nil, err
	}
	err = impl.SubscribeAutoCdTriggerEventHandler()
	if err != nil {
		logger.Errorw("error while subscribing", "topic", util.AUTO_TRIGGER_STAGES_AFTER_CI_COMPLETE_TOPIC, "err", err)
		return nil, err
	}
	return impl, nil
}

func (impl *AutoCdTriggerEventHandlerImpl) SubscribeAutoCdTriggerEventHandler() error {
	_, err := impl.pubsubClient.JetStrCtxt.QueueSubscribe(util.AUTO_TRIGGER_STAGES_AFTER_CI_COMPLETE_TOPIC, util.AUTO_TRIGGER_STAGES_AFTER_CI_COMPLETE_GROUP, func(msg *nats.Msg) {
		impl.logger.Debug("received auto trigger stages event after ci completion")
		defer msg.AckSync()
		eventPayload := pipeline.AutoTriggerStagesAfterCiCompleteEvent{}
		err := json.Unmarshal([]byte(string(msg.Data)), &eventPayload)
		if err != nil {
			impl.logger.Errorw("unmarshal error on auto trigger stages event event", "err", err)
			return
		}
		// get pipeline
		pipelineId := eventPayload.PipelineId
		pipeline, err := impl.pipelineRepository.FindById(pipelineId)
		if err != nil {
			impl.logger.Errorw("error while getting pipeline from DB", "pipelineId", pipelineId, "err", err)
			return
		}

		// get ciArtifact
		ciArtifactId := eventPayload.CiArtifactId
		ciArtifact, err := impl.ciArtifactRepository.Get(ciArtifactId)
		if err != nil {
			impl.logger.Errorw("error while getting ci-artifact from DB", "ciArtifactId", ciArtifactId, "err", err)
			return
		}

		err = impl.workflowDagExecutor.TriggerStage(pipeline, ciArtifact, eventPayload.ApplyAuth, eventPayload.TriggeredBy)
		if err != nil {
			impl.logger.Errorw("error on triggering stages", "err", err, "msg", string(msg.Data))
			return
		}
	}, nats.AckWait(time.Duration(impl.config.AutoCdTriggerConsumerAckWaitInSecs)*time.Second), nats.Durable(util.AUTO_TRIGGER_STAGES_AFTER_CI_COMPLETE_DURABLE), nats.DeliverLast(), nats.ManualAck(), nats.BindStream(util.ORCHESTRATOR_STREAM))
	if err != nil {
		impl.logger.Errorw("error in subscribing", "topic", util.AUTO_TRIGGER_STAGES_AFTER_CI_COMPLETE_TOPIC, "err", err)
		return err
	}
	return err
}
