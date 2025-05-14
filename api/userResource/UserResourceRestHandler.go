/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

package userResource

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/api/userResource/adapter"
	apiBean "github.com/devtron-labs/devtron/api/userResource/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/userResource"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type RestHandler interface {
	GetResourceOptions(w http.ResponseWriter, r *http.Request)
}
type RestHandlerImpl struct {
	logger              *zap.SugaredLogger
	userService         user.UserService
	userResourceService userResource.UserResourceService
}

func NewUserResourceRestHandler(logger *zap.SugaredLogger,
	userService user.UserService,
	userResourceService userResource.UserResourceService) *RestHandlerImpl {
	return &RestHandlerImpl{
		logger:              logger,
		userService:         userService,
		userResourceService: userResourceService,
	}
}

func (handler *RestHandlerImpl) GetResourceOptions(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	pathParams, caughtError := decodePathParams(w, r)
	if caughtError {
		return
	}
	decoder := json.NewDecoder(r.Body)
	var reqBean apiBean.ResourceOptionsReqDto
	err = decoder.Decode(&reqBean)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	// rbac enforcement is managed at service level based on entity and kind
	data, err := handler.userResourceService.GetResourceOptions(r.Context(), token, &reqBean, pathParams)
	if err != nil {
		handler.logger.Errorw("service error, GetResourceOptions", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, data, http.StatusOK)
	return

}

func decodePathParams(w http.ResponseWriter, r *http.Request) (pathParams *apiBean.PathParams, caughtError bool) {
	vars := mux.Vars(r)
	kindVar := vars[apiBean.PathParamKind]
	versionVar := vars[apiBean.PathParamVersion]
	pathParams = adapter.BuildPathParams(kindVar, versionVar)
	return pathParams, caughtError
}
