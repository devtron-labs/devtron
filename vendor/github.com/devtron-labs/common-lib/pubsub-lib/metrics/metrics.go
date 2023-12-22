package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var NatsPublishingCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "nats_publish_count",
	Help: "count of published events on nats",
}, []string{"topic"})

var NatsPublishingErrorCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "nats_publish_error_count",
	Help: "count of errored published events on nats",
}, []string{"topic"})

var NatsConsumptionCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "nats_consumption_count",
	Help: "count of consumed events on nats ",
}, []string{"topic"})

var NatsConsumingCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "nats_consuming_count",
	Help: "count of nats events whose consumption is in progress",
}, []string{"topic"})

var NatsEventConsumptionTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "nats_event_consumption_time",
}, []string{"topic"})

var NatsEventPublishTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "nats_event_publish_time",
}, []string{"topic"})

func IncPublishCount(topic string) {
	NatsPublishingCount.WithLabelValues(topic).Inc()
}

func IncConsumptionCount(topic string) {
	NatsConsumptionCount.WithLabelValues(topic).Inc()
}

func IncConsumingCount(topic string) {
	NatsConsumingCount.WithLabelValues(topic).Inc()
}

func IncPublishErrorCount(topic string) {
	NatsPublishingErrorCount.WithLabelValues(topic).Inc()
}
