/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package sse

import (
	"net/http"
	"time"
)

type Connection struct {
	request         *http.Request
	response        http.ResponseWriter
	createdTime     time.Time
	outboundMessage chan []byte
	namespace       string
}

func NewConnection(response http.ResponseWriter, request *http.Request, namespace string) *Connection {
	conn := Connection{
		request:         request,
		response:        response,
		createdTime:     time.Now(),
		outboundMessage: make(chan []byte, 256),
		namespace:       namespace,
	}
	return &conn
}

func (conn *Connection) BroadcastMessage(receive <-chan int) {
	keepAliveTime := time.NewTicker(5 * time.Second)
	keepAliveMsg := []byte(":keepalive\n")

	defer keepAliveTime.Stop()

	for {
		select {
		case msg, ok := <-conn.outboundMessage:
			if !ok {
				return
			}
			_, err := conn.response.Write(msg)
			if err != nil {
				return
			}
			if flusher, ok := conn.response.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-keepAliveTime.C:
			_, err := conn.response.Write(keepAliveMsg)
			if err != nil {
				return
			}
			if flusher, ok := conn.response.(http.Flusher); ok {
				flusher.Flush()
			}
		case <-conn.request.Context().Done():
			return
		case <-receive:
			return
		}
	}
}
