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
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type GitHostRestHandler interface {
	GetGitHosts(w http.ResponseWriter, r *http.Request)
	GetGitHostById(w http.ResponseWriter, r *http.Request)
}

type GitHostRestHandlerImpl struct {
	logger               *zap.SugaredLogger
	gitHostConfig    	 pipeline.GitHostConfig
}

func NewGitHostRestHandlerImpl(logger *zap.SugaredLogger,
	gitHostConfig pipeline.GitHostConfig) *GitHostRestHandlerImpl {
	return &GitHostRestHandlerImpl{
		logger:               logger,
		gitHostConfig:    gitHostConfig,
	}
}

func (impl GitHostRestHandlerImpl) GetGitHosts(w http.ResponseWriter, r *http.Request) {
	res, err := impl.gitHostConfig.GetAll()
	if err != nil {
		impl.logger.Errorw("service err, GetGitHosts", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl GitHostRestHandlerImpl) GetGitHostById(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err :=  strconv.Atoi(params["id"])

	if err != nil {
		impl.logger.Errorw("service err in parsing Id , GetGitHostById", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	res, err := impl.gitHostConfig.GetById(id)

	if err != nil {
		impl.logger.Errorw("service err, GetGitHostById", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	writeJsonResp(w, err, res, http.StatusOK)
}
