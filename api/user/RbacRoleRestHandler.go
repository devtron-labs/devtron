package user

import (
	"encoding/json"
	"errors"
	"fmt"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/bean"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/devtron-labs/devtron/util"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
)

type RbacRoleRestHandler interface {
	GetDefaultRoleDetailById(w http.ResponseWriter, r *http.Request)
	GetAllDefaultRolesByEntityAccessType(w http.ResponseWriter, r *http.Request)
	GetAllDefaultRoles(w http.ResponseWriter, r *http.Request)
	GetRbacPolicyResourceListForAllEntityAccessTypes(w http.ResponseWriter, r *http.Request)
	GetRbacPolicyResourceListByEntityAndAccessType(w http.ResponseWriter, r *http.Request)
	CreateDefaultRole(w http.ResponseWriter, r *http.Request)
	UpdateDefaultRole(w http.ResponseWriter, r *http.Request)
}

type RbacRoleRestHandlerImpl struct {
	logger          *zap.SugaredLogger
	validator       *validator.Validate
	rbacRoleService user.RbacRoleService
	userService     user.UserService
	enforcer        casbin.Enforcer
	enforcerUtil    rbac.EnforcerUtil
}

func NewRbacRoleHandlerImpl(logger *zap.SugaredLogger,
	validator *validator.Validate, rbacRoleService user.RbacRoleService,
	userService user.UserService, enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil) *RbacRoleRestHandlerImpl {
	rbacRoleRestHandlerImpl := &RbacRoleRestHandlerImpl{
		logger:          logger,
		validator:       validator,
		rbacRoleService: rbacRoleService,
		userService:     userService,
		enforcer:        enforcer,
		enforcerUtil:    enforcerUtil,
	}
	return rbacRoleRestHandlerImpl
}

