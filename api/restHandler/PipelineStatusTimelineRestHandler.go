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
)

type PipelineStatusTimelineRestHandler interface {
	FetchTimelines(w http.ResponseWriter, r *http.Request)
}

type PipelineStatusTimelineRestHandlerImpl struct {
	logger                        *zap.SugaredLogger
	pipelineStatusTimelineService app.PipelineStatusTimelineService
	enforcerUtil                  rbac.EnforcerUtil
	enforcer                      casbin.Enforcer
}

func NewPipelineStatusTimelineRestHandlerImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineService app.PipelineStatusTimelineService, enforcerUtil rbac.EnforcerUtil,
	enforcer casbin.Enforcer) *PipelineStatusTimelineRestHandlerImpl {
	return &PipelineStatusTimelineRestHandlerImpl{
		logger:                        logger,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
		enforcerUtil:                  enforcerUtil,
		enforcer:                      enforcer,
	}
}

func (handler *PipelineStatusTimelineRestHandlerImpl) FetchTimelines(w http.ResponseWriter, r *http.Request) {
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

	timelines, err := handler.pipelineStatusTimelineService.FetchTimelines(appId, envId, wfrId)
	if err != nil {
		handler.logger.Errorw("error in getting cd pipeline status timelines by wfrId", "err", err, "wfrId", wfrId)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
		return
	}
	common.WriteJsonResp(w, err, timelines, http.StatusOK)
	return
}
