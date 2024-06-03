/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package restHandler

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/pkg/pipeline/types"
	"net/http"
	"strings"

	pubsub "github.com/devtron-labs/common-lib/pubsub-lib"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"go.uber.org/zap"
)

type PubSubClientRestHandler interface {
	PublishEventsToNats(w http.ResponseWriter, r *http.Request)
}

type PubSubClientRestHandlerImpl struct {
	pubsubClient *pubsub.PubSubClientServiceImpl
	logger       *zap.SugaredLogger
	cdConfig     *types.CiCdConfig
}

type PublishRequest struct {
	Topic   string          `json:"topic"`
	Payload json.RawMessage `json:"payload"`
}

func NewPubSubClientRestHandlerImpl(pubsubClient *pubsub.PubSubClientServiceImpl, logger *zap.SugaredLogger, cdConfig *types.CiCdConfig) *PubSubClientRestHandlerImpl {
	return &PubSubClientRestHandlerImpl{
		pubsubClient: pubsubClient,
		logger:       logger,
		cdConfig:     cdConfig,
	}
}

func (impl *PubSubClientRestHandlerImpl) PublishEventsToNats(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var publishRequest PublishRequest
	err := decoder.Decode(&publishRequest)
	if err != nil {
		impl.logger.Errorw("request err, HandleExternalCiWebhook", "err", err, "payload", publishRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		impl.logger.Debugw("request err, HandleExternalCiWebhook", "payload", publishRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	reqToken = strings.TrimSpace(splitToken[1])
	if impl.cdConfig.OrchestratorToken != reqToken {
		common.WriteJsonResp(w, err, "Unauthorized req", http.StatusUnauthorized)
		return
	}
	data, err := json.Marshal(publishRequest.Payload)
	if err != nil {
		impl.logger.Errorw("error occurred in un-marshaling publishResquest in publishEvents to Nats", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = impl.pubsubClient.Publish(publishRequest.Topic, string(data))
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", publishRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// result := make(map[string]string)
	// result["id"] = id
	common.WriteJsonResp(w, err, nil, http.StatusAccepted)
}
