package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/connector"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/client/k8s/application"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/k8sObjectsUtil"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	errors2 "github.com/juju/errors"
	"go.uber.org/zap"
	errors3 "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"net/http"
	"strconv"
	"strings"
)

type K8sApplicationRestHandler interface {
	GetResource(w http.ResponseWriter, r *http.Request)
	CreateResource(w http.ResponseWriter, r *http.Request)
	UpdateResource(w http.ResponseWriter, r *http.Request)
	DeleteResource(w http.ResponseWriter, r *http.Request)
	ListEvents(w http.ResponseWriter, r *http.Request)
	GetPodLogs(w http.ResponseWriter, r *http.Request)
	GetTerminalSession(w http.ResponseWriter, r *http.Request)
	GetResourceInfo(w http.ResponseWriter, r *http.Request)
	GetHostUrlsByBatch(w http.ResponseWriter, r *http.Request)
	GetAllApiResources(w http.ResponseWriter, r *http.Request)
	GetResourceList(w http.ResponseWriter, r *http.Request)
	ApplyResources(w http.ResponseWriter, r *http.Request)
	RotatePod(w http.ResponseWriter, r *http.Request)
}

type K8sApplicationRestHandlerImpl struct {
	logger                 *zap.SugaredLogger
	k8sApplicationService  K8sApplicationService
	pump                   connector.Pump
	terminalSessionHandler terminal.TerminalSessionHandler
	enforcer               casbin.Enforcer
	enforcerUtil           rbac.EnforcerUtil
	enforcerUtilHelm       rbac.EnforcerUtilHelm
	helmAppService         client.HelmAppService
	userService            user.UserService
}

