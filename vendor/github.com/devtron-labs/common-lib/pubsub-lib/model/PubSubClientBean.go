package model

const PUBLISH_SUCCESS = "SUCCESS"
const PUBLISH_FAILURE = "FAILURE"
const NatsMsgId = "Nats-Msg-Id"

type PubSubMsg struct {
	Data            string
	MsgDeliverCount uint64
	MsgId           *string
}

type LogsConfig struct {
	DefaultLogTimeLimit int64 `env:"DEFAULT_LOG_TIME_LIMIT" envDefault:"1"`
}

// PublishPanicEvent is used for PANIC_ON_PROCESSING_TOPIC payload
type PublishPanicEvent struct {
	Topic   string               `json:"topic"`   // PANIC_ON_PROCESSING_TOPIC
	Payload PanicEventIdentifier `json:"payload"` // Panic Info structure
}

// PanicEventIdentifier is used to describe panic info
type PanicEventIdentifier struct {
	Topic     string `json:"topic"`
	Data      string `json:"data"`
	PanicInfo string `json:"panicInfo"`
}
