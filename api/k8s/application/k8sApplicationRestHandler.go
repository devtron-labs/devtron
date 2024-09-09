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

package application

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/common-lib/utils"
	util3 "github.com/devtron-labs/common-lib/utils/k8s"
	k8sCommonBean "github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/common-lib/utils/k8sObjectsUtil"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/connector"
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	client "github.com/devtron-labs/devtron/api/helm-app/service"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/argoApplication"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/cluster"
	clientErrors "github.com/devtron-labs/devtron/pkg/errors"
	"github.com/devtron-labs/devtron/pkg/fluxApplication"
	"github.com/devtron-labs/devtron/pkg/k8s"
	application2 "github.com/devtron-labs/devtron/pkg/k8s/application"
	bean2 "github.com/devtron-labs/devtron/pkg/k8s/application/bean"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	errors2 "github.com/juju/errors"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"io"
	errors3 "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type K8sApplicationRestHandler interface {
	GetResource(w http.ResponseWriter, r *http.Request)
	CreateResource(w http.ResponseWriter, r *http.Request)
	UpdateResource(w http.ResponseWriter, r *http.Request)
	DeleteResource(w http.ResponseWriter, r *http.Request)
	ListEvents(w http.ResponseWriter, r *http.Request)
	GetPodLogs(w http.ResponseWriter, r *http.Request)
	DownloadPodLogs(w http.ResponseWriter, r *http.Request)
	GetTerminalSession(w http.ResponseWriter, r *http.Request)
	GetResourceInfo(w http.ResponseWriter, r *http.Request)
	GetHostUrlsByBatch(w http.ResponseWriter, r *http.Request)
	GetAllApiResources(w http.ResponseWriter, r *http.Request)
	GetResourceList(w http.ResponseWriter, r *http.Request)
	ApplyResources(w http.ResponseWriter, r *http.Request)
	RotatePod(w http.ResponseWriter, r *http.Request)
	CreateEphemeralContainer(w http.ResponseWriter, r *http.Request)
	DeleteEphemeralContainer(w http.ResponseWriter, r *http.Request)
	GetAllApiResourceGVKWithoutAuthorization(w http.ResponseWriter, r *http.Request)
}

type K8sApplicationRestHandlerImpl struct {
	logger                 *zap.SugaredLogger
	k8sApplicationService  application2.K8sApplicationService
	pump                   connector.Pump
	terminalSessionHandler terminal.TerminalSessionHandler
	enforcer               casbin.Enforcer
	validator              *validator.Validate
	enforcerUtil           rbac.EnforcerUtil
	enforcerUtilHelm       rbac.EnforcerUtilHelm
	helmAppService         client.HelmAppService
	userService            user.UserService
	k8sCommonService       k8s.K8sCommonService
	terminalEnvVariables   *util.TerminalEnvVariables
	fluxAppService         fluxApplication.FluxApplicationService
	argoApplication        argoApplication.ArgoApplicationService
}

func NewK8sApplicationRestHandlerImpl(logger *zap.SugaredLogger, k8sApplicationService application2.K8sApplicationService, pump connector.Pump, terminalSessionHandler terminal.TerminalSessionHandler, enforcer casbin.Enforcer, enforcerUtilHelm rbac.EnforcerUtilHelm, enforcerUtil rbac.EnforcerUtil, helmAppService client.HelmAppService, userService user.UserService, k8sCommonService k8s.K8sCommonService, validator *validator.Validate, envVariables *util.EnvironmentVariables, fluxAppService fluxApplication.FluxApplicationService, argoApplication argoApplication.ArgoApplicationService,
) *K8sApplicationRestHandlerImpl {
	return &K8sApplicationRestHandlerImpl{
		logger:                 logger,
		k8sApplicationService:  k8sApplicationService,
		pump:                   pump,
		terminalSessionHandler: terminalSessionHandler,
		enforcer:               enforcer,
		validator:              validator,
		enforcerUtilHelm:       enforcerUtilHelm,
		enforcerUtil:           enforcerUtil,
		helmAppService:         helmAppService,
		userService:            userService,
		k8sCommonService:       k8sCommonService,
		terminalEnvVariables:   envVariables.TerminalEnvVariables,
		fluxAppService:         fluxAppService,
		argoApplication:        argoApplication,
	}
}

