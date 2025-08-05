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

package configure

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/bean"
	"net/http"
	"strconv"
)

// Helper functions for common operations in BuildPipelineRestHandler.go

// getUserIdOrUnauthorized gets the logged-in user ID or returns an unauthorized response
func (handler *PipelineConfigRestHandlerImpl) getUserIdOrUnauthorized(w http.ResponseWriter, r *http.Request) (int32, bool) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.HandleUnauthorized(w, r)
		return 0, false
	}
	return userId, true
}

// getIntPathParam gets an integer path parameter from the request
// DEPRECATED: Use common.ExtractIntPathParamWithContext() for new code
func (handler *PipelineConfigRestHandlerImpl) getIntPathParam(w http.ResponseWriter, vars map[string]string, paramName string) (int, bool) {
	paramValue, err := strconv.Atoi(vars[paramName])
	if err != nil {
		// Use enhanced error handling
		apiErr := util.NewInvalidPathParameterError(paramName, vars[paramName])
		handler.Logger.Errorw("Invalid path parameter", "paramName", paramName, "paramValue", vars[paramName], "err", err)
		common.WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
		return 0, false
	}
	if paramValue <= 0 {
		apiErr := util.NewValidationErrorForField(paramName, "must be a positive integer")
		handler.Logger.Errorw("Invalid path parameter value", "paramName", paramName, "paramValue", paramValue)
		common.WriteJsonResp(w, apiErr, nil, apiErr.HttpStatusCode)
		return 0, false
	}
	return paramValue, true
}

// decodeJsonBody decodes the request body into the provided struct
func (handler *PipelineConfigRestHandlerImpl) decodeJsonBody(w http.ResponseWriter, r *http.Request, obj interface{}, logContext string) bool {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(obj)
	if err != nil {
		handler.Logger.Errorw("request err, decode json body", "err", err, "context", logContext, "payload", obj)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return false
	}
	return true
}

// validateRequestBody validates the request body against struct validation tags
func (handler *PipelineConfigRestHandlerImpl) validateRequestBody(w http.ResponseWriter, obj interface{}, logContext string) bool {
	err := handler.validator.Struct(obj)
	if err != nil {
		handler.Logger.Errorw("validation err", "err", err, "context", logContext, "payload", obj)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return false
	}
	return true
}

// getAppAndCheckAuthForAction gets the app and checks if the user has the required permission
func (handler *PipelineConfigRestHandlerImpl) getAppAndCheckAuthForAction(w http.ResponseWriter, appId int, token string, action string) (app *bean.CreateAppDTO, authorized bool) {
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.Logger.Errorw("service err, GetApp", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return nil, false
	}

	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, action, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return nil, false
	}

	return app, true
}

// checkAppRbacForAppOrJob checks if the user has the required permission for app or job
func (handler *PipelineConfigRestHandlerImpl) checkAppRbacForAppOrJob(w http.ResponseWriter, token string, appId int, action string) bool {
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, action)
	if !ok {
		common.WriteJsonResp(w, nil, "Unauthorized User", http.StatusForbidden)
		return false
	}
	return true
}

// getCiPipelineWithAuth gets the CI pipeline and checks if the user has the required permission
func (handler *PipelineConfigRestHandlerImpl) getCiPipelineWithAuth(w http.ResponseWriter, pipelineId int, token string, action string) (*pipelineConfig.CiPipeline, bool) {
	ciPipeline, err := handler.ciPipelineRepository.FindById(pipelineId)
	if err != nil {
		handler.Logger.Errorw("service err, FindById", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return nil, false
	}

	object := handler.enforcerUtil.GetAppRBACNameByAppId(ciPipeline.AppId)
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, action)
	if !ok {
		common.WriteJsonResp(w, nil, "Unauthorized User", http.StatusForbidden)
		return nil, false
	}

	return ciPipeline, true
}

// getQueryParamBool gets a boolean query parameter from the request
func (handler *PipelineConfigRestHandlerImpl) getQueryParamBool(r *http.Request, paramName string, defaultValue bool) bool {
	v := r.URL.Query()
	paramValue := defaultValue
	paramStr := v.Get(paramName)
	if len(paramStr) > 0 {
		var err error
		paramValue, err = strconv.ParseBool(paramStr)
		if err != nil {
			paramValue = defaultValue
		}
	}
	return paramValue
}
