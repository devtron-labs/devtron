package pubsub_lib

import (
	"github.com/devtron-labs/common-lib/utils"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"sync"
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
	processingBatchSize := natsClient.NatsMsgProcessingBatchSize
	channel := make(chan *nats.Msg, 64)
	_, err := natsClient.JetStrCtxt.ChanQueueSubscribe(topic, queueName, channel, nats.Durable(consumerName), deliveryOption, nats.ManualAck(),
		nats.BindStream(streamName))
	if err != nil {
		impl.logger.Fatalw("error while subscribing", "stream", streamName, "topic", topic, "error", err)
		return err
	}
	go impl.startListeningForEvents(processingBatchSize, channel, callback)
	return nil
}

func (impl PubSubClientServiceImpl) startListeningForEvents(processingBatchSize int, channel chan *nats.Msg, callback func(msg *PubSubMsg)) {
	wg := new(sync.WaitGroup)
	wg.Add(processingBatchSize)
	index := 0
	for msg := range channel {
		go processMsg(wg, msg, callback)
		index++
		if index == processingBatchSize {
			wg.Wait()
			wg = new(sync.WaitGroup)
			wg.Add(processingBatchSize)
			index = 0
		}
	}
}

func processMsg(wg *sync.WaitGroup, msg *nats.Msg, callback func(msg *PubSubMsg)) {
	defer completeState(wg, msg)
	subMsg := &PubSubMsg{Data: string(msg.Data)}
	callback(subMsg)
}

func completeState(wg *sync.WaitGroup, msg *nats.Msg) {
	msg.Ack()
	wg.Done()
}