func (handler *K8sApplicationRestHandlerImpl) RotatePod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appIdString := vars["appId"]
	if appIdString == "" {
		common.WriteJsonResp(w, fmt.Errorf("empty appid in request"), nil, http.StatusBadRequest)
		return
	}
	decoder := json.NewDecoder(r.Body)
	podRotateRequest := &k8s.RotatePodRequest{}
	err := decoder.Decode(podRotateRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appIdentifier, err := handler.helmAppService.DecodeAppId(appIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject2)
	if !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	handler.logger.Infow("rotate pod request", "payload", podRotateRequest)
	rotatePodRequest := &k8s.RotatePodRequest{
		ClusterId: appIdentifier.ClusterId,
		Resources: podRotateRequest.Resources,
	}
	response, err := handler.k8sCommonService.RotatePods(r.Context(), rotatePodRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) GetResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request k8s.ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")

	//rbac validation for the apps requests
	if request.AppId != "" {
		ok, err := handler.verifyRbacForAppRequests(token, &request, r, casbin.ActionGet)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else if !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	}
	// Invalid cluster id
	if request.ClusterId <= 0 {
		common.WriteJsonResp(w, errors.New("can not resource manifest as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}
	// Fetching requested resource
	resource, err := handler.k8sCommonService.GetResource(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in getting resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	if resource != nil && resource.ManifestResponse != nil {
		err = resource.ManifestResponse.SetRunningEphemeralContainers()
		if err != nil {
			handler.logger.Errorw("error in setting running ephemeral containers and setting them in resource response", "err", err)
			common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
			return
		}
	}

	canUpdate := false
	// Obfuscate secret if user does not have edit access
	if request.AppIdentifier == nil && request.DevtronAppIdentifier == nil && request.AppType != bean2.ArgoAppType && request.ClusterId > 0 { // if the appType is not argoAppType,then verify logic w.r.t resource browser, when rbac for argoApp is introduced, handle rbac accordingly
		// Verify update access for Resource Browser
		canUpdate = handler.k8sApplicationService.ValidateClusterResourceBean(r.Context(), request.ClusterId, resource.ManifestResponse.Manifest, request.K8sRequest.ResourceIdentifier.GroupVersionKind, handler.getRbacCallbackForResource(token, casbin.ActionUpdate))
		if !canUpdate {
			// Verify read access for Resource Browser
			readAllowed := handler.k8sApplicationService.ValidateClusterResourceBean(r.Context(), request.ClusterId, resource.ManifestResponse.Manifest, request.K8sRequest.ResourceIdentifier.GroupVersionKind, handler.getRbacCallbackForResource(token, casbin.ActionGet))
			if !readAllowed {
				common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
				return
			}
		}
	}
	if !canUpdate && resource != nil {
		// Hide secret for read only access
		modifiedManifest, err := k8sObjectsUtil.HideValuesIfSecret(&resource.ManifestResponse.Manifest)
		if err != nil {
			handler.logger.Errorw("error in hiding secret values", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		resource.ManifestResponse.Manifest = *modifiedManifest
	}
	// setting flag for secret view access only for resource browser
	resource.SecretViewAccess = canUpdate

	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}
func (handler *K8sApplicationRestHandlerImpl) GetHostUrlsByBatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appIdString := vars["appId"]
	if appIdString == "" {
		common.WriteJsonResp(w, fmt.Errorf("empty appid in request"), nil, http.StatusBadRequest)
		return
	}
	appTypeString := vars["appType"]
	if appTypeString == "" {
		common.WriteJsonResp(w, fmt.Errorf("empty appType in request"), nil, http.StatusBadRequest)
		return
	}
	appType, err := strconv.Atoi(appTypeString)
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("invalid appType in request"), nil, http.StatusBadRequest)
		return
	}

	token := r.Header.Get("token")
	var k8sAppDetail bean.AppDetailContainer
	var resourceTreeResponse *gRPC.ResourceTreeResponse
	var clusterId int
	var namespace string
	var resourceTreeInf map[string]interface{}
	var externalArgoApplicationName string

	if appType == bean2.HelmAppType {
		appIdentifier, err := handler.helmAppService.DecodeAppId(appIdString)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		// RBAC enforcer applying
		rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)

		ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)

		if !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends
		appDetail, err := handler.helmAppService.GetApplicationDetail(r.Context(), appIdentifier)
		if err != nil {
			apiError := clientErrors.ConvertToApiError(err)
			if apiError != nil {
				err = apiError
			}
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}

		clusterId = appIdentifier.ClusterId
		namespace = appIdentifier.Namespace
		resourceTreeResponse = appDetail.ResourceTreeResponse

	} else if appType == bean2.ArgoAppType {
		appIdentifier, err := argoApplication.DecodeExternalArgoAppId(appIdString)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		// RBAC enforcer applying
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends

		appDetail, err := handler.argoApplication.GetAppDetail(appIdentifier.AppName, appIdentifier.Namespace, appIdentifier.ClusterId)
		if err != nil {
			apiError := clientErrors.ConvertToApiError(err)
			if apiError != nil {
				err = apiError
			}
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		clusterId = appIdentifier.ClusterId
		namespace = appIdentifier.Namespace
		resourceTreeResponse = appDetail.ResourceTree
		externalArgoApplicationName = appIdentifier.AppName

	} else if appType == bean2.FluxAppType {
		appIdentifier, err := fluxApplication.DecodeFluxExternalAppId(appIdString)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		// RBAC enforcer applying
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends

		appDetail, err := handler.fluxAppService.GetFluxAppDetail(r.Context(), appIdentifier)
		if err != nil {
			apiError := clientErrors.ConvertToApiError(err)
			if apiError != nil {
				err = apiError
			}
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		clusterId = appIdentifier.ClusterId
		namespace = appIdentifier.Namespace
		resourceTreeResponse = appDetail.ResourceTreeResponse
	}

	k8sAppDetail = bean.AppDetailContainer{
		DeploymentDetailContainer: bean.DeploymentDetailContainer{
			ClusterId: clusterId,
			Namespace: namespace,
		},
	}

	bytes, _ := json.Marshal(resourceTreeResponse)
	err = json.Unmarshal(bytes, &resourceTreeInf)
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("unmarshal error of resource tree response"), nil, http.StatusInternalServerError)
		return
	}

	validRequests := handler.k8sCommonService.FilterK8sResources(r.Context(), resourceTreeInf, k8sAppDetail, appIdString, []string{k8sCommonBean.ServiceKind, k8sCommonBean.IngressKind}, externalArgoApplicationName)
	if len(validRequests) == 0 {
		handler.logger.Error("neither service nor ingress found for this app", "appId", appIdString)
		common.WriteJsonResp(w, err, nil, http.StatusNoContent)
		return
	}

	resp, err := handler.k8sCommonService.GetManifestsByBatch(r.Context(), validRequests)
	if err != nil {
		handler.logger.Errorw("error in getting manifests in batch", "err", err, "clusterId", k8sAppDetail.ClusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	result := handler.k8sApplicationService.GetUrlsByBatchForIngress(r.Context(), resp)
	common.WriteJsonResp(w, nil, result, http.StatusOK)

}
func (handler *K8sApplicationRestHandlerImpl) CreateResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request k8s.ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appIdentifier, err := handler.helmAppService.DecodeAppId(request.AppId)
	if err != nil {
		handler.logger.Errorw("error in decoding appId", "err", err, "appId", request.AppId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//setting appIdentifier value in request
	request.AppIdentifier = appIdentifier
	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
	token := r.Header.Get("token")
	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject2)
	if !ok {
		common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	resource, err := handler.k8sApplicationService.RecreateResource(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in creating resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}
func (handler *K8sApplicationRestHandlerImpl) UpdateResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request k8s.ResourceRequestBean
	token := r.Header.Get("token")
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//rbac validation for the apps requests
	if request.AppId != "" {
		ok, err := handler.verifyRbacForAppRequests(token, &request, r, casbin.ActionUpdate)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else if !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	} else if request.ClusterId > 0 {
		// RBAC enforcer applying for Resource Browser
		if ok := handler.handleRbac(r, w, request, token, casbin.ActionUpdate); !ok {
			return
		}
		// RBAC enforcer Ends
	} else {
		common.WriteJsonResp(w, errors.New("can not update resource as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}

	resource, err := handler.k8sCommonService.UpdateResource(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in updating resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}
func (handler *K8sApplicationRestHandlerImpl) handleRbac(r *http.Request, w http.ResponseWriter, request k8s.ResourceRequestBean, token string, casbinAction string) bool {
	// assume direct update in cluster
	allowed, err := handler.k8sApplicationService.ValidateClusterResourceRequest(r.Context(), &request, handler.getRbacCallbackForResource(token, casbinAction))
	if err != nil {
		common.WriteJsonResp(w, errors.New("invalid request"), nil, http.StatusBadRequest)
		return false
	}
	if !allowed {
		common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
	}
	return allowed
}
func (handler *K8sApplicationRestHandlerImpl) DeleteResource(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request k8s.ResourceRequestBean
	err = json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	vars := r.URL.Query()
	request.ExternalArgoApplicationName = vars.Get("externalArgoApplicationName")

	//rbac handle for the apps requests
	if request.AppId != "" {
		ok, err := handler.verifyRbacForAppRequests(token, &request, r, casbin.ActionDelete)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else if !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	} else if request.ClusterId > 0 {
		// RBAC enforcer applying for resource Browser
		if ok := handler.handleRbac(r, w, request, token, casbin.ActionDelete); !ok {
			return
		}
		// RBAC enforcer Ends
	} else {
		common.WriteJsonResp(w, errors.New("can not delete resource as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}

	resource, err := handler.k8sApplicationService.DeleteResourceWithAudit(r.Context(), &request, userId)
	if err != nil {
		errCode := http.StatusInternalServerError
		if apiErr, ok := err.(*utils.ApiError); ok {
			errCode = apiErr.HttpStatusCode
			switch errCode {
			case http.StatusNotFound:
				errorMessage := k8s.ResourceNotFoundErr
				err = fmt.Errorf("%s: %w", errorMessage, err)
			}
		}
		handler.logger.Errorw("error in deleting resource", "err", err)
		common.WriteJsonResp(w, err, resource, errCode)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}
func (handler *K8sApplicationRestHandlerImpl) ListEvents(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	token := r.Header.Get("token")
	var request k8s.ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//rbac validation for the apps requests
	if request.AppId != "" {
		ok, err := handler.verifyRbacForAppRequests(token, &request, r, casbin.ActionGet)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else if !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized user"), nil, http.StatusForbidden)
			return
		}
	} else if request.ClusterId > 0 {
		// RBAC enforcer applying for resource Browser
		if ok := handler.handleRbac(r, w, request, token, casbin.ActionGet); !ok {
			return
		}
		// RBAC enforcer Ends
	} else {
		common.WriteJsonResp(w, errors.New("can not get resource as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}
	events, err := handler.k8sCommonService.ListEvents(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in getting events list", "err", err)
		common.WriteJsonResp(w, err, events, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, events, http.StatusOK)
}
func (handler *K8sApplicationRestHandlerImpl) GetPodLogs(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	request, err := handler.k8sApplicationService.ValidatePodLogsRequestQuery(r)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("get pod logs request", "request", request)
	handler.requestValidationAndRBAC(w, r, token, request)
	lastEventId := r.Header.Get(bean2.LastEventID)
	isReconnect := false
	if len(lastEventId) > 0 {
		lastSeenMsgId, err := strconv.ParseInt(lastEventId, bean2.IntegerBase, bean2.IntegerBitSize)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		lastSeenMsgId = lastSeenMsgId + bean2.TimestampOffsetToAvoidDuplicateLogs //increased by one ns to avoid duplicate
		t := v1.Unix(0, lastSeenMsgId)
		request.K8sRequest.PodLogsRequest.SinceTime = &t
		isReconnect = true
	}
	stream, err := handler.k8sApplicationService.GetPodLogs(r.Context(), request)
	//err is handled inside StartK8sStreamWithHeartBeat method
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
	defer cancel()
	defer util.Close(stream, handler.logger)
	handler.pump.StartK8sStreamWithHeartBeat(w, isReconnect, stream, err)
}

func (handler *K8sApplicationRestHandlerImpl) DownloadPodLogs(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	request, err := handler.k8sApplicationService.ValidatePodLogsRequestQuery(r)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.requestValidationAndRBAC(w, r, token, request)

	// just to make sure follow flag is set to false when downloading logs
	request.K8sRequest.PodLogsRequest.Follow = false

	stream, err := handler.k8sApplicationService.GetPodLogs(r.Context(), request)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
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
	defer cancel()
	defer util.Close(stream, handler.logger)

	var dataBuffer bytes.Buffer
	bufReader := bufio.NewReader(stream)
	eof := false
	for !eof {
		log, err := bufReader.ReadString('\n')
		log = strings.TrimSpace(log) // Remove trailing line ending
		a := regexp.MustCompile(" ")
		var res []byte
		splitLog := a.Split(log, 2)
		if len(splitLog[0]) > 0 {
			parsedTime, err := time.Parse(time.RFC3339, splitLog[0])
			if err != nil {
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
			gmtTimeLoc := time.FixedZone(bean2.LocalTimezoneInGMT, bean2.LocalTimeOffset)
			humanReadableTime := parsedTime.In(gmtTimeLoc).Format(time.RFC1123)
			res = append(res, humanReadableTime...)
		}

		if len(splitLog) == 2 {
			res = append(res, " "...)
			res = append(res, splitLog[1]...)
		}
		res = append(res, "\n"...)
		if err == io.EOF {
			eof = true
			// stop if we reached end of stream and the next line is empty
			if log == "" {
				break
			}
		} else if err != nil && err != io.EOF {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		_, err = dataBuffer.Write(res)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	if len(dataBuffer.Bytes()) == 0 {
		common.WriteJsonResp(w, nil, nil, http.StatusNoContent)
		return
	}
	podLogsFilename := generatePodLogsFilename(request.K8sRequest.ResourceIdentifier.Name)
	common.WriteOctetStreamResp(w, r, dataBuffer.Bytes(), podLogsFilename)
	return
}

func generatePodLogsFilename(filename string) string {
	return fmt.Sprintf("podlogs-%s-%s.log", filename, uuid.New().String())
}

func (handler *K8sApplicationRestHandlerImpl) requestValidationAndRBAC(w http.ResponseWriter, r *http.Request, token string, request *k8s.ResourceRequestBean) {
	if request.AppType == bean2.HelmAppType && request.AppIdentifier != nil {
		if request.DeploymentType == bean2.HelmInstalledType {
			if err := handler.k8sApplicationService.ValidateResourceRequest(r.Context(), request.AppIdentifier, request.K8sRequest); err != nil {
				common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
				return
			}
		} else if request.DeploymentType == bean2.ArgoInstalledType {
			//TODO Implement ResourceRequest Validation for ArgoCD Installed APPs From ResourceTree
		}
		// RBAC enforcer applying for Helm App
		rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
		ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)

		if !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends
	} else if request.AppType == bean2.DevtronAppType && request.DevtronAppIdentifier != nil {
		if request.DeploymentType == bean2.HelmInstalledType {
			//TODO Implement ResourceRequest Validation for Helm Installed Devtron APPs
		} else if request.DeploymentType == bean2.ArgoInstalledType {
			//TODO Implement ResourceRequest Validation for ArgoCD Installed APPs From ResourceTree
		}
		// RBAC enforcer applying For Devtron App
		envObject := handler.enforcerUtil.GetEnvRBACNameByAppId(request.DevtronAppIdentifier.AppId, request.DevtronAppIdentifier.EnvId)
		if !handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionGet, envObject) {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends
	} else if request.AppType == bean2.FluxAppType && request.ExternalFluxAppIdentifier != nil {
		valid, err := handler.k8sApplicationService.ValidateFluxResourceRequest(r.Context(), request.ExternalFluxAppIdentifier, request.K8sRequest)
		if err != nil || !valid {
			handler.logger.Errorw("error in validating resource request", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		//RBAC enforcer starts here
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer ends here
	} else if request.AppType == bean2.ArgoAppType && request.ExternalArgoApplicationName != "" {
		appIdentifier, err := argoApplication.DecodeExternalArgoAppId(request.AppId)
		if err != nil {
			handler.logger.Errorw(bean2.AppIdDecodingError, "err", err, "appIdentifier", request.AppIdentifier)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		}
		valid, err := handler.k8sApplicationService.ValidateArgoResourceRequest(r.Context(), appIdentifier, request.K8sRequest)
		if err != nil || !valid {
			handler.logger.Errorw("error in validating resource request", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}

		//RBAC enforcer starts here
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer ends here
	} else if request.AppIdentifier == nil && request.DevtronAppIdentifier == nil && request.ClusterId > 0 && request.ExternalArgoApplicationName == "" {
		//RBAC enforcer applying For Resource Browser
		if !handler.handleRbac(r, w, *request, token, casbin.ActionGet) {
			return
		}
		//RBAC enforcer Ends
	} else if request.ClusterId <= 0 {
		common.WriteJsonResp(w, errors.New("can not get pod logs as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}
}

func (handler *K8sApplicationRestHandlerImpl) restrictTerminalAccessForNonSuperUsers(w http.ResponseWriter, token string) bool {
	// if RESTRICT_TERMINAL_ACCESS_FOR_NON_SUPER_USER is set to true, only super admins can access terminal/ephemeral containers
	if handler.terminalEnvVariables.RestrictTerminalAccessForNonSuperUser {
		if isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
			common.WriteJsonResp(w, errors.New("unauthorized, only super-admins can access terminal"), nil, http.StatusForbidden)
			return true
		}
	}
	return false
}

func (handler *K8sApplicationRestHandlerImpl) GetTerminalSession(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	request, resourceRequestBean, err := handler.k8sApplicationService.ValidateTerminalRequestQuery(r)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// check for super admin
	restricted := handler.restrictTerminalAccessForNonSuperUsers(w, token)
	if restricted {
		return
	}
	if resourceRequestBean.AppIdentifier != nil {
		// RBAC enforcer applying For Helm App
		rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(resourceRequestBean.AppIdentifier.ClusterId, resourceRequestBean.AppIdentifier.Namespace, resourceRequestBean.AppIdentifier.ReleaseName)
		ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, "*", rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, "*", rbacObject2)

		if !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends
	} else if resourceRequestBean.DevtronAppIdentifier != nil {
		// RBAC enforcer applying For Devtron App
		envObject := handler.enforcerUtil.GetEnvRBACNameByAppId(resourceRequestBean.DevtronAppIdentifier.AppId, resourceRequestBean.DevtronAppIdentifier.EnvId)
		if !handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, envObject) {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends
	} else if resourceRequestBean.ExternalFluxAppIdentifier != nil {
		// RBAC enforcer applying For external flux app
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends

	} else if resourceRequestBean.ExternalArgoApplicationName != "" {
		// RBAC enforcer applying For external Argo app
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends

	} else if resourceRequestBean.AppIdentifier == nil && resourceRequestBean.DevtronAppIdentifier == nil && resourceRequestBean.ExternalFluxAppIdentifier == nil && resourceRequestBean.ExternalArgoApplicationName == "" && resourceRequestBean.ClusterId > 0 {
		//RBAC enforcer applying for Resource Browser
		if !handler.handleRbac(r, w, *resourceRequestBean, token, casbin.ActionUpdate) {
			return
		}
		//RBAC enforcer Ends
	} else if resourceRequestBean.ClusterId <= 0 {
		common.WriteJsonResp(w, errors.New("can not get terminal session as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	status, message, err := handler.terminalSessionHandler.GetTerminalSession(request)
	common.WriteJsonResp(w, err, message, status)
}

func (handler *K8sApplicationRestHandlerImpl) GetResourceInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// this is auth free api
	response, err := handler.k8sApplicationService.GetResourceInfo(r.Context())
	if err != nil {
		handler.logger.Errorw("error on resource info", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, response, http.StatusOK)
	return
}

// GetAllApiResourceGVKWithoutAuthorization  This function will the all the available api resource GVK list for specific cluster
func (handler *K8sApplicationRestHandlerImpl) GetAllApiResourceGVKWithoutAuthorization(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// get clusterId from request
	vars := mux.Vars(r)
	clusterId, err := strconv.Atoi(vars["clusterId"])
	if err != nil {
		handler.logger.Errorw("request err in getting clusterId in GetAllApiResources", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// get data from service
	response, err := handler.k8sApplicationService.GetAllApiResourceGVKWithoutAuthorization(r.Context(), clusterId)
	if err != nil {
		handler.logger.Errorw("error in getting api-resources", "clusterId", clusterId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) GetAllApiResources(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// get clusterId from request
	vars := mux.Vars(r)
	clusterId, err := strconv.Atoi(vars["clusterId"])
	if err != nil {
		handler.logger.Errorw("request err in getting clusterId in GetAllApiResources", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	isSuperAdmin := false
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isSuperAdmin = true
	}

	// get data from service
	response, err := handler.k8sApplicationService.GetAllApiResources(r.Context(), clusterId, isSuperAdmin, userId)
	if err != nil {
		handler.logger.Errorw("error in getting api-resources", "clusterId", clusterId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	// send unauthorised if response is empty
	if !isSuperAdmin && (response == nil || len(response.ApiResources) == 0) {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) GetResourceList(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	token := r.Header.Get("token")
	var request k8s.ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	isSuperAdmin := false
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); ok {
		isSuperAdmin = true
	}
	clusterRbacFunc := handler.verifyRbacForCluster
	if isSuperAdmin {
		clusterRbacFunc = func(token, clusterName string, request k8s.ResourceRequestBean, casbinAction string) bool {
			return true
		}
	}
	response, err := handler.k8sApplicationService.GetResourceList(r.Context(), token, &request, clusterRbacFunc)
	if err != nil {
		handler.logger.Errorw("error in getting resource list", "err", err)
		if statusErr, ok := err.(*errors3.StatusError); ok && statusErr.Status().Code == 404 {
			err = &util2.ApiError{Code: "404", HttpStatusCode: 404, UserMessage: "no resource found", InternalMessage: err.Error()}
		}
		common.WriteJsonResp(w, err, response, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) ApplyResources(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request util3.ApplyResourcesRequest
	token := r.Header.Get("token")
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	response, err := handler.k8sApplicationService.ApplyResources(r.Context(), token, &request, handler.verifyRbacForCluster)
	if err != nil {
		handler.logger.Errorw("error in applying resource", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) getRbacCallbackForResource(token string, casbinAction string) func(clusterName string, resourceIdentifier util3.ResourceIdentifier) bool {
	return func(clusterName string, resourceIdentifier util3.ResourceIdentifier) bool {
		return handler.verifyRbacForResource(token, clusterName, resourceIdentifier, casbinAction)
	}
}

func (handler *K8sApplicationRestHandlerImpl) verifyRbacForResource(token string, clusterName string, resourceIdentifier util3.ResourceIdentifier, casbinAction string) bool {
	resourceName, objectName := handler.enforcerUtil.GetRBACNameForClusterEntity(clusterName, resourceIdentifier)
	return handler.enforcer.Enforce(token, strings.ToLower(resourceName), casbinAction, objectName)
}

func (handler *K8sApplicationRestHandlerImpl) verifyRbacForCluster(token string, clusterName string, request k8s.ResourceRequestBean, casbinAction string) bool {
	k8sRequest := request.K8sRequest
	return handler.verifyRbacForResource(token, clusterName, k8sRequest.ResourceIdentifier, casbinAction)
}

func (handler *K8sApplicationRestHandlerImpl) CreateEphemeralContainer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request cluster.EphemeralContainerRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if err = handler.validator.Struct(request); err != nil || (request.BasicData == nil && request.AdvancedData == nil) {
		if err != nil {
			err = errors.New("invalid request payload")
		}
		handler.logger.Errorw("invalid request payload", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// check for super admin
	restricted := handler.restrictTerminalAccessForNonSuperUsers(w, token)
	if restricted {
		return
	}
	//rbac applied in below function
	resourceRequestBean := handler.handleEphemeralRBAC(request.PodName, request.Namespace, w, r)
	if resourceRequestBean == nil {
		return
	}
	if resourceRequestBean.ClusterId != request.ClusterId {
		common.WriteJsonResp(w, errors.New("clusterId mismatch in param and request body"), nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	err = handler.k8sApplicationService.CreatePodEphemeralContainers(&request)
	if err != nil {
		handler.logger.Errorw("error occurred in creating ephemeral container", "err", err, "requestPayload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, request.BasicData.ContainerName, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) DeleteEphemeralContainer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var request cluster.EphemeralContainerRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if err = handler.validator.Struct(request); err != nil || request.BasicData == nil {
		if err != nil {
			err = errors.New("invalid request payload")
		}
		handler.logger.Errorw("invalid request payload", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// check for super admin
	restricted := handler.restrictTerminalAccessForNonSuperUsers(w, token)
	if restricted {
		return
	}
	//rbac applied in below function
	resourceRequestBean := handler.handleEphemeralRBAC(request.PodName, request.Namespace, w, r)
	if resourceRequestBean == nil {
		return
	}
	if resourceRequestBean.ClusterId != request.ClusterId {
		common.WriteJsonResp(w, errors.New("clusterId mismatch in param and request body"), nil, http.StatusBadRequest)
		return
	}
	request.UserId = userId
	_, err = handler.k8sApplicationService.TerminatePodEphemeralContainer(request)
	if err != nil {
		handler.logger.Errorw("error occurred in terminating ephemeral container", "err", err, "requestPayload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, request.BasicData.ContainerName, http.StatusOK)

}

func (handler *K8sApplicationRestHandlerImpl) handleEphemeralRBAC(podName, namespace string, w http.ResponseWriter, r *http.Request) *k8s.ResourceRequestBean {
	token := r.Header.Get("token")
	_, resourceRequestBean, err := handler.k8sApplicationService.ValidateTerminalRequestQuery(r)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return resourceRequestBean
	}
	if resourceRequestBean.AppIdentifier != nil {
		// RBAC enforcer applying For Helm App
		rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(resourceRequestBean.AppIdentifier.ClusterId, resourceRequestBean.AppIdentifier.Namespace, resourceRequestBean.AppIdentifier.ReleaseName)
		ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject2)

		if !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return resourceRequestBean
		}
		//RBAC enforcer Ends
	} else if resourceRequestBean.DevtronAppIdentifier != nil {
		// RBAC enforcer applying For Devtron App
		envObject := handler.enforcerUtil.GetEnvRBACNameByAppId(resourceRequestBean.DevtronAppIdentifier.AppId, resourceRequestBean.DevtronAppIdentifier.EnvId)
		if !handler.enforcer.Enforce(token, casbin.ResourceEnvironment, casbin.ActionUpdate, envObject) {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return resourceRequestBean
		}
		//RBAC enforcer Ends
	} else if resourceRequestBean.ExternalFluxAppIdentifier != nil {
		//RBAC enforcer starts here
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return resourceRequestBean
		}
		//RBAC enforcer ends here
	} else if resourceRequestBean.ExternalArgoApplicationName != "" {
		//RBAC enforcer starts here
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return resourceRequestBean
		}
		//RBAC enforcer ends here

	} else if resourceRequestBean.AppIdentifier == nil && resourceRequestBean.DevtronAppIdentifier == nil && resourceRequestBean.ExternalArgoApplicationName == "" && resourceRequestBean.ExternalFluxAppIdentifier == nil && resourceRequestBean.ClusterId > 0 {
		//RBAC enforcer applying for Resource Browser
		resourceRequestBean.K8sRequest.ResourceIdentifier.Name = podName
		resourceRequestBean.K8sRequest.ResourceIdentifier.Namespace = namespace
		if !handler.handleRbac(r, w, *resourceRequestBean, token, casbin.ActionUpdate) {
			return resourceRequestBean
		}
		//RBAC enforcer Ends
	} else if resourceRequestBean.ClusterId <= 0 {
		common.WriteJsonResp(w, errors.New("can not create/terminate ephemeral containers as target cluster is not provided"), nil, http.StatusBadRequest)
		return resourceRequestBean
	}
	return resourceRequestBean
}

/*
	    true and err =!nil  --> not possible [indicates that authorized but error has occurred too.]
		true and err ==nil -->  Denotes that user is authorized without any error, we can proceed
		false and err !=nil --> during the validation of resources, we got an error, resulting the StatusBadRequest
		false and err == nil --> denotes that user is not authorized, resulting in Unauthorized
*/
func (handler *K8sApplicationRestHandlerImpl) verifyRbacForAppRequests(token string, request *k8s.ResourceRequestBean, r *http.Request, actionType string) (bool, error) {
	rbacObject := ""
	rbacObject2 := ""
	envObject := ""
	switch request.AppType {
	case bean2.ArgoAppType:
		argoAppIdentifier, err := argoApplication.DecodeExternalArgoAppId(request.AppId)
		if err != nil {
			handler.logger.Errorw("error in decoding appId", "err", err, "appId", request.AppId)
			return false, err
		}
		request.ClusterId = argoAppIdentifier.ClusterId
		request.ExternalArgoApplicationName = argoAppIdentifier.AppName
		valid, err := handler.k8sApplicationService.ValidateArgoResourceRequest(r.Context(), argoAppIdentifier, request.K8sRequest)
		if err != nil || !valid {
			handler.logger.Errorw("error in validating resource request", "err", err)
			return false, err
		}
		//RBAC enforcer starts here
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, actionType, "*"); !ok {
			return false, nil
		}
		return true, nil
		//RBAC enforcer ends here

	case bean2.HelmAppType:
		appIdentifier, err := handler.helmAppService.DecodeAppId(request.AppId)
		if err != nil {
			handler.logger.Errorw("error in decoding appId", "err", err, "appId", request.AppId)
			return false, err
		}
		//setting appIdentifier value in request
		request.AppIdentifier = appIdentifier
		request.ClusterId = request.AppIdentifier.ClusterId
		if request.DeploymentType == bean2.HelmInstalledType {
			if err := handler.k8sApplicationService.ValidateResourceRequest(r.Context(), request.AppIdentifier, request.K8sRequest); err != nil {
				return false, err
			}
		} else if request.DeploymentType == bean2.ArgoInstalledType {
			//TODO Implement ResourceRequest Validation for ArgoCD Installed APPs From ResourceTree
		}
		// RBAC enforcer applying for Helm App
		rbacObject, rbacObject2 = handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
		ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, actionType, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, actionType, rbacObject2)
		if !ok {
			return false, nil
		}
		return true, nil
		// RBAC enforcer Ends
	case bean2.DevtronAppType:
		devtronAppIdentifier, err := handler.k8sApplicationService.DecodeDevtronAppId(request.AppId)
		if err != nil {
			handler.logger.Errorw("error in decoding appId", "err", err, "appId", request.AppId)
			return false, err
		}
		//setting devtronAppIdentifier value in request
		request.DevtronAppIdentifier = devtronAppIdentifier
		request.ClusterId = request.DevtronAppIdentifier.ClusterId
		if request.DeploymentType == bean2.HelmInstalledType {
			//TODO Implement ResourceRequest Validation for Helm Installed Devtron APPs
		} else if request.DeploymentType == bean2.ArgoInstalledType {
			//TODO Implement ResourceRequest Validation for ArgoCD Installed APPs From ResourceTree
		}
		// RBAC enforcer applying for Devtron App
		envObject = handler.enforcerUtil.GetEnvRBACNameByAppId(request.DevtronAppIdentifier.AppId, request.DevtronAppIdentifier.EnvId)
		hasReadAccessForEnv := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, actionType, envObject)
		if !hasReadAccessForEnv {
			return false, nil
		}
		// RBAC enforcer Ends
		return true, nil
	case bean2.FluxAppType:
		// For flux app resource
		appIdentifier, err := fluxApplication.DecodeFluxExternalAppId(request.AppId)
		if err != nil {
			handler.logger.Errorw("error in decoding appId", "err", err, "appId", request.AppId)
			return false, err
		}
		//setting fluxAppIdentifier value in request
		request.ExternalFluxAppIdentifier = appIdentifier
		request.ClusterId = appIdentifier.ClusterId
		valid, err := handler.k8sApplicationService.ValidateFluxResourceRequest(r.Context(), request.ExternalFluxAppIdentifier, request.K8sRequest)
		if err != nil || !valid {
			handler.logger.Errorw("error in validating resource request", "err", err)
			return false, err
		}
		//RBAC enforcer starts here
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, actionType, "*"); !ok {
			return false, nil
		}
		return true, nil
		//RBAC enforcer ends here
	default:
		handler.logger.Errorw("appType not recognized", "appType", request.AppType)
		return false, errors.New("appType not founded in request")
	}
}
