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
	"github.com/devtron-labs/devtron/client/pubsub"
	"github.com/devtron-labs/devtron/pkg/sso"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type TelemetryRestHandler interface {
	GetUCID(w http.ResponseWriter, r *http.Request)
}

type TelemetryRestHandlerImpl struct {
	userAuthService user.UserAuthService
	validator       *validator.Validate
	logger          *zap.SugaredLogger
	enforcer        rbac.Enforcer
	natsClient      *pubsub.PubSubClient
	userService     user.UserService
	ssoLoginService sso.SSOLoginService
}

func NewTelemetryRestHandlerImpl(userAuthService user.UserAuthService, validator *validator.Validate,
	logger *zap.SugaredLogger, enforcer rbac.Enforcer, natsClient *pubsub.PubSubClient, userService user.UserService,
	ssoLoginService sso.SSOLoginService) *TelemetryRestHandlerImpl {
	handler := &TelemetryRestHandlerImpl{userAuthService: userAuthService, validator: validator, logger: logger,
		enforcer: enforcer, natsClient: natsClient, userService: userService, ssoLoginService: ssoLoginService}
	return handler
}

func (handler TelemetryRestHandlerImpl) GetUCID(w http.ResponseWriter, r *http.Request) {
	res, err := handler.ssoLoginService.GetUCID()
	if err != nil {
		handler.logger.Errorw("service err, GetUCID", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, nil, res, http.StatusOK)
}
