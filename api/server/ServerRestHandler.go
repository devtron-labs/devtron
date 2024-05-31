/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/server"
	serverBean "github.com/devtron-labs/devtron/pkg/server/bean"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

type ServerRestHandler interface {
	GetServerInfo(w http.ResponseWriter, r *http.Request)
	HandleServerAction(w http.ResponseWriter, r *http.Request)
}

type ServerRestHandlerImpl struct {
	logger        *zap.SugaredLogger
	serverService server.ServerService
	userService   user.UserService
	enforcer      casbin.Enforcer
	validator     *validator.Validate
}

func NewServerRestHandlerImpl(logger *zap.SugaredLogger,
	serverService server.ServerService,
	userService user.UserService,
	enforcer casbin.Enforcer,
	validator *validator.Validate,
) *ServerRestHandlerImpl {
	return &ServerRestHandlerImpl{
		logger:        logger,
		serverService: serverService,
		userService:   userService,
		enforcer:      enforcer,
		validator:     validator,
	}
}

func (impl ServerRestHandlerImpl) GetServerInfo(w http.ResponseWriter, r *http.Request) {
	// check if user is logged in or not
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	showServerStatus := true
	showServerStatusQueryParam := r.URL.Query().Get("showServerStatus")
	if len(showServerStatusQueryParam) != 0 {
		if showServerStatusQueryParam == "false" {
			showServerStatus = false
		} else if showServerStatusQueryParam == "true" {
			showServerStatus = true
		}
	}

	// service call
	res, err := impl.serverService.GetServerInfo(showServerStatus)
	if err != nil {
		impl.logger.Errorw("service err, GetServerInfo", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ServerRestHandlerImpl) HandleServerAction(w http.ResponseWriter, r *http.Request) {
	// check if user is logged in or not
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	var serverActionRequestDto *serverBean.ServerActionRequestDto
	err = decoder.Decode(&serverActionRequestDto)
	if err != nil {
		impl.logger.Errorw("error in decoding request in HandleServerAction", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = impl.validator.Struct(serverActionRequestDto)
	if err != nil {
		impl.logger.Errorw("error in validating request in HandleServerAction", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// service call
	res, err := impl.serverService.HandleServerAction(userId, serverActionRequestDto)
	if err != nil {
		impl.logger.Errorw("service err, HandleServerAction", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
