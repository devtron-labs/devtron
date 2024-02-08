package restHandler

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/config"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type DeploymentConfigurationRestHandler interface {
	ConfigAutoComplete(w http.ResponseWriter, r *http.Request)
}
type DeploymentConfigurationRestHandlerImpl struct {
	logger                         *zap.SugaredLogger
	userAuthService                user.UserService
	validator                      *validator.Validate
	enforcerUtil                   rbac.EnforcerUtil
	deploymentConfigurationService config.DeploymentConfigurationService
}

func NewDeploymentConfigurationRestHandlerImpl(logger *zap.SugaredLogger,
	userAuthService user.UserService,
	enforcerUtil rbac.EnforcerUtil,
	deploymentConfigurationService config.DeploymentConfigurationService,
) *DeploymentConfigurationRestHandlerImpl {
	return &DeploymentConfigurationRestHandlerImpl{
		logger:                         logger,
		userAuthService:                userAuthService,
		enforcerUtil:                   enforcerUtil,
		deploymentConfigurationService: deploymentConfigurationService,
	}
}

func (handler *DeploymentConfigurationRestHandlerImpl) ConfigAutoComplete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	appId, err := common.ExtractIntQueryParam(w, r, "appId", nil)
	if err != nil {
		return
	}
	envId, err := common.ExtractIntQueryParam(w, r, "envId", nil)
	if err != nil {
		return
	}

	//RBAC START
	token := r.Header.Get(common.TokenHeaderKey)
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.deploymentConfigurationService.ConfigAutoComplete(appId, envId)
	if err != nil {
		handler.logger.Errorw("service err, ConfigAutoComplete ", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
