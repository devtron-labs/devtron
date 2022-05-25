package restHandler

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	history2 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type PipelineHistoryRestHandler interface {
	FetchDeployedConfigurationsForWorkflow(w http.ResponseWriter, r *http.Request)
	FetchDeployedHistoryComponentList(w http.ResponseWriter, r *http.Request)
	FetchDeployedHistoryComponentDetail(w http.ResponseWriter, r *http.Request)
}

type PipelineHistoryRestHandlerImpl struct {
	logger                              *zap.SugaredLogger
	userAuthService                     user.UserService
	enforcer                            casbin.Enforcer
	strategyHistoryService              history2.PipelineStrategyHistoryService
	deploymentTemplateHistoryService    history2.DeploymentTemplateHistoryService
	configMapHistoryService             history2.ConfigMapHistoryService
	prePostCiScriptHistoryService       history2.PrePostCiScriptHistoryService
	prePostCdScriptHistoryService       history2.PrePostCdScriptHistoryService
	pipelineBuilder                     pipeline.PipelineBuilder
	enforcerUtil                        rbac.EnforcerUtil
	deployedConfigurationHistoryService history2.DeployedConfigurationHistoryService
}

func NewPipelineHistoryRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService,
	enforcer casbin.Enforcer, strategyHistoryService history2.PipelineStrategyHistoryService,
	deploymentTemplateHistoryService history2.DeploymentTemplateHistoryService,
	configMapHistoryService history2.ConfigMapHistoryService,
	prePostCiScriptHistoryService history2.PrePostCiScriptHistoryService,
	prePostCdScriptHistoryService history2.PrePostCdScriptHistoryService,
	pipelineBuilder pipeline.PipelineBuilder,
	enforcerUtil rbac.EnforcerUtil,
	deployedConfigurationHistoryService history2.DeployedConfigurationHistoryService) *PipelineHistoryRestHandlerImpl {
	return &PipelineHistoryRestHandlerImpl{
		logger:                              logger,
		userAuthService:                     userAuthService,
		enforcer:                            enforcer,
		strategyHistoryService:              strategyHistoryService,
		deploymentTemplateHistoryService:    deploymentTemplateHistoryService,
		configMapHistoryService:             configMapHistoryService,
		prePostCdScriptHistoryService:       prePostCdScriptHistoryService,
		prePostCiScriptHistoryService:       prePostCiScriptHistoryService,
		pipelineBuilder:                     pipelineBuilder,
		enforcerUtil:                        enforcerUtil,
		deployedConfigurationHistoryService: deployedConfigurationHistoryService,
	}
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedConfigurationsForWorkflow(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedConfigurationsForWorkflow", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedConfigurationsForWorkflow", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	wfrId, err := strconv.Atoi(vars["wfrId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedConfigurationsForWorkflow", "err", err, "wfrId", wfrId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedConfigurationsForWorkflow", "pipelineId", pipelineId, "wfrId", wfrId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.deployedConfigurationHistoryService.GetDeployedConfigurationByWfrId(pipelineId, wfrId)
	if err != nil {
		handler.logger.Errorw("service err, GetDeployedConfigurationByWfrId", "err", err, "pipelineId", pipelineId, "wfrId", wfrId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedHistoryComponentList(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedHistoryComponentList", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedHistoryComponentList", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}

	historyComponent := r.URL.Query().Get("historyComponent")
	if historyComponent == "" || err != nil {
		handler.logger.Errorw("request err, FetchDeployedHistoryComponentList", "err", err, "historyComponent", historyComponent)
		common.WriteJsonResp(w, err, "invalid historyComponent", http.StatusBadRequest)
		return
	}
	historyComponentName := r.URL.Query().Get("historyComponentName")
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedHistoryComponentList", "err", err, "historyComponentName", historyComponentName)
		common.WriteJsonResp(w, err, "invalid historyComponentName", http.StatusBadRequest)
		return
	}
	baseConfigurationIdParam := r.URL.Query().Get("baseConfigurationId")
	baseConfigurationId, err := strconv.Atoi(baseConfigurationIdParam)
	if baseConfigurationId == 0 || err != nil {
		handler.logger.Errorw("request err, FetchDeployedHistoryComponentList", "err", err, "baseConfigurationId", baseConfigurationId)
		common.WriteJsonResp(w, err, "invalid baseConfigurationId", http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedHistoryComponentList", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.deployedConfigurationHistoryService.GetDeployedHistoryComponentList(pipelineId, baseConfigurationId, historyComponent, historyComponentName)
	if err != nil {
		handler.logger.Errorw("service err, GetDeployedHistoryComponentList", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler *PipelineHistoryRestHandlerImpl) FetchDeployedHistoryComponentDetail(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedHistoryComponentDetail", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedHistoryComponentDetail", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedHistoryComponentDetail", "err", err, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	historyComponent := r.URL.Query().Get("historyComponent")
	if historyComponent == "" || err != nil {
		handler.logger.Errorw("request err, FetchDeployedHistoryComponentDetail", "err", err, "historyComponent", historyComponent)
		common.WriteJsonResp(w, err, "invalid historyComponent", http.StatusBadRequest)
		return
	}
	historyComponentName := r.URL.Query().Get("historyComponentName")
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedHistoryComponentDetail", "err", err, "historyComponentName", historyComponentName)
		common.WriteJsonResp(w, err, "invalid historyComponentName", http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedHistoryComponentDetail", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END
	//checking if user has admin access
	userHasAdminAccess := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionUpdate, resourceName)
	res, err := handler.deployedConfigurationHistoryService.GetDeployedHistoryComponentDetail(pipelineId, id, historyComponent, historyComponentName, userHasAdminAccess)
	if err != nil {
		handler.logger.Errorw("service err, GetDeployedHistoryComponentDetail", "err", err, "pipelineId", pipelineId, "id", id)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
