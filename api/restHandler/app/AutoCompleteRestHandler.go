package app

import (
	"context"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
	"net/http"
	"strconv"
)

type DevtronAppAutoCompleteRestHandler interface {
	GitListAutocomplete(w http.ResponseWriter, r *http.Request)
	DockerListAutocomplete(w http.ResponseWriter, r *http.Request)
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
	isActionUserSuperAdmin, err := handler.userAuthService.IsSuperAdmin(int(userId))
	if err != nil {
		common.WriteJsonResp(w, err, "Failed to check admin check", http.StatusInternalServerError)
		return
	}

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
	token := r.Header.Get("token")
	userEmailId, err := handler.userAuthService.GetEmailFromToken(token)
	if err != nil {
		handler.Logger.Errorw("error in getting user emailId from token", "userId", userId, "err", err)
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	accessedApps := make([]*pipeline.AppBean, 0)
	rbacObjects := make([]string, 0)

	var appIdToObjectMap map[int]string
	if len(teamId) == 0 {
		appIdToObjectMap = handler.enforcerUtil.GetRbacObjectsForAllAppsWithMatchingAppName(appName)
	} else {
		appIdToObjectMap = handler.enforcerUtil.GetRbacObjectsForAllAppsWithTeamID(teamIdInt)
	}

	for _, app := range apps {
		object := appIdToObjectMap[app.Id]
		rbacObjects = append(rbacObjects, object)
	}

	enforcedMap := handler.enforcer.EnforceByEmailInBatch(userEmailId, casbin.ResourceApplications, casbin.ActionGet, rbacObjects)
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
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, object); !ok {
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

func (handler PipelineConfigRestHandlerImpl) DockerListAutocomplete(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
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
	res, err := handler.dockerRegistryConfig.ListAllActive()
	if err != nil {
		handler.Logger.Errorw("service err, DockerListAutocomplete", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}

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
