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
	"context"
	cluster2 "github.com/devtron-labs/devtron/client/argocdServer/cluster"
	"github.com/devtron-labs/devtron/internal/constants"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/appstore"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"encoding/json"
	"errors"
	"fmt"
	cluster3 "github.com/argoproj/argo-cd/pkg/apiclient/cluster"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

type ClusterRestHandler interface {
	Save(w http.ResponseWriter, r *http.Request)
	FindOne(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)

	FindById(w http.ResponseWriter, r *http.Request)
	FindByEnvId(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)

	ClusterListFromACD(w http.ResponseWriter, r *http.Request)
	DeleteClusterFromACD(w http.ResponseWriter, r *http.Request)

	FindAllForAutoComplete(w http.ResponseWriter, r *http.Request)
	DefaultComponentInstallation(w http.ResponseWriter, r *http.Request)
}

type ClusterRestHandlerImpl struct {
	clusterService         cluster.ClusterService
	clusterServiceCD       cluster2.ServiceClient
	logger                 *zap.SugaredLogger
	envService             cluster.EnvironmentService
	clusterAccountsService cluster.ClusterAccountsService
	userService            user.UserService
	validator              *validator.Validate
	enforcer               rbac.Enforcer
	installedAppService    appstore.InstalledAppService
}

func NewClusterRestHandlerImpl(clusterService cluster.ClusterService,
	logger *zap.SugaredLogger,
	clusterServiceCD cluster2.ServiceClient,
	envService cluster.EnvironmentService,
	clusterAccountsService cluster.ClusterAccountsService,
	userService user.UserService,
	validator *validator.Validate,
	enforcer rbac.Enforcer, installedAppService appstore.InstalledAppService) *ClusterRestHandlerImpl {
	return &ClusterRestHandlerImpl{
		clusterService:         clusterService,
		logger:                 logger,
		clusterServiceCD:       clusterServiceCD,
		envService:             envService,
		clusterAccountsService: clusterAccountsService,
		userService:            userService,
		validator:              validator,
		enforcer:               enforcer,
		installedAppService:    installedAppService,
	}
}

