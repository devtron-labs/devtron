package restHandler

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/pkg/app"
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
}

func NewPipelineStatusTimelineRestHandlerImpl(logger *zap.SugaredLogger,
	pipelineStatusTimelineService app.PipelineStatusTimelineService) *PipelineStatusTimelineRestHandlerImpl {
	return &PipelineStatusTimelineRestHandlerImpl{
		logger:                        logger,
		pipelineStatusTimelineService: pipelineStatusTimelineService,
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
	//TODO: update rbac
	var timelines []*app.PipelineStatusTimelineDto
	if wfrId == 0 {
		timelines, err = handler.pipelineStatusTimelineService.FetchTimelinesForLatestTriggerByAppIdAndEnvId(appId, envId)
		if err != nil {
			handler.logger.Errorw("error in getting cd pipeline status timelines by appId & envId", "err", err, "appId", appId, "envId", envId)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	} else {
		timelines, err = handler.pipelineStatusTimelineService.FetchTimelinesByWfrId(wfrId)
		if err != nil {
			handler.logger.Errorw("error in getting cd pipeline status timelines by wfrId", "err", err, "wfrId", wfrId)
			common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
			return
		}
	}
	common.WriteJsonResp(w, err, timelines, http.StatusOK)
	return
}
