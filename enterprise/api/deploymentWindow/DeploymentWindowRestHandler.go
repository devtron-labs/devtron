package deploymentWindow

import (
	"encoding/json"
	"errors"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	util2 "github.com/devtron-labs/devtron/util"
	"golang.org/x/exp/maps"
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
	pipelineConfigService   pipeline.CdPipelineConfigService
}

func NewDeploymentWindowRestHandlerImpl(
	logger *zap.SugaredLogger,
	userService user.UserService,
	enforcer casbin.Enforcer,
	enforcerUtil rbac.EnforcerUtil,
	validator *validator.Validate,
	deploymentWindowService deploymentWindow.DeploymentWindowService,
	pipelineConfigService pipeline.CdPipelineConfigService,
) *DeploymentWindowRestHandlerImpl {
	return &DeploymentWindowRestHandlerImpl{
		logger:                  logger,
		userService:             userService,
		enforcer:                enforcer,
		enforcerUtil:            enforcerUtil,
		validator:               validator,
		deploymentWindowService: deploymentWindowService,
		pipelineConfigService:   pipelineConfigService,
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

	var request deploymentWindow.DeploymentWindowProfile
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

	profile, err := handler.deploymentWindowService.CreateDeploymentWindowProfile(&request, userId)
	if err != nil {
		handler.logger.Errorw("error occurred creating DeploymentWindowProfile", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, profile, http.StatusOK)
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

	var request deploymentWindow.DeploymentWindowProfile
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

	profile, err := handler.deploymentWindowService.UpdateDeploymentWindowProfile(&request, userId)
	if err != nil {
		handler.logger.Errorw("error occurred updating DeploymentWindowProfile", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, profile, http.StatusOK)
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
	profileName := v.Get("profileName")
	if len(profileName) != 0 {
		err = handler.deploymentWindowService.DeleteDeploymentWindowProfileForName(profileName, userId)
		if err != nil {
			handler.logger.Errorw("error occurred updating DeploymentWindowProfile", "err", err, "profileName", profileName)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, err, "", http.StatusOK)
	}

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

	profileName := v.Get("profileName")
	if len(profileName) != 0 {
		response, err := handler.deploymentWindowService.GetDeploymentWindowProfileForName(profileName)
		if err != nil {
			handler.logger.Errorw("error occurred fetching DeploymentWindowProfile", "err", err, "profileName", profileName)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
		common.WriteJsonResp(w, err, response, http.StatusOK)
	}

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

	authorizedEnvs := handler.filterAuthorizedResources(envIds, appId, token)
	if len(authorizedEnvs) == 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	response, err := handler.deploymentWindowService.GetDeploymentWindowProfileOverview(appId, authorizedEnvs)
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

	authorizedEnvs := handler.filterAuthorizedResources(envIds, appId, token)
	if len(authorizedEnvs) == 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}

	requestTime := time.Now()
	response, err := handler.deploymentWindowService.GetDeploymentWindowProfileState(requestTime, appId, authorizedEnvs, userId)
	if err != nil {
		handler.logger.Errorw("error occurred fetching DeploymentWindowProfileState", "err", err, "request time", requestTime, "appId", appId, "envIds", envIds, "userId", userId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, response, http.StatusOK)

}

func (handler *DeploymentWindowRestHandlerImpl) filterAuthorizedResourcesForGroup(appEnvs []deploymentWindow.AppEnvSelector, token string) []deploymentWindow.AppEnvSelector {

	appToEnvs := make(map[int][]int)
	for _, appEnv := range appEnvs {
		appToEnvs[appEnv.AppId] = append(appToEnvs[appEnv.AppId], appEnv.EnvId)
	}
	rbacObjectToAppEnv := make(map[string]deploymentWindow.AppEnvSelector)
	rbacObjects := make([]string, 0)
	for appId, envIds := range appToEnvs {
		objectMap, _ := handler.enforcerUtil.GetRbacObjectsByEnvIdsAndAppId(envIds, appId)
		rbacObjects = append(rbacObjects, maps.Values(objectMap)...)
		for _, envId := range envIds {
			rbacObjectToAppEnv[objectMap[envId]] = deploymentWindow.AppEnvSelector{
				AppId: appId,
				EnvId: envId,
			}
		}
	}
	authorizedAppEnvSelectors := make([]deploymentWindow.AppEnvSelector, 0)
	results := handler.enforcer.EnforceInBatch(token, casbin.ResourceEnvironment, casbin.ActionGet, rbacObjects)
	for object, isAllowed := range results {
		if isAllowed {
			authorizedAppEnvSelectors = append(authorizedAppEnvSelectors, rbacObjectToAppEnv[object])
		}
	}

	return authorizedAppEnvSelectors
}

func (handler *DeploymentWindowRestHandlerImpl) filterAuthorizedResources(envIds []int, appId int, token string) []int {
	objects, _ := handler.enforcerUtil.GetRbacObjectsByEnvIdsAndAppId(envIds, appId)
	rbacObjectArr := maps.Values(objects)

	authorizedResourceIds := make([]int, 0)
	results := handler.enforcer.EnforceInBatch(token, casbin.ResourceEnvironment, casbin.ActionGet, rbacObjectArr)
	for _, resourceId := range envIds {
		resourceObject := objects[resourceId]
		if results[resourceObject] {
			authorizedResourceIds = append(authorizedResourceIds, resourceId)
		}
	}
	return authorizedResourceIds
}

func (handler *DeploymentWindowRestHandlerImpl) getFilterDays(w http.ResponseWriter, v url.Values) (int, error) {
	daysString := v.Get("days")
	days := 0
	var err error
	if len(daysString) == 0 {
		return days, nil
	}
	days, err = strconv.Atoi(daysString)
	if err != nil {
		common.WriteJsonResp(w, err, "please provide valid filter days", http.StatusBadRequest)
		return 0, nil
	}
	return days, err
}

func (handler *DeploymentWindowRestHandlerImpl) getAppIdAndEnvIdsFromQueryParam(w http.ResponseWriter, v url.Values) (int, []int, error) {
	appId, err := strconv.Atoi(v.Get("appId"))
	if err != nil {
		common.WriteJsonResp(w, err, "please provide valid envIds", http.StatusBadRequest)
		return appId, nil, err
	}
	envIdsString := v.Get("envIds")
	if len(envIdsString) == 0 {
		envIds, err := handler.pipelineConfigService.GetPipelineEnvironmentsForApplication(appId)
		if err != nil {
			common.WriteJsonResp(w, err, "error finding pipelines for app Id", http.StatusBadRequest)
			return 0, nil, err
		}
		return appId, envIds, nil
	}
	envIds, err := util2.SplitCommaSeparatedIntValues(envIdsString)
	if err != nil {
		common.WriteJsonResp(w, err, "please provide valid envIds", http.StatusBadRequest)
	}
	return appId, envIds, nil
}

type payload struct {
	selectors []deploymentWindow.AppEnvSelector
}

func (handler *DeploymentWindowRestHandlerImpl) GetDeploymentWindowProfileStateForAppGroup(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")

	var request []deploymentWindow.AppEnvSelector
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&request)
	if err != nil {
		handler.logger.Errorw("err in decoding request in GetDeploymentWindowProfileStateForAppGroup", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	requestPayload := &payload{selectors: request}

	// validate request
	err = handler.validator.Struct(requestPayload)
	if err != nil {
		handler.logger.Errorw("validation err in GetDeploymentWindowProfileStateForAppGroup", "err", err, "request", request)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	request = handler.filterAuthorizedResourcesForGroup(request, token)
	if len(request) == 0 {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
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
