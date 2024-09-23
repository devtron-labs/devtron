/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metrics

import (
	"github.com/devtron-labs/common-lib/constants"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var NatsPublishingCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: constants.NATS_PUBLISH_COUNT,
	Help: "count of successfully published events on nats",
}, []string{constants.TOPIC, constants.STATUS})

var NatsConsumptionCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: constants.NATS_CONSUMPTION_COUNT,
	Help: "count of consumed events on nats ",
}, []string{constants.TOPIC})

var NatsConsumingCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: constants.NATS_CONSUMING_COUNT,
	Help: "count of nats events whose consumption is in progress",
}, []string{constants.TOPIC})

var NatsEventConsumptionTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: constants.NATS_EVENT_CONSUMPTION_TIME,
}, []string{constants.TOPIC})

var NatsEventPublishTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: constants.NATS_EVENT_PUBLISH_TIME,
}, []string{constants.TOPIC})

var NatsEventDeliveryCount = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: constants.NATS_EVENT_DELIVERY_COUNT,
}, []string{constants.TOPIC})

var PanicRecoveryCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: constants.PANIC_RECOVERY_COUNT,
}, []string{constants.PANIC_TYPE, constants.HOST, constants.METHOD, constants.PATH})

var ReverseProxyPanicRecoveryCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: constants.REVERSE_PROXY_PANIC_RECOVERY_COUNT,
}, []string{constants.PANIC_TYPE, constants.HOST, constants.METHOD, constants.PATH})

func IncPublishCount(topic, status string) {
	NatsPublishingCount.WithLabelValues(topic, status).Inc()
}

func IncConsumptionCount(topic string) {
	NatsConsumptionCount.WithLabelValues(topic).Inc()
}

func IncConsumingCount(topic string) {
	NatsConsumingCount.WithLabelValues(topic).Inc()
}

func IncPanicRecoveryCount(panicType, host, method, path string) {
	PanicRecoveryCount.WithLabelValues(panicType, host, method, path).Inc()
}

func IncReverseProxyPanicRecoveryCount(panicType, host, method, path string) {
	ReverseProxyPanicRecoveryCount.WithLabelValues(panicType, host, method, path).Inc()
}
