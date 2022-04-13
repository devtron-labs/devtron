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
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

type ExternalLinkoutRestHandler interface {
	CreateExternalLinks(w http.ResponseWriter, r *http.Request)
	GetAllTools(w http.ResponseWriter, r *http.Request)
	GetAllLinks(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request) // Update is_active to false link

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
	decoder := json.NewDecoder(r.Body)

	var bean externalLinkout.ExternalLinkoutRequest
	err := decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, SaveTeam", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	res, err := impl.externalLinkoutService.Create(&bean)
	if err != nil {
		impl.logger.Errorw("service err, SaveTeam", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ExternalLinkoutRestHandlerImpl) GetAllTools(w http.ResponseWriter, r *http.Request) {
	res, err := impl.externalLinkoutService.GetAllActiveTools()
	if err != nil {
		impl.logger.Errorw("service err, GetAllActiveTools", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)

}

func (impl ExternalLinkoutRestHandlerImpl) GetAllLinks(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	clusterIdString := v.Get("clusterId")
	var clusterIds []int
	if clusterIdString != "" {
		clusterIdSlices := strings.Split(clusterIdString, ",")
		for _, clusterId := range clusterIdSlices {
			id, err := strconv.Atoi(clusterId)
			if err != nil {
				common.WriteJsonResp(w, err, "please send valid cluster Ids", http.StatusBadRequest)
				return
			}
			clusterIds = append(clusterIds, id)
		}
	}
	res, err := impl.externalLinkoutService.FetchAllActiveLinks(clusterIds)
	if err != nil {
		impl.logger.Errorw("service err, FetchAllActive", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ExternalLinkoutRestHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var bean externalLinkout.ExternalLinkoutRequest
	err := decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, Update Link", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	impl.logger.Infow("request payload, UpdateTeam", "err", err, "bean", bean)

	res, err := impl.externalLinkoutService.Update(&bean)
	if err != nil {
		impl.logger.Errorw("service err, Update Links", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl ExternalLinkoutRestHandlerImpl) Delete(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var bean externalLinkout.ExternalLinkoutRequest
	err := decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, UpdateTeam", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, UpdateTeam", "err", err, "bean", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, UpdateTeam", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	res, err := impl.externalLinkoutService.Delete(&bean)
	if err != nil {
		impl.logger.Errorw("service err, UpdateLink", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
