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
	"fmt"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/externalLink"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"github.com/juju/errors"
	"go.uber.org/zap"
)

type ExternalLinkRestHandler interface {
	CreateExternalLinks(w http.ResponseWriter, r *http.Request)
	GetExternalLinkMonitoringTools(w http.ResponseWriter, r *http.Request)
	GetExternalLinks(w http.ResponseWriter, r *http.Request)
	GetExternalLinksV2(w http.ResponseWriter, r *http.Request)
	UpdateExternalLink(w http.ResponseWriter, r *http.Request)
	DeleteExternalLink(w http.ResponseWriter, r *http.Request) // Update is_active to false link
}
type ExternalLinkRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	externalLinkService externalLink.ExternalLinkService
	userService         user.UserService
	enforcer            casbin.Enforcer
	enforcerUtil        rbac.EnforcerUtil
}

func NewExternalLinkRestHandlerImpl(logger *zap.SugaredLogger,
	externalLinkService externalLink.ExternalLinkService,
	userService user.UserService,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil,
) *ExternalLinkRestHandlerImpl {
	return &ExternalLinkRestHandlerImpl{
		logger:              logger,
		externalLinkService: externalLinkService,
		userService:         userService,
		enforcer:            enforcer,
		enforcerUtil:        enforcerUtil,
	}
}

func (impl ExternalLinkRestHandlerImpl) roleCheckHelper(w http.ResponseWriter, r *http.Request, action string) (int32, string, error) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return userId, "", fmt.Errorf("unauthorized error")
	}
	userRole := ""
	v := r.URL.Query()
	//put this check from identifiers itself,don't get this appname from query params
	appId := v.Get("appId")
	token := r.Header.Get("token")
	if v.Has("appId") {
		id, err := strconv.Atoi(appId)
		if err != nil {
			impl.logger.Errorw("error occurred while converting appId to integer", "err", err, "appId", appId)
			common.WriteJsonResp(w, errors.New("Invalid request"), nil, http.StatusBadRequest)
			return userId, "", fmt.Errorf("invalid request query param appId = %s", appId)
		}
		object := impl.enforcerUtil.GetAppRBACNameByAppId(id)
		if ok := impl.enforcer.Enforce(token, casbin.ResourceApplications, action, object); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return userId, "", fmt.Errorf("unauthorized error")
		}
		userRole = externalLink.ADMIN_ROLE
	} else {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, action, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return userId, "", fmt.Errorf("unauthorized error")
		}
		userRole = externalLink.SUPER_ADMIN_ROLE
	}
	return userId, userRole, nil
}
func (impl ExternalLinkRestHandlerImpl) CreateExternalLinks(w http.ResponseWriter, r *http.Request) {
	userId, userRole, err := impl.roleCheckHelper(w, r, casbin.ActionCreate)
	if err != nil {
		impl.logger.Errorw("error in CreateExternalLinks ", "err", err)
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
	res, err := impl.externalLinkService.Create(beans, userId, userRole)
	if err != nil && err != pg.ErrNoRows {
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

	// auth free api as we are using this for multiple places
	res, err := impl.externalLinkService.GetAllActiveTools()
	if err != nil && err != pg.ErrNoRows {
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
		clusterIdNumber := 0
		res, err := impl.externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, clusterIdNumber)
		if err != nil {
			impl.logger.Errorw("service err, FetchAllActive", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, err, res, http.StatusOK)
		return

	} else if len(identifier) != 0 && len(linkType) != 0 { //api to get external links from app-level external links tab and from app-details page
		clusterIdNumber := 0
		if len(clusterId) != 0 { //api call from app-detail page
			clusterIdNumber, err = strconv.Atoi(clusterId)
			if err != nil {
				impl.logger.Errorw("error occurred while parsing cluster_id", "clusterId", clusterId, "err", err)
				common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
				return
			}
		}
		linkIdentifier := &externalLink.LinkIdentifier{
			Type:       linkType,
			Identifier: identifier,
			ClusterId:  0,
		}
		res, err := impl.externalLinkService.FetchAllActiveLinksByLinkIdentifier(linkIdentifier, clusterIdNumber)
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

func (impl ExternalLinkRestHandlerImpl) GetExternalLinksV2(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	v := r.URL.Query()
	clusterId := v.Get("clusterId")
	linkType := v.Get("type")
	identifier := v.Get("identifier")

	externalLinkAndMonitoringTools := externalLink.ExternalLinkAndMonitoringToolDTO{}
	externalLinks := []*externalLink.ExternalLinkDto{}

	tools, err := impl.externalLinkService.GetAllActiveTools()
	if err != nil && err != pg.ErrNoRows {
		impl.logger.Errorw("service err, GetAllActiveTools", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	token := r.Header.Get("token")
	if len(identifier) == 0 && len(linkType) == 0 && len(clusterId) == 0 {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		clusterIdNumber := 0
		externalLinks, err = impl.externalLinkService.FetchAllActiveLinksByLinkIdentifier(nil, clusterIdNumber)
		if err != nil {
			impl.logger.Errorw("service err, FetchAllActive", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}

	} else if len(identifier) != 0 && len(linkType) != 0 { //api to get external links from app-level external links tab and from app-details page
		clusterIdNumber := 0
		if len(clusterId) != 0 { //api call from app-detail page
			clusterIdNumber, err = strconv.Atoi(clusterId)
			if err != nil {
				impl.logger.Errorw("error occurred while parsing cluster_id", "clusterId", clusterId, "err", err)
				common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
				return
			}
		}
		linkIdentifier := &externalLink.LinkIdentifier{
			Type:       linkType,
			Identifier: identifier,
			ClusterId:  0,
		}
		externalLinks, err = impl.externalLinkService.FetchAllActiveLinksByLinkIdentifier(linkIdentifier, clusterIdNumber)
		if err != nil {
			impl.logger.Errorw("service err, FetchAllActive", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	externalLinkAndMonitoringTools.ExternalLinks = externalLinks
	externalLinkAndMonitoringTools.Tools = tools

	common.WriteJsonResp(w, err, externalLinkAndMonitoringTools, http.StatusOK)
	return

}

func (impl ExternalLinkRestHandlerImpl) UpdateExternalLink(w http.ResponseWriter, r *http.Request) {
	userId, userRole, err := impl.roleCheckHelper(w, r, casbin.ActionUpdate)
	if err != nil {
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

	res, err := impl.externalLinkService.Update(&bean, userRole)
	if err != nil {
		impl.logger.Errorw("service err, Update Links", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ExternalLinkRestHandlerImpl) DeleteExternalLink(w http.ResponseWriter, r *http.Request) {
	userId, userRole, err := impl.roleCheckHelper(w, r, casbin.ActionDelete)
	if err != nil {
		return
	}
	params := mux.Vars(r)
	id := params["id"]
	linkId, err := strconv.Atoi(id)
	if err != nil {
		impl.logger.Errorw("request err, DeleteExternalLink", "id", id, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	res, err := impl.externalLinkService.DeleteLink(linkId, userId, userRole)
	if err != nil {
		impl.logger.Errorw("service err, delete Links", "err", err, "linkId", linkId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
