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

package constants

const (
	PanicLogIdentifier         = "DEVTRON_PANIC_RECOVER"
	GoRoutinePanicMsgLogPrefix = "GO_ROUTINE_PANIC_LOG:"
)

// service names constant

type ServiceName string

func (m ServiceName) ToString() string {
	return string(m)
}

const (
	Orchestrator ServiceName = "ORCHESTRATOR"
	Kubelink     ServiceName = "KUBELINK"
	GitSensor    ServiceName = "GITSENSOR"
	Kubewatch    ServiceName = "KUBEWATCH"
)

// metrics name constants
const (
	NATS_PUBLISH_COUNT                 = "nats_publish_count"
	NATS_CONSUMPTION_COUNT             = "nats_consumption_count"
	NATS_CONSUMING_COUNT               = "nats_consuming_count"
	NATS_EVENT_CONSUMPTION_TIME        = "nats_event_consumption_time"
	NATS_EVENT_PUBLISH_TIME            = "nats_event_publish_time"
	NATS_EVENT_DELIVERY_COUNT          = "nats_event_delivery_count"
	PANIC_RECOVERY_COUNT               = "panic_recovery_count"
	REVERSE_PROXY_PANIC_RECOVERY_COUNT = "reverse_proxy_panic_recovery_count"
)

// metrics labels constant
const (
	PANIC_TYPE = "panic_type"
	HOST       = "host"
	METHOD     = "method"
	PATH       = "path"
	TOPIC      = "topic"
	STATUS     = "status"
)
