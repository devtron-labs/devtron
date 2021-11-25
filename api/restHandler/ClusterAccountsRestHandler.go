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
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	request "github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type ClusterAccountsRestHandler interface {
	Save(w http.ResponseWriter, r *http.Request)
	Get(w http.ResponseWriter, r *http.Request)
	GetByEnvironment(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)

	FindById(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
}

type ClusterAccountsRestHandlerImpl struct {
	clusterAccountsService request.ClusterAccountsService
	logger                 *zap.SugaredLogger
	userService            user.UserService
}

func NewClusterAccountsRestHandlerImpl(svc request.ClusterAccountsService, logger *zap.SugaredLogger, userService user.UserService) *ClusterAccountsRestHandlerImpl {
	return &ClusterAccountsRestHandlerImpl{
		clusterAccountsService: svc,
		logger:                 logger,
		userService:            userService,
	}
}

func (impl ClusterAccountsRestHandlerImpl) Save(w http.ResponseWriter, r *http.Request) {
	impl.logger.Debugw("save cluster account request")
	decoder := json.NewDecoder(r.Body)
	//userId := getLoggedInUser(r)
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugf("request by user %s \n", userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean request.ClusterAccountsBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = impl.clusterAccountsService.Save(&bean, userId)
	if err != nil {
		impl.logger.Errorw("error in saving cluster account details", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusCreated)
}

func (impl ClusterAccountsRestHandlerImpl) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterName := vars["clusterName"]
	bean, err := impl.clusterAccountsService.FindOne(clusterName)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, bean, http.StatusOK)
}

func (impl ClusterAccountsRestHandlerImpl) GetByEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	environment := vars["environment"]
	bean, err := impl.clusterAccountsService.FindOneByEnvironment(environment)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, bean, http.StatusOK)
}

func (impl ClusterAccountsRestHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	//userId := getLoggedInUser(r)
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugf("request by user %s \n", userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean request.ClusterAccountsBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = impl.clusterAccountsService.Update(&bean, userId)
	if err != nil {
		impl.logger.Errorw("error in updating cluster account details", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusCreated)
}

func (impl ClusterAccountsRestHandlerImpl) FindById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	bean, err := impl.clusterAccountsService.FindById(id)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, bean, http.StatusOK)
}

func (impl ClusterAccountsRestHandlerImpl) FindAll(w http.ResponseWriter, r *http.Request) {
	beans, err := impl.clusterAccountsService.FindAll()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, beans, http.StatusOK)
}
