package pubsub_lib

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/devtron-labs/common-lib-private/utils"
	"github.com/devtron-labs/common-lib-private/pubsub-lib/metrics"
	"github.com/devtron-labs/common-lib-private/pubsub-lib/model"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"runtime/debug"
	"sync"
	"time"
)

const NATS_MSG_LOG_PREFIX = "NATS_LOG"

type ValidateMsg func(msg model.PubSubMsg) bool

// LoggerFunc is used to log the message before passing to callback function.
// it expects logg message and key value pairs to be returned.
// if keysAndValues is empty, it will log whole model.PubSubMsg
type LoggerFunc func(msg model.PubSubMsg) (logMsg string, keysAndValues []interface{})

type PubSubClientService interface {
	Publish(topic string, msg string) error
	Subscribe(topic string, callback func(msg *model.PubSubMsg), loggerFunc LoggerFunc, validations ...ValidateMsg) error
}

type PubSubClientServiceImpl struct {
	Logger     *zap.SugaredLogger
	NatsClient *NatsClient
	logsConfig *model.LogsConfig
}

func NewPubSubClientServiceImpl(logger *zap.SugaredLogger) *PubSubClientServiceImpl {
	natsClient, err := NewNatsClient(logger)
	if err != nil {
		logger.Fatalw("error occurred while creating nats client stopping now!!")
	}
	logsConfig := &model.LogsConfig{}
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
	status := model.PUBLISH_FAILURE
	defer metrics.IncPublishCount(topic, status)
	natsClient := impl.NatsClient
	jetStrCtxt := natsClient.JetStrCtxt
	natsTopic := GetNatsTopic(topic)
	streamName := natsTopic.streamName
	streamConfig := impl.getStreamConfig(streamName)
	// streamConfig := natsClient.streamConfig
	_ = AddStream(jetStrCtxt, streamConfig, streamName)
	// Generate random string for passing as Header Id in message
	randString := "MsgHeaderId-" + utils.Generate(10)

	// track time taken to publish msg to nats server
	t1 := time.Now()
	defer func() {
		// wrapping this function in defer as directly calling Observe() will run immediately
		metrics.NatsEventPublishTime.WithLabelValues(topic).Observe(float64(time.Since(t1).Milliseconds()))
	}()

	_, err := jetStrCtxt.Publish(topic, []byte(msg), nats.MsgId(randString))
	if err != nil {
		// TODO need to handle retry specially for timeout cases
		impl.Logger.Errorw("error while publishing message", "stream", streamName, "topic", topic, "error", err)
		return err
	}

	// if reached here, means publish was successful
	status = model.PUBLISH_SUCCESS
	return nil
}

// Subscribe method is used to subscribe to the given topic(+required),
// this creates blocking process to continuously fetch messages from nats server published on this topic.
// invokes callback(+required) func for each message received.
// loggerFunc(+optional) is invoked before passing the message to the callback function.
// validations(+optional) methods were called before passing the message to the callback func.
func (impl PubSubClientServiceImpl) Subscribe(topic string, callback func(msg *model.PubSubMsg), loggerFunc LoggerFunc, validations ...ValidateMsg) error {
	impl.Logger.Infow("Subscribed to pubsub client", "topic", topic)
	natsTopic := GetNatsTopic(topic)
	streamName := natsTopic.streamName
	queueName := natsTopic.queueName
	consumerName := natsTopic.consumerName
	natsClient := impl.NatsClient
	streamConfig := impl.getStreamConfig(streamName)
	// streamConfig := natsClient.streamConfig
	_ = AddStream(natsClient.JetStrCtxt, streamConfig, streamName)
	deliveryOption := nats.DeliverLast()
	if streamConfig.Retention == nats.WorkQueuePolicy {
		deliveryOption = nats.DeliverAll()
	}

	consumerConfig := NatsConsumerWiseConfigMapping[consumerName]
	processingBatchSize := consumerConfig.NatsMsgProcessingBatchSize
	msgBufferSize := consumerConfig.GetNatsMsgBufferSize()

	// Converting provided ack wait (int) into duration for comparing with nats-server config
	ackWait := time.Duration(consumerConfig.AckWaitInSecs) * time.Second

	// Update consumer config if new changes detected
	impl.updateConsumer(natsClient, streamName, consumerName, &consumerConfig)

	channel := make(chan *nats.Msg, msgBufferSize)
	_, err := natsClient.JetStrCtxt.ChanQueueSubscribe(topic, queueName, channel,
		nats.Durable(consumerName),
		deliveryOption,
		nats.ManualAck(),
		nats.AckWait(ackWait), // if ackWait is 0 , nats sets this option to 30secs by default
		nats.BindStream(streamName))
	if err != nil {
		impl.Logger.Fatalw("error while subscribing to nats ", "stream", streamName, "topic", topic, "error", err)
		return err
	}
	go impl.startListeningForEvents(processingBatchSize, channel, callback, loggerFunc, validations...)
	impl.Logger.Infow("Successfully subscribed with Nats", "stream", streamName, "topic", topic, "queue", queueName, "consumer", consumerName)
	return nil
}