func (impl ClusterRestHandlerImpl) Save(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	bean := new(cluster.ClusterBean)
	err = decoder.Decode(bean)
	if err != nil {
		impl.logger.Errorw("request err, Save", "error", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Errorw("request payload, Save", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, Save", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionCreate, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	bean, err = impl.clusterService.Save(bean, userId)
	if err != nil {
		impl.logger.Errorw("service err, Save", "err", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	//create it into argo cd as well
	configMap := bean.Config
	serverUrl := bean.ServerUrl
	bearerToken := ""
	if configMap["bearer_token"] != "" {
		bearerToken = configMap["bearer_token"]
	}
	tlsConfig := v1alpha1.TLSClientConfig{
		Insecure: true,
	}
	cdClusterConfig := v1alpha1.ClusterConfig{
		BearerToken:     bearerToken,
		TLSClientConfig: tlsConfig,
	}

	cl := &v1alpha1.Cluster{
		Name:   bean.ClusterName,
		Server: serverUrl,
		Config: cdClusterConfig,
	}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(ctx, "token", token)
	_, err = impl.clusterServiceCD.Create(ctx, &cluster3.ClusterCreateRequest{Upsert: true, Cluster: cl})
	if err != nil {
		impl.logger.Errorw("service err, Save", "err", err, "payload", cl)
		err1 := impl.clusterService.Delete(bean, userId)
		if err1 != nil {
			impl.logger.Errorw("service err, Save, delete on rollback", "err", err, "payload", bean)
			err = &util.ApiError{
				Code:            constants.ClusterDBRollbackFailed,
				InternalMessage: err.Error(),
				UserMessage:     "failed to rollback cluster from db as it has failed in registering on ACD",
			}
			writeJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		err = &util.ApiError{
			Code:            constants.ClusterCreateACDFailed,
			InternalMessage: err.Error(),
			UserMessage:     "failed to register on ACD, rollback completed from db",
		}
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	isTriggered, err := impl.installedAppService.DeployDefaultChartOnCluster(bean, userId)
	if err != nil {
		impl.logger.Errorw("service err, Save, on DeployDefaultChartOnCluster", "err", err, "payload", bean)
	}
	if isTriggered {
		bean.AgentInstallationStage = 1
	} else {
		bean.AgentInstallationStage = 0
	}
	writeJsonResp(w, err, bean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cName := vars["cluster_name"]
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionGet, strings.ToLower(cName)); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	envBean, err := impl.clusterService.FindOne(cName)
	if err != nil {
		impl.logger.Errorw("service err, FindOne", "error", err, "cluster name", cName)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, envBean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindAll(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	clusterList, err := impl.clusterService.FindAll()
	if err != nil {
		impl.logger.Errorw("service err, FindAll", "err", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	var result []*cluster.ClusterBean
	for _, item := range clusterList {
		if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionGet, strings.ToLower(item.ClusterName)); ok {
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends

	writeJsonResp(w, err, result, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	i, err := strconv.Atoi(id)
	if err != nil {
		impl.logger.Errorw("request err, FindById", "error", err, "clusterId", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean, err := impl.clusterService.FindById(i)
	if err != nil {
		impl.logger.Errorw("service err, FindById", "err", err, "clusterId", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionGet, strings.ToLower(bean.ClusterName)); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	writeJsonResp(w, err, bean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindByEnvId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	idi, err := strconv.Atoi(id)
	if err != nil {
		impl.logger.Errorw("request err, FindByEnvId", "error", err, "clusterId", id)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envBean, err := impl.envService.FindClusterByEnvId(idi)
	if err != nil {
		impl.logger.Errorw("service err, FindByEnvId", "error", err, "clusterId", id)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionGet, strings.ToLower(envBean.ClusterName)); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer ends

	writeJsonResp(w, err, envBean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		impl.logger.Errorw("service err, Update", "error", err, "userId", userId)
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean cluster.ClusterBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, Update", "error", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Errorw("request payload, Update", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validate err, Update", "error", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionUpdate, strings.ToLower(bean.ClusterName)); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer ends

	_, err = impl.clusterService.Update(&bean, userId)
	if err != nil {
		impl.logger.Errorw("service err, Update", "error", err, "payload", bean)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	configMap := bean.Config
	serverUrl := bean.ServerUrl
	bearerToken := ""
	if configMap["bearer_token"] != "" {
		bearerToken = configMap["bearer_token"]
	}

	tlsConfig := v1alpha1.TLSClientConfig{
		Insecure: true,
	}
	cdClusterConfig := v1alpha1.ClusterConfig{
		BearerToken:     bearerToken,
		TLSClientConfig: tlsConfig,
	}

	cl := &v1alpha1.Cluster{
		Name:   bean.ClusterName,
		Server: serverUrl,
		Config: cdClusterConfig,
	}
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(r.Context(), "token", token)
	_, err = impl.clusterServiceCD.Update(ctx, &cluster3.ClusterUpdateRequest{Cluster: cl})

	if err != nil {
		impl.logger.Errorw("service err, Update", "error", err, "payload", cl)
		userMsg := "failed to update on cluster via ACD"
		if strings.Contains(err.Error(), "https://kubernetes.default.svc") {
			userMsg = fmt.Sprintf("%s, %s", err.Error(), ", sucessfully updated in ACD")
		}
		err = &util.ApiError{
			Code:            constants.ClusterUpdateACDFailed,
			InternalMessage: err.Error(),
			UserMessage:     userMsg,
		}
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, bean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) ClusterListFromACD(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(r.Context(), "token", token)
	cList, err := impl.clusterServiceCD.List(ctx, &cluster3.ClusterQuery{})
	if err != nil {
		impl.logger.Errorw("service err, ClusterListFromACD", "error", err)
		writeJsonResp(w, err, "failed to fetch list from ACD:", http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying
	if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionGet, "*"); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	// RBAC enforcer ends

	resJson, err := json.Marshal(cList)
	_, err = w.Write(resJson)
	if err != nil {
		impl.logger.Errorw("marshal err, ClusterListFromACD", "error", err, "cList", cList)
	}

}

func (impl ClusterRestHandlerImpl) DeleteClusterFromACD(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	serverUrl := vars["server_url"]
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	impl.logger.Errorw("request payload, DeleteClusterFromACD", "serverUrl", serverUrl)
	ctx = context.WithValue(r.Context(), "token", token)
	res, err := impl.clusterServiceCD.Delete(ctx, &cluster3.ClusterQuery{Server: serverUrl})
	if err != nil {
		impl.logger.Errorw("service err, DeleteClusterFromACD", "error", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionDelete, "*"); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer ends

	resJson, err := json.Marshal(res)
	_, err = w.Write(resJson)
	if err != nil {
		impl.logger.Errorw("service err, DeleteClusterFromACD", "error", err, "res", res)
	}

}

func (impl ClusterRestHandlerImpl) GetClusterFromACD(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	serverUrl := vars["server_url"]
	ctx, cancel := context.WithCancel(r.Context())
	if cn, ok := w.(http.CloseNotifier); ok {
		go func(done <-chan struct{}, closed <-chan bool) {
			select {
			case <-done:
			case <-closed:
				cancel()
			}
		}(ctx.Done(), cn.CloseNotify())
	}
	ctx = context.WithValue(r.Context(), "token", token)
	res, err := impl.clusterServiceCD.Get(ctx, &cluster3.ClusterQuery{Server: serverUrl})
	if err != nil {
		impl.logger.Errorw("service err, GetClusterFromACD", "error", err, "serverUrl", serverUrl)
		writeJsonResp(w, err, "failed to fetch from ACD:", http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionGet, "*"); !ok {
		writeJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	// RBAC enforcer ends

	resJson, err := json.Marshal(res)
	_, err = w.Write(resJson)
	if err != nil {
		impl.logger.Errorw("service err, GetClusterFromACD", "error", err, "res", res)
	}
}

func (impl ClusterRestHandlerImpl) FindAllForAutoComplete(w http.ResponseWriter, r *http.Request) {
	clusterList, err := impl.clusterService.FindAllForAutoComplete()
	if err != nil {
		impl.logger.Errorw("service err, FindAllForAutoComplete", "error", err)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	var result []cluster.ClusterBean
	v := r.URL.Query()
	authEnabled := true
	auth := v.Get("auth")
	if len(auth) > 0 {
		authEnabled, err = strconv.ParseBool(auth)
		if err != nil {
			authEnabled = true
			err = nil
			//ignore error, apply rbac by default
		}
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	for _, item := range clusterList {
		if authEnabled == true {
			if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionGet, item.ClusterName); ok {
				result = append(result, item)
			}
		} else {
			result = append(result, item)
		}

	}
	//RBAC enforcer Ends

	if result == nil || len(result) == 0 {
		result = make([]cluster.ClusterBean, 0)
	}
	writeJsonResp(w, err, result, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) DefaultComponentInstallation(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		impl.logger.Errorw("service err, DefaultComponentInstallation", "error", err, "userId", userId)
		writeJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	clusterId, err := strconv.Atoi(vars["clusterId"])
	if err != nil {
		impl.logger.Errorw("request err, DefaultComponentInstallation", "error", err, "clusterId", clusterId)
		writeJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Errorw("request payload, DefaultComponentInstallation", "clusterId", clusterId)
	cluster, err := impl.clusterService.FindById(clusterId)
	if err != nil {
		impl.logger.Errorw("service err, DefaultComponentInstallation", "error", err, "clusterId", clusterId)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	if ok := impl.enforcer.Enforce(token, rbac.ResourceCluster, rbac.ActionUpdate, strings.ToLower(cluster.ClusterName)); !ok {
		writeJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer ends
	isTriggered, err := impl.installedAppService.DeployDefaultChartOnCluster(cluster, userId)
	if err != nil {
		impl.logger.Errorw("service err, DefaultComponentInstallation", "error", err, "cluster", cluster)
		writeJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	writeJsonResp(w, err, isTriggered, http.StatusOK)
}
