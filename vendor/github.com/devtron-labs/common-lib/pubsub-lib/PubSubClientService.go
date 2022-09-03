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
	Logger     *zap.SugaredLogger
	NatsClient *NatsClient
}

func NewPubSubClientServiceImpl(logger *zap.SugaredLogger) *PubSubClientServiceImpl {
	natsClient, err := NewNatsClient(logger)
	if err != nil {
		logger.Fatalw("error occurred while creating nats client stopping now!!")
	}
	pubSubClient := &PubSubClientServiceImpl{
		Logger:     logger,
		NatsClient: natsClient,
	}
	return pubSubClient
}

func (impl PubSubClientServiceImpl) Publish(topic string, msg string) error {
	natsClient := impl.NatsClient
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
		impl.Logger.Errorw("error while publishing message", "stream", streamName, "topic", topic, "error", err)
		return err
	}
	return nil
}

func (impl PubSubClientServiceImpl) Subscribe(topic string, callback func(msg *PubSubMsg)) error {
	natsTopic := GetNatsTopic(topic)
	streamName := natsTopic.streamName
	queueName := natsTopic.queueName
	consumerName := natsTopic.consumerName
	natsClient := impl.NatsClient
	streamConfig := natsClient.streamConfig
	_ = AddStream(natsClient.JetStrCtxt, streamConfig, streamName)
	deliveryOption := nats.DeliverLast()
	if streamConfig.Retention == nats.WorkQueuePolicy {
		deliveryOption = nats.DeliverAll()
	}
	processingBatchSize := natsClient.NatsMsgProcessingBatchSize
	msgBufferSize := natsClient.NatsMsgBufferSize
	channel := make(chan *nats.Msg, msgBufferSize)
	_, err := natsClient.JetStrCtxt.ChanQueueSubscribe(topic, queueName, channel, nats.Durable(consumerName), deliveryOption, nats.ManualAck(),
		nats.BindStream(streamName))
	if err != nil {
		impl.Logger.Fatalw("error while subscribing to nats ", "stream", streamName, "topic", topic, "error", err)
		return err
	}
	go impl.startListeningForEvents(processingBatchSize, channel, callback)
	impl.Logger.Infow("Successfully subscribed with Nats", "stream", streamName, "topic", topic, "queue", queueName, "consumer", consumerName)
	return nil
}

func (impl PubSubClientServiceImpl) startListeningForEvents(processingBatchSize int, channel chan *nats.Msg, callback func(msg *PubSubMsg)) {
	wg := new(sync.WaitGroup)

	for index := 0; index < processingBatchSize; index++ {
		wg.Add(1)
		go processMessages(wg, channel, callback)
	}
	wg.Wait()
	impl.Logger.Warn("msgs received Done from Nats side, going to end listening!!")
}

func processMessages(wg *sync.WaitGroup, channel chan *nats.Msg, callback func(msg *PubSubMsg)) {
	defer wg.Done()
	for msg := range channel {
		processMsg(msg, callback)
	}
}

//TODO need to extend msg ack depending upon response from callback like error scenario
func processMsg(msg *nats.Msg, callback func(msg *PubSubMsg)) {
	defer msg.Ack()
	subMsg := &PubSubMsg{Data: string(msg.Data)}
	callback(subMsg)
}
