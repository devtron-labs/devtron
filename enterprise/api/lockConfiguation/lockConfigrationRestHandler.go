package lockConfiguation

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration"
	"github.com/devtron-labs/devtron/enterprise/pkg/lockConfiguration/bean"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
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
}

func NewLockConfigRestHandlerImpl(logger *zap.SugaredLogger,
	userService user.UserService,
	enforcer casbin.Enforcer,
	validator *validator.Validate,
	lockConfigurationService lockConfiguration.LockConfigurationService) *LockConfigRestHandlerImpl {
	return &LockConfigRestHandlerImpl{
		logger:                   logger,
		userService:              userService,
		enforcer:                 enforcer,
		validator:                validator,
		lockConfigurationService: lockConfigurationService,
	}
}

func (impl LockConfigRestHandlerImpl) GetLockConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)

	if err != nil || userId == 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

}

func (impl LockConfigRestHandlerImpl) CreateLockConfig(w http.ResponseWriter, r *http.Request) {
	userId, err := impl.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// handle super-admin RBAC
	token := r.Header.Get("token")
	if ok := impl.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionUpdate, "*"); !ok {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	// decode request
	decoder := json.NewDecoder(r.Body)
	var request *bean.LockConfigRequest
	err = decoder.Decode(&request)
	if err != nil {
		impl.logger.Errorw("err in decoding request in CreateTags", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = impl.validator.Struct(request)
	if err != nil {
		impl.logger.Errorw("validation err in CreateTags", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// service call
	err = impl.lockConfigurationService.SaveLockConfiguration(request, userId)
	if err != nil {
		impl.logger.Errorw("service err, CreateTags", "err", err, "payload", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, nil, http.StatusOK)
}

func (impl LockConfigRestHandlerImpl) DeleteLockConfig(w http.ResponseWriter, r *http.Request) {

}

type Payload struct {
	Allowed bool     `json:"allowed"`
	Config  []string `json:"config"`
}
