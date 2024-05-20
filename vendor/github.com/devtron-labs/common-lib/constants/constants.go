package constants

const PanicLogIdentifier = "DEVTRON_PANIC_RECOVER"

// metrics name constants

const (
	NATS_PUBLISH_COUNT          = "Nats_Publish_Count"
	NATS_CONSUMPTION_COUNT      = "Nats_Consumption_Count"
	NATS_CONSUMING_COUNT        = "Nats_Consuming_Count"
	NATS_EVENT_CONSUMPTION_TIME = "Nats_Event_Consumption_Time"
	NATS_EVENT_PUBLISH_TIME     = "Nats_Event_Publish_Time"
	NATS_EVENT_DELIVERY_COUNT   = "Nats_Event_Delivery_Count"
	PANIC_RECOVERY_COUNT        = "Panic_Recovery_Count"
)

// metrcis lables constant
const (
	PANIC_TYPE = "panicType"
	HOST       = "host"
	METHOD     = "method"
	PATH       = "path"
	TOPIC      = "topic"
	STATUS     = "status"
)
