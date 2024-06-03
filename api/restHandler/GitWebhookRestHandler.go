/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package restHandler

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/pkg/git"
	"go.uber.org/zap"
	"net/http"
)

type GitWebhookRestHandler interface {
	HandleGitWebhook(w http.ResponseWriter, r *http.Request)
}

type GitWebhookRestHandlerImpl struct {
	logger            *zap.SugaredLogger
	gitWebhookService git.GitWebhookService
}

func NewGitWebhookRestHandlerImpl(logger *zap.SugaredLogger, gitWebhookService git.GitWebhookService) *GitWebhookRestHandlerImpl {
	return &GitWebhookRestHandlerImpl{
		gitWebhookService: gitWebhookService,
		logger:            logger,
	}
}

func (impl GitWebhookRestHandlerImpl) HandleGitWebhook(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var bean gitSensor.CiPipelineMaterial
	err := decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, HandleGitWebhook", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, HandleGitWebhook", "payload", bean)
	resp, err := impl.gitWebhookService.HandleGitWebhook(bean)
	if err != nil {
		impl.logger.Errorw("service err, HandleGitWebhook", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res := map[string]int{"id": resp}
	common.WriteJsonResp(w, err, res, http.StatusCreated)
}
