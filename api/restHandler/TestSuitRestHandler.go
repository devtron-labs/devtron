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
	RedirectTestSuites(w http.ResponseWriter, r *http.Request)
	RedirectDetailedTestSuites(w http.ResponseWriter, r *http.Request)
	RedirectSuitByID(w http.ResponseWriter, r *http.Request)
	RedirectTestCases(w http.ResponseWriter, r *http.Request)
	RedirectTestCaseByID(w http.ResponseWriter, r *http.Request)
	RedirectTriggerForPipeline(w http.ResponseWriter, r *http.Request)
	RedirectTriggerForBuild(w http.ResponseWriter, r *http.Request)
	RedirectFilterForPipeline(w http.ResponseWriter, r *http.Request)
	RedirectFilterForBuild(w http.ResponseWriter, r *http.Request)
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

func (impl TestSuitRestHandlerImpl) RedirectTestSuites(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	link := fmt.Sprintf("%s/%s", impl.config.TestSuitURL, "testsuite")
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) RedirectDetailedTestSuites(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	link := fmt.Sprintf("%s/%s/all", impl.config.TestSuitURL, "testsuites")
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) RedirectSuitByID(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	token := r.Header.Get("token")
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		impl.logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	link := fmt.Sprintf("%s/%s/%d", impl.config.TestSuitURL, "testsuite", pipelineId)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) RedirectTestCases(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	link := fmt.Sprintf("%s/%s", impl.config.TestSuitURL, "testcase")
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) RedirectTestCaseByID(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		impl.logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	link := fmt.Sprintf("%s/%s/%d", impl.config.TestSuitURL, "testcase", pipelineId)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) RedirectTriggerForPipeline(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		impl.logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	link := fmt.Sprintf("%s/%s/%d?%s", impl.config.TestSuitURL, "triggers", pipelineId, r.URL.RawQuery)
	impl.logger.Debugw("redirect to link", "link", link)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) RedirectTriggerForBuild(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	//v := r.URL.Query()
	//startDate := v.Get("auth")
	//endDate := v.Get("auth")
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		impl.logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	triggerId, err := strconv.Atoi(vars["triggerId"])
	if err != nil {
		impl.logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	link := fmt.Sprintf("%s/%s/%d/%d?%s", impl.config.TestSuitURL, "triggers", pipelineId, triggerId, r.URL.RawQuery)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) RedirectFilterForPipeline(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	token := r.Header.Get("token")
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		impl.logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	link := fmt.Sprintf("%s/%s/%d", impl.config.TestSuitURL, "filters", pipelineId)
	impl.logger.Debugw("redirect to link", "link", link)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	writeJsonResp(w, err, res, http.StatusOK)
}

func (impl TestSuitRestHandlerImpl) RedirectFilterForBuild(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	impl.logger.Debugw("request for user", "userId", userId)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	token := r.Header.Get("token")
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := impl.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := impl.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, resourceName); !ok {
		writeJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		impl.logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	triggerId, err := strconv.Atoi(vars["triggerId"])
	if err != nil {
		impl.logger.Error(err)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	link := fmt.Sprintf("%s/%s/%d/%d", impl.config.TestSuitURL, "filters", pipelineId, triggerId)
	res, err := impl.HttpGet(link)
	if err != nil {
		impl.logger.Error(err)
	}
	writeJsonResp(w, err, res, http.StatusOK)
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
