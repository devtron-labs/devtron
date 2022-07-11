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

package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"net/http"
	"strconv"
	"strings"

	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const CLUSTER_DELETE_SUCCESS_RESP = "Cluster deleted successfully."

type ClusterRestHandler interface {
	Save(w http.ResponseWriter, r *http.Request)
	FindOne(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)

	FindById(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)

	FindAllForAutoComplete(w http.ResponseWriter, r *http.Request)
	DeleteCluster(w http.ResponseWriter, r *http.Request)
}

type ClusterRestHandlerImpl struct {
	clusterService  cluster.ClusterService
	logger          *zap.SugaredLogger
	userService     user.UserService
	validator       *validator.Validate
	enforcer        casbin.Enforcer
	deleteService   delete2.DeleteService
	argoUserService argo.ArgoUserService
}

func NewClusterRestHandlerImpl(clusterService cluster.ClusterService,
	logger *zap.SugaredLogger,
	userService user.UserService,
	validator *validator.Validate,
	enforcer casbin.Enforcer,
	deleteService delete2.DeleteService,
	argoUserService argo.ArgoUserService) *ClusterRestHandlerImpl {
	return &ClusterRestHandlerImpl{
		clusterService:  clusterService,
		logger:          logger,
		userService:     userService,
		validator:       validator,
		enforcer:        enforcer,
		deleteService:   deleteService,
		argoUserService: argoUserService,
	}
}

func (impl ClusterRestHandlerImpl) Save(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	bean := new(cluster.ClusterBean)
	err = decoder.Decode(bean)
	if err != nil {
		impl.logger.Errorw("request err, Save", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Errorw("request payload, Save", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validation err, Save", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
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
	if util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_HYPERION {
		ctx = context.WithValue(ctx, "token", token)
	} else {
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.logger.Errorw("error in getting acd token", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		ctx = context.WithValue(ctx, "token", acdToken)
	}
	bean, err = impl.clusterService.Save(ctx, bean, userId)
	if err != nil {
		impl.logger.Errorw("service err, Save", "err", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	/*	isTriggered, err := impl.installedAppService.DeployDefaultChartOnCluster(bean, userId)
		if err != nil {
			impl.logger.Errorw("service err, Save, on DeployDefaultChartOnCluster", "err", err, "payload", bean)
		}
		if isTriggered {
			bean.AgentInstallationStage = 1
		} else {
			bean.AgentInstallationStage = 0
		}*/
	common.WriteJsonResp(w, err, bean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindOne(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	cName := vars["cluster_name"]
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionGet, strings.ToLower(cName)); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	envBean, err := impl.clusterService.FindOne(cName)
	if err != nil {
		impl.logger.Errorw("service err, FindOne", "error", err, "cluster name", cName)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, envBean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindAll(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	clusterList, err := impl.clusterService.FindAll()
	if err != nil {
		impl.logger.Errorw("service err, FindAll", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	var result []*cluster.ClusterBean
	for _, item := range clusterList {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionGet, strings.ToLower(item.ClusterName)); ok {
			result = append(result, item)
		}
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	i, err := strconv.Atoi(id)
	if err != nil {
		impl.logger.Errorw("request err, FindById", "error", err, "clusterId", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean, err := impl.clusterService.FindById(i)
	if err != nil {
		impl.logger.Errorw("service err, FindById", "err", err, "clusterId", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionGet, strings.ToLower(bean.ClusterName)); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, err, bean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) Update(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		impl.logger.Errorw("service err, Update", "error", err, "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean cluster.ClusterBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, Update", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, Update", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validate err, Update", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionUpdate, strings.ToLower(bean.ClusterName)); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer ends
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
	if util2.GetDevtronVersion().ServerMode == util2.SERVER_MODE_HYPERION {
		ctx = context.WithValue(ctx, "token", token)
	} else {
		acdToken, err := impl.argoUserService.GetLatestDevtronArgoCdUserToken()
		if err != nil {
			impl.logger.Errorw("error in getting acd token", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		ctx = context.WithValue(ctx, "token", acdToken)
	}
	_, err = impl.clusterService.Update(ctx, &bean, userId)
	if err != nil {
		impl.logger.Errorw("service err, Update", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, bean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindAllForAutoComplete(w http.ResponseWriter, r *http.Request) {
	clusterList, err := impl.clusterService.FindAllForAutoComplete()
	if err != nil {
		impl.logger.Errorw("service err, FindAllForAutoComplete", "error", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
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
			if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionGet, item.ClusterName); ok {
				result = append(result, item)
			}
		} else {
			result = append(result, item)
		}

	}
	//RBAC enforcer Ends

	if len(result) == 0 {
		result = make([]cluster.ClusterBean, 0)
	}
	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		impl.logger.Errorw("service err, Delete", "error", err, "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean cluster.ClusterBean
	err = decoder.Decode(&bean)
	if err != nil {
		impl.logger.Errorw("request err, Delete", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Debugw("request payload, Delete", "payload", bean)
	err = impl.validator.Struct(bean)
	if err != nil {
		impl.logger.Errorw("validate err, Delete", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	err = impl.deleteService.DeleteCluster(&bean, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting cluster", "err", err, "id", bean.Id, "name", bean.ClusterName)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, CLUSTER_DELETE_SUCCESS_RESP, http.StatusOK)
}
