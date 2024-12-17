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

package terminal

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/pkg/cluster/rbac"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/sql/models"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/clusterTerminalAccess"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type UserTerminalAccessRestHandler interface {
	StartTerminalSession(w http.ResponseWriter, r *http.Request)
	UpdateTerminalSession(w http.ResponseWriter, r *http.Request)
	UpdateTerminalShellSession(w http.ResponseWriter, r *http.Request)
	FetchTerminalStatus(w http.ResponseWriter, r *http.Request)
	StopTerminalSession(w http.ResponseWriter, r *http.Request)
	DisconnectTerminalSession(w http.ResponseWriter, r *http.Request)
	DisconnectAllTerminalSessionAndRetry(w http.ResponseWriter, r *http.Request)
	FetchTerminalPodEvents(w http.ResponseWriter, r *http.Request)
	FetchTerminalPodManifest(w http.ResponseWriter, r *http.Request)
	ValidateShell(w http.ResponseWriter, r *http.Request)
	EditPodManifest(w http.ResponseWriter, r *http.Request)
}

type validShellResponse struct {
	IsValidShell bool   `json:"isValidShell"`
	ErrorReason  string `json:"errorReason"`
	ShellName    string `json:"shellName"`
}

type UserTerminalAccessRestHandlerImpl struct {
	Logger                    *zap.SugaredLogger
	UserTerminalAccessService clusterTerminalAccess.UserTerminalAccessService
	clusterRbacService        rbac.ClusterRbacService
	Enforcer                  casbin.Enforcer
	UserService               user.UserService
	validator                 *validator.Validate
}

