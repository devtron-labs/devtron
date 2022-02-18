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
	"github.com/devtron-labs/devtron/client/argocdServer/application"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"go.uber.org/zap"
	"net/http"
)

type CicdRestHandler interface {
	SaveCICD(w http.ResponseWriter, r *http.Request)
	UpdateCICD(w http.ResponseWriter, r *http.Request)
	FindByCICD(w http.ResponseWriter, r *http.Request)
	DeleteCICD(w http.ResponseWriter, r *http.Request)
}

type CicdRestHandlerImpl struct {
	application application.ServiceClient
	logger      *zap.SugaredLogger
	repository  repository.CicdRepository
}

type cICD struct {
	Id   int    `json:"Id"`
	Name string `json:"Name"`
}

func NewCicdRestHandlerImpl(application application.ServiceClient, logger *zap.SugaredLogger, repository repository.CicdRepository) *CicdRestHandlerImpl {
	cicdRestHandler := &CicdRestHandlerImpl{
		application: application,
		logger:      logger,
		repository:  repository,
	}
	return cicdRestHandler
}

func (handler CicdRestHandlerImpl) SaveCICD(w http.ResponseWriter, r *http.Request) {
	//for checking
	decoder := json.NewDecoder(r.Body)
	println(decoder)
	var bean cICD
	err := decoder.Decode(&bean)
	if err != nil {
		println(err)
	}
	test := &repository.CICD{
		Id:   bean.Id,
		Name: bean.Name,
	}
	err = handler.repository.Save(test)
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, err, "", http.StatusOK)
}

func (handler CicdRestHandlerImpl) UpdateCICD(w http.ResponseWriter, r *http.Request) {
	//for checking
	decoder := json.NewDecoder(r.Body)
	println(decoder)
	var bean cICD
	err := decoder.Decode(&bean)
	if err != nil {
		println(err)
	}

	test := &repository.CICD{
		Id:   bean.Id,
		Name: bean.Name,
	}
	err = handler.repository.Update(test)
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, err, "", http.StatusOK)
}

func (handler CicdRestHandlerImpl) FindByCICD(w http.ResponseWriter, r *http.Request) {
	//for checking
	decoder := json.NewDecoder(r.Body)
	println(decoder)
	var bean cICD
	err := decoder.Decode(&bean)
	if err != nil {
		println(err)
	}

	values, err := handler.repository.FindByAppId(bean.Id)
	println(values)
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, err, "", http.StatusOK)
}

func (handler CicdRestHandlerImpl) DeleteCICD(w http.ResponseWriter, r *http.Request) {
	//for checking
	decoder := json.NewDecoder(r.Body)
	println(decoder)
	var bean cICD
	err := decoder.Decode(&bean)
	if err != nil {
		println(err)
	}

	test := &repository.CICD{
		Id:   bean.Id,
		Name: bean.Name,
	}
	err = handler.repository.Delete(test)
	if err != nil {
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, err, "", http.StatusOK)
}
