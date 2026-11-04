/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package argoApplication

import (
	"context"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/argoApplication"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/pkg/argoApplication/read"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type ArgoApplicationRestHandler interface {
	ListApplications(w http.ResponseWriter, r *http.Request)
	GetApplicationDetail(w http.ResponseWriter, r *http.Request)
}

type ArgoApplicationRestHandlerImpl struct {
	argoApplicationService argoApplication.ArgoApplicationService
	readService            read.ArgoApplicationReadService
	logger                 *zap.SugaredLogger
	enforcer               casbin.Enforcer
	enforcerUtilGitOps     rbac.EnforcerUtilGitOps
}

func NewArgoApplicationRestHandlerImpl(argoApplicationService argoApplication.ArgoApplicationService,
	readService read.ArgoApplicationReadService, logger *zap.SugaredLogger, enforcer casbin.Enforcer,
	enforcerUtilGitOps rbac.EnforcerUtilGitOps) *ArgoApplicationRestHandlerImpl {
	return &ArgoApplicationRestHandlerImpl{
		argoApplicationService: argoApplicationService,
		readService:            readService,
		logger:                 logger,
		enforcer:               enforcer,
		enforcerUtilGitOps:     enforcerUtilGitOps,
	}

}

func (handler *ArgoApplicationRestHandlerImpl) ListApplications(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	v := r.URL.Query()
	clusterIdString := v.Get("clusterIds")
	var clusterIds []int
	if clusterIdString != "" {
		clusterIdSlices := strings.Split(clusterIdString, ",")
		for _, clusterId := range clusterIdSlices {
			id, err := strconv.Atoi(clusterId)
			if err != nil {
				handler.logger.Errorw("error in converting clusterId", "err", err, "clusterIdString", clusterIdString)
				common.WriteJsonResp(w, err, "please send valid cluster Ids", http.StatusBadRequest)
				return
			}
			clusterIds = append(clusterIds, id)
		}
	}
	resp, err := handler.argoApplicationService.ListApplications(clusterIds)
	if err != nil {
		handler.logger.Errorw("error in listing all argo applications", "err", err, "clusterIds", clusterIds)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying: filter the listing to the applications the caller may see.
	// Batched rather than a per-app Enforce loop; an app whose object cannot be built is dropped.
	objects := make([]string, 0, len(resp))
	objectByApp := make(map[*bean.ArgoApplicationListDto]string, len(resp))
	for _, app := range resp {
		object := handler.enforcerUtilGitOps.GetExternalGitOpsAppObjectByClusterName(app.ClusterName, app.Namespace, app.Name)
		if len(object) == 0 {
			continue
		}
		objectByApp[app] = object
		objects = append(objects, object)
	}
	authorisedObjects := make(map[string]bool)
	if len(objects) > 0 {
		authorisedObjects = handler.enforcer.EnforceInBatch(token, casbin.ResourceArgoApp, casbin.ActionGet, objects)
	}
	authorisedApps := make([]*bean.ArgoApplicationListDto, 0, len(resp))
	for _, app := range resp {
		if object, ok := objectByApp[app]; ok && authorisedObjects[strings.ToLower(object)] {
			authorisedApps = append(authorisedApps, app)
		}
	}
	//RBAC enforcer Ends
	common.WriteJsonResp(w, nil, authorisedApps, http.StatusOK)
}

func (handler *ArgoApplicationRestHandlerImpl) GetApplicationDetail(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	ctx := r.Context()
	ctx = context.WithValue(ctx, "token", token)

	var err error
	v := r.URL.Query()
	resourceName := v.Get("name")
	namespace := v.Get("namespace")
	clusterIdString := v.Get("clusterId")

	var clusterId int
	if clusterIdString != "" {
		clusterId, err = strconv.Atoi(clusterIdString)
		if err != nil {
			handler.logger.Errorw("error in converting clusterId", "err", err, "clusterIdString", clusterIdString)
			common.WriteJsonResp(w, err, "please send valid cluster Ids", http.StatusBadRequest)
			return
		}
	}
	// RBAC enforcer applying
	object := handler.enforcerUtilGitOps.GetExternalGitOpsAppObject(clusterId, namespace, resourceName)
	if len(object) == 0 || !handler.enforcer.Enforce(token, casbin.ResourceArgoApp, casbin.ActionGet, object) {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	resp, err := handler.readService.GetAppDetailEA(ctx, resourceName, namespace, clusterId)
	if err != nil {
		handler.logger.Errorw("error in getting argo application app detail", "err", err, "resourceName", resourceName, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
