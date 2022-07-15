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

package webhookHelm

import (
	"context"
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	webhookHelm "github.com/devtron-labs/devtron/pkg/webhook/helm"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type WebhookHelmRestHandler interface {
	InstallOrUpdateApplication(w http.ResponseWriter, r *http.Request)
}

type WebhookHelmRestHandlerImpl struct {
	logger             *zap.SugaredLogger
	webhookHelmService webhookHelm.WebhookHelmService
	userService        user.UserService
	enforcer           casbin.Enforcer
	validator          *validator.Validate
}

func NewWebhookHelmRestHandlerImpl(logger *zap.SugaredLogger, webhookHelmService webhookHelm.WebhookHelmService, userService user.UserService,
	enforcer casbin.Enforcer, validator *validator.Validate) *WebhookHelmRestHandlerImpl {
	return &WebhookHelmRestHandlerImpl{
		logger:             logger,
		webhookHelmService: webhookHelmService,
		userService:        userService,
		enforcer:           enforcer,
		validator:          validator,
	}
}

func (impl WebhookHelmRestHandlerImpl) InstallOrUpdateApplication(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteApiJsonResponse(w, nil, http.StatusUnauthorized, common.UnAuthenticated, "")
		return
	}

	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteApiJsonResponse(w, nil, http.StatusForbidden, common.UnAuthorized, "")
		return
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	var request *webhookHelm.HelmAppCreateUpdateRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("err in decoding request in InstallOrUpdateHelmApplication", "err", err)
		common.WriteApiJsonResponse(w, nil, http.StatusBadRequest, common.BadRequest, err.Error())
		return
	}

	// validate request
	err = impl.validator.Struct(request)
	if err != nil {
		impl.logger.Errorw("validation err in InstallOrUpdateHelmApplication", "err", err, "request", request)
		common.WriteApiJsonResponse(w, nil, http.StatusBadRequest, common.BadRequest, err.Error())
		return
	}

	// service call
	result, errCode, errMsg, statusCode := impl.webhookHelmService.CreateOrUpdateHelmApplication(context.Background(), request)
	if len(errCode) > 0 {
		impl.logger.Errorw("service err in InstallOrUpdateHelmApplication", "request", request, "errCode", errCode, "errMsg", errMsg)
		common.WriteApiJsonResponse(w, nil, statusCode, errCode, errMsg)
		return
	}

	common.WriteApiJsonResponse(w, result, http.StatusOK, "", "")
}
