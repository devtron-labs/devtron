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
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/api/router/pubsub"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type ExternalCiRestHandler interface {
	HandleExternalCiWebhook(w http.ResponseWriter, r *http.Request)
}

type ExternalCiRestHandlerImpl struct {
	logger         *zap.SugaredLogger
	webhookService pipeline.WebhookService
	ciEventHandler pubsub.CiEventHandler
	validator   *validator.Validate
	userService user.UserService
	enforcer    casbin.Enforcer
	enforcerUtil   rbac.EnforcerUtil
}

func NewExternalCiRestHandlerImpl(logger *zap.SugaredLogger, webhookService pipeline.WebhookService,
	ciEventHandler pubsub.CiEventHandler, validator *validator.Validate, userService user.UserService,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil) *ExternalCiRestHandlerImpl {
	return &ExternalCiRestHandlerImpl{
		webhookService: webhookService,
		logger:         logger,
		ciEventHandler: ciEventHandler,
		validator:      validator,
		userService:    userService,
		enforcer:       enforcer,
		enforcerUtil:   enforcerUtil,
	}
}

func (impl ExternalCiRestHandlerImpl) HandleExternalCiWebhook(w http.ResponseWriter, r *http.Request) {
	setupResponse(&w, r)
	vars := mux.Vars(r)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized", http.StatusUnauthorized)
		return
	}
	externalCiId, err := strconv.Atoi(vars["externalCiId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var req pubsub.CiCompleteEvent
	err = decoder.Decode(&req)
	if err != nil {
		impl.logger.Errorw("request err, HandleExternalCiWebhook", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	req.TriggeredBy = userId
	impl.logger.Infow("request payload, HandleExternalCiWebhook", "payload", req)

	err = impl.validator.Struct(req)
	if err != nil {
		impl.logger.Errorw("validation err, HandleExternalCiWebhook", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//fetching request
	ciArtifactReq, err := impl.ciEventHandler.BuildCiArtifactRequestForWebhook(req)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	_, err = impl.webhookService.HandleExternalCiWebhook(externalCiId, ciArtifactReq, impl.checkExternalCiDeploymentAuth)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (impl ExternalCiRestHandlerImpl) checkExternalCiDeploymentAuth(email string, projectObject string, envObject string) bool {
	if ok := impl.enforcer.EnforceByEmail(email, casbin.ResourceApplications, casbin.ActionTrigger, projectObject); !ok {
		return false
	}
	if ok := impl.enforcer.EnforceByEmail(email, casbin.ResourceEnvironment, casbin.ActionTrigger, envObject); !ok {
		return false
	}
	return true
}
