package deploymentWindow

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	util2 "github.com/devtron-labs/devtron/util"
	"net/url"
	"strconv"
	"time"

	"github.com/devtron-labs/devtron/enterprise/pkg/deploymentWindow"
	"github.com/devtron-labs/devtron/pkg/auth/user"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"gopkg.in/go-playground/validator.v9"
	"net/http"
)

type DeploymentWindowRestHandler interface {
	CreateDeploymentWindowProfile(w http.ResponseWriter, r *http.Request)
	UpdateDeploymentWindowProfile(w http.ResponseWriter, r *http.Request)
	DeleteDeploymentWindowProfile(w http.ResponseWriter, r *http.Request)
	GetDeploymentWindowProfile(w http.ResponseWriter, r *http.Request)
	ListAppDeploymentWindowProfiles(w http.ResponseWriter, r *http.Request)

	GetDeploymentWindowProfileAppOverview(w http.ResponseWriter, r *http.Request)

	GetDeploymentWindowProfileStateForApp(w http.ResponseWriter, r *http.Request)
	GetDeploymentWindowProfileStateForAppGroup(w http.ResponseWriter, r *http.Request)
}

type DeploymentWindowRestHandlerImpl struct {
	logger                  *zap.SugaredLogger
	userService             user.UserService
	enforcer                casbin.Enforcer
	enforcerUtil            rbac.EnforcerUtil
	validator               *validator.Validate
	deploymentWindowService deploymentWindow.DeploymentWindowService
}

func NewDeploymentWindowRestHandlerImpl(
	logger *zap.SugaredLogger,
	userService user.UserService,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil,
	validator *validator.Validate,
	deploymentWindowService deploymentWindow.DeploymentWindowService,
) *DeploymentWindowRestHandlerImpl {
	return &DeploymentWindowRestHandlerImpl{
		logger:                  logger,
		userService:             userService,
		enforcer:                enforcer,
		enforcerUtil:            enforcerUtil,
		validator:               validator,
		deploymentWindowService: deploymentWindowService,
	}
}