func (handler *RbacRoleRestHandlerImpl) GetDefaultRoleDetailById(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	roleId, err := strconv.Atoi(vars["roleId"])
	if err != nil {
		handler.logger.Errorw("request err, GetDefaultRoleDetail", "err", err, "roleId", roleId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, GetDefaultRoleDetail", "roleId", roleId)
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, bean.AllObjectAccessPlaceholder); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	roleDetail, err := handler.rbacRoleService.GetDefaultRoleDetail(roleId)
	if err != nil {
		handler.logger.Errorw("service error, GetDefaultRoleDetail", "err", err, "roleId", roleId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, roleDetail, http.StatusOK)
}

func (handler *RbacRoleRestHandlerImpl) GetAllDefaultRolesByEntityAccessType(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	entity := vars["entity"]
	v := r.URL.Query()
	accessType := v.Get("accessType")
	handler.logger.Debugw("request payload, GetAllDefaultRolesByEntityAccessType", "entity", entity, "accessType", accessType)
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, bean.AllObjectAccessPlaceholder); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	roles, err := handler.rbacRoleService.GetAllDefaultRolesByEntityAccessType(entity, accessType)
	if err != nil {
		handler.logger.Errorw("service error, GetAllDefaultRolesByEntityAccessType", "err", err, "entity", entity, "accessType", accessType)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, roles, http.StatusOK)
}

func (handler *RbacRoleRestHandlerImpl) GetAllDefaultRoles(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	handler.logger.Debugw("request payload, GetAllDefaultRoles")
	// RBAC enforcer applying
	token := r.Header.Get("token")
	emailId, err := handler.userService.GetEmailFromToken(token)
	if err != nil {
		handler.logger.Errorw("error in getting user emailId from token", "userId", userId, "err", err)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	teamNames, err := handler.enforcerUtil.GetAllActiveTeamNames()
	if err != nil {
		handler.logger.Errorw("error in finding all active team names", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if len(teamNames) > 0 {
		rbacResultMap := handler.enforcer.EnforceByEmailInBatch(emailId, casbin.ResourceUser, casbin.ActionGet, teamNames)
		isAuthorized := false
		for _, authorizedOnTeam := range rbacResultMap {
			if authorizedOnTeam {
				isAuthorized = true
				break
			}
		}
		if !isAuthorized {
			common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
			return
		}
	}
	roles, err := handler.rbacRoleService.GetAllDefaultRoles()
	if err != nil {
		handler.logger.Errorw("service error, GetAllDefaultRoles", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, roles, http.StatusOK)
}

func (handler *RbacRoleRestHandlerImpl) GetRbacPolicyResourceListForAllEntityAccessTypes(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	handler.logger.Debugw("request payload, GetRbacPolicyResourceListForAllEntityAccessTypes", "userId", userId)
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, bean.AllObjectAccessPlaceholder); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	policyEntityGroupList, err := handler.rbacRoleService.GetRbacPolicyResourceListForAllEntityAccessTypes()
	if err != nil {
		handler.logger.Errorw("service error, GetRbacPolicyResourceListForAllEntityAccessTypes", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, policyEntityGroupList, http.StatusOK)
}

func (handler *RbacRoleRestHandlerImpl) GetRbacPolicyResourceListByEntityAndAccessType(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	handler.logger.Debugw("request payload, GetRbacPolicyResourceListByEntityAndAccessType", "userId", userId)
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, bean.AllObjectAccessPlaceholder); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	entity := vars["entity"]
	v := r.URL.Query()
	accessType := v.Get("accessType")
	policyResourceList, err := handler.rbacRoleService.GetPolicyResourceListByEntityAccessType(entity, accessType)
	if err != nil {
		handler.logger.Errorw("service error, GetPolicyResourceListByEntityAccessType", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, policyResourceList, http.StatusOK)
}

func (handler *RbacRoleRestHandlerImpl) CreateDefaultRole(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var rbacRolePayload bean.RbacRoleDto
	err = decoder.Decode(&rbacRolePayload)
	if err != nil {
		handler.logger.Errorw("request err, CreateDefaultRole", "err", err, "payload", rbacRolePayload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, CreateDefaultRole", "payload", rbacRolePayload)
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, bean.AllObjectAccessPlaceholder); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	isValidationError, err := handler.validateRbacRoleCreateUpdateRequest(rbacRolePayload)
	if err != nil {
		handler.logger.Errorw("error, validateRbacRoleCreateUpdateRequest", "payload", rbacRolePayload, "err", err)
		if isValidationError {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	err = handler.rbacRoleService.CreateDefaultRole(&rbacRolePayload, userId)
	if err != nil {
		handler.logger.Errorw("service error, CreateDefaultRole", "payload", rbacRolePayload, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "Custom role creation succeeded.", http.StatusOK)
}

func (handler *RbacRoleRestHandlerImpl) validateRbacRoleCreateUpdateRequest(rbacRolePayload bean.RbacRoleDto) (bool, error) {
	isValidationError := true
	err := handler.validator.Struct(rbacRolePayload)
	if err != nil {
		handler.logger.Errorw("err, validateRbacRoleCreateUpdateRequest", "payload", rbacRolePayload, "err", err)
		return isValidationError, err
	}
	if rbacRolePayload.Entity == bean.ENTITY_APPS &&
		rbacRolePayload.AccessType != bean.DEVTRON_APP && rbacRolePayload.AccessType != bean2.APP_ACCESS_TYPE_HELM {

		return isValidationError, fmt.Errorf("invalid access type, please update and retry")
	}
	policyResourceEntityGroup, err := handler.rbacRoleService.GetPolicyResourceListByEntityAccessType(rbacRolePayload.Entity, rbacRolePayload.AccessType)
	if err != nil {
		handler.logger.Errorw("service err, GetRbacPolicyResourceListForAllEntityAccessTypes", "entity", rbacRolePayload.Entity, "accessType", rbacRolePayload.AccessType, "err", err)
		isValidationError = false
		return isValidationError, err
	}
	policyResourceListMap := make(map[string][]string, len(policyResourceEntityGroup.ResourceDetailList))
	for _, policyResource := range policyResourceEntityGroup.ResourceDetailList {
		policyResource.Actions = append(policyResource.Actions, bean.AllObjectAccessPlaceholder)
		policyResourceListMap[policyResource.Resource] = policyResource.Actions
	}
	for _, reqResourceDetail := range rbacRolePayload.ResourceDetailList {
		if actions, ok := policyResourceListMap[reqResourceDetail.Resource]; !ok {
			return isValidationError, fmt.Errorf("invalid resource in request: %s, please update and retry", reqResourceDetail.Resource)
		} else if !util.IsSubset(reqResourceDetail.Actions, actions) {
			return isValidationError, fmt.Errorf("invalid actions in request: %v for resource: %s, please update and retry", reqResourceDetail.Actions, reqResourceDetail.Resource)
		}
	}
	isValidationError = false
	return isValidationError, nil
}
func (handler *RbacRoleRestHandlerImpl) UpdateDefaultRole(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var rbacRolePayload bean.RbacRoleDto
	err = decoder.Decode(&rbacRolePayload)
	if err != nil {
		handler.logger.Errorw("request err, UpdateDefaultRole", "err", err, "payload", rbacRolePayload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, UpdateDefaultRole", "payload", rbacRolePayload)
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionCreate, bean.AllObjectAccessPlaceholder); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	isValidationError, err := handler.validateRbacRoleCreateUpdateRequest(rbacRolePayload)
	if err != nil {
		handler.logger.Errorw("error, validateRbacRoleCreateUpdateRequest", "payload", rbacRolePayload, "err", err)
		if isValidationError {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	err = handler.rbacRoleService.UpdateDefaultRole(&rbacRolePayload, userId)
	if err != nil {
		handler.logger.Errorw("service error, UpdateDefaultRole", "payload", rbacRolePayload, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, "Custom role update succeeded.", http.StatusOK)
}
