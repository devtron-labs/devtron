package lockConfiguation

import (
	"encoding/json"
	"errors"
	bean2 "github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strings"
)

type LockConfigRestHandler interface {
	GetLockConfig(w http.ResponseWriter, r *http.Request)
	CreateLockConfig(w http.ResponseWriter, r *http.Request)
	DeleteLockConfig(w http.ResponseWriter, r *http.Request)
}

type LockConfigRestHandlerImpl struct {
	logger                   *zap.SugaredLogger
	userService              user.UserService
	enforcer                 casbin.Enforcer
	validator                *validator.Validate
	lockConfigurationService lockConfiguration.LockConfigurationService
	userCommonService        user.UserCommonService
	enforcerUtil             rbac.EnforcerUtil
}

func NewLockConfigRestHandlerImpl(logger *zap.SugaredLogger,
	userService user.UserService,
	enforcer casbin.Enforcer,
	validator *validator.Validate,
	lockConfigurationService lockConfiguration.LockConfigurationService,
	userCommonService user.UserCommonService,
	enforcerUtil rbac.EnforcerUtil) *LockConfigRestHandlerImpl {
	return &LockConfigRestHandlerImpl{
		logger:                   logger,
		userService:              userService,
		enforcer:                 enforcer,
		validator:                validator,
		lockConfigurationService: lockConfigurationService,
		userCommonService:        userCommonService,
		enforcerUtil:             enforcerUtil,
	}

}

func (handler LockConfigRestHandlerImpl) GetLockConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if err != nil || userId == 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	isAuthorised := false
	//checking superAdmin access
	isAuthorised, err = handler.userService.IsSuperAdminForDevtronManaged(int(userId))
	if err != nil {
		handler.logger.Errorw("error in checking superAdmin access of user", "err", err)
		common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
		return
	}
	if !isAuthorised {
		user, err := handler.userService.GetRoleFiltersForAUserById(userId)
		if err != nil {
			handler.logger.Errorw("error in getting user by id", "err", err)
			common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
			return
		}
		// ApplicationResource pe Create
		var roleFilters []bean2.RoleFilter
		if len(user.Groups) > 0 {
			groupRoleFilters, err := handler.userService.GetRoleFiltersByGroupNames(user.Groups)
			if err != nil {
				handler.logger.Errorw("Error in getting role filters by group names", "err", err, "groupNames", user.Groups)
				common.WriteJsonResp(w, err, "", http.StatusInternalServerError)
				return
			}
			if len(groupRoleFilters) > 0 {
				roleFilters = append(roleFilters, groupRoleFilters...)
			}
		}
		if user.RoleFilters != nil && len(user.RoleFilters) > 0 {
			roleFilters = append(roleFilters, user.RoleFilters...)
		}
		if len(roleFilters) > 0 {
			for _, filter := range roleFilters {
				if len(filter.Team) > 0 {
					entityNames := strings.Split(filter.EntityName, ",")
					if len(entityNames) > 0 {
						for _, val := range entityNames {
							resourceName := handler.enforcerUtil.GetProjectOrAppAdminRBACNameByAppNameAndTeamName(val, filter.Team)
							if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); ok {
								isAuthorised = true
								break
							}
						}
					}

				}
			}
		}
	}
	if !isAuthorised {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	resp, err := handler.lockConfigurationService.GetLockConfiguration()
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, resp, http.StatusOK)
}

func (handler LockConfigRestHandlerImpl) CreateLockConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	var request *bean.LockConfigRequest
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("err in decoding request in LockConfigRequest", "err", err, "body", r.Body)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err in LockConfigRequest", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// service call
	err = handler.lockConfigurationService.SaveLockConfiguration(request, userId)
	if err != nil {
		handler.logger.Errorw("service err, SaveLockConfiguration", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, request, http.StatusOK)
}

func (handler LockConfigRestHandlerImpl) DeleteLockConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	// service call
	err = handler.lockConfigurationService.DeleteActiveLockConfiguration(userId)
	if err != nil {
		handler.logger.Errorw("service err, DeleteActiveLockConfiguration", "err", err, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (handler LockConfigRestHandlerImpl) CheckAdminAuth(resource, token string, object string) bool {
	if ok := handler.enforcer.Enforce(token, resource, casbin.ActionCreate, object); !ok {
		return false
	}
	return true
}
