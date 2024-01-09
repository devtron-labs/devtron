package app

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/devtron-labs/devtron/api/restHandler/common"
	repository "github.com/devtron-labs/devtron/internal/sql/repository/dockerRegistry"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"k8s.io/utils/strings/slices"
)

type DevtronAppAutoCompleteRestHandler interface {
	GitListAutocomplete(w http.ResponseWriter, r *http.Request)
	RegistriesListAutocomplete(w http.ResponseWriter, r *http.Request)
	TeamListAutocomplete(w http.ResponseWriter, r *http.Request)
	EnvironmentListAutocomplete(w http.ResponseWriter, r *http.Request)
	GetAppListForAutocomplete(w http.ResponseWriter, r *http.Request)
}

func (handler PipelineConfigRestHandlerImpl) GetAppListForAutocomplete(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	token := r.Header.Get("token")
	isActionUserSuperAdmin := handler.enforcer.Enforce(token, casbin.ResourceGlobal, casbin.ActionGet, "*")

	v := r.URL.Query()
	teamId := v.Get("teamId")
	appName := v.Get("appName")
	appTypeParam := v.Get("appType")
	var appType int
	if appTypeParam != "" {
		appType, err = strconv.Atoi(appTypeParam)
		if err != nil {
			handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId, "appTypeParam", appTypeParam)
			common.WriteJsonResp(w, err, "Failed to parse appType param", http.StatusInternalServerError)
			return
		}
	} else {
		// if appType not provided we are considering it as customApp for now, doing this because to get all apps by team id rbac objects
		appType = int(helper.CustomApp)
	}
	var teamIdInt int
	handler.Logger.Infow("request payload, GetAppListForAutocomplete", "teamId", teamId)
	var apps []*pipeline.AppBean
	if len(teamId) == 0 {
		apps, err = handler.pipelineBuilder.FindAllMatchesByAppName(appName, helper.AppType(appType))
		if err != nil {
			handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else {
		teamIdInt, err = strconv.Atoi(teamId)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else {
			apps, err = handler.pipelineBuilder.FindAppsByTeamId(teamIdInt)
			if err != nil {
				handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId)
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
		}
	}
	if isActionUserSuperAdmin {
		common.WriteJsonResp(w, err, apps, http.StatusOK)
		return
	}

	// RBAC
	_, span := otel.Tracer("autoCompleteAppAPI").Start(context.Background(), "RBACForAutoCompleteAppAPI")
	accessedApps := make([]*pipeline.AppBean, 0)
	rbacObjects := make([]string, 0)

	var appIdToObjectMap map[int]string
	if len(teamId) == 0 {
		appIdToObjectMap = handler.enforcerUtil.GetRbacObjectsForAllAppsWithMatchingAppName(appName, helper.AppType(appType))
	} else {
		appIdToObjectMap = handler.enforcerUtil.GetRbacObjectsForAllAppsWithTeamID(teamIdInt, helper.AppType(appType))
	}

	for _, app := range apps {
		object := appIdToObjectMap[app.Id]
		rbacObjects = append(rbacObjects, object)
	}

	enforcedMap := handler.enforcerUtil.CheckAppRbacForAppOrJobInBulk(token, casbin.ActionGet, rbacObjects, helper.AppType(appType))
	for _, app := range apps {
		object := appIdToObjectMap[app.Id]
		if enforcedMap[object] {
			accessedApps = append(accessedApps, app)
		}
	}
	span.End()
	// RBAC
	common.WriteJsonResp(w, err, accessedApps, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) EnvironmentListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, EnvironmentListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	showDeploymentOptionsParam := false
	param := r.URL.Query().Get("showDeploymentOptions")
	if param != "" {
		showDeploymentOptionsParam, _ = strconv.ParseBool(param)
	}
	result, err := handler.envService.GetEnvironmentListForAutocomplete(showDeploymentOptionsParam)
	if err != nil {
		handler.Logger.Errorw("service err, EnvironmentListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, result, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) GitListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, GitListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	ok := handler.enforcerUtil.CheckAppRbacForAppOrJob(token, object, casbin.ActionGet)
	if !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	res, err := handler.gitRegistryConfig.GetAll()
	if err != nil {
		handler.Logger.Errorw("service err, GitListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) RegistriesListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	v := r.URL.Query()
	storageType := v.Get("storageType")
	if storageType == "" {
		storageType = repository.OCI_REGISRTY_REPO_TYPE_CONTAINER
	}
	if !slices.Contains(repository.OCI_REGISRTY_REPO_TYPE_LIST, storageType) {
		common.WriteJsonResp(w, fmt.Errorf("invalid query parameters"), nil, http.StatusBadRequest)
		return
	}
	storageAction := v.Get("storageAction")
	if storageAction == "" {
		storageAction = repository.STORAGE_ACTION_TYPE_PUSH
	}
	if !(storageAction == repository.STORAGE_ACTION_TYPE_PULL || storageAction == repository.STORAGE_ACTION_TYPE_PUSH) {
		common.WriteJsonResp(w, fmt.Errorf("invalid query parameters"), nil, http.StatusBadRequest)
		return
	}

	handler.Logger.Infow("request payload, DockerListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	registryConfigs, err := handler.dockerRegistryConfig.ListAllActive()
	if err != nil {
		handler.Logger.Errorw("service err, DockerListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	res := handler.dockerRegistryConfig.FilterRegistryBeanListBasedOnStorageTypeAndAction(registryConfigs, storageType, storageAction, repository.STORAGE_ACTION_TYPE_PULL_AND_PUSH)
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineConfigRestHandlerImpl) TeamListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.Logger.Infow("request payload, TeamListAutocomplete", "appId", appId)
	//RBAC
	object := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	result, err := handler.teamService.FetchForAutocomplete()
	if err != nil {
		handler.Logger.Errorw("service err, TeamListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

	common.WriteJsonResp(w, err, result, http.StatusOK)
}
