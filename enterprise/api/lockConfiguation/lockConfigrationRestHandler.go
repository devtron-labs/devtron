package lockConfiguation

import (
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
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
	logger      *zap.SugaredLogger
	userService user.UserService
	enforcer    casbin.Enforcer
	validator   validator.Validate
}

func NewLockConfigRestHandlerImpl(logger *zap.SugaredLogger,
	userService user.UserService,
	enforcer casbin.Enforcer,
	validator validator.Validate) *LockConfigRestHandlerImpl {
	return &LockConfigRestHandlerImpl{
		logger:      logger,
		userService: userService,
		enforcer:    enforcer,
		validator:   validator,
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

}

func (impl LockConfigRestHandlerImpl) DeleteLockConfig(w http.ResponseWriter, r *http.Request) {

}
