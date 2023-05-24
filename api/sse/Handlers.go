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
