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
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/events"
	"github.com/devtron-labs/devtron/client/grafana"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"io/ioutil"
	"net/http"
	"strconv"
)

type TestSuitRestHandler interface {
	SuitesProxy(w http.ResponseWriter, r *http.Request)
	GetTestSuites(w http.ResponseWriter, r *http.Request)
	DetailedTestSuites(w http.ResponseWriter, r *http.Request)
	GetAllSuitByID(w http.ResponseWriter, r *http.Request)
	GetAllTestCases(w http.ResponseWriter, r *http.Request)
	GetTestCaseByID(w http.ResponseWriter, r *http.Request)
	RedirectTriggerForApp(w http.ResponseWriter, r *http.Request)
	RedirectTriggerForEnv(w http.ResponseWriter, r *http.Request)
}

type TestSuitRestHandlerImpl struct {
	logger       *zap.SugaredLogger
	userService  user.UserService
	validator    *validator.Validate
	enforcer     rbac.Enforcer
	enforcerUtil rbac.EnforcerUtil
	config       *client.EventClientConfig
	client       *http.Client
}

func NewTestSuitRestHandlerImpl(logger *zap.SugaredLogger, userService user.UserService,
	validator *validator.Validate, enforcer rbac.Enforcer, enforcerUtil rbac.EnforcerUtil,
	config *client.EventClientConfig, client *http.Client) *TestSuitRestHandlerImpl {
	return &TestSuitRestHandlerImpl{
		logger:       logger,
		userService:  userService,
		enforcer:     enforcer,
		enforcerUtil: enforcerUtil,
		config:       config,
		client:       client,
	}
}

type TestSuiteBean struct {
	Link       string `json:"link,omitempty"`
	PipelineId int    `json:"PipelineId"`
	TriggerId  int    `json:"triggerId"`
}

func (impl TestSuitRestHandlerImpl) SuitesProxy(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var bean TestSuiteBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("decode err", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	link := fmt.Sprintf("%s/%s", impl.config.TestSuitURL, bean.Link)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) GetTestSuites(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	link := fmt.Sprintf("%s/%s", impl.config.TestSuitURL, "testsuite")
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) DetailedTestSuites(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	link := fmt.Sprintf("%s/%s", impl.config.TestSuitURL, "testsuites/all")
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) GetAllSuitByID(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		impl.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	link := fmt.Sprintf("%s/%s/%d", impl.config.TestSuitURL, "testsuite", id)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) GetAllTestCases(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	link := fmt.Sprintf("%s/%s", impl.config.TestSuitURL, "testcase")
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) GetTestCaseByID(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		impl.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	link := fmt.Sprintf("%s/%s/%d", impl.config.TestSuitURL, "testcase", id)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) RedirectTriggerForApp(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		impl.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	link := fmt.Sprintf("%s/%s/%d", impl.config.TestSuitURL, "triggers", appId)
	impl.logger.Debugw("redirect to link", "link", link)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) RedirectTriggerForEnv(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		impl.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["triggerId"])
	if err != nil {
		impl.logger.Error(err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	link := fmt.Sprintf("%s/%s/%d/%d", impl.config.TestSuitURL, "triggers", appId, envId)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) HttpGet(url string) (map[string]interface{}, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		impl.logger.Errorw("error while fetching data", "err", err)
		return nil, err
	}
	resp, err := impl.client.Do(req)
	if err != nil {
		impl.logger.Errorw("err", err)
		return nil, err
	}
	status := grafana.StatusCode(resp.StatusCode)
	resBody, err := ioutil.ReadAll(resp.Body)
	var apiRes map[string]interface{}
	if status.IsSuccess() {
		if err != nil {
			impl.logger.Errorw("error in grafana communication ", "err", err)
			return nil, err
		}
		err = json.Unmarshal(resBody, &apiRes)
		if err != nil {
			impl.logger.Errorw("error in grafana resp unmarshalling ", "err", err)
			return nil, err
		}
	} else {
		impl.logger.Errorw("api err", "res", string(resBody))
		return nil, fmt.Errorf("res not success, code: %d ,response body: %s", status, string(resBody))
	}
	return apiRes, nil
}