func NewUserTerminalAccessRestHandlerImpl(logger *zap.SugaredLogger, userTerminalAccessService clusterTerminalAccess.UserTerminalAccessService, Enforcer casbin.Enforcer,
	UserService user.UserService, validator *validator.Validate,
	clusterRbacService rbac.ClusterRbacService) *UserTerminalAccessRestHandlerImpl {
	return &UserTerminalAccessRestHandlerImpl{
		Logger:                    logger,
		UserTerminalAccessService: userTerminalAccessService,
		Enforcer:                  Enforcer,
		UserService:               UserService,
		validator:                 validator,
		clusterRbacService:        clusterRbacService,
	}
}
func (handler UserTerminalAccessRestHandlerImpl) ValidateShell(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	podName := vars["podName"]
	namespace := vars["namespace"]
	shellName := vars["shellName"]
	containerName := vars["containerName"]
	clusterId, err := strconv.Atoi(vars["clusterId"])
	if err != nil {
		handler.Logger.Errorw("error in parsing clusterId from request", "clusterId", clusterId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	authenticated, err := handler.clusterRbacService.CheckAuthorisationForNodeWithClusterId(token, clusterId, "", casbin.ActionCreate)
	if err != nil {
		handler.Logger.Errorw("error in CheckAuthorisationForNodeWithClusterId", "clusterId", clusterId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	res, shell, err := handler.UserTerminalAccessService.ValidateShell(podName, namespace, shellName, containerName, clusterId)
	reason := ""
	if err != nil {
		reason = err.Error()
	}
	resp := validShellResponse{IsValidShell: res, ErrorReason: reason, ShellName: shell}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
func (handler UserTerminalAccessRestHandlerImpl) StartTerminalSession(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request models.UserTerminalSessionRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, StartTerminalSession", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, StartTerminalSession", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	authenticated, err := handler.clusterRbacService.CheckAuthorisationForNodeWithClusterId(token, request.ClusterId, request.NodeName, casbin.ActionCreate)
	if err != nil {
		handler.Logger.Errorw("error in CheckAuthorisationForNodeWithClusterId", "clusterId", request.ClusterId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	sessionResponse, err := handler.UserTerminalAccessService.StartTerminalSession(r.Context(), &request)
	if err != nil {
		handler.Logger.Errorw("service err, StartTerminalSession", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, sessionResponse, http.StatusOK)
}

func (handler UserTerminalAccessRestHandlerImpl) UpdateTerminalSession(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request models.UserTerminalSessionRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, UpdateTerminalSession", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateTerminalSession", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	authenticated, err := handler.clusterRbacService.CheckAuthorisationForNodeWithClusterId(token, request.ClusterId, request.NodeName, casbin.ActionUpdate)
	if err != nil {
		handler.Logger.Errorw("error in CheckAuthorisationForNodeWithClusterId", "clusterId", request.ClusterId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	sessionResponse, err := handler.UserTerminalAccessService.UpdateTerminalSession(r.Context(), &request)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateTerminalSession", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, sessionResponse, http.StatusOK)
}

func (handler UserTerminalAccessRestHandlerImpl) UpdateTerminalShellSession(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request models.UserTerminalShellSessionRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, UpdateTerminalShellSession", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, UpdateTerminalShellSession", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	authenticated, err := handler.checkRbacForTerminalWithTerminalAccessId(token, request.TerminalAccessId)
	if err != nil {
		handler.Logger.Errorw("error in UpdateTerminalShellSession", "terminalAccessId", request.TerminalAccessId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	sessionResponse, err := handler.UserTerminalAccessService.UpdateTerminalShellSession(r.Context(), &request)
	if err != nil {
		handler.Logger.Errorw("service err, UpdateTerminalShellSession", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, sessionResponse, http.StatusOK)
}

func (handler UserTerminalAccessRestHandlerImpl) FetchTerminalStatus(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	terminalAccessId, err := strconv.Atoi(vars["terminalAccessId"])
	namespace := vars["namespace"]
	shellName := vars["shellName"]
	containerName := vars["containerName"]
	if err != nil {
		handler.Logger.Errorw("request err, FetchTerminalStatus", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	authenticated, err := handler.checkRbacForTerminalWithTerminalAccessId(token, terminalAccessId)
	if err != nil {
		handler.Logger.Errorw("error in FetchTerminalStatus", "terminalAccessId", terminalAccessId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	sessionResponse, err := handler.UserTerminalAccessService.FetchTerminalStatus(r.Context(), terminalAccessId, namespace, containerName, shellName)
	if err != nil {
		handler.Logger.Errorw("service err, FetchTerminalStatus", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, sessionResponse, http.StatusOK)
}

func (handler UserTerminalAccessRestHandlerImpl) FetchTerminalPodEvents(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	terminalAccessId, err := strconv.Atoi(vars["terminalAccessId"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchTerminalPodEvents", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	authenticated, err := handler.checkRbacForTerminalWithTerminalAccessId(token, terminalAccessId)
	if err != nil {
		handler.Logger.Errorw("error in FetchTerminalPodEvents", "terminalAccessId", terminalAccessId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	resp, err := handler.UserTerminalAccessService.FetchPodEvents(r.Context(), terminalAccessId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchTerminalPodEvents", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler UserTerminalAccessRestHandlerImpl) FetchTerminalPodManifest(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	terminalAccessId, err := strconv.Atoi(vars["terminalAccessId"])
	if err != nil {
		handler.Logger.Errorw("request err, FetchTerminalPodManifest", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	authenticated, err := handler.checkRbacForTerminalWithTerminalAccessId(token, terminalAccessId)
	if err != nil {
		handler.Logger.Errorw("error in FetchTerminalPodManifest", "terminalAccessId", terminalAccessId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	podManifest, err := handler.UserTerminalAccessService.FetchPodManifest(r.Context(), terminalAccessId)
	if err != nil {
		handler.Logger.Errorw("service err, FetchTerminalPodManifest", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, podManifest, http.StatusOK)
}

func (handler UserTerminalAccessRestHandlerImpl) DisconnectTerminalSession(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	terminalAccessId, err := strconv.Atoi(vars["terminalAccessId"])
	if err != nil {
		handler.Logger.Errorw("request err, DisconnectTerminalSession", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	authenticated, err := handler.checkRbacForTerminalWithTerminalAccessId(token, terminalAccessId)
	if err != nil {
		handler.Logger.Errorw("error in DisconnectTerminalSession", "terminalAccessId", terminalAccessId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	err = handler.UserTerminalAccessService.DisconnectTerminalSession(r.Context(), terminalAccessId)
	if err != nil {
		handler.Logger.Errorw("service err, DisconnectTerminalSession", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler UserTerminalAccessRestHandlerImpl) StopTerminalSession(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	terminalAccessId, err := strconv.Atoi(vars["terminalAccessId"])
	if err != nil {
		handler.Logger.Errorw("request err, StopTerminalSession", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	authenticated, err := handler.checkRbacForTerminalWithTerminalAccessId(token, terminalAccessId)
	if err != nil {
		handler.Logger.Errorw("error in StopTerminalSession", "terminalAccessId", terminalAccessId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	handler.UserTerminalAccessService.StopTerminalSession(r.Context(), terminalAccessId)
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler UserTerminalAccessRestHandlerImpl) DisconnectAllTerminalSessionAndRetry(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request models.UserTerminalSessionRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, DisconnectAllTerminalSessionAndRetry", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	err = handler.validator.Struct(request)
	if err != nil {
		handler.Logger.Errorw("validation err, DisconnectAllTerminalSessionAndRetry", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	authenticated, err := handler.clusterRbacService.CheckAuthorisationForNodeWithClusterId(token, request.ClusterId, request.NodeName, casbin.ActionUpdate)
	if err != nil {
		handler.Logger.Errorw("error in CheckAuthorisationForNodeWithClusterId", "clusterId", request.ClusterId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	handler.UserTerminalAccessService.DisconnectAllSessionsForUser(r.Context(), userId)
	sessionResponse, err := handler.UserTerminalAccessService.StartTerminalSession(r.Context(), &request)
	if err != nil {
		handler.Logger.Errorw("service err, DisconnectAllTerminalSessionAndRetry", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, sessionResponse, http.StatusOK)
}

func (handler UserTerminalAccessRestHandlerImpl) EditPodManifest(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.UserService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var request models.UserTerminalSessionRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.Logger.Errorw("request err, StartTerminalSession", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	authenticated, err := handler.clusterRbacService.CheckAuthorisationForNodeWithClusterId(token, request.ClusterId, request.NodeName, casbin.ActionUpdate)
	if err != nil {
		handler.Logger.Errorw("error in CheckAuthorisationForNodeWithClusterId", "clusterId", request.ClusterId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	manifest, err := handler.UserTerminalAccessService.EditTerminalPodManifest(r.Context(), &request, false)
	if err != nil {
		handler.Logger.Errorw("service err, FetchTerminalPodManifest", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, manifest, http.StatusOK)
}

func (handler UserTerminalAccessRestHandlerImpl) checkRbacForTerminalWithTerminalAccessId(token string, terminalAccessId int) (bool, error) {
	terminalAccessSessionData, present := handler.UserTerminalAccessService.GetTerminalAccessSessionDataFromCacheById(terminalAccessId)
	if !present {
		return false, errors.New("terminal access session not found")
	}

	authenticated, err := handler.clusterRbacService.CheckAuthorisationForNodeWithClusterId(token, terminalAccessSessionData.ClusterId, terminalAccessSessionData.NodeName, casbin.ActionUpdate)
	if err != nil {
		handler.Logger.Errorw("error encountered in checkRbacForTerminalWithTerminalAccessId", "err", err)
		return false, err

	}
	return authenticated, nil
}
