package restHandler

import (
	"encoding/json"
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
	PatchPlugin(w http.ResponseWriter, r *http.Request)

	GetAllGlobalVariables(w http.ResponseWriter, r *http.Request)
	ListAllPlugins(w http.ResponseWriter, r *http.Request)
	GetPluginDetailById(w http.ResponseWriter, r *http.Request)
	GetDetailedPluginInfoByPluginId(w http.ResponseWriter, r *http.Request)
	GetAllDetailedPluginInfo(w http.ResponseWriter, r *http.Request)
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

func (handler *GlobalPluginRestHandlerImpl) PatchPlugin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	var pluginDataDto plugin.PluginMetadataDto
	err = decoder.Decode(&pluginDataDto)
	if err != nil {
		handler.logger.Errorw("request err, PatchPlugin", "error", err, "payload", pluginDataDto)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload received for patching plugins", pluginDataDto, "userId", userId)
	// RBAC enforcer applying
	isSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if !isSuperAdmin || err != nil {
		if err != nil {
			handler.logger.Errorw("request err, CheckSuperAdmin", "err", err, "isSuperAdmin", isSuperAdmin)
		}
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	pluginData, err := handler.globalPluginService.PatchPlugin(&pluginDataDto, userId)
	if err != nil {
		handler.logger.Errorw("error in patching plugin data", "action", pluginDataDto.Action, "pluginMetadataPayloadDto", pluginDataDto, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, pluginData, http.StatusOK)

}
func (handler *GlobalPluginRestHandlerImpl) GetDetailedPluginInfoByPluginId(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pluginId, err := strconv.Atoi(vars["pluginId"])
	if err != nil {
		handler.logger.Errorw("error in converting from string to integer", "userId", userId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	// RBAC enforcer applying
	isSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if !isSuperAdmin || err != nil {
		if err != nil {
			handler.logger.Errorw("request err, CheckSuperAdmin", "err", err, "isSuperAdmin", isSuperAdmin)
		}
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	pluginMetaData, err := handler.globalPluginService.GetDetailedPluginInfoByPluginId(pluginId)
	if err != nil {
		handler.logger.Errorw("error in getting plugin metadata", "pluginId", pluginId, "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, pluginMetaData, http.StatusOK)
}
func (handler *GlobalPluginRestHandlerImpl) GetAllDetailedPluginInfo(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	// RBAC enforcer applying
	isSuperAdmin, err := handler.userService.IsSuperAdmin(int(userId))
	if !isSuperAdmin || err != nil {
		if err != nil {
			handler.logger.Errorw("request err, CheckSuperAdmin", "err", err, "isSuperAdmin", isSuperAdmin)
		}
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	pluginMetaData, err := handler.globalPluginService.GetAllDetailedPluginInfo()
	if err != nil {
		handler.logger.Errorw("error in getting all plugins metadata", "err", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, nil, pluginMetaData, http.StatusOK)
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
	//using appId for rbac in plugin(global resource), because this data must be visible to person having create permission
	//on atleast one app & we can't check this without iterating through every app
	//TODO: update plugin as a resource in casbin and make rbac independent of appId
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
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
	appId, err := strconv.Atoi(appIdQueryParam)
	if appIdQueryParam == "" || err != nil {
		common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
		return
	}
	stageType := r.URL.Query().Get("stage")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.logger.Infow("service error, ListAllPlugins", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	//using appId for rbac in plugin(global resource), because this data must be visible to person having create permission
	//on atleast one app & we can't check this without iterating through every app
	//TODO: update plugin as a resource in casbin and make rbac independent of appId
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
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

	common.WriteJsonResp(w, err, plugins, http.StatusOK)
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
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionCreate, resourceName); !ok {
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
