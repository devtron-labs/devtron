package k8s

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type K8sCapacityRestHandler interface {
	GetClusterList(w http.ResponseWriter, r *http.Request)
	GetClusterDetail(w http.ResponseWriter, r *http.Request)
	GetNodeList(w http.ResponseWriter, r *http.Request)
	GetNodeDetail(w http.ResponseWriter, r *http.Request)
	UpdateNodeManifest(w http.ResponseWriter, r *http.Request)
}
type K8sCapacityRestHandlerImpl struct {
	logger             *zap.SugaredLogger
	k8sCapacityService K8sCapacityService
	userService        user.UserService
	enforcer           casbin.Enforcer
}

func NewK8sCapacityRestHandlerImpl(logger *zap.SugaredLogger,
	k8sCapacityService K8sCapacityService, userService user.UserService,
	enforcer casbin.Enforcer) *K8sCapacityRestHandlerImpl {
	return &K8sCapacityRestHandlerImpl{
		logger:             logger,
		k8sCapacityService: k8sCapacityService,
		userService:        userService,
		enforcer:           enforcer,
	}
}

func (handler *K8sCapacityRestHandlerImpl) GetClusterList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	clusterDetailList, err := handler.k8sCapacityService.GetClusterCapacityDetailList()
	if err != nil {
		handler.logger.Errorw("error in getting cluster capacity detail list", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, clusterDetailList, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) GetClusterDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	clusterId, err := strconv.Atoi(vars["clusterId"])
	if err != nil {
		handler.logger.Errorw("request err, GetClusterDetail", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	clusterDetail, err := handler.k8sCapacityService.GetClusterCapacityDetailById(clusterId, false)
	if err != nil {
		handler.logger.Errorw("error in getting cluster capacity detail by id", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, clusterDetail, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) GetNodeList(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	clusterId, err := strconv.Atoi(vars.Get("clusterId"))
	if err != nil {
		handler.logger.Errorw("request err, GetNodeList", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	nodeList, err := handler.k8sCapacityService.GetNodeCapacityDetailsListByClusterId(clusterId)
	if err != nil {
		handler.logger.Errorw("error in getting node detail list by clusterId", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nodeList, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) GetNodeDetail(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	clusterId, err := strconv.Atoi(vars.Get("clusterId"))
	if err != nil {
		handler.logger.Errorw("request err, GetNodeDetail", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	name := vars.Get("name")
	if err != nil {
		handler.logger.Errorw("request err, GetNodeDetail", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	nodeDetail, err := handler.k8sCapacityService.GetNodeCapacityDetailByNameAndClusterId(clusterId, name)
	if err != nil {
		handler.logger.Errorw("error in getting node detail by clusterId ", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nodeDetail, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) UpdateNodeManifest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var manifestUpdateReq NodeManifestUpdateDto
	err := decoder.Decode(&manifestUpdateReq)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	updatedManifest, err := handler.k8sCapacityService.UpdateNodeManifest(&manifestUpdateReq)
	if err != nil {
		handler.logger.Errorw("error in updating node manifest", "err", err, "updateRequest", manifestUpdateReq)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, updatedManifest, http.StatusOK)
}
