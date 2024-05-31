/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
