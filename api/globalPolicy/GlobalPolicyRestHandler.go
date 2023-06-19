package globalPolicy

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/globalPolicy"
	"github.com/devtron-labs/devtron/pkg/globalPolicy/bean"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
	"strings"
)

type GlobalPolicyRestHandler interface {
	GetById(w http.ResponseWriter, r *http.Request)
	GetAllGlobalPolicies(w http.ResponseWriter, r *http.Request)
	CreateGlobalPolicy(w http.ResponseWriter, r *http.Request)
	UpdateGlobalPolicy(w http.ResponseWriter, r *http.Request)
	DeleteGlobalPolicy(w http.ResponseWriter, r *http.Request)
	GetPolicyOffendingPipelinesWfTree(w http.ResponseWriter, r *http.Request)
	GetOnlyBlockageStateOfACiPipeline(w http.ResponseWriter, r *http.Request)
	GetMandatoryPluginsForACiPipeline(w http.ResponseWriter, r *http.Request)
}

type GlobalPolicyRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	globalPolicyService globalPolicy.GlobalPolicyService
	userService         user.UserService
	enforcer            casbin.Enforcer
	validator           *validator.Validate
	enforcerUtil        rbac.EnforcerUtil
}

func NewGlobalPolicyRestHandlerImpl(logger *zap.SugaredLogger,
	globalPolicyService globalPolicy.GlobalPolicyService, userService user.UserService,
	enforcer casbin.Enforcer, validator *validator.Validate,
	enforcerUtil rbac.EnforcerUtil) *GlobalPolicyRestHandlerImpl {
	return &GlobalPolicyRestHandlerImpl{
		logger:              logger,
		globalPolicyService: globalPolicyService,
		userService:         userService,
		enforcer:            enforcer,
		validator:           validator,
		enforcerUtil:        enforcerUtil,
	}
}

