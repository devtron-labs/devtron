package restHandler

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/appStore/history"
	"github.com/devtron-labs/devtron/pkg/pipeline"
	history2 "github.com/devtron-labs/devtron/pkg/pipeline/history"
	"github.com/devtron-labs/devtron/pkg/pipeline/history/repository"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type PipelineHistoryRestHandler interface {
	FetchDeployedTemplatesHistory(w http.ResponseWriter, r *http.Request)
	FetchDeployedStrategyHistory(w http.ResponseWriter, r *http.Request)
	FetchDeployedCMHistory(w http.ResponseWriter, r *http.Request)
	FetchDeployedCSHistory(w http.ResponseWriter, r *http.Request)
	FetchDeployedPrePostCdScriptHistory(w http.ResponseWriter, r *http.Request)
}

type PipelineHistoryRestHandlerImpl struct {
	logger                           *zap.SugaredLogger
	userAuthService                  user.UserService
	enforcer                         casbin.Enforcer
	strategyHistoryService           history2.PipelineStrategyHistoryService
	deploymentTemplateHistoryService history2.DeploymentTemplateHistoryService
	configMapHistoryService          history2.ConfigMapHistoryService
	prePostCiScriptHistoryService    history2.PrePostCiScriptHistoryService
	prePostCdScriptHistoryService    history2.PrePostCdScriptHistoryService
	appStoreChartsHistoryService     history.AppStoreChartsHistoryService
	pipelineBuilder                  pipeline.PipelineBuilder
	enforcerUtil                     rbac.EnforcerUtil
}

func NewPipelineHistoryRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService,
	enforcer casbin.Enforcer, strategyHistoryService history2.PipelineStrategyHistoryService,
	deploymentTemplateHistoryService history2.DeploymentTemplateHistoryService,
	configMapHistoryService history2.ConfigMapHistoryService,
	prePostCiScriptHistoryService history2.PrePostCiScriptHistoryService,
	prePostCdScriptHistoryService history2.PrePostCdScriptHistoryService,
	appStoreChartsHistoryService history.AppStoreChartsHistoryService,
	pipelineBuilder pipeline.PipelineBuilder,
	enforcerUtil rbac.EnforcerUtil) *PipelineHistoryRestHandlerImpl {
	return &PipelineHistoryRestHandlerImpl{
		logger:                           logger,
		userAuthService:                  userAuthService,
		enforcer:                         enforcer,
		strategyHistoryService:           strategyHistoryService,
		deploymentTemplateHistoryService: deploymentTemplateHistoryService,
		configMapHistoryService:          configMapHistoryService,
		prePostCdScriptHistoryService:    prePostCdScriptHistoryService,
		prePostCiScriptHistoryService:    prePostCiScriptHistoryService,
		appStoreChartsHistoryService:     appStoreChartsHistoryService,
		pipelineBuilder:                  pipelineBuilder,
		enforcerUtil:                     enforcerUtil,
	}
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedTemplatesHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedTemplatesHistory", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedTemplatesHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedTemplatesHistory", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.deploymentTemplateHistoryService.GetHistoryForDeployedTemplates(pipelineId)
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedTemplates", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedStrategyHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedStrategyHistory", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedStrategyHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedStrategyHistory", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.strategyHistoryService.GetHistoryForDeployedStrategy(pipelineId)
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedStrategy", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedCMHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCMHistory", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCMHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedCMHistory", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapHistoryService.GetHistoryForDeployedCMCS(pipelineId, repository.CONFIGMAP_TYPE)
	if err != nil {
		handler.logger.Errorw("service err, FetchDeployedCMHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedCSHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCSHistory", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCSHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Debugw("request payload, FetchDeployedCSHistory", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.configMapHistoryService.GetHistoryForDeployedCMCS(pipelineId, repository.SECRET_TYPE)
	if err != nil {
		handler.logger.Errorw("service err, FetchDeployedCSHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler PipelineHistoryRestHandlerImpl) FetchDeployedPrePostCdScriptHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedPrePostCdScriptHistory", "err", err, "appId", appId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedPrePostCdScriptHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	stage := r.URL.Query().Get("stage")
	handler.logger.Debugw("request payload, FetchDeployedPrePostCdScriptHistory", "pipelineId", pipelineId)

	//RBAC START
	token := r.Header.Get("token")
	app, err := handler.pipelineBuilder.GetApp(appId)
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	resourceName := handler.enforcerUtil.GetAppRBACName(app.AppName)
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	//RBAC END

	res, err := handler.prePostCdScriptHistoryService.GetHistoryForDeployedPrePostCdScript(pipelineId, repository.CdStageType(stage))
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedPrePostCdScript", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}
