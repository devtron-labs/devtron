package client

import (
	"context"
	"encoding/json"
	"errors"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type HelmAppRestHandler interface {
	ListApplications(w http.ResponseWriter, r *http.Request)
	GetApplicationDetail(w http.ResponseWriter, r *http.Request)
	Hibernate(w http.ResponseWriter, r *http.Request)
	UnHibernate(w http.ResponseWriter, r *http.Request)
	GetDeploymentHistory(w http.ResponseWriter, r *http.Request)
	GetValuesYaml(w http.ResponseWriter, r *http.Request)
	GetDesiredManifest(w http.ResponseWriter, r *http.Request)
}
type HelmAppRestHandlerImpl struct {
	logger         *zap.SugaredLogger
	helmAppService HelmAppService
	enforcer       casbin.Enforcer
	clusterService cluster.ClusterService
	enforcerUtil   rbac.EnforcerUtilHelm
}

func NewHelmAppRestHandlerImpl(logger *zap.SugaredLogger,
	helmAppService HelmAppService, enforcer casbin.Enforcer,
	clusterService cluster.ClusterService, enforcerUtil rbac.EnforcerUtilHelm) *HelmAppRestHandlerImpl {
	return &HelmAppRestHandlerImpl{
		logger:         logger,
		helmAppService: helmAppService,
		enforcer:       enforcer,
		clusterService: clusterService,
		enforcerUtil:   enforcerUtil,
	}
}

func (handler *HelmAppRestHandlerImpl) ListApplications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["clusterIds"]
	clusterIdSlices := strings.Split(clusterIdString, ",")
	var clusterIds []int
	for _, is := range clusterIdSlices {
		if len(is) == 0 {
			continue
		}
		j, err := strconv.Atoi(is)
		if err != nil {
			handler.logger.Errorw("request err, CreateUser", "err", err, "payload", clusterIds)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		clusterIds = append(clusterIds, j)
	}
	token := r.Header.Get("token")
	handler.helmAppService.ListHelmApplications(clusterIds, w, token, handler.CheckHelmAuth)
}

func (handler *HelmAppRestHandlerImpl) GetApplicationDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["appId"]

	appIdentifier, err := handler.helmAppService.DecodeAppId(clusterIdString)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	appdetail, err := handler.helmAppService.GetApplicationDetail(context.Background(), appIdentifier)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, appdetail, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) Hibernate(w http.ResponseWriter, r *http.Request) {
	hibernateRequest := &openapi.HibernateRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(hibernateRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appIdentifier, err := handler.helmAppService.DecodeAppId(*hibernateRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.HibernateApplication(context.Background(), appIdentifier, hibernateRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) UnHibernate(w http.ResponseWriter, r *http.Request) {
	var hibernateRequest *openapi.HibernateRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(hibernateRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appIdentifier, err := handler.helmAppService.DecodeAppId(*hibernateRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.UnHibernateApplication(context.Background(), appIdentifier, hibernateRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) GetDeploymentHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId := vars["appId"]
	appIdentifier, err := handler.helmAppService.DecodeAppId(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.GetDeploymentHistory(context.Background(), appIdentifier)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) GetValuesYaml(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId := vars["appId"]
	appIdentifier, err := handler.helmAppService.DecodeAppId(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.GetValuesYaml(context.Background(), appIdentifier)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) GetDesiredManifest(w http.ResponseWriter, r *http.Request) {
	desiredManifestRequest := &openapi.DesiredManifestRequest{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(desiredManifestRequest)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	appIdentifier, err := handler.helmAppService.DecodeAppId(*desiredManifestRequest.AppId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	rbacObject := handler.enforcerUtil.GetHelmObjectByClusterId(appIdentifier.ClusterId, appIdentifier.Namespace, appIdentifier.ReleaseName)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, rbacObject); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	res, err := handler.helmAppService.GetDesiredManifest(context.Background(), appIdentifier, desiredManifestRequest.Resource)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *HelmAppRestHandlerImpl) CheckHelmAuth(token string, object string) bool {
	if ok := handler.enforcer.Enforce(token, casbin.ResourceHelmApp, casbin.ActionGet, strings.ToLower(object)); !ok {
		return false
	}
	return true
}
