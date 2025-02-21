/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package team

import (
	"encoding/json"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/pkg/team/bean"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	user2 "github.com/devtron-labs/devtron/pkg/auth/user"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/team"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const PROJECT_DELETE_SUCCESS_RESP = "Project deleted successfully."

type TeamRestHandler interface {
	SaveTeam(w http.ResponseWriter, r *http.Request)
	FetchAll(w http.ResponseWriter, r *http.Request)
	FetchOne(w http.ResponseWriter, r *http.Request)
	UpdateTeam(w http.ResponseWriter, r *http.Request)
	DeleteTeam(w http.ResponseWriter, r *http.Request)

	FetchForAutocomplete(w http.ResponseWriter, r *http.Request)
}

type TeamRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	teamService     team.TeamService
	userService     user2.UserService
	validator       *validator.Validate
	enforcer        casbin.Enforcer
	userAuthService user2.UserAuthService
	deleteService   delete2.DeleteService
	cfg             *bean.Config
}

func NewTeamRestHandlerImpl(logger *zap.SugaredLogger,
	teamService team.TeamService,
	userService user2.UserService,
	enforcer casbin.Enforcer,
	validator *validator.Validate, userAuthService user2.UserAuthService,
	deleteService delete2.DeleteService,
) *TeamRestHandlerImpl {
	cfg := &bean.Config{}
	err := env.Parse(cfg)
	if err != nil {
		logger.Errorw("error occurred while parsing config ", "err", err)
		cfg.IgnoreAuthCheck = false
	}

	logger.Infow("team rest handler initialized", "ignoreAuthCheckValue", cfg.IgnoreAuthCheck)
	return &TeamRestHandlerImpl{
		logger:          logger,
		teamService:     teamService,
		userService:     userService,
		validator:       validator,
		enforcer:        enforcer,
		userAuthService: userAuthService,
		deleteService:   deleteService,
		cfg:             cfg,
	}
}

func (impl TeamRestHandlerImpl) SaveTeam(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean bean2.TeamRequest
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, SaveTeam", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	impl.logger.Infow("request payload, SaveTeam", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, SaveTeam", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceTeam, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	res, err := impl.teamService.Create(&bean)
	if err != nil {
		impl.logger.Errorw("service err, SaveTeam", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TeamRestHandlerImpl) FetchAll(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	res, err := impl.teamService.FetchAllActive()
	if err != nil {
		impl.logger.Errorw("service err, FetchAllActive", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying
	var result []bean2.TeamRequest
	for _, item := range res {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceTeam, casbin.ActionGet, item.Name); ok {
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (impl TeamRestHandlerImpl) FetchOne(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idi, err := strconv.Atoi(id)
	if err != nil {
		impl.logger.Errorw("request err, FetchOne", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	res, err := impl.teamService.FetchOne(idi)
	if err != nil {
		impl.logger.Errorw("service err, FetchOne", "err", err, "id", idi)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceTeam, casbin.ActionGet, res.Name); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TeamRestHandlerImpl) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean bean2.TeamRequest
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, UpdateTeam", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean.UserId = userId
	impl.logger.Infow("request payload, UpdateTeam", "err", err, "bean", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, UpdateTeam", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceTeam, casbin.ActionUpdate, bean.Name); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	res, err := impl.teamService.Update(&bean)
	if err != nil {
		impl.logger.Errorw("service err, UpdateTeam", "err", err, "bean", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TeamRestHandlerImpl) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var deleteRequest bean2.TeamRequest
	err = decoder.Decode(&deleteRequest)
	if err != nil {
		impl.logger.Errorw("request err, DeleteTeam", "err", err, "deleteRequest", deleteRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	deleteRequest.UserId = userId
	impl.logger.Infow("request payload, DeleteTeam", "err", err, "deleteRequest", deleteRequest)
	err = impl.validator.Struct(deleteRequest)
	if err != nil {
		impl.logger.Errorw("validation err, DeleteTeam", "err", err, "deleteRequest", deleteRequest)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//rbac starts
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceTeam, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//rbac ends
	err = impl.deleteService.DeleteTeam(&deleteRequest)
	if err != nil {
		impl.logger.Errorw("service err, DeleteTeam", "err", err, "deleteRequest", deleteRequest)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, PROJECT_DELETE_SUCCESS_RESP, http.StatusOK)
}

func (impl TeamRestHandlerImpl) FetchForAutocomplete(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	start := time.Now()
	teams, err := impl.teamService.FetchForAutocomplete()
	if err != nil {
		impl.logger.Errorw("service err, FetchForAutocomplete", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	dbElapsedTime := time.Since(start)
	token := r.Header.Get("token")
	var grantedTeams = teams
	start = time.Now()
	if !impl.cfg.IgnoreAuthCheck {
		grantedTeams = make([]bean2.TeamRequest, 0)
		// RBAC enforcer applying
		var teamNameList []string
		for _, item := range teams {
			teamNameList = append(teamNameList, strings.ToLower(item.Name))
		}

		result := impl.enforcer.EnforceInBatch(token, casbin.ResourceTeam, casbin.ActionGet, teamNameList)

		for _, item := range teams {
			if hasAccess := result[strings.ToLower(item.Name)]; hasAccess {
				grantedTeams = append(grantedTeams, item)
			}
		}
	}
	impl.logger.Infow("Team elapsed Time for enforcer", "dbElapsedTime", dbElapsedTime, "elapsedTime", time.Since(start),
		"envSize", len(grantedTeams))

	//RBAC enforcer Ends
	common.WriteJsonResp(w, err, grantedTeams, http.StatusOK)
}
