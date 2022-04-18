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

package externalLinkout

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/externalLinkout"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type ExternalLinkoutRestHandler interface {
	CreateExternalLinks(w http.ResponseWriter, r *http.Request)
	GetExternalLinksTools(w http.ResponseWriter, r *http.Request)
	GetExternalLinks(w http.ResponseWriter, r *http.Request)
	UpdateExternalLinks(w http.ResponseWriter, r *http.Request)
	DeleteExternalLinks(w http.ResponseWriter, r *http.Request) // Update is_active to false link
}
type ExternalLinkoutRestHandlerImpl struct {
	logger                 *zap.SugaredLogger
	externalLinkoutService externalLinkout.ExternalLinkoutService
	userService            user.UserService
	validator              *validator.Validate
	enforcer               casbin.Enforcer
	userAuthService        user.UserAuthService
	deleteService          delete2.DeleteService
}

func NewExternalLinkoutRestHandlerImpl(logger *zap.SugaredLogger,
	externalLinkoutService externalLinkout.ExternalLinkoutService,
	userService user.UserService,
	enforcer casbin.Enforcer,
	validator *validator.Validate, userAuthService user.UserAuthService,
	deleteService delete2.DeleteService,
) *ExternalLinkoutRestHandlerImpl {
	return &ExternalLinkoutRestHandlerImpl{
		logger:                 logger,
		externalLinkoutService: externalLinkoutService,
		userService:            userService,
		validator:              validator,
		enforcer:               enforcer,
		userAuthService:        userAuthService,
		deleteService:          deleteService,
	}
}
func (impl ExternalLinkoutRestHandlerImpl) CreateExternalLinks(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var beans []*externalLinkout.ExternalLinkoutRequest
	err = decoder.Decode(&beans)
	if err != nil {
		impl.logger.Errorw("request err, SaveLink", "err", err, "payload", beans)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	res, err := impl.externalLinkoutService.Create(beans, userId)
	if err != nil {
		impl.logger.Errorw("service err, SaveLink", "err", err, "payload", beans)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
func (impl ExternalLinkoutRestHandlerImpl) GetExternalLinksTools(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// auth free api as we using this for multiple places
	/*
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	*/

	res, err := impl.externalLinkoutService.GetAllActiveTools()
	if err != nil {
		impl.logger.Errorw("service err, GetAllActiveTools", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
func (impl ExternalLinkoutRestHandlerImpl) GetExternalLinks(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	v := r.URL.Query()
	id := v.Get("clusterId")
	clusterId := 0
	if len(id) > 0 {
		clusterId, err = strconv.Atoi(id)
		if err != nil {
			impl.logger.Errorw("request err, FetchOne", "err", err, "id", id)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}

	//apply auth only in case when requested for all links
	if clusterId == 0 {
		token := r.Header.Get("token")
		if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	}

	res, err := impl.externalLinkoutService.FetchAllActiveLinks(clusterId)
	if err != nil {
		impl.logger.Errorw("service err, FetchAllActive", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
func (impl ExternalLinkoutRestHandlerImpl) UpdateExternalLinks(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var bean externalLinkout.ExternalLinkoutRequest
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, Update Link", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId

	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	impl.logger.Infow("request payload, UpdateLink", "err", err, "bean", bean)
	res, err := impl.externalLinkoutService.Update(&bean)
	if err != nil {
		impl.logger.Errorw("service err, Update Links", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
func (impl ExternalLinkoutRestHandlerImpl) DeleteExternalLinks(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	id := params["id"]
	idi, err := strconv.Atoi(id)
	if err != nil {
		impl.logger.Errorw("request err, FetchOne", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	var bean externalLinkout.ExternalLinkoutRequest
	bean.Id = idi
	res, err := impl.externalLinkoutService.DeleteLink(idi, userId)
	if err != nil {
		impl.logger.Errorw("service err, Update Links", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
