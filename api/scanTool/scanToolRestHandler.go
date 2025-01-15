/*
 * Copyright (c) 2024. Devtron Inc.
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

package scanTool

import (
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/scanTool"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type ScanToolRestHandler interface {
	MartToolActiveOrInActive(w http.ResponseWriter, r *http.Request)
}

type ScanToolRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	userService     user.UserService
	enforcer        casbin.Enforcer
	enforcerUtil    rbac.EnforcerUtil
	validator       *validator.Validate
	scanToolService scanTool.ScanToolMetadataService
}

func NewScanToolRestHandlerImpl(
	logger *zap.SugaredLogger,
	userService user.UserService,
	scanToolService scanTool.ScanToolMetadataService,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil,
	validator *validator.Validate,
) *ScanToolRestHandlerImpl {
	return &ScanToolRestHandlerImpl{
		logger:          logger,
		userService:     userService,
		scanToolService: scanToolService,
		enforcer:        enforcer,
		enforcerUtil:    enforcerUtil,
		validator:       validator,
	}
}

func (impl *ScanToolRestHandlerImpl) MartToolActiveOrInActive(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// since adding/registering a scan tool operates at global level hence super admin check
	// RBAC
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized User"), nil, http.StatusForbidden)
		return
	}
	// RBAC
	queryParams := r.URL.Query()
	toolName := queryParams.Get("toolName")
	toolVersion := queryParams.Get("toolVersion")
	activeStr := queryParams.Get("active")
	if len(toolVersion) == 0 || len(toolName) == 0 || len(activeStr) == 0 {
		common.WriteJsonResp(w, errors.New("please provide toolName, toolVersion and active query params to update"), nil, http.StatusBadRequest)
		return
	}
	active, err := strconv.ParseBool(activeStr)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = impl.scanToolService.MartToolActiveOrInActiveByNameAndVersion(toolName, toolVersion, active)
	if err != nil {
		impl.logger.Errorw("service err, MartToolActiveOrInActiveByNameAndVersion", "toolName", toolName, "toolVersion", toolVersion, "active", active, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)

}
