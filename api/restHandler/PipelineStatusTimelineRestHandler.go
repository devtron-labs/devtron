package restHandler

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/user/casbin"
	"github.com/devtron-labs/devtron/util/rbac"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type PipelineStatusTimelineRestHandler interface {
	FetchTimelines(w http.ResponseWriter, r *http.Request)
}

type PipelineStatusTimelineRestHandlerImpl struct {
	logger                        *zap.SugaredLogger
	pipelineStatusTimelineService app.PipelineStatusTimelineService
	enforcerUtil                  rbac.EnforcerUtil
	enforcer                      casbin.Enforcer
	//config                        sql.Config
}

func NewPipelineStatusTimelineRestHandlerImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineService app.PipelineStatusTimelineService, enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer) *PipelineStatusTimelineRestHandlerImpl {
	return &PipelineStatusTimelineRestHandlerImpl{
		logger:                        logger,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
		enforcerUtil:                  enforcerUtil,
		enforcer:                      enforcer,
		//config:                        config,
	}
}

func (handler *PipelineStatusTimelineRestHandlerImpl) FetchTimelines(w http.ResponseWriter, r *http.Request) {
	logTimeStart := time.Now()
	logPrintDuration := 5
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["appId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	envId, err := strconv.Atoi(vars["envId"])
	if err != nil {
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	wfrId := 0
	wfrIdParam := r.URL.Query().Get("wfrId")
	if len(wfrIdParam) != 0 {
		wfrId, err = strconv.Atoi(wfrIdParam)
		if err != nil {
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
	}
	resourceName := handler.enforcerUtil.GetAppRBACNameByAppId(appId)
	token := r.Header.Get("token")
	if ok := handler.enforcer.Enforce(token, casbin.ResourceApplications, casbin.ActionGet, resourceName); !ok {
		common.WriteJsonResp(w, fmt.Errorf("unauthorized user"), "Unauthorized User", http.StatusForbidden)
		return
	}
	if time.Since(logTimeStart) > (time.Second * time.Duration(logPrintDuration)) {
		handler.logger.Errorw("pipelineStatusTimelineRestHandler processing time high, FetchTimelines.Enforcer", "timeDuration", time.Since(logTimeStart))
		logTimeStart = time.Now()
	}
	timelines, err := handler.pipelineStatusTimelineService.FetchTimelines(appId, envId, wfrId, logPrintDuration)
	if err != nil {
		handler.logger.Errorw("error in getting cd pipeline status timelines by wfrId", "err", err, "wfrId", wfrId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	if time.Since(logTimeStart) > (time.Second * time.Duration(logPrintDuration)) {
		handler.logger.Errorw("pipelineStatusTimelineRestHandler processing time high, FetchTimelines.FetchTimelines", "timeDuration", time.Since(logTimeStart))
		logTimeStart = time.Now()
	}
	common.WriteJsonResp(w, err, timelines, http.StatusOK)
	return
}
