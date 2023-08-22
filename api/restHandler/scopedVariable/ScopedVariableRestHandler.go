package scopedVariable

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/pkg/variables"
	"github.com/devtron-labs/devtron/pkg/variables/repository"
	"github.com/devtron-labs/devtron/util/rbac"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type ScopedVariableRestHandler interface {
	CreateVariables(w http.ResponseWriter, r *http.Request)
	GetScopedVariables(w http.ResponseWriter, r *http.Request)
}

type ScopedVariableRestHandlerImpl struct {
	logger                *zap.SugaredLogger
	userAuthService       user.UserService
	pipelineBuilder       pipeline.PipelineBuilder
	enforcerUtil          rbac.EnforcerUtil
	enforcer              casbin.Enforcer
	scopedVariableService variables.ScopedVariableService
}

func NewScopedVariableRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService, pipelineBuilder pipeline.PipelineBuilder, enforcerUtil rbac.EnforcerUtil, enforcer casbin.Enforcer, scopedVariableService variables.ScopedVariableService) *ScopedVariableRestHandlerImpl {
	return &ScopedVariableRestHandlerImpl{
		logger:                logger,
		userAuthService:       userAuthService,
		pipelineBuilder:       pipelineBuilder,
		enforcerUtil:          enforcerUtil,
		enforcer:              enforcer,
		scopedVariableService: scopedVariableService,
	}
}
func (handler *ScopedVariableRestHandlerImpl) CreateVariables(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	paylaod := repository.Payload{}
	err = decoder.Decode(&paylaod)
	if err != nil {
		handler.logger.Errorw("request err, Save", "error", err, "payload", paylaod)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	paylaod.UserId = userId
	// not logging bean object as it contains sensitive data
	handler.logger.Infow("request payload received for variables")

	// RBAC enforcer applying
	isSuperAdmin, err := handler.userAuthService.IsSuperAdmin(int(userId))
	if !isSuperAdmin || err != nil {
		if err != nil {
			handler.logger.Errorw("request err, CheckSuperAdmin", "err", err, "isSuperAdmin", isSuperAdmin)
		}
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC enforcer Ends
	err = handler.scopedVariableService.CreateVariables(paylaod)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	}
	common.WriteJsonResp(w, nil, nil, http.StatusOK)
}
func (handler *ScopedVariableRestHandlerImpl) GetScopedVariables(w http.ResponseWriter, r *http.Request) {
	appIdQueryParam := r.URL.Query().Get("appId")
	var appId int
	var err error
	if appIdQueryParam != "" {
		appId, err = strconv.Atoi(appIdQueryParam)
		if err != nil {
			common.WriteJsonResp(w, err, "invalid appId", http.StatusBadRequest)
			return
		}
	}

	var envId int
	envIdQueryParam := r.URL.Query().Get("envId")
	if envIdQueryParam != "" {
		envId, err = strconv.Atoi(envIdQueryParam)
		if err != nil {
			common.WriteJsonResp(w, err, "invalid envId", http.StatusBadRequest)
			return
		}
	}

	var clusterId int
	clusterIdQueryParam := r.URL.Query().Get("clusterId")
	if clusterIdQueryParam != "" {
		clusterId, err = strconv.Atoi(clusterIdQueryParam)
		if err != nil {
			common.WriteJsonResp(w, err, "invalid clusterId", http.StatusBadRequest)
			return
		}
	}

	//vars := mux.Vars(r)
	//environmentId, err := strconv.Atoi(vars["environmentId"])
	//if err != nil {
	//	common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	//	return
	//}
	//appId, err := strconv.Atoi(vars["appId"])
	//if err != nil {
	//	common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	//	return
	//}
	//clusterId, err := strconv.Atoi(vars["clusterId"])
	//if err != nil {
	//	common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
	//	return
	//}

	token := r.Header.Get("token")
	var app *bean.CreateAppDTO
	app, err = handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		handler.logger.Errorw("service err, GetScopedVariables", "err", err, "payload", appId, envId, clusterId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Infow("request payload, GetScopedVariables", "payload", appId, envId, clusterId)
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if appId == 0 && envId == 0 && clusterId == 0 {
		return
	}
	var scopedVariableData []*variables.ScopedVariableData
	scope := variables.Scope{
		AppId:     appId,
		EnvId:     envId,
		ClusterId: clusterId,
	}
	scopedVariableData, err = handler.scopedVariableService.GetScopedVariables(scope, nil)
	common.WriteJsonResp(w, err, scopedVariableData, http.StatusOK)

}
