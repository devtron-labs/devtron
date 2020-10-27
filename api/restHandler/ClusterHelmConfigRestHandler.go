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
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"encoding/json"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type ClusterHelmConfigRestHandler interface {
	Save(w http.ResponseWriter, r *http.Request)
	GetByEnvironment(w http.ResponseWriter, r *http.Request)
}

type ClusterHelmConfigRestHandlerImpl struct {
	clusterHelmConfigService cluster.ClusterHelmConfigService
	logger                   *zap.SugaredLogger
	userAuthService          user.UserService
}

func NewClusterHelmConfigRestHandlerImpl(service cluster.ClusterHelmConfigService, logger *zap.SugaredLogger, userAuthService user.UserService) *ClusterHelmConfigRestHandlerImpl {
	return &ClusterHelmConfigRestHandlerImpl{
		clusterHelmConfigService: service,
		logger:                   logger,
		userAuthService:          userAuthService,
	}
}

func (impl ClusterHelmConfigRestHandlerImpl) Save(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean cluster.ClusterHelmConfigBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, Save", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Errorw("request payload, Save", "err", err, "payload", bean)
	err = impl.clusterHelmConfigService.Save(&bean, userId)
	if err != nil {
		impl.logger.Errorw("service err, Save", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, nil, http.StatusCreated)
}

func (impl ClusterHelmConfigRestHandlerImpl) GetByEnvironment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	environment := vars["environment"]
	bean, err := impl.clusterHelmConfigService.FindOneByEnvironment(environment)
	if err != nil {
		impl.logger.Errorw("service err, GetByEnvironment", "err", err, "environment", environment)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, bean, http.StatusOK)
}
