package pubsub_lib

import (
	"encoding/json"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib/natsMetrics"
	"github.com/devtron-labs/common-lib/utils"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"sync"
	"time"
)

type PubSubClientService interface {
	Publish(topic string, msg string) error
	Subscribe(topic string, callback func(msg *PubSubMsg)) error
}

type PubSubMsg struct {
	Data string
}

type LogsConfig struct {
	DefaultLogTimeLimit int64 `env:"DEFAULT_LOG_TIME_LIMIT" envDefault:"1"`
}

type PubSubClientServiceImpl struct {
	Logger     *zap.SugaredLogger
	NatsClient *NatsClient
	logsConfig *LogsConfig
}

func NewPubSubClientServiceImpl(logger *zap.SugaredLogger) *PubSubClientServiceImpl {
	natsClient, err := NewNatsClient(logger)
	if err != nil {
		logger.Fatalw("error occurred while creating nats client stopping now!!")
	}
	logsConfig := &LogsConfig{}
	err = env.Parse(logsConfig)
	if err != nil {
		logger.Errorw("error occurred while parsing LogsConfig", "err", err)
	}
	ParseAndFillStreamWiseAndConsumerWiseConfigMaps()
	pubSubClient := &PubSubClientServiceImpl{
		Logger:     logger,
		NatsClient: natsClient,
		logsConfig: logsConfig,
	}
	return pubSubClient
}

func (impl PubSubClientServiceImpl) Publish(topic string, msg string) error {
	impl.Logger.Debugw("Published message on pubsub client", "topic", topic, "msg", msg)
	defer natsMetrics.IncPublishCount(topic)
	natsClient := impl.NatsClient
	jetStrCtxt := natsClient.JetStrCtxt
	natsTopic := GetNatsTopic(topic)
	streamName := natsTopic.streamName
	streamConfig := impl.getStreamConfig(streamName)
	//streamConfig := natsClient.streamConfig
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
	impl.Logger.Infow("Subscribed to pubsub client", "topic", topic)
	natsTopic := GetNatsTopic(topic)
	streamName := natsTopic.streamName
	queueName := natsTopic.queueName
	consumerName := natsTopic.consumerName
	natsClient := impl.NatsClient
	streamConfig := impl.getStreamConfig(streamName)
	//streamConfig := natsClient.streamConfig
	_ = AddStream(natsClient.JetStrCtxt, streamConfig, streamName)
	deliveryOption := nats.DeliverLast()
	if streamConfig.Retention == nats.WorkQueuePolicy {
		deliveryOption = nats.DeliverAll()
	}
	processingBatchSize := NatsConsumerWiseConfigMapping[consumerName].NatsMsgProcessingBatchSize
	msgBufferSize := NatsConsumerWiseConfigMapping[consumerName].NatsMsgBufferSize

	// Converting provided ack wait (int) into duration for comparing with nats-server config
	ackWait := time.Duration(NatsConsumerWiseConfigMapping[consumerName].AckWaitInSecs) * time.Second

	// Get the current Consumer config from NATS-server
	info, err := natsClient.JetStrCtxt.ConsumerInfo(streamName, consumerName)

	if err != nil {
		impl.Logger.Errorw("unable to retrieve consumer info from NATS-server",
			"stream", streamName,
			"consumer", consumerName,
			"err", err)

	} else {
		// Update NATS Consumer config if new changes detected
		// Currently only checking for AckWait, but can be done for other editable properties as well

		if ackWait > 0 && info.Config.AckWait != ackWait {

			updatedConfig := info.Config
			updatedConfig.AckWait = ackWait

			_, err = natsClient.JetStrCtxt.UpdateConsumer(streamName, &updatedConfig)

			if err != nil {
				impl.Logger.Errorw("failed to update Consumer config",
					"received consumer config", info.Config,
					"err", err)
			}
		}
	}

	channel := make(chan *nats.Msg, msgBufferSize)
	_, err = natsClient.JetStrCtxt.ChanQueueSubscribe(topic, queueName, channel, nats.Durable(consumerName), deliveryOption, nats.ManualAck(),
		nats.BindStream(streamName))
	if err != nil {
		impl.Logger.Fatalw("error while subscribing to nats ", "stream", streamName, "topic", topic, "error", err)
		return err
	}
	go impl.startListeningForEvents(processingBatchSize, channel, callback, topic)
	impl.Logger.Infow("Successfully subscribed with Nats", "stream", streamName, "topic", topic, "queue", queueName, "consumer", consumerName)
	return nil
}

func (impl PubSubClientServiceImpl) startListeningForEvents(processingBatchSize int, channel chan *nats.Msg, callback func(msg *PubSubMsg), topic string) {
	wg := new(sync.WaitGroup)

	for index := 0; index < processingBatchSize; index++ {
		wg.Add(1)
		go impl.processMessages(wg, channel, callback, topic)
	}
	wg.Wait()
	impl.Logger.Warn("msgs received Done from Nats side, going to end listening!!")
}

func (impl PubSubClientServiceImpl) processMessages(wg *sync.WaitGroup, channel chan *nats.Msg, callback func(msg *PubSubMsg), topic string) {
	defer wg.Done()
	for msg := range channel {
		impl.processMsg(msg, callback, topic)
	}
}

// TODO need to extend msg ack depending upon response from callback like error scenario
func (impl PubSubClientServiceImpl) processMsg(msg *nats.Msg, callback func(msg *PubSubMsg), topic string) {
	timeLimitInMillSecs := impl.logsConfig.DefaultLogTimeLimit * 1000
	t1 := time.Now()
	defer natsMetrics.IncConsumptionCount(topic)
	defer impl.printTimeDiff(t1, msg, timeLimitInMillSecs)
	defer msg.Ack()
	subMsg := &PubSubMsg{Data: string(msg.Data)}
	callback(subMsg)
}

func (impl PubSubClientServiceImpl) printTimeDiff(t0 time.Time, msg *nats.Msg, timeLimitInMillSecs int64) {
	t1 := time.Since(t0)
	if t1.Milliseconds() > timeLimitInMillSecs {
		impl.Logger.Debugw("time took to process msg: ", msg, "time :", t1)
	}
}
func (impl PubSubClientServiceImpl) getStreamConfig(streamName string) *nats.StreamConfig {
	configJson := NatsStreamWiseConfigMapping[streamName].StreamConfig
	streamCfg := &nats.StreamConfig{}
	data, err := json.Marshal(configJson)
	if err == nil {
		err = json.Unmarshal(data, streamCfg)
		if err != nil {
			impl.Logger.Errorw("error occurred while parsing streamConfigJson ", "streamCfg", streamCfg, "reason", err)
		}
	} else {
		impl.Logger.Errorw("error occurred while parsing streamConfigJson ", "configJson", configJson, "reason", err)
	}

	return streamCfg
}
