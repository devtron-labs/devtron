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
	"context"
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/apis/devtron/v1"
	"github.com/devtron-labs/devtron/pkg/apis/devtron/v1/validation"
	"github.com/devtron-labs/devtron/pkg/appClone/batch"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"net/http"
)

type BatchOperationRestHandler interface {
	Operate(w http.ResponseWriter, r *http.Request)
}

type BatchOperationRestHandlerImpl struct {
	userAuthService user.UserService
	enforcer        casbin.Enforcer
	workflowAction  batch.WorkflowAction
	teamService     team.TeamService
	logger          *zap.SugaredLogger
	enforcerUtil    rbac.EnforcerUtil
}

func NewBatchOperationRestHandlerImpl(userAuthService user.UserService, enforcer casbin.Enforcer, workflowAction batch.WorkflowAction,
	teamService team.TeamService, logger *zap.SugaredLogger, enforcerUtil rbac.EnforcerUtil) *BatchOperationRestHandlerImpl {
	return &BatchOperationRestHandlerImpl{
		userAuthService: userAuthService,
		enforcer:        enforcer,
		workflowAction:  workflowAction,
		teamService:     teamService,
		logger:          logger,
		enforcerUtil:    enforcerUtil,
	}
}

func (handler BatchOperationRestHandlerImpl) Operate(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var data map[string]interface{}
	err = decoder.Decode(&data)
	if err != nil {
		handler.logger.Errorw("request err, Operate", "err", err, "payload", data)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//validate request
	emptyProps := v1.InheritedProps{}

	if wf, ok := data["workflow"]; ok {
		var workflow v1.Workflow
		wfd, err := json.Marshal(wf)
		if err != nil {
			handler.logger.Errorw("marshaling err, Operate", "err", err, "wf", wf)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(wfd, &workflow)
		if err != nil {
			handler.logger.Errorw("marshaling err, Operate", "err", err, "workflow", workflow)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		if workflow.Destination.App == nil || len(*workflow.Destination.App) == 0 {
			common.WriteJsonResp(w, errors.New("app name cannot be empty"), nil, http.StatusBadRequest)
		}
		rbacString := handler.enforcerUtil.GetProjectAdminRBACNameBYAppName(*workflow.Destination.App)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, rbacString); !ok {
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}

		ctx, cancel := context.WithCancel(r.Context())
		if cn, ok := w.(http.CloseNotifier); ok {
			go func(done <-chan struct{}, closed <-chan bool) {
				select {
				case <-done:
				case <-closed:
					cancel()
				}
			}(ctx.Done(), cn.CloseNotify())
		}

		ctx = context.WithValue(ctx, "token", token)

		err = handler.workflowAction.Execute(&workflow, emptyProps, ctx)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	common.WriteJsonResp(w, nil, `{"result": "ok"}`, http.StatusOK)
	//panic("implement me")
}

func validatePipeline(pipeline *v1.Pipeline, props v1.InheritedProps) error {
	if pipeline.Build == nil && pipeline.Deployment == nil {
		return nil
	} else if pipeline.Build != nil {
		pipeline.Build.UpdateMissingProps(props)
		return validation.ValidateBuild(pipeline.Build)
	} else if pipeline.Deployment != nil {
		return validation.ValidateDeployment(pipeline.Deployment, props)
	}
	return nil
}

func executePipeline(pipeline *v1.Pipeline, props v1.InheritedProps) error {
	if pipeline.Build == nil && pipeline.Deployment == nil {
		return nil
	} else if pipeline.Build != nil {
		pipeline.Build.UpdateMissingProps(props)
		return validation.ValidateBuild(pipeline.Build)
	} else if pipeline.Deployment != nil {
		//return batch.ExecuteDeployment(pipeline.Deployment, props)
	}
	return nil
}
