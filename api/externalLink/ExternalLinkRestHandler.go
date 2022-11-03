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

package externalLink

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/externalLink"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type ExternalLinkRestHandler interface {
	CreateExternalLinks(w http.ResponseWriter, r *http.Request)
	GetExternalLinkMonitoringTools(w http.ResponseWriter, r *http.Request)
	GetExternalLinks(w http.ResponseWriter, r *http.Request)
	UpdateExternalLink(w http.ResponseWriter, r *http.Request)
	DeleteExternalLink(w http.ResponseWriter, r *http.Request) // Update is_active to false link
}
type ExternalLinkRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	externalLinkService externalLink.ExternalLinkService
	userService         user.UserService
	enforcer            casbin.Enforcer
}

func NewExternalLinkRestHandlerImpl(logger *zap.SugaredLogger,
	externalLinkService externalLink.ExternalLinkService,
	userService user.UserService,
	enforcer casbin.Enforcer,
) *ExternalLinkRestHandlerImpl {
	return &ExternalLinkRestHandlerImpl{
		logger:              logger,
		externalLinkService: externalLinkService,
		userService:         userService,
		enforcer:            enforcer,
	}
}

func (impl ExternalLinkRestHandlerImpl) CreateExternalLinks(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var beans []*externalLink.ExternalLinkDto
	err = decoder.Decode(&beans)
	if err != nil {
		impl.logger.Errorw("request err, SaveLink", "err", err, "payload", beans)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//find user role
	userRole := ""
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); ok {
		userRole = externalLink.SUPER_ADMIN_ROLE
	} else if ok = impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, "*"); ok {
		userRole = externalLink.ADMIN_ROLE
	}
	if userRole == "" {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	res, err := impl.externalLinkService.Create(beans, userId, userRole)
	if err != nil {
		impl.logger.Errorw("service err, SaveLink", "err", err, "payload", beans)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
func (impl ExternalLinkRestHandlerImpl) GetExternalLinkMonitoringTools(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// auth free api as we using this for multiple places

	res, err := impl.externalLinkService.GetAllActiveTools()
	if err != nil {
		impl.logger.Errorw("service err, GetAllActiveTools", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
func (impl ExternalLinkRestHandlerImpl) GetExternalLinks(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	v := r.URL.Query()
	clusterId := v.Get("clusterId")
	linkType := v.Get("type")
	identifier := v.Get("identifier")

	token := r.Header.Get("token")
	if len(identifier) == 0 && len(linkType) == 0 && len(clusterId) == 0 {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		id, err := strconv.Atoi(clusterId)
		res, err := impl.externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, id, externalLink.SUPER_ADMIN_ROLE, int(userId))
		if err != nil {
			impl.logger.Errorw("service err, FetchAllActive", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, err, res, http.StatusOK)
		return

	} else if len(identifier) != 0 && len(linkType) != 0 && len(clusterId) != 0 {
		//if user access level for this app
		//if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, "*"); !ok {
		//	common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		//	return
		//}
		id, err := strconv.Atoi(clusterId)
		if err != nil {
			impl.logger.Errorw("error occurred while parsing cluster_id", "clusterId", clusterId, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		linkIdentifier := &externalLink.LinkIdentifier{
			Type:       linkType,
			Identifier: identifier,
			ClusterId:  0,
		}
		res, err := impl.externalLinkService.FetchAllActiveLinksByLinkIdentifier(linkIdentifier, id, externalLink.ADMIN_ROLE, int(userId))
		if err != nil {
			impl.logger.Errorw("service err, FetchAllActive", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, err, res, http.StatusOK)
		return
	}

	impl.logger.Errorw("invalid request, FetchAllActive external links", "err", err)
	common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	return

}

func (impl ExternalLinkRestHandlerImpl) UpdateExternalLink(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var bean externalLink.ExternalLinkDto
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, Update Link", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	//find user role
	userRole := ""
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); ok {
		userRole = externalLink.SUPER_ADMIN_ROLE
	} else if ok = impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, "*"); ok {
		userRole = externalLink.ADMIN_ROLE
	}
	if userRole == "" {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	res, err := impl.externalLinkService.Update(&bean, userRole)
	if err != nil {
		impl.logger.Errorw("service err, Update Links", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ExternalLinkRestHandlerImpl) DeleteExternalLink(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	params := mux.Vars(r)
	id := params["id"]
	linkId, err := strconv.Atoi(id)
	if err != nil {
		impl.logger.Errorw("request err, DeleteExternalLink", "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	userRole := ""
	//find user role
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); ok {
		userRole = externalLink.SUPER_ADMIN_ROLE
	} else if ok = impl.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionDelete, "*"); ok {
		userRole = externalLink.ADMIN_ROLE
	}
	if userRole == "" {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//request for delete link
	res, err := impl.externalLinkService.DeleteLink(linkId, userId, userRole)
	if err != nil {
		impl.logger.Errorw("service err, delete Links", "err", err, "linkId", linkId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
