/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package sse

import (
	"encoding/json"
	"fmt"
	"github.com/juju/errors"
	"net/http"
)

type Validator func(r *http.Request) (string, error)
type Processor func(r *http.Request, receive <-chan int, send chan<- int)

func SubscribeHandler(br *Broker, validator Validator, processor Processor) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receive := make(chan int)
		send := make(chan int)
		namespace, err := validator(r)
		if err != nil {
			respondWithError(w, 400, errors.Details(err))
			return
		}

		// write headers
		headers := w.Header()
		//headers.Set("Access-Control-Allow-Origin", "*") // TODO: make optional
		headers.Set("Content-Type", "text/event-stream; charset=utf-8")
		headers.Set("Cache-Control", "no-cache")
		headers.Set("Connection", "keep-alive")

		//var namespace strings.Builder
		//namespace.WriteString("/")
		//namespace.WriteString(topic)
		connection := NewConnection(w, r, namespace)
		br.register <- connection
		defer func() {
			br.unregister <- connection
			exit(send)
		}()

		go processor(r, send, receive)
		connection.BroadcastMessage(receive)
	})

}

func exit(status chan int) {
	status <- 1
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		fmt.Println(err)
	}
}
