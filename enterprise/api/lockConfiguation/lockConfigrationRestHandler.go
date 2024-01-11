package lockConfiguation

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
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
	isAuthorised, err := handler.userService.CheckRoleForAppAdminAndManager(userId, token)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
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
