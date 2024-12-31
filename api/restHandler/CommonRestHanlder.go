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
	"net/http"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/commonService"
	"go.uber.org/zap"
)

type CommonRestHanlder interface {
	GlobalChecklist(w http.ResponseWriter, r *http.Request)
}

type CommonRestHanlderImpl struct {
	logger          *zap.SugaredLogger
	userAuthService user.UserService
	commonService   commonService.CommonService
}

func NewCommonRestHanlderImpl(
	logger *zap.SugaredLogger,
	userAuthService user.UserService,
	commonService commonService.CommonService) *CommonRestHanlderImpl {
	return &CommonRestHanlderImpl{
		logger:          logger,
		userAuthService: userAuthService,
		commonService:   commonService,
	}
}

func (impl CommonRestHanlderImpl) GlobalChecklist(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	res, err := impl.commonService.GlobalChecklist()
	if err != nil {
		impl.logger.Errorw("service err, GlobalChecklist", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}