func NewK8sApplicationRestHandlerImpl(logger *zap.SugaredLogger,
	k8sApplicationService K8sApplicationService, pump connector.Pump,
	terminalSessionHandler terminal.TerminalSessionHandler,
	enforcer casbin.Enforcer, enforcerUtilHelm rbac.EnforcerUtilHelm, enforcerUtil rbac.EnforcerUtil,
	helmAppService client.HelmAppService, userService user.UserService) *K8sApplicationRestHandlerImpl {
	return &K8sApplicationRestHandlerImpl{
		logger:                 logger,
		k8sApplicationService:  k8sApplicationService,
		pump:                   pump,
		terminalSessionHandler: terminalSessionHandler,
		enforcer:               enforcer,
		enforcerUtilHelm:       enforcerUtilHelm,
		enforcerUtil:           enforcerUtil,
		helmAppService:         helmAppService,
		userService:            userService,
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
	podRotateRequest := &RotatePodRequest{}
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
	rotatePodRequest := &RotatePodRequest{
		ClusterId: appIdentifier.ClusterId,
		Resources: podRotateRequest.Resources,
	}
	response, err := handler.k8sApplicationService.RotatePods(r.Context(), rotatePodRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) GetResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	rbacObject := ""
	rbacObject2 := ""

	token := r.Header.Get("token")
	if request.AppId != "" {
		appIdentifier, err := handler.helmAppService.DecodeAppId(request.AppId)
		if err != nil {
			handler.logger.Errorw("error in decoding appId", "err", err, "appId", request.AppId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		//setting appIdentifier value in request
		request.AppIdentifier = appIdentifier
		request.ClusterId = request.AppIdentifier.ClusterId
		valid, err := handler.k8sApplicationService.ValidateResourceRequest(r.Context(), request.AppIdentifier, request.K8sRequest)
		if err != nil || !valid {
			handler.logger.Errorw("error in validating resource request", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		// TODO: this rbac is commented because we are only checking helm apps access whereas this api is being used in devtron apps too
		// this  needs to be updated with conditional rbac depending on where the call came from,until then this will get prevented with the view page permission
		//rbacObject, rbacObject2 = handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
		//token := r.Header.Get("token")
		//ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)
		//if !ok {
		//	common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
		//	return
		//}
	} else if request.ClusterId <= 0 {
		common.WriteJsonResp(w, errors.New("can not resource manifest as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}

	resource, err := handler.k8sApplicationService.GetResource(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in getting resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}

	canUpdate := true
	if request.AppId != "" {
		// Obfuscate secret if user does not have edit access
		canUpdate = handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject2)
	} else if request.ClusterId > 0 {
		canUpdate = handler.k8sApplicationService.ValidateClusterResourceBean(r.Context(), request.ClusterId, resource.Manifest, request.K8sRequest.ResourceIdentifier.GroupVersionKind, handler.getRbacCallbackForResource(token, casbin.ActionUpdate))
		if !canUpdate {
			readAllowed := handler.k8sApplicationService.ValidateClusterResourceBean(r.Context(), request.ClusterId, resource.Manifest, request.K8sRequest.ResourceIdentifier.GroupVersionKind, handler.getRbacCallbackForResource(token, casbin.ActionGet))
			if !readAllowed {
				common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
				return
			}
		}
	}
	if !canUpdate && resource != nil {
		modifiedManifest, err := k8sObjectsUtil.HideValuesIfSecret(&resource.Manifest)
		if err != nil {
			handler.logger.Errorw("error in hiding secret values", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		resource.Manifest = *modifiedManifest
	}

	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) GetHostUrlsByBatch(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["appId"]
	if clusterIdString == "" {
		common.WriteJsonResp(w, fmt.Errorf("empty appid in request"), nil, http.StatusBadRequest)
		return
	}
	appIdentifier, err := handler.helmAppService.DecodeAppId(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")

	ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)

	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	appDetail, err := handler.helmAppService.GetApplicationDetail(r.Context(), appIdentifier)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	k8sAppDetail := bean.AppDetailContainer{
		DeploymentDetailContainer: bean.DeploymentDetailContainer{
			ClusterId: appIdentifier.ClusterId,
			Namespace: appIdentifier.Namespace,
		},
	}
	validRequests := make([]ResourceRequestBean, 0)
	var resourceTreeInf map[string]interface{}
	bytes, _ := json.Marshal(appDetail.ResourceTreeResponse)
	err = json.Unmarshal(bytes, &resourceTreeInf)
	if err != nil {
		common.WriteJsonResp(w, fmt.Errorf("unmarshal error of resource tree response"), nil, http.StatusInternalServerError)
		return
	}
	validRequests = handler.k8sApplicationService.FilterServiceAndIngress(r.Context(), resourceTreeInf, validRequests, k8sAppDetail, clusterIdString)
	if len(validRequests) == 0 {
		handler.logger.Error("neither service nor ingress found for this app", "appId", clusterIdString)
		common.WriteJsonResp(w, err, nil, http.StatusNoContent)
		return
	}

	resp, err := handler.k8sApplicationService.GetManifestsByBatch(r.Context(), validRequests)
	if err != nil {
		handler.logger.Errorw("error in getting manifests in batch", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	result := handler.k8sApplicationService.GetUrlsByBatch(r.Context(), resp)
	common.WriteJsonResp(w, nil, result, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) CreateResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request ResourceRequestBean
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
	resource, err := handler.k8sApplicationService.CreateResource(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in creating resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) UpdateResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	token := r.Header.Get("token")
	var request ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if len(request.AppId) > 0 {
		// assume it as helm release case in which appId is supplied
		appIdentifier, err := handler.helmAppService.DecodeAppId(request.AppId)
		if err != nil {
			handler.logger.Errorw("error in decoding appId", "err", err, "appId", request.AppId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		//setting appIdentifier value in request
		request.AppIdentifier = appIdentifier
		request.ClusterId = appIdentifier.ClusterId
		valid, err := handler.k8sApplicationService.ValidateResourceRequest(r.Context(), request.AppIdentifier, request.K8sRequest)
		if err != nil || !valid {
			handler.logger.Errorw("error in validating resource request", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		// RBAC enforcer applying
		rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
		token := r.Header.Get("token")
		ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionUpdate, rbacObject2)
		if !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends
	} else if request.ClusterId > 0 {
		// assume direct update in cluster
		if ok := handler.handleRbac(r, w, request, token, casbin.ActionUpdate); !ok {
			return
		}
	} else {
		common.WriteJsonResp(w, errors.New("can not update resource as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}

	resource, err := handler.k8sApplicationService.UpdateResource(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in updating resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) handleRbac(r *http.Request, w http.ResponseWriter, request ResourceRequestBean, token string, casbinAction string) bool {
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

	decoder := json.NewDecoder(r.Body)
	token := r.Header.Get("token")
	var request ResourceRequestBean
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	if len(request.AppId) > 0 {
		// assume it as helm release case in which appId is supplied
		appIdentifier, err := handler.helmAppService.DecodeAppId(request.AppId)
		if err != nil {
			handler.logger.Errorw("error in decoding appId", "err", err, "appId", request.AppId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		//setting appIdentifier value in request
		request.AppIdentifier = appIdentifier
		request.ClusterId = appIdentifier.ClusterId
		valid, err := handler.k8sApplicationService.ValidateResourceRequest(r.Context(), request.AppIdentifier, request.K8sRequest)
		if err != nil || !valid {
			handler.logger.Errorw("error in validating resource request", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		// RBAC enforcer applying
		rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
		token := r.Header.Get("token")

		ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionDelete, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionDelete, rbacObject2)

		if !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends
	} else if request.ClusterId > 0 {
		if ok := handler.handleRbac(r, w, request, token, casbin.ActionDelete); !ok {
			return
		}
	} else {
		common.WriteJsonResp(w, errors.New("can not delete resource as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}

	resource, err := handler.k8sApplicationService.DeleteResource(r.Context(), &request, userId)
	if err != nil {
		handler.logger.Errorw("error in deleting resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) ListEvents(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	token := r.Header.Get("token")
	var request ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	if len(request.AppId) > 0 {
		// assume it as helm release case in which appId is supplied
		appIdentifier, err := handler.helmAppService.DecodeAppId(request.AppId)
		if err != nil {
			handler.logger.Errorw("error in decoding appId", "err", err, "appId", request.AppId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		//setting appIdentifier value in request
		request.AppIdentifier = appIdentifier
		request.ClusterId = appIdentifier.ClusterId
		valid, err := handler.k8sApplicationService.ValidateResourceRequest(r.Context(), request.AppIdentifier, request.K8sRequest)
		if err != nil || !valid {
			handler.logger.Errorw("error in validating resource request", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		// TODO: this rbac is commented because we are only checking helm apps access whereas this api is being used in devtron apps too
		// this  needs to be updated with conditional rbac depending on where the call came from,until then this will get prevented with the view page permission
		//rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
		//token := r.Header.Get("token")
		//
		//ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)
		//
		//if !ok {
		//	common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
		//	return
		//}
		//RBAC enforcer Ends
	} else if request.ClusterId > 0 {
		if ok := handler.handleRbac(r, w, request, token, casbin.ActionGet); !ok {
			return
		}
	} else {
		common.WriteJsonResp(w, errors.New("can not get resource as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}
	events, err := handler.k8sApplicationService.ListEvents(r.Context(), &request)
	if err != nil {
		handler.logger.Errorw("error in getting events list", "err", err)
		common.WriteJsonResp(w, err, events, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, events, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) GetPodLogs(w http.ResponseWriter, r *http.Request) {
	v := r.URL.Query()
	vars := mux.Vars(r)
	podName := vars["podName"]
	containerName := v.Get("containerName")
	appId := v.Get("appId")
	clusterIdString := v.Get("clusterId")
	namespace := v.Get("namespace")
	/*sinceSeconds, err := strconv.Atoi(v.Get("sinceSeconds"))
	if err != nil {
		sinceSeconds = 0
	}*/
	token := r.Header.Get("token")
	follow, err := strconv.ParseBool(v.Get("follow"))
	if err != nil {
		follow = false
	}
	tailLines, err := strconv.Atoi(v.Get("tailLines"))
	if err != nil {
		tailLines = 0
	}
	var request *ResourceRequestBean
	if appId != "" {
		appIdentifier, err := handler.helmAppService.DecodeAppId(appId)
		if err != nil {
			handler.logger.Errorw("error in decoding appId", "err", err, "appId", appId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		request = &ResourceRequestBean{
			AppIdentifier: appIdentifier,
			ClusterId:     appIdentifier.ClusterId,
			K8sRequest: &application.K8sRequestBean{
				ResourceIdentifier: application.ResourceIdentifier{
					Name:             podName,
					Namespace:        appIdentifier.Namespace,
					GroupVersionKind: schema.GroupVersionKind{},
				},
				PodLogsRequest: application.PodLogsRequest{
					//SinceTime:     sinceSeconds,
					TailLines:     tailLines,
					Follow:        follow,
					ContainerName: containerName,
				},
			},
		}

		valid, err := handler.k8sApplicationService.ValidateResourceRequest(r.Context(), request.AppIdentifier, request.K8sRequest)
		if err != nil || !valid {
			handler.logger.Errorw("error in validating resource request", "err", err)
			apiError := util2.ApiError{
				InternalMessage: "failed to validate the resource with error " + err.Error(),
				UserMessage:     "Failed to validate resource",
			}
			if !valid {
				apiError.InternalMessage = "failed to validate the resource"
				apiError.UserMessage = "requested Pod or Container doesn't exist"
			}
			common.WriteJsonResp(w, &apiError, nil, http.StatusBadRequest)
			return
		}
		// TODO: this rbac is commented because we are only checking helm apps access whereas this api is being used in devtron apps too
		// this  needs to be updated with conditional rbac depending on where the call came from,until then this will get prevented with the view page permission
		//rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
		//token := r.Header.Get("token")
		//
		//ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)
		//
		//if !ok {
		//	common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
		//	return
		//}
		//RBAC enforcer Ends
	} else if clusterIdString != "" && namespace != "" {
		clusterId, err := strconv.Atoi(clusterIdString)
		if err != nil {
			handler.logger.Errorw("invalid cluster id", "clusterId", clusterIdString, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		request = &ResourceRequestBean{
			ClusterId: clusterId,
			K8sRequest: &application.K8sRequestBean{
				ResourceIdentifier: application.ResourceIdentifier{
					Name:      podName,
					Namespace: namespace,
					GroupVersionKind: schema.GroupVersionKind{
						Group:   "",
						Kind:    "Pod",
						Version: "v1",
					},
				},
				PodLogsRequest: application.PodLogsRequest{
					//SinceTime:     sinceSeconds,
					TailLines:     tailLines,
					Follow:        follow,
					ContainerName: containerName,
				},
			},
		}
		if ok := handler.handleRbac(r, w, *request, token, casbin.ActionGet); !ok {
			return
		}
	} else {
		common.WriteJsonResp(w, errors.New("can not get pod logs as target cluster or namespace is not provided"), nil, http.StatusBadRequest)
		return
	}

	lastEventId := r.Header.Get("Last-Event-ID")
	isReconnect := false
	if len(lastEventId) > 0 {
		lastSeenMsgId, err := strconv.ParseInt(lastEventId, 10, 64)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		lastSeenMsgId = lastSeenMsgId + 1 //increased by one ns to avoid duplicate
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

func (handler *K8sApplicationRestHandlerImpl) GetTerminalSession(w http.ResponseWriter, r *http.Request) {
	request := &terminal.TerminalSessionRequest{}
	vars := mux.Vars(r)
	token := r.Header.Get("token")
	request.ContainerName = vars["container"]
	request.Namespace = vars["namespace"]
	request.PodName = vars["pod"]
	request.Shell = vars["shell"]
	clusterIdString := ""
	appId := ""
	identifier := vars["identifier"]
	if strings.Contains(identifier, "|") {
		appId = identifier
	} else {
		clusterIdString = identifier
	}

	if appId != "" {
		request.ApplicationId = appId
		app, err := handler.helmAppService.DecodeAppId(request.ApplicationId)
		if err != nil {
			handler.logger.Errorw("invalid app id", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		request.ClusterId = app.ClusterId

		// RBAC enforcer applying
		rbacObject, rbacObject2 := handler.enforcerUtilHelm.GetHelmObjectByClusterIdNamespaceAndAppName(app.ClusterId, app.Namespace, app.ReleaseName)
		token := r.Header.Get("token")

		ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject) || handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject2)

		if !ok {
			common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
		//RBAC enforcer Ends
	} else if clusterIdString != "" {
		clusterId, err := strconv.Atoi(clusterIdString)
		if err != nil {
			handler.logger.Errorw("invalid cluster id", "clusterId", clusterIdString, "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		request.ClusterId = clusterId
		resourceRequestBean := ResourceRequestBean{
			ClusterId: clusterId,
			K8sRequest: &application.K8sRequestBean{
				ResourceIdentifier: application.ResourceIdentifier{
					Name:      request.PodName,
					Namespace: request.Namespace,
					GroupVersionKind: schema.GroupVersionKind{
						Group:   "",
						Kind:    "Pod",
						Version: "v1",
					},
				},
			},
		}
		if ok := handler.handleRbac(r, w, resourceRequestBean, token, casbin.ActionUpdate); !ok {
			return
		}
	} else {
		common.WriteJsonResp(w, errors.New("can not get terminal session as target cluster is not provided"), nil, http.StatusBadRequest)
		return
	}

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
	var request ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	response, err := handler.k8sApplicationService.GetResourceList(r.Context(), token, &request, handler.verifyRbacForCluster)
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
	var request application.ApplyResourcesRequest
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

func (handler *K8sApplicationRestHandlerImpl) getRbacCallbackForResource(token string, casbinAction string) func(clusterName string, resourceIdentifier application.ResourceIdentifier) bool {
	return func(clusterName string, resourceIdentifier application.ResourceIdentifier) bool {
		return handler.verifyRbacForResource(token, clusterName, resourceIdentifier, casbinAction)
	}
}

func (handler *K8sApplicationRestHandlerImpl) verifyRbacForResource(token string, clusterName string, resourceIdentifier application.ResourceIdentifier, casbinAction string) bool {
	resourceName, objectName := handler.enforcerUtil.GetRBACNameForClusterEntity(clusterName, resourceIdentifier)
	return handler.enforcer.Enforce(token, strings.ToLower(resourceName), casbinAction, strings.ToLower(objectName))
}

func (handler *K8sApplicationRestHandlerImpl) verifyRbacForCluster(token string, clusterName string, request ResourceRequestBean, casbinAction string) bool {
	k8sRequest := request.K8sRequest
	return handler.verifyRbacForResource(token, clusterName, k8sRequest.ResourceIdentifier, casbinAction)
}
