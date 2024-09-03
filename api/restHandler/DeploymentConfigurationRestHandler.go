package restHandler

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/configDiff"
	"github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/schema"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type DeploymentConfigurationRestHandler interface {
	ConfigAutoComplete(w http.ResponseWriter, r *http.Request)
	GetConfigData(w http.ResponseWriter, r *http.Request)
}
type DeploymentConfigurationRestHandlerImpl struct {
	logger                         *zap.SugaredLogger
	userAuthService                user.UserService
	validator                      *validator.Validate
	enforcerUtil                   rbac.EnforcerUtil
	deploymentConfigurationService configDiff.DeploymentConfigurationService
	enforcer                       casbin.Enforcer
}

func NewDeploymentConfigurationRestHandlerImpl(logger *zap.SugaredLogger,
	userAuthService user.UserService,
	enforcerUtil rbac.EnforcerUtil,
	deploymentConfigurationService configDiff.DeploymentConfigurationService,
	enforcer casbin.Enforcer,
) *DeploymentConfigurationRestHandlerImpl {
	return &DeploymentConfigurationRestHandlerImpl{
		logger:                         logger,
		userAuthService:                userAuthService,
		enforcerUtil:                   enforcerUtil,
		deploymentConfigurationService: deploymentConfigurationService,
		enforcer:                       enforcer,
	}
}

func (handler *DeploymentConfigurationRestHandlerImpl) ConfigAutoComplete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	appId, err := common.ExtractIntQueryParam(w, r, "appId", 0)
	if err != nil {
		return
	}
	envId, err := common.ExtractIntQueryParam(w, r, "envId", 0)
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
		handler.logger.Errorw("service err, ConfigAutoComplete ", "appId", appId, "envId", envId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *DeploymentConfigurationRestHandlerImpl) GetConfigData(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	configDataQueryParams, err := getConfigDataQueryParams(r)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	//RBAC START
	token := r.Header.Get(common.TokenHeaderKey)
	object := handler.enforcerUtil.GetAppRBACName(configDataQueryParams.AppName)
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.deploymentConfigurationService.GetAllConfigData(r.Context(), configDataQueryParams)
	if err != nil {
		handler.logger.Errorw("service err, GetAllConfigData ", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res.IsAppAdmin = handler.enforceForAppAndEnv(configDataQueryParams.AppName, configDataQueryParams.EnvName, token, casbin.ActionUpdate)

	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (handler *DeploymentConfigurationRestHandlerImpl) enforceForAppAndEnv(appName, envName string, token string, action string) bool {
	object := handler.enforcerUtil.GetAppRBACNameByAppName(appName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, action, object); !ok {
		return false
	}

	if len(envName) > 0 {
		object = handler.enforcerUtil.GetEnvRBACNameByAppAndEnvName(appName, envName)
		if ok := handler.enforcer.Enforce(token, casbin.ResourceEnvironment, action, object); !ok {
			return false
		}
	}
	return true
}
func getConfigDataQueryParams(r *http.Request) (*bean.ConfigDataQueryParams, error) {
	v := r.URL.Query()
	var decoder = schema.NewDecoder()
	decoder.IgnoreUnknownKeys(true)
	queryParams := bean.ConfigDataQueryParams{}
	err := decoder.Decode(&queryParams, v)
	if err != nil {
		return nil, err
	}

	return &queryParams, nil
}
