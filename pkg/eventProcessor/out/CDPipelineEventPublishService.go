package out

import (
	"encoding/json"
	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/pkg/eventProcessor/bean"
	"go.uber.org/zap"
)

type CDPipelineEventPublishService interface {
	PublishArgoTypePipelineSyncEvent(pipelineId, installedAppVersionId int,
		userId int32, isAppStoreApplication bool) error
}

type CDPipelineEventPublishServiceImpl struct {
	logger       *zap.SugaredLogger
	pubSubClient *pubsub.PubSubClientServiceImpl
}

func NewCDPipelineEventPublishServiceImpl(logger *zap.SugaredLogger,
	pubSubClient *pubsub.PubSubClientServiceImpl) *CDPipelineEventPublishServiceImpl {
	return &CDPipelineEventPublishServiceImpl{
		logger:       logger,
		pubSubClient: pubSubClient,
	}
}

func (impl *CDPipelineEventPublishServiceImpl) PublishArgoTypePipelineSyncEvent(pipelineId, installedAppVersionId int,
	userId int32, isAppStoreApplication bool) error {
	statusUpdateEvent := bean.ArgoPipelineStatusSyncEvent{
		PipelineId:            pipelineId,
		InstalledAppVersionId: installedAppVersionId,
		UserId:                userId,
		IsAppStoreApplication: isAppStoreApplication,
	}
	data, err := json.Marshal(statusUpdateEvent)
	if err != nil {
		impl.logger.Errorw("error while writing cd pipeline delete event to nats", "err", err, "req", statusUpdateEvent)
		return err
	} else {
		err = impl.pubSubClient.Publish(pubsub.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, string(data))
		if err != nil {
			impl.logger.Errorw("error, PublishArgoTypePipelineSyncEvent", "topic", pubsub.ARGO_PIPELINE_STATUS_UPDATE_TOPIC, "error", err, "data", data)
			return err
		}
	}
	return nil
}
