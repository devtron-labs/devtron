package model

type PubSubMsg struct {
	Data string
}

type LogsConfig struct {
	DefaultLogTimeLimit int64 `env:"DEFAULT_LOG_TIME_LIMIT" envDefault:"1"`
}

type PublishPanicEvent struct {
	Topic   string               `json:"topic"`
	Payload PanicEventIdentifier `json:"payload"`
}

type PanicEventIdentifier struct {
	Topic     string `json:"topic"`
	Data      string `json:"data"`
	PanicInfo string `json:"panicInfo"`
}
