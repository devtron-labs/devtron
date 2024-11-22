package configDiff

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/configDiff"
	"github.com/devtron-labs/devtron/pkg/configDiff/bean"
	util2 "github.com/devtron-labs/devtron/util"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"time"
)

type DeploymentConfigurationRestHandler interface {
	ConfigAutoComplete(w http.ResponseWriter, r *http.Request)
	GetConfigData(w http.ResponseWriter, r *http.Request)
	CompareCategoryWiseConfigData(w http.ResponseWriter, r *http.Request)
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

	configDataQueryParams.UserId = userId
	//RBAC START
	token := r.Header.Get(common.TokenHeaderKey)
	object := handler.enforcerUtil.GetAppRBACName(configDataQueryParams.AppName)
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	userHasAdminAccess := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, object)
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()
	ctx = util2.SetSuperAdminInContext(ctx, isSuperAdmin)
	res, err := handler.deploymentConfigurationService.GetAllConfigData(ctx, configDataQueryParams, userHasAdminAccess)
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

func (handler *DeploymentConfigurationRestHandlerImpl) CompareCategoryWiseConfigData(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	configCategory := vars["category"]
	var comparisonRequestDto bean.ComparisonRequestDto
	err = json.NewDecoder(r.Body).Decode(&comparisonRequestDto)
	if err != nil {
		handler.logger.Errorw("error in decoding request body", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	err = validateComparisonRequest(configCategory, comparisonRequestDto)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	comparisonRequestDto.UpdateUserIdInComparisonItems(userId)

	//RBAC START
	token := r.Header.Get(common.TokenHeaderKey)
	object := handler.enforcerUtil.GetAppRBACName(comparisonRequestDto.AppName)

	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
	if !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
		return
	}
	//RBAC END
	//isSuperAdmin is required to make decision if a sensitive data(as defined by super admin) needs to be redacted
	//or not while resolving scope variable.
	isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")
	//userHasAdminAccess is required to mask secrets in the response after scope resolution.
	userHasAdminAccess := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, object)

	ctx := util2.SetSuperAdminInContext(r.Context(), isSuperAdmin)
	res, err := handler.deploymentConfigurationService.CompareCategoryWiseConfigData(ctx, comparisonRequestDto, userHasAdminAccess)
	if err != nil {
		handler.logger.Errorw("service err, GetAllConfigData ", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	//res.IsAppAdmin = handler.enforceForAppAndEnv(configDataQueryParams.AppName, configDataQueryParams.EnvName, token, casbin.ActionUpdate)

	common.WriteJsonResp(w, nil, res, http.StatusOK)
}
