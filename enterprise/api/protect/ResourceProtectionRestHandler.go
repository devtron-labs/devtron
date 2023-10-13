package protect

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/protect"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type ResourceProtectionRestHandler interface {
	ConfigureResourceProtect(w http.ResponseWriter, r *http.Request)
	GetResourceProtectMetadata(w http.ResponseWriter, r *http.Request)
	GetResourceProtectMetadataForEnv(w http.ResponseWriter, r *http.Request)
}

type ResourceProtectionRestHandlerImpl struct {
	logger                    *zap.SugaredLogger
	userService               user.UserService
	enforcer                  casbin.Enforcer
	enforcerUtil              rbac.EnforcerUtil
	validator                 *validator.Validate
	resourceProtectionService protect.ResourceProtectionService
}

func NewResourceProtectionRestHandlerImpl(logger *zap.SugaredLogger, resourceProtectionService protect.ResourceProtectionService,
	userService user.UserService, enforcer casbin.Enforcer, enforcerUtil rbac.EnforcerUtil,
	validator *validator.Validate) *ResourceProtectionRestHandlerImpl {
	return &ResourceProtectionRestHandlerImpl{
		logger:                    logger,
		userService:               userService,
		enforcer:                  enforcer,
		enforcerUtil:              enforcerUtil,
		validator:                 validator,
		resourceProtectionService: resourceProtectionService,
	}
}

func (handler *ResourceProtectionRestHandlerImpl) ConfigureResourceProtect(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var request protect.ResourceProtectModel
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("err in decoding request in ResourceProtectModel", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// validate request
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err in ResourceProtectModel", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	request.UserId = userId
	err = handler.resourceProtectionService.ConfigureResourceProtection(&request)
	if err != nil {
		handler.logger.Errorw("error occurred while configuring resource protection", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}

func (handler *ResourceProtectionRestHandlerImpl) GetResourceProtectMetadata(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	appId, err := common.ExtractIntQueryParam(w, r, "appId", nil)
	if err != nil {
		return
	}

	token := r.Header.Get("token")
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	protectModels, err := handler.resourceProtectionService.GetResourceProtectMetadata(appId)
	if err != nil {
		handler.logger.Errorw("error occurred while fetching resource protection", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, protectModels, http.StatusOK)
}

func (handler *ResourceProtectionRestHandlerImpl) GetResourceProtectMetadataForEnv(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	envId, err := common.ExtractIntQueryParam(w, r, "envId", nil)
	if err != nil {
		return
	}

	token := r.Header.Get("token")
	userEmailId, err := handler.userService.GetEmailFromToken(token)
	if err != nil {
		handler.logger.Errorw("error in getting user emailId from token", "userId", userId, "err", err)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	appStatues := handler.resourceProtectionService.ResourceProtectionEnabledForEnv(envId)
	if err != nil {
		handler.logger.Errorw("error occurred while fetching resource protection", "err", err, "envId", envId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	var rbacObjectArray []string
	rbacObjectVsAppIdMap := make(map[string]int)
	for appId, _ := range appStatues {
		rbacObject := handler.enforcerUtil.GetTeamEnvRBACNameByAppId(appId, envId)
		rbacObjectArray = append(rbacObjectArray, rbacObject)
		rbacObjectVsAppIdMap[rbacObject] = appId
	}

	rbacResponse := handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceApplications, casbin.ActionGet, rbacObjectArray)
	appStatusResponse := make(map[int]bool)
	for rbacObj := range rbacResponse {
		appId := rbacObjectVsAppIdMap[rbacObj]
		appStatusResponse[appId] = appStatues[appId]
	}
	if len(appStatusResponse) == 0 && len(appStatues) != 0 {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}

	common.WriteJsonResp(w, nil, appStatusResponse, http.StatusOK)
}