func (impl PubSubClientServiceImpl) startListeningForEvents(processingBatchSize int, channel chan *nats.Msg, callback func(msg *model.PubSubMsg), loggerFunc LoggerFunc, validations ...ValidateMsg) {
	wg := new(sync.WaitGroup)

	for index := 0; index < processingBatchSize; index++ {
		wg.Add(1)
		go impl.processMessages(wg, channel, callback, loggerFunc, validations...)
	}
	wg.Wait()
	impl.Logger.Warn("msgs received Done from Nats side, going to end listening!!")
}

func (impl PubSubClientServiceImpl) processMessages(wg *sync.WaitGroup, channel chan *nats.Msg, callback func(msg *model.PubSubMsg), loggerFunc LoggerFunc, validations ...ValidateMsg) {
	defer wg.Done()
	for msg := range channel {
		impl.processMsg(msg, callback, loggerFunc, validations...)
	}
}

// TODO need to extend msg ack depending upon response from callback like error scenario
func (impl PubSubClientServiceImpl) processMsg(msg *nats.Msg, callback func(msg *model.PubSubMsg), loggerFunc LoggerFunc, validations ...ValidateMsg) {
	t1 := time.Now()
	metrics.IncConsumingCount(msg.Subject)
	defer metrics.IncConsumptionCount(msg.Subject)
	defer func() {
		// wrapping this function in defer as directly calling Observe() will run immediately
		metrics.NatsEventConsumptionTime.WithLabelValues(msg.Subject).Observe(float64(time.Since(t1).Milliseconds()))
	}()
	impl.TryCatchCallBack(msg, callback, loggerFunc, validations...)
}

func (impl PubSubClientServiceImpl) publishPanicError(msg *nats.Msg, panicErr error) (err error) {
	publishPanicEvent := model.PublishPanicEvent{
		Topic: PANIC_ON_PROCESSING_TOPIC,
		Payload: model.PanicEventIdentifier{
			Topic:     msg.Subject,
			Data:      string(msg.Data),
			PanicInfo: panicErr.Error(),
		},
	}
	data, err := json.Marshal(publishPanicEvent.Payload)
	if err != nil {
		impl.Logger.Errorw("error in marshalling data! unable to publish panic error", "err", err)
		return err
	}
	err = impl.Publish(publishPanicEvent.Topic, string(data))
	if err != nil {
		impl.Logger.Errorw("error in publishing panic error", "err", err)
		return err
	}
	return nil
}

