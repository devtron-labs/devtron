package commonPolicyActions

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/policyGovernance"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type ListResponse struct {
	TotalCount                   int                                      `json:"totalCount"`
	AppEnvironmentPolicyMappings []policyGovernance.AppEnvPolicyContainer `json:"appEnvironmentPolicyMappings"`
}

type CommonPolicyRestHandler interface {
	ListAppEnvPolicies(w http.ResponseWriter, r *http.Request)
	ApplyPolicyToIdentifiers(w http.ResponseWriter, r *http.Request)
}

type CommonPolicyRestHandlerImpl struct {
	commonPolicyActionService policyGovernance.CommonPolicyActionsService
	userService               user.UserService
	enforcer                  casbin.Enforcer
	enforcerUtil              rbac.EnforcerUtil
	validator                 *validator.Validate
	logger                    *zap.SugaredLogger
}

func NewCommonPolicyRestHandlerImpl(commonPolicyActionService policyGovernance.CommonPolicyActionsService,
	userService user.UserService,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil,
	validator *validator.Validate,
	logger *zap.SugaredLogger) *CommonPolicyRestHandlerImpl {
	return &CommonPolicyRestHandlerImpl{
		commonPolicyActionService: commonPolicyActionService,
		userService:               userService,
		enforcer:                  enforcer,
		enforcerUtil:              enforcerUtil,
		validator:                 validator,
		logger:                    logger,
	}
}

func (handler *CommonPolicyRestHandlerImpl) ListAppEnvPolicies(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	policyTypeVar := vars[policyGovernance.PathVariablePolicyTypeVariable]
	policyType := policyGovernance.PathVariablePolicyType(policyTypeVar)
	if lo.Contains(policyGovernance.ExistingPolicyTypes, policyType) {
		common.WriteJsonResp(w, errors.New("profileType not found"), nil, http.StatusNotFound)
		return
	}

	payload := &policyGovernance.AppEnvPolicyMappingsListFilter{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(payload)
	if err != nil {
		handler.logger.Errorw("error in decoding the request payload", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	payload.PolicyType = policyGovernance.PathPolicyTypeGlobalPolicyTypeMap[policyType]
	err = handler.validator.Struct(payload)
	if err != nil {
		handler.logger.Errorw("error in validating the request payload", "err", err, "payload", payload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	result, totalCount, err := handler.commonPolicyActionService.ListAppEnvPolicies(payload)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	response := &ListResponse{
		TotalCount:                   totalCount,
		AppEnvironmentPolicyMappings: result,
	}

	common.WriteJsonResp(w, nil, response, http.StatusOK)
}

func (handler *CommonPolicyRestHandlerImpl) ApplyPolicyToIdentifiers(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionDelete, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	policyTypeVar := vars[policyGovernance.PathVariablePolicyTypeVariable]
	policyType := policyGovernance.PathVariablePolicyType(policyTypeVar)
	if lo.Contains(policyGovernance.ExistingPolicyTypes, policyType) {
		common.WriteJsonResp(w, errors.New("profileType not found"), nil, http.StatusNotFound)
		return
	}

	payload := &policyGovernance.BulkPromotionPolicyApplyRequest{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(payload)
	if err != nil {
		handler.logger.Errorw("error in decoding the request payload", "err", err, "requestBody", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	payload.PolicyType = policyGovernance.PathPolicyTypeGlobalPolicyTypeMap[policyType]
	err = handler.validator.Struct(payload)
	if err != nil {
		handler.logger.Errorw("error in validating the request payload", "err", err, "payload", payload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	err = handler.commonPolicyActionService.ApplyPolicyToIdentifiers(userId, payload)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
