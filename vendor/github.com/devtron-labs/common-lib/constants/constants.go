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
