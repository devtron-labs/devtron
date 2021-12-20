package client

import (
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type HelmAppRestHandler interface {
	ListApplications(w http.ResponseWriter, r *http.Request)
}
type HelmAppRestHandlerImpl struct {
	logger         *zap.SugaredLogger
	helmAppService HelmAppService
}

func NewHelmAppRestHandlerImpl(logger *zap.SugaredLogger,
	helmAppService HelmAppService) *HelmAppRestHandlerImpl {
	return &HelmAppRestHandlerImpl{
		logger:         logger,
		helmAppService: helmAppService,
	}
}

func (handler *HelmAppRestHandlerImpl) ListApplications(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	clusterIdString := vars["clusterIds"]
	clusterIdSlices := strings.Split(clusterIdString, ",")
	var clusterIds []int
	for _, is := range clusterIdSlices {
		j, err := strconv.Atoi(is)
		if err != nil {
			handler.logger.Errorw("request err, CreateUser", "err", err, "payload", clusterIds)
			common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
			return
		}
		clusterIds = append(clusterIds, j)
	}
	handler.helmAppService.ListHelmApplications(clusterIds, w)
}
