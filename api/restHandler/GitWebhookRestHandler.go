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
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, HandleGitWebhook", "payload", bean)
	resp, err := impl.gitWebhookService.HandleGitWebhook(bean)
	if err != nil {
		impl.logger.Errorw("service err, HandleGitWebhook", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res := map[string]int{"id": resp}
	writeJsonResp(w, err, res, http.StatusCreated)
}
