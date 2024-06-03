/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package restHandler

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/pkg/attributes/bean"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/attributes"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type AttributesRestHandler interface {
	AddAttributes(w http.ResponseWriter, r *http.Request)
	UpdateAttributes(w http.ResponseWriter, r *http.Request)
	GetAttributesById(w http.ResponseWriter, r *http.Request)
	GetAttributesActiveList(w http.ResponseWriter, r *http.Request)
	GetAttributesByKey(w http.ResponseWriter, r *http.Request)
	AddDeploymentEnforcementConfig(w http.ResponseWriter, r *http.Request)
}

type AttributesRestHandlerImpl struct {
	logger            *zap.SugaredLogger
	enforcer          casbin.Enforcer
	userService       user.UserService
	attributesService attributes.AttributesService
}

func NewAttributesRestHandlerImpl(logger *zap.SugaredLogger, enforcer casbin.Enforcer,
	userService user.UserService, attributesService attributes.AttributesService) *AttributesRestHandlerImpl {
	userAuthHandler := &AttributesRestHandlerImpl{
		logger:            logger,
		enforcer:          enforcer,
		userService:       userService,
		attributesService: attributesService,
	}
	return userAuthHandler
}
func (handler AttributesRestHandlerImpl) AddAttributes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto bean.AttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, AddAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, AddAttributes", "payload", dto)
	resp, err := handler.attributesService.AddAttributes(&dto)
	if err != nil {
		handler.logger.Errorw("service err, AddAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) UpdateAttributes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var dto bean.AttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, UpdateAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, UpdateAttributes", "payload", dto)
	resp, err := handler.attributesService.UpdateAttributes(&dto)
	if err != nil {
		handler.logger.Errorw("service err, UpdateAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) GetAttributesById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusBadRequest)
		return
	}
	res, err := handler.attributesService.GetById(id)
	if err != nil {
		handler.logger.Errorw("service err, GetAttributesById", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) GetAttributesActiveList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	res, err := handler.attributesService.GetActiveList()
	if err != nil {
		handler.logger.Errorw("service err, GetHostUrlActive", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) GetAttributesByKey(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	/*token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, rbac.ResourceGlobal, rbac.ActionGet, "*"); !ok {
		WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}*/

	vars := mux.Vars(r)
	key := vars["key"]
	res, err := handler.attributesService.GetByKey(key)
	if err != nil {
		handler.logger.Errorw("service err, GetAttributesById", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler AttributesRestHandlerImpl) AddDeploymentEnforcementConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var dto bean.AttributesDto
	err = decoder.Decode(&dto)
	if err != nil {
		handler.logger.Errorw("request err, AddAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	handler.logger.Infow("request payload, AddAttributes", "payload", dto)
	resp, err := handler.attributesService.AddDeploymentEnforcementConfig(&dto)
	if err != nil {
		handler.logger.Errorw("service err, AddAttributes", "err", err, "payload", dto)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
