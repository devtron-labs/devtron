package k8s

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/cluster"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type K8sCapacityRestHandler interface {
	GetClusterListRaw(w http.ResponseWriter, r *http.Request)
	GetClusterListWithDetail(w http.ResponseWriter, r *http.Request)
	GetClusterDetail(w http.ResponseWriter, r *http.Request)
	GetNodeList(w http.ResponseWriter, r *http.Request)
	GetNodeDetail(w http.ResponseWriter, r *http.Request)
	UpdateNodeManifest(w http.ResponseWriter, r *http.Request)
	DeleteNode(w http.ResponseWriter, r *http.Request)
	CordonOrUnCordonNode(w http.ResponseWriter, r *http.Request)
	DrainNode(w http.ResponseWriter, r *http.Request)
	EditNodeTaints(w http.ResponseWriter, r *http.Request)
}
type K8sCapacityRestHandlerImpl struct {
	logger             *zap.SugaredLogger
	k8sCapacityService K8sCapacityService
	userService        user.UserService
	enforcer           casbin.Enforcer
	clusterService     cluster.ClusterService
	environmentService cluster.EnvironmentService
}

func NewK8sCapacityRestHandlerImpl(logger *zap.SugaredLogger,
	k8sCapacityService K8sCapacityService, userService user.UserService,
	enforcer casbin.Enforcer,
	clusterService cluster.ClusterService,
	environmentService cluster.EnvironmentService) *K8sCapacityRestHandlerImpl {
	return &K8sCapacityRestHandlerImpl{
		logger:             logger,
		k8sCapacityService: k8sCapacityService,
		userService:        userService,
		enforcer:           enforcer,
		clusterService:     clusterService,
		environmentService: environmentService,
	}
}

