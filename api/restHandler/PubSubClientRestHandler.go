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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"go.uber.org/zap"
)

type PubSubClientRestHandler interface {
	PublishEventsToNats(w http.ResponseWriter, r *http.Request)
}

type PubSubClientRestHandlerImpl struct {
	natsPublishClient pubsub.NatsPublishClient
	logger            *zap.SugaredLogger
	cdConfig          *pipeline.CdConfig
}

func NewPubSubClientRestHandlerImpl(natsPublishClient pubsub.NatsPublishClient, logger *zap.SugaredLogger, cdConfig *pipeline.CdConfig) *PubSubClientRestHandlerImpl {
	return &PubSubClientRestHandlerImpl{
		natsPublishClient: natsPublishClient,
		logger:            logger,
		cdConfig:          cdConfig,
	}
}

func (impl *PubSubClientRestHandlerImpl) PublishEventsToNats(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var publishRequest pubsub.PublishRequest
	err := decoder.Decode(&publishRequest)
	if err != nil {
		impl.logger.Errorw("request err, HandleExternalCiWebhook", "err", err, "payload", publishRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		impl.logger.Debugw("request err, HandleExternalCiWebhook", "payload", publishRequest, "token", reqToken)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	reqToken = strings.TrimSpace(splitToken[1])
	if impl.cdConfig.OrchestratorToken != reqToken {
		common.WriteJsonResp(w, err, "Unauthorized req", http.StatusUnauthorized)
		return
	}

	err = impl.natsPublishClient.Publish(&publishRequest)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", publishRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// result := make(map[string]string)
	// result["id"] = id
	common.WriteJsonResp(w, err, nil, http.StatusAccepted)
}
