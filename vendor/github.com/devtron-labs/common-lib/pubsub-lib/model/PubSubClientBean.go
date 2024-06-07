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

package model

const PUBLISH_SUCCESS = "SUCCESS"
const PUBLISH_FAILURE = "FAILURE"
const NatsMsgId = "Nats-Msg-Id"

type PubSubMsg struct {
	Data            string
	MsgDeliverCount uint64
	MsgId           string
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
