/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package restHandler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/telemetry"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"go.uber.org/zap"
)

type TelemetryRestHandler interface {
	GetTelemetryMetaInfo(w http.ResponseWriter, r *http.Request)
	SendTelemetryData(w http.ResponseWriter, r *http.Request)
	SendSummaryEvent(w http.ResponseWriter, r *http.Request)
}

type TelemetryRestHandlerImpl struct {
	logger               *zap.SugaredLogger
	telemetryEventClient telemetry.TelemetryEventClient
	enforcer             casbin.Enforcer
	userService          user.UserService
}

type TelemetryGenericEvent struct {
	eventType    string
	eventPayload map[string]interface{}
}

func NewTelemetryRestHandlerImpl(logger *zap.SugaredLogger,
	telemetryEventClient telemetry.TelemetryEventClient, enforcer casbin.Enforcer, userService user.UserService) *TelemetryRestHandlerImpl {
	handler := &TelemetryRestHandlerImpl{logger: logger, telemetryEventClient: telemetryEventClient, enforcer: enforcer, userService: userService}
	return handler
}

func (handler TelemetryRestHandlerImpl) GetTelemetryMetaInfo(w http.ResponseWriter, r *http.Request) {
	res, err := handler.telemetryEventClient.GetTelemetryMetaInfo()
	if err != nil {
		handler.logger.Errorw("service err, GetTelemetryMetaInfo", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler TelemetryRestHandlerImpl) SendTelemetryData(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var payload map[string]interface{}
	err = decoder.Decode(&payload)
	if err != nil {
		handler.logger.Errorw("request err, SendTelemetryData", "err", err, "payload", payload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//token := r.Header.Get("token")
	//if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
	//	common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
	//	return
	//}
	var eventTypeString string
	eventType := payload["eventType"]
	if eventType != nil {
		eventTypeStr, ok := eventType.(string)
		if ok {
			eventTypeString = eventTypeStr
		}
	}
	err = handler.telemetryEventClient.SendGenericTelemetryEvent(eventTypeString, payload)

	if err != nil {
		handler.logger.Errorw("service err, SendTelemetryData", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "success", http.StatusOK)

}

func (handler TelemetryRestHandlerImpl) SendSummaryEvent(w http.ResponseWriter, r *http.Request) {
	handler.logger.Info("Handling SendSummaryEvent request")
	decoder := json.NewDecoder(r.Body)
	var payload map[string]interface{}
	err := decoder.Decode(&payload)
	if err != nil && err != io.EOF {
		handler.logger.Errorw("request err, SendSummaryEvent", "err", err, "payload", payload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	var eventType telemetry.TelemetryEventType
	if err == io.EOF {
		eventType = telemetry.Summary
	} else {
		event := payload["eventType"].(string)
		eventType = telemetry.TelemetryEventType(event)
		if eventType == "" {
			eventType = telemetry.Summary
		}
	}

	err = handler.telemetryEventClient.SendSummaryEvent(string(eventType))
	if err != nil {
		handler.logger.Errorw("error occurred in SendSummaryEvent handler in TelemetryRestHandler", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "success", http.StatusOK)
}
