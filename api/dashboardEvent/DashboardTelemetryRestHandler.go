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

package dashboardEvent

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/telemetry"
	"go.uber.org/zap"
	"net/http"
)

type DashboardTelemetryRestHandler interface {
	SendDashboardAccessedEvent(w http.ResponseWriter, r *http.Request)
	SendDashboardLoggedInEvent(w http.ResponseWriter, r *http.Request)
}

type DashboardTelemetryRestHandlerImpl struct {
	logger    *zap.SugaredLogger
	telemetry telemetry.TelemetryEventClient
}

func NewDashboardTelemetryRestHandlerImpl(logger *zap.SugaredLogger,
	telemetry telemetry.TelemetryEventClient) *DashboardTelemetryRestHandlerImpl {
	return &DashboardTelemetryRestHandlerImpl{
		logger:    logger,
		telemetry: telemetry,
	}
}

func (handler *DashboardTelemetryRestHandlerImpl) SendDashboardAccessedEvent(w http.ResponseWriter, r *http.Request) {
	err := handler.telemetry.SendTelemetryDashboardAccessEvent()
	if err != nil {
		handler.logger.Warnw("Sending Telemetry Event failed", "error", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("Event Sent Successfully")
	common.WriteJsonResp(w, err, "Event Sent Successfully", http.StatusOK)
	return
}

func (handler *DashboardTelemetryRestHandlerImpl) SendDashboardLoggedInEvent(w http.ResponseWriter, r *http.Request) {
	err := handler.telemetry.SendTelemetryDashboardLoggedInEvent()
	if err != nil {
		handler.logger.Warnw("Sending Telemetry Event failed", "error", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("Event Sent Successfully")
	common.WriteJsonResp(w, err, "Event Sent Successfully", http.StatusOK)
	return
}
