/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
