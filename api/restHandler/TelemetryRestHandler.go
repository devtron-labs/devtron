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

package restHandler

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/telemetry"
	"go.uber.org/zap"
	"net/http"
)

type TelemetryRestHandler interface {
	GetTelemetryMetaInfo(w http.ResponseWriter, r *http.Request)
}

type TelemetryRestHandlerImpl struct {
	logger               *zap.SugaredLogger
	telemetryEventClient telemetry.TelemetryEventClient
}

func NewTelemetryRestHandlerImpl(logger *zap.SugaredLogger,
	telemetryEventClient telemetry.TelemetryEventClient) *TelemetryRestHandlerImpl {
	handler := &TelemetryRestHandlerImpl{logger: logger, telemetryEventClient: telemetryEventClient}
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