func (handler *DeploymentWindowRestHandlerImpl) CreateDeploymentWindowProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	// RBAC enforcer applying
	token := r.Header.Get("token")
	if isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	var request *deploymentWindow.DeploymentWindowProfile
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("err in decoding request in DeploymentWindowProfile", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err in DeploymentWindowProfile", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	configDrafts, err := handler.deploymentWindowService.CreateDeploymentWindowProfile(request, userId)
	if err != nil {
		handler.logger.Errorw("error occurred creating DeploymentWindowProfile", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, configDrafts, http.StatusOK)
}

func (handler *DeploymentWindowRestHandlerImpl) UpdateDeploymentWindowProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	var request *deploymentWindow.DeploymentWindowProfile
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("err in decoding request in DeploymentWindowProfile", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err in DeploymentWindowProfile", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	configDrafts, err := handler.deploymentWindowService.UpdateDeploymentWindowProfile(request, userId)
	if err != nil {
		handler.logger.Errorw("error occurred updating DeploymentWindowProfile", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, configDrafts, http.StatusOK)
}

func (handler *DeploymentWindowRestHandlerImpl) DeleteDeploymentWindowProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}
	v := r.URL.Query()
	id, err := strconv.Atoi(v.Get("profileId"))
	if err != nil {
		common.WriteJsonResp(w, err, "please provide valid profileId", http.StatusBadRequest)
		return
	}
	err = handler.deploymentWindowService.DeleteDeploymentWindowProfileForId(id, userId)
	if err != nil {
		handler.logger.Errorw("error occurred updating DeploymentWindowProfile", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, "", http.StatusOK)

}

func (handler *DeploymentWindowRestHandlerImpl) GetDeploymentWindowProfile(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	v := r.URL.Query()
	id, err := strconv.Atoi(v.Get("profileId"))
	if err != nil {
		common.WriteJsonResp(w, err, "please provide valid profileId", http.StatusBadRequest)
		return
	}
	response, err := handler.deploymentWindowService.GetDeploymentWindowProfileForId(id)
	if err != nil {
		handler.logger.Errorw("error occurred fetching DeploymentWindowProfile", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, response, http.StatusOK)

}

func (handler *DeploymentWindowRestHandlerImpl) ListAppDeploymentWindowProfiles(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	token := r.Header.Get("token")
	if isSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*"); !isSuperAdmin {
		common.WriteJsonResp(w, errors.New("unauthorized"), nil, http.StatusForbidden)
		return
	}

	response, err := handler.deploymentWindowService.ListDeploymentWindowProfiles()
	if err != nil {
		handler.logger.Errorw("error occurred fetching DeploymentWindowProfile", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, response, http.StatusOK)
}

func (handler *DeploymentWindowRestHandlerImpl) GetDeploymentWindowProfileAppOverview(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	v := r.URL.Query()
	appId, envIds, err := handler.getAppIdAndEnvIdsFromQueryParam(w, v)
	if err != nil {
		return
	}

	objects, envObjectToName := handler.enforcerUtil.GetRbacObjectsByEnvIdsAndAppId(envIds, appId)
	var rbacObjectArr []string
	for _, object := range objects {
		rbacObjectArr = append(rbacObjectArr, object)
	}
	unauthorizedResources := make([]string, 0)
	results := handler.enforcer.EnforceInBatch(token, casbin.ResourceApplications, casbin.ActionGet, rbacObjectArr)
	for _, resourceId := range envIds {
		resourceObject := objects[resourceId]
		if !results[resourceObject] {
			unauthorizedResources = append(unauthorizedResources, envObjectToName[resourceObject])
		}
	}
	if len(unauthorizedResources) > 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	response, err := handler.deploymentWindowService.GetDeploymentWindowProfileOverview(appId, envIds)
	if err != nil {
		handler.logger.Errorw("error occurred fetching DeploymentWindowProfileOverview", "err", err, "appId", appId, "envIds", envIds, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, response, http.StatusOK)

}

func (handler *DeploymentWindowRestHandlerImpl) GetDeploymentWindowProfileStateForApp(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	token := r.Header.Get("token")
	v := r.URL.Query()
	appId, envIds, err := handler.getAppIdAndEnvIdsFromQueryParam(w, v)
	if err != nil {
		return
	}

	objects, envObjectToName := handler.enforcerUtil.GetRbacObjectsByEnvIdsAndAppId(envIds, appId)
	var rbacObjectArr []string
	for _, object := range objects {
		rbacObjectArr = append(rbacObjectArr, object)
	}
	unauthorizedResources := make([]string, 0)
	results := handler.enforcer.EnforceInBatch(token, casbin.ResourceApplications, casbin.ActionGet, rbacObjectArr)
	for _, resourceId := range envIds {
		resourceObject := objects[resourceId]
		if !results[resourceObject] {
			unauthorizedResources = append(unauthorizedResources, envObjectToName[resourceObject])
		}
	}
	if len(unauthorizedResources) > 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	requestTime := time.Now()
	response, err := handler.deploymentWindowService.GetDeploymentWindowProfileState(requestTime, appId, envIds, userId)
	if err != nil {
		handler.logger.Errorw("error occurred fetching DeploymentWindowProfileState", "err", err, "request time", requestTime, "appId", appId, "envIds", envIds, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, response, http.StatusOK)

}

func (handler *DeploymentWindowRestHandlerImpl) getAppIdAndEnvIdsFromQueryParam(w http.ResponseWriter, v url.Values) (int, []int, error) {
	appId, err := strconv.Atoi(v.Get("appId"))
	if err != nil {
		common.WriteJsonResp(w, err, "please provide valid envIds", http.StatusBadRequest)
		return appId, nil, err
	}
	envIdsString := v.Get("envIds")
	envIds, err := util2.SplitCommaSeparatedIntValues(envIdsString)
	if err != nil {
		common.WriteJsonResp(w, err, "please provide valid envIds", http.StatusBadRequest)
	}
	return appId, envIds, err
}

func (handler *DeploymentWindowRestHandlerImpl) GetDeploymentWindowProfileStateForAppGroup(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	//TODO RBAC batch

	var request []deploymentWindow.AppEnvSelector
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("err in decoding request in GetDeploymentWindowProfileStateForAppGroup", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	// validate request
	err = handler.validator.Struct(request)
	if err != nil {
		handler.logger.Errorw("validation err in GetDeploymentWindowProfileStateForAppGroup", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	requestTime := time.Now()
	response, err := handler.deploymentWindowService.GetDeploymentWindowProfileStateAppGroup(requestTime, request, userId)
	if err != nil {
		handler.logger.Errorw("error occurred fetching DeploymentWindowProfileState", "err", err, "request time", requestTime, "request", request, userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, response, http.StatusOK)

}
