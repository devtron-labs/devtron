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
	util3 "github.com/devtron-labs/devtron/api/util"
	"github.com/devtron-labs/devtron/pkg/workflow/dag"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type ExternalCiRestHandler interface {
	HandleExternalCiWebhook(w http.ResponseWriter, r *http.Request)
}

type ExternalCiRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	validator           *validator.Validate
	userService         user.UserService
	enforcer            casbin.Enforcer
	workflowDagExecutor dag.WorkflowDagExecutor
}

func NewExternalCiRestHandlerImpl(logger *zap.SugaredLogger, validator *validator.Validate,
	userService user.UserService, enforcer casbin.Enforcer,
	workflowDagExecutor dag.WorkflowDagExecutor) *ExternalCiRestHandlerImpl {
	return &ExternalCiRestHandlerImpl{
		logger:              logger,
		validator:           validator,
		userService:         userService,
		enforcer:            enforcer,
		workflowDagExecutor: workflowDagExecutor,
	}
}

func (impl ExternalCiRestHandlerImpl) HandleExternalCiWebhook(w http.ResponseWriter, r *http.Request) {
	util3.SetupCorsOriginHeader(&w)
	vars := mux.Vars(r)
	token := r.Header.Get("api-token")
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
	var req pipeline.ExternalCiWebhookDto
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
	ciArtifactReq, err := impl.workflowDagExecutor.BuildCiArtifactRequestForWebhook(req)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	err = impl.validator.Struct(ciArtifactReq)
	if err != nil {
		impl.logger.Errorw("validation err, HandleExternalCiWebhook", "err", err, "payload", ciArtifactReq)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	_, err = impl.workflowDagExecutor.HandleExternalCiWebhook(externalCiId, ciArtifactReq, impl.checkExternalCiDeploymentAuth, token)
	if err != nil {
		impl.logger.Errorw("service err, HandleExternalCiWebhook", "err", err, "payload", req)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (impl ExternalCiRestHandlerImpl) checkExternalCiDeploymentAuth(token string, projectObject string, envObject string) bool {
	if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionTrigger, projectObject); !ok {
		return false
	}
	if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionTrigger, envObject); !ok {
		return false
	}
	return true
}
