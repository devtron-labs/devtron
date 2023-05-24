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
	"strings"
	"time"
)

type SSE struct {
	Broker          *Broker
	OutboundChannel chan<- SSEMessage
}

type Broker struct {
	notifier    chan SSEMessage
	connections map[*Connection]bool
	register    chan *Connection
	unregister  chan *Connection
	shutdown    chan bool
	createdTime time.Time
}

func NewSSE() *SSE {
	sse := SSE{
		Broker: NewBroker(),
	}
	sse.Broker.Start()
	sse.OutboundChannel = sse.Broker.notifier
	return &sse
}

func NewBroker() *Broker {
	broker := Broker{
		notifier:    make(chan SSEMessage),
		connections: make(map[*Connection]bool),
		register:    make(chan *Connection),
		unregister:  make(chan *Connection),
		shutdown:    make(chan bool),
		createdTime: time.Now(),
	}
	return &broker
}

func (br *Broker) shutDown() {
	br.shutdown <- true
}

func (br *Broker) Start() {
	go br.run()
}

func (br *Broker) run() {
	for {
		select {
		case <-br.shutdown:
			for conn := range br.connections {
				br.shutdownConnection(conn)
			}
			return
		case conn := <-br.register:
			br.connections[conn] = true
		case conn := <-br.unregister:
			br.unregisterConnection(conn)
		case msg := <-br.notifier:
			br.broadcastMessage(msg)
		}
	}
}

func (br *Broker) shutdownConnection(conn *Connection) {
	br.unregisterConnection(conn)
	close(conn.outboundMessage)
}

func (br *Broker) unregisterConnection(conn *Connection) {
	delete(br.connections, conn)
}

func (br *Broker) broadcastMessage(message SSEMessage) {
	fmtMsg := message.format()
	for conn := range br.connections {
		if strings.HasPrefix(message.Namespace, conn.namespace) {
			select {
			case conn.outboundMessage <- fmtMsg:
			default:
				br.shutdownConnection(conn)
			}
		}

	}
}
