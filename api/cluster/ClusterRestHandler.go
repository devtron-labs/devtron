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
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/cluster"
	delete2 "github.com/devtron-labs/devtron/pkg/delete"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/argo"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
)

const CLUSTER_DELETE_SUCCESS_RESP = "Cluster deleted successfully."

type ClusterRestHandler interface {
	Save(w http.ResponseWriter, r *http.Request)
	SaveVirtualCluster(w http.ResponseWriter, r *http.Request)
	FindAll(w http.ResponseWriter, r *http.Request)
	FindById(w http.ResponseWriter, r *http.Request)
	FindNoteByClusterId(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	UpdateVirtualCluster(w http.ResponseWriter, r *http.Request)
	UpdateClusterNote(w http.ResponseWriter, r *http.Request)
	FindAllForAutoComplete(w http.ResponseWriter, r *http.Request)
	DeleteCluster(w http.ResponseWriter, r *http.Request)
	DeleteVirtualCluster(w http.ResponseWriter, r *http.Request)
	GetClusterNamespaces(w http.ResponseWriter, r *http.Request)
	GetAllClusterNamespaces(w http.ResponseWriter, r *http.Request)
	FindAllForClusterPermission(w http.ResponseWriter, r *http.Request)
}

type ClusterRestHandlerImpl struct {
	clusterService            cluster.ClusterService
	clusterNoteService        cluster.ClusterNoteService
	clusterDescriptionService cluster.ClusterDescriptionService
	logger                    *zap.SugaredLogger
	userService               user.UserService
	validator                 *validator.Validate
	enforcer                  casbin.Enforcer
	deleteService             delete2.DeleteService
	argoUserService           argo.ArgoUserService
	environmentService        cluster.EnvironmentService
}

func NewClusterRestHandlerImpl(clusterService cluster.ClusterService,
	clusterNoteService cluster.ClusterNoteService,
	clusterDescriptionService cluster.ClusterDescriptionService,
	logger *zap.SugaredLogger,
	userService user.UserService,
	validator *validator.Validate,
	enforcer casbin.Enforcer,
	deleteService delete2.DeleteService,
	argoUserService argo.ArgoUserService,
	environmentService cluster.EnvironmentService) *ClusterRestHandlerImpl {
	return &ClusterRestHandlerImpl{
		clusterService:            clusterService,
		clusterNoteService:        clusterNoteService,
		clusterDescriptionService: clusterDescriptionService,
		logger:                    logger,
		userService:               userService,
		validator:                 validator,
		enforcer:                  enforcer,
		deleteService:             deleteService,
		argoUserService:           argoUserService,
		environmentService:        environmentService,
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
	impl.logger.Infow("request payload, Save", "payload", bean)
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
	if util2.IsBaseStack() {
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

func (impl ClusterRestHandlerImpl) SaveVirtualCluster(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	bean := new(cluster.VirtualClusterBean)
	err = decoder.Decode(bean)
	if err != nil {
		impl.logger.Errorw("request err, Save", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, Save", "payload", bean)
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
	clusterBean, err := impl.clusterService.SaveVirtualCluster(bean, userId)
	if err != nil {
		impl.logger.Errorw("error in saving cluster", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, clusterBean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindAll(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	clusterList, err := impl.clusterService.FindAllWithoutConfig()
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
	bean, err := impl.clusterService.FindByIdWithoutConfig(i)
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

func (impl ClusterRestHandlerImpl) FindNoteByClusterId(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	i, err := strconv.Atoi(id)
	if err != nil {
		impl.logger.Errorw("request err, FindById", "error", err, "clusterId", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	bean, err := impl.clusterDescriptionService.FindByClusterIdWithClusterDetails(i)
	if err != nil {
		if err == pg.ErrNoRows {
			impl.logger.Errorw("cluster not found, FindById", "err", err, "clusterId", id)
			common.WriteJsonResp(w, errors.New("invalid cluster id"), nil, http.StatusNotFound)
			return
		}
		impl.logger.Errorw("service err, FindById", "err", err, "clusterId", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	authenticated, err := impl.CheckRbacForClusterDetails(bean.ClusterId, token)
	if err != nil {
		impl.logger.Errorw("error in checking rbac for cluster", "err", err, "clusterId", bean.ClusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
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
	if util2.IsBaseStack() {
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

func (impl ClusterRestHandlerImpl) UpdateVirtualCluster(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	bean := new(cluster.VirtualClusterBean)
	err = decoder.Decode(bean)
	if err != nil {
		impl.logger.Errorw("request err, Save", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	impl.logger.Infow("request payload, Save", "payload", bean)
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

	bean, err = impl.clusterService.UpdateVirtualCluster(bean, userId)
	if err != nil {
		impl.logger.Errorw("service err, Update", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, bean, http.StatusOK)

}

func (impl ClusterRestHandlerImpl) UpdateClusterNote(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		impl.logger.Errorw("service err, Update", "error", err, "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean cluster.ClusterNoteBean
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
	clusterDescription, err := impl.clusterDescriptionService.FindByClusterIdWithClusterDetails(bean.ClusterId)
	if err != nil {
		impl.logger.Errorw("service err, FindById", "err", err, "clusterId", bean.ClusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying
	if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionUpdate, strings.ToLower(clusterDescription.ClusterName)); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// RBAC enforcer ends

	_, err = impl.clusterNoteService.Update(&bean, userId)
	if err == pg.ErrNoRows {
		_, err = impl.clusterNoteService.Save(&bean, userId)
		if err != nil {
			impl.logger.Errorw("service err, Save", "error", err, "payload", bean)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	if err != nil {
		impl.logger.Errorw("service err, Update", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	userInfo, err := impl.userService.GetById(bean.UpdatedBy)
	if err != nil {
		impl.logger.Errorw("user service err, FindById", "err", err, "userId", bean.UpdatedBy)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	clusterNoteResponseBean := &cluster.ClusterNoteResponseBean{
		Id:          bean.Id,
		Description: bean.Description,
		UpdatedOn:   bean.UpdatedOn,
		UpdatedBy:   userInfo.EmailId,
	}
	if err != nil {
		impl.logger.Errorw("cluster note service err, Update", "error", err, "payload", bean)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, clusterNoteResponseBean, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindAllForAutoComplete(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	clusterList, err := impl.clusterService.FindAllForAutoComplete()
	dbOperationTime := time.Since(start)
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
	start = time.Now()
	for _, item := range clusterList {
		if authEnabled == true {
			if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionGet, item.ClusterName); ok {
				result = append(result, item)
			}
		} else {
			result = append(result, item)
		}

	}
	impl.logger.Infow("Cluster elapsed Time for enforcer", "dbElapsedTime", dbOperationTime, "enforcerTime", time.Since(start), "envSize", len(result))
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

func (impl ClusterRestHandlerImpl) GetAllClusterNamespaces(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	clusterNamespaces := impl.clusterService.GetAllClusterNamespaces()

	// RBAC enforcer applying
	for clusterName, _ := range clusterNamespaces {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceCluster, casbin.ActionGet, strings.ToLower(clusterName)); !ok {
			delete(clusterNamespaces, clusterName)
		}
	}
	//RBAC enforcer Ends

	common.WriteJsonResp(w, nil, clusterNamespaces, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) GetClusterNamespaces(w http.ResponseWriter, r *http.Request) {
	//token := r.Header.Get("token")
	vars := mux.Vars(r)
	clusterIdString := vars["clusterId"]

	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		impl.logger.Errorw("user not authorized", "error", err, "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	clusterId, err := strconv.Atoi(clusterIdString)
	if err != nil {
		impl.logger.Errorw("failed to extract clusterId from param", "error", err, "clusterId", clusterIdString)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	allClusterNamespaces, err := impl.clusterService.FindAllNamespacesByUserIdAndClusterId(userId, clusterId, isActionUserSuperAdmin)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, allClusterNamespaces, http.StatusOK)
}

func (impl ClusterRestHandlerImpl) FindAllForClusterPermission(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		impl.logger.Errorw("user not authorized", "error", err, "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	isActionUserSuperAdmin := false
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isActionUserSuperAdmin = true
	}
	clusterList, err := impl.clusterService.FindAllForClusterByUserId(userId, isActionUserSuperAdmin)
	if err != nil {
		impl.logger.Errorw("error in deleting cluster", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying
	// Already applied at service layer
	//RBAC enforcer Ends

	if len(clusterList) == 0 {
		// assumption is that if list is empty, then it can happen only in case of Unauthorized (but not sending Unauthorized for super-admin user)
		if isActionUserSuperAdmin {
			clusterList = make([]cluster.ClusterBean, 0)
		} else {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	}
	common.WriteJsonResp(w, err, clusterList, http.StatusOK)
}

func (impl *ClusterRestHandlerImpl) CheckRbacForClusterDetails(clusterId int, token string) (authenticated bool, err error) {
	//getting all environments for this cluster
	envs, err := impl.environmentService.GetByClusterId(clusterId)
	if err != nil {
		impl.logger.Errorw("error in getting environments by clusterId", "err", err, "clusterId", clusterId)
		return false, err
	}
	if len(envs) == 0 {
		if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			return false, nil
		}
		return true, nil
	}
	emailId, err := impl.userService.GetEmailFromToken(token)
	if err != nil {
		impl.logger.Errorw("error in getting emailId from token", "err", err)
		return false, err
	}

	var envIdentifierList []string
	envIdentifierMap := make(map[string]bool)
	for _, env := range envs {
		envIdentifier := strings.ToLower(env.EnvironmentIdentifier)
		envIdentifierList = append(envIdentifierList, envIdentifier)
		envIdentifierMap[envIdentifier] = true
	}
	if len(envIdentifierList) == 0 {
		return false, errors.New("environment identifier list for rbac batch enforcing contains zero environments")
	}
	// RBAC enforcer applying
	rbacResultMap := impl.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceGlobalEnvironment, casbin.ActionGet, envIdentifierList)
	for envIdentifier, _ := range envIdentifierMap {
		if rbacResultMap[envIdentifier] {
			//if user has view permission to even one environment of this cluster, authorise the request
			return true, nil
		}
	}
	return false, nil
}

func (impl *ClusterRestHandlerImpl) DeleteVirtualCluster(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		impl.logger.Errorw("service err, Delete", "error", err, "userId", userId)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var bean cluster.VirtualClusterBean
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
	err = impl.clusterService.DeleteFromDbVirtualCluster(&bean, userId)
	if err != nil {
		impl.logger.Errorw("error in deleting cluster", "err", err, "id", bean.Id, "name", bean.ClusterName)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, CLUSTER_DELETE_SUCCESS_RESP, http.StatusOK)
}
