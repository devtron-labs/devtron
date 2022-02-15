package restHandler

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	history2 "github.com/devtron-labs/devtron/internal/sql/repository/history"
	"github.com/devtron-labs/devtron/pkg/history"
	"github.com/devtron-labs/devtron/pkg/user"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
)

type HistoryRestHandler interface {
	FetchDeployedChartsHistory(w http.ResponseWriter, r *http.Request)
	FetchDeployedStrategyHistory(w http.ResponseWriter, r *http.Request)
	FetchDeployedCmCsHistory(w http.ResponseWriter, r *http.Request)
	FetchDeployedCdConfigHistory(w http.ResponseWriter, r *http.Request)
}

type HistoryRestHandlerImpl struct {
	logger                     *zap.SugaredLogger
	userAuthService            user.UserService
	enforcer                   casbin.Enforcer
	strategyHistoryService     history.PipelineStrategyHistoryService
	chartsHistoryService       history.ChartsHistoryService
	configMapHistoryService    history.ConfigMapHistoryService
	ciScriptHistoryService     history.CiScriptHistoryService
	cdConfigHistoryService     history.CdConfigHistoryService
	installedAppHistoryService history.InstalledAppHistoryService
}

func NewHistoryRestHandlerImpl(logger *zap.SugaredLogger, userAuthService user.UserService,
	enforcer casbin.Enforcer) *HistoryRestHandlerImpl {
	return &HistoryRestHandlerImpl{
		logger:          logger,
		userAuthService: userAuthService,
		enforcer:        enforcer,
	}
}

func (handler HistoryRestHandlerImpl) FetchDeployedChartsHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedChartsHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Errorw("request payload, FetchDeployedChartsHistory", "pipelineId", pipelineId)

	//RBAC START

	//RBAC END

	res, err := handler.chartsHistoryService.GetHistoryForDeployedCharts(pipelineId)
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedCharts", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler HistoryRestHandlerImpl) FetchDeployedStrategyHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedStrategyHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Errorw("request payload, FetchDeployedStrategyHistory", "pipelineId", pipelineId)

	//RBAC START

	//RBAC END

	res, err := handler.strategyHistoryService.GetHistoryForDeployedStrategy(pipelineId)
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedStrategy", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler HistoryRestHandlerImpl) FetchDeployedCmCsHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCmCsHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.logger.Errorw("request payload, FetchDeployedCmCsHistory", "pipelineId", pipelineId)

	//RBAC START

	//RBAC END

	res, err := handler.configMapHistoryService.GetHistoryForDeployedCMCS(pipelineId)
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedCMCS", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}

func (handler HistoryRestHandlerImpl) FetchDeployedCdConfigHistory(w http.ResponseWriter, r *http.Request) {
	userId, err := handler.userAuthService.GetLoggedInUser(r)
	if userId == 0 || err != nil {
		common.WriteJsonResp(w, err, "Unauthorized User", http.StatusUnauthorized)
		return
	}
	vars := mux.Vars(r)
	pipelineId, err := strconv.Atoi(vars["pipelineId"])
	if err != nil {
		handler.logger.Errorw("request err, FetchDeployedCdConfigHistory", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	stage := r.URL.Query().Get("stage")
	handler.logger.Errorw("request payload, FetchDeployedCdConfigHistory", "pipelineId", pipelineId)

	//RBAC START

	//RBAC END

	res, err := handler.cdConfigHistoryService.GetHistoryForDeployedCdConfig(pipelineId, history2.CdStageType(stage))
	if err != nil {
		handler.logger.Errorw("service err, GetHistoryForDeployedCdConfig", "err", err, "pipelineId", pipelineId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, res, http.StatusOK)
}