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

package resourceScan

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning"
	"github.com/devtron-labs/devtron/pkg/policyGovernance/security/imageScanning/bean"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type ScanningResultRestHandler interface {
	ScanResults(w http.ResponseWriter, r *http.Request)
}

type ScanningResultRestHandlerImpl struct {
	logger       *zap.SugaredLogger
	userService  user.UserService
	scanService  imageScanning.ImageScanService
	enforcer     casbin.Enforcer
	enforcerUtil rbac.EnforcerUtil
	validator    *validator.Validate
}

func NewScanningResultRestHandlerImpl(
	logger *zap.SugaredLogger,
	userService user.UserService,
	scanService imageScanning.ImageScanService,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil,
	validator *validator.Validate,
) *ScanningResultRestHandlerImpl {
	return &ScanningResultRestHandlerImpl{
		logger:       logger,
		userService:  userService,
		scanService:  scanService,
		enforcer:     enforcer,
		enforcerUtil: enforcerUtil,
		validator:    validator,
	}
}

func getResourceScanQueryParams(w http.ResponseWriter, r *http.Request) (*bean.ResourceScanQueryParams, error) {
	queryParams := &bean.ResourceScanQueryParams{}
	var appId, envId, installedAppId, artifactId, installedAppVersionHistoryId int
	var err error
	appId, err = common.ExtractIntQueryParam(w, r, "appId", 0)
	if err != nil {
		return queryParams, err
	}
	queryParams.AppId = appId

	installedAppId, err = common.ExtractIntQueryParam(w, r, "installedAppId", 0)
	if err != nil {
		return queryParams, err
	}
	queryParams.InstalledAppId = installedAppId

	installedAppVersionHistoryId, err = common.ExtractIntQueryParam(w, r, "installedAppVersionHistoryId", 0)
	if err != nil {
		return queryParams, err
	}
	queryParams.InstalledAppVersionHistoryId = installedAppVersionHistoryId

	envId, err = common.ExtractIntQueryParam(w, r, "envId", 0)
	if err != nil {
		return queryParams, err
	}
	queryParams.EnvId = envId

	artifactId, err = common.ExtractIntQueryParam(w, r, "artifactId", 0)
	if err != nil {
		return queryParams, err
	}
	queryParams.ArtifactId = artifactId
	return queryParams, nil
}

func (impl ScanningResultRestHandlerImpl) ScanResults(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	resourceScanQueryParams, err := getResourceScanQueryParams(w, r)
	if err != nil {
		return
	}
	// RBAC
	token := r.Header.Get("token")
	object := impl.enforcerUtil.GetAppRBACNameByAppId(resourceScanQueryParams.AppId)
	if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if resourceScanQueryParams.EnvId > 0 {
		object = impl.enforcerUtil.GetEnvRBACNameByAppId(resourceScanQueryParams.AppId, resourceScanQueryParams.EnvId)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, object); !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	// RBAC
	resp, err := impl.scanService.GetScanResults(resourceScanQueryParams)
	if err != nil {
		impl.logger.Errorw("service err, GetScanResults", "resourceScanQueryParams", resourceScanQueryParams, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, resp, http.StatusOK)

}
