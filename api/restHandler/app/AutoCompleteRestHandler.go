package app

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
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
	v := r.URL.Query()
	teamId := v.Get("teamId")
	handler.Logger.Infow("request payload, GetAppListForAutocomplete", "teamId", teamId)
	var apps []pipeline.AppBean
	if len(teamId) == 0 {
		apps, err = handler.pipelineBuilder.GetAppList()
		if err != nil {
			handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else {
		teamId, err := strconv.Atoi(teamId)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		} else {
			apps, err = handler.pipelineBuilder.FindAppsByTeamId(teamId)
			if err != nil {
				handler.Logger.Errorw("service err, GetAppListForAutocomplete", "err", err, "teamId", teamId)
				common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
				return
			}
		}
	}

	token := r.Header.Get("token")
	var accessedApps []pipeline.AppBean
	// RBAC
	objects := handler.enforcerUtil.GetRbacObjectsForAllApps()
	for _, app := range apps {
		object := objects[app.Id]
		if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); ok {
			accessedApps = append(accessedApps, app)
		}
	}
	// RBAC
	if len(accessedApps) == 0 {
		accessedApps = make([]pipeline.AppBean, 0)
	}
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
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC
	result, err := handler.envService.GetEnvironmentListForAutocomplete()
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
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
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
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
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
	if ok := handler.enforcer.Enforce(token, rbac.ResourceApplications, rbac.ActionGet, object); !ok {
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