// TryCatchCallBack is a fail-safe method to use callback function
func (impl PubSubClientServiceImpl) TryCatchCallBack(msg *nats.Msg, callback func(msg *model.PubSubMsg), loggerFunc LoggerFunc, validations ...ValidateMsg) {
	var msgDeliveryCount uint64 = 0
	if metadata, err := msg.Metadata(); err == nil {
		msgDeliveryCount = metadata.NumDelivered
	}
	natsMsgId := msg.Header.Get(model.NatsMsgId)
	subMsg := &model.PubSubMsg{Data: string(msg.Data), MsgDeliverCount: msgDeliveryCount, MsgId: natsMsgId}

	// call loggersFunc
	impl.Log(loggerFunc, msg.Subject, *subMsg)

	// run validations
	for _, validation := range validations {
		if !validation(*subMsg) {
			impl.Logger.Warnw("nats: message validation failed, not processing the message...", "subject", msg.Subject, "msg", string(msg.Data))
			return
		}
	}
	defer func() {
		// Acknowledge the message delivery
		err := msg.Ack()
		if err != nil {
			impl.Logger.Errorw("nats: unable to acknowledge the message", "subject", msg.Subject, "msg", string(msg.Data))
		}

		// publish metrics for msg delivery count if msgDeliveryCount > 1
		if msgDeliveryCount > 1 {
			metrics.NatsEventDeliveryCount.WithLabelValues(msg.Subject, natsMsgId).Observe(float64(msgDeliveryCount))
		}

		// Panic recovery handling
		if panicInfo := recover(); panicInfo != nil {
			impl.Logger.Warnw("nats: found panic error", "subject", msg.Subject, "payload", string(msg.Data), "logs", string(debug.Stack()))
			err = fmt.Errorf("%v\nPanic Logs:\n%s", panicInfo, string(debug.Stack()))
			// Publish the panic info to PANIC_ON_PROCESSING_TOPIC
			publishErr := impl.publishPanicError(msg, err)
			if publishErr != nil {
				impl.Logger.Errorw("error in publishing Panic Event topic", "err", publishErr)
			}
			return
		}
	}()
	// Process the event message
	callback(subMsg)
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

// Updates NATS Consumer config if new changes detected
// if consumer didn't exist, this will just return
func (impl PubSubClientServiceImpl) updateConsumer(natsClient *NatsClient, streamName string, consumerName string, overrideConfig *NatsConsumerConfig) {

	// Get the current Consumer config from NATS-server
	info, err := natsClient.JetStrCtxt.ConsumerInfo(streamName, consumerName)
	if err != nil {
		impl.Logger.Errorw("unable to retrieve consumer info from NATS-server", "stream", streamName, "consumer", consumerName, "err", err)
		return
	}

	existingConfig := info.Config
	updatesDetected := false

	// Currently only checking for AckWait,MaxAckPending but can be done for other editable properties as well
	if ackWaitOverride := time.Duration(overrideConfig.AckWaitInSecs) * time.Second; ackWaitOverride > 0 && existingConfig.AckWait != ackWaitOverride {
		existingConfig.AckWait = ackWaitOverride
		updatesDetected = true
	}

	if messageBufferSize := overrideConfig.GetNatsMsgBufferSize(); messageBufferSize > 0 && existingConfig.MaxAckPending != messageBufferSize {
		existingConfig.MaxAckPending = messageBufferSize
		updatesDetected = true
	}

	if updatesDetected {
		_, err = natsClient.JetStrCtxt.UpdateConsumer(streamName, &existingConfig)
		if err != nil {
			impl.Logger.Errorw("failed to update Consumer config", "received consumer config", info.Config, "err", err)
		}
	}
	return
}

func (impl PubSubClientServiceImpl) Log(loggerFunc LoggerFunc, topic string, subMsg model.PubSubMsg) {
	logMsg, metaSlice := loggerFunc(subMsg)
	logMsg = fmt.Sprintf("%s:%s", NATS_MSG_LOG_PREFIX, logMsg)
	if len(metaSlice) == 0 {
		metaSlice = []interface{}{"msgId", subMsg.MsgId, "msg", subMsg.Data}
	}
	metaSlice = append(metaSlice, "topic", topic)
	impl.Logger.Infow(logMsg, metaSlice...)
}
