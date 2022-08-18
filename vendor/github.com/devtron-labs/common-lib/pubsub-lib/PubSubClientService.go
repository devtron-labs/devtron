package pubsub_lib

import (
	"github.com/devtron-labs/common-lib/utils"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type PubSubClientService interface {
	Publish(topic string, msg string) error
	Subscribe(topic string, callback func(msg *PubSubMsg)) error
}

type PubSubMsg struct {
	Data string
}

type PubSubClientServiceImpl struct {
	logger     *zap.SugaredLogger
	natsClient *NatsClient
}

func NewPubSubClientServiceImpl(logger *zap.SugaredLogger) *PubSubClientServiceImpl {
	natsClient, err := NewNatsClient(logger)
	if err != nil {
		logger.Fatalw("error occurred while creating nats client stopping now!!")
	}
	pubSubClient := &PubSubClientServiceImpl{
		logger:     logger,
		natsClient: natsClient,
	}
	return pubSubClient
}

func (impl PubSubClientServiceImpl) Publish(topic string, msg string) error {
	natsClient := impl.natsClient
	jetStrCtxt := natsClient.JetStrCtxt
	natsTopic := GetNatsTopic(topic)
	streamName := natsTopic.streamName
	streamConfig := natsClient.streamConfig
	_ = AddStream(jetStrCtxt, streamConfig, streamName)
	//Generate random string for passing as Header Id in message
	randString := "MsgHeaderId-" + utils.Generate(10)
	_, err := jetStrCtxt.Publish(topic, []byte(msg), nats.MsgId(randString))
	if err != nil {
		//TODO need to handle retry specially for timeout cases
		impl.logger.Errorw("error while publishing message", "stream", streamName, "topic", topic, "error", err)
		return err
	}
	return nil
}

func (impl PubSubClientServiceImpl) Subscribe(topic string, callback func(msg *PubSubMsg)) error {
	natsTopic := GetNatsTopic(topic)
	streamName := natsTopic.streamName
	queueName := natsTopic.queueName
	consumerName := natsTopic.consumerName
	natsClient := impl.natsClient
	streamConfig := natsClient.streamConfig
	_ = AddStream(natsClient.JetStrCtxt, streamConfig, streamName)
	deliveryOption := nats.DeliverLast()
	if streamConfig.Retention == nats.WorkQueuePolicy {
		deliveryOption = nats.DeliverAll()
	}
	_, err := natsClient.JetStrCtxt.QueueSubscribe(topic, queueName, func(msg *nats.Msg) {
		defer msg.Ack()
		subMsg := &PubSubMsg{Data: string(msg.Data)}
		callback(subMsg)
	}, nats.Durable(consumerName), deliveryOption, nats.ManualAck(), nats.BindStream(streamName))
	if err != nil {
		impl.logger.Fatalw("error while subscribing", "stream", streamName, "topic", topic, "error", err)
		return err
	}

	return nil
}
