package restHandler

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/pkg/configDiff"
	"github.com/devtron-labs/devtron/pkg/configDiff/bean"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
	"strconv"
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
}

func NewDeploymentConfigurationRestHandlerImpl(logger *zap.SugaredLogger,
	userAuthService user.UserService,
	enforcerUtil rbac.EnforcerUtil,
	deploymentConfigurationService configDiff.DeploymentConfigurationService,
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

	if len(configDataQueryParams.AppName) > 0 {
		//RBAC START
		token := r.Header.Get(common.TokenHeaderKey)
		object := handler.enforcerUtil.GetAppRBACName(configDataQueryParams.AppName)
		ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
		if !ok {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), nil, http.StatusForbidden)
			return
		}
		//RBAC END
	}

	res, err := handler.deploymentConfigurationService.GetAllConfigData(r.Context(), configDataQueryParams)
	if err != nil {
		handler.logger.Errorw("service err, GetAllConfigData ", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func getConfigDataQueryParams(r *http.Request) (*bean.ConfigDataQueryParams, error) {
	v := r.URL.Query()

	var identifierId int
	var pipelineId int
	var resourceId int
	var err error

	appName := v.Get("appName")
	envName := v.Get("envName")
	configType := v.Get("configType")
	identifierIdStr := v.Get("identifierId")

	if len(identifierIdStr) > 0 {
		identifierId, err = strconv.Atoi(identifierIdStr)
		if err != nil {
			return nil, err
		}
	}

	pipelineIdStr := v.Get("pipelineId")

	if len(pipelineIdStr) > 0 {
		pipelineId, err = strconv.Atoi(pipelineIdStr)
		if err != nil {
			return nil, err
		}
	}

	resourceName := v.Get("resourceName")
	resourceType := v.Get("resourceType")
	resourceIdStr := v.Get("resourceId")
	if len(resourceIdStr) > 0 {
		resourceId, err = strconv.Atoi(resourceIdStr)
		if err != nil {
			return nil, err
		}
	}
	return &bean.ConfigDataQueryParams{
		AppName:      appName,
		EnvName:      envName,
		ConfigType:   configType,
		IdentifierId: identifierId,
		ResourceName: resourceName,
		ResourceType: resourceType,
		PipelineId:   pipelineId,
		ResourceId:   resourceId,
	}, nil
}
