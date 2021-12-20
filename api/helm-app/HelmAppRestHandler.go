package client

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"go.uber.org/zap"
	"net/http"
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

type ClusterIds struct {
	ClusterIds []int `json:"clusterIds"`
}

func (handler *HelmAppRestHandlerImpl) ListApplications(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var clusterIds ClusterIds
	err := decoder.Decode(&clusterIds)
	if err != nil {
		handler.logger.Errorw("request err, CreateUser", "err", err, "payload", clusterIds)
		common.WriteJsonResp(w, err, nil, http.StatusBadRequest)
		return
	}
	handler.helmAppService.ListHelmApplications(clusterIds.ClusterIds, w)
}
