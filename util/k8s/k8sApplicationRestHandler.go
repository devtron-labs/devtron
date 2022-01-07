package k8s

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/connector"
	client "github.com/devtron-labs/devtron/api/helm-app"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/terminal"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	errors2 "github.com/juju/errors"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strconv"
)

type K8sApplicationRestHandler interface {
	GetResource(w http.ResponseWriter, r *http.Request)
	UpdateResource(w http.ResponseWriter, r *http.Request)
	DeleteResource(w http.ResponseWriter, r *http.Request)
	ListEvents(w http.ResponseWriter, r *http.Request)
	GetPodLogs(w http.ResponseWriter, r *http.Request)
	GetTerminalSession(w http.ResponseWriter, r *http.Request)
}
type K8sApplicationRestHandlerImpl struct {
	logger                 *zap.SugaredLogger
	k8sApplicationService  K8sApplicationService
	pump                   connector.Pump
	terminalSessionHandler terminal.TerminalSessionHandler
	enforcer               casbin.Enforcer
	enforcerUtil           rbac.EnforcerUtilHelm
	clusterService         cluster.ClusterService
	helmAppService         client.HelmAppService
}

func NewK8sApplicationRestHandlerImpl(logger *zap.SugaredLogger,
	k8sApplicationService K8sApplicationService, pump connector.Pump,
	terminalSessionHandler terminal.TerminalSessionHandler,
	enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtilHelm, clusterService cluster.ClusterService,
	 helmAppService client.HelmAppService) *K8sApplicationRestHandlerImpl {
	return &K8sApplicationRestHandlerImpl{
		logger:                 logger,
		k8sApplicationService:  k8sApplicationService,
		pump:                   pump,
		terminalSessionHandler: terminalSessionHandler,
		enforcer:               enforcer,
		enforcerUtil:           enforcerUtil,
		helmAppService:         helmAppService,
		clusterService:         clusterService,
	}
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
	valid, err := handler.k8sApplicationService.ValidateResourceRequest(request.AppIdentifier, request.K8sRequest)
	if err != nil || !valid {
		handler.logger.Errorw("error in validating resource request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmAppWF, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends

	resource, err := handler.k8sApplicationService.GetResource(&request)
	if err != nil {
		handler.logger.Errorw("error in getting resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) UpdateResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	valid, err := handler.k8sApplicationService.ValidateResourceRequest(request.AppIdentifier, request.K8sRequest)
	if err != nil || !valid {
		handler.logger.Errorw("error in validating resource request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmAppWF, casbin.ActionUpdate, rbacObject); !ok {
		common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	resource, err := handler.k8sApplicationService.UpdateResource(&request)
	if err != nil {
		handler.logger.Errorw("error in updating resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) DeleteResource(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	valid, err := handler.k8sApplicationService.ValidateResourceRequest(request.AppIdentifier, request.K8sRequest)
	if err != nil || !valid {
		handler.logger.Errorw("error in validating resource request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmAppWF, casbin.ActionDelete, rbacObject); !ok {
		common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	resource, err := handler.k8sApplicationService.DeleteResource(&request)
	if err != nil {
		handler.logger.Errorw("error in deleting resource", "err", err)
		common.WriteJsonResp(w, err, resource, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resource, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) ListEvents(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	valid, err := handler.k8sApplicationService.ValidateResourceRequest(request.AppIdentifier, request.K8sRequest)
	if err != nil || !valid {
		handler.logger.Errorw("error in validating resource request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmAppWF, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	events, err := handler.k8sApplicationService.ListEvents(&request)
	if err != nil {
		handler.logger.Errorw("error in getting events list", "err", err)
		common.WriteJsonResp(w, err, events, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, events, http.StatusOK)
}

func (handler *K8sApplicationRestHandlerImpl) GetPodLogs(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request ResourceRequestBean
	err := decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	valid, err := handler.k8sApplicationService.ValidateResourceRequest(request.AppIdentifier, request.K8sRequest)
	if err != nil || !valid {
		handler.logger.Errorw("error in validating resource request", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(request.AppIdentifier.ClusterId, request.AppIdentifier.Namespace, request.AppIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmAppWF, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors2.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
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
	stream, err := handler.k8sApplicationService.GetPodLogs(&request)
	defer util.Close(stream, handler.logger)
	handler.pump.StartK8sStreamWithHeartBeat(w, isReconnect, stream, err)
}

func (handler *K8sApplicationRestHandlerImpl) GetTerminalSession(w http.ResponseWriter, r *http.Request) {
	request := &terminal.TerminalSessionRequest{}
	vars := mux.Vars(r)
	request.ContainerName = vars["container"]
	request.Namespace = vars["namespace"]
	request.PodName = vars["pod"]
	request.Shell = vars["shell"]
	request.ApplicationId = vars["applicationId"]

	//---------auth
	app, err := handler.helmAppService.DecodeAppId(request.ApplicationId)
	if err != nil {
		handler.logger.Errorw("invalid app id", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.ClusterId = app.ClusterId
	//---------auth end

	status, message, err := handler.terminalSessionHandler.GetTerminalSession(request)
	common.WriteJsonResp(w, err, message, status)
}
