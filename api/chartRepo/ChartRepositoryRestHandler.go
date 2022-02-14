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

package chartRepo

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/chartRepo"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

const CHART_REPO_DELETE_SUCCESS_RESP = "Chart repo deleted successfully."

type ChartRepositoryRestHandler interface {
	GetChartRepoById(w http.ResponseWriter, r *http.Request)
	GetChartRepoList(w http.ResponseWriter, r *http.Request)
	CreateChartRepo(w http.ResponseWriter, r *http.Request)
	UpdateChartRepo(w http.ResponseWriter, r *http.Request)
	ValidateChartRepo(w http.ResponseWriter, r *http.Request)
	TriggerChartSyncManual(w http.ResponseWriter, r *http.Request)
	DeleteChartRepo(w http.ResponseWriter, r *http.Request)
}

type ChartRepositoryRestHandlerImpl struct {
	Logger                 *zap.SugaredLogger
	chartRepositoryService chartRepo.ChartRepositoryService
	userAuthService        user.UserService
	enforcer               casbin.Enforcer
	validator              *validator.Validate
	deleteService          delete2.DeleteService
}

func NewChartRepositoryRestHandlerImpl(Logger *zap.SugaredLogger, userAuthService user.UserService, chartRepositoryService chartRepo.ChartRepositoryService,
	enforcer casbin.Enforcer, validator *validator.Validate, deleteService delete2.DeleteService) *ChartRepositoryRestHandlerImpl {
	return &ChartRepositoryRestHandlerImpl{
		Logger:                 Logger,
		chartRepositoryService: chartRepositoryService,
		userAuthService:        userAuthService,
		enforcer:               enforcer,
		validator:              validator,
		deleteService:          deleteService,
	}
}

func (handler *ChartRepositoryRestHandlerImpl) GetChartRepoById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.Logger.Errorw("request err, GetChartRepoById", "err", err, "chart repo id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GetChartRepoById, app store", "chart repo id", id)
	res, err := handler.chartRepositoryService.GetChartRepoById(id)
	if err != nil {
		handler.Logger.Errorw("service err, GetChartRepoById, app store", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *ChartRepositoryRestHandlerImpl) GetChartRepoList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	handler.Logger.Infow("request payload, GetChartRepoList, app store")
	res, err := handler.chartRepositoryService.GetChartRepoList()
	if err != nil {
		handler.Logger.Errorw("service err, GetChartRepoList, app store", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *ChartRepositoryRestHandlerImpl) CreateChartRepo(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request *chartRepo.ChartRepoDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, CreateChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, CreateChartRepo", "err", err, "payload", request)
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "data validation error", InternalMessage: err.Error()}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	//rback block ends here
	request.UserId = userId
	handler.Logger.Infow("request payload, CreateChartRepo", "payload", request)
	res, err, validationResult := handler.chartRepositoryService.ValidateAndCreateChartRepo(request)
	if validationResult.CustomErrMsg != chartRepo.ValidationSuccessMsg {
		common.WriteJsonResp(w, nil, validationResult, http.StatusOK)
		return
	}
	if err != nil {
		handler.Logger.Errorw("service err, CreateChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *ChartRepositoryRestHandlerImpl) UpdateChartRepo(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request *chartRepo.ChartRepoDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, UpdateChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateChartRepo", "err", err, "payload", request)
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "data validation error", InternalMessage: err.Error()}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	//rback block ends here
	request.UserId = userId
	handler.Logger.Infow("request payload, UpdateChartRepo", "payload", request)
	res, err, validationResult := handler.chartRepositoryService.ValidateAndUpdateChartRepo(request)
	if validationResult.CustomErrMsg != chartRepo.ValidationSuccessMsg {
		common.WriteJsonResp(w, nil, validationResult, http.StatusOK)
		return
	}
	if err != nil {
		handler.Logger.Errorw("service err, UpdateChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *ChartRepositoryRestHandlerImpl) ValidateChartRepo(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request *chartRepo.ChartRepoDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, ValidateChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, ValidateChartRepo", "err", err, "payload", request)
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "data validation error", InternalMessage: err.Error()}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	request.UserId = userId
	handler.Logger.Infow("request payload, ValidateChartRepo", "payload", request)
	validationResult := handler.chartRepositoryService.ValidateChartRepo(request)
	common.WriteJsonResp(w, nil, validationResult, http.StatusOK)
}

func (handler *ChartRepositoryRestHandlerImpl) TriggerChartSyncManual(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	err2 := handler.chartRepositoryService.TriggerChartSyncManual()
	if err2 != nil {
		common.WriteJsonResp(w, err2, nil, http.StatusInternalServerError)
	} else {
		common.WriteJsonResp(w, nil, map[string]string{"status": "ok"}, http.StatusOK)
	}
}

func (handler *ChartRepositoryRestHandlerImpl) DeleteChartRepo(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request *chartRepo.ChartRepoDto
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, DeleteChartRepo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, DeleteChartRepo", "err", err, "payload", request)
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "data validation error", InternalMessage: err.Error()}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//rbac starts here
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//rbac ends here
	request.UserId = userId
	handler.Logger.Infow("request payload, DeleteChartRepo", "payload", request)
	err = handler.deleteService.DeleteChartRepo(request)
	if err != nil {
		handler.Logger.Errorw("err in deleting chart repo", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, CHART_REPO_DELETE_SUCCESS_RESP, http.StatusOK)
}