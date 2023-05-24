package user

import (
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type RbacRoleRestHandler interface {
	GetAllDefaultRoles(w http.ResponseWriter, r *http.Request)
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
