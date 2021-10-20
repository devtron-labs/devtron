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

package connector

import (
	"encoding/json"
	"fmt"
	"github.com/argoproj/argo-cd/pkg/apiclient/application"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/gogo/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var delimiter = []byte("\n\n")

type Pump interface {
	StartStream(w http.ResponseWriter, recv func() (proto.Message, error), err error)
	StartStreamWithHeartBeat(w http.ResponseWriter, isReconnect bool, recv func() (*application.LogEntry, error), err error)
	StartMessage(w http.ResponseWriter, resp proto.Message, perr error)
}

type PumpImpl struct {
	logger *zap.SugaredLogger
}

func NewPumpImpl(logger *zap.SugaredLogger) *PumpImpl {
	return &PumpImpl{
		logger: logger,
	}
}

func (impl PumpImpl) StartStreamWithHeartBeat(w http.ResponseWriter, isReconnect bool, recv func() (*application.LogEntry, error), err error) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "unexpected server doesnt support streaming", http.StatusInternalServerError)
	}
	if err != nil {
		http.Error(w, errors.Details(err), http.StatusInternalServerError)
	}
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	var wroteHeader bool
	if isReconnect {
		err := impl.sendEvent(nil, []byte("RECONNECT_STREAM"), []byte("RECONNECT_STREAM"), w)
		if err != nil {
			impl.logger.Errorw("error in writing data over sse", "err", err)
			return
		}
	}
	// heartbeat start
	ticker := time.NewTicker(30 * time.Second)
	done := make(chan bool)
	var mux sync.Mutex
	go func() error {
		for {
			select {
			case <-done:
				return nil
			case t := <-ticker.C:
				mux.Lock()
				err := impl.sendEvent(nil, []byte("PING"), []byte(t.String()), w)
				mux.Unlock()
				if err != nil {
					impl.logger.Errorw("error in writing PING over sse", "err", err)
					return err
				}
				f.Flush()
			}
		}
	}()
	defer func() {
		ticker.Stop()
		done <- true
	}()

	// heartbeat end

	for {
		resp, err := recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			impl.logger.Errorf("Error occurred while reading data from argocd %+v\n", err)
			impl.handleForwardResponseStreamError(wroteHeader, w, err)
			return
		}
		response := bean.Response{}
		response.Result = resp
		buf, err := json.Marshal(response)
		if err != nil {
			impl.logger.Errorw("error in marshaling data", "err", err)
			return
		}
		mux.Lock()
		err = impl.sendEvent([]byte(strconv.FormatInt(resp.GetTimeStamp().UnixNano(), 10)), nil, buf, w)
		mux.Unlock()
		if err != nil {
			impl.logger.Errorw("error in writing data over sse", "err", err)
			return
		}
		wroteHeader = true
		f.Flush()
	}
}

func (impl *PumpImpl) sendEvent(eventId []byte, eventName []byte, payload []byte, w http.ResponseWriter) error {
	var res []byte
	if len(eventId) > 0 {
		res = append(res, "id: "...)
		res = append(res, eventId...)
		res = append(res, '\n')
	}
	if len(eventName) > 0 {
		res = append(res, "event:"...)
		res = append(res, eventName...)
		res = append(res, '\n')
	}
	if len(payload) > 0 {
		res = append(res, "data:"...)
		res = append(res, payload...)
	}
	res = append(res, '\n', '\n')
	if i, err := w.Write(res); err != nil {
		impl.logger.Errorf("Failed to send response chunk: %v", err)
		return err
	} else {
		impl.logger.Debugw("msg written", "count", i)
	}

	return nil
}

func (impl PumpImpl) StartStream(w http.ResponseWriter, recv func() (proto.Message, error), err error) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "unexpected server doesnt support streaming", http.StatusInternalServerError)
	}
	if err != nil {
		http.Error(w, errors.Details(err), http.StatusInternalServerError)
	}
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("X-Accel-Buffering", "no")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	var wroteHeader bool
	for {
		resp, err := recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			impl.logger.Errorf("Error occurred while reading data from argocd %+v\n", err)
			impl.handleForwardResponseStreamError(wroteHeader, w, err)
			return
		}
		response := bean.Response{}
		response.Result = resp
		buf, err := json.Marshal(response)
		data := "data: " + string(buf)
		if _, err = w.Write([]byte(data)); err != nil {
			impl.logger.Errorf("Failed to send response chunk: %v", err)
			return
		}
		wroteHeader = true
		if _, err = w.Write(delimiter); err != nil {
			impl.logger.Errorf("Failed to send delimiter chunk: %v", err)
			return
		}
		f.Flush()
	}
}

func (impl PumpImpl) handleForwardResponseStreamError(wroteHeader bool, w http.ResponseWriter, err error) {
	code := "000"
	if !wroteHeader {
		s, ok := status.FromError(err)
		if !ok {
			s = status.New(codes.Unknown, err.Error())
		}
		w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
		code = fmt.Sprint(s.Code())
	}
	response := bean.Response{}
	apiErr := bean.ApiError{}
	apiErr.Code = code // 000=unknown
	apiErr.InternalMessage = errors.Details(err)
	response.Errors = []bean.ApiError{apiErr}
	buf, merr := json.Marshal(response)
	if merr != nil {
		impl.logger.Errorf("Failed to marshal response %+v\n", merr)
	}
	if _, werr := w.Write(buf); werr != nil {
		impl.logger.Errorf("Failed to notify error to client: %v", werr)
		return
	}
}

func (impl PumpImpl) StartMessage(w http.ResponseWriter, resp proto.Message, perr error) {
	//w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("Content-Type", "application/json")

	response := bean.Response{}
	if perr != nil {
		impl.handleForwardResponseMessageError(w, perr)
		return
	}
	var buf []byte
	var err error
	if rb, ok := resp.(responseBody); ok {
		response.Result = rb.XXX_ResponseBody()
		buf, err = json.Marshal(response)
	} else {
		response.Result = resp
		buf, err = json.Marshal(response)
	}
	if err != nil {
		impl.logger.Errorf("Marshal error: %v", err)
		return
	}

	if _, err = w.Write(buf); err != nil {
		impl.logger.Errorf("Failed to write response: %v", err)
	}
}

func (impl PumpImpl) handleForwardResponseMessageError(w http.ResponseWriter, err error) {
	code := "000"
	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}
	w.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
	code = fmt.Sprint(s.Code())
	response := bean.Response{}
	apiErr := bean.ApiError{}
	apiErr.Code = code // 000=unknown
	apiErr.InternalMessage = errors.Details(err)
	response.Errors = []bean.ApiError{apiErr}
	buf, merr := json.Marshal(response)
	if merr != nil {
		impl.logger.Errorf("Failed to marshal response %+v\n", merr)
	}
	if _, werr := w.Write(buf); werr != nil {
		impl.logger.Errorf("Failed to notify error to client: %v", werr)
		return
	}
}

type responseBody interface {
	XXX_ResponseBody() interface{}
}