func (handler *K8sCapacityRestHandlerImpl) GetClusterListRaw(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	clusters, err := handler.clusterService.FindAll()
	if err != nil {
		handler.logger.Errorw("error in getting all clusters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying
	var authenticatedClusters []*cluster.ClusterBean
	var clusterDetailList []*ClusterCapacityDetail
	for _, cluster := range clusters {
		authenticated, err := handler.CheckRbacForCluster(cluster, token)
		if err != nil {
			handler.logger.Errorw("error in checking rbac for cluster", "err", err, "clusterId", cluster.Id)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		if authenticated {
			authenticatedClusters = append(authenticatedClusters, cluster)
			clusterDetail := &ClusterCapacityDetail{
				Id:                cluster.Id,
				Name:              cluster.ClusterName,
				ErrorInConnection: cluster.ErrorInConnecting,
			}
			clusterDetailList = append(clusterDetailList, clusterDetail)
		}
	}
	if len(clusters) != 0 && len(clusterDetailList) == 0 {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	common.WriteJsonResp(w, nil, clusterDetailList, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) GetClusterListWithDetail(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	clusters, err := handler.clusterService.FindAll()
	if err != nil {
		handler.logger.Errorw("error in getting all clusters", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	// RBAC enforcer applying
	var authenticatedClusters []*cluster.ClusterBean
	for _, cluster := range clusters {
		authenticated, err := handler.CheckRbacForCluster(cluster, token)
		if err != nil {
			handler.logger.Errorw("error in checking rbac for cluster", "err", err, "clusterId", cluster.Id)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		if authenticated {
			authenticatedClusters = append(authenticatedClusters, cluster)
		}
	}
	if len(authenticatedClusters) == 0 {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	clusterDetailList, err := handler.k8sCapacityService.GetClusterCapacityDetailList(r.Context(), authenticatedClusters)
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
	token := r.Header.Get("token")
	// RBAC enforcer applying
	cluster, err := handler.clusterService.FindById(clusterId)
	if err != nil {
		handler.logger.Errorw("error in getting cluster by id", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	authenticated, err := handler.CheckRbacForCluster(cluster, token)
	if err != nil {
		handler.logger.Errorw("error in checking rbac for cluster", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	clusterDetail, err := handler.k8sCapacityService.GetClusterCapacityDetail(r.Context(), cluster, false)
	if err != nil {
		handler.logger.Errorw("error in getting cluster capacity detail", "err", err, "clusterId", clusterId)
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
	cluster, err := handler.clusterService.FindById(clusterId)
	if err != nil {
		handler.logger.Errorw("error in getting cluster by id", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	authenticated, err := handler.CheckRbacForCluster(cluster, token)
	if err != nil {
		handler.logger.Errorw("error in checking rbac for cluster", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	nodeList, err := handler.k8sCapacityService.GetNodeCapacityDetailsListByCluster(r.Context(), cluster)
	if err != nil {
		handler.logger.Errorw("error in getting node detail list by cluster", "err", err, "clusterId", clusterId)
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
	cluster, err := handler.clusterService.FindById(clusterId)
	if err != nil {
		handler.logger.Errorw("error in getting cluster by id", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	authenticated, err := handler.CheckRbacForCluster(cluster, token)
	if err != nil {
		handler.logger.Errorw("error in checking rbac for cluster", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if !authenticated {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	nodeDetail, err := handler.k8sCapacityService.GetNodeCapacityDetailByNameAndCluster(r.Context(), cluster, name)
	if err != nil {
		handler.logger.Errorw("error in getting node detail by cluster", "err", err, "clusterId", clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nodeDetail, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) UpdateNodeManifest(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var manifestUpdateReq NodeUpdateRequestDto
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
	updatedManifest, err := handler.k8sCapacityService.UpdateNodeManifest(r.Context(), &manifestUpdateReq)
	if err != nil {
		handler.logger.Errorw("error in updating node manifest", "err", err, "updateRequest", manifestUpdateReq)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, updatedManifest, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) DeleteNode(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var nodeDelReq NodeUpdateRequestDto
	err := decoder.Decode(&nodeDelReq)
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
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	updatedManifest, err := handler.k8sCapacityService.DeleteNode(r.Context(), &nodeDelReq)
	if err != nil {
		handler.logger.Errorw("error in deleting node", "err", err, "deleteRequest", nodeDelReq)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, updatedManifest, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) CordonOrUnCordonNode(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var nodeCordonReq NodeUpdateRequestDto
	err := decoder.Decode(&nodeCordonReq)
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
	resp, err := handler.k8sCapacityService.CordonOrUnCordonNode(r.Context(), &nodeCordonReq)
	if err != nil {
		handler.logger.Errorw("error in cordon/unCordon node", "err", err, "req", nodeCordonReq)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) DrainNode(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var nodeDrainReq NodeUpdateRequestDto
	err := decoder.Decode(&nodeDrainReq)
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
	resp, err := handler.k8sCapacityService.DrainNode(r.Context(), &nodeDrainReq)
	if err != nil {
		handler.logger.Errorw("error in draining node", "err", err, "req", nodeDrainReq)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) EditNodeTaints(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var nodeTaintReq NodeUpdateRequestDto
	err := decoder.Decode(&nodeTaintReq)
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
	resp, err := handler.k8sCapacityService.EditNodeTaints(r.Context(), &nodeTaintReq)
	if err != nil {
		handler.logger.Errorw("error in editing node taints", "err", err, "req", nodeTaintReq)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}

func (handler *K8sCapacityRestHandlerImpl) CheckRbacForCluster(cluster *cluster.ClusterBean, token string) (authenticated bool, err error) {
	//getting all environments for this cluster
	envs, err := handler.environmentService.GetByClusterId(cluster.Id)
	if err != nil {
		handler.logger.Errorw("error in getting environments by clusterId", "err", err, "clusterId", cluster.Id)
		return false, err
	}
	if len(envs) == 0 {
		if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
			return false, nil
		}
		return true, nil
	}
	emailId, err := handler.userService.GetEmailFromToken(token)
	if err != nil {
		handler.logger.Errorw("error in getting emailId from token", "err", err)
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
	rbacResultMap := handler.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceGlobalEnvironment, casbin.ActionGet, envIdentifierList)
	for envIdentifier, _ := range envIdentifierMap {
		if rbacResultMap[envIdentifier] {
			//if user has view permission to even one environment of this cluster, authorise the request
			return true, nil
		}
	}
	return false, nil
}