func (handler *GlobalPolicyRestHandlerImpl) GetById(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, GetById", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	policy, err := handler.globalPolicyService.GetById(id)
	if err != nil {
		handler.logger.Errorw("service error, GetById", "err", err, "policyId", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, policy, http.StatusOK)
}

func (handler *GlobalPolicyRestHandlerImpl) GetAllGlobalPolicies(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	policyOf := r.URL.Query().Get("policyOf")
	policyVersion := r.URL.Query().Get("policyVersion")
	policies, err := handler.globalPolicyService.GetAllGlobalPolicies(bean.GlobalPolicyType(policyOf), bean.GlobalPolicyVersion(policyVersion))
	if err != nil {
		handler.logger.Errorw("service error, GetAllGlobalPolicies", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, policies, http.StatusOK)
}

func (handler *GlobalPolicyRestHandlerImpl) CreateGlobalPolicy(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	policyOf := r.URL.Query().Get("policyOf")
	policyVersion := r.URL.Query().Get("policyVersion")

	var request bean.GlobalPolicyDto
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, CreateOrUpdateGlobalPolicy", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.PolicyOf = bean.GlobalPolicyType(policyOf)
	request.PolicyVersion = bean.GlobalPolicyVersion(policyVersion)
	request.UserId = userId
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err, CreateOrUpdateGlobalPolicy", "err", err, "payload", request)
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "data validation error", InternalMessage: err.Error()}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.globalPolicyService.CreateOrUpdateGlobalPolicy(&request)
	if err != nil {
		handler.logger.Errorw("service error, CreateOrUpdateGlobalPolicy", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "Policy created successfully.", http.StatusOK)
}

func (handler *GlobalPolicyRestHandlerImpl) UpdateGlobalPolicy(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	policyOf := r.URL.Query().Get("policyOf")
	policyVersion := r.URL.Query().Get("policyVersion")

	var request bean.GlobalPolicyDto
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("request err, UpdateGlobalPolicy", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	request.PolicyOf = bean.GlobalPolicyType(policyOf)
	request.PolicyVersion = bean.GlobalPolicyVersion(policyVersion)
	request.UserId = userId

	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err, UpdateGlobalPolicy", "err", err, "payload", request)
		err = &util.ApiError{Code: "400", HttpStatusCode: 400, UserMessage: "data validation error", InternalMessage: err.Error()}
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = handler.globalPolicyService.CreateOrUpdateGlobalPolicy(&request)
	if err != nil {
		handler.logger.Errorw("service error, UpdateGlobalPolicy", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "Policy updated successfully.", http.StatusOK)

}

func (handler *GlobalPolicyRestHandlerImpl) DeleteGlobalPolicy(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, DeleteGlobalPolicy", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.globalPolicyService.DeleteGlobalPolicy(id, userId)
	if err != nil {
		handler.logger.Errorw("service error, DeleteGlobalPolicy", "err", err, "policyId", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "Policy deleted successfully.", http.StatusOK)
}

func (handler *GlobalPolicyRestHandlerImpl) GetPolicyOffendingPipelinesWfTree(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	policyIdStr := r.URL.Query().Get("policyId")
	policyId := 0
	if len(policyIdStr) > 0 {
		policyId, err = strconv.Atoi(policyIdStr)
		if err != nil {
			handler.logger.Errorw("request err, GetMandatoryPluginsForACiPipeline", "err", err, "policyId", policyId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	wfTrees, err := handler.globalPolicyService.GetPolicyOffendingPipelinesWfTree(policyId)
	if err != nil {
		handler.logger.Errorw("service error, GetPolicyOffendingPipelinesWfTree", "err", err, "policyId", policyId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, wfTrees, http.StatusOK)
}

func (handler *GlobalPolicyRestHandlerImpl) GetOnlyBlockageStateOfACiPipeline(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	ciPipelineIdStr := r.URL.Query().Get("ciPipelineId")
	appIdStr := r.URL.Query().Get("appId")
	branchValuesStr := r.URL.Query().Get("branchValues")
	branchValues := strings.Split(branchValuesStr, ",")
	ciPipelineId := 0
	appId := 0
	if len(ciPipelineIdStr) > 0 {
		ciPipelineId, err = strconv.Atoi(ciPipelineIdStr)
		if err != nil {
			handler.logger.Errorw("request err, GetOnlyBlockageStateOfACiPipeline", "err", err, "ciPipelineId", ciPipelineId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	if len(appIdStr) > 0 {
		appId, err = strconv.Atoi(appIdStr)
		if err != nil {
			handler.logger.Errorw("request err, GetOnlyBlockageStateOfACiPipeline", "err", err, "appId", appId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	token := r.Header.Get("token")
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	isOffendingMandatoryPlugin, isCIPipelineTriggerBlocked, blockageState, err := handler.globalPolicyService.GetOnlyBlockageStateForCiPipeline(ciPipelineId, branchValues)
	if err != nil {
		handler.logger.Errorw("service error, GetOnlyBlockageStateOfACiPipeline", "err", err, "ciPipelineId", ciPipelineId, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	responseMap := make(map[string]interface{})
	responseMap["isOffendingMandatoryPlugin"] = isOffendingMandatoryPlugin
	responseMap["isCITriggerBlocked"] = isCIPipelineTriggerBlocked
	responseMap["ciBlockState"] = blockageState
	common.WriteJsonResp(w, nil, responseMap, http.StatusOK)
}

func (handler *GlobalPolicyRestHandlerImpl) GetMandatoryPluginsForACiPipeline(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusUnauthorized)
		return
	}
	ciPipelineIdStr := r.URL.Query().Get("ciPipelineId")
	appIdStr := r.URL.Query().Get("appId")
	branchValuesStr := r.URL.Query().Get("branchValues")
	branchValues := strings.Split(branchValuesStr, ",")
	ciPipelineId := 0
	appId := 0
	if len(ciPipelineIdStr) > 0 {
		ciPipelineId, err = strconv.Atoi(ciPipelineIdStr)
		if err != nil {
			handler.logger.Errorw("request err, GetMandatoryPluginsForACiPipeline", "err", err, "ciPipelineId", ciPipelineId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	if len(appIdStr) > 0 {
		appId, err = strconv.Atoi(appIdStr)
		if err != nil {
			handler.logger.Errorw("request err, GetMandatoryPluginsForACiPipeline", "err", err, "appId", appId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	token := r.Header.Get("token")
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	mandatoryPlugins, _, err := handler.globalPolicyService.GetMandatoryPluginsForACiPipeline(ciPipelineId, appId, branchValues, false)
	if err != nil {
		handler.logger.Errorw("service error, GetMandatoryPluginsForACiPipeline", "err", err, "ciPipelineId", ciPipelineId, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, mandatoryPlugins, http.StatusOK)
}
