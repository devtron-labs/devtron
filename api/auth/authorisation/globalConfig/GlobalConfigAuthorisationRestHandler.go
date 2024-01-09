package globalConfig

import (
	"encoding/json"
	"errors"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	auth "github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig/bean"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type AuthorisationConfigRestHandler interface {
	CreateOrUpdateAuthorisationConfig(w http.ResponseWriter, r *http.Request)
	GetAllActiveAuthorisationConfig(w http.ResponseWriter, r *http.Request)
}

type AuthorisationConfigRestHandlerImpl struct {
	validator                        *validator.Validate
	logger                           *zap.SugaredLogger
	enforcer                         casbin.Enforcer
	userService                      user.UserService
	userCommonService                user.UserCommonService
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService
}

func NewGlobalAuthorisationConfigRestHandlerImpl(validator *validator.Validate,
	logger *zap.SugaredLogger, enforcer casbin.Enforcer,
	userService user.UserService,
	globalAuthorisationConfigService auth.GlobalAuthorisationConfigService,
	userCommonService user.UserCommonService,
) *AuthorisationConfigRestHandlerImpl {
	handler := &AuthorisationConfigRestHandlerImpl{
		validator:                        validator,
		logger:                           logger,
		enforcer:                         enforcer,
		userService:                      userService,
		globalAuthorisationConfigService: globalAuthorisationConfigService,
		userCommonService:                userCommonService,
	}
	return handler
}

func (handler *AuthorisationConfigRestHandlerImpl) CreateOrUpdateAuthorisationConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var globalConfigPayload bean.GlobalAuthorisationConfig
	err = decoder.Decode(&globalConfigPayload)
	if err != nil {
		handler.logger.Errorw("request err, CreateOrUpdateAuthorisationConfig", "err", err, "payload", globalConfigPayload)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// Rbac Enforcement
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionCreate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	//Validation
	isValidationError, err := handler.validateGlobalAuthorisationConfigPayload(globalConfigPayload)
	if err != nil {
		handler.logger.Errorw("error, validateGlobalAuthorisationConfigPayload", "payload", globalConfigPayload, "err", err)
		if isValidationError {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else {
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	// Setting User Id for Internal use
	globalConfigPayload.UserId = userId
	// Service Call
	resp, err := handler.globalAuthorisationConfigService.CreateOrUpdateGlobalAuthConfig(globalConfigPayload, nil)
	if err != nil {
		handler.logger.Errorw("service error, CreateOrUpdateAuthorisationConfig", "err", err, "payload", globalConfigPayload)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, nil, resp, http.StatusOK)
}
func (handler *AuthorisationConfigRestHandlerImpl) GetAllActiveAuthorisationConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	isAuthorised := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	if !isAuthorised {
		user, err := handler.userService.GetRoleFiltersForAUserById(userId)
		if err != nil {
			handler.logger.Errorw("error in getting user by id", "err", err)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return
		}
		var roleFilters []bean2.RoleFilter
		if user.RoleFilters != nil && len(user.RoleFilters) > 0 {
			roleFilters = append(roleFilters, user.RoleFilters...)
		}
		if len(roleFilters) > 0 {
			for _, filter := range roleFilters {
				if len(filter.Team) > 0 {
					if ok := handler.enforcer.Enforce(token, casbin.ResourceUser, casbin.ActionGet, filter.Team); ok {
						isAuthorised = true
						break
					}
				}
				if filter.Entity == bean2.CLUSTER_ENTITIY {
					if ok := handler.userCommonService.CheckRbacForClusterEntity(filter.Cluster, filter.Namespace, filter.Group, filter.Kind, filter.Resource, token, handler.checkActionUpdateAuth); ok {
						isAuthorised = true
						break
					}
				}
			}
		}
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// Service Call
	resp, err := handler.globalAuthorisationConfigService.GetAllActiveAuthorisationConfig()
	if err != nil {
		handler.logger.Errorw("service error, GetAllActiveAuthorisationConfig", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, resp, http.StatusOK)

}

func (handler *AuthorisationConfigRestHandlerImpl) validateGlobalAuthorisationConfigPayload(globalConfigPayload bean.GlobalAuthorisationConfig) (bool, error) {
	isValidationError := true
	err := handler.validator.Struct(globalConfigPayload)
	if err != nil {
		handler.logger.Errorw("err, validateGlobalAuthorisationConfigPayload", "payload", globalConfigPayload, "err", err)
		return isValidationError, err
	}
	if len(globalConfigPayload.ConfigTypes) == 0 {
		handler.logger.Errorw("err, validation failed on validateGlobalAuthorisationConfigPayload due to no configType provided", "payload", globalConfigPayload, "err", err)
		return isValidationError, errors.New("no configTypes provided in request")
	}
	isValidationError = false
	return isValidationError, nil
}

func (handler *AuthorisationConfigRestHandlerImpl) checkActionUpdateAuth(resource, token string, object string) bool {
	if ok := handler.enforcer.Enforce(token, resource, casbin.ActionUpdate, object); !ok {
		return false
	}
	return true

}
