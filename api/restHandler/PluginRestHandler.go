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
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type PluginRestHandler interface {
	SavePlugin(w http.ResponseWriter, r *http.Request)
	UpdatePlugin(w http.ResponseWriter, r *http.Request)
	FindByPlugin(w http.ResponseWriter, r *http.Request)
	DeletePlugin(w http.ResponseWriter, r *http.Request)
}

type PluginRestHandlerImpl struct {
	logger     *zap.SugaredLogger
	repository repository.PluginRepository
}

type plugin struct {
	Id                   int            `json:"pluginId"`
	Name                 string         `json:"name"`
	Description          string         `json:"description"`
	Body                 string         `json:"body"`
	StepTemplateLanguage string         `json:"stepTemplateLanguage"`
	StepTemplate         string         `json:"stepTemplate"`
	PluginInputs         []pluginInputs `json:"pluginInputs"`
}

type pluginInputs struct {
	Id          int    `json:"pluginId"`
	Name        string `json:"keyName"`
	Value       string `json:"defaultValue"`
	Description string `json:"pluginKeyDescription"`
}

func NewPluginRestHandlerImpl(logger *zap.SugaredLogger, repository repository.PluginRepository) *PluginRestHandlerImpl {
	pluginRestHandler := &PluginRestHandlerImpl{
		logger:     logger,
		repository: repository,
	}
	return pluginRestHandler
}

func (handler PluginRestHandlerImpl) SavePlugin(w http.ResponseWriter, r *http.Request) {
	//for checking
	decoder := json.NewDecoder(r.Body)
	var bean plugin
	err := decoder.Decode(&bean)
	if err != nil {
		common.WriteJsonResp(w, err, "Plugin Id couldn't be parsed from input", http.StatusBadRequest)
	}

	data := &repository.Plugin{
		Id:                   bean.Id,
		Name:                 bean.Name,
		Description:          bean.Description,
		Body:                 bean.Body,
		StepTemplateLanguage: bean.StepTemplateLanguage,
		StepTemplate:         bean.StepTemplate,
	}
	var inputData []*repository.PluginInputs
	for _, eachInput := range bean.PluginInputs {
		input := &repository.PluginInputs{
			Id:           bean.Id,
			Name:         eachInput.Name,
			DefaultValue: eachInput.Value,
			Description:  eachInput.Description,
		}
		inputData = append(inputData, input)
	}

	err = handler.repository.Save(data, inputData)
	if err != nil {
		common.WriteJsonResp(w, err, "Plugin couldn't be saved", http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, err, "Stored Successfully", http.StatusOK)
}

func (handler PluginRestHandlerImpl) UpdatePlugin(w http.ResponseWriter, r *http.Request) {
	//for checking
	decoder := json.NewDecoder(r.Body)
	println(decoder)
	var bean plugin
	err := decoder.Decode(&bean)
	if err != nil {
		handler.logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, "Plugin Id couldn't be parsed from input", http.StatusBadRequest)
	}

	test := &repository.Plugin{
		Id:                   bean.Id,
		Name:                 bean.Name,
		Description:          bean.Description,
		Body:                 bean.Body,
		StepTemplateLanguage: bean.StepTemplateLanguage,
		StepTemplate:         bean.StepTemplate,
	}

	var inputData []*repository.PluginInputs
	for _, eachInput := range bean.PluginInputs {
		input := &repository.PluginInputs{
			Id:           bean.Id,
			Name:         eachInput.Name,
			DefaultValue: eachInput.Value,
			Description:  eachInput.Description,
		}
		inputData = append(inputData, input)
	}

	err = handler.repository.Update(test, inputData)
	if err != nil {
		common.WriteJsonResp(w, err, "Plugin couldn't be updated", http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, err, "Update Successful", http.StatusOK)
}

func (handler PluginRestHandlerImpl) FindByPlugin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	/* #nosec */
	id, err := strconv.Atoi(vars["Id"])
	if err != nil {
		handler.logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, "Plugin Id couldn't be parsed from input", http.StatusBadRequest)
	}
	values, params, err := handler.repository.FindByAppId(id)
	if err != nil {
		common.WriteJsonResp(w, err, "Plugin not found", http.StatusInternalServerError)
	}

	var inputValues []pluginInputs
	for _, eachInput := range params {
		input := pluginInputs{
			Id:          eachInput.Id,
			Name:        eachInput.Name,
			Value:       eachInput.DefaultValue,
			Description: eachInput.Description,
		}
		inputValues = append(inputValues, input)
	}

	pluginData := &plugin{
		Id:                   values.Id,
		Name:                 values.Name,
		Description:          values.Description,
		Body:                 values.Body,
		StepTemplateLanguage: values.StepTemplateLanguage,
		StepTemplate:         values.StepTemplate,
		PluginInputs:         inputValues,
	}
	common.WriteJsonResp(w, err, pluginData, http.StatusOK)
}

func (handler PluginRestHandlerImpl) DeletePlugin(w http.ResponseWriter, r *http.Request) {
	//for checking
	vars := mux.Vars(r)
	/* #nosec */
	id, err := strconv.Atoi(vars["Id"])
	if err != nil {
		handler.logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, "Plugin Id couldn't be parsed from input", http.StatusBadRequest)
	}

	err = handler.repository.Delete(id)
	if err != nil {
		common.WriteJsonResp(w, err, "Plugin couldn't be deleted", http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, err, "Delete Successful", http.StatusOK)
}
