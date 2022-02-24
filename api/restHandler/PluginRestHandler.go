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
	Id           int            `json:"pluginId"`
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	Body         string         `json:"body"`
	PluginInputs []pluginFields `json:"pluginInputs"`
	PluginSteps  []pluginSteps  `json:"pluginSteps"`
	PluginTags   []string       `json:"pluginTags"`
}

type pluginFields struct {
	Id              int    `json:"pluginId"`
	Name            string `json:"keyName"`
	Value           string `json:"defaultValue"`
	Description     string `json:"pluginKeyDescription"`
	PluginFieldType string `json:"pluginFieldType"`
}

type pluginSteps struct {
	StepId               string         `json:"stepId"`
	StepName             string         `json:"stepName"`
	StepTemplateLanguage string         `json:"stepTemplateLanguage"`
	StepTemplate         string         `json:"stepTemplate"`
	PluginInputs         []pluginFields `json:"pluginInputs"`
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

	//StepTemplateLanguage: bean.StepTemplateLanguage,
	//	StepTemplate:         bean.StepTemplate,

	data := &repository.Plugin{
		PluginId:          bean.Id,
		PluginName:        bean.Name,
		PluginDescription: bean.Description,
		PluginBody:        bean.Body,
	}

	pluginResp, err := handler.repository.SavePlugin(data)
	if err != nil {
		common.WriteJsonResp(w, err, "Plugin couldn't be saved", http.StatusInternalServerError)
		return
	}

	bean.Id = pluginResp.PluginId
	inputData, err := SaveFieldsList(bean.Id, bean.PluginInputs)

	if err != nil {
		common.WriteJsonResp(w, err, "Plugin fields couldn't be saved", http.StatusInternalServerError)
	}

	var pluginstepslist []*repository.PluginSteps
	for _, eachInput := range bean.PluginSteps {
		input := &repository.PluginSteps{
			StepName:              eachInput.StepName,
			StepsTemplateLanguage: eachInput.StepTemplateLanguage,
			StepsTemplate:         eachInput.StepTemplate,
		}
		step, err := handler.repository.SaveSteps(input)
		pluginStepsFields, err := SaveFieldsList(step.StepId, eachInput.PluginInputs)
		stepSeq := &repository.PluginStepsSequence{
			StepsId:  step.StepId,
			PluginId: pluginResp.PluginId,
		}
		err = handler.repository.SaveStepsSequence(stepSeq)
		if err != nil {
			common.WriteJsonResp(w, err, "Steps fields couldn't be saved", http.StatusInternalServerError)
		}

		pluginstepslist = append(pluginstepslist, input)
		inputData = append(inputData, pluginStepsFields...)
	}

	err = handler.SaveFLags(bean.Id, bean.PluginTags)
	if err != nil {
		common.WriteJsonResp(w, err, "Plugin Tags Map couldn't be saved", http.StatusInternalServerError)
	}

	err = handler.repository.SaveFields(inputData)
	if err != nil {
		common.WriteJsonResp(w, err, "Plugin fields couldn't be saved", http.StatusInternalServerError)
	}

	common.WriteJsonResp(w, err, "Stored Successfully", http.StatusOK)
}

func SaveFieldsList(id int, pluginfields []pluginFields) ([]*repository.PluginFields, error) {
	var inputData []*repository.PluginFields
	for _, eachInput := range pluginfields {
		input := &repository.PluginFields{
			PluginId:          id,
			FieldName:         eachInput.Name,
			FieldDefaultValue: eachInput.Value,
			FieldDescription:  eachInput.Description,
			FieldType:         eachInput.PluginFieldType,
		}
		inputData = append(inputData, input)
	}
	return inputData, nil
}

func (handler PluginRestHandlerImpl) SaveFLags(id int, plugintags []string) error {
	for _, eachInput := range plugintags {
		plugintag, err := handler.repository.FindTagId(eachInput)
		if err != nil {
			tag := &repository.Tags{
				TagName: eachInput,
			}
			plugintag, err = handler.repository.SaveTag(tag)
			if err != nil {
				return err
			}
		}
		tagsMap := &repository.PluginTagsMap{
			TagId:    plugintag.TagId,
			PluginId: id,
		}
		err = handler.repository.SavePluginTagsMap(tagsMap)
		if err != nil {
			return err
		}
	}
	return nil
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
		PluginId:          bean.Id,
		PluginName:        bean.Name,
		PluginDescription: bean.Description,
		PluginBody:        bean.Body,
	}

	var inputData []*repository.PluginFields
	for _, eachInput := range bean.PluginInputs {
		input := &repository.PluginFields{
			PluginId:          bean.Id,
			FieldName:         eachInput.Name,
			FieldDefaultValue: eachInput.Value,
			FieldDescription:  eachInput.Description,
			FieldType:         eachInput.PluginFieldType,
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
	id, _ := strconv.Atoi(vars["Id"])
	values, _, _ := handler.repository.FindByAppId(id)
	common.WriteJsonResp(w, nil, values, http.StatusOK)
}

//func (handler PluginRestHandlerImpl) FindByPlugin(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	/* #nosec */
//	id, err := strconv.Atoi(vars["Id"])
//	if err != nil {
//		handler.logger.Errorw("decode err", "err", err)
//		common.WriteJsonResp(w, err, "Plugin Id couldn't be parsed from input", http.StatusBadRequest)
//	}
//	values, params, err := handler.repository.FindByAppId(id)
//	if err != nil {
//		common.WriteJsonResp(w, err, "Plugin not found", http.StatusInternalServerError)
//	}
//
//	var inputValues []pluginFields
//	for _, eachInput := range params {
//		input := pluginFields{
//			Id:          eachInput.Id,
//			Name:        eachInput.Name,
//			Value:       eachInput.DefaultValue,
//			Description: eachInput.Description,
//		}
//		inputValues = append(inputValues, input)
//	}
//
//	pluginData := &plugin{
//		Id:                   values.Id,
//		Name:                 values.Name,
//		Description:          values.Description,
//		Body:                 values.Body,
//		StepTemplateLanguage: values.StepTemplateLanguage,
//		StepTemplate:         values.StepTemplate,
//		PluginInputs:         inputValues,
//	}
//	common.WriteJsonResp(w, err, pluginData, http.StatusOK)
//}

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
