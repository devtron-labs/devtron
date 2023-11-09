package restHandler

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/plugin"
	"github.com/devtron-labs/devtron/pkg/plugin/repository"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type GlobalPluginRestHandler interface {
	GetAllGlobalVariables(w http.ResponseWriter, r *http.Request)
	ListAllPlugins(w http.ResponseWriter, r *http.Request)
	GetPluginDetailById(w http.ResponseWriter, r *http.Request)
}

func NewGlobalPluginRestHandler(logger *zap.SugaredLogger, globalPluginService plugin.GlobalPluginService,
	enforcerUtil rbac.EnforcerUtil, enforcer casbin.Enforcer, pipelineBuilder pipeline.PipelineBuilder,
	userService user.UserService) *GlobalPluginRestHandlerImpl {
	return &GlobalPluginRestHandlerImpl{
		logger:              logger,
		globalPluginService: globalPluginService,
		enforcerUtil:        enforcerUtil,
		enforcer:            enforcer,
		pipelineBuilder:     pipelineBuilder,
		userService:         userService,
	}
}

type GlobalPluginRestHandlerImpl struct {
	logger              *zap.SugaredLogger
	globalPluginService plugin.GlobalPluginService
	enforcerUtil        rbac.EnforcerUtil
	enforcer            casbin.Enforcer
	pipelineBuilder     pipeline.PipelineBuilder
	userService         user.UserService
}

func (handler *GlobalPluginRestHandlerImpl) GetAllGlobalVariables(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	appIdQueryParam := r.URL.Query().Get("appId")
	appId, err := strconv.Atoi(appIdQueryParam)
	if appIdQueryParam == "" || err != nil {
		common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.logger.Infow("service error, GetAllGlobalVariables", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//iteration 1 -
	//using appId for rbac in plugin(global resource), because this data must be visible to person having create permission
	//on atleast one app & we can't check this without iterating through every app
	//TODO: update plugin as a resource in casbin and make rbac independent of appId
	//iteration 2 -
	//adding rbac for branch change resource too, to be removed with implementation on above TODO comment
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	ok1 := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName)
	noEnvObject := handler.enforcerUtil.GetTeamNoEnvRBACNameByAppName(app.AppName)
	ok2 := handler.enforcer.Enforce(token, casbin.ResourceCiPipelineSourceValue, casbin.ActionUpdate, noEnvObject)
	if !ok1 && !ok2 {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	globalVariables, err := handler.globalPluginService.GetAllGlobalVariables()
	if err != nil {
		handler.logger.Errorw("error in getting global variable list", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, globalVariables, http.StatusOK)
}

func (handler *GlobalPluginRestHandlerImpl) ListAllPlugins(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	appIdQueryParam := r.URL.Query().Get("appId")
	appId := 0
	var err error
	if appIdQueryParam != "" {
		appId, err = strconv.Atoi(appIdQueryParam)
		if err != nil {
			common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
			return
		}
	}
	stageType := r.URL.Query().Get("stage")
	if appId > 0 { //check rbac for app, being used for ci pipeline
		app, err := handler.pipelineBuilder.GetApp(appId)
		if err != nil {
			handler.logger.Infow("service error, ListAllPlugins", "err", err, "appId", appId)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		//using appId for rbac in plugin(global resource), because this data must be visible to person having create permission
		//on atleast one app & we can't check this without iterating through every app
		//TODO: update plugin as a resource in casbin and make rbac independent of appId
		//iteration 2 -
		//adding rbac for branch change resource too, to be removed with implementation on above TODO comment
		resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
		ok1 := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName)
		noEnvObject := handler.enforcerUtil.GetTeamNoEnvRBACNameByAppName(app.AppName)
		ok2 := handler.enforcer.Enforce(token, casbin.ResourceCiPipelineSourceValue, casbin.ActionUpdate, noEnvObject)
		if !ok1 && !ok2 {
			common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
			return
		}
	} else { //check for super-admin, to be used in global policy
		userId, err := handler.userService.GetLoggedInUser(r)
		if userId == 0 || err != nil {
			handler.logger.Errorw("request err, userId", "err", err, "payload", userId)
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
			return
		}
		isSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
		if !isSuperAdmin || err != nil {
			if err != nil {
				handler.logger.Errorw("request err, CheckSuperAdmin", "err", isSuperAdmin, "isSuperAdmin", isSuperAdmin)
			}
			common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
			return
		}
	}
	var plugins []*plugin.PluginListComponentDto
	if stageType == repository.CD_STAGE_TYPE {
		plugins, err = handler.globalPluginService.ListAllPlugins(repository.CD)
		if err != nil {
			handler.logger.Errorw("error in getting cd plugin list", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else {
		plugins, err = handler.globalPluginService.ListAllPlugins(repository.CI)
		if err != nil {
			handler.logger.Errorw("error in getting ci plugin list", "err", err)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	common.WriteJsonResp(w, nil, plugins, http.StatusOK)
}

func (handler *GlobalPluginRestHandlerImpl) GetPluginDetailById(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("token")
	appIdQueryParam := r.URL.Query().Get("appId")
	appId, err := strconv.Atoi(appIdQueryParam)
	if appIdQueryParam == "" || err != nil {
		common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
		return
	}
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.logger.Infow("service error, GetPluginDetailById", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//using appId for rbac in plugin(global resource), because this data must be visible to person having create permission
	//on atleast one app & we can't check this without iterating through every app
	//TODO: update plugin as a resource in casbin and make rbac independent of appId
	//iteration 2 -
	//adding rbac for branch change resource too, to be removed with implementation on above TODO comment
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	ok1 := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName)
	noEnvObject := handler.enforcerUtil.GetTeamNoEnvRBACNameByAppName(app.AppName)
	ok2 := handler.enforcer.Enforce(token, casbin.ResourceCiPipelineSourceValue, casbin.ActionUpdate, noEnvObject)
	if !ok1 && !ok2 {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	vars := mux.Vars(r)
	pluginId, err := strconv.Atoi(vars["pluginId"])
	if err != nil {
		handler.logger.Errorw("received invalid pluginId, GetPluginDetailById", "err", err, "pluginId", pluginId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pluginDetail, err := handler.globalPluginService.GetPluginDetailById(pluginId)
	if err != nil {
		handler.logger.Errorw("error in getting plugin detail by id", "err", err, "pluginId", pluginId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, pluginDetail, http.StatusOK)
}
